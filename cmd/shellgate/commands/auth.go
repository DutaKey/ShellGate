package commands

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type providerStatus struct {
	Name      string
	Connected bool
	Detail    string
}

func checkProviderStatus(name string) providerStatus {
	s := providerStatus{Name: name}
	switch name {
	case "codex":
		p, ok := providers["codex"]
		if !ok {
			s.Detail = "unknown provider"
			return s
		}
		if _, err := exec.LookPath(p.binary); err != nil {
			s.Detail = "not installed"
			return s
		}
		out, err := exec.Command(p.binary, p.statusArgs...).CombinedOutput()
		if err != nil {
			s.Detail = "not authenticated"
			return s
		}
		line := strings.TrimSpace(string(out))
		if line == "" {
			line = "authenticated"
		}
		s.Connected = true
		s.Detail = line

	case "kimi":
		p, ok := providers["kimi"]
		if !ok {
			s.Detail = "unknown provider"
			return s
		}
		if _, err := exec.LookPath(p.binary); err != nil {
			s.Detail = "not installed"
			return s
		}
		// kimi has no login status command — check credentials file
		home, _ := os.UserHomeDir()
		credFile := filepath.Join(home, ".kimi", "credentials", "kimi-code.json")
		if _, err := os.Stat(credFile); err != nil {
			s.Detail = "not authenticated"
			return s
		}
		s.Connected = true
		s.Detail = "authenticated"

	case "claude":
		p, ok := providers["claude"]
		if !ok {
			s.Detail = "unknown provider"
			return s
		}
		if _, err := exec.LookPath(p.binary); err != nil {
			s.Detail = "not installed"
			return s
		}
		// claude stores OAuth credentials at ~/.claude/.credentials.json
		home, _ := os.UserHomeDir()
		credFile := filepath.Join(home, ".claude", ".credentials.json")
		data, err := os.ReadFile(credFile)
		if err != nil {
			s.Detail = "not authenticated"
			return s
		}
		var creds struct {
			ClaudeAIOauth *struct {
				AccessToken string `json:"accessToken"`
			} `json:"claudeAiOauth"`
		}
		if err := json.Unmarshal(data, &creds); err != nil || creds.ClaudeAIOauth == nil || creds.ClaudeAIOauth.AccessToken == "" {
			s.Detail = "not authenticated"
			return s
		}
		s.Connected = true
		s.Detail = "authenticated"
	}
	return s
}

func allProviderStatuses() []providerStatus {
	names := []string{"codex", "kimi", "claude"}
	out := make([]providerStatus, len(names))
	for i, name := range names {
		out[i] = checkProviderStatus(name)
	}
	return out
}
