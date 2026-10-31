package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lcaohoanq/bit-learning-be-v2/internal/app"
	"github.com/lcaohoanq/bit-learning-be-v2/internal/config"
	"github.com/lcaohoanq/bit-learning-be-v2/internal/database"
	db "github.com/lcaohoanq/bit-learning-be-v2/internal/database/db"
	"github.com/lcaohoanq/bit-learning-be-v2/internal/modules/auth"
	"github.com/lcaohoanq/bit-learning-be-v2/internal/modules/user"
	"github.com/lcaohoanq/bit-learning-be-v2/internal/telemetry"
)

const ScalarUi = "http://localhost:8080/docs"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	shutdownTracing, err := telemetry.Setup(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
	if err != nil {
		logger.Error("setup telemetry", "error", err)
		os.Exit(1)
	}
	defer func() { _ = shutdownTracing(context.Background()) }()

	if cfg.AutoMigrate {
		if err := database.Migrate(cfg.DatabaseURL); err != nil {
			logger.Error("migrate database", "error", err)
			os.Exit(1)
		}
	}
	pool, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	userRepository := user.NewRepository(db.New(pool))
	userService := user.NewService(userRepository)
	tokenManager := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTTTL)
	authService := auth.NewService(userService, tokenManager)
	authHandler := auth.NewHandler(authService, logger)
	userHandler := user.NewHandler(userService, logger)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           app.NewRouter(authHandler, userHandler, tokenManager, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		logger.Info("server started", "address", cfg.HTTPAddr, "environment", cfg.Environment)
		logger.Debug("Scalar UI: " + ScalarUi)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server stopped unexpectedly", "error", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
