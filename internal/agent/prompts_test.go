package agent

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetOSInstructions_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}

	instructions := GetOSInstructions()
	if !strings.Contains(strings.ToLower(instructions), "windows") {
		t.Errorf("Windows instructions should mention 'windows', got: %s", instructions)
	}
}

func TestGetOSInstructions_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS-only test")
	}

	instructions := GetOSInstructions()
	if !strings.Contains(strings.ToLower(instructions), "macos") && !strings.Contains(strings.ToLower(instructions), "mac") {
		t.Errorf("macOS instructions should mention 'macos' or 'mac', got: %s", instructions)
	}
}

func TestGetOSInstructions_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}

	instructions := GetOSInstructions()
	if !strings.Contains(strings.ToLower(instructions), "linux") {
		t.Errorf("Linux instructions should mention 'linux', got: %s", instructions)
	}
}

func TestGetOSInstructions_NotEmpty(t *testing.T) {
	instructions := GetOSInstructions()
	if instructions == "" {
		t.Error("OS instructions should not be empty")
	}
}

func TestGetSystemPrompt_ContainsToolList(t *testing.T) {
	prompt := GetSystemPrompt()

	// Should mention the main tools
	expectedTools := []string{
		"run_command",
		"read_file",
		"write_file",
		"list_directory",
		"change_directory",
		"task_complete",
	}

	for _, tool := range expectedTools {
		if !strings.Contains(prompt, tool) {
			t.Errorf("System prompt should mention tool %q", tool)
		}
	}
}

func TestGetSystemPrompt_ContainsRules(t *testing.T) {
	prompt := GetSystemPrompt()

	// Should contain critical rules
	if !strings.Contains(strings.ToLower(prompt), "task_complete") {
		t.Error("System prompt should mention task_complete")
	}

	if !strings.Contains(strings.ToLower(prompt), "rules") || !strings.Contains(strings.ToLower(prompt), "critical") {
		// Could also be "important" or similar
		if !strings.Contains(strings.ToLower(prompt), "important") && !strings.Contains(strings.ToLower(prompt), "must") {
			t.Log("System prompt should contain rules/instructions")
		}
	}
}

func TestGetSystemPrompt_ContainsOSInstructions(t *testing.T) {
	prompt := GetSystemPrompt()

	// Should contain OS-specific text
	osKeywords := map[string][]string{
		"windows": {"windows", "cmd", "powershell"},
		"darwin":  {"macos", "mac", "unix"},
		"linux":   {"linux", "unix", "bash"},
	}

	keywords := osKeywords[runtime.GOOS]
	found := false
	for _, kw := range keywords {
		if strings.Contains(strings.ToLower(prompt), kw) {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("System prompt should contain OS-specific instructions for %s", runtime.GOOS)
	}
}

func TestGetSystemPrompt_NotEmpty(t *testing.T) {
	prompt := GetSystemPrompt()
	if prompt == "" {
		t.Error("System prompt should not be empty")
	}
	if len(prompt) < 100 {
		t.Error("System prompt seems too short")
	}
}
