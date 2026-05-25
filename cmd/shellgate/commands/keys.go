package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/dutakey/shellgate/config"
	"github.com/dutakey/shellgate/internal/store"
)

func newKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage ShellGate API keys",
		// No args = launch interactive manager
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKeysInteractive()
		},
	}

	cmd.AddCommand(
		newKeysCreateCmd(),
		newKeysListCmd(),
		newKeysRevokeCmd(),
	)

	return cmd
}

func runKeysInteractive() error {
	for {
		ks, err := loadKeyStore()
		if err != nil {
			return err
		}
		keys := ks.List()

		var action string
		options := []huh.Option[string]{
			huh.NewOption("Create new key", "create"),
		}
		for _, k := range keys {
			lastUsed := "never"
			if k.LastUsedAt != nil {
				lastUsed = "used " + k.LastUsedAt.Format("2006-01-02")
			}
			label := fmt.Sprintf("Revoke: %-20s %s  [%s]", k.Name, k.Key[:20]+"...", lastUsed)
			options = append(options, huh.NewOption(label, "revoke:"+k.ID+":"+k.Name))
		}
		options = append(options, huh.NewOption("Exit", "exit"))

		err = huh.NewSelect[string]().
			Title("API Key Manager").
			Description(fmt.Sprintf("%d key(s) active", len(keys))).
			Options(options...).
			Value(&action).
			Run()
		if err != nil {
			return nil // Ctrl+C = exit
		}

		switch {
		case action == "exit":
			return nil

		case action == "create":
			var name string
			err = huh.NewInput().
				Title("Key name").
				Placeholder("my-project").
				Value(&name).
				Run()
			if err != nil {
				continue
			}
			name = strings.TrimSpace(name)
			if name == "" {
				name = "unnamed"
			}
			key, err := ks.Create(name)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				continue
			}
			fmt.Printf("\n  ✓ Created key %q\n", key.Name)
			fmt.Printf("  Key: %s\n\n", key.Key)
			fmt.Println("  Save this key — it cannot be recovered.")
			fmt.Println()

		case strings.HasPrefix(action, "revoke:"):
			parts := strings.SplitN(action, ":", 3)
			if len(parts) < 3 {
				continue
			}
			id, name := parts[1], parts[2]

			var confirm bool
			err = huh.NewConfirm().
				Title(fmt.Sprintf("Revoke key %q?", name)).
				Description("This will immediately invalidate the key.").
				Affirmative("Revoke").
				Negative("Cancel").
				Value(&confirm).
				Run()
			if err != nil || !confirm {
				continue
			}

			if !ks.Revoke(id) {
				fmt.Printf("  Key not found: %s\n", id)
				continue
			}
			fmt.Printf("  ✓ Key %q revoked\n\n", name)
		}
	}
}

func newKeysCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new API key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ks, err := loadKeyStore()
			if err != nil {
				return err
			}
			key, err := ks.Create(args[0])
			if err != nil {
				return fmt.Errorf("create key: %w", err)
			}
			fmt.Printf("Created key for %q\n\n", key.Name)
			fmt.Printf("  Key: %s\n", key.Key)
			fmt.Printf("  ID:  %s\n", key.ID)
			fmt.Println("\nSave this key — it will not be shown again in plain form.")
			return nil
		},
	}
}

func newKeysListCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all API keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			ks, err := loadKeyStore()
			if err != nil {
				return err
			}
			keys := ks.List()
			if len(keys) == 0 {
				fmt.Println("No keys found. Run: shellgate keys create <name>")
				return nil
			}
			if jsonOut {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(keys)
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tID\tCREATED\tLAST USED")
			for _, k := range keys {
				lastUsed := "never"
				if k.LastUsedAt != nil {
					lastUsed = k.LastUsedAt.Format(time.RFC3339)
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					k.Name, k.ID, k.CreatedAt.Format(time.RFC3339), lastUsed)
			}
			w.Flush()
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")
	return cmd
}

func newKeysRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <id>",
		Short: "Revoke an API key by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ks, err := loadKeyStore()
			if err != nil {
				return err
			}
			if !ks.Revoke(args[0]) {
				return fmt.Errorf("key %q not found", args[0])
			}
			fmt.Printf("Key %q revoked.\n", args[0])
			return nil
		},
	}
}

func loadKeyStore() (*store.KeyStore, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return store.NewKeyStore(cfg.Auth.KeysFile)
}
