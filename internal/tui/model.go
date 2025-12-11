package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mounis-bhat/commit-gen/internal/config"
	"github.com/mounis-bhat/commit-gen/internal/git"
)

// AppState represents the current state of the application.
type AppState int

const (
	StateCheckingConfig AppState = iota
	StateSelectProvider
	StateSelectModel
	StateInputKey
	StateSelectFiles
	StateStaging
	StateGenerating
	StateShowResult
	StatePushing
	StateError
	StateSuccess
)

// Message types for Bubble Tea
type (
	ConfigLoadedMsg    struct{ Config *config.Config }
	APIKeySavedMsg     struct{}
	DiffReadyMsg       struct{ Diff string }
	CommitGeneratedMsg struct{ Commit string }
	ErrorMsg           struct{ Err error }
	CommitExecutedMsg  struct{}
	ModelsFetchedMsg   struct{ Models []string }
	FilesReadyMsg      struct {
		Unstaged  []git.FileStatus
		Untracked []git.FileStatus
	}
	FilesStagedMsg  struct{}
	PushExecutedMsg struct{}
)

// Model for Bubble Tea
type Model struct {
	State          AppState
	Spinner        spinner.Model
	TextInput      textinput.Model
	Provider       string
	APIKey         string
	OllamaModel    string
	OllamaModels   []string
	Diff           string
	CommitMsg      string
	Err            error
	Width          int
	Height         int
	SelectedItem   int
	MenuItems      []string
	UnstagedFiles  []git.FileStatus
	UntrackedFiles []git.FileStatus
	SelectedFiles  map[int]bool // map of file index to selected state
	SuccessAction  string       // tracks what action succeeded for success message
}

// NewModel creates and returns an initialized Model.
func NewModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF79C6"))

	ti := textinput.New()
	ti.Placeholder = "Enter your Gemini API key..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	return Model{
		State:         StateCheckingConfig,
		Spinner:       s,
		TextInput:     ti,
		MenuItems:     []string{"Copy to clipboard", "Execute commit", "Execute commit and push", "Regenerate", "Quit"},
		SelectedItem:  0,
		SelectedFiles: make(map[int]bool),
	}
}

// GetAllFiles returns all files (unstaged + untracked) as a combined slice.
func (m Model) GetAllFiles() []git.FileStatus {
	all := make([]git.FileStatus, 0, len(m.UnstagedFiles)+len(m.UntrackedFiles))
	all = append(all, m.UnstagedFiles...)
	all = append(all, m.UntrackedFiles...)
	return all
}

// GetSelectedFilePaths returns the paths of selected files.
func (m Model) GetSelectedFilePaths() []string {
	allFiles := m.GetAllFiles()
	var paths []string
	for idx, selected := range m.SelectedFiles {
		if selected && idx < len(allFiles) {
			paths = append(paths, allFiles[idx].Path)
		}
	}
	return paths
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Tick,
		LoadConfig(),
	)
}
