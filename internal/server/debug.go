package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/config"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/database"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/handler"
)

type DebugServer struct {
	cfg        *config.ServerConfig
	db         *database.Database
	logger     *slog.Logger
	httpServer *http.Server
}

func NewDebugServer(cfg *config.ServerConfig, db *database.Database, logger *slog.Logger) *DebugServer {
	return &DebugServer{
		cfg:    cfg,
		db:     db,
		logger: logger.With("component", "debug_server"),
	}
}

func (s *DebugServer) Run() error {
	debugHandler := handler.NewDebugHandler(s.db, s.logger)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.DebugPort),
		Handler: debugHandler,
	}

	s.logger.Info("debug HTTP server starting", "host", s.cfg.Host, "port", s.cfg.DebugPort)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error("debug server stopped with error", "error", err)
		return fmt.Errorf("failed to serve debug http: %w", err)
	}

	return nil
}

func (s *DebugServer) Stop() {
	if s.httpServer != nil {
		s.logger.Info("stopping debug server")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("error during debug server shutdown", "error", err)
		} else {
			s.logger.Info("debug server stopped")
		}
	}
}
