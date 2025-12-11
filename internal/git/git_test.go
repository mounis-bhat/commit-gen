package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestHelper provides utility functions for git testing
type TestHelper struct {
	t       *testing.T
	repoDir string
	oldDir  string
}

// NewTestHelper creates a new test helper with a temporary git repository
func NewTestHelper(t *testing.T) *TestHelper {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Save current directory
	oldDir, err := os.Getwd()
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to get current dir: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to change to temp dir: %v", err)
	}

	// Initialize git repository
	if err := exec.Command("git", "init").Run(); err != nil {
		os.Chdir(oldDir)
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git user for commits
	exec.Command("git", "config", "user.email", "test@test.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	return &TestHelper{
		t:       t,
		repoDir: tmpDir,
		oldDir:  oldDir,
	}
}

// Cleanup removes the temporary repository and restores the original directory
func (h *TestHelper) Cleanup() {
	os.Chdir(h.oldDir)
	os.RemoveAll(h.repoDir)
}

// CreateFile creates a file with the given content
func (h *TestHelper) CreateFile(name, content string) {
	h.t.Helper()
	path := filepath.Join(h.repoDir, name)

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		h.t.Fatalf("failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		h.t.Fatalf("failed to create file %s: %v", name, err)
	}
}

// StageFile stages a file
func (h *TestHelper) StageFile(name string) {
	h.t.Helper()
	if err := exec.Command("git", "add", name).Run(); err != nil {
		h.t.Fatalf("failed to stage file %s: %v", name, err)
	}
}

// Commit creates a commit with the given message
func (h *TestHelper) Commit(message string) {
	h.t.Helper()
	if err := exec.Command("git", "commit", "-m", message).Run(); err != nil {
		h.t.Fatalf("failed to commit: %v", err)
	}
}

// DeleteFile deletes a file
func (h *TestHelper) DeleteFile(name string) {
	h.t.Helper()
	path := filepath.Join(h.repoDir, name)
	if err := os.Remove(path); err != nil {
		h.t.Fatalf("failed to delete file %s: %v", name, err)
	}
}

// ModifyFile modifies a file with new content
func (h *TestHelper) ModifyFile(name, content string) {
	h.t.Helper()
	h.CreateFile(name, content)
}

// TestGetDiff tests the GetDiff function
func TestGetDiff(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("empty diff when no staged changes", func(t *testing.T) {
		diff, err := GetDiff()
		if err != nil {
			t.Fatalf("GetDiff failed: %v", err)
		}
		if diff != "" {
			t.Errorf("expected empty diff, got: %s", diff)
		}
	})

	t.Run("returns diff for staged new file", func(t *testing.T) {
		h.CreateFile("test.txt", "hello world\n")
		h.StageFile("test.txt")

		diff, err := GetDiff()
		if err != nil {
			t.Fatalf("GetDiff failed: %v", err)
		}
		if diff == "" {
			t.Error("expected non-empty diff for staged file")
		}
		if !contains(diff, "test.txt") {
			t.Errorf("diff should contain filename, got: %s", diff)
		}
		if !contains(diff, "hello world") {
			t.Errorf("diff should contain file content, got: %s", diff)
		}
	})

	t.Run("returns diff for staged modification", func(t *testing.T) {
		// First commit the file
		h.Commit("initial commit")

		// Modify the file
		h.ModifyFile("test.txt", "hello world modified\n")
		h.StageFile("test.txt")

		diff, err := GetDiff()
		if err != nil {
			t.Fatalf("GetDiff failed: %v", err)
		}
		if diff == "" {
			t.Error("expected non-empty diff for staged modification")
		}
		if !contains(diff, "modified") {
			t.Errorf("diff should contain modification, got: %s", diff)
		}
	})
}

// TestHasStagedChanges tests the HasStagedChanges function
func TestHasStagedChanges(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("returns false when no staged changes", func(t *testing.T) {
		hasChanges, err := HasStagedChanges()
		if err != nil {
			t.Fatalf("HasStagedChanges failed: %v", err)
		}
		if hasChanges {
			t.Error("expected no staged changes")
		}
	})

	t.Run("returns true when file is staged", func(t *testing.T) {
		h.CreateFile("test.txt", "content")
		h.StageFile("test.txt")

		hasChanges, err := HasStagedChanges()
		if err != nil {
			t.Fatalf("HasStagedChanges failed: %v", err)
		}
		if !hasChanges {
			t.Error("expected staged changes to be detected")
		}
	})

	t.Run("returns false after commit", func(t *testing.T) {
		h.Commit("test commit")

		hasChanges, err := HasStagedChanges()
		if err != nil {
			t.Fatalf("HasStagedChanges failed: %v", err)
		}
		if hasChanges {
			t.Error("expected no staged changes after commit")
		}
	})
}

