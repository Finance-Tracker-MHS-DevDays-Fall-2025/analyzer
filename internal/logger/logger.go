package logger

import (
	"io"
	"log/slog"
	"os"
)

func New(level string, logFile string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	if logFile == "" {
		handler := slog.NewJSONHandler(os.Stdout, opts)
		return slog.New(handler)
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		handler := slog.NewTextHandler(os.Stdout, opts)
		return slog.New(handler)
	}

	multiWriter := io.MultiWriter(os.Stdout, file)
	handler := slog.NewTextHandler(multiWriter, opts)
	return slog.New(handler)
}
