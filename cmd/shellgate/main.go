package main

import (
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/dutakey/shellgate/config"
	"github.com/dutakey/shellgate/internal/api"
	"github.com/dutakey/shellgate/internal/store"
)

func main() {
	configPath := flag.String("config", "config.toml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	logger := buildLogger(cfg.Logging.Level, cfg.Logging.Format)
	defer logger.Sync()

	keys, err := store.NewKeyStore(cfg.Auth.KeysFile)
	if err != nil {
		logger.Fatal("failed to load key store", zap.Error(err))
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
	if err := srv.Shutdown(15 * time.Second); err != nil {
		logger.Error("shutdown error", zap.Error(err))
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
