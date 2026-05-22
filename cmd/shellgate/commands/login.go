package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var providers = map[string]providerDef{
	"codex": {
		binary:      "codex",
		loginArgs:   []string{"login"},
		statusArgs:  []string{"login", "status"},
		installHint: "npm install -g @openai/codex",
	},
}

type providerDef struct {
	binary      string
	loginArgs   []string
	statusArgs  []string
	installHint string
}

func newLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login <provider>",
		Short: "Authenticate a CLI provider (e.g. codex)",
		Long: `Log in to a supported CLI provider so ShellGate can proxy requests through it.

Supported providers:
  codex    OpenAI Codex CLI`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			p, ok := providers[name]
			if !ok {
				return fmt.Errorf("unknown provider %q — run `shellgate login --help` for supported providers", name)
			}

			// check binary exists
			if _, err := exec.LookPath(p.binary); err != nil {
				return fmt.Errorf("%s CLI not found in PATH\nInstall: %s", p.binary, p.installHint)
			}

			fmt.Printf("Logging in to %s...\n\n", name)

			// passthrough stdin/stdout so OAuth flow works interactively
			loginCmd := exec.Command(p.binary, p.loginArgs...)
			loginCmd.Stdin = os.Stdin
			loginCmd.Stdout = os.Stdout
			loginCmd.Stderr = os.Stderr

			if err := loginCmd.Run(); err != nil {
				return fmt.Errorf("%s login failed: %w", name, err)
			}

			fmt.Printf("\n%s login successful.\n", name)
			fmt.Println("\nNext: shellgate serve")
			return nil
		},
	}

	return cmd
}
