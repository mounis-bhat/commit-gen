package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mounis-bhat/commit-gen/internal/config"
)

// AppState represents the current state of the application.
type AppState int

const (
	StateCheckingConfig AppState = iota
	StateSelectProvider
	StateSelectModel
	StateInputKey
	StateGenerating
	StateShowResult
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
)

// Model for Bubble Tea
type Model struct {
	State        AppState
	Spinner      spinner.Model
	TextInput    textinput.Model
	Provider     string
	APIKey       string
	OllamaModel  string
	Diff         string
	CommitMsg    string
	Err          error
	Width        int
	Height       int
	SelectedItem int
	MenuItems    []string
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
		State:        StateCheckingConfig,
		Spinner:      s,
		TextInput:    ti,
		MenuItems:    []string{"Copy to clipboard", "Execute commit", "Regenerate", "Quit"},
		SelectedItem: 0,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Tick,
		LoadConfig(),
	)
}
