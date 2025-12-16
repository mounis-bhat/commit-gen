package tui

import (
	"testing"

	"github.com/mounis-bhat/commit-gen/internal/git"
)

// TestNewModel tests the NewModel constructor
func TestNewModel(t *testing.T) {
	model := NewModel()

	t.Run("initial state", func(t *testing.T) {
		if model.State != StateCheckingConfig {
			t.Errorf("expected initial state StateCheckingConfig, got %v", model.State)
		}
	})

	t.Run("menu items", func(t *testing.T) {
		expectedItems := []string{
			"Copy to clipboard",
			"Execute commit",
			"Execute commit and push",
			"Regenerate",
			"Quit",
		}
		if len(model.MenuItems) != len(expectedItems) {
			t.Errorf("expected %d menu items, got %d", len(expectedItems), len(model.MenuItems))
		}
		for i, expected := range expectedItems {
			if model.MenuItems[i] != expected {
				t.Errorf("expected menu item[%d] '%s', got '%s'", i, expected, model.MenuItems[i])
			}
		}
	})

	t.Run("selected item", func(t *testing.T) {
		if model.SelectedItem != 0 {
			t.Errorf("expected SelectedItem 0, got %d", model.SelectedItem)
		}
	})

	t.Run("dimensions", func(t *testing.T) {
		if model.Width != defaultWidth {
			t.Errorf("expected Width %d, got %d", defaultWidth, model.Width)
		}
		if model.Height != defaultHeight {
			t.Errorf("expected Height %d, got %d", defaultHeight, model.Height)
		}
	})

	t.Run("ready flag", func(t *testing.T) {
		if model.Ready != false {
			t.Error("expected Ready to be false")
		}
	})
}

// TestGetSteps tests the GetSteps method
func TestGetSteps(t *testing.T) {
	tests := []struct {
		name     string
		state    AppState
		expected []Step
	}{
		{
			name:  "checking config",
			state: StateCheckingConfig,
			expected: []Step{
				{Name: "Provider", Completed: false, Active: false},
				{Name: "Files", Completed: false, Active: false},
				{Name: "Generate", Completed: false, Active: false},
				{Name: "Action", Completed: false, Active: false},
			},
		},
		{
			name:  "select provider",
			state: StateSelectProvider,
			expected: []Step{
				{Name: "Provider", Completed: false, Active: true},
				{Name: "Files", Completed: false, Active: false},
				{Name: "Generate", Completed: false, Active: false},
				{Name: "Action", Completed: false, Active: false},
			},
		},
		{
			name:  "select files",
			state: StateSelectFiles,
			expected: []Step{
				{Name: "Provider", Completed: true, Active: false},
				{Name: "Files", Completed: false, Active: true},
				{Name: "Generate", Completed: false, Active: false},
				{Name: "Action", Completed: false, Active: false},
			},
		},
		{
			name:  "generating",
			state: StateGenerating,
			expected: []Step{
				{Name: "Provider", Completed: true, Active: false},
				{Name: "Files", Completed: true, Active: false},
				{Name: "Generate", Completed: false, Active: true},
				{Name: "Action", Completed: false, Active: false},
			},
		},
		{
			name:  "show result",
			state: StateShowResult,
			expected: []Step{
				{Name: "Provider", Completed: true, Active: false},
				{Name: "Files", Completed: true, Active: false},
				{Name: "Generate", Completed: true, Active: false},
				{Name: "Action", Completed: false, Active: true},
			},
		},
		{
			name:  "success",
			state: StateSuccess,
			expected: []Step{
				{Name: "Provider", Completed: true, Active: false},
				{Name: "Files", Completed: true, Active: false},
				{Name: "Generate", Completed: true, Active: false},
				{Name: "Action", Completed: true, Active: false},
			},
		},
		{
			name:  "error with provider set",
			state: StateError,
			expected: []Step{
				{Name: "Provider", Completed: true, Active: false},
				{Name: "Files", Completed: false, Active: false},
				{Name: "Generate", Completed: false, Active: false},
				{Name: "Action", Completed: false, Active: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			model.State = tt.state
			if tt.state == StateError {
				model.Provider = "gemini" // Set provider for error case
			}

			steps := model.GetSteps()

			if len(steps) != len(tt.expected) {
				t.Fatalf("expected %d steps, got %d", len(tt.expected), len(steps))
			}

			for i, expected := range tt.expected {
				if steps[i].Name != expected.Name {
					t.Errorf("step[%d] name: expected '%s', got '%s'", i, expected.Name, steps[i].Name)
				}
				if steps[i].Completed != expected.Completed {
					t.Errorf("step[%d] completed: expected %v, got %v", i, expected.Completed, steps[i].Completed)
				}
				if steps[i].Active != expected.Active {
					t.Errorf("step[%d] active: expected %v, got %v", i, expected.Active, steps[i].Active)
				}
			}
		})
	}
}

// TestGetCurrentStepNumber tests the GetCurrentStepNumber method
func TestGetCurrentStepNumber(t *testing.T) {
	tests := []struct {
		state    AppState
		expected int
	}{
		{StateCheckingConfig, 0},
		{StateSelectProvider, 1},
		{StateSelectModel, 1},
		{StateInputKey, 1},
		{StateSelectFiles, 2},
		{StateStaging, 2},
		{StateGenerating, 3},
		{StateShowResult, 4},
		{StatePushing, 4},
		{StateSuccess, 4},
		{StateError, 4},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			model := NewModel()
			model.State = tt.state

			result := model.GetCurrentStepNumber()
			if result != tt.expected {
				t.Errorf("state %v: expected step number %d, got %d", tt.state, tt.expected, result)
			}
		})
	}
}

