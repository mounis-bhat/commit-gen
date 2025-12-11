package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
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
