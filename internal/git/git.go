package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// GetDiff returns the staged git diff.
func GetDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}

	return out.String(), nil
}

// HasStagedChanges checks if there are any staged changes.
func HasStagedChanges() (bool, error) {
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	err := cmd.Run()
	if err != nil {
		// Exit code 1 means there are changes
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return true, nil
			}
		}
		return false, fmt.Errorf("git diff failed: %w", err)
	}
	// Exit code 0 means no changes
	return false, nil
}

// FileStatus represents the status of a file in git.
type FileStatus struct {
	Path   string
	Status string // "modified", "deleted", "renamed", "untracked"
}

// GetUnstagedFiles returns a list of modified/deleted files that are not staged.
func GetUnstagedFiles() ([]FileStatus, error) {
	cmd := exec.Command("git", "diff", "--name-status")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	var files []FileStatus
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 2 {
			continue
		}
		status := "modified"
		switch parts[0] {
		case "D":
			status = "deleted"
		case "R":
			status = "renamed"
		case "M":
			status = "modified"
		}
		files = append(files, FileStatus{Path: parts[1], Status: status})
	}

	return files, nil
}

// GetUntrackedFiles returns a list of untracked files.
func GetUntrackedFiles() ([]FileStatus, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git ls-files failed: %w", err)
	}

	var files []FileStatus
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		files = append(files, FileStatus{Path: line, Status: "untracked"})
	}

	return files, nil
}

// StageFiles stages the given files using git add.
func StageFiles(files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files to stage")
	}

	args := append([]string{"add"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	return nil
}

// Push pushes commits to the remote repository.
func Push() error {
	cmd := exec.Command("git", "push")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	return nil
}
