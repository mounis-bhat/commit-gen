package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mounis-bhat/commit-gen/internal/git"
)

// View implements tea.Model.
func (m Model) View() string {
	var content strings.Builder

	// Build the main content based on state
	switch m.State {
	case StateCheckingConfig:
		content.WriteString(m.viewCheckingConfig())
	case StateSelectProvider:
		content.WriteString(m.viewSelectProvider())
	case StateSelectModel:
		content.WriteString(m.viewSelectModel())
	case StateInputKey:
		content.WriteString(m.viewInputKey())
	case StateSelectFiles:
		content.WriteString(m.viewSelectFiles())
	case StateStaging:
		content.WriteString(m.viewStaging())
	case StateGenerating:
		content.WriteString(m.viewGenerating())
	case StateShowResult:
		content.WriteString(m.viewShowResult())
	case StatePushing:
		content.WriteString(m.viewPushing())
	case StateError:
		content.WriteString(m.viewError())
	case StateSuccess:
		content.WriteString(m.viewSuccess())
	}

	// Build the full layout
	return m.buildLayout(content.String())
}

// buildLayout creates the full UI layout with header, content, and footer
func (m Model) buildLayout(content string) string {
	var s strings.Builder
	width := m.GetContentWidth()

	// Header
	s.WriteString(m.renderHeader(width))
	s.WriteString("\n")

	// Progress bar (only show after config check)
	if m.State != StateCheckingConfig {
		s.WriteString(m.renderProgressBar(width))
		s.WriteString("\n\n")
	}

	// Main content
	s.WriteString(content)

	// Wrap in container
	containerStyle := GetContainerStyle(m.Width)
	return containerStyle.Render(s.String())
}

// renderHeader renders the application header
func (m Model) renderHeader(_ int) string {
	title := TitleStyle.Render(fmt.Sprintf(" %s Commit Generator ", IconGit))
	subtitle := SubtitleStyle.Render("AI-powered git commit messages")

	return lipgloss.JoinVertical(lipgloss.Left, title, subtitle)
}

// renderProgressBar renders the step progress indicator
func (m Model) renderProgressBar(width int) string {
	steps := m.GetSteps()
	var parts []string

	for i, step := range steps {
		var stepStr string
		if step.Completed {
			stepStr = StepCompletedStyle.Render(fmt.Sprintf("%s %s", IconCheck, step.Name))
		} else if step.Active {
			stepStr = StepActiveStyle.Render(fmt.Sprintf("%s %s", IconArrow, step.Name))
		} else {
			stepStr = StepPendingStyle.Render(fmt.Sprintf("%s %s", IconBullet, step.Name))
		}
		parts = append(parts, stepStr)

		// Add connector between steps
		if i < len(steps)-1 {
			if step.Completed {
				parts = append(parts, StepCompletedStyle.Render(" ─── "))
			} else {
				parts = append(parts, StepPendingStyle.Render(" ─── "))
			}
		}
	}

	progressLine := strings.Join(parts, "")

	// Add step counter
	currentStep := m.GetCurrentStepNumber()
	if currentStep > 0 {
		stepCounter := ProgressTextStyle.Render(fmt.Sprintf("Step %d of %d", currentStep, len(steps)))
		return lipgloss.JoinVertical(lipgloss.Left, progressLine, stepCounter)
	}

	return progressLine
}

// renderHelp renders the help/footer section
func (m Model) renderHelp(keys [][]string) string {
	var parts []string
	for _, kv := range keys {
		key := HelpKeyStyle.Render(kv[0])
		desc := HelpDescStyle.Render(kv[1])
		parts = append(parts, fmt.Sprintf("%s %s", key, desc))
	}
	return HelpStyle.Render(strings.Join(parts, HelpSepStyle.Render(" • ")))
}

// viewCheckingConfig renders the config checking state
func (m Model) viewCheckingConfig() string {
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(m.Spinner.View())
	s.WriteString(InfoStyle.Render(" Checking for saved configuration..."))
	s.WriteString("\n")
	return s.String()
}

