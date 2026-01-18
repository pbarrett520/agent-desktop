package tools

import (
	"fmt"
)

// ToolFunction represents a function definition in OpenAI format.
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolDefinition represents a tool definition in OpenAI function calling format.
type ToolDefinition struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// toolDefinitions contains all available tool definitions.
var toolDefinitions = []ToolDefinition{
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "run_command",
			Description: "Execute a shell command and return the output. Use this to run any command-line operation.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "The shell command to execute",
					},
					"working_dir": map[string]interface{}{
						"type":        "string",
						"description": "Directory to run the command in. If not specified, uses the current working directory.",
					},
					"timeout": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum time in seconds to wait for the command. Default is 60.",
						"default":     60,
					},
				},
				"required": []string{"command"},
			},
		},
	},
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "read_file",
			Description: "Read the contents of a file.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file to read",
					},
					"max_lines": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of lines to read. If not specified, reads entire file.",
					},
				},
				"required": []string{"path"},
			},
		},
	},
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "write_file",
			Description: "Write content to a file. Creates the file if it doesn't exist.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file to write",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Content to write to the file",
					},
					"append": map[string]interface{}{
						"type":        "boolean",
						"description": "If true, append to the file instead of overwriting. Default is false.",
						"default":     false,
					},
				},
				"required": []string{"path", "content"},
			},
		},
	},
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "list_directory",
			Description: "List files and directories in a path.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the directory to list. Defaults to current working directory.",
					},
					"show_hidden": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether to show hidden files (starting with .). Default is false.",
						"default":     false,
					},
				},
				"required": []string{},
			},
		},
	},
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "get_current_directory",
			Description: "Get the current working directory.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
	},
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "change_directory",
			Description: "Change the current working directory.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to change to",
					},
				},
				"required": []string{"path"},
			},
		},
	},
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "task_complete",
			Description: "Call this when you have completed the user's task. Provide a summary of what was done.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"summary": map[string]interface{}{
						"type":        "string",
						"description": "A summary of what was accomplished",
					},
					"files_modified": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "List of files that were created or modified",
					},
				},
				"required": []string{"summary"},
			},
		},
	},
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "delete_file",
			Description: "Delete a file. Use with caution.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file to delete",
					},
					"confirm": map[string]interface{}{
						"type":        "boolean",
						"description": "Must be true to confirm deletion",
					},
				},
				"required": []string{"path", "confirm"},
			},
		},
	},
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "copy_file",
			Description: "Copy a file to a new location.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"source": map[string]interface{}{
						"type":        "string",
						"description": "Path to the source file",
					},
					"destination": map[string]interface{}{
						"type":        "string",
						"description": "Path to the destination",
					},
				},
				"required": []string{"source", "destination"},
			},
		},
	},
	{
		Type: "function",
		Function: ToolFunction{
			Name:        "move_file",
			Description: "Move or rename a file.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"source": map[string]interface{}{
						"type":        "string",
						"description": "Path to the source file",
					},
					"destination": map[string]interface{}{
						"type":        "string",
						"description": "Path to the destination",
					},
				},
				"required": []string{"source", "destination"},
			},
		},
	},
}

// GetToolDefinitions returns all available tool definitions in OpenAI format.
func GetToolDefinitions() []ToolDefinition {
	return toolDefinitions
}

// ExecuteTool executes a tool by name with the given arguments.
func ExecuteTool(name string, args map[string]interface{}) ToolResult {
	switch name {
	case "run_command":
		command, ok := args["command"].(string)
		if !ok {
			return ToolResult{Success: false, Error: "run_command requires 'command' argument"}
		}
		workingDir, _ := args["working_dir"].(string)
		timeout := 60
		if t, ok := args["timeout"].(float64); ok {
			timeout = int(t)
		} else if t, ok := args["timeout"].(int); ok {
			timeout = t
		}
		return RunCommand(command, workingDir, timeout)

	case "read_file":
		path, ok := args["path"].(string)
		if !ok {
			return ToolResult{Success: false, Error: "read_file requires 'path' argument"}
		}
		var maxLines *int
		if ml, ok := args["max_lines"].(float64); ok {
			mlInt := int(ml)
			maxLines = &mlInt
		} else if ml, ok := args["max_lines"].(int); ok {
			maxLines = &ml
		}
		return ReadFile(path, maxLines)

	case "write_file":
		path, ok := args["path"].(string)
		if !ok {
			return ToolResult{Success: false, Error: "write_file requires 'path' argument"}
		}
		content, ok := args["content"].(string)
		if !ok {
			return ToolResult{Success: false, Error: "write_file requires 'content' argument"}
		}
		appendFlag := false
		if a, ok := args["append"].(bool); ok {
			appendFlag = a
		}
		return WriteFile(path, content, appendFlag)

	case "list_directory":
		path, _ := args["path"].(string)
		showHidden := false
		if sh, ok := args["show_hidden"].(bool); ok {
			showHidden = sh
		}
		return ListDirectory(path, showHidden)

	case "get_current_directory":
		return GetCurrentDirectory()

	case "change_directory":
		path, ok := args["path"].(string)
		if !ok {
			return ToolResult{Success: false, Error: "change_directory requires 'path' argument"}
		}
		return ChangeDirectory(path)

	case "task_complete":
		summary, ok := args["summary"].(string)
		if !ok {
			return ToolResult{Success: false, Error: "task_complete requires 'summary' argument"}
		}
		var filesModified []string
		if fm, ok := args["files_modified"].([]interface{}); ok {
			for _, f := range fm {
				if s, ok := f.(string); ok {
					filesModified = append(filesModified, s)
				}
			}
		}
		return TaskComplete(summary, filesModified)

	case "delete_file":
		path, ok := args["path"].(string)
		if !ok {
			return ToolResult{Success: false, Error: "delete_file requires 'path' argument"}
		}
		confirm := false
		if c, ok := args["confirm"].(bool); ok {
			confirm = c
		}
		return DeleteFile(path, confirm)

	case "copy_file":
		source, ok := args["source"].(string)
		if !ok {
			return ToolResult{Success: false, Error: "copy_file requires 'source' argument"}
		}
		destination, ok := args["destination"].(string)
		if !ok {
			return ToolResult{Success: false, Error: "copy_file requires 'destination' argument"}
		}
		return CopyFile(source, destination)

	case "move_file":
		source, ok := args["source"].(string)
		if !ok {
			return ToolResult{Success: false, Error: "move_file requires 'source' argument"}
		}
		destination, ok := args["destination"].(string)
		if !ok {
			return ToolResult{Success: false, Error: "move_file requires 'destination' argument"}
		}
		return MoveFile(source, destination)

	default:
		return ToolResult{Success: false, Error: fmt.Sprintf("Unknown tool: %s", name)}
	}
}
