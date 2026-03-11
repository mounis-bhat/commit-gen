package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
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

// Step represents a workflow step for progress tracking
type Step struct {
	Name      string
	Completed bool
	Active    bool
}

// GetSteps returns the workflow steps based on current state
func (m Model) GetSteps() []Step {
	steps := []Step{
		{Name: "Provider", Completed: false, Active: false},
		{Name: "Files", Completed: false, Active: false},
		{Name: "Generate", Completed: false, Active: false},
		{Name: "Action", Completed: false, Active: false},
	}

	switch m.State {
	case StateCheckingConfig:
		// No steps active yet
	case StateSelectProvider, StateSelectModel, StateInputKey:
		steps[0].Active = true
	case StateSelectFiles, StateStaging:
		steps[0].Completed = true
		steps[1].Active = true
	case StateGenerating:
		steps[0].Completed = true
		steps[1].Completed = true
		steps[2].Active = true
	case StateShowResult:
		steps[0].Completed = true
		steps[1].Completed = true
		steps[2].Completed = true
		steps[3].Active = true
	case StatePushing:
		steps[0].Completed = true
		steps[1].Completed = true
		steps[2].Completed = true
		steps[3].Active = true
	case StateSuccess:
		steps[0].Completed = true
		steps[1].Completed = true
		steps[2].Completed = true
		steps[3].Completed = true
	case StateError:
		// Mark completed steps based on where we got to
		if m.Provider != "" {
			steps[0].Completed = true
		}
		if len(m.StagedFiles) > 0 || len(m.UnstagedFiles) > 0 {
			steps[1].Completed = true
		}
		if m.CommitMsg != "" {
			steps[2].Completed = true
		}
	}

	return steps
}

// GetCurrentStepNumber returns the current step number (1-based)
func (m Model) GetCurrentStepNumber() int {
	switch m.State {
	case StateCheckingConfig:
		return 0
	case StateSelectProvider, StateSelectModel, StateInputKey:
		return 1
	case StateSelectFiles, StateStaging:
		return 2
	case StateGenerating:
		return 3
	case StateShowResult, StatePushing:
		return 4
	case StateSuccess, StateError:
		return 4
	}
	return 0
}

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
		Staged    []git.FileStatus
		Unstaged  []git.FileStatus
		Untracked []git.FileStatus
	}
	FilesStagedMsg   struct{}
	FilesUnstagedMsg struct{}
	PushExecutedMsg  struct{}
)

// Model for Bubble Tea
type Model struct {
	State          AppState
	Spinner        spinner.Model
	TextInput      textinput.Model
	Viewport       viewport.Model
	Provider       string
	GeminiAPIKey   string
	ClaudeAPIKey   string
	OpenAIAPIKey   string
	OllamaModel    string
	OllamaModels   []string
	OpenAIModel    string
	OpenAIModels   []string
	Diff           string
	CommitMsg      string
	Err            error
	Width          int
	Height         int
	SelectedItem   int
	MenuItems      []string
	StagedFiles    []git.FileStatus
	UnstagedFiles  []git.FileStatus
	UntrackedFiles []git.FileStatus
	SelectedFiles  map[int]bool // map of file index to selected state
	SuccessAction  string       // tracks what action succeeded for success message
	Ready          bool         // whether viewport is ready
}

// Default dimensions
const (
	defaultWidth  = 80
	defaultHeight = 24
	minWidth      = 40
	minHeight     = 15
)

// NewModel creates and returns an initialized Model.
func NewModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF79C6"))

	ti := textinput.New()
	ti.Placeholder = "Enter your API key..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 50
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	vp := viewport.New(defaultWidth, defaultHeight-10)
	vp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))

	return Model{
		State:         StateCheckingConfig,
		Spinner:       s,
		TextInput:     ti,
		Viewport:      vp,
		MenuItems:     []string{"Copy to clipboard", "Execute commit", "Execute commit and push", "Regenerate", "Quit"},
		SelectedItem:  0,
		SelectedFiles: make(map[int]bool),
		Width:         defaultWidth,
		Height:        defaultHeight,
		Ready:         false,
	}
}

// currentAPIKey returns the API key for the currently selected provider.
func (m Model) currentAPIKey() string {
	switch m.Provider {
	case "claude":
		return m.ClaudeAPIKey
	case "openai":
		return m.OpenAIAPIKey
	default:
		return m.GeminiAPIKey
	}
}

// setCurrentAPIKey sets the API key for the currently selected provider.
func (m *Model) setCurrentAPIKey(key string) {
	switch m.Provider {
	case "claude":
		m.ClaudeAPIKey = key
	case "openai":
		m.OpenAIAPIKey = key
	default:
		m.GeminiAPIKey = key
	}
}

// GetAllFiles returns all files (staged + unstaged + untracked) as a combined slice.
func (m Model) GetAllFiles() []git.FileStatus {
	all := make([]git.FileStatus, 0, len(m.StagedFiles)+len(m.UnstagedFiles)+len(m.UntrackedFiles))
	all = append(all, m.StagedFiles...)
	all = append(all, m.UnstagedFiles...)
	all = append(all, m.UntrackedFiles...)
	return all
}

// GetStagedFileCount returns the number of staged files.
func (m Model) GetStagedFileCount() int {
	return len(m.StagedFiles)
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

// GetFilesToStage returns paths of files that need to be staged (selected but not currently staged).
func (m Model) GetFilesToStage() []string {
	stagedCount := len(m.StagedFiles)
	allFiles := m.GetAllFiles()
	var paths []string
	for idx, selected := range m.SelectedFiles {
		// Only include files that are not already staged (index >= stagedCount)
		if selected && idx >= stagedCount && idx < len(allFiles) {
			paths = append(paths, allFiles[idx].Path)
		}
	}
	return paths
}

// GetFilesToUnstage returns paths of files that need to be unstaged (staged but not selected).
func (m Model) GetFilesToUnstage() []string {
	var paths []string
	for idx, file := range m.StagedFiles {
		// If this staged file is not selected, it should be unstaged
		if !m.SelectedFiles[idx] {
			paths = append(paths, file.Path)
		}
	}
	return paths
}

// GetContentWidth returns the usable content width
func (m Model) GetContentWidth() int {
	w := m.Width
	if w < minWidth {
		w = minWidth
	}
	return w - 6 // Account for container padding and border
}

// GetContentHeight returns the usable content height
func (m Model) GetContentHeight() int {
	h := m.Height
	if h < minHeight {
		h = minHeight
	}
	return h - 10 // Account for header, footer, progress bar
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Tick,
		LoadConfig(),
	)
}
