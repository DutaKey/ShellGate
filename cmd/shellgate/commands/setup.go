package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/dutakey/shellgate/config"
	"github.com/dutakey/shellgate/internal/store"
)

func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Guided first-time setup — init, login, create key, start server",
		Long: `Interactive wizard that walks through the full ShellGate setup:
  1. Create config file
  2. Authenticate your CLI provider
  3. Create your first API key
  4. Start the server in the background`,
		RunE: func(cmd *cobra.Command, args []string) error {
			printStep(1, "Configure ShellGate")
			provider, err := runInit(configPath, false)
			if err != nil {
				return err
			}
			if provider == "" {
				return nil // user aborted
			}

			printStep(2, fmt.Sprintf("Login to %s", provider))
			if err := runProviderLogin(provider); err != nil {
				return err
			}

			printStep(3, "Create your first API key")
			key, err := createFirstKey(configPath)
			if err != nil {
				return fmt.Errorf("create API key: %w", err)
			}

			printStep(4, "Start ShellGate in background")
			if err := runDetached(); err != nil {
				return err
			}

			cfg, _ := config.Load(configPath)
			port := 8080
			if cfg != nil {
				port = cfg.Server.Port
			}

			fmt.Println()
			fmt.Println("╔═══════════════════════════════════════════╗")
			fmt.Println("║        ShellGate is ready!                ║")
			fmt.Println("╚═══════════════════════════════════════════╝")
			fmt.Println()
			fmt.Printf("  API endpoint: http://localhost:%d/v1\n", port)
			fmt.Printf("  API key:      %s\n", key)
			fmt.Printf("  Provider:     %s\n", provider)
			fmt.Println()
			fmt.Println("  Example:")
			fmt.Printf("    curl http://localhost:%d/v1/chat/completions \\\n", port)
			fmt.Println("      -H \"Content-Type: application/json\" \\")
			fmt.Printf("      -H \"Authorization: Bearer %s\" \\\n", key)
			fmt.Println("      -d '{\"model\":\"gpt-5.4\",\"messages\":[{\"role\":\"user\",\"content\":\"hello\"}]}'")
			fmt.Println()
			fmt.Println("  shellgate status   — check server status")
			fmt.Println("  shellgate logs     — view server logs")
			fmt.Println("  shellgate stop     — stop the server")

			return nil
		},
	}
}

func printStep(n int, label string) {
	fmt.Printf("\n[%d/4] %s\n", n, label)
	fmt.Println("────────────────────────────────────")
}

func runProviderLogin(provider string) error {
	p, ok := providers[provider]
	if !ok {
		return fmt.Errorf("unknown provider %q", provider)
	}

	if _, err := exec.LookPath(p.binary); err != nil {
		return fmt.Errorf("%s CLI not found in PATH\nInstall: %s", p.binary, p.installHint)
	}

	fmt.Printf("Logging in to %s...\n\n", provider)
	loginCmd := exec.Command(p.binary, p.loginArgs...)
	loginCmd.Stdin = os.Stdin
	loginCmd.Stdout = os.Stdout
	loginCmd.Stderr = os.Stderr
	if err := loginCmd.Run(); err != nil {
		return fmt.Errorf("%s login failed: %w", provider, err)
	}
	return nil
}

func createFirstKey(cfgPath string) (string, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return "", err
	}

	ks, err := store.NewKeyStore(cfg.Auth.KeysFile)
	if err != nil {
		return "", err
	}

	entry, err := ks.Create("default")
	if err != nil {
		return "", err
	}

	fmt.Printf("API key created: %s\n", entry.Key)
	return entry.Key, nil
}
