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
	"governance-api/internal/finops"
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

	aiProviders := map[string]provider.Provider{
		"mock": provider.NewMock(),
	}

	aiRoutes := map[string]airouter.Route{
		"fast-general": {
			RoutedModel: "mock-fast-general",
			Provider:    "mock",
			Reason:      "default local mock route",
		},
	}

	pricingCatalog, err := finops.NewStaticCatalog(
		[]finops.Rate{
			{
				Provider:            "mock",
				Model:               "mock-fast-general",
				InputPerMillionUSD:  0,
				OutputPerMillionUSD: 0,
			},
		},
	)
	if err != nil {
		logger.Error(
			"FinOps pricing catalog initialization failed",
			"error", err,
		)
		os.Exit(1)
	}

	costCalculator, err := finops.NewCalculator(
		pricingCatalog,
	)
	if err != nil {
		logger.Error(
			"FinOps cost calculator initialization failed",
			"error", err,
		)
		os.Exit(1)
	}

	switch os.Getenv("AI_PROVIDER") {
	case "", "mock":
		logger.Info(
			"AI provider configured",
			"provider", "mock",
		)

	case "azure-openai":
		endpoint := os.Getenv(
			"AZURE_OPENAI_ENDPOINT",
		)
		deployment := os.Getenv(
			"AZURE_OPENAI_DEPLOYMENT",
		)
		managedIdentityClientID := os.Getenv(
			"AZURE_CLIENT_ID",
		)

		if endpoint == "" ||
			deployment == "" ||
			managedIdentityClientID == "" {
			logger.Error(
				"Azure OpenAI configuration incomplete",
			)
			os.Exit(1)
		}

		azureOpenAIProvider, err :=
			provider.NewAzureOpenAIManagedIdentity(
				endpoint,
				managedIdentityClientID,
			)
		if err != nil {
			logger.Error(
				"create Azure OpenAI provider",
				"error", err,
			)
			os.Exit(1)
		}

		aiProviders["azure-openai"] =
			azureOpenAIProvider

		aiRoutes["fast-general"] =
			airouter.Route{
				RoutedModel: deployment,
				Provider:    "azure-openai",
				Reason:      "default Azure OpenAI route",
			}

		logger.Info(
			"AI provider configured",
			"provider", "azure-openai",
			"deployment", deployment,
		)

	default:
		logger.Error(
			"unsupported AI_PROVIDER",
			"provider",
			os.Getenv("AI_PROVIDER"),
		)
		os.Exit(1)
	}

	aiService := airouter.NewService(
		governanceService,
		aiRepository,
		costCalculator,
		aiProviders,
		aiRoutes,
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
