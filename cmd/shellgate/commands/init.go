package commands

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

const configTemplate = `[server]
host = "0.0.0.0"
port = {{.Port}}
read_timeout = "30s"
write_timeout = "120s"

[auth]
admin_secret = "{{.AdminSecret}}"
keys_file = "keys.json"

[executor]
codex_binary = "codex"
default_sandbox = "read-only"
timeout = "120s"
working_dir = ""

[logging]
level = "info"
format = "json"
`

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Interactive setup — create config.toml",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(configPath); err == nil {
				fmt.Printf("Config already exists at %s. Overwrite? [y/N] ", configPath)
				var answer string
				fmt.Scanln(&answer)
				if strings.ToLower(strings.TrimSpace(answer)) != "y" {
					fmt.Println("Aborted.")
					return nil
				}
			}

			reader := bufio.NewReader(os.Stdin)

			fmt.Print("Port [8080]: ")
			portStr, _ := reader.ReadString('\n')
			portStr = strings.TrimSpace(portStr)
			if portStr == "" {
				portStr = "8080"
			}

			fmt.Print("Admin secret (leave blank to generate): ")
			secret, _ := reader.ReadString('\n')
			secret = strings.TrimSpace(secret)
			if secret == "" {
				b := make([]byte, 16)
				rand.Read(b)
				secret = "sg-admin-" + hex.EncodeToString(b)
			}

			f, err := os.Create(configPath)
			if err != nil {
				return fmt.Errorf("create config: %w", err)
			}
			defer f.Close()

			tmpl := template.Must(template.New("config").Parse(configTemplate))
			if err := tmpl.Execute(f, map[string]string{
				"Port":        portStr,
				"AdminSecret": secret,
			}); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			fmt.Printf("\nConfig written to %s\n", configPath)
			fmt.Printf("Admin secret: %s\n\n", secret)
			fmt.Println("Next steps:")
			fmt.Println("  shellgate login codex   # authenticate your CLI provider")
			fmt.Println("  shellgate serve         # start the API server")
			return nil
		},
	}
}
