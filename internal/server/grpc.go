package server

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/config"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/handler"
	pb "github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/pkg/api/proto/analyzer"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	cfg             *config.ServerConfig
	logger          *slog.Logger
	analyzerHandler *handler.AnalyzerHandler
	grpcServer      *grpc.Server
}

func NewGRPCServer(
	cfg *config.ServerConfig,
	analyzerHandler *handler.AnalyzerHandler,
	logger *slog.Logger,
) *GRPCServer {
	return &GRPCServer{
		cfg:             cfg,
		analyzerHandler: analyzerHandler,
		logger:          logger.With("component", "grpc_server"),
	}
}

func (s *GRPCServer) Run() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port))
	if err != nil {
		s.logger.Error("failed to create listener", "error", err)
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(UnaryServerInterceptor(s.logger)),
	)

	pb.RegisterAnalyzerServiceServer(s.grpcServer, s.analyzerHandler)

	s.logger.Info("gRPC server starting", "host", s.cfg.Host, "port", s.cfg.Port)

	if err := s.grpcServer.Serve(listener); err != nil {
		s.logger.Error("grpc server stopped with error", "error", err)
		return fmt.Errorf("failed to serve grpc: %w", err)
	}

	return nil
}

func (s *GRPCServer) Stop() {
	if s.grpcServer != nil {
		s.logger.Info("stopping gRPC server")
		s.grpcServer.GracefulStop()
		s.logger.Info("gRPC server stopped")
	}
}
