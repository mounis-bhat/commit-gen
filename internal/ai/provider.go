package ai

import (
	"context"
	"fmt"

	"github.com/mounis-bhat/commit-gen/internal/ai/providers"
	"github.com/mounis-bhat/commit-gen/internal/config"
)

// Provider defines the interface that all AI providers must implement
type Provider interface {
	GenerateCommitMessage(ctx context.Context, diff string, os string) (string, error)
	Name() string
}

// NewProvider creates a new provider based on the configuration
func NewProvider(cfg *config.Config) (Provider, error) {
	switch cfg.Provider {
	case "gemini":
		return providers.NewGeminiProvider(cfg.GeminiAPIKey)
	case "claude":
		return providers.NewClaudeProvider(cfg.ClaudeAPIKey)
	case "ollama":
		return providers.NewOllamaProvider(cfg.OllamaURL, cfg.OllamaModel)
	case "openai":
		return providers.NewOpenAIProvider(cfg.OpenAIAPIKey, cfg.OpenAIModel)
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider)
	}
}
