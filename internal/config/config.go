// Package config handles configuration management for Agent Desktop.
// It provides functionality to load, save, and validate Azure OpenAI configuration.
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

// Config holds the Azure OpenAI configuration and execution settings.
type Config struct {
	// Azure OpenAI settings
	OpenAISubscriptionKey string `json:"openai_subscription_key"`
	OpenAIEndpoint        string `json:"openai_endpoint"`
	OpenAIDeployment      string `json:"openai_deployment"`
	OpenAIModelName       string `json:"openai_model_name"`

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
	if c.OpenAISubscriptionKey == "" {
		return errors.New("openai_subscription_key is required")
	}
	if c.OpenAIEndpoint == "" {
		return errors.New("openai_endpoint is required")
	}
	if c.OpenAIDeployment == "" {
		return errors.New("openai_deployment is required")
	}
	if c.OpenAIModelName == "" {
		return errors.New("openai_model_name is required")
	}
	return nil
}

// IsConfigured returns true if all required Azure OpenAI fields are set.
func (c *Config) IsConfigured() bool {
	return c.OpenAISubscriptionKey != "" &&
		c.OpenAIEndpoint != "" &&
		c.OpenAIDeployment != "" &&
		c.OpenAIModelName != ""
}
