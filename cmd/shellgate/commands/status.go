package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/dutakey/shellgate/config"
	"github.com/dutakey/shellgate/internal/store"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show server and configuration status",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("ShellGate Status")
			fmt.Println("────────────────────────────────────")

			// Server status from PID file
			pidFile := pidFilePath()
			if data, err := os.ReadFile(pidFile); err == nil {
				if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
					if isRunning(pid) {
						fmt.Printf("Server:   running (PID %d)\n", pid)
					} else {
						fmt.Println("Server:   stopped (stale PID file)")
						os.Remove(pidFile)
					}
				}
			} else {
				fmt.Println("Server:   stopped")
			}

			// Config details
			fmt.Printf("Config:   %s\n", configPath)
			cfg, err := config.Load(configPath)
			if err != nil {
				fmt.Printf("          (cannot load: %s)\n", err)
			} else {
				fmt.Printf("Provider: %s\n", cfg.Executor.Provider)
				fmt.Printf("Port:     %d\n", cfg.Server.Port)

				if ks, err := store.NewKeyStore(cfg.Auth.KeysFile); err == nil {
					fmt.Printf("Keys:     %d active\n", len(ks.List()))
				}
			}

			// Log file
			logFile := logFilePath()
			if _, err := os.Stat(logFile); err == nil {
				fmt.Printf("Logs:     %s\n", logFile)
			}

			fmt.Println("────────────────────────────────────")
			return nil
		},
	}
}

func isRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}
