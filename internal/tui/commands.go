package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mounis-bhat/commit-gen/internal/ai"
	"github.com/mounis-bhat/commit-gen/internal/ai/providers"
	"github.com/mounis-bhat/commit-gen/internal/config"
	"github.com/mounis-bhat/commit-gen/internal/git"
)

// LoadConfig loads the configuration from the config file.
func LoadConfig() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load()
		if err != nil {
			// Return empty config with defaults
			return ConfigLoadedMsg{Config: &config.Config{
				OllamaURL:   "http://localhost:11434",
				OllamaModel: "qwen2.5-coder:3b",
			}}
		}
		return ConfigLoadedMsg{Config: cfg}
	}
}

// SaveConfig saves the configuration to the config file.
func SaveConfig(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		if err := config.Save(cfg); err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to save config: %w", err)}
		}
		return APIKeySavedMsg{}
	}
}

// ReadGitDiff reads the staged git diff.
// Returns empty diff if no staged changes (caller should handle this).
func ReadGitDiff() tea.Cmd {
	return func() tea.Msg {
		diff, err := git.GetDiff()
		if err != nil {
			return ErrorMsg{Err: err}
		}
		// Return diff even if empty - caller will handle showing file picker
		return DiffReadyMsg{Diff: diff}
	}
}

// ReadAvailableFiles reads staged, unstaged, and untracked files.
func ReadAvailableFiles() tea.Cmd {
	return func() tea.Msg {
		staged, err := git.GetStagedFiles()
		if err != nil {
			return ErrorMsg{Err: err}
		}

		unstaged, err := git.GetUnstagedFiles()
		if err != nil {
			return ErrorMsg{Err: err}
		}

		untracked, err := git.GetUntrackedFiles()
		if err != nil {
			return ErrorMsg{Err: err}
		}

		return FilesReadyMsg{Staged: staged, Unstaged: unstaged, Untracked: untracked}
	}
}

// StageSelectedFiles stages the given files.
func StageSelectedFiles(files []string) tea.Cmd {
	return func() tea.Msg {
		if err := git.StageFiles(files); err != nil {
			return ErrorMsg{Err: err}
		}
		return FilesStagedMsg{}
	}
}

// UnstageSelectedFiles unstages the given files.
func UnstageSelectedFiles(files []string) tea.Cmd {
	return func() tea.Msg {
		if err := git.UnstageFiles(files); err != nil {
			return ErrorMsg{Err: err}
		}
		return FilesUnstagedMsg{}
	}
}

// ApplyStagingChanges applies both staging and unstaging in one operation.
func ApplyStagingChanges(toStage, toUnstage []string) tea.Cmd {
	return func() tea.Msg {
		// First unstage files that should be removed
		if len(toUnstage) > 0 {
			if err := git.UnstageFiles(toUnstage); err != nil {
				return ErrorMsg{Err: err}
			}
		}

		// Then stage files that should be added
		if len(toStage) > 0 {
			if err := git.StageFiles(toStage); err != nil {
				return ErrorMsg{Err: err}
			}
		}

		return FilesStagedMsg{}
	}
}

// ExecutePush pushes commits to the remote repository.
func ExecutePush() tea.Cmd {
	return func() tea.Msg {
		if err := git.Push(); err != nil {
			return ErrorMsg{Err: err}
		}
		return PushExecutedMsg{}
	}
}

// GenerateCommit generates a commit message using the AI provider.
func GenerateCommit(provider ai.Provider, diff string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		commitMsg, err := provider.GenerateCommitMessage(ctx, diff, runtime.GOOS)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return CommitGeneratedMsg{Commit: commitMsg}
	}
}

