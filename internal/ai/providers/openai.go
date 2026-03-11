package providers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenAIProvider struct {
	apiKey string
	model  string
}

func NewOpenAIProvider(apiKey, model string) (*OpenAIProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}
	if model == "" {
		model = "gpt-4o"
	}
	return &OpenAIProvider{apiKey: apiKey, model: model}, nil
}

func (p *OpenAIProvider) ListModels(ctx context.Context) ([]string, error) {
	client := openai.NewClient(option.WithAPIKey(p.apiKey))
	page, err := client.Models.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	var models []string
	for _, m := range page.Data {
		id := m.ID
		// Keep only chat-capable models
		if !strings.HasPrefix(id, "gpt-") &&
			!strings.HasPrefix(id, "o1") &&
			!strings.HasPrefix(id, "o3") &&
			!strings.HasPrefix(id, "o4") {
			continue
		}
		// Exclude non-chat models
		if strings.Contains(id, "-instruct") ||
			strings.Contains(id, "embedding") ||
			strings.Contains(id, "whisper") ||
			strings.Contains(id, "tts") ||
			strings.Contains(id, "dall-e") {
			continue
		}
		models = append(models, id)
	}
	sort.Strings(models)
	return models, nil
}

func (p *OpenAIProvider) GenerateCommitMessage(ctx context.Context, diff string, os string) (string, error) {
	client := openai.NewClient(option.WithAPIKey(p.apiKey))

	if len(diff) > MaxDiffSize {
		diff = diff[:MaxDiffSize] + "\n... (diff truncated)"
	}

	systemPrompt := GetOSAwarePrompt(os)
	userPrompt := fmt.Sprintf("Analyze this git diff and generate a properly formatted commit message:\n\n```\n%s\n```", diff)

	params := openai.ChatCompletionNewParams{
		Model: p.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
	}
	// o1/o3/o4 models require MaxCompletionTokens; others use MaxTokens
	if strings.HasPrefix(p.model, "o1") || strings.HasPrefix(p.model, "o3") || strings.HasPrefix(p.model, "o4") {
		params.MaxCompletionTokens = openai.Int(1024)
	} else {
		params.MaxTokens = openai.Int(1024)
	}
	resp, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	result := resp.Choices[0].Message.Content
	if result == "" {
		return "", fmt.Errorf("empty response from API")
	}

	return result, nil
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}
