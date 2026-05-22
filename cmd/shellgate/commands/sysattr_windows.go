//go:build windows

package commands

import (
	"os/exec"
	"syscall"
)

func newSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: 0x00000008} // DETACHED_PROCESS
}

func newStopCmd() *exec.Cmd {
	return nil
}
