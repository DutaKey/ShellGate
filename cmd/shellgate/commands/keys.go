package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/dutakey/shellgate/config"
	"github.com/dutakey/shellgate/internal/store"
)

func newKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage ShellGate API keys",
	}

	cmd.AddCommand(
		newKeysCreateCmd(),
		newKeysListCmd(),
		newKeysRevokeCmd(),
	)

	return cmd
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
			fmt.Printf("  Key:  %s\n", key.Key)
			fmt.Printf("  ID:   %s\n", key.ID)
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
				fmt.Println("No keys found. Create one: shellgate keys create <name>")
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
