package commands

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/dutakey/shellgate/config"
	"github.com/dutakey/shellgate/internal/api"
	"github.com/dutakey/shellgate/internal/store"
)

func newServeCmd() *cobra.Command {
	var detach bool

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the ShellGate API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if detach {
				return runDetached()
			}
			return runForeground()
		},
	}

	cmd.Flags().BoolVarP(&detach, "detach", "d", false, "run in background (detached)")
	return cmd
}

func runForeground() error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := buildLogger(cfg.Logging.Level, cfg.Logging.Format)
	defer logger.Sync()

	keys, err := store.NewKeyStore(cfg.Auth.KeysFile)
	if err != nil {
		return fmt.Errorf("load key store: %w", err)
	}

	srv := api.NewServer(cfg, keys, logger)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	<-quit
	return srv.Shutdown(15 * time.Second)
}

func runDetached() error {
	logFile := logFilePath()
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer f.Close()

	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}

	args := []string{"serve", "--config", configPath}
	proc := exec.Command(self, args...)
	proc.Stdout = f
	proc.Stderr = f
	proc.Stdin = nil

	proc.SysProcAttr = newSysProcAttr()

	if err := proc.Start(); err != nil {
		return fmt.Errorf("start background process: %w", err)
	}

	pid := proc.Process.Pid
	fmt.Printf("ShellGate running in background (PID %d)\n", pid)
	fmt.Printf("Logs:   %s\n", logFile)
	fmt.Printf("Status: shellgate status\n")
	fmt.Printf("Stop:   shellgate stop\n")

	return os.WriteFile(pidFilePath(), []byte(fmt.Sprintf("%d", pid)), 0644)
}

func buildLogger(level, format string) *zap.Logger {
	var cfg zap.Config
	if format == "json" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}
	switch level {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "warn":
		cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	logger, _ := cfg.Build()
	return logger
}
