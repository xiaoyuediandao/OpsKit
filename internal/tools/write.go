package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile writes content to a file, creating parent directories as needed.
func WriteFile(path, content string) ToolResult {
	if path == "" {
		return ToolResult{Error: "path is required", IsErr: true}
	}
	if content == "" {
		return ToolResult{Error: "content is required", IsErr: true}
	}

	path = expandTilde(path)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return ToolResult{Error: fmt.Sprintf("cannot create directories: %v", err), IsErr: true}
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return ToolResult{Error: fmt.Sprintf("cannot write file: %v", err), IsErr: true}
	}

	return ToolResult{Output: fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path)}
}
