// Package config handles configuration management for Agent Desktop.
// It provides functionality to load, save, and validate configuration for
// OpenAI-compatible endpoints (OpenAI, LM Studio, OpenRouter, etc.).
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
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

// Mode selects which tool surface the agent exposes.
const (
	ModeGeneral  = "GENERAL"
	ModeCloudOps = "CLOUD_OPS"
)

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

	// Mode selects the agent's tool surface: ModeGeneral or ModeCloudOps.
	// Defaults to ModeCloudOps.
	Mode string `json:"mode"`
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
				Mode:             ModeCloudOps,
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

	// Default to cloud ops mode if not set (covers configs saved before Mode existed).
	if cfg.Mode == "" {
		cfg.Mode = ModeCloudOps
	}

	// Set default endpoint if not set
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://api.openai.com/v1"
	}

	// Normalize endpoint to ensure /v1 suffix
	cfg.Endpoint = normalizeEndpoint(cfg.Endpoint)

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

// isLocalEndpoint checks if the endpoint points to a local LLM server.
func isLocalEndpoint(endpoint string) bool {
	return strings.Contains(endpoint, "localhost") || strings.Contains(endpoint, "127.0.0.1")
}

// normalizeEndpoint ensures the endpoint has the /v1 path suffix
// for known providers. LM Studio, OpenAI, and OpenRouter all require it.
func normalizeEndpoint(endpoint string) string {
	endpoint = strings.TrimSuffix(endpoint, "/")
	if !strings.HasSuffix(endpoint, "/v1") {
		// For local endpoints (LM Studio) and known providers, append /v1
		if isLocalEndpoint(endpoint) ||
			strings.Contains(endpoint, "api.openai.com") ||
			strings.Contains(endpoint, "openrouter.ai") {
			endpoint += "/v1"
		}
	}
	return endpoint
}

// Validate checks if the configuration has all required fields.
// It also normalizes the endpoint URL.
func (c *Config) Validate() error {
	if c.APIKey == "" && !isLocalEndpoint(c.Endpoint) {
		return errors.New("api_key is required")
	}
	if c.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	if c.Model == "" {
		return errors.New("model is required")
	}
	// Normalize endpoint to ensure /v1 suffix
	c.Endpoint = normalizeEndpoint(c.Endpoint)
	return nil
}

// IsConfigured returns true if all required fields are set.
func (c *Config) IsConfigured() bool {
	return (c.APIKey != "" || isLocalEndpoint(c.Endpoint)) &&
		c.Endpoint != "" &&
		c.Model != ""
}
