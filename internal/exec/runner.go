package exec

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// Runner executes shell commands.
type Runner struct {
	Timeout time.Duration
}

// NewRunner creates a Runner with the given timeout.
func NewRunner(timeout time.Duration) *Runner {
	return &Runner{Timeout: timeout}
}

// Run executes the shell command and returns combined stdout+stderr output.
// It kills the entire process group on timeout to clean up interactive child processes.
func (r *Runner) Run(cmd string) (string, error) {
	c := newShellCommand(cmd)
	setSysProcAttr(c)
	// No stdin — non-interactive
	c.Stdin = nil

	var buf bytes.Buffer
	c.Stdout = &buf
	c.Stderr = &buf

	if err := c.Start(); err != nil {
		return "", fmt.Errorf("start command: %w", err)
	}

	done := make(chan error, 1)
	go func() { done <- c.Wait() }()

	var runErr error
	select {
	case runErr = <-done:
		// Completed normally
	case <-time.After(r.Timeout):
		// Kill entire process group
		killProcessGroup(c)
		<-done
		runErr = fmt.Errorf("command timed out after %v", r.Timeout)
	}

	output := buf.String()
	if len(output) > 8000 {
		half := 4000
		output = output[:half] + "\n...[output truncated]...\n" + output[len(output)-half:]
	}
	return strings.TrimRight(output, "\n"), runErr
}
