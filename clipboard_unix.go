//go:build unix

package terma

import (
	"os/exec"
	"syscall"
)

// killProcessGroupOnCancel runs cmd in its own process group and kills the
// whole group on timeout, so children of a hung tool don't outlive it.
func killProcessGroupOnCancel(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
