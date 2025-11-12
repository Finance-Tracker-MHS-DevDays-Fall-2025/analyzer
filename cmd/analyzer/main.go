package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/config"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/database"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/handler"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/logger"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/server"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/service"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/storage"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	log := logger.New(cfg.Log.Level)
	log.Info("application starting", "config", *configPath)
	log.Info("database configuration",
		"host", cfg.DB.Host,
		"port", cfg.DB.Port,
		"user", cfg.DB.User,
		"dbname", cfg.DB.DBName,
		"sslmode", cfg.DB.SSLMode)

	ctx := context.Background()

	db, err := database.New(ctx, &cfg.DB)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	log.Info("database connected successfully")

	transactionStorage := storage.NewPostgresStorage(db.Pool())

	analyzerService := service.NewAnalyzerService(transactionStorage, log)

	analyzerHandler := handler.NewAnalyzerHandler(analyzerService, log)

	grpcServer := server.NewGRPCServer(&cfg.Server, analyzerHandler, log)
	debugServer := server.NewDebugServer(&cfg.Server, db, log)

	srv := server.New(cfg, log, db, grpcServer, debugServer)

	go func() {
		if err := srv.Run(); err != nil {
			log.Error("server run failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutdown signal received")
	srv.Stop()
	log.Info("application stopped gracefully")
}
