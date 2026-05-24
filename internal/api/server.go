package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/dutakey/shellgate/config"
	"github.com/dutakey/shellgate/internal/api/handler"
	"github.com/dutakey/shellgate/internal/executor"
	"github.com/dutakey/shellgate/internal/middleware"
	"github.com/dutakey/shellgate/internal/store"
)

type Server struct {
	http   *http.Server
	logger *zap.Logger
}

func newExecutor(cfg *config.Config) executor.Executor {
	codex := executor.NewCodexExecutor(cfg.Executor.CodexBinary, cfg.Executor.Sandbox, cfg.Executor.WorkingDir)
	kimi := executor.NewKimiExecutor(cfg.Executor.KimiBinary, cfg.Executor.WorkingDir)

	return executor.NewRouterExecutor(codex,
		executor.Route{Prefix: "kimi", Executor: kimi},
	)
}

func NewServer(cfg *config.Config, keys *store.KeyStore, logger *zap.Logger) *Server {
	exec := newExecutor(cfg)

	chatHandler := handler.NewChatHandler(exec, logger)
	responsesHandler := handler.NewResponsesHandler(exec, logger)
	adminHandler := handler.NewAdminHandler(keys)
	modelsList, modelsById := handler.NewModelsHandlers()

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(requestLogger(logger))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// OpenAI-compatible API — protected by ShellGate API key
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(keys))
		r.Route("/v1", func(r chi.Router) {
			r.Post("/chat/completions", chatHandler.ChatCompletions())
			r.Post("/responses", responsesHandler.Create())
			r.Get("/models", modelsList)
			r.Get("/models/{id}", modelsById)
		})
	})

	// Admin API — protected by admin secret
	r.Group(func(r chi.Router) {
		r.Use(middleware.AdminAuth(cfg.Auth.AdminSecret))
		r.Route("/admin", func(r chi.Router) {
			r.Post("/keys", adminHandler.CreateKey())
			r.Get("/keys", adminHandler.ListKeys())
			r.Delete("/keys/{id}", adminHandler.RevokeKey())
		})
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	return &Server{
		http: &http.Server{
			Addr:         addr,
			Handler:      r,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		},
		logger: logger,
	}
}

func (s *Server) Start() error {
	s.logger.Info("ShellGate starting", zap.String("addr", s.http.Addr))
	return s.http.ListenAndServe()
}

func requestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("request", zap.String("method", r.Method), zap.String("path", r.URL.Path))
			next.ServeHTTP(w, r)
		})
	}
}

func (s *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	s.logger.Info("ShellGate shutting down")
	return s.http.Shutdown(ctx)
}
