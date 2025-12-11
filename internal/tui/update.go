package tui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mounis-bhat/commit-gen/internal/ai"
	"github.com/mounis-bhat/commit-gen/internal/config"
)

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.State != StateInputKey && m.State != StateSelectFiles {
				return m, tea.Quit
			}
			if m.State == StateSelectFiles {
				return m, tea.Quit
			}
		case "enter":
			switch m.State {
			case StateSelectProvider:
				if m.SelectedItem == 0 {
					// Ollama selected
					m.Provider = "ollama"
					m.State = StateSelectModel
					m.SelectedItem = 0
					if len(m.OllamaModels) == 0 {
						return m, FetchOllamaModels("http://localhost:11434")
					}
				} else {
					// Gemini selected
					m.Provider = "gemini"
					if m.APIKey != "" {
						// API key already set, proceed to generating
						m.State = StateGenerating
						return m, tea.Batch(
							ReadGitDiff(),
						)
					} else {
						m.State = StateInputKey
					}
				}
				return m, nil
			case StateSelectModel:
				if len(m.OllamaModels) == 0 || m.SelectedItem >= len(m.OllamaModels) {
					return m, nil // Wait for models to load
				}
				model := m.OllamaModels[m.SelectedItem]
				cfg := &config.Config{
					Provider:    "ollama",
					OllamaURL:   "http://localhost:11434",
					OllamaModel: model,
				}
				m.OllamaModel = model
				m.State = StateGenerating
				return m, tea.Batch(
					SaveConfig(cfg),
					ReadGitDiff(),
				)
			case StateInputKey:
				key := strings.TrimSpace(m.TextInput.Value())
				if key == "" {
					return m, nil
				}
				m.APIKey = key
				m.State = StateGenerating
				return m, ReadGitDiff()
			case StateSelectFiles:
				// Stage selected files and proceed
				selectedPaths := m.GetSelectedFilePaths()
				if len(selectedPaths) == 0 {
					return m, nil // No files selected
				}
				m.State = StateStaging
				return m, StageSelectedFiles(selectedPaths)
			case StateShowResult:
				return m.handleMenuSelection()
			case StateError:
				return m, tea.Quit
			case StateSuccess:
				return m, tea.Quit
			}
		case "up", "k":
			if m.State == StateSelectProvider {
				m.SelectedItem--
				if m.SelectedItem < 0 {
					m.SelectedItem = 1
				}
			} else if m.State == StateSelectModel {
				m.SelectedItem--
				if m.SelectedItem < 0 {
					m.SelectedItem = len(m.OllamaModels) - 1
				}
			} else if m.State == StateShowResult {
				m.SelectedItem--
				if m.SelectedItem < 0 {
					m.SelectedItem = len(m.MenuItems) - 1
				}
			} else if m.State == StateSelectFiles {
				allFiles := m.GetAllFiles()
				m.SelectedItem--
				if m.SelectedItem < 0 {
					m.SelectedItem = len(allFiles) - 1
				}
			}
		case "down", "j":
			if m.State == StateSelectProvider {
				m.SelectedItem++
				if m.SelectedItem > 1 {
					m.SelectedItem = 0
				}
			} else if m.State == StateSelectModel {
				m.SelectedItem++
				if m.SelectedItem >= len(m.OllamaModels) {
					m.SelectedItem = 0
				}
			} else if m.State == StateShowResult {
				m.SelectedItem++
				if m.SelectedItem >= len(m.MenuItems) {
					m.SelectedItem = 0
				}
			} else if m.State == StateSelectFiles {
				allFiles := m.GetAllFiles()
				m.SelectedItem++
				if m.SelectedItem >= len(allFiles) {
					m.SelectedItem = 0
				}
			}
		case " ": // Space to toggle file selection
			if m.State == StateSelectFiles {
				m.SelectedFiles[m.SelectedItem] = !m.SelectedFiles[m.SelectedItem]
			}
		case "a": // Select all files
			if m.State == StateSelectFiles {
				allFiles := m.GetAllFiles()
				allSelected := true
				for i := 0; i < len(allFiles); i++ {
					if !m.SelectedFiles[i] {
						allSelected = false
						break
					}
				}
				// Toggle all: if all selected, deselect all; otherwise select all
				for i := 0; i < len(allFiles); i++ {
					m.SelectedFiles[i] = !allSelected
				}
			}
		case "1":
			if m.State == StateShowResult {
				m.SelectedItem = 0
				return m.handleMenuSelection()
			}
		case "2":
			if m.State == StateShowResult {
				m.SelectedItem = 1
				return m.handleMenuSelection()
			}
		case "3":
			if m.State == StateShowResult {
				m.SelectedItem = 2
				return m.handleMenuSelection()
			}
		case "4":
			if m.State == StateShowResult {
				m.SelectedItem = 3
				return m.handleMenuSelection()
			}
		case "5":
			if m.State == StateShowResult {
				m.SelectedItem = 4
				return m.handleMenuSelection()
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case ConfigLoadedMsg:
		cfg := msg.Config
		// Always show provider selection, but load saved config for convenience
		m.APIKey = cfg.APIKey
		m.OllamaModel = cfg.OllamaModel
		if cfg.Provider != "" {
			m.Provider = cfg.Provider
			// Pre-select the saved provider
			if cfg.Provider == "ollama" {
				m.SelectedItem = 0
			} else if cfg.Provider == "gemini" {
				m.SelectedItem = 1
			}
		}
		m.State = StateSelectProvider
		return m, nil

	case APIKeySavedMsg:
		// Key saved successfully, continue with generation
		return m, nil

	case DiffReadyMsg:
		m.Diff = msg.Diff
		// Check if diff is empty - need to show file picker
		if strings.TrimSpace(msg.Diff) == "" {
			// No staged changes, fetch available files
			return m, ReadAvailableFiles()
		}
		provider, err := ai.NewProvider(&config.Config{
			Provider:    m.Provider,
			APIKey:      m.APIKey,
			OllamaURL:   "http://localhost:11434",
			OllamaModel: m.OllamaModel,
		})
		if err != nil {
			return m, func() tea.Msg { return ErrorMsg{Err: err} }
		}
		return m, GenerateCommit(provider, m.Diff)

	case FilesReadyMsg:
		m.UnstagedFiles = msg.Unstaged
		m.UntrackedFiles = msg.Untracked
		// Check if there are any files to stage
		if len(msg.Unstaged) == 0 && len(msg.Untracked) == 0 {
			return m, func() tea.Msg {
				return ErrorMsg{Err: fmt.Errorf("no changes detected. Nothing to commit")}
			}
		}
		m.State = StateSelectFiles
		m.SelectedItem = 0
		m.SelectedFiles = make(map[int]bool)
		return m, nil

	case FilesStagedMsg:
		// Files staged successfully, now read the diff
		m.State = StateGenerating
		return m, ReadGitDiff()

	case CommitGeneratedMsg:
		m.CommitMsg = msg.Commit
		// Save config after successful generation
		cfg := &config.Config{
			Provider:    m.Provider,
			APIKey:      m.APIKey,
			OllamaURL:   "http://localhost:11434",
			OllamaModel: m.OllamaModel,
		}
		m.State = StateShowResult
		m.SelectedItem = 0
		return m, SaveConfig(cfg)

	case CommitExecutedMsg:
		// Check if we should also push
		if m.SuccessAction == "commit_and_push" {
			m.State = StatePushing
			return m, ExecutePush()
		}
		m.SuccessAction = "commit"
		m.State = StateSuccess
		return m, nil

	case PushExecutedMsg:
		m.SuccessAction = "commit_and_push"
		m.State = StateSuccess
		return m, nil

	case ModelsFetchedMsg:
		m.OllamaModels = msg.Models
		return m, nil

	case ErrorMsg:
		m.Err = msg.Err
		m.State = StateError
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}

	// Update text input
	if m.State == StateInputKey {
		var cmd tea.Cmd
		m.TextInput, cmd = m.TextInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleMenuSelection() (tea.Model, tea.Cmd) {
	switch m.SelectedItem {
	case 0: // Copy to clipboard
		if err := clipboard.WriteAll(m.CommitMsg); err != nil {
			m.Err = fmt.Errorf("failed to copy: %w", err)
			m.State = StateError
		} else {
			m.SuccessAction = "copy"
			m.State = StateSuccess
		}
		return m, nil
	case 1: // Execute commit
		m.SuccessAction = "commit"
		return m, ExecuteCommit(m.CommitMsg)
	case 2: // Execute commit and push
		m.SuccessAction = "commit_and_push"
		return m, ExecuteCommit(m.CommitMsg)
	case 3: // Regenerate
		m.State = StateGenerating
		m.SelectedItem = 0
		provider, err := ai.NewProvider(&config.Config{
			Provider:    m.Provider,
			APIKey:      m.APIKey,
			OllamaURL:   "http://localhost:11434",
			OllamaModel: m.OllamaModel,
		})
		if err != nil {
			return m, func() tea.Msg { return ErrorMsg{Err: err} }
		}
		return m, GenerateCommit(provider, m.Diff)
	case 4: // Quit
		return m, tea.Quit
	}
	return m, nil
}
