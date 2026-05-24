package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

func newRestartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: "Restart the background ShellGate server",
		RunE: func(cmd *cobra.Command, args []string) error {
			pidFile := pidFilePath()
			data, err := os.ReadFile(pidFile)
			if err == nil {
				pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
				if err == nil {
					if proc, err := os.FindProcess(pid); err == nil {
						_ = proc.Signal(syscall.SIGTERM)
						fmt.Printf("Stopped ShellGate (PID %d)\n", pid)
						time.Sleep(500 * time.Millisecond)
					}
				}
				os.Remove(pidFile)
			}

			fmt.Println("Starting ShellGate in background...")
			return runDetached()
		},
	}
}
