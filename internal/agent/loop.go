package agent

import (
	"context"
	"encoding/json"
	"strings"

	"agent-desktop/internal/llm"
	"agent-desktop/internal/tools"
)

// Client interface for the LLM client (allows mocking in tests)
type Client interface {
	ChatCompletion(ctx context.Context, messages []llm.Message, toolDefs []tools.ToolDefinition) (*llm.Response, error)
}

// RunLoop runs the agent loop to complete a task.
// It yields Steps through the returned channel.
func RunLoop(ctx context.Context, client Client, task string, taskContext string, maxSteps int) <-chan Step {
	steps := make(chan Step)

	go func() {
		defer close(steps)

		// Reset session for fresh start
		tools.ResetSession()

		// Build initial messages
		messages := []llm.Message{
			{Role: "system", Content: GetSystemPrompt()},
			{Role: "user", Content: BuildUserMessage(task, taskContext)},
		}

		toolDefs := tools.GetToolDefinitions()
		stepNumber := 0
		consecutiveTextResponses := 0
		maxTextResponses := 2

		for stepNumber < maxSteps {
			stepNumber++

			// Check context cancellation
			select {
			case <-ctx.Done():
				steps <- NewErrorStep(stepNumber, "Task cancelled")
				return
			default:
			}

			// Call LLM
			resp, err := client.ChatCompletion(ctx, messages, toolDefs)
			if err != nil {
				steps <- NewErrorStep(stepNumber, "Error: "+err.Error())
				return
			}

			// Emit usage if available
			if resp.Usage != nil {
				steps <- NewUsageStep(stepNumber, &TokenUsage{
					PromptTokens:     resp.Usage.PromptTokens,
					CompletionTokens: resp.Usage.CompletionTokens,
					TotalTokens:      resp.Usage.TotalTokens,
				})
			}

			// Process tool calls if present
			if len(resp.ToolCalls) > 0 {
				consecutiveTextResponses = 0

				// Build assistant message with tool calls
				assistantMsg := llm.Message{
					Role:    "assistant",
					Content: resp.Content,
					ToolCalls: make([]llm.ToolCall, len(resp.ToolCalls)),
				}
				for i, tc := range resp.ToolCalls {
					assistantMsg.ToolCalls[i] = llm.ToolCall{
						ID:        tc.ID,
						Name:      tc.Name,
						Arguments: tc.Arguments,
					}
				}
				messages = append(messages, assistantMsg)

				// If there's thinking content, emit it
				if resp.Content != "" {
					steps <- NewThinkingStep(stepNumber, resp.Content)
				}

				// Process each tool call
				for _, tc := range resp.ToolCalls {
					// Parse tool arguments
					var toolArgs map[string]interface{}
					if err := json.Unmarshal([]byte(tc.Arguments), &toolArgs); err != nil {
						toolArgs = make(map[string]interface{})
					}

					// Emit tool call step
					steps <- NewToolCallStep(stepNumber, tc.Name, toolArgs)

					// Execute the tool
					result := tools.ExecuteTool(tc.Name, toolArgs)

					// Add tool result to messages
					resultContent := result.Output
					if result.Error != "" {
						resultContent += "\n\nError: " + result.Error
					}
					messages = append(messages, llm.Message{
						Role:       "tool",
						Content:    resultContent,
						ToolCallID: tc.ID,
					})

					// Emit tool result step
					steps <- NewToolResultStep(stepNumber, tc.Name, &result)

					// Check if task_complete was called
					if tc.Name == "task_complete" {
						steps <- NewCompleteStep(stepNumber, result.Output)
						return
					}
				}
			} else {
				// No tool calls - model wants to respond with text
				consecutiveTextResponses++

				if resp.Content != "" {
					// Check if this looks like a completion
					content := strings.ToLower(resp.Content)
					isComplete := strings.Contains(content, "completed") ||
						strings.Contains(content, "done") ||
						strings.Contains(content, "finished") ||
						strings.Contains(content, "task complete") ||
						strings.Contains(content, "let me know") ||
						strings.Contains(content, "anything else") ||
						strings.Contains(content, "help you with")

					if isComplete || consecutiveTextResponses >= maxTextResponses {
						steps <- NewCompleteStep(stepNumber, resp.Content)
						return
					}

					// Model wants to say something without tools
					steps <- NewThinkingStep(stepNumber, resp.Content)
					messages = append(messages, llm.Message{
						Role:    "assistant",
						Content: resp.Content,
					})
				} else {
					// Empty response - something went wrong
					steps <- NewErrorStep(stepNumber, "Received empty response from model")
					return
				}
			}
		}

	// Max steps reached
	steps <- NewErrorStep(stepNumber, "Maximum steps reached without completing the task")
	}()

	return steps
}

