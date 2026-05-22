//go:build !windows

package commands

import (
	"os/exec"
	"syscall"
)

func newSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}

// stopCmd returns a command to stop a running background process
func newStopCmd() *exec.Cmd {
	return nil
}
