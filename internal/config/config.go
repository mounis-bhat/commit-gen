package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the application configuration.
type Config struct {
	Provider    string `json:"provider"`     // "gemini" or "ollama"
	APIKey      string `json:"api_key"`      // for gemini
	OllamaURL   string `json:"ollama_url"`   // default: http://localhost:11434
	OllamaModel string `json:"ollama_model"` // e.g., "qwen2.5-coder:3b"
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
