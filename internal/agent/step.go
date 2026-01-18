// Package agent provides the agent loop implementation for Agent Desktop.
package agent

import (
	"agent-desktop/internal/llm"
	"agent-desktop/internal/tools"
)

// Step type constants
const (
	StepTypeThinking         = "thinking"
	StepTypeToolCall         = "tool_call"
	StepTypeToolResult       = "tool_result"
	StepTypeComplete         = "complete"
	StepTypeError            = "error"
	StepTypeUsage            = "usage"
	StepTypeAssistantMessage = "assistant_message" // Conversational response (not task completion)
)

// Step represents a single step in the agent's execution.
type Step struct {
	StepNumber int                    `json:"step_number"`
	Type       string                 `json:"type"` // thinking, tool_call, tool_result, complete, error, usage, assistant_message
	Content    string                 `json:"content"`
	ToolName   string                 `json:"tool_name,omitempty"`
	ToolArgs   map[string]interface{} `json:"tool_args,omitempty"`
	ToolResult *tools.ToolResult      `json:"tool_result,omitempty"`
	Usage      *TokenUsage            `json:"usage,omitempty"`
	Messages   []llm.Message          `json:"messages,omitempty"` // Updated conversation messages (for multi-turn)
}

// TokenUsage represents token usage information for a step.
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewThinkingStep creates a new thinking step.
func NewThinkingStep(stepNumber int, content string) Step {
	return Step{
		StepNumber: stepNumber,
		Type:       StepTypeThinking,
		Content:    content,
	}
}

// NewToolCallStep creates a new tool call step.
func NewToolCallStep(stepNumber int, toolName string, toolArgs map[string]interface{}) Step {
	return Step{
		StepNumber: stepNumber,
		Type:       StepTypeToolCall,
		Content:    "Calling " + toolName,
		ToolName:   toolName,
		ToolArgs:   toolArgs,
	}
}

// NewToolResultStep creates a new tool result step.
func NewToolResultStep(stepNumber int, toolName string, result *tools.ToolResult) Step {
	content := result.Output
	if result.Error != "" {
		if content != "" {
			content += "\n\nError: " + result.Error
		} else {
			content = "Error: " + result.Error
		}
	}

	return Step{
		StepNumber: stepNumber,
		Type:       StepTypeToolResult,
		Content:    content,
		ToolName:   toolName,
		ToolResult: result,
	}
}

// NewCompleteStep creates a new completion step.
func NewCompleteStep(stepNumber int, content string) Step {
	return Step{
		StepNumber: stepNumber,
		Type:       StepTypeComplete,
		Content:    content,
	}
}

// NewErrorStep creates a new error step.
func NewErrorStep(stepNumber int, content string) Step {
	return Step{
		StepNumber: stepNumber,
		Type:       StepTypeError,
		Content:    content,
	}
}

// NewUsageStep creates a new usage step.
func NewUsageStep(stepNumber int, usage *TokenUsage) Step {
	return Step{
		StepNumber: stepNumber,
		Type:       StepTypeUsage,
		Content:    "",
		Usage:      usage,
	}
}

// NewAssistantMessageStep creates a step for a conversational assistant response.
// This is used in multi-turn conversations where the assistant responds without
// completing a task. It includes the updated messages for the conversation.
func NewAssistantMessageStep(stepNumber int, content string, messages []llm.Message) Step {
	return Step{
		StepNumber: stepNumber,
		Type:       StepTypeAssistantMessage,
		Content:    content,
		Messages:   messages,
	}
}
