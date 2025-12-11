package tui

import "github.com/charmbracelet/lipgloss"

// Styles using Lip Gloss
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF79C6")).
			Background(lipgloss.Color("#282A36")).
			Padding(0, 2).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			Italic(true)

	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#50FA7B"))

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF5555"))

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB86C"))

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BD93F9"))

	CodeBlockStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#44475A")).
			Foreground(lipgloss.Color("#F8F8F2")).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	MenuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")).
			PaddingLeft(2)

	SelectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#50FA7B")).
				Bold(true).
				PaddingLeft(2)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			MarginTop(1)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#BD93F9")).
			Padding(1, 2).
			MarginTop(1)

	LinkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			Underline(true)
)
