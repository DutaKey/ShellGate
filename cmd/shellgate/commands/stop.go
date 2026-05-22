package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

func newStopServerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop a background ShellGate server",
		RunE: func(cmd *cobra.Command, args []string) error {
			pidFile := "shellgate.pid"
			data, err := os.ReadFile(pidFile)
			if err != nil {
				return fmt.Errorf("no PID file found (%s) — is ShellGate running in background?", pidFile)
			}

			pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err != nil {
				return fmt.Errorf("invalid PID in %s", pidFile)
			}

			proc, err := os.FindProcess(pid)
			if err != nil {
				return fmt.Errorf("process %d not found", pid)
			}

			if err := proc.Signal(syscall.SIGTERM); err != nil {
				return fmt.Errorf("send SIGTERM to %d: %w", pid, err)
			}

			os.Remove(pidFile)
			fmt.Printf("ShellGate (PID %d) stopped.\n", pid)
			return nil
		},
	}
}