// viewSelectProvider renders the provider selection state
func (m Model) viewSelectProvider() string {
	var s strings.Builder
	width := m.GetContentWidth()

	s.WriteString(SectionTitleStyle.Render("Select AI Provider"))
	s.WriteString("\n\n")

	providers := []struct {
		name    string
		desc    string
		key     string
		current bool
	}{
		{
			name:    "Ollama",
			desc:    "Local, free, private",
			key:     "ollama",
			current: m.Provider == "ollama",
		},
		{
			name:    "Gemini",
			desc:    "Cloud, API key required",
			key:     "gemini",
			current: m.Provider == "gemini",
		},
		{
			name:    "Claude",
			desc:    "Cloud, API key required",
			key:     "claude",
			current: m.Provider == "claude",
		},
		{
			name:    "OpenAI",
			desc:    "Cloud, API key required",
			key:     "openai",
			current: m.Provider == "openai",
		},
	}

	for i, provider := range providers {
		var cardStyle lipgloss.Style
		cursor := "  "
		if i == m.SelectedItem {
			cardStyle = ProviderCardSelectedStyle.Width(width - 4)
			cursor = SelectedMenuItemStyle.Render(IconArrow + " ")
		} else {
			cardStyle = ProviderCardStyle.Width(width - 4)
		}

		providerName := provider.name
		if provider.current {
			providerName += " " + ProviderCurrentStyle.Render("(current)")
		}

		cardContent := fmt.Sprintf("%s\n%s", providerName, ProviderDescStyle.Render(provider.desc))
		s.WriteString(cursor + cardStyle.Render(cardContent))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(m.renderHelp([][]string{
		{"↑/↓", "navigate"},
		{"Enter", "select"},
		{"q", "quit"},
	}))

	return s.String()
}

// viewSelectModel renders the model selection state for Ollama or OpenAI
func (m Model) viewSelectModel() string {
	var s strings.Builder

	var title, loadingHint string
	var models []string
	if m.Provider == "openai" {
		title = "Select OpenAI Model"
		loadingHint = "Validating API key and fetching models..."
		models = m.OpenAIModels
	} else {
		title = "Select Ollama Model"
		loadingHint = "Make sure Ollama is running on localhost:11434"
		models = m.OllamaModels
	}

	s.WriteString(SectionTitleStyle.Render(title))
	s.WriteString("\n\n")

	if len(models) == 0 {
		s.WriteString(m.Spinner.View())
		s.WriteString(InfoStyle.Render(" Fetching available models..."))
		s.WriteString("\n\n")
		s.WriteString(HelpStyle.Render(loadingHint))
	} else {
		// Show models in a scrollable list
		maxVisible := m.GetContentHeight() - 4
		if maxVisible < 5 {
			maxVisible = 5
		}

		start := 0
		if m.SelectedItem >= maxVisible {
			start = m.SelectedItem - maxVisible + 1
		}
		end := start + maxVisible
		if end > len(models) {
			end = len(models)
		}

		// Show scroll indicator if needed
		if start > 0 {
			s.WriteString(HelpStyle.Render("  ↑ more models above"))
			s.WriteString("\n")
		}

		for i := start; i < end; i++ {
			model := models[i]
			if i == m.SelectedItem {
				s.WriteString(SelectedMenuItemStyle.Render(fmt.Sprintf("%s %s", IconArrow, model)))
			} else {
				s.WriteString(fmt.Sprintf("  %s", model))
			}
			s.WriteString("\n")
		}

		if end < len(models) {
			s.WriteString(HelpStyle.Render("  ↓ more models below"))
			s.WriteString("\n")
		}

		s.WriteString("\n")
		s.WriteString(FileCountStyle.Render(fmt.Sprintf("%d models available", len(models))))
	}

	s.WriteString("\n\n")
	s.WriteString(m.renderHelp([][]string{
		{"↑/↓", "navigate"},
		{"Enter", "select"},
		{"q", "quit"},
	}))

	return s.String()
}
// viewInputKey renders the API key input state
func (m Model) viewInputKey() string {
	var s strings.Builder
	width := m.GetContentWidth()

	s.WriteString(SectionTitleStyle.Render(fmt.Sprintf("%s Enter API Key", IconKey)))
	s.WriteString("\n\n")

	// Warning box
	warningBox := BoxStyle.Width(width - 4).BorderForeground(colorOrange)
	var apiKeyLink string
	if m.Provider == "claude" {
		apiKeyLink = "https://console.anthropic.com/settings/keys"
	} else if m.Provider == "openai" {
		apiKeyLink = "https://platform.openai.com/api-keys"
	} else {
		apiKeyLink = "https://aistudio.google.com/apikey"
	}
	warningContent := fmt.Sprintf("%s No API key found!\n\nGet your API key from:\n%s",
		IconWarning,
		LinkStyle.Render(apiKeyLink))
	s.WriteString(warningBox.Render(warningContent))
	s.WriteString("\n\n")

	// Input field
	s.WriteString(InputLabelStyle.Render("API Key:"))
	s.WriteString("\n")
	s.WriteString(m.TextInput.View())
	s.WriteString("\n\n")

	s.WriteString(HelpStyle.Render("Your API key will be securely stored for future use"))
	s.WriteString("\n\n")
	s.WriteString(m.renderHelp([][]string{
		{"Enter", "continue"},
		{"Ctrl+C", "quit"},
	}))

	return s.String()
}

// viewSelectFiles renders the file selection state
func (m Model) viewSelectFiles() string {
	var s strings.Builder
	width := m.GetContentWidth()

	s.WriteString(SectionTitleStyle.Render("Select Files to Commit"))
	s.WriteString("\n\n")

	allFiles := m.GetAllFiles()
	fileIdx := 0
	maxPathWidth := width - 20 // Account for checkbox, cursor, badges

	// Calculate visible range for scrolling
	maxVisible := m.GetContentHeight() - 8
	if maxVisible < 5 {
		maxVisible = 5
	}

	// Show staged files section
	if len(m.StagedFiles) > 0 {
		header := fmt.Sprintf("%s Staged (%d)", IconCheck, len(m.StagedFiles))
		s.WriteString(SuccessStyle.Render(header))
		s.WriteString("\n")

		for _, file := range m.StagedFiles {
			s.WriteString(m.renderFileItem(fileIdx, file, maxPathWidth, true))
			s.WriteString("\n")
			fileIdx++
		}
		s.WriteString("\n")
	}

	// Show modified files (unstaged)
	if len(m.UnstagedFiles) > 0 {
		header := fmt.Sprintf("%s Modified (%d)", IconWarning, len(m.UnstagedFiles))
		s.WriteString(WarningStyle.Render(header))
		s.WriteString("\n")

		for _, file := range m.UnstagedFiles {
			s.WriteString(m.renderFileItem(fileIdx, file, maxPathWidth, false))
			s.WriteString("\n")
			fileIdx++
		}
		s.WriteString("\n")
	}

	// Show untracked files
	if len(m.UntrackedFiles) > 0 {
		header := fmt.Sprintf("? Untracked (%d)", len(m.UntrackedFiles))
		s.WriteString(HelpStyle.Render(header))
		s.WriteString("\n")

		for _, file := range m.UntrackedFiles {
			s.WriteString(m.renderFileItem(fileIdx, file, maxPathWidth, false))
			s.WriteString("\n")
			fileIdx++
		}
		s.WriteString("\n")
	}

	// Selection summary
	selectedCount := 0
	for i := range allFiles {
		if m.SelectedFiles[i] {
			selectedCount++
		}
	}

	summaryStyle := InfoStyle.Bold(true)
	s.WriteString(summaryStyle.Render(fmt.Sprintf("Selected: %d/%d files", selectedCount, len(allFiles))))
	s.WriteString("\n\n")

	s.WriteString(m.renderHelp([][]string{
		{"↑/↓", "navigate"},
		{"Space", "toggle"},
		{"a", "toggle all"},
		{"Enter", "continue"},
		{"q", "quit"},
	}))

	return s.String()
}

// renderFileItem renders a single file item with checkbox and status badge
func (m Model) renderFileItem(idx int, file FileStatus, maxWidth int, _ bool) string {
	var s strings.Builder

	// Cursor
	if idx == m.SelectedItem {
		s.WriteString(SelectedMenuItemStyle.Render(IconArrow + " "))
	} else {
		s.WriteString("  ")
	}

	// Checkbox
	if m.SelectedFiles[idx] {
		s.WriteString(CheckboxCheckedStyle.Render("[" + IconCheck + "]"))
	} else {
		s.WriteString(CheckboxStyle.Render("[ ]"))
	}
	s.WriteString(" ")

	// File path (truncated if needed)
	path := TruncatePath(file.Path, maxWidth)
	if idx == m.SelectedItem {
		s.WriteString(FileItemSelectedStyle.Render(path))
	} else {
		s.WriteString(FileItemStyle.Render(path))
	}

	// Status badge
	s.WriteString(" ")
	s.WriteString(m.renderStatusBadge(file.Status))

	return s.String()
}

// renderStatusBadge renders a colored badge for file status
func (m Model) renderStatusBadge(status string) string {
	switch status {
	case "added":
		return BadgeNewStyle.Render("NEW")
	case "modified":
		return BadgeModifiedStyle.Render("MOD")
	case "deleted":
		return BadgeDeletedStyle.Render("DEL")
	case "renamed":
		return BadgeRenamedStyle.Render("REN")
	case "untracked":
		return BadgeUntrackedStyle.Render("untracked")
	default:
		return ""
	}
}

// viewStaging renders the staging state
func (m Model) viewStaging() string {
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(m.Spinner.View())
	s.WriteString(InfoStyle.Render(" Staging selected files..."))
	s.WriteString("\n")
	return s.String()
}

// viewGenerating renders the generating state
func (m Model) viewGenerating() string {
	var s strings.Builder
	width := m.GetContentWidth()

	s.WriteString("\n")

	// Animated loading box
	loadingBox := BoxStyle.Width(width - 4)
	loadingContent := fmt.Sprintf("%s %s\n\n%s",
		m.Spinner.View(),
		InfoStyle.Render("Analyzing your changes..."),
		HelpStyle.Render("Generating commit message with AI"))

	s.WriteString(loadingBox.Render(loadingContent))
	s.WriteString("\n\n")

	s.WriteString(ProgressTextStyle.Render("This may take a few seconds depending on your provider"))
	s.WriteString("\n")

	return s.String()
}

// viewShowResult renders the result state with commit message and actions
func (m Model) viewShowResult() string {
	var s strings.Builder
	width := m.GetContentWidth()

	s.WriteString(SuccessStyle.Render(fmt.Sprintf("%s Commit Message Generated!", IconSuccess)))
	s.WriteString("\n\n")

	// Commit message in a styled box
	commitBox := GetCommitBoxStyle(width)
	formattedMsg := m.formatCommitMessage()
	s.WriteString(commitBox.Render(formattedMsg))
	s.WriteString("\n\n")

	// Action menu
	s.WriteString(SectionTitleStyle.Render("Choose an Action"))
	s.WriteString("\n\n")

	menuIcons := []string{IconClipboard, IconCommit, IconPush, IconReload, IconCross}
	menuDescs := []string{
		"Copy message to clipboard",
		"Create commit with this message",
		"Commit and push to remote",
		"Generate a new message",
		"Exit without action",
	}

	for i, item := range m.MenuItems {
		cursor := "  "
		var itemStyle lipgloss.Style
		if i == m.SelectedItem {
			cursor = SelectedMenuItemStyle.Render(IconArrow + " ")
			itemStyle = SelectedMenuItemStyle
		} else {
			itemStyle = MenuItemStyle
		}

		number := MenuNumberStyle.Render(fmt.Sprintf("[%d]", i+1))
		icon := menuIcons[i]
		name := itemStyle.Render(item)
		desc := HelpStyle.Render(menuDescs[i])

		s.WriteString(fmt.Sprintf("%s%s %s %s\n", cursor, number, icon, name))
		s.WriteString(fmt.Sprintf("      %s\n", desc))
	}

	s.WriteString("\n")
	s.WriteString(m.renderHelp([][]string{
		{"↑/↓", "navigate"},
		{"1-5", "quick select"},
		{"Enter", "confirm"},
		{"q", "quit"},
	}))

	return s.String()
}

// viewPushing renders the pushing state
func (m Model) viewPushing() string {
	var s strings.Builder
	width := m.GetContentWidth()

	s.WriteString("\n")

	loadingBox := BoxStyle.Width(width - 4)
	loadingContent := fmt.Sprintf("%s %s\n\n%s",
		m.Spinner.View(),
		InfoStyle.Render("Pushing to remote repository..."),
		HelpStyle.Render("Please wait"))

	s.WriteString(loadingBox.Render(loadingContent))
	s.WriteString("\n")

	return s.String()
}

// viewError renders the error state
func (m Model) viewError() string {
	var s strings.Builder
	width := m.GetContentWidth()

	s.WriteString(ErrorStyle.Render(fmt.Sprintf("%s Error Occurred", IconError)))
	s.WriteString("\n\n")

	// Error box
	errorBox := GetErrorBoxStyle(width)
	errorContent := m.Err.Error()
	s.WriteString(errorBox.Render(errorContent))
	s.WriteString("\n\n")

	s.WriteString(m.renderHelp([][]string{
		{"Enter", "exit"},
		{"q", "exit"},
	}))

	return s.String()
}

// viewSuccess renders the success state
func (m Model) viewSuccess() string {
	var s strings.Builder
	width := m.GetContentWidth()

	s.WriteString(SuccessStyle.Render(fmt.Sprintf("%s Success!", IconSuccess)))
	s.WriteString("\n\n")

	// Success message box
	successBox := BoxStyle.Width(width - 4).BorderForeground(colorGreen)
	var message string
	switch m.SuccessAction {
	case "copy":
		message = fmt.Sprintf("%s Commit message copied to clipboard\n\nYou can now paste it anywhere.", IconClipboard)
	case "commit":
		message = fmt.Sprintf("%s Commit created successfully\n\nYour changes have been committed.", IconCommit)
	case "commit_and_push":
		message = fmt.Sprintf("%s Commit created and pushed!\n\nYour changes are now on the remote repository.", IconRocket)
	default:
		message = "Operation completed successfully"
	}
	s.WriteString(successBox.Render(message))
	s.WriteString("\n\n")

	s.WriteString(m.renderHelp([][]string{
		{"Enter", "exit"},
		{"q", "exit"},
	}))

	return s.String()
}

// FileStatus is imported from git package, but we need a local reference for the view
type FileStatus = git.FileStatus

// formatCommitMessage formats the commit command for display
func (m Model) formatCommitMessage() string {
	return strings.TrimSpace(m.CommitMsg)
}
