package tools

import (
	"opskit/internal/exec"
	"time"
)

// RunBash executes a shell command with a 30-second timeout.
func RunBash(command string) ToolResult {
	if command == "" {
		return ToolResult{Error: "command is required", IsErr: true}
	}
	runner := exec.NewRunner(30 * time.Second)
	output, err := runner.Run(command)
	if err != nil {
		// Even on error, include any output that was produced
		if output != "" {
			return ToolResult{
				Output: output,
				Error:  err.Error(),
				IsErr:  true,
			}
		}
		return ToolResult{Error: err.Error(), IsErr: true}
	}
	return ToolResult{Output: output}
}
