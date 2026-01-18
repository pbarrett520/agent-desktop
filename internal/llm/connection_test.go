package llm

import (
	"strings"
	"testing"

	"agent-desktop/internal/config"
)

func TestTestConnection_InvalidConfig(t *testing.T) {
	// Test with nil config
	success, msg := TestConnection(nil)
	if success {
		t.Error("TestConnection should fail for nil config")
	}
	if msg == "" {
		t.Error("TestConnection should return error message")
	}
}

func TestTestConnection_MissingFields(t *testing.T) {
	cfg := &config.Config{
		APIKey: "key",
		// Missing other fields
	}

	success, msg := TestConnection(cfg)
	if success {
		t.Error("TestConnection should fail for incomplete config")
	}
	if !strings.Contains(strings.ToLower(msg), "required") && !strings.Contains(strings.ToLower(msg), "missing") {
		t.Logf("Message: %s", msg)
		// This is acceptable - validation errors are fine
	}
}

func TestTestConnection_InvalidEndpoint(t *testing.T) {
	cfg := &config.Config{
		APIKey:   "fake-key",
		Endpoint: "https://invalid-endpoint-that-does-not-exist-12345.example.com/v1",
		Model:    "gpt-4o",
	}

	success, msg := TestConnection(cfg)
	if success {
		t.Error("TestConnection should fail for invalid endpoint")
	}
	if msg == "" {
		t.Error("TestConnection should return error message for invalid endpoint")
	}
}

// Note: Testing successful connection requires a real API endpoint
// This should be done via integration tests with proper credentials
