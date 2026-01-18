package tools

import (
	"os"
	"testing"
)

func TestToolResult_Success(t *testing.T) {
	result := ToolResult{
		Success: true,
		Output:  "command output",
		Error:   "",
	}

	if !result.Success {
		t.Error("expected Success to be true")
	}
	if result.Output != "command output" {
		t.Errorf("expected Output='command output', got %q", result.Output)
	}
	if result.Error != "" {
		t.Errorf("expected empty Error, got %q", result.Error)
	}
}

func TestToolResult_Error(t *testing.T) {
	result := ToolResult{
		Success: false,
		Output:  "",
		Error:   "something went wrong",
	}

	if result.Success {
		t.Error("expected Success to be false")
	}
	if result.Error != "something went wrong" {
		t.Errorf("expected Error='something went wrong', got %q", result.Error)
	}
}

func TestShellSession_DefaultValues(t *testing.T) {
	session := NewShellSession()

	// Should start with home directory
	home, _ := os.UserHomeDir()
	if session.CWD != home {
		t.Errorf("expected CWD=%q, got %q", home, session.CWD)
	}

	// Should have environment variables
	if len(session.Env) == 0 {
		t.Error("expected Env to have values from os.Environ")
	}

	// Should have empty history
	if len(session.History) != 0 {
		t.Errorf("expected empty History, got %d items", len(session.History))
	}
}

func TestShellSession_RecordCommand(t *testing.T) {
	session := NewShellSession()

	session.RecordCommand("ls -la", 0)

	if len(session.History) != 1 {
		t.Fatalf("expected 1 history item, got %d", len(session.History))
	}

	record := session.History[0]
	if record.Command != "ls -la" {
		t.Errorf("expected Command='ls -la', got %q", record.Command)
	}
	if record.ExitCode != 0 {
		t.Errorf("expected ExitCode=0, got %d", record.ExitCode)
	}
	if record.CWD != session.CWD {
		t.Errorf("expected CWD=%q, got %q", session.CWD, record.CWD)
	}
}

func TestShellSession_Reset(t *testing.T) {
	session := NewShellSession()

	// Modify the session
	session.CWD = "/some/other/path"
	session.RecordCommand("test", 0)

	// Reset
	session.Reset()

	home, _ := os.UserHomeDir()
	if session.CWD != home {
		t.Errorf("after Reset, expected CWD=%q, got %q", home, session.CWD)
	}
	if len(session.History) != 0 {
		t.Errorf("after Reset, expected empty History, got %d items", len(session.History))
	}
}

func TestCommandRecord(t *testing.T) {
	record := CommandRecord{
		Command:  "git status",
		CWD:      "/repo",
		ExitCode: 0,
	}

	if record.Command != "git status" {
		t.Errorf("expected Command='git status', got %q", record.Command)
	}
	if record.CWD != "/repo" {
		t.Errorf("expected CWD='/repo', got %q", record.CWD)
	}
	if record.ExitCode != 0 {
		t.Errorf("expected ExitCode=0, got %d", record.ExitCode)
	}
}
