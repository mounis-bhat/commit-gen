package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OllamaProvider struct {
	baseURL string
	model   string
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	System string `json:"system"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func NewOllamaProvider(baseURL, model string) (*OllamaProvider, error) {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaProvider{baseURL: baseURL, model: model}, nil
}

func (p *OllamaProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	// Limit diff size to avoid token limits
	if len(diff) > MaxDiffSize {
		diff = diff[:MaxDiffSize] + "\n... (diff truncated)"
	}

	reqBody := ollamaRequest{
		Model:  p.model,
		Prompt: fmt.Sprintf("Analyze this git diff and generate a properly formatted commit message:\n\n```\n%s\n```", diff),
		System: SystemPrompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result ollamaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.Response, nil
}

func (p *OllamaProvider) Name() string {
	return "ollama"
}
