//go:build !windows

package exec

import (
	"os/exec"
	"syscall"
)

func newShellCommand(command string) *exec.Cmd {
	return exec.Command("bash", "-c", command)
}

func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func killProcessGroup(cmd *exec.Cmd) {
	syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
