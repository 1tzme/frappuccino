package main

// TODO: Transition State: JSON → PostgreSQL
// DEPRECATED: Replace JSON-based repositories with PostgreSQL-backed repositories
// 2. Replace orderRepo := repositories.NewOrderRepository(appLogger, flagConfig.DataDir)
//    with orderRepo := repositories.NewOrderRepository(appLogger, db)
// 3. Replace menuRepo := repositories.NewMenuRepository(appLogger, flagConfig.DataDir)
//    with menuRepo := repositories.NewMenuRepository(appLogger, db)
// 4. Replace inventoryRepo := repositories.NewInventoryRepository(appLogger, flagConfig.DataDir)
//    with inventoryRepo := repositories.NewInventoryRepository(appLogger, db)
// 5. Remove flagConfig.DataDir dependency entirely

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hot-coffee/internal/handler"
	"hot-coffee/internal/repositories"
	"hot-coffee/internal/router"
	"hot-coffee/internal/service"
	"hot-coffee/pkg/database"
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
		Level:        envconfig.GetLogLevel(),
		Format:       envconfig.GetEnv("LOG_FORMAT", "json"),
		Output:       envconfig.GetEnv("LOG_OUTPUT", "stdout"),
		EnableCaller: envconfig.GetEnv("LOG_ENABLE_CALLER", "true") == "true",
		Environment:  envconfig.GetEnv("ENVIRONMENT", "development"),
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

	// TODO: Transition State: JSON → PostgreSQL
	// Initialize database connection to replace JSON file storage

	// Parse database port with default value
	dbPort := 5432
	if portStr := envconfig.GetEnv("DB_PORT", "5432"); portStr != "" {
		if parsedPort, err := strconv.Atoi(portStr); err == nil {
			dbPort = parsedPort
		}
	}

	dbConfig := database.Config{
		Host:     envconfig.GetEnv("DB_HOST", "localhost"),
		Port:     dbPort,
		User:     envconfig.GetEnv("DB_USER", "postgres"),
		Password: envconfig.GetEnv("DB_PASSWORD", ""),
		DBName:   envconfig.GetEnv("DB_NAME", "hotcoffee"),
		SSLMode:  envconfig.GetEnv("DB_SSL_MODE", "disable"),
		// Use default connection pool settings from database package
	}

	// Establish database connection
	db, err := database.NewConnection(dbConfig, appLogger)
	if err != nil {
		appLogger.Error("Failed to establish database connection", "error", err)
		// For now, fall back to JSON storage during transition
		appLogger.Warn("Falling back to JSON storage during transition period")
		db = nil
	} else {
		appLogger.Info("Database connection established successfully")

		// Perform initial health check
		if err := db.HealthCheck(); err != nil {
			appLogger.Error("Database health check failed", "error", err)
			appLogger.Warn("Continuing with JSON storage due to database health issues")
			db.Close()
			db = nil
		} else {
			appLogger.Info("Database health check passed")
		}

		// Ensure database connection is closed on shutdown
		if db != nil {
			defer func() {
				if err := db.Close(); err != nil {
					appLogger.Error("Failed to close database connection", "error", err)
				}
			}()
		}
	}

	// Initialize repositories with logger and data directory from flags
	// TODO: Transition State: JSON → PostgreSQL - Replace these with database-backed repositories
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
