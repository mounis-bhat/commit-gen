package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"

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

// ReadAvailableFiles reads unstaged and untracked files.
func ReadAvailableFiles() tea.Cmd {
	return func() tea.Msg {
		unstaged, err := git.GetUnstagedFiles()
		if err != nil {
			return ErrorMsg{Err: err}
		}

		untracked, err := git.GetUntrackedFiles()
		if err != nil {
			return ErrorMsg{Err: err}
		}

		return FilesReadyMsg{Unstaged: unstaged, Untracked: untracked}
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
		commitMsg, err := provider.GenerateCommitMessage(ctx, diff)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return CommitGeneratedMsg{Commit: commitMsg}
	}
}

// ExecuteCommit executes the generated commit command.
func ExecuteCommit(commitMsg string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("bash", "-c", commitMsg)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to execute commit: %w", err)}
		}

		return CommitExecutedMsg{}
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
