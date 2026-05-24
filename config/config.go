package config

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server   ServerConfig   `toml:"server"`
	Auth     AuthConfig     `toml:"auth"`
	Executor ExecutorConfig `toml:"executor"`
	Logging  LoggingConfig  `toml:"logging"`
}

type ServerConfig struct {
	Host         string        `toml:"host"`
	Port         int           `toml:"port"`
	ReadTimeout  time.Duration `toml:"read_timeout"`
	WriteTimeout time.Duration `toml:"write_timeout"`
}

type AuthConfig struct {
	AdminSecret string `toml:"admin_secret"`
	KeysFile    string `toml:"keys_file"`
}

type ExecutorConfig struct {
	Provider string `toml:"provider"` // "codex" (default) | "kimi"

	CodexBinary string        `toml:"codex_binary"`
	Sandbox     string        `toml:"default_sandbox"`
	Timeout     time.Duration `toml:"timeout"`
	WorkingDir  string        `toml:"working_dir"`

	KimiBinary string `toml:"kimi_binary"`
}

type LoggingConfig struct {
	Level  string `toml:"level"`
	Format string `toml:"format"`
}

func Load(path string) (*Config, error) {
	cfg := defaults()

	if path != "" {
		if _, err := os.Stat(path); err == nil {
			if _, err := toml.DecodeFile(path, cfg); err != nil {
				return nil, fmt.Errorf("parse config: %w", err)
			}
		}
	}

	applyEnvOverrides(cfg)

	if cfg.Auth.AdminSecret == "" {
		return nil, fmt.Errorf("auth.admin_secret is required (or SHELLGATE_ADMIN_SECRET env var)")
	}

	return cfg, nil
}

func defaults() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 120 * time.Second,
		},
		Auth: AuthConfig{
			KeysFile: "keys.json",
		},
		Executor: ExecutorConfig{
			Provider:    "codex",
			CodexBinary: "codex",
			Sandbox:     "read-only",
			Timeout:     120 * time.Second,
			KimiBinary:  "kimi",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("SHELLGATE_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Server.Port)
	}
	if v := os.Getenv("SHELLGATE_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("SHELLGATE_ADMIN_SECRET"); v != "" {
		cfg.Auth.AdminSecret = v
	}
	if v := os.Getenv("SHELLGATE_KEYS_FILE"); v != "" {
		cfg.Auth.KeysFile = v
	}
	if v := os.Getenv("SHELLGATE_EXECUTOR_PROVIDER"); v != "" {
		cfg.Executor.Provider = v
	}
	if v := os.Getenv("SHELLGATE_EXECUTOR_CODEX_BINARY"); v != "" {
		cfg.Executor.CodexBinary = v
	}
	if v := os.Getenv("SHELLGATE_EXECUTOR_SANDBOX"); v != "" {
		cfg.Executor.Sandbox = v
	}
	if v := os.Getenv("SHELLGATE_EXECUTOR_KIMI_BINARY"); v != "" {
		cfg.Executor.KimiBinary = v
	}
	if v := os.Getenv("SHELLGATE_LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
}
