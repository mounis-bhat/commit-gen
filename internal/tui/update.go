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
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.State != StateInputKey {
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
						// API key already set, proceed to file selection
						return m, ReadAvailableFiles()
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
				// Proceed to file selection instead of directly generating
				return m, tea.Batch(
					SaveConfig(cfg),
					ReadAvailableFiles(),
				)
			case StateInputKey:
				key := strings.TrimSpace(m.TextInput.Value())
				if key == "" {
					return m, nil
				}
				m.APIKey = key
				// Proceed to file selection instead of directly generating
				return m, ReadAvailableFiles()
			case StateSelectFiles:
				// Check if any files are selected
				selectedPaths := m.GetSelectedFilePaths()
				if len(selectedPaths) == 0 {
					return m, nil // No files selected
				}

				// Get files that need to be staged and unstaged
				toStage := m.GetFilesToStage()
				toUnstage := m.GetFilesToUnstage()

				// If no changes needed (all staged files selected, no new files to stage)
				if len(toStage) == 0 && len(toUnstage) == 0 {
					// Proceed directly to generating
					m.State = StateGenerating
					return m, ReadGitDiff()
				}

				m.State = StateStaging
				return m, ApplyStagingChanges(toStage, toUnstage)
			case StateShowResult:
				return m.handleMenuSelection()
			case StateError:
				return m, tea.Quit
			case StateSuccess:
				return m, tea.Quit
			}
		case "up", "k":
			switch m.State {
			case StateSelectProvider:
				m.SelectedItem--
				if m.SelectedItem < 0 {
					m.SelectedItem = 1
				}
			case StateSelectModel:
				if len(m.OllamaModels) > 0 {
					m.SelectedItem--
					if m.SelectedItem < 0 {
						m.SelectedItem = len(m.OllamaModels) - 1
					}
				}
			case StateShowResult:
				m.SelectedItem--
				if m.SelectedItem < 0 {
					m.SelectedItem = len(m.MenuItems) - 1
				}
			case StateSelectFiles:
				allFiles := m.GetAllFiles()
				if len(allFiles) > 0 {
					m.SelectedItem--
					if m.SelectedItem < 0 {
						m.SelectedItem = len(allFiles) - 1
					}
				}
			}
		case "down", "j":
			switch m.State {
			case StateSelectProvider:
				m.SelectedItem++
				if m.SelectedItem > 1 {
					m.SelectedItem = 0
				}
			case StateSelectModel:
				if len(m.OllamaModels) > 0 {
					m.SelectedItem++
					if m.SelectedItem >= len(m.OllamaModels) {
						m.SelectedItem = 0
					}
				}
			case StateShowResult:
				m.SelectedItem++
				if m.SelectedItem >= len(m.MenuItems) {
					m.SelectedItem = 0
				}
			case StateSelectFiles:
				allFiles := m.GetAllFiles()
				if len(allFiles) > 0 {
					m.SelectedItem++
					if m.SelectedItem >= len(allFiles) {
						m.SelectedItem = 0
					}
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
				for i := range allFiles {
					if !m.SelectedFiles[i] {
						allSelected = false
						break
					}
				}
				// Toggle all: if all selected, deselect all; otherwise select all
				for i := range allFiles {
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

		// Update viewport dimensions
		contentHeight := m.GetContentHeight()
		m.Viewport.Width = m.GetContentWidth()
		m.Viewport.Height = contentHeight

		// Update text input width
		inputWidth := m.GetContentWidth() - 10
		if inputWidth < 20 {
			inputWidth = 20
		}
		m.TextInput.Width = inputWidth

		if !m.Ready {
			m.Ready = true
		}

	case ConfigLoadedMsg:
		cfg := msg.Config
		// Always show provider selection, but load saved config for convenience
		m.APIKey = cfg.APIKey
		m.OllamaModel = cfg.OllamaModel
		if cfg.Provider != "" {
			m.Provider = cfg.Provider
			// Pre-select the saved provider
			switch cfg.Provider {
			case "ollama":
				m.SelectedItem = 0
			case "gemini":
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
		// Check if diff is empty (shouldn't happen as we validate in file selection)
		if strings.TrimSpace(msg.Diff) == "" {
			return m, func() tea.Msg {
				return ErrorMsg{Err: fmt.Errorf("no staged changes found")}
			}
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
		m.StagedFiles = msg.Staged
		m.UnstagedFiles = msg.Unstaged
		m.UntrackedFiles = msg.Untracked
		// Check if there are any files at all
		if len(msg.Staged) == 0 && len(msg.Unstaged) == 0 && len(msg.Untracked) == 0 {
			return m, func() tea.Msg {
				return ErrorMsg{Err: fmt.Errorf("no changes detected. Nothing to commit")}
			}
		}
		m.State = StateSelectFiles
		m.SelectedItem = 0
		m.SelectedFiles = make(map[int]bool)
		// Pre-select all staged files
		for i := 0; i < len(msg.Staged); i++ {
			m.SelectedFiles[i] = true
		}
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
