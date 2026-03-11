//go:build windows

package exec

import (
	"fmt"
	"os/exec"
	"syscall"
)

func newShellCommand(command string) *exec.Cmd {
	return exec.Command("powershell", "-NoProfile", "-Command", command)
}

func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x00000200} // CREATE_NEW_PROCESS_GROUP
}

func killProcessGroup(cmd *exec.Cmd) {
	exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", cmd.Process.Pid)).Run()
}
