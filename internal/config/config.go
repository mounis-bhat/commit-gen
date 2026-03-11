package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the application configuration.
type Config struct {
	Provider     string `json:"provider"`      // "gemini", "claude", or "ollama"
	APIKey       string `json:"api_key"`       // deprecated: legacy Gemini key, migrated on load
	GeminiAPIKey string `json:"gemini_api_key"` // for Gemini
	ClaudeAPIKey string `json:"claude_api_key"` // for Claude
	OpenAIAPIKey string `json:"openai_api_key"` // for OpenAI
	OllamaURL    string `json:"ollama_url"`    // default: http://localhost:11434
	OllamaModel  string `json:"ollama_model"`  // e.g., "qwen2.5-coder:3b"
	OpenAIModel  string `json:"openai_model"`  // e.g., "gpt-4o"
}

// GetConfigPath returns the path to the configuration file.
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".commit-gen-config.json"
	}
	return filepath.Join(homeDir, ".commit-gen-config.json")
}

// Load reads the configuration from the config file.
func Load() (*Config, error) {
	configPath := GetConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Migrate legacy api_key to gemini_api_key
	if config.APIKey != "" && config.GeminiAPIKey == "" {
		config.GeminiAPIKey = config.APIKey
		config.APIKey = ""
	}

	// Set defaults for Ollama
	if config.OllamaURL == "" {
		config.OllamaURL = "http://localhost:11434"
	}
	if config.OllamaModel == "" {
		config.OllamaModel = "qwen2.5-coder:3b"
	}

	return &config, nil
}

// Save writes the configuration to the config file.
func Save(config *Config) error {
	configPath := GetConfigPath()
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}
