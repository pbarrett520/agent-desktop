package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteTool_ValidTool(t *testing.T) {
	// Test get_current_directory which is simple
	result := ExecuteTool("get_current_directory", map[string]interface{}{})

	if !result.Success {
		t.Errorf("ExecuteTool failed: %s", result.Error)
	}
}

func TestExecuteTool_UnknownTool(t *testing.T) {
	result := ExecuteTool("nonexistent_tool", map[string]interface{}{})

	if result.Success {
		t.Error("ExecuteTool should fail for unknown tool")
	}
	if !strings.Contains(strings.ToLower(result.Error), "unknown") {
		t.Errorf("error should mention unknown tool, got: %q", result.Error)
	}
}

func TestExecuteTool_InvalidArgs(t *testing.T) {
	// read_file requires a path argument
	result := ExecuteTool("read_file", map[string]interface{}{})

	if result.Success {
		t.Error("ExecuteTool should fail for missing required args")
	}
}

func TestExecuteTool_ReadFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dispatcher-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("hello"), 0644)

	result := ExecuteTool("read_file", map[string]interface{}{
		"path": testFile,
	})

	if !result.Success {
		t.Errorf("ExecuteTool read_file failed: %s", result.Error)
	}
	if result.Output != "hello" {
		t.Errorf("output = %q, want %q", result.Output, "hello")
	}
}

func TestExecuteTool_WriteFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dispatcher-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "output.txt")

	result := ExecuteTool("write_file", map[string]interface{}{
		"path":    testFile,
		"content": "test content",
	})

	if !result.Success {
		t.Errorf("ExecuteTool write_file failed: %s", result.Error)
	}

	data, _ := os.ReadFile(testFile)
	if string(data) != "test content" {
		t.Errorf("file content = %q, want %q", string(data), "test content")
	}
}

func TestExecuteTool_RunCommand(t *testing.T) {
	ResetSession() // Ensure clean state
	result := ExecuteTool("run_command", map[string]interface{}{
		"command": "echo hello",
	})

	if !result.Success {
		t.Errorf("ExecuteTool run_command failed: %s", result.Error)
	}
	if !strings.Contains(result.Output, "hello") {
		t.Errorf("output should contain 'hello', got: %q", result.Output)
	}
}

func TestExecuteTool_TaskComplete(t *testing.T) {
	result := ExecuteTool("task_complete", map[string]interface{}{
		"summary": "All done!",
	})

	if !result.Success {
		t.Errorf("ExecuteTool task_complete failed: %s", result.Error)
	}
	if !strings.Contains(result.Output, "All done!") {
		t.Errorf("output should contain summary, got: %q", result.Output)
	}
}

func TestGetToolDefinitions(t *testing.T) {
	defs := GetToolDefinitions()

	if len(defs) == 0 {
		t.Error("GetToolDefinitions should return tool definitions")
	}

	// Check that expected tools are present
	expectedTools := []string{
		"run_command",
		"read_file",
		"write_file",
		"list_directory",
		"get_current_directory",
		"change_directory",
		"task_complete",
		"delete_file",
		"copy_file",
		"move_file",
	}

	toolNames := make(map[string]bool)
	for _, def := range defs {
		toolNames[def.Function.Name] = true
	}

	for _, expected := range expectedTools {
		if !toolNames[expected] {
			t.Errorf("missing tool definition: %s", expected)
		}
	}
}

func TestGetToolDefinitions_HasRequiredFields(t *testing.T) {
	defs := GetToolDefinitions()

	for _, def := range defs {
		if def.Type != "function" {
			t.Errorf("tool %s: type should be 'function', got %q", def.Function.Name, def.Type)
		}
		if def.Function.Name == "" {
			t.Error("tool has empty name")
		}
		if def.Function.Description == "" {
			t.Errorf("tool %s: has empty description", def.Function.Name)
		}
	}
}

func TestResetSession_ResetsState(t *testing.T) {
	// Modify the session
	session := GetSession()
	session.CWD = "/some/path"
	session.RecordCommand("test", 0)

	// Reset
	ResetSession()

	// Verify reset
	home, _ := os.UserHomeDir()
	if GetSession().CWD != home {
		t.Errorf("after reset, CWD = %q, want %q", GetSession().CWD, home)
	}
	if len(GetSession().History) != 0 {
		t.Errorf("after reset, history should be empty, got %d items", len(GetSession().History))
	}
}

func TestGetSessionInfo_ReturnsInfo(t *testing.T) {
	ResetSession()

	info := GetSessionInfo()

	if info["cwd"] == nil {
		t.Error("session info should have 'cwd'")
	}
	if info["history_count"] == nil {
		t.Error("session info should have 'history_count'")
	}
}
