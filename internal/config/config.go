// Package config handles configuration management for Agent Desktop.
// It provides functionality to load, save, and validate configuration for
// OpenAI-compatible endpoints (OpenAI, LM Studio, OpenRouter, etc.).
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// configDir is the directory where configuration files are stored.
// It can be overridden for testing.
var configDir = ""

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	configDir = filepath.Join(home, ".agent_desktop")
}

// Config holds the LLM configuration and execution settings.
// It supports any OpenAI-compatible endpoint including:
// - OpenAI (https://api.openai.com/v1)
// - LM Studio (http://localhost:1234/v1)
// - OpenRouter (https://openrouter.ai/api/v1)
// - Any other OpenAI-compatible API
type Config struct {
	// LLM API settings
	APIKey   string `json:"api_key"`
	Endpoint string `json:"endpoint"`   // Base URL (e.g., https://api.openai.com/v1)
	Model    string `json:"model"`      // Model name (e.g., gpt-4o, deepseek-chat)

	// Execution settings
	ExecutionTimeout int `json:"execution_timeout"`
}

// getConfigPath returns the full path to the config file.
func getConfigPath() string {
	return filepath.Join(configDir, "config.json")
}

// Load loads the configuration from disk.
// If the config file doesn't exist, it returns a default configuration.
func Load() (*Config, error) {
	configPath := getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return &Config{
				Endpoint:         "https://api.openai.com/v1",
				ExecutionTimeout: 60,
			}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Ensure default timeout if not set
	if cfg.ExecutionTimeout == 0 {
		cfg.ExecutionTimeout = 60
	}

	// Set default endpoint if not set
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://api.openai.com/v1"
	}

	return &cfg, nil
}

// Save saves the configuration to disk.
// It creates the config directory if it doesn't exist.
func (c *Config) Save() error {
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(getConfigPath(), data, 0644)
}

// Validate checks if the configuration has all required fields.
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return errors.New("api_key is required")
	}
	if c.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	if c.Model == "" {
		return errors.New("model is required")
	}
	return nil
}

// IsConfigured returns true if all required fields are set.
func (c *Config) IsConfigured() bool {
	return c.APIKey != "" &&
		c.Endpoint != "" &&
		c.Model != ""
}
