package server

import (
	"fmt"
	"log/slog"

	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/config"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/database"
)

type Server struct {
	cfg         *config.Config
	logger      *slog.Logger
	db          *database.Database
	grpcServer  *GRPCServer
	debugServer *DebugServer
}

func New(
	cfg *config.Config,
	logger *slog.Logger,
	db *database.Database,
	grpcServer *GRPCServer,
	debugServer *DebugServer,
) *Server {
	return &Server{
		cfg:         cfg,
		logger:      logger,
		db:          db,
		grpcServer:  grpcServer,
		debugServer: debugServer,
	}
}

func (s *Server) Run() error {
	errChan := make(chan error, 2)

	go func() {
		s.logger.Info("starting gRPC server in goroutine")
		if err := s.grpcServer.Run(); err != nil {
			errChan <- fmt.Errorf("grpc server error: %w", err)
		}
	}()

	go func() {
		s.logger.Info("starting debug server in goroutine")
		if err := s.debugServer.Run(); err != nil {
			errChan <- fmt.Errorf("debug server error: %w", err)
		}
	}()

	err := <-errChan
	s.logger.Error("server component failed", "error", err)
	return err
}

func (s *Server) Stop() {
	s.logger.Info("shutting down all server components")

	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}

	if s.debugServer != nil {
		s.debugServer.Stop()
	}

	if s.db != nil {
		s.logger.Info("closing database connection")
		s.db.Close()
	}

	s.logger.Info("all components stopped")
}
