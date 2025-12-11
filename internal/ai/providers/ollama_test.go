package providers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNewOllamaProvider tests the NewOllamaProvider constructor
func TestNewOllamaProvider(t *testing.T) {
	t.Run("creates provider with custom URL", func(t *testing.T) {
		provider, err := NewOllamaProvider("http://custom:8080", "llama3")
		if err != nil {
			t.Fatalf("NewOllamaProvider failed: %v", err)
		}
		if provider == nil {
			t.Fatal("expected non-nil provider")
		}
		if provider.baseURL != "http://custom:8080" {
			t.Errorf("expected baseURL 'http://custom:8080', got %s", provider.baseURL)
		}
		if provider.model != "llama3" {
			t.Errorf("expected model 'llama3', got %s", provider.model)
		}
	})

	t.Run("uses default URL when empty", func(t *testing.T) {
		provider, err := NewOllamaProvider("", "llama3")
		if err != nil {
			t.Fatalf("NewOllamaProvider failed: %v", err)
		}
		if provider.baseURL != "http://localhost:11434" {
			t.Errorf("expected default baseURL 'http://localhost:11434', got %s", provider.baseURL)
		}
	})

	t.Run("stores model name", func(t *testing.T) {
		provider, err := NewOllamaProvider("http://localhost:11434", "qwen2.5-coder:3b")
		if err != nil {
			t.Fatalf("NewOllamaProvider failed: %v", err)
		}
		if provider.model != "qwen2.5-coder:3b" {
			t.Errorf("expected model 'qwen2.5-coder:3b', got %s", provider.model)
		}
	})
}

// TestOllamaProviderName tests the Name method
func TestOllamaProviderName(t *testing.T) {
	provider, _ := NewOllamaProvider("http://localhost:11434", "llama3")

	name := provider.Name()
	if name != "ollama" {
		t.Errorf("expected name 'ollama', got %s", name)
	}
}

// TestOllamaGenerateCommitMessage tests the GenerateCommitMessage method
func TestOllamaGenerateCommitMessage(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		expectedResponse := `git commit -m "feat(api): ✨ add new endpoint"`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request method
			if r.Method != "POST" {
				t.Errorf("expected POST method, got %s", r.Method)
			}

			// Verify request path
			if r.URL.Path != "/api/generate" {
				t.Errorf("expected path '/api/generate', got %s", r.URL.Path)
			}

			// Verify content type
			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got %s", contentType)
			}

			// Read and verify request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read request body: %v", err)
			}

			var reqBody map[string]any
			if err := json.Unmarshal(body, &reqBody); err != nil {
				t.Fatalf("failed to parse request body: %v", err)
			}

			// Verify request fields
			if reqBody["model"] != "test-model" {
				t.Errorf("expected model 'test-model', got %v", reqBody["model"])
			}
			if reqBody["stream"] != false {
				t.Errorf("expected stream false, got %v", reqBody["stream"])
			}
			if reqBody["system"] == nil || reqBody["system"] == "" {
				t.Error("expected system prompt to be set")
			}

			// Send response
			resp := map[string]any{
				"response": expectedResponse,
				"done":     true,
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		provider, _ := NewOllamaProvider(server.URL, "test-model")
		ctx := context.Background()

		result, err := provider.GenerateCommitMessage(ctx, "diff --git a/test.go")
		if err != nil {
			t.Fatalf("GenerateCommitMessage failed: %v", err)
		}
		if result != expectedResponse {
			t.Errorf("expected response '%s', got '%s'", expectedResponse, result)
		}
	})

	t.Run("truncates large diffs", func(t *testing.T) {
		var receivedPrompt string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			var reqBody map[string]any
			json.Unmarshal(body, &reqBody)
			receivedPrompt = reqBody["prompt"].(string)

			resp := map[string]any{
				"response": "test response",
				"done":     true,
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		provider, _ := NewOllamaProvider(server.URL, "test-model")
		ctx := context.Background()

		// Create a diff larger than MaxDiffSize
		largeDiff := strings.Repeat("a", MaxDiffSize+1000)

		_, err := provider.GenerateCommitMessage(ctx, largeDiff)
		if err != nil {
			t.Fatalf("GenerateCommitMessage failed: %v", err)
		}

		// Check that diff was truncated
		if !strings.Contains(receivedPrompt, "... (diff truncated)") {
			t.Error("expected diff to be truncated")
		}
	})

	t.Run("handles HTTP errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		provider, _ := NewOllamaProvider(server.URL, "test-model")
		ctx := context.Background()

		_, err := provider.GenerateCommitMessage(ctx, "test diff")
		// The error might come from JSON parsing the empty response
		if err == nil {
			t.Error("expected error for HTTP error response")
		}
	})

	t.Run("handles connection errors", func(t *testing.T) {
		provider, _ := NewOllamaProvider("http://localhost:99999", "test-model")
		ctx := context.Background()

		_, err := provider.GenerateCommitMessage(ctx, "test diff")
		if err == nil {
			t.Error("expected error for connection failure")
		}
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			time.Sleep(100 * time.Millisecond)
			resp := map[string]any{"response": "test", "done": true}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		provider, _ := NewOllamaProvider(server.URL, "test-model")
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := provider.GenerateCommitMessage(ctx, "test diff")
		if err == nil {
			t.Error("expected error for cancelled context")
		}
	})

	t.Run("handles malformed JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		provider, _ := NewOllamaProvider(server.URL, "test-model")
		ctx := context.Background()

		_, err := provider.GenerateCommitMessage(ctx, "test diff")
		if err == nil {
			t.Error("expected error for malformed JSON")
		}
	})

	t.Run("includes diff in prompt", func(t *testing.T) {
		var receivedPrompt string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			var reqBody map[string]any
			json.Unmarshal(body, &reqBody)
			receivedPrompt = reqBody["prompt"].(string)

			resp := map[string]any{"response": "test", "done": true}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		provider, _ := NewOllamaProvider(server.URL, "test-model")
		ctx := context.Background()

		testDiff := "diff --git a/main.go b/main.go\n+func main() {}"
		provider.GenerateCommitMessage(ctx, testDiff)

		if !strings.Contains(receivedPrompt, testDiff) {
			t.Errorf("prompt should contain the diff, got: %s", receivedPrompt)
		}
	})
}

