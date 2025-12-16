package providers

import (
	"context"
	"strings"
	"testing"
)

// TestNewGeminiProvider tests the NewGeminiProvider constructor
func TestNewGeminiProvider(t *testing.T) {
	t.Run("creates provider with API key", func(t *testing.T) {
		provider, err := NewGeminiProvider("test-api-key")
		if err != nil {
			t.Fatalf("NewGeminiProvider failed: %v", err)
		}
		if provider == nil {
			t.Fatal("expected non-nil provider")
		}
		if provider.apiKey != "test-api-key" {
			t.Errorf("expected apiKey 'test-api-key', got %s", provider.apiKey)
		}
	})

	t.Run("accepts empty API key", func(t *testing.T) {
		provider, err := NewGeminiProvider("")
		if err != nil {
			t.Fatalf("NewGeminiProvider failed: %v", err)
		}
		if provider.apiKey != "" {
			t.Errorf("expected empty apiKey, got %s", provider.apiKey)
		}
	})
}

// TestGeminiProviderName tests the Name method
func TestGeminiProviderName(t *testing.T) {
	provider, _ := NewGeminiProvider("test-key")

	name := provider.Name()
	if name != "gemini" {
		t.Errorf("expected name 'gemini', got %s", name)
	}
}

// TestGeminiGenerateCommitMessage tests the GenerateCommitMessage method
func TestGeminiGenerateCommitMessage(t *testing.T) {
	t.Run("handles invalid API key", func(t *testing.T) {
		provider, _ := NewGeminiProvider("invalid-key")
		ctx := context.Background()

		// This will fail due to invalid API key, but we can test the error handling
		_, err := provider.GenerateCommitMessage(ctx, "test diff", "linux")
		if err == nil {
			t.Error("expected error with invalid API key")
		}
	})

	t.Run("handles empty API key", func(t *testing.T) {
		provider, _ := NewGeminiProvider("")
		ctx := context.Background()

		_, err := provider.GenerateCommitMessage(ctx, "test diff", "linux")
		if err == nil {
			t.Error("expected error with empty API key")
		}
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		provider, _ := NewGeminiProvider("invalid-key")
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := provider.GenerateCommitMessage(ctx, "test diff", "linux")
		if err == nil {
			t.Error("expected error for cancelled context")
		}
	})

	t.Run("handles empty diff", func(t *testing.T) {
		provider, _ := NewGeminiProvider("invalid-key")
		ctx := context.Background()

		_, err := provider.GenerateCommitMessage(ctx, "", "linux")
		if err == nil {
			t.Error("expected error with empty diff")
		}
	})
}

// TestGeminiConstants tests the constants used by Gemini provider
func TestGeminiConstants(t *testing.T) {
	t.Run("ModelName is set", func(t *testing.T) {
		if ModelName == "" {
			t.Error("ModelName should not be empty")
		}
		if !strings.Contains(ModelName, "gemini") {
			t.Errorf("ModelName should contain 'gemini', got %s", ModelName)
		}
	})

	t.Run("MaxDiffSize is reasonable", func(t *testing.T) {
		if MaxDiffSize <= 0 {
			t.Error("MaxDiffSize should be positive")
		}
		if MaxDiffSize < 1000 {
			t.Error("MaxDiffSize should be at least 1000 characters")
		}
		if MaxDiffSize > 100000 {
			t.Error("MaxDiffSize should not be too large to avoid API limits")
		}
	})

	t.Run("GetOSAwarePrompt works", func(t *testing.T) {
		prompt := GetOSAwarePrompt("linux")
		if prompt == "" {
			t.Error("GetOSAwarePrompt should not be empty")
		}
	})
}

// TestGeminiProviderStruct tests the GeminiProvider struct
func TestGeminiProviderStruct(t *testing.T) {
	t.Run("struct has apiKey field", func(t *testing.T) {
		provider := &GeminiProvider{apiKey: "test"}
		if provider.apiKey != "test" {
			t.Errorf("expected apiKey 'test', got %s", provider.apiKey)
		}
	})
}
