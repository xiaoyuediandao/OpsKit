package tools

import (
	"fmt"
	"os"
	"strings"
)

// ListDir lists the contents of a directory.
func ListDir(path string) ToolResult {
	if path == "" {
		path = "."
	}
	path = expandTilde(path)

	entries, err := os.ReadDir(path)
	if err != nil {
		return ToolResult{Error: fmt.Sprintf("cannot read directory: %v", err), IsErr: true}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Contents of %s:\n", path))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		if e.IsDir() {
			sb.WriteString(fmt.Sprintf("  [DIR]  %s/\n", e.Name()))
		} else {
			sb.WriteString(fmt.Sprintf("  [FILE] %s  (%d bytes)\n", e.Name(), info.Size()))
		}
	}
	if len(entries) == 0 {
		sb.WriteString("  (empty directory)\n")
	}
	return ToolResult{Output: sb.String()}
}
