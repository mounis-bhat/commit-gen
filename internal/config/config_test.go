package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestHelper provides utility functions for config testing
type TestHelper struct {
	t          *testing.T
	tempDir    string
	oldHome    string
	configPath string
}

// NewTestHelper creates a new test helper with a temporary home directory
func NewTestHelper(t *testing.T) *TestHelper {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Save and override HOME environment variable
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	return &TestHelper{
		t:          t,
		tempDir:    tmpDir,
		oldHome:    oldHome,
		configPath: filepath.Join(tmpDir, ".commit-gen-config.json"),
	}
}

// Cleanup restores the original HOME and removes temp directory
func (h *TestHelper) Cleanup() {
	os.Setenv("HOME", h.oldHome)
	os.RemoveAll(h.tempDir)
}

// WriteConfig writes a config file with the given content
func (h *TestHelper) WriteConfig(cfg *Config) {
	h.t.Helper()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		h.t.Fatalf("failed to marshal config: %v", err)
	}
	if err := os.WriteFile(h.configPath, data, 0600); err != nil {
		h.t.Fatalf("failed to write config: %v", err)
	}
}

// WriteRawConfig writes raw JSON content to the config file
func (h *TestHelper) WriteRawConfig(content string) {
	h.t.Helper()
	if err := os.WriteFile(h.configPath, []byte(content), 0600); err != nil {
		h.t.Fatalf("failed to write raw config: %v", err)
	}
}

// TestGetConfigPath tests the GetConfigPath function
func TestGetConfigPath(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("returns path in home directory", func(t *testing.T) {
		path := GetConfigPath()
		expectedPath := filepath.Join(h.tempDir, ".commit-gen-config.json")
		if path != expectedPath {
			t.Errorf("expected %s, got %s", expectedPath, path)
		}
	})

	t.Run("path ends with correct filename", func(t *testing.T) {
		path := GetConfigPath()
		filename := filepath.Base(path)
		if filename != ".commit-gen-config.json" {
			t.Errorf("expected filename .commit-gen-config.json, got %s", filename)
		}
	})
}

// TestLoad tests the Load function
func TestLoad(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("returns error when config file does not exist", func(t *testing.T) {
		_, err := Load()
		if err == nil {
			t.Error("expected error when config file does not exist")
		}
	})

	t.Run("loads valid gemini config", func(t *testing.T) {
		cfg := &Config{
			Provider: "gemini",
			APIKey:   "test-api-key-12345",
		}
		h.WriteConfig(cfg)

		loaded, err := Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if loaded.Provider != "gemini" {
			t.Errorf("expected provider 'gemini', got %s", loaded.Provider)
		}
		if loaded.APIKey != "test-api-key-12345" {
			t.Errorf("expected API key 'test-api-key-12345', got %s", loaded.APIKey)
		}
	})

	t.Run("loads valid ollama config", func(t *testing.T) {
		cfg := &Config{
			Provider:    "ollama",
			OllamaURL:   "http://localhost:11434",
			OllamaModel: "llama3:8b",
		}
		h.WriteConfig(cfg)

		loaded, err := Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if loaded.Provider != "ollama" {
			t.Errorf("expected provider 'ollama', got %s", loaded.Provider)
		}
		if loaded.OllamaURL != "http://localhost:11434" {
			t.Errorf("expected OllamaURL 'http://localhost:11434', got %s", loaded.OllamaURL)
		}
		if loaded.OllamaModel != "llama3:8b" {
			t.Errorf("expected OllamaModel 'llama3:8b', got %s", loaded.OllamaModel)
		}
	})

	t.Run("sets default OllamaURL when empty", func(t *testing.T) {
		cfg := &Config{
			Provider:    "ollama",
			OllamaURL:   "", // Empty
			OllamaModel: "llama3:8b",
		}
		h.WriteConfig(cfg)

		loaded, err := Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if loaded.OllamaURL != "http://localhost:11434" {
			t.Errorf("expected default OllamaURL 'http://localhost:11434', got %s", loaded.OllamaURL)
		}
	})

	t.Run("sets default OllamaModel when empty", func(t *testing.T) {
		cfg := &Config{
			Provider:    "ollama",
			OllamaURL:   "http://localhost:11434",
			OllamaModel: "", // Empty
		}
		h.WriteConfig(cfg)

		loaded, err := Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if loaded.OllamaModel != "qwen2.5-coder:3b" {
			t.Errorf("expected default OllamaModel 'qwen2.5-coder:3b', got %s", loaded.OllamaModel)
		}
	})

	t.Run("returns error for malformed JSON", func(t *testing.T) {
		h.WriteRawConfig("{invalid json")

		_, err := Load()
		if err == nil {
			t.Error("expected error for malformed JSON")
		}
	})

	t.Run("loads config with all fields", func(t *testing.T) {
		cfg := &Config{
			Provider:    "gemini",
			APIKey:      "my-api-key",
			OllamaURL:   "http://custom:8080",
			OllamaModel: "custom-model",
		}
		h.WriteConfig(cfg)

		loaded, err := Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if loaded.Provider != cfg.Provider {
			t.Errorf("Provider mismatch: expected %s, got %s", cfg.Provider, loaded.Provider)
		}
		if loaded.APIKey != cfg.APIKey {
			t.Errorf("APIKey mismatch: expected %s, got %s", cfg.APIKey, loaded.APIKey)
		}
		if loaded.OllamaURL != cfg.OllamaURL {
			t.Errorf("OllamaURL mismatch: expected %s, got %s", cfg.OllamaURL, loaded.OllamaURL)
		}
		if loaded.OllamaModel != cfg.OllamaModel {
			t.Errorf("OllamaModel mismatch: expected %s, got %s", cfg.OllamaModel, loaded.OllamaModel)
		}
	})
}

