package agent

import (
	"context"
	"strings"
	"testing"

	"agent-desktop/internal/llm"
	"agent-desktop/internal/tools"
)

// mockClient is a mock LLM client for testing
type mockClient struct {
	responses []mockResponse
	callCount int
}

type mockResponse struct {
	content   string
	toolCalls []llm.ToolCall
	err       error
}

func (m *mockClient) ChatCompletion(ctx context.Context, messages []llm.Message, toolDefs []tools.ToolDefinition) (*llm.Response, error) {
	if m.callCount >= len(m.responses) {
		return &llm.Response{Content: "Done"}, nil
	}
	resp := m.responses[m.callCount]
	m.callCount++
	if resp.err != nil {
		return nil, resp.err
	}
	return &llm.Response{
		Content:   resp.content,
		ToolCalls: resp.toolCalls,
		Usage: &llm.TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}, nil
}

func TestRunLoop_TaskComplete(t *testing.T) {
	// Mock client that calls task_complete on first turn
	client := &mockClient{
		responses: []mockResponse{
			{
				toolCalls: []llm.ToolCall{
					{
						ID:        "call_1",
						Name:      "task_complete",
						Arguments: `{"summary": "Task done"}`,
					},
				},
			},
		},
	}

	tools.ResetSession()
	ctx := context.Background()

	var steps []Step
	for step := range RunLoop(ctx, client, "Do something", "", 20) {
		steps = append(steps, step)
	}

	// Should have usage step, tool call, tool result, and complete step
	hasComplete := false
	for _, step := range steps {
		if step.Type == StepTypeComplete {
			hasComplete = true
		}
	}

	if !hasComplete {
		t.Error("RunLoop should emit a complete step when task_complete is called")
	}
}

func TestRunLoop_MaxSteps(t *testing.T) {
	// Mock client that keeps calling tools but never task_complete
	client := &mockClient{
		responses: []mockResponse{
			{
				toolCalls: []llm.ToolCall{
					{ID: "call_1", Name: "get_current_directory", Arguments: `{}`},
				},
			},
			{
				toolCalls: []llm.ToolCall{
					{ID: "call_2", Name: "get_current_directory", Arguments: `{}`},
				},
			},
			{
				toolCalls: []llm.ToolCall{
					{ID: "call_3", Name: "get_current_directory", Arguments: `{}`},
				},
			},
			{
				toolCalls: []llm.ToolCall{
					{ID: "call_4", Name: "get_current_directory", Arguments: `{}`},
				},
			},
		},
	}

	tools.ResetSession()
	ctx := context.Background()

	var steps []Step
	maxSteps := 3
	for step := range RunLoop(ctx, client, "Do something", "", maxSteps) {
		steps = append(steps, step)
	}

	// Should have an error step about max steps
	hasMaxStepsError := false
	for _, step := range steps {
		if step.Type == StepTypeError && strings.Contains(step.Content, "Maximum") {
			hasMaxStepsError = true
		}
	}

	if !hasMaxStepsError {
		t.Error("RunLoop should emit error when max steps reached")
	}
}

func TestRunLoop_EmitsUsage(t *testing.T) {
	client := &mockClient{
		responses: []mockResponse{
			{
				toolCalls: []llm.ToolCall{
					{
						ID:        "call_1",
						Name:      "task_complete",
						Arguments: `{"summary": "Done"}`,
					},
				},
			},
		},
	}

	tools.ResetSession()
	ctx := context.Background()

	hasUsage := false
	for step := range RunLoop(ctx, client, "test", "", 20) {
		if step.Type == StepTypeUsage && step.Usage != nil {
			hasUsage = true
		}
	}

	if !hasUsage {
		t.Error("RunLoop should emit usage steps")
	}
}

func TestRunLoop_ToolExecution(t *testing.T) {
	// Mock client that calls get_current_directory then completes
	client := &mockClient{
		responses: []mockResponse{
			{
				toolCalls: []llm.ToolCall{
					{
						ID:        "call_1",
						Name:      "get_current_directory",
						Arguments: `{}`,
					},
				},
			},
			{
				toolCalls: []llm.ToolCall{
					{
						ID:        "call_2",
						Name:      "task_complete",
						Arguments: `{"summary": "Got the directory"}`,
					},
				},
			},
		},
	}

	tools.ResetSession()
	ctx := context.Background()

	var steps []Step
	for step := range RunLoop(ctx, client, "Get current directory", "", 20) {
		steps = append(steps, step)
	}

	// Should have tool call and tool result steps
	hasToolCall := false
	hasToolResult := false
	for _, step := range steps {
		if step.Type == StepTypeToolCall && step.ToolName == "get_current_directory" {
			hasToolCall = true
		}
		if step.Type == StepTypeToolResult && step.ToolName == "get_current_directory" {
			hasToolResult = true
		}
	}

	if !hasToolCall {
		t.Error("RunLoop should emit tool call step")
	}
	if !hasToolResult {
		t.Error("RunLoop should emit tool result step")
	}
}

func TestRunLoop_ContextCancellation(t *testing.T) {
	client := &mockClient{
		responses: []mockResponse{
			{content: "thinking..."},
			{content: "thinking..."},
			{content: "thinking..."},
		},
	}

	tools.ResetSession()
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	var steps []Step
	for step := range RunLoop(ctx, client, "test", "", 20) {
		steps = append(steps, step)
	}

	// Should exit quickly due to cancellation
	if len(steps) > 5 {
		t.Errorf("RunLoop should exit quickly on context cancellation, got %d steps", len(steps))
	}
}

