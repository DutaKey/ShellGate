package commands

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
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
keys_file = "{{.KeysFile}}"

[executor]
# Provider to use: "codex" or "kimi"
provider = "{{.Provider}}"
codex_binary = "codex"
default_sandbox = "read-only"
timeout = "120s"
working_dir = ""
kimi_binary = "kimi"

[logging]
level = "info"
format = "json"
`

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Interactive setup — create config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := runInit(configPath, false)
			return err
		},
	}
}

func runInit(cfgPath string, skipExistsCheck bool) (string, error) {
	if !skipExistsCheck {
		if _, err := os.Stat(cfgPath); err == nil {
			fmt.Printf("Config already exists at %s. Overwrite? [y/N] ", cfgPath)
			var answer string
			fmt.Scanln(&answer)
			if strings.ToLower(strings.TrimSpace(answer)) != "y" {
				fmt.Println("Aborted.")
				return "", nil
			}
		}
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Provider [codex/kimi] (default: codex): ")
	provider, _ := reader.ReadString('\n')
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" || (provider != "codex" && provider != "kimi") {
		provider = "codex"
	}

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

	keysFile := filepath.Join(shellgateDir(), "keys.json")

	if err := os.MkdirAll(filepath.Dir(cfgPath), 0700); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}

	f, err := os.Create(cfgPath)
	if err != nil {
		return "", fmt.Errorf("create config: %w", err)
	}
	defer f.Close()

	tmpl := template.Must(template.New("config").Parse(configTemplate))
	if err := tmpl.Execute(f, map[string]string{
		"Port":        portStr,
		"AdminSecret": secret,
		"Provider":    provider,
		"KeysFile":    keysFile,
	}); err != nil {
		return "", fmt.Errorf("write config: %w", err)
	}

	fmt.Printf("\nConfig written to %s\n", cfgPath)
	fmt.Printf("Admin secret: %s\n", secret)
	fmt.Printf("Provider:     %s\n\n", provider)
	fmt.Println("Next steps:")
	fmt.Printf("  shellgate login %s   # authenticate your CLI provider\n", provider)
	fmt.Println("  shellgate serve      # start the API server")

	return provider, nil
}
