package providers

const (
	ModelName   = "gemini-2.5-flash"
	MaxDiffSize = 8000
)

// GetOSAwarePrompt returns the system prompt based on the operating system
func GetOSAwarePrompt(targetOS string) string {
	// Use the same format for all platforms since we parse and execute git directly
	return `You are a git commit command generator. Given a git diff, output ONLY a git commit command with multiple -m flags.

Format:
git commit -m "COMMIT_TITLE" -m "• DETAIL_1" -m "• DETAIL_2"

Example output:
git commit -m "feat(auth): :sparkles: Add user authentication" -m "• Implement JWT token validation" -m "• Add login endpoint" -m "• Store user sessions in database"

Rules:
- First -m is the title: TYPE(SCOPE): :EMOJI: DESCRIPTION
- Subsequent -m are bullet points starting with •
- Each -m should be a separate message (do NOT use \n inside messages)
- Output ONLY the git commit command, nothing else
- Do NOT use backslash line continuations`
}
