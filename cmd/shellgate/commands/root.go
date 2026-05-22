package commands

import (
	"github.com/spf13/cobra"
)

var configPath string

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "shellgate",
		Short: "CLI proxy — turn authenticated CLI tools into OpenAI-compatible APIs",
		Long: `ShellGate wraps locally authenticated CLI tools (Codex, Claude, and others)
into an OpenAI-compatible REST API. No extra API keys needed — use your existing CLI login.`,
	}

	root.PersistentFlags().StringVarP(&configPath, "config", "c", "config.toml", "config file path")

	root.AddCommand(
		newServeCmd(),
		newInitCmd(),
		newLoginCmd(),
		newKeysCmd(),
	)

	return root
}
