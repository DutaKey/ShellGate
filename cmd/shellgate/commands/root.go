package commands

import (
	"github.com/spf13/cobra"
)

var configPath string

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "shellgate",
		Short: "CLI proxy — turn authenticated CLI tools into OpenAI-compatible APIs",
		Long: `ShellGate wraps locally authenticated CLI tools (Codex, Kimi, and others)
into an OpenAI-compatible REST API. No extra API keys needed — use your existing CLI login.`,
	}

	root.PersistentFlags().StringVarP(&configPath, "config", "c", defaultConfigPath(), "config file path")

	root.AddCommand(
		newSetupCmd(),
		newServeCmd(),
		newStopServerCmd(),
		newRestartCmd(),
		newInitCmd(),
		newLoginCmd(),
		newKeysCmd(),
		newStatusCmd(),
		newLogsCmd(),
	)

	return root
}