// TestGetAllFiles tests the GetAllFiles method
func TestGetAllFiles(t *testing.T) {
	model := NewModel()

	// Set up test data
	model.StagedFiles = []git.FileStatus{
		{Path: "staged1.txt", Status: "modified"},
		{Path: "staged2.txt", Status: "added"},
	}
	model.UnstagedFiles = []git.FileStatus{
		{Path: "unstaged1.txt", Status: "modified"},
		{Path: "unstaged2.txt", Status: "deleted"},
	}
	model.UntrackedFiles = []git.FileStatus{
		{Path: "untracked1.txt", Status: "untracked"},
	}

	allFiles := model.GetAllFiles()

	expected := []git.FileStatus{
		{Path: "staged1.txt", Status: "modified"},
		{Path: "staged2.txt", Status: "added"},
		{Path: "unstaged1.txt", Status: "modified"},
		{Path: "unstaged2.txt", Status: "deleted"},
		{Path: "untracked1.txt", Status: "untracked"},
	}

	if len(allFiles) != len(expected) {
		t.Fatalf("expected %d files, got %d", len(expected), len(allFiles))
	}

	for i, expectedFile := range expected {
		if allFiles[i].Path != expectedFile.Path {
			t.Errorf("file[%d] path: expected '%s', got '%s'", i, expectedFile.Path, allFiles[i].Path)
		}
		if allFiles[i].Status != expectedFile.Status {
			t.Errorf("file[%d] status: expected '%s', got '%s'", i, expectedFile.Status, allFiles[i].Status)
		}
	}
}

// TestGetStagedFileCount tests the GetStagedFileCount method
func TestGetStagedFileCount(t *testing.T) {
	model := NewModel()

	t.Run("empty staged files", func(t *testing.T) {
		count := model.GetStagedFileCount()
		if count != 0 {
			t.Errorf("expected count 0, got %d", count)
		}
	})

	t.Run("with staged files", func(t *testing.T) {
		model.StagedFiles = []git.FileStatus{
			{Path: "file1.txt", Status: "modified"},
			{Path: "file2.txt", Status: "added"},
			{Path: "file3.txt", Status: "deleted"},
		}

		count := model.GetStagedFileCount()
		if count != 3 {
			t.Errorf("expected count 3, got %d", count)
		}
	})
}

// TestGetSelectedFilePaths tests the GetSelectedFilePaths method
func TestGetSelectedFilePaths(t *testing.T) {
	model := NewModel()

	// Set up test data
	model.StagedFiles = []git.FileStatus{
		{Path: "staged1.txt", Status: "modified"},
		{Path: "staged2.txt", Status: "added"},
	}
	model.UnstagedFiles = []git.FileStatus{
		{Path: "unstaged1.txt", Status: "modified"},
		{Path: "unstaged2.txt", Status: "deleted"},
	}
	model.UntrackedFiles = []git.FileStatus{
		{Path: "untracked1.txt", Status: "untracked"},
	}

	t.Run("no selections", func(t *testing.T) {
		paths := model.GetSelectedFilePaths()
		if len(paths) != 0 {
			t.Errorf("expected 0 selected paths, got %d", len(paths))
		}
	})

	t.Run("select some files", func(t *testing.T) {
		// Select staged1.txt (index 0), unstaged2.txt (index 3), untracked1.txt (index 4)
		model.SelectedFiles[0] = true
		model.SelectedFiles[3] = true
		model.SelectedFiles[4] = true

		paths := model.GetSelectedFilePaths()
		expected := []string{"staged1.txt", "unstaged2.txt", "untracked1.txt"}

		if len(paths) != len(expected) {
			t.Fatalf("expected %d paths, got %d", len(expected), len(paths))
		}

		for i, expectedPath := range expected {
			if paths[i] != expectedPath {
				t.Errorf("path[%d]: expected '%s', got '%s'", i, expectedPath, paths[i])
			}
		}
	})
}

