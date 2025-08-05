package shutdownsetup

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hot-coffee/pkg/logger"
)

// SetupGracefulShutdown handles graceful server shutdown
func SetupGracefulShutdown(server *http.Server, logger *logger.Logger) {
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	logger.Info("Received shutdown signal", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info("Shutting down server gracefully...")

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		return
	}

	logger.Info("Server shutdown completed successfully")
}
