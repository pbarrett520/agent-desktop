package llm

import (
	"testing"

	"agent-desktop/internal/config"
)

func TestNewClient_ValidConfig(t *testing.T) {
	cfg := &config.Config{
		OpenAISubscriptionKey: "test-key",
		OpenAIEndpoint:        "https://test.openai.azure.com",
		OpenAIDeployment:      "gpt-4o",
		OpenAIModelName:       "gpt-4o",
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.deployment != "gpt-4o" {
		t.Errorf("deployment = %q, want %q", client.deployment, "gpt-4o")
	}
	if client.model != "gpt-4o" {
		t.Errorf("model = %q, want %q", client.model, "gpt-4o")
	}
}

func TestNewClient_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name: "missing endpoint",
			config: &config.Config{
				OpenAISubscriptionKey: "key",
				OpenAIDeployment:      "deploy",
				OpenAIModelName:       "model",
			},
		},
		{
			name: "missing key",
			config: &config.Config{
				OpenAIEndpoint:   "https://test.openai.azure.com",
				OpenAIDeployment: "deploy",
				OpenAIModelName:  "model",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.config)
			if err == nil {
				t.Error("NewClient should fail for invalid config")
			}
		})
	}
}

func TestMessage_Types(t *testing.T) {
	// Test creating different message types
	systemMsg := Message{Role: "system", Content: "You are helpful"}
	userMsg := Message{Role: "user", Content: "Hello"}
	assistantMsg := Message{Role: "assistant", Content: "Hi there"}
	toolMsg := Message{Role: "tool", Content: "result", ToolCallID: "call_123"}

	if systemMsg.Role != "system" {
		t.Errorf("systemMsg.Role = %q, want %q", systemMsg.Role, "system")
	}
	if userMsg.Role != "user" {
		t.Errorf("userMsg.Role = %q, want %q", userMsg.Role, "user")
	}
	if assistantMsg.Role != "assistant" {
		t.Errorf("assistantMsg.Role = %q, want %q", assistantMsg.Role, "assistant")
	}
	if toolMsg.Role != "tool" {
		t.Errorf("toolMsg.Role = %q, want %q", toolMsg.Role, "tool")
	}
	if toolMsg.ToolCallID != "call_123" {
		t.Errorf("toolMsg.ToolCallID = %q, want %q", toolMsg.ToolCallID, "call_123")
	}
}

func TestToolCall(t *testing.T) {
	tc := ToolCall{
		ID:        "call_abc123",
		Name:      "read_file",
		Arguments: `{"path": "/tmp/test.txt"}`,
	}

	if tc.ID != "call_abc123" {
		t.Errorf("ToolCall.ID = %q, want %q", tc.ID, "call_abc123")
	}
	if tc.Name != "read_file" {
		t.Errorf("ToolCall.Name = %q, want %q", tc.Name, "read_file")
	}
}

func TestTokenUsage(t *testing.T) {
	usage := TokenUsage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}

	if usage.PromptTokens != 100 {
		t.Errorf("PromptTokens = %d, want %d", usage.PromptTokens, 100)
	}
	if usage.CompletionTokens != 50 {
		t.Errorf("CompletionTokens = %d, want %d", usage.CompletionTokens, 50)
	}
	if usage.TotalTokens != 150 {
		t.Errorf("TotalTokens = %d, want %d", usage.TotalTokens, 150)
	}
}

func TestResponse(t *testing.T) {
	resp := Response{
		Content: "Hello!",
		ToolCalls: []ToolCall{
			{ID: "call_1", Name: "test_tool", Arguments: "{}"},
		},
		Usage: &TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	if resp.Content != "Hello!" {
		t.Errorf("Response.Content = %q, want %q", resp.Content, "Hello!")
	}
	if len(resp.ToolCalls) != 1 {
		t.Errorf("len(Response.ToolCalls) = %d, want %d", len(resp.ToolCalls), 1)
	}
	if resp.Usage == nil {
		t.Error("Response.Usage should not be nil")
	}
}

// Note: Actual API call tests would require mocking or integration test setup
// The ChatCompletion method will be tested via integration tests with a real Azure endpoint
