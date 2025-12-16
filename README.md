# commit-gen

An interactive CLI tool that generates structured git commit messages using Google's Gemini AI, following the Conventional Commits standard with emojis. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for a beautiful terminal UI experience.

## Features

- Interactive TUI powered by Bubble Tea and Lip Gloss
- Analyzes staged git changes (`git diff --cached`)
- Generates commit messages in Conventional Commits format
- Includes relevant emojis and bullet-point details
- Supports grouping related changes
- API key stored securely for future use (no need to enter it every time)
- Multiple actions: copy to clipboard, execute commit directly, or regenerate

## Requirements

- Git installed and available in PATH
- Google Gemini API key (or Ollama for local inference)

## Installation

### Linux / macOS

```bash
curl -sSfL https://raw.githubusercontent.com/mounis-bhat/commit-gen/main/install.sh | sh
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/mounis-bhat/commit-gen/main/install.ps1 | iex
```

### From Source (requires Go 1.21+)

```bash
go install github.com/mounis-bhat/commit-gen/cmd/commit-gen@latest
```

Or clone and build manually:

```bash
git clone https://github.com/mounis-bhat/commit-gen.git
cd commit-gen
make build
# Binary will be in ./bin/commit-gen
```

## Usage

1. Stage your changes:

   ```bash
   git add .
   ```

2. Run the tool:

   ```bash
   commit-gen
   ```

3. On first run, you'll be prompted to enter your Gemini API key. The key will be saved to `~/.commit-gen-config.json` for future use.

4. Once the commit message is generated, use the interactive menu to:
   - **Copy to clipboard** - Copy the generated command to paste manually
   - **Execute commit** - Run the git commit directly
   - **Regenerate** - Generate a new commit message
   - **Quit** - Exit the application

### Controls

- `↑/↓` or `j/k` - Navigate menu options
- `Enter` - Select option
- `1-4` - Quick select menu items
- `q` or `Ctrl+C` - Quit

## API Key Setup

Get your Gemini API key from [Google AI Studio](https://aistudio.google.com/apikey).

The API key is stored in `~/.commit-gen-config.json` and is only requested once.

## Examples

The tool generates commit commands like:

```bash
git commit -m "feat(auth): Add user authentication" \
-m "• Implement JWT token generation" \
-m "• Add login/logout endpoints" \
-m "• Validate user credentials against database"
```

```bash
git commit -m "fix(api): Handle null pointer in user service" \
-m "• Add nil checks in UserService.GetUser()" \
-m "• Return appropriate error responses"
```

## Uninstall

### Linux / macOS

```bash
rm ~/.local/bin/commit-gen
```

### Windows

```powershell
Remove-Item "$env:USERPROFILE\.local\bin\commit-gen.exe"
```

## License

MIT