// ============================================================================
// ContinueConversation Tests
// ============================================================================

func TestContinueConversation_WithExistingMessages(t *testing.T) {
	// Mock client that responds to the continuation
	client := &mockClient{
		responses: []mockResponse{
			{
				toolCalls: []llm.ToolCall{
					{
						ID:        "call_1",
						Name:      "task_complete",
						Arguments: `{"summary": "Continuing..."}`,
					},
				},
			},
		},
	}

	tools.ResetSession()
	ctx := context.Background()

	// Existing conversation history
	existingMessages := []llm.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
		{Role: "user", Content: "What time is it?"}, // New message
	}

	var steps []Step
	for step := range ContinueConversation(ctx, client, existingMessages, 20) {
		steps = append(steps, step)
	}

	// Should complete
	hasComplete := false
	for _, step := range steps {
		if step.Type == StepTypeComplete {
			hasComplete = true
		}
	}

	if !hasComplete {
		t.Error("ContinueConversation should emit a complete step")
	}
}

func TestContinueConversation_ReturnsAssistantMessage(t *testing.T) {
	// Mock client that gives a text response (no tools)
	client := &mockClient{
		responses: []mockResponse{
			{content: "Here's my response to your question."},
		},
	}

	tools.ResetSession()
	ctx := context.Background()

	existingMessages := []llm.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Tell me a joke"},
	}

	var steps []Step
	for step := range ContinueConversation(ctx, client, existingMessages, 20) {
		steps = append(steps, step)
	}

	// Should have an assistant message step
	hasAssistantMessage := false
	for _, step := range steps {
		if step.Type == StepTypeAssistantMessage {
			hasAssistantMessage = true
			if step.Content != "Here's my response to your question." {
				t.Errorf("Unexpected content: %s", step.Content)
			}
		}
	}

	if !hasAssistantMessage {
		t.Error("ContinueConversation should emit assistant message step for text responses")
	}
}

func TestContinueConversation_ReturnsUpdatedMessages(t *testing.T) {
	// Mock client that responds with a tool call then completes
	client := &mockClient{
		responses: []mockResponse{
			{
				toolCalls: []llm.ToolCall{
					{ID: "call_1", Name: "get_current_directory", Arguments: `{}`},
				},
			},
			{
				toolCalls: []llm.ToolCall{
					{ID: "call_2", Name: "task_complete", Arguments: `{"summary": "Done"}`},
				},
			},
		},
	}

	tools.ResetSession()
	ctx := context.Background()

	existingMessages := []llm.Message{
		{Role: "system", Content: "You are helpful."},
		{Role: "user", Content: "Get my directory"},
	}

	var finalMessages []llm.Message
	for step := range ContinueConversation(ctx, client, existingMessages, 20) {
		if step.Messages != nil {
			finalMessages = step.Messages
		}
	}

	// Should have more messages than we started with
	if len(finalMessages) <= len(existingMessages) {
		t.Errorf("Expected more messages, got %d (started with %d)", len(finalMessages), len(existingMessages))
	}
}

func TestContinueConversation_DoesNotAutoComplete(t *testing.T) {
	// Mock client that says "let me know" - should NOT auto-complete in conversation mode
	client := &mockClient{
		responses: []mockResponse{
			{content: "I'm done! Let me know if you need anything else."},
		},
	}

	tools.ResetSession()
	ctx := context.Background()

	existingMessages := []llm.Message{
		{Role: "system", Content: "You are helpful."},
		{Role: "user", Content: "Help me"},
	}

	var steps []Step
	for step := range ContinueConversation(ctx, client, existingMessages, 20) {
		steps = append(steps, step)
	}

	// Should NOT have a "complete" step - that's only for task_complete tool
	// Instead should have assistant_message step
	hasAssistantMessage := false
	hasComplete := false
	for _, step := range steps {
		if step.Type == StepTypeAssistantMessage {
			hasAssistantMessage = true
		}
		if step.Type == StepTypeComplete {
			hasComplete = true
		}
	}

	if !hasAssistantMessage {
		t.Error("Should emit assistant_message step")
	}
	if hasComplete {
		t.Error("Should NOT auto-complete based on phrases like 'let me know' in conversation mode")
	}
}

func TestContinueConversation_ToolCallsWork(t *testing.T) {
	// Mock client that executes tools
	client := &mockClient{
		responses: []mockResponse{
			{
				content: "Let me check that for you.",
				toolCalls: []llm.ToolCall{
					{ID: "call_1", Name: "get_current_directory", Arguments: `{}`},
				},
			},
			{content: "Your current directory is shown above."},
		},
	}

	tools.ResetSession()
	ctx := context.Background()

	existingMessages := []llm.Message{
		{Role: "system", Content: "You are helpful."},
		{Role: "user", Content: "What directory am I in?"},
	}

	var steps []Step
	for step := range ContinueConversation(ctx, client, existingMessages, 20) {
		steps = append(steps, step)
	}

	hasToolCall := false
	hasToolResult := false
	for _, step := range steps {
		if step.Type == StepTypeToolCall {
			hasToolCall = true
		}
		if step.Type == StepTypeToolResult {
			hasToolResult = true
		}
	}

	if !hasToolCall {
		t.Error("Should emit tool_call step")
	}
	if !hasToolResult {
		t.Error("Should emit tool_result step")
	}
}