// ContinueConversation continues an existing conversation with new messages.
// Unlike RunLoop, this function:
// - Does not reset the tools session (session persists across turns)
// - Does not auto-complete based on phrases like "let me know"
// - Only completes when task_complete tool is called
// - Returns assistant_message steps for conversational responses
// - Includes updated messages in step for conversation persistence
func ContinueConversation(ctx context.Context, client Client, messages []llm.Message, maxSteps int) <-chan Step {
	steps := make(chan Step)

	go func() {
		defer close(steps)

		// Make a copy of messages to avoid mutating the input
		msgs := make([]llm.Message, len(messages))
		copy(msgs, messages)

		toolDefs := tools.GetToolDefinitions()
		stepNumber := 0

		for stepNumber < maxSteps {
			stepNumber++

			// Check context cancellation
			select {
			case <-ctx.Done():
				steps <- NewErrorStep(stepNumber, "Task cancelled")
				return
			default:
			}

			// Call LLM
			resp, err := client.ChatCompletion(ctx, msgs, toolDefs)
			if err != nil {
				steps <- NewErrorStep(stepNumber, "Error: "+err.Error())
				return
			}

			// Emit usage if available
			if resp.Usage != nil {
				steps <- NewUsageStep(stepNumber, &TokenUsage{
					PromptTokens:     resp.Usage.PromptTokens,
					CompletionTokens: resp.Usage.CompletionTokens,
					TotalTokens:      resp.Usage.TotalTokens,
				})
			}

			// Process tool calls if present
			if len(resp.ToolCalls) > 0 {
				// Build assistant message with tool calls
				assistantMsg := llm.Message{
					Role:    "assistant",
					Content: resp.Content,
					ToolCalls: make([]llm.ToolCall, len(resp.ToolCalls)),
				}
				for i, tc := range resp.ToolCalls {
					assistantMsg.ToolCalls[i] = llm.ToolCall{
						ID:        tc.ID,
						Name:      tc.Name,
						Arguments: tc.Arguments,
					}
				}
				msgs = append(msgs, assistantMsg)

				// If there's thinking content, emit it
				if resp.Content != "" {
					steps <- NewThinkingStep(stepNumber, resp.Content)
				}

				// Process each tool call
				for _, tc := range resp.ToolCalls {
					// Parse tool arguments
					var toolArgs map[string]interface{}
					if err := json.Unmarshal([]byte(tc.Arguments), &toolArgs); err != nil {
						toolArgs = make(map[string]interface{})
					}

					// Emit tool call step
					steps <- NewToolCallStep(stepNumber, tc.Name, toolArgs)

					// Execute the tool
					result := tools.ExecuteTool(tc.Name, toolArgs)

					// Add tool result to messages
					resultContent := result.Output
					if result.Error != "" {
						resultContent += "\n\nError: " + result.Error
					}
					msgs = append(msgs, llm.Message{
						Role:       "tool",
						Content:    resultContent,
						ToolCallID: tc.ID,
					})

					// Emit tool result step with updated messages
					toolResultStep := NewToolResultStep(stepNumber, tc.Name, &result)
					toolResultStep.Messages = msgs
					steps <- toolResultStep

					// Check if task_complete was called
					if tc.Name == "task_complete" {
						completeStep := NewCompleteStep(stepNumber, result.Output)
						completeStep.Messages = msgs
						steps <- completeStep
						return
					}
				}
			} else {
				// No tool calls - model responded with text
				if resp.Content != "" {
					// Add assistant message to conversation
					msgs = append(msgs, llm.Message{
						Role:    "assistant",
						Content: resp.Content,
					})

					// In conversation mode, text responses are just messages, not completions
					// Return assistant message step with updated messages
					steps <- NewAssistantMessageStep(stepNumber, resp.Content, msgs)
					return
				} else {
					// Empty response
					steps <- NewErrorStep(stepNumber, "Received empty response from model")
					return
				}
			}
		}

		// Max steps reached
		errorStep := NewErrorStep(stepNumber, "Maximum steps reached")
		errorStep.Messages = msgs
		steps <- errorStep
	}()

	return steps
}