// TestGetFilesToStage tests the GetFilesToStage method
func TestGetFilesToStage(t *testing.T) {
	model := NewModel()

	// Set up test data
	model.StagedFiles = []git.FileStatus{
		{Path: "staged1.txt", Status: "modified"},
		{Path: "staged2.txt", Status: "added"},
	}
	model.UnstagedFiles = []git.FileStatus{
		{Path: "unstaged1.txt", Status: "modified"},
		{Path: "unstaged2.txt", Status: "deleted"},
	}
	model.UntrackedFiles = []git.FileStatus{
		{Path: "untracked1.txt", Status: "untracked"},
	}

	t.Run("no selections", func(t *testing.T) {
		paths := model.GetFilesToStage()
		if len(paths) != 0 {
			t.Errorf("expected 0 files to stage, got %d", len(paths))
		}
	})

	t.Run("select unstaged and untracked files", func(t *testing.T) {
		// Select unstaged1.txt (index 2), untracked1.txt (index 4)
		model.SelectedFiles[2] = true
		model.SelectedFiles[4] = true

		paths := model.GetFilesToStage()
		expected := []string{"unstaged1.txt", "untracked1.txt"}

		if len(paths) != len(expected) {
			t.Fatalf("expected %d paths, got %d", len(expected), len(paths))
		}

		for i, expectedPath := range expected {
			if paths[i] != expectedPath {
				t.Errorf("path[%d]: expected '%s', got '%s'", i, expectedPath, paths[i])
			}
		}
	})

	t.Run("select staged file (should not be included)", func(t *testing.T) {
		// Select staged1.txt (index 0) - should not be included in files to stage
		model.SelectedFiles[0] = true

		paths := model.GetFilesToStage()
		// Should not include staged files
		for _, path := range paths {
			if path == "staged1.txt" {
				t.Error("staged file should not be included in files to stage")
			}
		}
	})
}

// TestGetFilesToUnstage tests the GetFilesToUnstage method
func TestGetFilesToUnstage(t *testing.T) {
	model := NewModel()

	// Set up test data
	model.StagedFiles = []git.FileStatus{
		{Path: "staged1.txt", Status: "modified"},
		{Path: "staged2.txt", Status: "added"},
		{Path: "staged3.txt", Status: "deleted"},
	}

	t.Run("no selections (all staged files should be unstaged)", func(t *testing.T) {
		paths := model.GetFilesToUnstage()
		expected := []string{"staged1.txt", "staged2.txt", "staged3.txt"}

		if len(paths) != len(expected) {
			t.Fatalf("expected %d paths, got %d", len(expected), len(paths))
		}

		for i, expectedPath := range expected {
			if paths[i] != expectedPath {
				t.Errorf("path[%d]: expected '%s', got '%s'", i, expectedPath, paths[i])
			}
		}
	})

	t.Run("select some staged files (only unselected should be unstaged)", func(t *testing.T) {
		// Select staged1.txt (index 0) and staged3.txt (index 2)
		model.SelectedFiles[0] = true
		model.SelectedFiles[2] = true

		paths := model.GetFilesToUnstage()
		expected := []string{"staged2.txt"} // Only the unselected one

		if len(paths) != len(expected) {
			t.Fatalf("expected %d paths, got %d", len(expected), len(paths))
		}

		if paths[0] != expected[0] {
			t.Errorf("expected '%s', got '%s'", expected[0], paths[0])
		}
	})
}

// TestGetContentWidth tests the GetContentWidth method
func TestGetContentWidth(t *testing.T) {
	tests := []struct {
		width    int
		expected int
	}{
		{minWidth - 1, minWidth - 6}, // Below min, should use min
		{minWidth, minWidth - 6},
		{80, 80 - 6},
		{120, 120 - 6},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			model := NewModel()
			model.Width = tt.width

			result := model.GetContentWidth()
			if result != tt.expected {
				t.Errorf("width %d: expected %d, got %d", tt.width, tt.expected, result)
			}
		})
	}
}

// TestGetContentHeight tests the GetContentHeight method
func TestGetContentHeight(t *testing.T) {
	tests := []struct {
		height   int
		expected int
	}{
		{minHeight - 1, minHeight - 10}, // Below min, should use min
		{minHeight, minHeight - 10},
		{24, 24 - 10},
		{40, 40 - 10},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			model := NewModel()
			model.Height = tt.height

			result := model.GetContentHeight()
			if result != tt.expected {
				t.Errorf("height %d: expected %d, got %d", tt.height, tt.expected, result)
			}
		})
	}
}

// TestAppState tests the AppState constants
func TestAppState(t *testing.T) {
	// Test that all states have unique values
	states := []AppState{
		StateCheckingConfig,
		StateSelectProvider,
		StateSelectModel,
		StateInputKey,
		StateSelectFiles,
		StateStaging,
		StateGenerating,
		StateShowResult,
		StatePushing,
		StateError,
		StateSuccess,
	}

	seen := make(map[AppState]bool)
	for _, state := range states {
		if seen[state] {
			t.Errorf("duplicate state value: %v", state)
		}
		seen[state] = true
	}
}

