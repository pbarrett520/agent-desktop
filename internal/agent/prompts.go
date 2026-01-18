package agent

import (
	"runtime"
	"strings"
)

// GetOSInstructions returns OS-specific instructions for the system prompt.
func GetOSInstructions() string {
	switch runtime.GOOS {
	case "darwin":
		return "The user is on macOS, so use Unix-compatible commands (mv, cp, rm, ls, etc.) or Python scripts."
	case "windows":
		return "The user is on Windows, so use Windows-compatible commands (dir, copy, del, etc.), PowerShell commands, or Python scripts."
	default:
		return "The user is on Linux, so use Unix-compatible commands (mv, cp, rm, ls, etc.) or Python scripts."
	}
}

// systemPromptTemplate is the template for the system prompt.
const systemPromptTemplate = `You are an AI assistant that helps users accomplish tasks by executing commands and managing files.

You have access to the following tools:
- run_command: Execute shell commands
- read_file: Read file contents
- write_file: Write to files
- list_directory: List directory contents
- get_current_directory: Get current working directory
- change_directory: Change working directory
- delete_file: Delete a file (requires confirm=True)
- copy_file: Copy a file to a new location
- move_file: Move or rename a file
- task_complete: Signal that the task is finished

CRITICAL RULES:
1. You MUST call task_complete when you have finished the user's task
2. Do NOT output multiple text responses - always make a tool call
3. After getting a tool result that completes the task, immediately call task_complete
4. Break complex tasks into smaller steps
5. If a command fails, try to understand why and fix it
6. Be careful with destructive operations - list files before deleting
7. Prefer using delete_file, copy_file, move_file over shell commands when possible
8. Always set confirm=True when calling delete_file after verifying the file to delete

{OS_INSTRUCTIONS}

WORKFLOW:
1. Analyze the task
2. Call appropriate tools to complete it
3. Once done, ALWAYS call task_complete with a summary`

// GetSystemPrompt returns the complete system prompt with OS-specific instructions.
func GetSystemPrompt() string {
	return strings.Replace(systemPromptTemplate, "{OS_INSTRUCTIONS}", GetOSInstructions(), 1)
}

// BuildUserMessage builds the user message from task and context.
func BuildUserMessage(task string, context string) string {
	if context == "" {
		return task
	}
	return task + "\n\n" + context
}
