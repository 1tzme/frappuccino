// pkg/database/connection.go
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"hot-coffee/pkg/logger"

	_ "github.com/lib/pq"
)

// Connection pool configuration constants
const (
	DefaultMaxOpenConns    = 25
	DefaultMaxIdleConns    = 5
	DefaultConnMaxLifetime = 5 * time.Minute
	DefaultConnMaxIdleTime = 30 * time.Second
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string

	// Connection pool settings (optional)
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultConfig returns a database configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "",
		DBName:          "hotcoffee",
		SSLMode:         "disable",
		MaxOpenConns:    DefaultMaxOpenConns,
		MaxIdleConns:    DefaultMaxIdleConns,
		ConnMaxLifetime: DefaultConnMaxLifetime,
		ConnMaxIdleTime: DefaultConnMaxIdleTime,
	}
}

// BuildConnectionString builds a PostgreSQL connection string from config
func (c Config) BuildConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

type DB struct {
	*sql.DB
	logger *logger.Logger
}

func NewConnection(config Config, log *logger.Logger) (*DB, error) {
	log.Info("Establishing database connection",
		"host", config.Host,
		"port", config.Port,
		"database", config.DBName,
		"ssl_mode", config.SSLMode)

	dsn := config.BuildConnectionString()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Error("Failed to open database connection", "error", err)
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Configure connection pool with defaults if not specified
	maxOpenConns := config.MaxOpenConns
	if maxOpenConns <= 0 {
		maxOpenConns = DefaultMaxOpenConns
	}

	maxIdleConns := config.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = DefaultMaxIdleConns
	}

	connMaxLifetime := config.ConnMaxLifetime
	if connMaxLifetime <= 0 {
		connMaxLifetime = DefaultConnMaxLifetime
	}

	connMaxIdleTime := config.ConnMaxIdleTime
	if connMaxIdleTime <= 0 {
		connMaxIdleTime = DefaultConnMaxIdleTime
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	log.Debug("Database connection pool configured",
		"max_open_conns", maxOpenConns,
		"max_idle_conns", maxIdleConns,
		"conn_max_lifetime", connMaxLifetime,
		"conn_max_idle_time", connMaxIdleTime)

	if err := db.Ping(); err != nil {
		db.Close()
		log.Error("Failed to ping database", "error", err)
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	log.Info("Database connection established successfully",
		"host", config.Host,
		"port", config.Port,
		"database", config.DBName)
	return &DB{DB: db, logger: log}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	db.logger.Info("Closing database connection")
	return db.DB.Close()
}

// Ping tests the database connection
func (db *DB) Ping() error {
	return db.DB.Ping()
}

// HealthCheck returns the database health status
func (db *DB) HealthCheck() error {
	db.logger.Debug("Performing database health check")

	if err := db.Ping(); err != nil {
		db.logger.Error("Database health check failed", "error", err)
		return fmt.Errorf("database ping failed: %v", err)
	}

	// Test with a simple query
	var result int
	err := db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		db.logger.Error("Database query test failed", "error", err)
		return fmt.Errorf("database query test failed: %v", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected query result: got %d, expected 1", result)
	}

	db.logger.Debug("Database health check passed")
	return nil
}

// GetStats returns database connection statistics
func (db *DB) GetStats() sql.DBStats {
	return db.DB.Stats()
}

// LogStats logs current database connection statistics
func (db *DB) LogStats() {
	stats := db.GetStats()
	db.logger.Info("Database connection stats",
		"open_connections", stats.OpenConnections,
		"in_use", stats.InUse,
		"idle", stats.Idle,
		"wait_count", stats.WaitCount,
		"wait_duration", stats.WaitDuration,
		"max_idle_closed", stats.MaxIdleClosed,
		"max_idle_time_closed", stats.MaxIdleTimeClosed,
		"max_lifetime_closed", stats.MaxLifetimeClosed)
}

// ValidateConnection validates the database connection with timeout
func (db *DB) ValidateConnection(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	db.logger.Debug("Validating database connection", "timeout", timeout)

	err := db.PingContext(ctx)
	if err != nil {
		db.logger.Error("Database connection validation failed", "error", err, "timeout", timeout)
		return fmt.Errorf("database ping failed within %v: %v", timeout, err)
	}

	db.logger.Debug("Database connection validation successful")
	return nil
}

// ExecuteInTransaction executes a function within a database transaction
func (db *DB) ExecuteInTransaction(fn func(*sql.Tx) error) error {
	db.logger.Debug("Starting database transaction")

	tx, err := db.Begin()
	if err != nil {
		db.logger.Error("Failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			db.logger.Error("Transaction panic, rolling back", "panic", p)
			panic(p)
		} else if err != nil {
			tx.Rollback()
			db.logger.Warn("Transaction failed, rolling back", "error", err)
		} else {
			err = tx.Commit()
			if err != nil {
				db.logger.Error("Failed to commit transaction", "error", err)
			} else {
				db.logger.Debug("Transaction committed successfully")
			}
		}
	}()

	err = fn(tx)
	return err
}
