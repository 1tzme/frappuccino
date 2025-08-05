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
	"sync"
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

// Config holds comprehensive logger configuration
type Config struct {
	Level         LogLevel `json:"level"`
	Format        string   `json:"format"`         // "json", "text", "console"
	Output        string   `json:"output"`         // "stdout", "stderr", file path
	EnableCaller  bool     `json:"enable_caller"`  // Include file and line info
	EnableColors  bool     `json:"enable_colors"`  // Enable colored output for console
	TimeFormat    string   `json:"time_format"`    // Custom time format
	Component     string   `json:"component"`      // Default component name
	Environment   string   `json:"environment"`    // Environment (dev, staging, prod)
	MaxFileSize   int64    `json:"max_file_size"`  // Max log file size in bytes
	MaxBackups    int      `json:"max_backups"`    // Max number of backup files
	EnableMetrics bool     `json:"enable_metrics"` // Enable performance metrics
	SensitiveKeys []string `json:"sensitive_keys"` // Keys to redact in logs
}

// Logger wraps slog.Logger with enhanced functionality
type Logger struct {
	*slog.Logger
	config    Config
	startTime time.Time
	mutex     sync.RWMutex
	output    io.Writer
	metrics   *LogMetrics
}

// LogMetrics tracks logging statistics
type LogMetrics struct {
	TotalLogs   int64            `json:"total_logs"`
	LogsByLevel map[string]int64 `json:"logs_by_level"`
	ErrorRate   float64          `json:"error_rate"`
	LastLogTime time.Time        `json:"last_log_time"`
	mutex       sync.RWMutex
}

// RequestContext holds request-specific logging context
type RequestContext struct {
	RequestID    string        `json:"request_id"`
	UserID       string        `json:"user_id,omitempty"`
	Method       string        `json:"method"`
	Path         string        `json:"path"`
	RemoteAddr   string        `json:"remote_addr"`
	UserAgent    string        `json:"user_agent"`
	StartTime    time.Time     `json:"start_time"`
	Duration     time.Duration `json:"duration,omitempty"`
	StatusCode   int           `json:"status_code,omitempty"`
	ResponseSize int64         `json:"response_size,omitempty"`
}

// Performance holds performance-related metrics (COMMENTED OUT - Future feature)
// type Performance struct {
// 	Operation      string        `json:"operation"`
// 	Duration       time.Duration `json:"duration"`
// 	MemoryUsage    uint64        `json:"memory_usage"`
// 	GoroutineCount int           `json:"goroutine_count"`
// }

// DefaultConfig returns a default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:         LevelInfo,
		Format:        "json",
		Output:        "stdout",
		EnableCaller:  true,
		EnableColors:  false,
		TimeFormat:    time.RFC822, //"02 Jan 06 15:04 MST"
		Environment:   "development",
		MaxFileSize:   100 * 1024 * 1024, // 100MB
		MaxBackups:    5,
		EnableMetrics: true,
		SensitiveKeys: []string{"password", "token", "secret", "key", "authorization"},
	}
}

// New creates a new enhanced logger instance
func New(config Config) *Logger {
	if config.TimeFormat == "" {
		config.TimeFormat = time.RFC822
	}

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
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Custom attribute replacement for sensitive data
			if contains(config.SensitiveKeys, strings.ToLower(a.Key)) {
				return slog.String(a.Key, "[REDACTED]")
			}

			// Custom time formatting
			if a.Key == slog.TimeKey && config.TimeFormat != "" {
				if t, ok := a.Value.Any().(time.Time); ok {
					return slog.String(a.Key, t.Format(config.TimeFormat))
				}
			}

			return a
		},
	}

	var handler slog.Handler
	switch config.Format {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	case "text":
		handler = slog.NewTextHandler(output, opts)
	case "console":
		handler = NewConsoleHandler(output, opts, config.EnableColors)
	default:
		handler = slog.NewJSONHandler(output, opts)
	}

	// Wrap with custom source handler if caller info is enabled
	if config.EnableCaller {
		handler = newCustomSourceHandler(handler, 0) // depth is handled inside the custom handler
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
		Logger:    slogLogger,
		config:    config,
		startTime: time.Now(),
		output:    output,
		metrics:   NewLogMetrics(),
	}

	return logger
}

// NewLogMetrics creates a new metrics tracker
func NewLogMetrics() *LogMetrics {
	return &LogMetrics{
		LogsByLevel: make(map[string]int64),
	}
}

