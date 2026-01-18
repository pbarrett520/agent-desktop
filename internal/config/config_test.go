package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Helper to create a temporary config directory for tests
func setupTestConfigDir(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "agent-desktop-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Override the config directory for testing
	originalConfigDir := configDir
	configDir = tmpDir

	cleanup := func() {
		configDir = originalConfigDir
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestLoadConfig_NotExists_ReturnsDefault(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Should return default values
	if cfg.APIKey != "" {
		t.Errorf("expected empty APIKey, got %q", cfg.APIKey)
	}
	if cfg.Endpoint != "https://api.openai.com/v1" {
		t.Errorf("expected default Endpoint, got %q", cfg.Endpoint)
	}
	if cfg.Model != "" {
		t.Errorf("expected empty Model, got %q", cfg.Model)
	}
	if cfg.ExecutionTimeout != 60 {
		t.Errorf("expected ExecutionTimeout=60, got %d", cfg.ExecutionTimeout)
	}
}

func TestLoadConfig_Exists_ParsesCorrectly(t *testing.T) {
	tmpDir, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Create a config file
	testConfig := Config{
		APIKey:           "sk-test-key-123",
		Endpoint:         "https://api.openai.com/v1",
		Model:            "gpt-4o",
		ExecutionTimeout: 120,
	}

	configPath := filepath.Join(tmpDir, "config.json")
	data, _ := json.MarshalIndent(testConfig, "", "  ")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.APIKey != "sk-test-key-123" {
		t.Errorf("expected APIKey='sk-test-key-123', got %q", cfg.APIKey)
	}
	if cfg.Endpoint != "https://api.openai.com/v1" {
		t.Errorf("expected Endpoint='https://api.openai.com/v1', got %q", cfg.Endpoint)
	}
	if cfg.Model != "gpt-4o" {
		t.Errorf("expected Model='gpt-4o', got %q", cfg.Model)
	}
	if cfg.ExecutionTimeout != 120 {
		t.Errorf("expected ExecutionTimeout=120, got %d", cfg.ExecutionTimeout)
	}
}

func TestLoadConfig_CustomEndpoints(t *testing.T) {
	tmpDir, cleanup := setupTestConfigDir(t)
	defer cleanup()

	tests := []struct {
		name     string
		endpoint string
		model    string
	}{
		{"OpenAI", "https://api.openai.com/v1", "gpt-4o"},
		{"LMStudio", "http://localhost:1234/v1", "local-model"},
		{"OpenRouter", "https://openrouter.ai/api/v1", "anthropic/claude-3-opus"},
		{"Custom", "https://my-custom-api.com/v1", "custom-model"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testConfig := Config{
				APIKey:           "test-key",
				Endpoint:         tt.endpoint,
				Model:            tt.model,
				ExecutionTimeout: 60,
			}

			configPath := filepath.Join(tmpDir, "config.json")
			data, _ := json.MarshalIndent(testConfig, "", "  ")
			if err := os.WriteFile(configPath, data, 0644); err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() returned error: %v", err)
			}

			if cfg.Endpoint != tt.endpoint {
				t.Errorf("expected Endpoint=%q, got %q", tt.endpoint, cfg.Endpoint)
			}
			if cfg.Model != tt.model {
				t.Errorf("expected Model=%q, got %q", tt.model, cfg.Model)
			}
		})
	}
}

func TestSaveConfig_CreatesDirectory(t *testing.T) {
	tmpDir, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Remove the directory to ensure Save creates it
	os.RemoveAll(tmpDir)

	cfg := &Config{
		APIKey:           "key",
		Endpoint:         "https://api.openai.com/v1",
		Model:            "gpt-4o",
		ExecutionTimeout: 60,
	}

	err := cfg.Save()
	if err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Check directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("config directory was not created")
	}

	// Check config file exists
	configPath := filepath.Join(tmpDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}

func TestSaveConfig_WritesValidJSON(t *testing.T) {
	_, cleanup := setupTestConfigDir(t)
	defer cleanup()

	original := &Config{
		APIKey:           "my-secret-key",
		Endpoint:         "https://openrouter.ai/api/v1",
		Model:            "anthropic/claude-3-sonnet",
		ExecutionTimeout: 90,
	}

	err := original.Save()
	if err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Load it back and verify round-trip
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if loaded.APIKey != original.APIKey {
		t.Errorf("round-trip failed for APIKey: got %q, want %q",
			loaded.APIKey, original.APIKey)
	}
	if loaded.Endpoint != original.Endpoint {
		t.Errorf("round-trip failed for Endpoint: got %q, want %q",
			loaded.Endpoint, original.Endpoint)
	}
	if loaded.Model != original.Model {
		t.Errorf("round-trip failed for Model: got %q, want %q",
			loaded.Model, original.Model)
	}
	if loaded.ExecutionTimeout != original.ExecutionTimeout {
		t.Errorf("round-trip failed for ExecutionTimeout: got %d, want %d",
			loaded.ExecutionTimeout, original.ExecutionTimeout)
	}
}

func TestConfig_Validate_AllFieldsRequired(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "empty config",
			config:  Config{},
			wantErr: true,
		},
		{
			name: "missing api key",
			config: Config{
				Endpoint: "https://api.openai.com/v1",
				Model:    "gpt-4o",
			},
			wantErr: true,
		},
		{
			name: "missing endpoint",
			config: Config{
				APIKey: "key",
				Model:  "gpt-4o",
			},
			wantErr: true,
		},
		{
			name: "missing model",
			config: Config{
				APIKey:   "key",
				Endpoint: "https://api.openai.com/v1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Validate_Success(t *testing.T) {
	cfg := Config{
		APIKey:           "key",
		Endpoint:         "https://api.openai.com/v1",
		Model:            "gpt-4o",
		ExecutionTimeout: 60,
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() returned error for valid config: %v", err)
	}
}

func TestConfig_IsConfigured(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   bool
	}{
		{
			name:   "empty config",
			config: Config{},
			want:   false,
		},
		{
			name: "partial config - missing model",
			config: Config{
				APIKey:   "key",
				Endpoint: "https://api.openai.com/v1",
			},
			want: false,
		},
		{
			name: "complete config",
			config: Config{
				APIKey:   "key",
				Endpoint: "https://api.openai.com/v1",
				Model:    "gpt-4o",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsConfigured()
			if got != tt.want {
				t.Errorf("IsConfigured() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigPath(t *testing.T) {
	tmpDir, cleanup := setupTestConfigDir(t)
	defer cleanup()

	expected := filepath.Join(tmpDir, "config.json")
	got := getConfigPath()

	if got != expected {
		t.Errorf("getConfigPath() = %q, want %q", got, expected)
	}
}
