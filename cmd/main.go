package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"hot-coffee/internal/handler"
	"hot-coffee/internal/repositories"
	"hot-coffee/internal/router"
	"hot-coffee/internal/service"
	"hot-coffee/pkg/envconfig"
	"hot-coffee/pkg/flags"
	"hot-coffee/pkg/logger"
	"hot-coffee/pkg/shutdownsetup"
)

func main() {
	// Parse command-line flags
	flagConfig := flags.Parse()

	// Validate flag configuration
	if err := flagConfig.Validate(); err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		return
	}

	envErr := envconfig.LoadEnvFile(".env")

	loggerConfig := logger.Config{
		Level:         envconfig.GetLogLevel(),
		Format:        envconfig.GetEnv("LOG_FORMAT", "json"),
		Output:        envconfig.GetEnv("LOG_OUTPUT", "stdout"),
		EnableCaller:  envconfig.GetEnv("LOG_ENABLE_CALLER", "true") == "true",
		EnableColors:  envconfig.GetEnv("LOG_ENABLE_COLORS", "false") == "true",
		Environment:   envconfig.GetEnv("ENVIRONMENT", "development"),
		EnableMetrics: true,
		SensitiveKeys: []string{"password", "token", "secret", "key", "authorization"}, // abstract for now
	}

	appLogger := logger.New(loggerConfig)

	if envErr != nil {
		appLogger.Warn("Failed to load .env file", "error", envErr)
	} else {
		appLogger.Debug(".env file loaded successfully")
	}

	appLogger.Info("Starting Hot Coffee application",
		"environment", loggerConfig.Environment,
		"log_level", loggerConfig.Level)

	// Initialize repositories with logger and data directory from flags
	orderRepo := repositories.NewOrderRepository(appLogger, flagConfig.DataDir)
	menuRepo := repositories.NewMenuRepository(appLogger, flagConfig.DataDir)
	inventoryRepo := repositories.NewInventoryRepository(appLogger, flagConfig.DataDir)
	aggregationRepo := repositories.NewAggregationRepository(orderRepo, menuRepo, appLogger)

	// Initialize services with logger
	orderService := service.NewOrderService(orderRepo, menuRepo, inventoryRepo, appLogger)
	menuService := service.NewMenuService(inventoryRepo, menuRepo, orderRepo, appLogger)
	inventoryService := service.NewInventoryService(inventoryRepo, orderRepo, menuRepo, appLogger)
	aggregationService := service.NewAggregationService(aggregationRepo, appLogger)

	// Initialize handlers with logger
	orderHandler := handler.NewOrderHandler(orderService, appLogger)
	menuHandler := handler.NewMenuHandler(menuService, appLogger)
	inventoryHandler := handler.NewInventoryHandler(inventoryService, appLogger)
	aggregationHandler := handler.NewAggregationHandler(aggregationService, appLogger)

	mux := router.NewRouter(orderHandler, menuHandler, inventoryHandler, aggregationHandler)

	handler := appLogger.HTTPMiddleware(mux)

	initialPort := flagConfig.Port
	if initialPort == "" {
		initialPort = envconfig.GetEnv("PORT", "8080")
	}
	host := envconfig.GetEnv("HOST", "localhost")

	port := initialPort

	server := &http.Server{
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		server.Addr = host + ":" + port

		go func() {
			appLogger.Info("Starting HTTP server",
				"host", host,
				"port", port,
				"address", server.Addr)

			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				appLogger.Error("Server error", "error", err)
				serverErrors <- err
			}
		}()

		select {
		case err := <-serverErrors:
			if strings.Contains(err.Error(), "address already in use") && i < maxRetries-1 {
				portNum := 8080 + i + 1
				port = fmt.Sprintf("%d", portNum)
				appLogger.Warn("Port already in use, trying alternative port",
					"current_port", server.Addr,
					"next_port", port)
				continue
			} else {
				appLogger.Error("Failed to start server after retries", "error", err)
				return
			}
		case <-time.After(200 * time.Millisecond):
			appLogger.Info("Server started successfully", "port", port)
		}

		break
	}

	select {
	case err := <-serverErrors:
		appLogger.Error("Could not start server", "error", err)
		return
	default:
		shutdownsetup.SetupGracefulShutdown(server, appLogger)
	}
}
