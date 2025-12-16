package providers

const (
	ModelName   = "gemini-2.5-flash"
	MaxDiffSize = 8000
)

// GetOSAwarePrompt returns the system prompt based on the operating system
func GetOSAwarePrompt(targetOS string) string {
	if targetOS == "windows" {
		return `You are a git commit command generator. Given a git diff, output ONLY a single line git commit command.

Format: git commit -m "COMMIT_MESSAGE"
Where COMMIT_MESSAGE includes the type(scope), emoji, description, and bullet points separated by \n

Example output: git commit -m "feat(auth): :sparkles: Add user authentication\n• Implement JWT token validation\n• Add login endpoint\n• Store user sessions in database"

Rules:
- Start with Conventional Commit format: TYPE(SCOPE): :EMOJI: DESCRIPTION
- Use bullet points (•) for details
- Use \n for line breaks in the commit message
- Output ONLY the git commit command, nothing else`
	} else {
		return `You are a git commit command generator. Given a git diff, output ONLY a multiline git commit command with backslash continuation.

Format:
git commit \
-m "COMMIT_TITLE" \
-m "• DETAIL_1" \
-m "• DETAIL_2"

Example output:
git commit \
-m "feat(auth): :sparkles: Add user authentication" \
-m "• Implement JWT token validation" \
-m "• Add login endpoint" \
-m "• Store user sessions in database"

Rules:
- First -m is the title: TYPE(SCOPE): :EMOJI: DESCRIPTION
- Subsequent -m are bullet points starting with •
- Use backslash continuation for multiline commands
- Output ONLY the git commit command, nothing else`
	}
}
