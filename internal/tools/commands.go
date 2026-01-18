package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// RunCommand executes a shell command and returns the output.
// It checks command safety before execution and records the command in history.
func RunCommand(command string, workingDir string, timeout int) ToolResult {
	// Check command safety first
	safe, reason := CheckCommandSafety(command)
	if !safe {
		return ToolResult{Success: false, Error: reason}
	}

	session := GetSession()

	// Determine working directory
	cwd := session.CWD
	if workingDir != "" {
		cwd = ExpandPath(workingDir, session.CWD)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// Create command based on OS
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "bash", "-c", command)
	}

	cmd.Dir = cwd

	// Set environment from session
	env := os.Environ()
	for k, v := range session.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	// Run command and capture output
	output, err := cmd.CombinedOutput()

	// Record in history
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	session.RecordCommand(command, exitCode)

	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		return ToolResult{
			Success: false,
			Output:  string(output),
			Error:   fmt.Sprintf("Command timed out after %d seconds", timeout),
		}
	}

	// Check for error
	if err != nil {
		return ToolResult{
			Success: false,
			Output:  string(output),
			Error:   fmt.Sprintf("Command failed with exit code %d: %s", exitCode, err.Error()),
		}
	}

	return ToolResult{
		Success: true,
		Output:  strings.TrimRight(string(output), "\r\n"),
	}
}

// GetCurrentDirectory returns the current working directory of the session.
func GetCurrentDirectory() ToolResult {
	return ToolResult{
		Success: true,
		Output:  GetSession().CWD,
	}
}

// ChangeDirectory changes the current working directory of the session.
func ChangeDirectory(path string) ToolResult {
	session := GetSession()

	// Expand path
	expandedPath := ExpandPath(path, session.CWD)

	// Get absolute path
	absPath, err := filepath.Abs(expandedPath)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}
	}

	// Check if directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ToolResult{Success: false, Error: fmt.Sprintf("Directory not found: %s", absPath)}
		}
		return ToolResult{Success: false, Error: err.Error()}
	}

	if !info.IsDir() {
		return ToolResult{Success: false, Error: fmt.Sprintf("Not a directory: %s", absPath)}
	}

	// Update session CWD
	session.mu.Lock()
	session.CWD = absPath
	session.mu.Unlock()

	return ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Changed directory to: %s", absPath),
	}
}

// TaskComplete signals that the agent has completed its task.
// It returns a formatted summary of what was accomplished.
func TaskComplete(summary string, filesModified []string) ToolResult {
	output := fmt.Sprintf("✅ Task completed!\n\n%s", summary)

	if len(filesModified) > 0 {
		output += "\n\nFiles modified:\n"
		for _, f := range filesModified {
			output += fmt.Sprintf("  • %s\n", f)
		}
	}

	return ToolResult{
		Success: true,
		Output:  output,
	}
}
