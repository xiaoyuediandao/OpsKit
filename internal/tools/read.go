package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const maxFileSize = 50 * 1024 // 50KB

// ReadFile reads a file and returns its contents.
func ReadFile(path string) ToolResult {
	if path == "" {
		return ToolResult{Error: "path is required", IsErr: true}
	}

	path = expandTilde(path)

	info, err := os.Stat(path)
	if err != nil {
		return ToolResult{Error: fmt.Sprintf("cannot stat file: %v", err), IsErr: true}
	}
	if info.IsDir() {
		return ToolResult{Error: "path is a directory, use list_dir instead", IsErr: true}
	}
	if info.Size() > maxFileSize {
		return ToolResult{Error: fmt.Sprintf("file too large (%d bytes, max %d bytes)", info.Size(), maxFileSize), IsErr: true}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return ToolResult{Error: fmt.Sprintf("cannot read file: %v", err), IsErr: true}
	}
	return ToolResult{Output: string(data)}
}

func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}
