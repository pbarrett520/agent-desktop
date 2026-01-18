// Package llm provides an Azure OpenAI client for chat completions with tool calling.
package llm

import (
	"context"
	"encoding/json"
	"errors"

	"agent-desktop/internal/config"
	"agent-desktop/internal/tools"

	openai "github.com/sashabaranov/go-openai"
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

// Client is an Azure OpenAI client.
type Client struct {
	client     *openai.Client
	deployment string
	model      string
}

// NewClient creates a new Azure OpenAI client from the given configuration.
func NewClient(cfg *config.Config) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Create Azure OpenAI config
	azureConfig := openai.DefaultAzureConfig(cfg.OpenAISubscriptionKey, cfg.OpenAIEndpoint)
	azureConfig.APIVersion = "2024-02-15-preview"

	client := openai.NewClientWithConfig(azureConfig)

	return &Client{
		client:     client,
		deployment: cfg.OpenAIDeployment,
		model:      cfg.OpenAIModelName,
	}, nil
}

// ChatCompletion sends a chat completion request with optional tool definitions.
func (c *Client) ChatCompletion(ctx context.Context, messages []Message, toolDefs []tools.ToolDefinition) (*Response, error) {
	// Convert messages to OpenAI format
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMsg := openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}

		// Handle tool call ID for tool messages
		if msg.Role == "tool" && msg.ToolCallID != "" {
			openaiMsg.ToolCallID = msg.ToolCallID
		}

		// Handle assistant messages with tool calls
		if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
			openaiMsg.ToolCalls = make([]openai.ToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				openaiMsg.ToolCalls[j] = openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      tc.Name,
						Arguments: tc.Arguments,
					},
				}
			}
		}

		openaiMessages[i] = openaiMsg
	}

	// Convert tool definitions to OpenAI format
	var openaiTools []openai.Tool
	if len(toolDefs) > 0 {
		openaiTools = make([]openai.Tool, len(toolDefs))
		for i, def := range toolDefs {
			// Marshal parameters to JSON bytes for FunctionDefinition
			paramsBytes, err := json.Marshal(def.Function.Parameters)
			if err != nil {
				return nil, err
			}

			openaiTools[i] = openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        def.Function.Name,
					Description: def.Function.Description,
					Parameters:  json.RawMessage(paramsBytes),
				},
			}
		}
	}

	// Build request
	req := openai.ChatCompletionRequest{
		Model:    c.deployment, // Azure uses deployment name as model
		Messages: openaiMessages,
	}

	if len(openaiTools) > 0 {
		req.Tools = openaiTools
	}

	// Make request
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	// Parse response
	if len(resp.Choices) == 0 {
		return nil, errors.New("no choices in response")
	}

	choice := resp.Choices[0]
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
	if resp.Usage.TotalTokens > 0 {
		result.Usage = &TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	return result, nil
}

// GetDeployment returns the Azure deployment name.
func (c *Client) GetDeployment() string {
	return c.deployment
}

// GetModel returns the model name.
func (c *Client) GetModel() string {
	return c.model
}