// TestOllamaListModels tests the ListModels method
func TestOllamaListModels(t *testing.T) {
	t.Run("successful model listing", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request method
			if r.Method != "GET" {
				t.Errorf("expected GET method, got %s", r.Method)
			}

			// Verify request path
			if r.URL.Path != "/api/tags" {
				t.Errorf("expected path '/api/tags', got %s", r.URL.Path)
			}

			resp := map[string]any{
				"models": []map[string]any{
					{"name": "llama3:8b"},
					{"name": "qwen2.5-coder:3b"},
					{"name": "codellama:7b"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		provider, _ := NewOllamaProvider(server.URL, "test-model")
		ctx := context.Background()

		models, err := provider.ListModels(ctx)
		if err != nil {
			t.Fatalf("ListModels failed: %v", err)
		}

		if len(models) != 3 {
			t.Errorf("expected 3 models, got %d", len(models))
		}

		expectedModels := []string{"llama3:8b", "qwen2.5-coder:3b", "codellama:7b"}
		for i, expected := range expectedModels {
			if models[i] != expected {
				t.Errorf("expected model[%d] '%s', got '%s'", i, expected, models[i])
			}
		}
	})

	t.Run("empty model list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]any{
				"models": []map[string]any{},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		provider, _ := NewOllamaProvider(server.URL, "test-model")
		ctx := context.Background()

		models, err := provider.ListModels(ctx)
		if err != nil {
			t.Fatalf("ListModels failed: %v", err)
		}

		if len(models) != 0 {
			t.Errorf("expected 0 models, got %d", len(models))
		}
	})

	t.Run("handles connection errors", func(t *testing.T) {
		provider, _ := NewOllamaProvider("http://localhost:99999", "test-model")
		ctx := context.Background()

		_, err := provider.ListModels(ctx)
		if err == nil {
			t.Error("expected error for connection failure")
		}
	})

	t.Run("handles malformed JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json"))
		}))
		defer server.Close()

		provider, _ := NewOllamaProvider(server.URL, "test-model")
		ctx := context.Background()

		_, err := provider.ListModels(ctx)
		if err == nil {
			t.Error("expected error for malformed JSON")
		}
	})
}

// TestOllamaConstants tests the constants used by Ollama provider
func TestOllamaConstants(t *testing.T) {
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

	t.Run("SystemPrompt is not empty", func(t *testing.T) {
		if SystemPrompt == "" {
			t.Error("SystemPrompt should not be empty")
		}
	})
}
