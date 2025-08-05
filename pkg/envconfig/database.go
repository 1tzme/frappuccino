package envconfig

import (
	"strconv"
	"time"

	"hot-coffee/pkg/database"
)

// LoadDatabaseConfig loads database configuration from environment variables
func LoadDatabaseConfig() database.Config {
	config := database.DefaultConfig()

	// Override with environment variables if they exist
	if host := GetEnv("DB_HOST", ""); host != "" {
		config.Host = host
	}

	if portStr := GetEnv("DB_PORT", ""); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Port = port
		}
	}

	if user := GetEnv("DB_USER", ""); user != "" {
		config.User = user
	}

	if password := GetEnv("DB_PASSWORD", ""); password != "" {
		config.Password = password
	}

	if dbname := GetEnv("DB_NAME", ""); dbname != "" {
		config.DBName = dbname
	}

	if sslmode := GetEnv("DB_SSL_MODE", ""); sslmode != "" {
		config.SSLMode = sslmode
	}

	// Connection pool settings
	if maxOpenConnsStr := GetEnv("DB_MAX_OPEN_CONNS", ""); maxOpenConnsStr != "" {
		if maxOpenConns, err := strconv.Atoi(maxOpenConnsStr); err == nil && maxOpenConns > 0 {
			config.MaxOpenConns = maxOpenConns
		}
	}

	if maxIdleConnsStr := GetEnv("DB_MAX_IDLE_CONNS", ""); maxIdleConnsStr != "" {
		if maxIdleConns, err := strconv.Atoi(maxIdleConnsStr); err == nil && maxIdleConns > 0 {
			config.MaxIdleConns = maxIdleConns
		}
	}

	if connMaxLifetimeStr := GetEnv("DB_CONN_MAX_LIFETIME", ""); connMaxLifetimeStr != "" {
		if connMaxLifetime, err := time.ParseDuration(connMaxLifetimeStr); err == nil {
			config.ConnMaxLifetime = connMaxLifetime
		}
	}

	if connMaxIdleTimeStr := GetEnv("DB_CONN_MAX_IDLE_TIME", ""); connMaxIdleTimeStr != "" {
		if connMaxIdleTime, err := time.ParseDuration(connMaxIdleTimeStr); err == nil {
			config.ConnMaxIdleTime = connMaxIdleTime
		}
	}

	return config
}