// TestGetUnstagedFiles tests the GetUnstagedFiles function
func TestGetUnstagedFiles(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Create and commit an initial file
	h.CreateFile("initial.txt", "initial content")
	h.StageFile("initial.txt")
	h.Commit("initial commit")

	t.Run("returns empty when no unstaged changes", func(t *testing.T) {
		files, err := GetUnstagedFiles()
		if err != nil {
			t.Fatalf("GetUnstagedFiles failed: %v", err)
		}
		if len(files) != 0 {
			t.Errorf("expected no unstaged files, got: %v", files)
		}
	})

	t.Run("detects modified files", func(t *testing.T) {
		h.ModifyFile("initial.txt", "modified content")

		files, err := GetUnstagedFiles()
		if err != nil {
			t.Fatalf("GetUnstagedFiles failed: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 unstaged file, got: %d", len(files))
		}
		if files[0].Path != "initial.txt" {
			t.Errorf("expected initial.txt, got: %s", files[0].Path)
		}
		if files[0].Status != "modified" {
			t.Errorf("expected status 'modified', got: %s", files[0].Status)
		}
	})

	t.Run("detects deleted files", func(t *testing.T) {
		// Reset the modification first
		h.StageFile("initial.txt")
		h.Commit("update")

		h.DeleteFile("initial.txt")

		files, err := GetUnstagedFiles()
		if err != nil {
			t.Fatalf("GetUnstagedFiles failed: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 unstaged file, got: %d", len(files))
		}
		if files[0].Status != "deleted" {
			t.Errorf("expected status 'deleted', got: %s", files[0].Status)
		}
	})
}

// TestGetUntrackedFiles tests the GetUntrackedFiles function
func TestGetUntrackedFiles(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("returns empty when no untracked files", func(t *testing.T) {
		files, err := GetUntrackedFiles()
		if err != nil {
			t.Fatalf("GetUntrackedFiles failed: %v", err)
		}
		if len(files) != 0 {
			t.Errorf("expected no untracked files, got: %v", files)
		}
	})

	t.Run("detects untracked files", func(t *testing.T) {
		h.CreateFile("untracked.txt", "untracked content")

		files, err := GetUntrackedFiles()
		if err != nil {
			t.Fatalf("GetUntrackedFiles failed: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 untracked file, got: %d", len(files))
		}
		if files[0].Path != "untracked.txt" {
			t.Errorf("expected untracked.txt, got: %s", files[0].Path)
		}
		if files[0].Status != "untracked" {
			t.Errorf("expected status 'untracked', got: %s", files[0].Status)
		}
	})

	t.Run("detects multiple untracked files", func(t *testing.T) {
		h.CreateFile("another.txt", "another content")

		files, err := GetUntrackedFiles()
		if err != nil {
			t.Fatalf("GetUntrackedFiles failed: %v", err)
		}
		if len(files) != 2 {
			t.Errorf("expected 2 untracked files, got: %d", len(files))
		}
	})

	t.Run("excludes staged files", func(t *testing.T) {
		h.StageFile("untracked.txt")

		files, err := GetUntrackedFiles()
		if err != nil {
			t.Fatalf("GetUntrackedFiles failed: %v", err)
		}
		// Should only have another.txt now
		for _, f := range files {
			if f.Path == "untracked.txt" {
				t.Error("staged file should not appear in untracked files")
			}
		}
	})
}

