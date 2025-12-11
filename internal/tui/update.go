package tui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.State != StateInputKey {
				return m, tea.Quit
			}
		case "enter":
			switch m.State {
			case StateInputKey:
				key := strings.TrimSpace(m.TextInput.Value())
				if key == "" {
					return m, nil
				}
				m.APIKey = key
				m.State = StateGenerating
				return m, tea.Batch(
					SaveAPIKey(key),
					ReadGitDiff(),
				)
			case StateShowResult:
				return m.handleMenuSelection()
			case StateError:
				return m, tea.Quit
			case StateSuccess:
				return m, tea.Quit
			}
		case "up", "k":
			if m.State == StateShowResult {
				m.SelectedItem--
				if m.SelectedItem < 0 {
					m.SelectedItem = len(m.MenuItems) - 1
				}
			}
		case "down", "j":
			if m.State == StateShowResult {
				m.SelectedItem++
				if m.SelectedItem >= len(m.MenuItems) {
					m.SelectedItem = 0
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
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case APIKeyLoadedMsg:
		if msg.Key != "" {
			m.APIKey = msg.Key
			m.State = StateGenerating
			return m, ReadGitDiff()
		}
		m.State = StateInputKey
		return m, textinput.Blink

	case APIKeySavedMsg:
		// Key saved successfully, continue with generation
		return m, nil

	case DiffReadyMsg:
		m.Diff = msg.Diff
		return m, GenerateCommit(m.APIKey, m.Diff)

	case CommitGeneratedMsg:
		m.CommitMsg = msg.Commit
		m.State = StateShowResult
		return m, nil

	case CommitExecutedMsg:
		m.State = StateSuccess
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
			m.State = StateSuccess
		}
		return m, nil
	case 1: // Execute commit
		return m, ExecuteCommit(m.CommitMsg)
	case 2: // Regenerate
		m.State = StateGenerating
		m.SelectedItem = 0
		return m, GenerateCommit(m.APIKey, m.Diff)
	case 3: // Quit
		return m, tea.Quit
	}
	return m, nil
}