// WithContext creates a new logger with additional context
func (l *Logger) WithContext(args ...interface{}) *Logger {
	return &Logger{
		Logger:    l.Logger.With(args...),
		config:    l.config,
		startTime: l.startTime,
		output:    l.output,
		metrics:   l.metrics,
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

// Debug logs at debug level with metrics tracking
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.trackMetrics("debug")
	l.Logger.Debug(msg, args...)
}

// Info logs at info level with metrics tracking
func (l *Logger) Info(msg string, args ...interface{}) {
	l.trackMetrics("info")
	l.Logger.Info(msg, args...)
}

// Warn logs at warn level with metrics tracking
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.trackMetrics("warn")
	l.Logger.Warn(msg, args...)
}

// Error logs at error level with metrics tracking
func (l *Logger) Error(msg string, args ...interface{}) {
	l.trackMetrics("error")

	// Add caller information for errors
	if l.config.EnableCaller {
		if _, file, line, ok := runtime.Caller(1); ok {
			args = append(args, "caller", fmt.Sprintf("%s:%d", filepath.Base(file), line))
		}
	}

	l.Logger.Error(msg, args...)
}

// Fatal logs at error level and exits the program with status code 1
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.Error(msg, args...)

	time.Sleep(100 * time.Millisecond)
	os.Exit(1)
}

// LogRequest logs HTTP request information
func (l *Logger) LogRequest(ctx *RequestContext) {
	l.WithRequest(ctx).Info("HTTP request started",
		"user_agent", ctx.UserAgent,
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
		"response_size", ctx.ResponseSize,
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

// LogPerformance logs performance metrics (COMMENTED OUT - Future feature)
// func (l *Logger) LogPerformance(perf Performance) {
// 	l.Debug("Performance metrics",
// 		"operation", perf.Operation,
// 		"duration_ms", perf.Duration.Milliseconds(),
// 		"memory_usage", perf.MemoryUsage,
// 		"goroutine_count", perf.GoroutineCount,
// 	)
// }

// LogBusinessEvent logs important business events
func (l *Logger) LogBusinessEvent(event string, entityID string, details map[string]interface{}) {
	args := []interface{}{
		"event_type", "business_event",
		"event", event,
		"entity_id", entityID,
		"timestamp", time.Now(),
	}

	for k, v := range details {
		args = append(args, k, v)
	}

	l.Info("Business event", args...)
}

// LogSecurity logs security-related events (COMMENTED OUT - Future feature)
// func (l *Logger) LogSecurity(event string, severity string, details map[string]interface{}) {
// 	args := []interface{}{
// 		"event_type", "security_event",
// 		"security_event", event,
// 		"severity", severity,
// 		"timestamp", time.Now(),
// 	}
//
// 	for k, v := range details {
// 		args = append(args, k, v)
// 	}
//
// 	if severity == "high" || severity == "critical" {
// 		l.Error("Security event", args...)
// 	} else {
// 		l.Warn("Security event", args...)
// 	}
// }

// GetMetrics returns current logging metrics
func (l *Logger) GetMetrics() *LogMetrics {
	l.metrics.mutex.RLock()
	defer l.metrics.mutex.RUnlock()

	// Calculate error rate
	totalErrors := l.metrics.LogsByLevel["error"]
	if l.metrics.TotalLogs > 0 {
		l.metrics.ErrorRate = float64(totalErrors) / float64(l.metrics.TotalLogs) * 100
	}

	return l.metrics
}

// ResetMetrics resets logging metrics
func (l *Logger) ResetMetrics() {
	l.metrics.mutex.Lock()
	defer l.metrics.mutex.Unlock()

	l.metrics.TotalLogs = 0
	l.metrics.LogsByLevel = make(map[string]int64)
	l.metrics.ErrorRate = 0
	l.metrics.LastLogTime = time.Time{}
}

// trackMetrics updates internal metrics
func (l *Logger) trackMetrics(level string) {
	if !l.config.EnableMetrics {
		return
	}

	l.metrics.mutex.Lock()
	defer l.metrics.mutex.Unlock()

	l.metrics.TotalLogs++
	l.metrics.LogsByLevel[level]++
	l.metrics.LastLogTime = time.Now()
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
			UserAgent:  r.UserAgent(),
			StartTime:  start,
		}

		// Log request start
		l.LogRequest(ctx)

		// Process request
		next.ServeHTTP(rw, r)

		// Log response
		ctx.StatusCode = rw.statusCode
		ctx.ResponseSize = int64(rw.size)
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

// HealthCheck returns logger health status
func (l *Logger) HealthCheck() map[string]interface{} {
	metrics := l.GetMetrics()
	uptime := time.Since(l.startTime)

	return map[string]interface{}{
		"status":     "healthy",
		"uptime":     uptime.String(),
		"total_logs": metrics.TotalLogs,
		"error_rate": fmt.Sprintf("%.2f%%", metrics.ErrorRate),
		"last_log":   metrics.LastLogTime,
		"config":     l.config,
	}
}

// Close properly closes the logger and any file handles
func (l *Logger) Close() error {
	if closer, ok := l.output.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
