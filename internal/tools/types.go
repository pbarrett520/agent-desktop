package tools

import (
	"os"
	"sync"
)

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

// CommandRecord represents a recorded command in the session history.
type CommandRecord struct {
	Command  string `json:"command"`
	CWD      string `json:"cwd"`
	ExitCode int    `json:"exit_code"`
}

// ShellSession maintains state for shell command execution.
type ShellSession struct {
	CWD     string            `json:"cwd"`
	Env     map[string]string `json:"env"`
	History []CommandRecord   `json:"history"`
	mu      sync.Mutex
}

// NewShellSession creates a new shell session with default values.
func NewShellSession() *ShellSession {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	// Copy current environment
	env := make(map[string]string)
	for _, e := range os.Environ() {
		for i := 0; i < len(e); i++ {
			if e[i] == '=' {
				env[e[:i]] = e[i+1:]
				break
			}
		}
	}

	return &ShellSession{
		CWD:     home,
		Env:     env,
		History: make([]CommandRecord, 0),
	}
}

// RecordCommand adds a command to the session history.
func (s *ShellSession) RecordCommand(command string, exitCode int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.History = append(s.History, CommandRecord{
		Command:  command,
		CWD:      s.CWD,
		ExitCode: exitCode,
	})
}

// Reset resets the shell session to its initial state.
func (s *ShellSession) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	s.CWD = home
	s.History = make([]CommandRecord, 0)
}

// GetInfo returns information about the current session.
func (s *ShellSession) GetInfo() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get last 5 commands
	lastCommands := s.History
	if len(lastCommands) > 5 {
		lastCommands = lastCommands[len(lastCommands)-5:]
	}

	return map[string]interface{}{
		"cwd":           s.CWD,
		"history_count": len(s.History),
		"last_commands": lastCommands,
	}
}

// globalSession is the global shell session used by tool implementations.
var globalSession = NewShellSession()

// GetSession returns the global shell session.
func GetSession() *ShellSession {
	return globalSession
}

// ResetSession resets the global shell session.
func ResetSession() {
	globalSession.Reset()
}

// GetSessionInfo returns information about the global session.
func GetSessionInfo() map[string]interface{} {
	return globalSession.GetInfo()
}
