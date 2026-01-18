package llm

import (
	"testing"

	"agent-desktop/internal/config"
)

func TestNewClient_ValidConfig(t *testing.T) {
	cfg := &config.Config{
		APIKey:   "sk-test-key",
		Endpoint: "https://api.openai.com/v1",
		Model:    "gpt-4o",
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.model != "gpt-4o" {
		t.Errorf("model = %q, want %q", client.model, "gpt-4o")
	}
	if client.endpoint != "https://api.openai.com/v1" {
		t.Errorf("endpoint = %q, want %q", client.endpoint, "https://api.openai.com/v1")
	}
}

func TestNewClient_DifferentEndpoints(t *testing.T) {
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
			cfg := &config.Config{
				APIKey:   "test-key",
				Endpoint: tt.endpoint,
				Model:    tt.model,
			}

			client, err := NewClient(cfg)
			if err != nil {
				t.Fatalf("NewClient failed: %v", err)
			}
			if client.endpoint != tt.endpoint {
				t.Errorf("endpoint = %q, want %q", client.endpoint, tt.endpoint)
			}
			if client.model != tt.model {
				t.Errorf("model = %q, want %q", client.model, tt.model)
			}
		})
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
				APIKey: "key",
				Model:  "gpt-4o",
			},
		},
		{
			name: "missing key",
			config: &config.Config{
				Endpoint: "https://api.openai.com/v1",
				Model:    "gpt-4o",
			},
		},
		{
			name: "missing model",
			config: &config.Config{
				APIKey:   "key",
				Endpoint: "https://api.openai.com/v1",
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

func TestClient_GetModel(t *testing.T) {
	cfg := &config.Config{
		APIKey:   "test-key",
		Endpoint: "https://api.openai.com/v1",
		Model:    "gpt-4o",
	}

	client, _ := NewClient(cfg)
	if client.GetModel() != "gpt-4o" {
		t.Errorf("GetModel() = %q, want %q", client.GetModel(), "gpt-4o")
	}
}

func TestClient_GetEndpoint(t *testing.T) {
	cfg := &config.Config{
		APIKey:   "test-key",
		Endpoint: "https://api.openai.com/v1",
		Model:    "gpt-4o",
	}

	client, _ := NewClient(cfg)
	if client.GetEndpoint() != "https://api.openai.com/v1" {
		t.Errorf("GetEndpoint() = %q, want %q", client.GetEndpoint(), "https://api.openai.com/v1")
	}
}

// Note: Actual API call tests would require mocking or integration test setup
// The ChatCompletion method will be tested via integration tests with a real endpoint
