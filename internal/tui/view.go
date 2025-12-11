package tui

import (
	"fmt"
	"strings"
)

// View implements tea.Model.
func (m Model) View() string {
	var s strings.Builder

	// Title
	s.WriteString(TitleStyle.Render("  Commit Generator  "))
	s.WriteString("\n")
	s.WriteString(SubtitleStyle.Render("AI-powered git commit messages"))
	s.WriteString("\n\n")

	switch m.State {
	case StateCheckingConfig:
		s.WriteString(m.Spinner.View())
		s.WriteString(InfoStyle.Render(" Checking for saved configuration..."))

	case StateSelectProvider:
		s.WriteString(InfoStyle.Render("Select AI Provider:"))
		s.WriteString("\n\n")
		providers := []string{"Ollama (local, free)", "Gemini (cloud, API key)"}
		for i, provider := range providers {
			prefix := "  "
			if i == m.SelectedItem {
				prefix = SelectedMenuItemStyle.Render("→ ")
			}
			providerText := provider
			// Show current/saved provider
			if (i == 0 && m.Provider == "ollama") || (i == 1 && m.Provider == "gemini") {
				providerText += " (current)"
			}
			s.WriteString(prefix + providerText)
			s.WriteString("\n")
		}
		s.WriteString("\n")
		s.WriteString(HelpStyle.Render("↑/↓ to navigate • Enter to select"))

	case StateSelectModel:
		s.WriteString(InfoStyle.Render("Select Ollama Model:"))
		s.WriteString("\n\n")
		if len(m.OllamaModels) == 0 {
			s.WriteString(m.Spinner.View())
			s.WriteString(InfoStyle.Render(" Fetching available models..."))
		} else {
			for i, model := range m.OllamaModels {
				if i == m.SelectedItem {
					s.WriteString(SelectedMenuItemStyle.Render("→ " + model))
				} else {
					s.WriteString("  " + model)
				}
				s.WriteString("\n")
			}
		}
		s.WriteString("\n")
		s.WriteString(HelpStyle.Render("↑/↓ to navigate • Enter to select"))

	case StateInputKey:
		s.WriteString(WarningStyle.Render("No API key found!"))
		s.WriteString("\n\n")
		s.WriteString(InfoStyle.Render("Get your free API key from: "))
		s.WriteString(LinkStyle.Render("https://aistudio.google.com/apikey"))
		s.WriteString("\n\n")
		s.WriteString(m.TextInput.View())
		s.WriteString("\n")
		s.WriteString(HelpStyle.Render("Your API key will be securely stored for future use"))
		s.WriteString("\n")
		s.WriteString(HelpStyle.Render("Press Enter to continue"))

	case StateSelectFiles:
		s.WriteString(InfoStyle.Render("Select files to include in commit:"))
		s.WriteString("\n\n")

		allFiles := m.GetAllFiles()
		fileIdx := 0

		// Show staged files
		if len(m.StagedFiles) > 0 {
			s.WriteString(SuccessStyle.Render("Staged:"))
			s.WriteString("\n")
			for _, file := range m.StagedFiles {
				checkbox := "[ ]"
				if m.SelectedFiles[fileIdx] {
					checkbox = "[x]"
				}
				cursor := "  "
				if fileIdx == m.SelectedItem {
					cursor = SelectedMenuItemStyle.Render("→ ")
				}
				statusLabel := ""
				switch file.Status {
				case "deleted":
					statusLabel = ErrorStyle.Render(" (deleted)")
				case "renamed":
					statusLabel = WarningStyle.Render(" (renamed)")
				case "added":
					statusLabel = SuccessStyle.Render(" (new)")
				}
				s.WriteString(fmt.Sprintf("%s%s %s%s\n", cursor, checkbox, file.Path, statusLabel))
				fileIdx++
			}
			s.WriteString("\n")
		}

		// Show modified files (unstaged)
		if len(m.UnstagedFiles) > 0 {
			s.WriteString(WarningStyle.Render("Modified (unstaged):"))
			s.WriteString("\n")
			for _, file := range m.UnstagedFiles {
				checkbox := "[ ]"
				if m.SelectedFiles[fileIdx] {
					checkbox = "[x]"
				}
				cursor := "  "
				if fileIdx == m.SelectedItem {
					cursor = SelectedMenuItemStyle.Render("→ ")
				}
				statusLabel := ""
				if file.Status == "deleted" {
					statusLabel = ErrorStyle.Render(" (deleted)")
				} else if file.Status == "renamed" {
					statusLabel = WarningStyle.Render(" (renamed)")
				}
				s.WriteString(fmt.Sprintf("%s%s %s%s\n", cursor, checkbox, file.Path, statusLabel))
				fileIdx++
			}
			s.WriteString("\n")
		}

		// Show untracked files
		if len(m.UntrackedFiles) > 0 {
			s.WriteString(HelpStyle.Render("Untracked:"))
			s.WriteString("\n")
			for _, file := range m.UntrackedFiles {
				checkbox := "[ ]"
				if m.SelectedFiles[fileIdx] {
					checkbox = "[x]"
				}
				cursor := "  "
				if fileIdx == m.SelectedItem {
					cursor = SelectedMenuItemStyle.Render("→ ")
				}
				s.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checkbox, file.Path))
				fileIdx++
			}
			s.WriteString("\n")
		}

		// Count selected files
		selectedCount := 0
		for i := 0; i < len(allFiles); i++ {
			if m.SelectedFiles[i] {
				selectedCount++
			}
		}
		s.WriteString(InfoStyle.Render(fmt.Sprintf("Selected: %d/%d files", selectedCount, len(allFiles))))
		s.WriteString("\n\n")
		s.WriteString(HelpStyle.Render("↑/↓ navigate • Space toggle • a toggle all • Enter continue • q quit"))

	case StateStaging:
		s.WriteString(m.Spinner.View())
		s.WriteString(InfoStyle.Render(" Staging selected files..."))

	case StateGenerating:
		s.WriteString(m.Spinner.View())
		s.WriteString(InfoStyle.Render(" Analyzing your changes and generating commit message..."))
		s.WriteString("\n")
		s.WriteString(HelpStyle.Render("This may take a few seconds"))

	case StatePushing:
		s.WriteString(m.Spinner.View())
		s.WriteString(InfoStyle.Render(" Pushing to remote repository..."))

	case StateShowResult:
		s.WriteString(SuccessStyle.Render("Commit message generated!"))
		s.WriteString("\n")

		// Show the commit message in a styled box
		s.WriteString(CodeBlockStyle.Render(m.CommitMsg))
		s.WriteString("\n\n")

		// Menu
		s.WriteString(InfoStyle.Render("What would you like to do?"))
		s.WriteString("\n\n")

		for i, item := range m.MenuItems {
			cursor := "  "
			style := MenuItemStyle
			if i == m.SelectedItem {
				cursor = "▸ "
				style = SelectedMenuItemStyle
			}
			s.WriteString(fmt.Sprintf("%s%s %s\n", cursor, WarningStyle.Render(fmt.Sprintf("[%d]", i+1)), style.Render(item)))
		}

		s.WriteString("\n")
		s.WriteString(HelpStyle.Render("Use ↑/↓ or j/k to navigate • Enter to select • q to quit"))

	case StateError:
		s.WriteString(ErrorStyle.Render("Error occurred!"))
		s.WriteString("\n\n")
		s.WriteString(BoxStyle.Render(m.Err.Error()))
		s.WriteString("\n\n")
		s.WriteString(HelpStyle.Render("Press Enter or q to exit"))

	case StateSuccess:
		s.WriteString(SuccessStyle.Render("Success!"))
		s.WriteString("\n\n")
		switch m.SuccessAction {
		case "copy":
			s.WriteString(InfoStyle.Render("Commit message copied to clipboard"))
		case "commit":
			s.WriteString(InfoStyle.Render("Commit executed successfully"))
		case "commit_and_push":
			s.WriteString(InfoStyle.Render("Commit executed and pushed to remote successfully"))
		default:
			s.WriteString(InfoStyle.Render("Operation completed successfully"))
		}
		s.WriteString("\n\n")
		s.WriteString(HelpStyle.Render("Press Enter or q to exit"))
	}

	return s.String()
}
