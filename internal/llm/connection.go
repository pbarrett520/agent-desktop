package llm

import (
	"context"
	"time"

	"agent-desktop/internal/config"
)

// TestConnection tests the LLM connection by making a minimal API call.
// Returns (true, "success message") on success, (false, "error message") on failure.
func TestConnection(cfg *config.Config) (bool, string) {
	if cfg == nil {
		return false, "Configuration is nil"
	}

	// Validate config first
	if err := cfg.Validate(); err != nil {
		return false, err.Error()
	}

	// Create client
	client, err := NewClient(cfg)
	if err != nil {
		return false, "Failed to create client: " + err.Error()
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Make a minimal chat completion request
	messages := []Message{
		{Role: "user", Content: "Hi"},
	}

	_, err = client.ChatCompletion(ctx, messages, nil)
	if err != nil {
		return false, "Connection failed: " + err.Error()
	}

	return true, "Connected successfully to " + cfg.Endpoint + "!"
}
