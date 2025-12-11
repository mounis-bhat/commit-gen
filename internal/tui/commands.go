package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mounis-bhat/commit-gen/internal/ai"
	"github.com/mounis-bhat/commit-gen/internal/config"
	"github.com/mounis-bhat/commit-gen/internal/git"
)

// LoadAPIKey loads the API key from the config file.
func LoadAPIKey() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load()
		if err != nil {
			return APIKeyLoadedMsg{Key: ""}
		}
		return APIKeyLoadedMsg{Key: cfg.APIKey}
	}
}

// SaveAPIKey saves the API key to the config file.
func SaveAPIKey(key string) tea.Cmd {
	return func() tea.Msg {
		cfg := &config.Config{APIKey: key}
		if err := config.Save(cfg); err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to save config: %w", err)}
		}
		return APIKeySavedMsg{}
	}
}

// ReadGitDiff reads the staged git diff.
func ReadGitDiff() tea.Cmd {
	return func() tea.Msg {
		diff, err := git.GetDiff()
		if err != nil {
			return ErrorMsg{Err: err}
		}
		if strings.TrimSpace(diff) == "" {
			return ErrorMsg{Err: fmt.Errorf("no staged changes detected. Use 'git add' to stage changes first")}
		}
		return DiffReadyMsg{Diff: diff}
	}
}

// GenerateCommit generates a commit message using the AI.
func GenerateCommit(apiKey, diff string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		commitMsg, err := ai.GenerateCommitMessage(ctx, apiKey, diff)
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
