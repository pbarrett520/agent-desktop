package agent

import (
	"testing"

	"agent-desktop/internal/tools"
)

func TestStep_Thinking(t *testing.T) {
	step := Step{
		StepNumber: 1,
		Type:       StepTypeThinking,
		Content:    "I need to analyze the task",
	}

	if step.StepNumber != 1 {
		t.Errorf("StepNumber = %d, want %d", step.StepNumber, 1)
	}
	if step.Type != StepTypeThinking {
		t.Errorf("Type = %q, want %q", step.Type, StepTypeThinking)
	}
	if step.Content != "I need to analyze the task" {
		t.Errorf("Content = %q, want %q", step.Content, "I need to analyze the task")
	}
}

func TestStep_ToolCall(t *testing.T) {
	step := Step{
		StepNumber: 2,
		Type:       StepTypeToolCall,
		Content:    "Calling read_file",
		ToolName:   "read_file",
		ToolArgs: map[string]interface{}{
			"path": "/tmp/test.txt",
		},
	}

	if step.Type != StepTypeToolCall {
		t.Errorf("Type = %q, want %q", step.Type, StepTypeToolCall)
	}
	if step.ToolName != "read_file" {
		t.Errorf("ToolName = %q, want %q", step.ToolName, "read_file")
	}
	if step.ToolArgs["path"] != "/tmp/test.txt" {
		t.Errorf("ToolArgs[path] = %v, want %q", step.ToolArgs["path"], "/tmp/test.txt")
	}
}

func TestStep_ToolResult(t *testing.T) {
	step := Step{
		StepNumber: 3,
		Type:       StepTypeToolResult,
		Content:    "file contents here",
		ToolName:   "read_file",
		ToolResult: &tools.ToolResult{
			Success: true,
			Output:  "file contents here",
		},
	}

	if step.Type != StepTypeToolResult {
		t.Errorf("Type = %q, want %q", step.Type, StepTypeToolResult)
	}
	if step.ToolResult == nil {
		t.Fatal("ToolResult should not be nil")
	}
	if !step.ToolResult.Success {
		t.Error("ToolResult.Success should be true")
	}
}

func TestStep_Complete(t *testing.T) {
	step := Step{
		StepNumber: 4,
		Type:       StepTypeComplete,
		Content:    "Task completed successfully",
	}

	if step.Type != StepTypeComplete {
		t.Errorf("Type = %q, want %q", step.Type, StepTypeComplete)
	}
}

func TestStep_Error(t *testing.T) {
	step := Step{
		StepNumber: 5,
		Type:       StepTypeError,
		Content:    "Something went wrong",
	}

	if step.Type != StepTypeError {
		t.Errorf("Type = %q, want %q", step.Type, StepTypeError)
	}
}

func TestStep_Usage(t *testing.T) {
	step := Step{
		StepNumber: 1,
		Type:       StepTypeUsage,
		Usage: &TokenUsage{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		},
	}

	if step.Type != StepTypeUsage {
		t.Errorf("Type = %q, want %q", step.Type, StepTypeUsage)
	}
	if step.Usage == nil {
		t.Fatal("Usage should not be nil")
	}
	if step.Usage.TotalTokens != 150 {
		t.Errorf("Usage.TotalTokens = %d, want %d", step.Usage.TotalTokens, 150)
	}
}

func TestTokenUsage(t *testing.T) {
	usage := TokenUsage{
		PromptTokens:     200,
		CompletionTokens: 100,
		TotalTokens:      300,
	}

	if usage.PromptTokens != 200 {
		t.Errorf("PromptTokens = %d, want %d", usage.PromptTokens, 200)
	}
	if usage.CompletionTokens != 100 {
		t.Errorf("CompletionTokens = %d, want %d", usage.CompletionTokens, 100)
	}
	if usage.TotalTokens != 300 {
		t.Errorf("TotalTokens = %d, want %d", usage.TotalTokens, 300)
	}
}

func TestStepTypeConstants(t *testing.T) {
	// Verify step type constants are defined
	types := []string{
		StepTypeThinking,
		StepTypeToolCall,
		StepTypeToolResult,
		StepTypeComplete,
		StepTypeError,
		StepTypeUsage,
	}

	for _, stepType := range types {
		if stepType == "" {
			t.Error("Step type constant should not be empty")
		}
	}

	// Verify they're distinct
	seen := make(map[string]bool)
	for _, stepType := range types {
		if seen[stepType] {
			t.Errorf("Duplicate step type: %s", stepType)
		}
		seen[stepType] = true
	}
}
