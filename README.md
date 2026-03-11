# commit-gen

An interactive terminal app that generates structured git commit commands using AI. It supports multiple providers, follows Conventional Commits style with emojis, and gives you a guided TUI workflow for selecting files, generating messages, and running commits.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and Lip Gloss.

## Features

- Multi-provider AI support: Ollama (local), Gemini, Claude, and OpenAI
- Provider and model selection in the TUI (Ollama and OpenAI models are fetched dynamically)
- Interactive file picker for staged, unstaged, and untracked files
- Built-in staging and unstaging based on your file selection
- Conventional Commit style output with emoji and bullet-point details
- Action menu after generation:
  - Copy to clipboard
  - Execute commit
  - Execute commit and push
  - Regenerate
  - Quit
- Saved configuration in `~/.commit-gen-config.json` (provider, keys, and selected models)

## Requirements

- Git installed and available in `PATH`
- One of the following:
  - Ollama running locally (default URL: `http://localhost:11434`)
  - Gemini API key from [Google AI Studio](https://aistudio.google.com/apikey)
  - Claude API key from [Anthropic Console](https://console.anthropic.com/settings/keys)
  - OpenAI API key from [OpenAI Platform](https://platform.openai.com/api-keys)

## Installation

### Linux / macOS

```bash
curl -sSfL https://raw.githubusercontent.com/mounis-bhat/commit-gen/trunk/install.sh | sh
```

The installer places the binary in `~/.local/bin`, upgrades existing installs, and adds `~/.local/bin` to your shell profile if needed.

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/mounis-bhat/commit-gen/trunk/install.ps1 | iex
```

The installer places the binary in `$env:USERPROFILE\.local\bin` and can help you add it to your user `PATH`.

### From Source (Go 1.22+ recommended)

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

1. From inside any git repository, run:

   ```bash
   commit-gen
   ```

2. In the TUI workflow:
   - Select an AI provider
   - Enter API key if required (cloud providers)
   - Select a model (Ollama/OpenAI)
   - Select files to include in the commit
   - Generate commit message
   - Choose action (copy, commit, commit+push, regenerate, quit)

3. The generated output is a `git commit` command with multiple `-m` flags, ready to execute directly.

## Controls

- `↑/↓` or `j/k`: Navigate lists and menus
- `Enter`: Confirm selection / continue
- `Space`: Toggle file selection
- `a`: Toggle select all files
- `1-5`: Quick select action menu items
- `q` or `Ctrl+C`: Quit

## Configuration

Configuration is stored at:

```text
~/.commit-gen-config.json
```

It can include:

- Selected provider (`ollama`, `gemini`, `claude`, `openai`)
- Provider API keys
- Ollama URL and selected Ollama model
- Selected OpenAI model

## Example Output

```bash
git commit -m "feat(auth): :sparkles: Add user authentication" -m "• Implement JWT token validation" -m "• Add login endpoint" -m "• Store user sessions in database"
```

```bash
git commit -m "fix(api): :bug: Handle nil pointer in user service" -m "• Add nil checks in UserService.GetUser()" -m "• Return consistent error responses"
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
