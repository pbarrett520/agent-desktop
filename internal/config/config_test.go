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
	if cfg.OpenAISubscriptionKey != "" {
		t.Errorf("expected empty OpenAISubscriptionKey, got %q", cfg.OpenAISubscriptionKey)
	}
	if cfg.OpenAIEndpoint != "" {
		t.Errorf("expected empty OpenAIEndpoint, got %q", cfg.OpenAIEndpoint)
	}
	if cfg.OpenAIDeployment != "" {
		t.Errorf("expected empty OpenAIDeployment, got %q", cfg.OpenAIDeployment)
	}
	if cfg.OpenAIModelName != "" {
		t.Errorf("expected empty OpenAIModelName, got %q", cfg.OpenAIModelName)
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
		OpenAISubscriptionKey: "test-key-123",
		OpenAIEndpoint:        "https://test.openai.azure.com",
		OpenAIDeployment:      "gpt-4o-deployment",
		OpenAIModelName:       "gpt-4o",
		ExecutionTimeout:      120,
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

	if cfg.OpenAISubscriptionKey != "test-key-123" {
		t.Errorf("expected OpenAISubscriptionKey='test-key-123', got %q", cfg.OpenAISubscriptionKey)
	}
	if cfg.OpenAIEndpoint != "https://test.openai.azure.com" {
		t.Errorf("expected OpenAIEndpoint='https://test.openai.azure.com', got %q", cfg.OpenAIEndpoint)
	}
	if cfg.OpenAIDeployment != "gpt-4o-deployment" {
		t.Errorf("expected OpenAIDeployment='gpt-4o-deployment', got %q", cfg.OpenAIDeployment)
	}
	if cfg.OpenAIModelName != "gpt-4o" {
		t.Errorf("expected OpenAIModelName='gpt-4o', got %q", cfg.OpenAIModelName)
	}
	if cfg.ExecutionTimeout != 120 {
		t.Errorf("expected ExecutionTimeout=120, got %d", cfg.ExecutionTimeout)
	}
}

func TestSaveConfig_CreatesDirectory(t *testing.T) {
	tmpDir, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Remove the directory to ensure Save creates it
	os.RemoveAll(tmpDir)

	cfg := &Config{
		OpenAISubscriptionKey: "key",
		OpenAIEndpoint:        "https://test.openai.azure.com",
		OpenAIDeployment:      "deploy",
		OpenAIModelName:       "model",
		ExecutionTimeout:      60,
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
		OpenAISubscriptionKey: "my-secret-key",
		OpenAIEndpoint:        "https://myresource.openai.azure.com",
		OpenAIDeployment:      "gpt-4o",
		OpenAIModelName:       "gpt-4o",
		ExecutionTimeout:      90,
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

	if loaded.OpenAISubscriptionKey != original.OpenAISubscriptionKey {
		t.Errorf("round-trip failed for OpenAISubscriptionKey: got %q, want %q",
			loaded.OpenAISubscriptionKey, original.OpenAISubscriptionKey)
	}
	if loaded.OpenAIEndpoint != original.OpenAIEndpoint {
		t.Errorf("round-trip failed for OpenAIEndpoint: got %q, want %q",
			loaded.OpenAIEndpoint, original.OpenAIEndpoint)
	}
	if loaded.OpenAIDeployment != original.OpenAIDeployment {
		t.Errorf("round-trip failed for OpenAIDeployment: got %q, want %q",
			loaded.OpenAIDeployment, original.OpenAIDeployment)
	}
	if loaded.OpenAIModelName != original.OpenAIModelName {
		t.Errorf("round-trip failed for OpenAIModelName: got %q, want %q",
			loaded.OpenAIModelName, original.OpenAIModelName)
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
			name: "missing subscription key",
			config: Config{
				OpenAIEndpoint:   "https://test.openai.azure.com",
				OpenAIDeployment: "deploy",
				OpenAIModelName:  "model",
			},
			wantErr: true,
		},
		{
			name: "missing endpoint",
			config: Config{
				OpenAISubscriptionKey: "key",
				OpenAIDeployment:      "deploy",
				OpenAIModelName:       "model",
			},
			wantErr: true,
		},
		{
			name: "missing deployment",
			config: Config{
				OpenAISubscriptionKey: "key",
				OpenAIEndpoint:        "https://test.openai.azure.com",
				OpenAIModelName:       "model",
			},
			wantErr: true,
		},
		{
			name: "missing model name",
			config: Config{
				OpenAISubscriptionKey: "key",
				OpenAIEndpoint:        "https://test.openai.azure.com",
				OpenAIDeployment:      "deploy",
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
		OpenAISubscriptionKey: "key",
		OpenAIEndpoint:        "https://test.openai.azure.com",
		OpenAIDeployment:      "deploy",
		OpenAIModelName:       "model",
		ExecutionTimeout:      60,
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
			name: "partial config",
			config: Config{
				OpenAISubscriptionKey: "key",
				OpenAIEndpoint:        "https://test.openai.azure.com",
			},
			want: false,
		},
		{
			name: "complete config",
			config: Config{
				OpenAISubscriptionKey: "key",
				OpenAIEndpoint:        "https://test.openai.azure.com",
				OpenAIDeployment:      "deploy",
				OpenAIModelName:       "model",
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
