package logger

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// LogLevel represents logging levels
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// Config holds essential logger configuration for PostgreSQL operations
type Config struct {
	Level        LogLevel `json:"level"`
	Format       string   `json:"format"`        // "json", "text"
	Output       string   `json:"output"`        // "stdout", "stderr", file path
	EnableCaller bool     `json:"enable_caller"` // Include file and line info
	Component    string   `json:"component"`     // Default component name
	Environment  string   `json:"environment"`   // Environment (dev, staging, prod)
}

// Logger wraps slog.Logger with PostgreSQL-focused functionality
type Logger struct {
	*slog.Logger
	config Config
	output io.Writer
}

// RequestContext holds essential request-specific logging context
type RequestContext struct {
	RequestID  string        `json:"request_id"`
	Method     string        `json:"method"`
	Path       string        `json:"path"`
	RemoteAddr string        `json:"remote_addr"`
	StartTime  time.Time     `json:"start_time"`
	Duration   time.Duration `json:"duration,omitempty"`
	StatusCode int           `json:"status_code,omitempty"`
}

// DefaultConfig returns a default logger configuration optimized for PostgreSQL operations
func DefaultConfig() Config {
	return Config{
		Level:        LevelInfo,
		Format:       "json",
		Output:       "stdout",
		EnableCaller: true,
		Environment:  "development",
	}
}

// New creates a simplified logger instance optimized for PostgreSQL operations
func New(config Config) *Logger {
	var level slog.Level
	switch config.Level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Determine output writer
	var output io.Writer
	switch config.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// File output
		if file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666); err == nil {
			output = file
		} else {
			output = os.Stdout // Fallback to stdout
		}
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: false, // We'll handle source manually
	}

	var handler slog.Handler
	switch config.Format {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	case "text":
		handler = slog.NewTextHandler(output, opts)
	default:
		handler = slog.NewJSONHandler(output, opts)
	}

	// Wrap with custom source handler if caller info is enabled
	if config.EnableCaller {
		handler = newCustomSourceHandler(handler, 0)
	}

	slogLogger := slog.New(handler)

	// Add default context
	if config.Component != "" {
		slogLogger = slogLogger.With("component", config.Component)
	}
	if config.Environment != "" {
		slogLogger = slogLogger.With("environment", config.Environment)
	}

	logger := &Logger{
		Logger: slogLogger,
		config: config,
		output: output,
	}

	return logger
}

// WithContext creates a new logger with additional context
func (l *Logger) WithContext(args ...interface{}) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
		config: l.config,
		output: l.output,
	}
}

// WithComponent creates a logger with component context
func (l *Logger) WithComponent(component string) *Logger {
	return l.WithContext("component", component)
}

// WithRequest creates a logger with request context
func (l *Logger) WithRequest(ctx *RequestContext) *Logger {
	return l.WithContext(
		"request_id", ctx.RequestID,
		"method", ctx.Method,
		"path", ctx.Path,
		"remote_addr", ctx.RemoteAddr,
	)
}

// Debug logs at debug level - essential for database operations
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.Logger.Debug(msg, args...)
}

// Info logs at info level - essential for database operations
func (l *Logger) Info(msg string, args ...interface{}) {
	l.Logger.Info(msg, args...)
}

// Warn logs at warn level - essential for database operations
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.Logger.Warn(msg, args...)
}

// Error logs at error level with caller information - critical for database errors
func (l *Logger) Error(msg string, args ...interface{}) {
	// Add caller information for errors
	if l.config.EnableCaller {
		if _, file, line, ok := runtime.Caller(1); ok {
			args = append(args, "caller", fmt.Sprintf("%s:%d", filepath.Base(file), line))
		}
	}
	l.Logger.Error(msg, args...)
}

// Fatal logs at error level and exits - for critical database failures
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.Error(msg, args...)
	time.Sleep(100 * time.Millisecond)
	os.Exit(1)
}

// LogRequest logs HTTP request information
func (l *Logger) LogRequest(ctx *RequestContext) {
	l.WithRequest(ctx).Info("HTTP request started",
		"start_time", ctx.StartTime,
	)
}

// LogResponse logs HTTP response information
func (l *Logger) LogResponse(ctx *RequestContext) {
	duration := time.Since(ctx.StartTime)
	ctx.Duration = duration

	logLevel := "info"
	if ctx.StatusCode >= 400 {
		logLevel = "error"
	} else if ctx.StatusCode >= 300 {
		logLevel = "warn"
	}

	logger := l.WithRequest(ctx)
	args := []interface{}{
		"status_code", ctx.StatusCode,
		"duration_ms", duration.Milliseconds(),
	}

	switch logLevel {
	case "error":
		logger.Error("HTTP request completed", args...)
	case "warn":
		logger.Warn("HTTP request completed", args...)
	default:
		logger.Info("HTTP request completed", args...)
	}
}

// HTTPMiddleware returns a standard HTTP middleware for request logging
func (l *Logger) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate request ID if not present
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Create response writer wrapper to capture response details
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		ctx := &RequestContext{
			RequestID:  requestID,
			Method:     r.Method,
			Path:       r.URL.Path,
			RemoteAddr: getClientIP(r),
			StartTime:  start,
		}

		// Log request start
		l.LogRequest(ctx)

		// Process request
		next.ServeHTTP(rw, r)

		// Log response
		ctx.StatusCode = rw.statusCode
		l.LogResponse(ctx)
	})
}

// responseWriter wraps http.ResponseWriter to capture response details
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.size += size
	return size, err
}

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if ips := strings.Split(xForwardedFor, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
		return xRealIP
	}

	// Fall back to RemoteAddr
	if ip := strings.Split(r.RemoteAddr, ":"); len(ip) > 0 {
		return ip[0]
	}

	return r.RemoteAddr
}

// Close properly closes the logger and any file handles
func (l *Logger) Close() error {
	if closer, ok := l.output.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
