// Package llm provides an OpenAI-compatible client for chat completions with tool calling.
// It supports any OpenAI-compatible endpoint including OpenAI, LM Studio, OpenRouter, etc.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"agent-desktop/internal/config"
	"agent-desktop/internal/tools"
)

// Message represents a chat message.
type Message struct {
	Role       string     `json:"role"` // system, user, assistant, tool
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolCall represents a tool call from the assistant.
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// TokenUsage represents token usage information.
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Response represents a chat completion response.
type Response struct {
	Content   string      `json:"content"`
	ToolCalls []ToolCall  `json:"tool_calls,omitempty"`
	Usage     *TokenUsage `json:"usage,omitempty"`
}

// Client is an OpenAI-compatible API client.
// It works with any endpoint that implements the OpenAI chat completions API:
// - OpenAI (https://api.openai.com/v1)
// - LM Studio (http://localhost:1234/v1)
// - OpenRouter (https://openrouter.ai/api/v1)
// - Any other OpenAI-compatible API
type Client struct {
	httpClient *http.Client
	endpoint   string
	apiKey     string
	model      string
}

// NewClient creates a new OpenAI-compatible client from the given configuration.
func NewClient(cfg *config.Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	endpoint := strings.TrimSuffix(cfg.Endpoint, "/")

	return &Client{
		httpClient: &http.Client{Timeout: 120 * time.Second},
		endpoint:   endpoint,
		apiKey:     cfg.APIKey,
		model:      cfg.Model,
	}, nil
}

// chatRequest is the request body for chat completions.
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Tools    []chatTool    `json:"tools,omitempty"`
}

type chatMessage struct {
	Role       string         `json:"role"`
	Content    string         `json:"content"`
	ToolCalls  []chatToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
}

type chatTool struct {
	Type     string             `json:"type"`
	Function chatToolDefinition `json:"function"`
}

type chatToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type chatToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function chatFunctionCall `json:"function"`
}

type chatFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// chatResponse is the response from chat completions.
type chatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Role      string         `json:"role"`
			Content   string         `json:"content"`
			ToolCalls []chatToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// ChatCompletion sends a chat completion request with optional tool definitions.
func (c *Client) ChatCompletion(ctx context.Context, messages []Message, toolDefs []tools.ToolDefinition) (*Response, error) {
	// Convert messages to API format
	chatMessages := make([]chatMessage, len(messages))
	for i, msg := range messages {
		chatMsg := chatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}

		if msg.ToolCallID != "" {
			chatMsg.ToolCallID = msg.ToolCallID
		}

		if len(msg.ToolCalls) > 0 {
			chatMsg.ToolCalls = make([]chatToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				chatMsg.ToolCalls[j] = chatToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: chatFunctionCall{
						Name:      tc.Name,
						Arguments: tc.Arguments,
					},
				}
			}
		}

		chatMessages[i] = chatMsg
	}

	// Convert tool definitions to API format
	var chatTools []chatTool
	if len(toolDefs) > 0 {
		chatTools = make([]chatTool, len(toolDefs))
		for i, def := range toolDefs {
			chatTools[i] = chatTool{
				Type: "function",
				Function: chatToolDefinition{
					Name:        def.Function.Name,
					Description: def.Function.Description,
					Parameters:  def.Function.Parameters,
				},
			}
		}
	}

	// Build request body
	reqBody := chatRequest{
		Model:    c.model,
		Messages: chatMessages,
	}
	if len(chatTools) > 0 {
		reqBody.Tools = chatTools
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL - standard OpenAI format
	url := fmt.Sprintf("%s/chat/completions", c.endpoint)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error in response
	if chatResp.Error != nil {
		return nil, fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	// Parse response
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := chatResp.Choices[0]
	result := &Response{
		Content: choice.Message.Content,
	}

	// Parse tool calls
	if len(choice.Message.ToolCalls) > 0 {
		result.ToolCalls = make([]ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			result.ToolCalls[i] = ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			}
		}
	}

	// Parse usage
	if chatResp.Usage.TotalTokens > 0 {
		result.Usage = &TokenUsage{
			PromptTokens:     chatResp.Usage.PromptTokens,
			CompletionTokens: chatResp.Usage.CompletionTokens,
			TotalTokens:      chatResp.Usage.TotalTokens,
		}
	}

	return result, nil
}

// GetModel returns the model name.
func (c *Client) GetModel() string {
	return c.model
}

// GetEndpoint returns the endpoint URL.
func (c *Client) GetEndpoint() string {
	return c.endpoint
}
