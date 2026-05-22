package commands

import (
	"errors"
	"fmt"
	"net/http"
	"os"
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
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the ShellGate API server",
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}
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