// parseCommitMessages extracts commit messages from the AI-generated git commit command.
// It handles both single-line format (Windows) and multi-line format (POSIX).
func parseCommitMessages(commitCmd string) ([]string, error) {
	commitCmd = strings.TrimSpace(commitCmd)
	if commitCmd == "" {
		return nil, fmt.Errorf("empty commit command")
	}

	// Remove "git commit" prefix if present
	commitCmd = strings.TrimPrefix(commitCmd, "git commit")
	commitCmd = strings.TrimSpace(commitCmd)

	// Remove backslash line continuations (POSIX multiline format)
	commitCmd = strings.ReplaceAll(commitCmd, "\\\n", " ")
	commitCmd = strings.ReplaceAll(commitCmd, "\\\r\n", " ")

	var messages []string

	// Parse -m "message" patterns
	// We need to handle both: -m "message" and -m 'message'
	remaining := commitCmd
	for {
		remaining = strings.TrimSpace(remaining)
		if remaining == "" {
			break
		}

		// Look for -m flag
		if !strings.HasPrefix(remaining, "-m") {
			// Skip unknown content
			idx := strings.Index(remaining, "-m")
			if idx == -1 {
				break
			}
			remaining = remaining[idx:]
			continue
		}

		// Skip "-m" and any whitespace
		remaining = strings.TrimPrefix(remaining, "-m")
		remaining = strings.TrimSpace(remaining)

		if remaining == "" {
			break
		}

		// Determine the quote character
		var quote byte
		if remaining[0] == '"' {
			quote = '"'
		} else if remaining[0] == '\'' {
			quote = '\''
		} else {
			// No quote, take until next space or -m
			endIdx := strings.Index(remaining, " -m")
			if endIdx == -1 {
				endIdx = len(remaining)
			}
			messages = append(messages, remaining[:endIdx])
			remaining = remaining[endIdx:]
			continue
		}

		// Find the closing quote, handling escaped quotes
		remaining = remaining[1:] // skip opening quote
		var msg strings.Builder
		escaped := false
		foundClose := false

		for i := 0; i < len(remaining); i++ {
			c := remaining[i]
			if escaped {
				// Handle common escape sequences
				switch c {
				case 'n':
					msg.WriteByte('\n')
				case 't':
					msg.WriteByte('\t')
				case '\\':
					msg.WriteByte('\\')
				case '"':
					msg.WriteByte('"')
				case '\'':
					msg.WriteByte('\'')
				default:
					// Keep the backslash for unknown escapes
					msg.WriteByte('\\')
					msg.WriteByte(c)
				}
				escaped = false
				continue
			}

			if c == '\\' {
				escaped = true
				continue
			}

			if c == quote {
				remaining = remaining[i+1:]
				foundClose = true
				break
			}

			msg.WriteByte(c)
		}

		if !foundClose {
			// No closing quote found, use rest of string
			remaining = ""
		}

		if msg.Len() > 0 {
			messages = append(messages, msg.String())
		}
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("no commit messages found in command")
	}

	return messages, nil
}

// ExecuteCommit executes the generated commit command.
func ExecuteCommit(commitMsg string) tea.Cmd {
	return func() tea.Msg {
		messages, err := parseCommitMessages(commitMsg)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to parse commit command: %w", err)}
		}

		// Build git commit command with -m arguments
		args := []string{"commit"}
		for _, msg := range messages {
			args = append(args, "-m", msg)
		}

		cmd := exec.Command("git", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to execute commit: %w", err)}
		}

		return CommitExecutedMsg{}
	}
}

// FetchOpenAIModels fetches available OpenAI models (also validates the API key).
func FetchOpenAIModels(apiKey string) tea.Cmd {
	return func() tea.Msg {
		provider, err := providers.NewOpenAIProvider(apiKey, "")
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("invalid OpenAI API key: %w", err)}
		}
		ctx := context.Background()
		models, err := provider.ListModels(ctx)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to fetch OpenAI models: %w", err)}
		}
		return ModelsFetchedMsg{Models: models}
	}
}

// FetchOllamaModels fetches available Ollama models.
func FetchOllamaModels(baseURL string) tea.Cmd {
	return func() tea.Msg {
		provider, err := providers.NewOllamaProvider(baseURL, "")
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to create Ollama provider: %w", err)}
		}
		ctx := context.Background()
		models, err := provider.ListModels(ctx)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to fetch Ollama models: %w", err)}
		}
		return ModelsFetchedMsg{Models: models}
	}
}