// TestGetStagedFiles tests the GetStagedFiles function
func TestGetStagedFiles(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("returns empty when no staged files", func(t *testing.T) {
		files, err := GetStagedFiles()
		if err != nil {
			t.Fatalf("GetStagedFiles failed: %v", err)
		}
		if len(files) != 0 {
			t.Errorf("expected no staged files, got: %v", files)
		}
	})

	t.Run("detects staged new file as added", func(t *testing.T) {
		h.CreateFile("new.txt", "new content")
		h.StageFile("new.txt")

		files, err := GetStagedFiles()
		if err != nil {
			t.Fatalf("GetStagedFiles failed: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 staged file, got: %d", len(files))
		}
		if files[0].Path != "new.txt" {
			t.Errorf("expected new.txt, got: %s", files[0].Path)
		}
		if files[0].Status != "added" {
			t.Errorf("expected status 'added', got: %s", files[0].Status)
		}
	})

	t.Run("detects staged modification", func(t *testing.T) {
		h.Commit("initial")
		h.ModifyFile("new.txt", "modified content")
		h.StageFile("new.txt")

		files, err := GetStagedFiles()
		if err != nil {
			t.Fatalf("GetStagedFiles failed: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 staged file, got: %d", len(files))
		}
		if files[0].Status != "modified" {
			t.Errorf("expected status 'modified', got: %s", files[0].Status)
		}
	})

	t.Run("detects staged deletion", func(t *testing.T) {
		h.Commit("update")
		h.DeleteFile("new.txt")
		exec.Command("git", "add", "new.txt").Run()

		files, err := GetStagedFiles()
		if err != nil {
			t.Fatalf("GetStagedFiles failed: %v", err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 staged file, got: %d", len(files))
		}
		if files[0].Status != "deleted" {
			t.Errorf("expected status 'deleted', got: %s", files[0].Status)
		}
	})
}

// TestStageFiles tests the StageFiles function
func TestStageFiles(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("returns error for empty file list", func(t *testing.T) {
		err := StageFiles([]string{})
		if err == nil {
			t.Error("expected error for empty file list")
		}
	})

	t.Run("stages single file", func(t *testing.T) {
		h.CreateFile("stage-me.txt", "content")

		err := StageFiles([]string{"stage-me.txt"})
		if err != nil {
			t.Fatalf("StageFiles failed: %v", err)
		}

		files, _ := GetStagedFiles()
		if len(files) != 1 {
			t.Errorf("expected 1 staged file, got: %d", len(files))
		}
	})

	t.Run("stages multiple files", func(t *testing.T) {
		h.Commit("initial")
		h.CreateFile("file1.txt", "content1")
		h.CreateFile("file2.txt", "content2")

		err := StageFiles([]string{"file1.txt", "file2.txt"})
		if err != nil {
			t.Fatalf("StageFiles failed: %v", err)
		}

		files, _ := GetStagedFiles()
		if len(files) != 2 {
			t.Errorf("expected 2 staged files, got: %d", len(files))
		}
	})
}

// TestUnstageFiles tests the UnstageFiles function
func TestUnstageFiles(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	// Create initial commit so we can test unstaging
	h.CreateFile("initial.txt", "initial")
	h.StageFile("initial.txt")
	h.Commit("initial")

	t.Run("returns nil for empty file list", func(t *testing.T) {
		err := UnstageFiles([]string{})
		if err != nil {
			t.Errorf("expected nil for empty file list, got: %v", err)
		}
	})

	t.Run("unstages single file", func(t *testing.T) {
		h.CreateFile("unstage-me.txt", "content")
		h.StageFile("unstage-me.txt")

		// Verify it's staged
		files, _ := GetStagedFiles()
		if len(files) != 1 {
			t.Fatalf("expected file to be staged first")
		}

		err := UnstageFiles([]string{"unstage-me.txt"})
		if err != nil {
			t.Fatalf("UnstageFiles failed: %v", err)
		}

		files, _ = GetStagedFiles()
		if len(files) != 0 {
			t.Errorf("expected 0 staged files after unstage, got: %d", len(files))
		}
	})

	t.Run("unstages multiple files", func(t *testing.T) {
		h.CreateFile("file1.txt", "content1")
		h.CreateFile("file2.txt", "content2")
		h.StageFile("file1.txt")
		h.StageFile("file2.txt")

		err := UnstageFiles([]string{"file1.txt", "file2.txt"})
		if err != nil {
			t.Fatalf("UnstageFiles failed: %v", err)
		}

		files, _ := GetStagedFiles()
		if len(files) != 0 {
			t.Errorf("expected 0 staged files, got: %d", len(files))
		}
	})
}

// TestFileStatus tests the FileStatus struct
func TestFileStatus(t *testing.T) {
	fs := FileStatus{
		Path:   "test/file.go",
		Status: "modified",
	}

	if fs.Path != "test/file.go" {
		t.Errorf("expected path 'test/file.go', got: %s", fs.Path)
	}
	if fs.Status != "modified" {
		t.Errorf("expected status 'modified', got: %s", fs.Status)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
