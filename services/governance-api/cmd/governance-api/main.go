package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"governance-api/internal/airouter"
	"governance-api/internal/config"
	"governance-api/internal/database"
	"governance-api/internal/governance"
	"governance-api/internal/httpserver"
	"governance-api/internal/provider"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	governanceRepository := governance.NewPostgresRepository(db)
	governanceService := governance.NewService(governanceRepository)

	aiRepository := airouter.NewPostgresRepository(db)
	mockProvider := provider.NewMock()

	aiService := airouter.NewService(
		governanceService,
		aiRepository,
		map[string]provider.Provider{
			"mock": mockProvider,
		},
		map[string]airouter.Route{
			"fast-general": {
				RoutedModel: "mock-fast-general",
				Provider:    "mock",
				Reason:      "default Stage 5 mock route",
			},
		},
	)

	api := httpserver.NewWithAIRouter(
		logger,
		db,
		governanceService,
		aiService,
	)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		logger.Info(
			"governance-api started",
			"port", cfg.Port,
		)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-signalCtx.Done()

	logger.Info("shutdown requested")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("governance-api stopped")
}
