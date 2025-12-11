# commit-gen

A CLI tool that generates structured git commit messages using Google's Gemini AI, following the Conventional Commits standard with emojis.

## Features

- Analyzes staged git changes (`git diff --cached`)
- Generates commit messages in Conventional Commits format
- Includes relevant emojis and bullet-point details
- Copies the message to clipboard automatically
- Supports grouping related changes

## Requirements

- Go 1.25.3 or later
- Google Gemini API key

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/mounis-bhat/commit-gen.git
   cd commit-gen
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Build the binary:
   ```bash
   make build
   ```

   Or install globally:
   ```bash
   make install
   ```

## Usage

1. Stage your changes:
    ```bash
    git add .
    ```

2. Run the tool with your Gemini API key (ensure the binary is in your PATH):
    ```bash
    git-commit-generator YOUR_GEMINI_API_KEY
    ```

    The generated commit message will be displayed and copied to your clipboard.

3. Commit using the generated message:
   ```bash
   git commit -m "feat(ui): ✨ Add dark mode toggle

   • Implement theme context for state management
   • Add CSS variables for light/dark themes
   • Update components to use theme variables"
   ```

## API Key Setup

Get your Gemini API key from [Google AI Studio](https://makersuite.google.com/app/apikey).

## Examples

The tool generates messages like:

```
feat(auth): 🔐 Add user authentication

• Implement JWT token generation
• Add login/logout endpoints
• Validate user credentials against database

fix(api): 🐛 Handle null pointer in user service
• Add nil checks in UserService.GetUser()
• Return appropriate error responses
```

## License

MIT