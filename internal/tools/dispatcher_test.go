package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agent-desktop/internal/config"
)

func TestExecuteTool_ValidTool(t *testing.T) {
	// Test get_current_directory which is simple
	result := ExecuteTool("get_current_directory", map[string]interface{}{}, config.ModeGeneral)

	if !result.Success {
		t.Errorf("ExecuteTool failed: %s", result.Error)
	}
}

func TestExecuteTool_UnknownTool(t *testing.T) {
	result := ExecuteTool("nonexistent_tool", map[string]interface{}{}, config.ModeGeneral)

	if result.Success {
		t.Error("ExecuteTool should fail for unknown tool")
	}
	if !strings.Contains(strings.ToLower(result.Error), "unknown") {
		t.Errorf("error should mention unknown tool, got: %q", result.Error)
	}
}

func TestExecuteTool_InvalidArgs(t *testing.T) {
	// read_file requires a path argument
	result := ExecuteTool("read_file", map[string]interface{}{}, config.ModeGeneral)

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
	}, config.ModeGeneral)

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
	}, config.ModeGeneral)

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
	}, config.ModeGeneral)

	if !result.Success {
		t.Errorf("ExecuteTool run_command failed: %s", result.Error)
	}
	if !strings.Contains(result.Output, "hello") {
		t.Errorf("output should contain 'hello', got: %q", result.Output)
	}
}

func TestExecuteTool_RunCommand_RefusedInCloudOpsMode(t *testing.T) {
	result := ExecuteTool("run_command", map[string]interface{}{
		"command": "echo hello",
	}, config.ModeCloudOps)

	if result.Success {
		t.Error("run_command should be refused in cloud ops mode")
	}
}

func TestExecuteTool_DeleteFile_RefusedInCloudOpsMode(t *testing.T) {
	result := ExecuteTool("delete_file", map[string]interface{}{
		"path":    "/tmp/whatever",
		"confirm": true,
	}, config.ModeCloudOps)

	if result.Success {
		t.Error("delete_file should be refused in cloud ops mode")
	}
}

func TestExecuteTool_AzQuery_RefusedInGeneralMode(t *testing.T) {
	result := ExecuteTool("az_query", map[string]interface{}{
		"command": "az vm list",
	}, config.ModeGeneral)

	if result.Success {
		t.Error("az_query should be refused in general mode")
	}
}

func TestExecuteTool_TaskComplete(t *testing.T) {
	result := ExecuteTool("task_complete", map[string]interface{}{
		"summary": "All done!",
	}, config.ModeGeneral)

	if !result.Success {
		t.Errorf("ExecuteTool task_complete failed: %s", result.Error)
	}
	if !strings.Contains(result.Output, "All done!") {
		t.Errorf("output should contain summary, got: %q", result.Output)
	}
}

func TestGetToolDefinitions_GeneralMode(t *testing.T) {
	defs := GetToolDefinitions(config.ModeGeneral)

	if len(defs) == 0 {
		t.Error("GetToolDefinitions should return tool definitions")
	}

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

	for _, forbidden := range []string{"az_query", "az_propose"} {
		if toolNames[forbidden] {
			t.Errorf("general mode should not expose %s", forbidden)
		}
	}
}

func TestGetToolDefinitions_CloudOpsMode(t *testing.T) {
	defs := GetToolDefinitions(config.ModeCloudOps)

	toolNames := make(map[string]bool)
	for _, def := range defs {
		toolNames[def.Function.Name] = true
	}

	for _, expected := range []string{"az_query", "az_propose", "read_file", "write_file", "list_directory", "copy_file", "move_file", "task_complete"} {
		if !toolNames[expected] {
			t.Errorf("cloud ops mode missing tool definition: %s", expected)
		}
	}

	for _, forbidden := range []string{"run_command", "delete_file"} {
		if toolNames[forbidden] {
			t.Errorf("cloud ops mode should not expose %s", forbidden)
		}
	}
}

func TestGetToolDefinitions_HasRequiredFields(t *testing.T) {
	defs := GetToolDefinitions(config.ModeGeneral)

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
