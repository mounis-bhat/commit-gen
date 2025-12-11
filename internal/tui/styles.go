package tui

import "github.com/charmbracelet/lipgloss"

// Color palette (Dracula theme)
var (
	colorPink       = lipgloss.Color("#FF79C6")
	colorCyan       = lipgloss.Color("#8BE9FD")
	colorGreen      = lipgloss.Color("#50FA7B")
	colorRed        = lipgloss.Color("#FF5555")
	colorOrange     = lipgloss.Color("#FFB86C")
	colorPurple     = lipgloss.Color("#BD93F9")
	colorYellow     = lipgloss.Color("#F1FA8C")
	colorForeground = lipgloss.Color("#F8F8F2")
	colorComment    = lipgloss.Color("#6272A4")
	colorBackground = lipgloss.Color("#282A36")
	colorCurrentBg  = lipgloss.Color("#44475A")
)

// Container styles
var (
	// Main container with border
	ContainerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Padding(1, 2)

	// Header section
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPink).
			Background(colorBackground).
			Padding(0, 1).
			MarginBottom(1)

	// Footer/help bar
	FooterStyle = lipgloss.NewStyle().
			Foreground(colorComment).
			MarginTop(1).
			Padding(0, 1)
)

// Title styles
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPink).
			Background(colorCurrentBg).
			Padding(0, 2)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Italic(true)

	// Section title style
	SectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPurple).
				MarginTop(1).
				MarginBottom(1)
)

// Status styles
var (
	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorGreen)

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorRed)

	WarningStyle = lipgloss.NewStyle().
			Foreground(colorOrange)

	InfoStyle = lipgloss.NewStyle().
			Foreground(colorPurple)
)

// Progress indicator styles
var (
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(colorPurple)

	ProgressTextStyle = lipgloss.NewStyle().
				Foreground(colorComment).
				Italic(true)

	StepActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCyan)

	StepCompletedStyle = lipgloss.NewStyle().
				Foreground(colorGreen)

	StepPendingStyle = lipgloss.NewStyle().
				Foreground(colorComment)
)

// File status badge styles
var (
	BadgeNewStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBackground).
			Background(colorGreen).
			Padding(0, 1)

	BadgeModifiedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorBackground).
				Background(colorOrange).
				Padding(0, 1)

	BadgeDeletedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorBackground).
				Background(colorRed).
				Padding(0, 1)

	BadgeRenamedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorBackground).
				Background(colorCyan).
				Padding(0, 1)

	BadgeStagedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorBackground).
				Background(colorPurple).
				Padding(0, 1)

	BadgeUntrackedStyle = lipgloss.NewStyle().
				Foreground(colorComment).
				Italic(true)
)

// Code and content blocks
var (
	CodeBlockStyle = lipgloss.NewStyle().
			Background(colorCurrentBg).
			Foreground(colorForeground).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	CommitBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGreen).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	ErrorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorRed).
			Padding(1, 2).
			MarginTop(1)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Padding(1, 2).
			MarginTop(1)
)

// Menu styles
var (
	MenuItemStyle = lipgloss.NewStyle().
			Foreground(colorForeground).
			PaddingLeft(2)

	SelectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(colorGreen).
				Bold(true)

	MenuNumberStyle = lipgloss.NewStyle().
			Foreground(colorOrange).
			Bold(true)

	// Card-style menu item
	MenuCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorComment).
			Padding(0, 2).
			MarginRight(1)

	MenuCardSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorGreen).
				Foreground(colorGreen).
				Bold(true).
				Padding(0, 2).
				MarginRight(1)
)

// File list styles
var (
	FileListHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPurple).
				MarginBottom(1)

	FileItemStyle = lipgloss.NewStyle().
			Foreground(colorForeground)

	FileItemSelectedStyle = lipgloss.NewStyle().
				Foreground(colorGreen).
				Bold(true)

	CheckboxStyle = lipgloss.NewStyle().
			Foreground(colorComment)

	CheckboxCheckedStyle = lipgloss.NewStyle().
				Foreground(colorGreen).
				Bold(true)

	FileCountStyle = lipgloss.NewStyle().
			Foreground(colorComment).
			Italic(true)
)

// Input styles
var (
	InputLabelStyle = lipgloss.NewStyle().
			Foreground(colorPurple).
			Bold(true)

	InputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorPurple).
			Padding(0, 1)
)

// Help and link styles
var (
	HelpStyle = lipgloss.NewStyle().
			Foreground(colorComment).
			MarginTop(1)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(colorComment)

	HelpSepStyle = lipgloss.NewStyle().
			Foreground(colorComment)

	LinkStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Underline(true)
)

// Provider selection styles
var (
	ProviderCardStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorComment).
				Padding(0, 2).
				MarginBottom(1)

	ProviderCardSelectedStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(colorGreen).
					Foreground(colorGreen).
					Bold(true).
					Padding(0, 2).
					MarginBottom(1)

	ProviderDescStyle = lipgloss.NewStyle().
				Foreground(colorComment).
				Italic(true)

	ProviderCurrentStyle = lipgloss.NewStyle().
				Foreground(colorCyan).
				Italic(true)
)

// Icon characters
const (
	IconCheck     = "✓"
	IconCross     = "✗"
	IconArrow     = "→"
	IconBullet    = "•"
	IconSpinner   = "◐"
	IconFolder    = "📁"
	IconFile      = "📄"
	IconGit       = "🔀"
	IconSuccess   = "✅"
	IconError     = "❌"
	IconWarning   = "⚠️"
	IconInfo      = "ℹ️"
	IconKey       = "🔑"
	IconRocket    = "🚀"
	IconClipboard = "📋"
	IconCommit    = "💾"
	IconPush      = "⬆️"
	IconReload    = "🔄"
)

// Helper functions for dynamic styling
func GetContainerStyle(width int) lipgloss.Style {
	return ContainerStyle.Width(width - 4) // Account for border
}

func GetCodeBlockStyle(width int) lipgloss.Style {
	return CodeBlockStyle.Width(width - 8) // Account for padding and border
}

func GetCommitBoxStyle(width int) lipgloss.Style {
	return CommitBoxStyle.Width(width - 8)
}

func GetErrorBoxStyle(width int) lipgloss.Style {
	return ErrorBoxStyle.Width(width - 8)
}

// TruncatePath truncates a file path to fit within maxWidth
func TruncatePath(path string, maxWidth int) string {
	if len(path) <= maxWidth {
		return path
	}
	if maxWidth <= 3 {
		return "..."
	}
	return "..." + path[len(path)-(maxWidth-3):]
}
