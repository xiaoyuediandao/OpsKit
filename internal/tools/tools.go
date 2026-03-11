package tools

import (
	"opskit/internal/ai"
	"opskit/internal/security"
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
		{
			Type: "function",
			Function: ai.ToolFunction{
				Name:        "security_scan",
				Description: "扫描 OpenClaw 安全配置，检查 SECURITY.md 部署、AGENTS.md 引用、SOUL.md 安全边界、文件权限和密钥泄露。输出安全评分报告（A/B/C/D/F）。",
				Parameters: ai.ToolParam{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: ai.ToolFunction{
				Name:        "security_fix",
				Description: "自动加固 OpenClaw 安全配置：部署 SECURITY.md、注入 AGENTS.md 安全引用、补丁 SOUL.md 安全边界、修复文件权限。幂等操作，可重复执行。",
				Parameters: ai.ToolParam{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
		},
	}
}

// Execute dispatches a tool call by name with the given arguments.
// All tool outputs are redacted to remove sensitive information.
func Execute(name string, args map[string]interface{}) ToolResult {
	var result ToolResult
	switch name {
	case "bash":
		cmd, _ := args["command"].(string)
		result = RunBash(cmd)
	case "read_file":
		path, _ := args["path"].(string)
		result = ReadFile(path)
	case "write_file":
		path, _ := args["path"].(string)
		content, _ := args["content"].(string)
		result = WriteFile(path, content)
	case "list_dir":
		path, _ := args["path"].(string)
		result = ListDir(path)
	case "security_scan":
		result = RunSecurityScan()
	case "security_fix":
		result = RunSecurityFix()
	default:
		return ToolResult{Error: "unknown tool: " + name, IsErr: true}
	}

	// Apply redaction to all tool outputs (defense in depth)
	result.Output = security.Redact(result.Output)
	if result.Error != "" {
		result.Error = security.Redact(result.Error)
	}
	return result
}
