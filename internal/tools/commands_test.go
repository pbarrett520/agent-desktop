package tools

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunCommand_Success(t *testing.T) {
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "echo hello"
	} else {
		cmd = "echo hello"
	}

	result := RunCommand(cmd, "", 30)

	if !result.Success {
		t.Errorf("RunCommand failed: %s", result.Error)
	}
	if !strings.Contains(result.Output, "hello") {
		t.Errorf("output should contain 'hello', got: %q", result.Output)
	}
}

func TestRunCommand_FailedCommand(t *testing.T) {
	// Try to run a nonexistent command
	result := RunCommand("nonexistent_command_12345", "", 30)

	if result.Success {
		t.Error("RunCommand should fail for nonexistent command")
	}
}

func TestRunCommand_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "ping -n 10 127.0.0.1"
	} else {
		cmd = "sleep 10"
	}

	result := RunCommand(cmd, "", 1)

	if result.Success {
		t.Error("RunCommand should fail due to timeout")
	}
	if !strings.Contains(strings.ToLower(result.Error), "timed out") && !strings.Contains(strings.ToLower(result.Error), "timeout") {
		t.Errorf("error should mention timeout, got: %q", result.Error)
	}
}

func TestRunCommand_BlockedCommand(t *testing.T) {
	result := RunCommand("rm -rf /", "", 30)

	if result.Success {
		t.Error("RunCommand should block dangerous commands")
	}
	if !strings.Contains(strings.ToLower(result.Error), "blocked") {
		t.Errorf("error should mention blocked, got: %q", result.Error)
	}
}

func TestRunCommand_WorkingDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cmd-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file in the temp directory
	testFile := filepath.Join(tmpDir, "testfile.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "dir"
	} else {
		cmd = "ls"
	}

	result := RunCommand(cmd, tmpDir, 30)

	if !result.Success {
		t.Errorf("RunCommand failed: %s", result.Error)
	}
	if !strings.Contains(result.Output, "testfile") {
		t.Errorf("output should contain 'testfile', got: %q", result.Output)
	}
}

func TestRunCommand_RecordsHistory(t *testing.T) {
	// Reset session first
	ResetSession()
	initialCount := len(GetSession().History)

	RunCommand("echo test", "", 30)

	newCount := len(GetSession().History)
	if newCount != initialCount+1 {
		t.Errorf("expected history count %d, got %d", initialCount+1, newCount)
	}
}

func TestGetCurrentDirectory(t *testing.T) {
	ResetSession()

	result := GetCurrentDirectory()

	if !result.Success {
		t.Errorf("GetCurrentDirectory failed: %s", result.Error)
	}

	home, _ := os.UserHomeDir()
	if result.Output != home {
		t.Errorf("GetCurrentDirectory = %q, want %q", result.Output, home)
	}
}

func TestChangeDirectory_Valid(t *testing.T) {
	ResetSession()

	tmpDir, err := os.MkdirTemp("", "cd-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	result := ChangeDirectory(tmpDir)

	if !result.Success {
		t.Errorf("ChangeDirectory failed: %s", result.Error)
	}

	if GetSession().CWD != tmpDir {
		t.Errorf("session CWD = %q, want %q", GetSession().CWD, tmpDir)
	}
}

func TestChangeDirectory_Invalid(t *testing.T) {
	result := ChangeDirectory("/nonexistent/directory/path")

	if result.Success {
		t.Error("ChangeDirectory should fail for nonexistent directory")
	}
}

func TestTaskComplete_FormatsOutput(t *testing.T) {
	result := TaskComplete("Task finished successfully", []string{"file1.txt", "file2.txt"})

	if !result.Success {
		t.Errorf("TaskComplete failed: %s", result.Error)
	}
	if !strings.Contains(result.Output, "Task finished successfully") {
		t.Error("output should contain summary")
	}
	if !strings.Contains(result.Output, "file1.txt") {
		t.Error("output should contain modified files")
	}
	if !strings.Contains(result.Output, "file2.txt") {
		t.Error("output should contain modified files")
	}
	if !strings.Contains(result.Output, "✅") {
		t.Error("output should contain checkmark")
	}
}

func TestTaskComplete_NoFiles(t *testing.T) {
	result := TaskComplete("Done", nil)

	if !result.Success {
		t.Errorf("TaskComplete failed: %s", result.Error)
	}
	if !strings.Contains(result.Output, "Done") {
		t.Error("output should contain summary")
	}
}
