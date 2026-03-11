package providers

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const ClaudeModelName = anthropic.ModelClaudeHaiku4_5_20251001

type ClaudeProvider struct {
	apiKey string
}

func NewClaudeProvider(apiKey string) (*ClaudeProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Claude API key is required")
	}
	return &ClaudeProvider{apiKey: apiKey}, nil
}

func (p *ClaudeProvider) GenerateCommitMessage(ctx context.Context, diff string, os string) (string, error) {
	client := anthropic.NewClient(option.WithAPIKey(p.apiKey))

	if len(diff) > MaxDiffSize {
		diff = diff[:MaxDiffSize] + "\n... (diff truncated)"
	}

	systemPrompt := GetOSAwarePrompt(os)
	userPrompt := fmt.Sprintf("Analyze this git diff and generate a properly formatted commit message:\n\n```\n%s\n```", diff)

	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     ClaudeModelName,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	var result string
	for _, block := range message.Content {
		if block.Type == "text" {
			result += block.Text
		}
	}

	if result == "" {
		return "", fmt.Errorf("empty response from API")
	}

	return result, nil
}

func (p *ClaudeProvider) Name() string {
	return "claude"
}
