package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// NewConsoleHandler creates a custom console handler (simplified)
func NewConsoleHandler(w io.Writer, opts *slog.HandlerOptions, enableColors bool) slog.Handler {
	return slog.NewTextHandler(w, opts)
}

// customSourceHandler wraps a handler to fix source attribution
type customSourceHandler struct {
	handler slog.Handler
	depth   int
}

func newCustomSourceHandler(handler slog.Handler, depth int) *customSourceHandler {
	return &customSourceHandler{
		handler: handler,
		depth:   depth,
	}
}

func (h *customSourceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *customSourceHandler) Handle(ctx context.Context, r slog.Record) error {
	// Find the correct caller by skipping logger wrapper methods
	for i := 3; i < 8; i++ { // Start from 3 and go up to 8 to find the right caller
		if _, file, line, ok := runtime.Caller(i); ok {
			filename := filepath.Base(file)
			// Skip logger-related files and Go runtime files to find actual application code
			if filename != "logger.go" &&
				filename != "logger_helper.go" &&
				!strings.HasSuffix(filename, "_test.go") &&
				!strings.HasSuffix(filename, ".s") && // Skip assembly files
				!strings.Contains(file, "runtime/") { // Skip Go runtime files
				r.AddAttrs(slog.String("source", fmt.Sprintf("%s:%d", filename, line)))
				break
			}
		}
	}
	return h.handler.Handle(ctx, r)
}

func (h *customSourceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &customSourceHandler{
		handler: h.handler.WithAttrs(attrs),
		depth:   h.depth,
	}
}

func (h *customSourceHandler) WithGroup(name string) slog.Handler {
	return &customSourceHandler{
		handler: h.handler.WithGroup(name),
		depth:   h.depth,
	}
}

// Default returns a default logger instance
func Default() *Logger {
	return New(DefaultConfig())
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// loggerKey is the key for storing/retrieving logger in context
const loggerKey contextKey = "logger"

// FromContext extracts logger from context
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(loggerKey).(*Logger); ok {
		return logger
	}
	return Default()
}

// WithLogger adds logger to context
func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