// TestSave tests the Save function
func TestSave(t *testing.T) {
	h := NewTestHelper(t)
	defer h.Cleanup()

	t.Run("saves config to file", func(t *testing.T) {
		cfg := &Config{
			Provider:    "gemini",
			APIKey:      "test-key",
			OllamaURL:   "http://localhost:11434",
			OllamaModel: "llama3",
		}

		err := Save(cfg)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(h.configPath); os.IsNotExist(err) {
			t.Error("config file was not created")
		}
	})

	t.Run("saved config can be loaded", func(t *testing.T) {
		cfg := &Config{
			Provider:    "ollama",
			APIKey:      "saved-key",
			OllamaURL:   "http://custom:9999",
			OllamaModel: "custom",
		}

		err := Save(cfg)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		loaded, err := Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if loaded.Provider != cfg.Provider {
			t.Errorf("Provider mismatch: expected %s, got %s", cfg.Provider, loaded.Provider)
		}
		if loaded.APIKey != cfg.APIKey {
			t.Errorf("APIKey mismatch: expected %s, got %s", cfg.APIKey, loaded.APIKey)
		}
	})

	t.Run("sets correct file permissions", func(t *testing.T) {
		cfg := &Config{Provider: "gemini", APIKey: "secret"}

		err := Save(cfg)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		info, err := os.Stat(h.configPath)
		if err != nil {
			t.Fatalf("failed to stat config file: %v", err)
		}

		// Check permissions are 0600 (owner read/write only)
		perm := info.Mode().Perm()
		if perm != 0600 {
			t.Errorf("expected permissions 0600, got %o", perm)
		}
	})

	t.Run("overwrites existing config", func(t *testing.T) {
		// Save first config
		cfg1 := &Config{Provider: "gemini", APIKey: "first-key"}
		Save(cfg1)

		// Save second config
		cfg2 := &Config{Provider: "ollama", APIKey: "second-key"}
		err := Save(cfg2)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		loaded, err := Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if loaded.Provider != "ollama" {
			t.Errorf("expected provider 'ollama', got %s", loaded.Provider)
		}
		if loaded.APIKey != "second-key" {
			t.Errorf("expected API key 'second-key', got %s", loaded.APIKey)
		}
	})

	t.Run("produces valid JSON", func(t *testing.T) {
		cfg := &Config{
			Provider:    "gemini",
			APIKey:      "json-key",
			OllamaURL:   "http://localhost:11434",
			OllamaModel: "model",
		}

		err := Save(cfg)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Read raw file content
		data, err := os.ReadFile(h.configPath)
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}

		// Verify it's valid JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Errorf("saved config is not valid JSON: %v", err)
		}
	})
}

// TestConfig tests the Config struct
func TestConfig(t *testing.T) {
	t.Run("struct fields", func(t *testing.T) {
		cfg := Config{
			Provider:    "gemini",
			APIKey:      "key123",
			OllamaURL:   "http://localhost:11434",
			OllamaModel: "llama3",
		}

		if cfg.Provider != "gemini" {
			t.Errorf("expected provider 'gemini', got %s", cfg.Provider)
		}
		if cfg.APIKey != "key123" {
			t.Errorf("expected API key 'key123', got %s", cfg.APIKey)
		}
		if cfg.OllamaURL != "http://localhost:11434" {
			t.Errorf("expected OllamaURL 'http://localhost:11434', got %s", cfg.OllamaURL)
		}
		if cfg.OllamaModel != "llama3" {
			t.Errorf("expected OllamaModel 'llama3', got %s", cfg.OllamaModel)
		}
	})

	t.Run("json tags", func(t *testing.T) {
		cfg := &Config{
			Provider:    "ollama",
			APIKey:      "test",
			OllamaURL:   "http://test:1234",
			OllamaModel: "test-model",
		}

		data, err := json.Marshal(cfg)
		if err != nil {
			t.Fatalf("failed to marshal config: %v", err)
		}

		jsonStr := string(data)

		// Verify JSON field names
		expectedFields := []string{
			`"provider"`,
			`"api_key"`,
			`"ollama_url"`,
			`"ollama_model"`,
		}

		for _, field := range expectedFields {
			if !containsString(jsonStr, field) {
				t.Errorf("JSON should contain field %s, got: %s", field, jsonStr)
			}
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
