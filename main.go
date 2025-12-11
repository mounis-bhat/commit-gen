package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"google.golang.org/genai"
)

const systemPrompt = `You are an expert at writing clear, structured git commit messages following the Conventional Commits standard with emojis.

When given a git diff, generate a commit message in this EXACT format:

git commit -m "TYPE(SCOPE): EMOJI DESCRIPTION" \
-m "• First bullet point detail" \
-m "• Second bullet point detail" \
\
-m "TYPE(SCOPE): EMOJI DESCRIPTION" \
-m "• Detail about this commit" \
\

Rules:
1. Use Conventional Commits: feat, fix, refactor, ui, docs, test, chore, perf, style, etc.
2. Add relevant emoji (sparkles :sparkles: for features, bug :bug: for fixes, recycle :recycle: for refactors, lipstick :lipstick: for UI, etc.)
3. Keep scope concise and descriptive
4. Use bullet points (•) for implementation details
5. Group related changes together with blank lines between groups
6. Each -m creates a separate commit message line
7. Start descriptions with action verbs
8. Be specific about what changed and why
9. Do not use backticks or markdown formatting for code; use plain text

Generate ONLY the git commit command, nothing else. No explanations or markdown.`

// Config file structure
type Config struct {
	APIKey string `json:"api_key"`
}

// Application states
type appState int

const (
	stateCheckingKey appState = iota
	stateInputKey
	stateGenerating
	stateShowResult
	stateError
	stateSuccess
)

// Messages
type apiKeyLoadedMsg struct{ key string }
type apiKeySavedMsg struct{}
type diffReadyMsg struct{ diff string }
type commitGeneratedMsg struct{ commit string }
type errorMsg struct{ err error }
type commitExecutedMsg struct{}

// Model for Bubble Tea
type model struct {
	state        appState
	spinner      spinner.Model
	textInput    textinput.Model
	apiKey       string
	diff         string
	commitMsg    string
	err          error
	width        int
	height       int
	selectedItem int
	menuItems    []string
}

// Styles using Lip Gloss
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF79C6")).
			Background(lipgloss.Color("#282A36")).
			Padding(0, 2).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			Italic(true)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#50FA7B"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF5555"))

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB86C"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BD93F9"))

	codeBlockStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#44475A")).
			Foreground(lipgloss.Color("#F8F8F2")).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")).
			PaddingLeft(2)

	selectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#50FA7B")).
				Bold(true).
				PaddingLeft(2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			MarginTop(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#BD93F9")).
			Padding(1, 2).
			MarginTop(1)
)

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".commit-gen-config.json"
	}
	return filepath.Join(homeDir, ".commit-gen-config.json")
}

func loadAPIKey() tea.Cmd {
	return func() tea.Msg {
		configPath := getConfigPath()
		data, err := os.ReadFile(configPath)
		if err != nil {
			return apiKeyLoadedMsg{key: ""}
		}

		var config Config
		if err := json.Unmarshal(data, &config); err != nil {
			return apiKeyLoadedMsg{key: ""}
		}

		return apiKeyLoadedMsg{key: config.APIKey}
	}
}

func saveAPIKey(key string) tea.Cmd {
	return func() tea.Msg {
		configPath := getConfigPath()
		config := Config{APIKey: key}
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to marshal config: %w", err)}
		}

		if err := os.WriteFile(configPath, data, 0600); err != nil {
			return errorMsg{err: fmt.Errorf("failed to save config: %w", err)}
		}

		return apiKeySavedMsg{}
	}
}

func readGitDiff() tea.Cmd {
	return func() tea.Msg {
		diff, err := getGitDiff()
		if err != nil {
			return errorMsg{err: err}
		}
		if strings.TrimSpace(diff) == "" {
			return errorMsg{err: fmt.Errorf("no staged changes detected. Use 'git add' to stage changes first")}
		}
		return diffReadyMsg{diff: diff}
	}
}

func generateCommit(apiKey, diff string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		client, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  apiKey,
			Backend: genai.BackendGeminiAPI,
		})
		if err != nil {
			return errorMsg{err: fmt.Errorf("failed to create Gemini client: %w", err)}
		}

		commitMsg, err := generateCommitMessage(ctx, client, diff)
		if err != nil {
			return errorMsg{err: err}
		}

		return commitGeneratedMsg{commit: commitMsg}
	}
}

func executeCommit(commitMsg string) tea.Cmd {
	return func() tea.Msg {
		// Extract the commit command and execute it
		cmd := exec.Command("bash", "-c", commitMsg)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return errorMsg{err: fmt.Errorf("failed to execute commit: %w", err)}
		}

		return commitExecutedMsg{}
	}
}

