package providers

const (
	ModelName   = "gemini-2.5-flash"
	MaxDiffSize = 8000
)

// GetOSAwarePrompt returns the system prompt based on the operating system
func GetOSAwarePrompt(targetOS string) string {

	basePrompt := `You are an expert at writing clear, structured git commit messages following the Conventional Commits standard with emojis.

Rules:
1. Use Conventional Commits: feat, fix, refactor, ui, docs, test, chore, perf, style, etc.
2. Add relevant emoji (sparkles :sparkles: for features, bug :bug: for fixes, recycle :recycle: for refactors, lipstick :lipstick: for UI, etc.)
3. Keep scope concise and descriptive
4. Use bullet points (•) for implementation details
5. Group related changes together with blank lines between groups
6. Start descriptions with action verbs
7. Be specific about what changed and why
8. Do not use backticks or markdown formatting for code; use plain text

Generate ONLY the commit message content, nothing else. No explanations or markdown.`

	if targetOS == "windows" {
		return basePrompt + `

When given a git diff, generate a SINGLE LINE git commit command in this EXACT format:
git commit -m "TYPE(SCOPE): EMOJI DESCRIPTION\n• First bullet point detail\n• Second bullet point detail\n\nTYPE(SCOPE): EMOJI DESCRIPTION\n• Detail about this commit"

Use \n for line breaks and \\n for blank lines within the single quoted -m parameter.`
	} else {
		return basePrompt + `

When given a git diff, generate a MULTILINE git commit command with backslash continuation in this EXACT format:
git commit \
-m "TYPE(SCOPE): EMOJI DESCRIPTION" \
-m "• First bullet point detail" \
-m "• Second bullet point detail" \
\
-m "TYPE(SCOPE): EMOJI DESCRIPTION" \
-m "• Detail about this commit"`
	}
}
