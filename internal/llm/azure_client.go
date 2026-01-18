// Package llm provides an Azure OpenAI client for chat completions with tool calling.
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

// AzureClient is a custom Azure OpenAI client that uses direct HTTP requests.
// This is needed because the go-openai library has issues with certain Azure endpoints.
type AzureClient struct {
	httpClient *http.Client
	endpoint   string
	apiKey     string
	deployment string
	model      string
	apiVersion string
}

// NewAzureClient creates a new Azure OpenAI client from the given configuration.
func NewAzureClient(cfg *config.Config) (*AzureClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	endpoint := strings.TrimSuffix(cfg.OpenAIEndpoint, "/")

	return &AzureClient{
		httpClient: &http.Client{Timeout: 120 * time.Second},
		endpoint:   endpoint,
		apiKey:     cfg.OpenAISubscriptionKey,
		deployment: cfg.OpenAIDeployment,
		model:      cfg.OpenAIModelName,
		apiVersion: "2024-10-21",
	}, nil
}

// azureChatRequest is the request body for Azure OpenAI chat completions.
type azureChatRequest struct {
	Messages []azureChatMessage `json:"messages"`
	Tools    []azureTool        `json:"tools,omitempty"`
}

type azureChatMessage struct {
	Role       string          `json:"role"`
	Content    string          `json:"content"`
	ToolCalls  []azureToolCall `json:"tool_calls,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
}

type azureTool struct {
	Type     string              `json:"type"`
	Function azureToolDefinition `json:"function"`
}

type azureToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type azureToolCall struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Function azureFunctionCall `json:"function"`
}

type azureFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// azureChatResponse is the response from Azure OpenAI chat completions.
type azureChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Role      string          `json:"role"`
			Content   string          `json:"content"`
			ToolCalls []azureToolCall `json:"tool_calls,omitempty"`
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
func (c *AzureClient) ChatCompletion(ctx context.Context, messages []Message, toolDefs []tools.ToolDefinition) (*Response, error) {
	// Convert messages to Azure format
	azureMessages := make([]azureChatMessage, len(messages))
	for i, msg := range messages {
		azureMsg := azureChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}

		if msg.ToolCallID != "" {
			azureMsg.ToolCallID = msg.ToolCallID
		}

		if len(msg.ToolCalls) > 0 {
			azureMsg.ToolCalls = make([]azureToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				azureMsg.ToolCalls[j] = azureToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: azureFunctionCall{
						Name:      tc.Name,
						Arguments: tc.Arguments,
					},
				}
			}
		}

		azureMessages[i] = azureMsg
	}

	// Convert tool definitions to Azure format
	var azureTools []azureTool
	if len(toolDefs) > 0 {
		azureTools = make([]azureTool, len(toolDefs))
		for i, def := range toolDefs {
			azureTools[i] = azureTool{
				Type: "function",
				Function: azureToolDefinition{
					Name:        def.Function.Name,
					Description: def.Function.Description,
					Parameters:  def.Function.Parameters,
				},
			}
		}
	}

	// Build request body
	reqBody := azureChatRequest{
		Messages: azureMessages,
	}
	if len(azureTools) > 0 {
		reqBody.Tools = azureTools
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL
	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s",
		c.endpoint, c.deployment, c.apiVersion)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.apiKey)

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
	var azureResp azureChatResponse
	if err := json.Unmarshal(respBody, &azureResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error in response
	if azureResp.Error != nil {
		return nil, fmt.Errorf("API error: %s", azureResp.Error.Message)
	}

	// Parse response
	if len(azureResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := azureResp.Choices[0]
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
	if azureResp.Usage.TotalTokens > 0 {
		result.Usage = &TokenUsage{
			PromptTokens:     azureResp.Usage.PromptTokens,
			CompletionTokens: azureResp.Usage.CompletionTokens,
			TotalTokens:      azureResp.Usage.TotalTokens,
		}
	}

	return result, nil
}

// GetDeployment returns the Azure deployment name.
func (c *AzureClient) GetDeployment() string {
	return c.deployment
}

// GetModel returns the model name.
func (c *AzureClient) GetModel() string {
	return c.model
}