// TestStep tests the Step struct
func TestStep(t *testing.T) {
	step := Step{
		Name:      "Test Step",
		Completed: true,
		Active:    false,
	}

	if step.Name != "Test Step" {
		t.Errorf("expected Name 'Test Step', got %s", step.Name)
	}
	if !step.Completed {
		t.Error("expected Completed true")
	}
	if step.Active {
		t.Error("expected Active false")
	}
}

// TestParseCommitMessages tests the parseCommitMessages function
func TestParseCommitMessages(t *testing.T) {
	t.Run("single -m flag with double quotes", func(t *testing.T) {
		cmd := `git commit -m "feat(auth): Add login"`
		messages, err := parseCommitMessages(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(messages))
		}
		if messages[0] != "feat(auth): Add login" {
			t.Errorf("expected 'feat(auth): Add login', got '%s'", messages[0])
		}
	})

	t.Run("multiple -m flags inline", func(t *testing.T) {
		cmd := `git commit -m "feat(auth): :sparkles: Add login" -m "• Implement JWT" -m "• Add endpoint"`
		messages, err := parseCommitMessages(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(messages) != 3 {
			t.Fatalf("expected 3 messages, got %d", len(messages))
		}
		expected := []string{
			"feat(auth): :sparkles: Add login",
			"• Implement JWT",
			"• Add endpoint",
		}
		for i, exp := range expected {
			if messages[i] != exp {
				t.Errorf("message[%d]: expected '%s', got '%s'", i, exp, messages[i])
			}
		}
	})

	t.Run("backslash line continuation (POSIX format)", func(t *testing.T) {
		cmd := "git commit \\\n-m \"feat: Add feature\" \\\n-m \"• Detail 1\" \\\n-m \"• Detail 2\""
		messages, err := parseCommitMessages(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(messages) != 3 {
			t.Fatalf("expected 3 messages, got %d", len(messages))
		}
		expected := []string{
			"feat: Add feature",
			"• Detail 1",
			"• Detail 2",
		}
		for i, exp := range expected {
			if messages[i] != exp {
				t.Errorf("message[%d]: expected '%s', got '%s'", i, exp, messages[i])
			}
		}
	})

	t.Run("handles escaped newlines in message", func(t *testing.T) {
		cmd := `git commit -m "feat: Title\n• Detail 1\n• Detail 2"`
		messages, err := parseCommitMessages(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(messages))
		}
		expected := "feat: Title\n• Detail 1\n• Detail 2"
		if messages[0] != expected {
			t.Errorf("expected '%s', got '%s'", expected, messages[0])
		}
	})

	t.Run("handles single quotes", func(t *testing.T) {
		cmd := `git commit -m 'feat: Add feature'`
		messages, err := parseCommitMessages(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(messages))
		}
		if messages[0] != "feat: Add feature" {
			t.Errorf("expected 'feat: Add feature', got '%s'", messages[0])
		}
	})

	t.Run("empty command returns error", func(t *testing.T) {
		_, err := parseCommitMessages("")
		if err == nil {
			t.Error("expected error for empty command")
		}
	})

	t.Run("no -m flag returns error", func(t *testing.T) {
		_, err := parseCommitMessages("git commit")
		if err == nil {
			t.Error("expected error for command without -m flag")
		}
	})

	t.Run("handles escaped quotes in message", func(t *testing.T) {
		cmd := `git commit -m "feat: Add \"quoted\" text"`
		messages, err := parseCommitMessages(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(messages))
		}
		expected := `feat: Add "quoted" text`
		if messages[0] != expected {
			t.Errorf("expected '%s', got '%s'", expected, messages[0])
		}
	})

	t.Run("handles emoji codes", func(t *testing.T) {
		cmd := `git commit -m "feat(ui): :sparkles: Add new UI" -m "• :bug: Fix edge case"`
		messages, err := parseCommitMessages(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(messages) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(messages))
		}
		if messages[0] != "feat(ui): :sparkles: Add new UI" {
			t.Errorf("message[0]: expected 'feat(ui): :sparkles: Add new UI', got '%s'", messages[0])
		}
		if messages[1] != "• :bug: Fix edge case" {
			t.Errorf("message[1]: expected '• :bug: Fix edge case', got '%s'", messages[1])
		}
	})

	t.Run("handles CRLF line endings", func(t *testing.T) {
		cmd := "git commit \\\r\n-m \"feat: Title\" \\\r\n-m \"• Detail\""
		messages, err := parseCommitMessages(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(messages) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(messages))
		}
	})
}