func initialModel() model {
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

	return model{
		state:        stateCheckingKey,
		spinner:      s,
		textInput:    ti,
		menuItems:    []string{"Copy to clipboard", "Execute commit", "Regenerate", "Quit"},
		selectedItem: 0,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		loadAPIKey(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state != stateInputKey {
				return m, tea.Quit
			}
		case "enter":
			switch m.state {
			case stateInputKey:
				key := strings.TrimSpace(m.textInput.Value())
				if key == "" {
					return m, nil
				}
				m.apiKey = key
				m.state = stateGenerating
				return m, tea.Batch(
					saveAPIKey(key),
					readGitDiff(),
				)
			case stateShowResult:
				return m.handleMenuSelection()
			case stateError:
				return m, tea.Quit
			case stateSuccess:
				return m, tea.Quit
			}
		case "up", "k":
			if m.state == stateShowResult {
				m.selectedItem--
				if m.selectedItem < 0 {
					m.selectedItem = len(m.menuItems) - 1
				}
			}
		case "down", "j":
			if m.state == stateShowResult {
				m.selectedItem++
				if m.selectedItem >= len(m.menuItems) {
					m.selectedItem = 0
				}
			}
		case "1":
			if m.state == stateShowResult {
				m.selectedItem = 0
				return m.handleMenuSelection()
			}
		case "2":
			if m.state == stateShowResult {
				m.selectedItem = 1
				return m.handleMenuSelection()
			}
		case "3":
			if m.state == stateShowResult {
				m.selectedItem = 2
				return m.handleMenuSelection()
			}
		case "4":
			if m.state == stateShowResult {
				m.selectedItem = 3
				return m.handleMenuSelection()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case apiKeyLoadedMsg:
		if msg.key != "" {
			m.apiKey = msg.key
			m.state = stateGenerating
			return m, readGitDiff()
		}
		m.state = stateInputKey
		return m, textinput.Blink

	case apiKeySavedMsg:
		// Key saved successfully, continue with generation
		return m, nil

	case diffReadyMsg:
		m.diff = msg.diff
		return m, generateCommit(m.apiKey, m.diff)

	case commitGeneratedMsg:
		m.commitMsg = msg.commit
		m.state = stateShowResult
		return m, nil

	case commitExecutedMsg:
		m.state = stateSuccess
		return m, nil

	case errorMsg:
		m.err = msg.err
		m.state = stateError
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Update text input
	if m.state == stateInputKey {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) handleMenuSelection() (tea.Model, tea.Cmd) {
	switch m.selectedItem {
	case 0: // Copy to clipboard
		if err := clipboard.WriteAll(m.commitMsg); err != nil {
			m.err = fmt.Errorf("failed to copy: %w", err)
			m.state = stateError
		} else {
			m.state = stateSuccess
		}
		return m, nil
	case 1: // Execute commit
		return m, executeCommit(m.commitMsg)
	case 2: // Regenerate
		m.state = stateGenerating
		m.selectedItem = 0
		return m, generateCommit(m.apiKey, m.diff)
	case 3: // Quit
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("  Commit Generator  "))
	s.WriteString("\n")
	s.WriteString(subtitleStyle.Render("AI-powered git commit messages"))
	s.WriteString("\n\n")

	switch m.state {
	case stateCheckingKey:
		s.WriteString(m.spinner.View())
		s.WriteString(infoStyle.Render(" Checking for saved API key..."))

	case stateInputKey:
		s.WriteString(warningStyle.Render("No API key found!"))
		s.WriteString("\n\n")
		s.WriteString(infoStyle.Render("Get your free API key from: "))
		s.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			Underline(true).
			Render("https://aistudio.google.com/apikey"))
		s.WriteString("\n\n")
		s.WriteString(m.textInput.View())
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("Your API key will be securely stored for future use"))
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("Press Enter to continue"))

	case stateGenerating:
		s.WriteString(m.spinner.View())
		s.WriteString(infoStyle.Render(" Analyzing your changes and generating commit message..."))
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("This may take a few seconds"))

	case stateShowResult:
		s.WriteString(successStyle.Render("Commit message generated!"))
		s.WriteString("\n")

		// Show the commit message in a styled box
		s.WriteString(codeBlockStyle.Render(m.commitMsg))
		s.WriteString("\n\n")

		// Menu
		s.WriteString(infoStyle.Render("What would you like to do?"))
		s.WriteString("\n\n")

		for i, item := range m.menuItems {
			cursor := "  "
			style := menuItemStyle
			if i == m.selectedItem {
				cursor = "▸ "
				style = selectedMenuItemStyle
			}
			s.WriteString(fmt.Sprintf("%s%s %s\n", cursor, warningStyle.Render(fmt.Sprintf("[%d]", i+1)), style.Render(item)))
		}

		s.WriteString("\n")
		s.WriteString(helpStyle.Render("Use ↑/↓ or j/k to navigate • Enter to select • q to quit"))

	case stateError:
		s.WriteString(errorStyle.Render("Error occurred!"))
		s.WriteString("\n\n")
		s.WriteString(boxStyle.Render(m.err.Error()))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Press Enter or q to exit"))

	case stateSuccess:
		s.WriteString(successStyle.Render("Success!"))
		s.WriteString("\n\n")
		if m.selectedItem == 0 {
			s.WriteString(infoStyle.Render("Commit message copied to clipboard"))
		} else if m.selectedItem == 1 {
			s.WriteString(infoStyle.Render("Commit executed successfully"))
		}
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Press Enter or q to exit"))
	}

	return s.String()
}

func getGitDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}

	return out.String(), nil
}

func generateCommitMessage(ctx context.Context, client *genai.Client, diff string) (string, error) {
	// Limit diff size to avoid token limits
	if len(diff) > 8000 {
		diff = diff[:8000] + "\n... (diff truncated)"
	}

	prompt := fmt.Sprintf("Analyze this git diff and generate a properly formatted commit message:\n\n```\n%s\n```", diff)

	response, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		[]*genai.Content{genai.NewContentFromText(prompt, genai.RoleUser)},
		&genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleUser),
		},
	)
	if err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}

	if len(response.Candidates) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	candidate := response.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	// Extract text from response
	var result strings.Builder
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			result.WriteString(part.Text)
		}
	}

	return strings.TrimSpace(result.String()), nil
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
