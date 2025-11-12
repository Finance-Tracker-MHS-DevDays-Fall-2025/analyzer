package server

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
)

func UnaryServerInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		logger.Info("grpc request started",
			"method", info.FullMethod)

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		if err != nil {
			logger.Error("grpc request failed",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
				"error", err)
		} else {
			logger.Info("grpc request completed",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds())
		}

		return resp, err
	}
}
