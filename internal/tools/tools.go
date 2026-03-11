package tools

import (
	"opskit/internal/ai"
)

// ToolResult holds the output of a tool execution.
type ToolResult struct {
	Output string
	Error  string
	IsErr  bool
}

// AllToolDefinitions returns all available tool definitions for the AI.
func AllToolDefinitions() []ai.ToolDefinition {
	return []ai.ToolDefinition{
		{
			Type: "function",
			Function: ai.ToolFunction{
				Name:        "bash",
				Description: "Execute a shell command (bash on macOS/Linux, PowerShell on Windows). Returns stdout and stderr combined. Use for running system commands, checking status, installing software, etc.",
				Parameters: ai.ToolParam{
					Type: "object",
					Properties: map[string]interface{}{
						"command": map[string]interface{}{
							"type":        "string",
							"description": "The shell command to execute",
						},
					},
					Required: []string{"command"},
				},
			},
		},
		{
			Type: "function",
			Function: ai.ToolFunction{
				Name:        "read_file",
				Description: "Read the contents of a file. Supports ~ for home directory. Returns file contents up to 50KB.",
				Parameters: ai.ToolParam{
					Type: "object",
					Properties: map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "The file path to read (supports ~ expansion)",
						},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: ai.ToolFunction{
				Name:        "write_file",
				Description: "Write content to a file, creating parent directories as needed. Supports ~ for home directory.",
				Parameters: ai.ToolParam{
					Type: "object",
					Properties: map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "The file path to write (supports ~ expansion)",
						},
						"content": map[string]interface{}{
							"type":        "string",
							"description": "The content to write to the file",
						},
					},
					Required: []string{"path", "content"},
				},
			},
		},
		{
			Type: "function",
			Function: ai.ToolFunction{
				Name:        "list_dir",
				Description: "List the contents of a directory. Supports ~ for home directory.",
				Parameters: ai.ToolParam{
					Type: "object",
					Properties: map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "The directory path to list (supports ~ expansion)",
						},
					},
					Required: []string{"path"},
				},
			},
		},
	}
}

// Execute dispatches a tool call by name with the given arguments.
func Execute(name string, args map[string]interface{}) ToolResult {
	switch name {
	case "bash":
		cmd, _ := args["command"].(string)
		return RunBash(cmd)
	case "read_file":
		path, _ := args["path"].(string)
		return ReadFile(path)
	case "write_file":
		path, _ := args["path"].(string)
		content, _ := args["content"].(string)
		return WriteFile(path, content)
	case "list_dir":
		path, _ := args["path"].(string)
		return ListDir(path)
	default:
		return ToolResult{Error: "unknown tool: " + name, IsErr: true}
	}
}
