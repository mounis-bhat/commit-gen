package ai

import (
	"strings"
	"testing"

	"github.com/mounis-bhat/commit-gen/internal/ai/providers"
	"github.com/mounis-bhat/commit-gen/internal/config"
)

// TestNewProvider tests the NewProvider factory function
func TestNewProvider(t *testing.T) {
	t.Run("creates gemini provider", func(t *testing.T) {
		cfg := &config.Config{
			Provider: "gemini",
			APIKey:   "test-api-key",
		}

		provider, err := NewProvider(cfg)
		if err != nil {
			t.Fatalf("NewProvider failed: %v", err)
		}
		if provider == nil {
			t.Fatal("expected non-nil provider")
		}
		if provider.Name() != "gemini" {
			t.Errorf("expected provider name 'gemini', got %s", provider.Name())
		}
	})

	t.Run("creates ollama provider", func(t *testing.T) {
		cfg := &config.Config{
			Provider:    "ollama",
			OllamaURL:   "http://localhost:11434",
			OllamaModel: "llama3",
		}

		provider, err := NewProvider(cfg)
		if err != nil {
			t.Fatalf("NewProvider failed: %v", err)
		}
		if provider == nil {
			t.Fatal("expected non-nil provider")
		}
		if provider.Name() != "ollama" {
			t.Errorf("expected provider name 'ollama', got %s", provider.Name())
		}
	})

	t.Run("creates ollama provider with default URL", func(t *testing.T) {
		cfg := &config.Config{
			Provider:    "ollama",
			OllamaURL:   "", // Empty should use default
			OllamaModel: "llama3",
		}

		provider, err := NewProvider(cfg)
		if err != nil {
			t.Fatalf("NewProvider failed: %v", err)
		}
		if provider == nil {
			t.Fatal("expected non-nil provider")
		}
		if provider.Name() != "ollama" {
			t.Errorf("expected provider name 'ollama', got %s", provider.Name())
		}
	})

	t.Run("returns error for unknown provider", func(t *testing.T) {
		cfg := &config.Config{
			Provider: "unknown",
		}

		provider, err := NewProvider(cfg)
		if err == nil {
			t.Error("expected error for unknown provider")
		}
		if provider != nil {
			t.Error("expected nil provider for error case")
		}
	})

	t.Run("returns error for empty provider", func(t *testing.T) {
		cfg := &config.Config{
			Provider: "",
		}

		provider, err := NewProvider(cfg)
		if err == nil {
			t.Error("expected error for empty provider")
		}
		if provider != nil {
			t.Error("expected nil provider for error case")
		}
	})

	t.Run("error message contains provider name", func(t *testing.T) {
		cfg := &config.Config{
			Provider: "invalid-provider",
		}

		_, err := NewProvider(cfg)
		if err == nil {
			t.Fatal("expected error")
		}
		errMsg := err.Error()
		if !containsString(errMsg, "invalid-provider") {
			t.Errorf("error message should contain provider name, got: %s", errMsg)
		}
	})
}

// TestProviderInterface tests that providers implement the Provider interface
func TestProviderInterface(t *testing.T) {
	t.Run("gemini provider implements interface", func(t *testing.T) {
		cfg := &config.Config{
			Provider: "gemini",
			APIKey:   "test-key",
		}

		var provider Provider
		var err error
		provider, err = NewProvider(cfg)
		if err != nil {
			t.Fatalf("NewProvider failed: %v", err)
		}

		// Verify Name() method exists and works
		name := provider.Name()
		if name != "gemini" {
			t.Errorf("expected 'gemini', got %s", name)
		}
	})

	t.Run("ollama provider implements interface", func(t *testing.T) {
		cfg := &config.Config{
			Provider:    "ollama",
			OllamaURL:   "http://localhost:11434",
			OllamaModel: "test-model",
		}

		var provider Provider
		var err error
		provider, err = NewProvider(cfg)
		if err != nil {
			t.Fatalf("NewProvider failed: %v", err)
		}

		// Verify Name() method exists and works
		name := provider.Name()
		if name != "ollama" {
			t.Errorf("expected 'ollama', got %s", name)
		}
	})
}

// TestOSAwarePrompt tests that OS-aware prompts are generated
func TestOSAwarePrompt(t *testing.T) {
	t.Run("windows prompt contains single line", func(t *testing.T) {
		prompt := providers.GetOSAwarePrompt("windows")
		if !strings.Contains(prompt, "SINGLE LINE") {
			t.Error("Windows prompt should mention SINGLE LINE")
		}
	})

	t.Run("posix prompt contains multiline", func(t *testing.T) {
		prompt := providers.GetOSAwarePrompt("linux")
		if !strings.Contains(prompt, "MULTILINE") {
			t.Error("POSIX prompt should mention MULTILINE")
		}
	})

	t.Run("prompt contains conventional commits", func(t *testing.T) {
		prompt := providers.GetOSAwarePrompt("linux")
		if !containsString(prompt, "Conventional Commits") {
			t.Error("Prompt should mention Conventional Commits")
		}
	})

	t.Run("prompt contains commit types", func(t *testing.T) {
		prompt := providers.GetOSAwarePrompt("linux")
		expectedTypes := []string{"feat", "fix", "refactor"}
		for _, typ := range expectedTypes {
			if !containsString(prompt, typ) {
				t.Errorf("Prompt should contain commit type '%s'", typ)
			}
		}
	})

	t.Run("system prompt mentions emojis", func(t *testing.T) {
		prompt := providers.GetOSAwarePrompt("linux")
		if !containsString(prompt, "emoji") {
			t.Error("Prompt should mention emojis")
		}
	})
}

// containsString checks if s contains substr
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
