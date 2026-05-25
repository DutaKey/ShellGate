package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/dutakey/shellgate/config"
	"github.com/dutakey/shellgate/internal/store"
)

func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Guided first-time setup — config, login, API key, start server",
		Long: `Interactive wizard that walks through the full ShellGate setup:
  1. Create config file
  2. Login to CLI providers
  3. Create your first API key
  4. Start the server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Welcome to ShellGate Setup")
			fmt.Println("────────────────────────────────────────")

			// Step 1: Config
			fmt.Println("\n[1/4] Configure ShellGate")
			_, err := runInit(configPath, false)
			if err != nil {
				return err
			}

			// Step 2: Provider login
			fmt.Println("\n[2/4] Connect CLI Providers")
			statuses := allProviderStatuses()
			if err := runProviderLoginWizard(statuses); err != nil {
				return err
			}

			// Step 3: API key
			fmt.Println("\n[3/4] Create API Key")
			key, err := createFirstKey(configPath)
			if err != nil {
				return fmt.Errorf("create API key: %w", err)
			}

			// Step 4: Start server
			fmt.Println("\n[4/4] Start Server")
			var startNow bool
			err = huh.NewConfirm().
				Title("Start ShellGate in background now?").
				Value(&startNow).
				Run()
			if err != nil {
				return err
			}

			if startNow {
				if err := runDetached(); err != nil {
					return err
				}
			}

			cfg, _ := config.Load(configPath)
			port := 8080
			if cfg != nil {
				port = cfg.Server.Port
			}

			fmt.Println()
			fmt.Println("╔══════════════════════════════════════════════╗")
			fmt.Println("║           ShellGate is ready!                ║")
			fmt.Println("╚══════════════════════════════════════════════╝")
			fmt.Println()
			fmt.Printf("  API endpoint:  http://localhost:%d/v1\n", port)
			fmt.Printf("  API key:       %s\n", key)
			fmt.Println()
			fmt.Println("  Model routing:")
			fmt.Println("    gpt-5.4, gpt-5.5, ...           → Codex")
			fmt.Println("    kimi-code/kimi-for-coding        → Kimi")
			fmt.Println()
			fmt.Println("  shellgate status   — check status")
			fmt.Println("  shellgate logs     — view logs")
			fmt.Println("  shellgate stop     — stop server")
			return nil
		},
	}
}

func runProviderLoginWizard(statuses []providerStatus) error {
	// Show current status
	for _, s := range statuses {
		icon := "✗"
		if s.Connected {
			icon = "✓"
		}
		fmt.Printf("  %s %-8s %s\n", icon, s.Name, s.Detail)
	}
	fmt.Println()

	// Build options for providers not yet connected
	var toLogin []string
	var options []huh.Option[string]
	for _, s := range statuses {
		label := s.Name
		if s.Connected {
			label += " (already authenticated)"
		} else if s.Detail == "not installed" {
			label += " (not installed — skip)"
		}
		options = append(options, huh.NewOption(label, s.Name))
		if !s.Connected && s.Detail != "not installed" {
			toLogin = append(toLogin, s.Name)
		}
	}

	var selected []string
	err := huh.NewMultiSelect[string]().
		Title("Which providers do you want to login to?").
		Description("Space to toggle, Enter to confirm").
		Options(options...).
		Value(&selected).
		Run()
	if err != nil {
		return err
	}

	_ = toLogin
	for _, name := range selected {
		p, ok := providers[name]
		if !ok {
			continue
		}
		if _, err := exec.LookPath(p.binary); err != nil {
			fmt.Printf("  %s not found in PATH — skipping\n", p.binary)
			continue
		}
		fmt.Printf("\nLogging in to %s...\n\n", name)
		loginCmd := exec.Command(p.binary, p.loginArgs...)
		loginCmd.Stdin = os.Stdin
		loginCmd.Stdout = os.Stdout
		loginCmd.Stderr = os.Stderr
		if err := loginCmd.Run(); err != nil {
			fmt.Printf("  Warning: %s login failed: %s\n", name, err)
		} else {
			fmt.Printf("  ✓ %s authenticated\n", name)
		}
	}
	return nil
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

	var keyName string
	err = huh.NewInput().
		Title("API key name").
		Placeholder("default").
		Value(&keyName).
		Run()
	if err != nil {
		return "", err
	}
	keyName = strings.TrimSpace(keyName)
	if keyName == "" {
		keyName = "default"
	}

	entry, err := ks.Create(keyName)
	if err != nil {
		return "", err
	}
	fmt.Printf("  ✓ API key created: %s\n", entry.Key)
	return entry.Key, nil
}
