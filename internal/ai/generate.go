package ai

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

const (
	// ModelName is the Gemini model to use for generation.
	ModelName = "gemini-2.5-flash"
	// MaxDiffSize is the maximum size of diff to send to the API.
	MaxDiffSize = 8000
)

// GenerateCommitMessage generates a commit message from the given diff using the Gemini API.
func GenerateCommitMessage(ctx context.Context, apiKey, diff string) (string, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Limit diff size to avoid token limits
	if len(diff) > MaxDiffSize {
		diff = diff[:MaxDiffSize] + "\n... (diff truncated)"
	}

	prompt := fmt.Sprintf("Analyze this git diff and generate a properly formatted commit message:\n\n```\n%s\n```", diff)

	response, err := client.Models.GenerateContent(
		ctx,
		ModelName,
		[]*genai.Content{genai.NewContentFromText(prompt, genai.RoleUser)},
		&genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(SystemPrompt, genai.RoleUser),
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
