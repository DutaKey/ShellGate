package commands

import (
	"os"
	"path/filepath"
)

// shellgateDir returns ~/.shellgate, creating it if needed.
func shellgateDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	dir := filepath.Join(home, ".shellgate")
	_ = os.MkdirAll(dir, 0700)
	return dir
}

func defaultConfigPath() string {
	return filepath.Join(shellgateDir(), "config.toml")
}

func pidFilePath() string {
	return filepath.Join(shellgateDir(), "shellgate.pid")
}

func logFilePath() string {
	return filepath.Join(shellgateDir(), "shellgate.log")
}
