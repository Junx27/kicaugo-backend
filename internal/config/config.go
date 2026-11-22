// Package config membaca konfigurasi proses dari environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	envPort            = "PORT"
	envDatabaseURL     = "DATABASE_URL"
	envJWTSecret       = "JWT_SECRET"
	envDBMaxConns      = "DB_MAX_CONNS"
	envDBMinConns      = "DB_MIN_CONNS"
	envDBMaxConnLife   = "DB_MAX_CONN_LIFETIME"
	envDBMaxConnIdle   = "DB_MAX_CONN_IDLE_TIME"
	envDBHealthCheck   = "DB_HEALTH_CHECK_PERIOD"
	defaultPort        = "8080"
	defaultMaxConns    = 10
	defaultMinConns    = 2
	defaultMaxConnLife = 1 * time.Hour
	defaultMaxConnIdle = 30 * time.Minute
	defaultHealthCheck = 1 * time.Minute
	tokenTTL           = 24 * time.Hour
)

// PoolConfig berisi konfigurasi connection pool pgxpool.
type PoolConfig struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// Config adalah konfigurasi runtime yang sudah divalidasi.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	TokenTTL    time.Duration
	Pool        PoolConfig
	DocsEnabled bool
	DocsUser    string
	DocsPass    string
}

// Load membaca environment dan gagal lebih awal bila konfigurasi wajib kosong.
func Load() (Config, error) {
	databaseURL := os.Getenv(envDatabaseURL)
	if databaseURL == "" {
		return Config{}, fmt.Errorf("%s is required", envDatabaseURL)
	}
	secret := os.Getenv(envJWTSecret)
	if secret == "" {
		return Config{}, fmt.Errorf("%s is required", envJWTSecret)
	}
	port := os.Getenv(envPort)
	if port == "" {
		port = defaultPort
	}

	maxConns := parseIntEnv(envDBMaxConns, defaultMaxConns)
	minConns := parseIntEnv(envDBMinConns, defaultMinConns)
	maxConnLife := parseDurationEnv(envDBMaxConnLife, defaultMaxConnLife)
	maxConnIdle := parseDurationEnv(envDBMaxConnIdle, defaultMaxConnIdle)
	healthCheck := parseDurationEnv(envDBHealthCheck, defaultHealthCheck)

	// Docs configuration
	docsEnabled := os.Getenv("DOCS_ENABLED")
	if docsEnabled == "" {
		docsEnabled = "true" // default enabled (dev)
	}
	docsUser := os.Getenv("DOCS_USER")
	if docsUser == "" {
		docsUser = "admin"
	}
	docsPass := os.Getenv("DOCS_PASS")
	if docsPass == "" {
		docsPass = "changeme"
	}

	return Config{
		Port:        port,
		DatabaseURL: databaseURL,
		JWTSecret:   secret,
		TokenTTL:    tokenTTL,
		Pool: PoolConfig{
			MaxConns:          int32(maxConns),
			MinConns:          int32(minConns),
			MaxConnLifetime:   maxConnLife,
			MaxConnIdleTime:   maxConnIdle,
			HealthCheckPeriod: healthCheck,
		},
		DocsEnabled: docsEnabled == "true" || docsEnabled == "1",
		DocsUser:    docsUser,
		DocsPass:    docsPass,
	}, nil
}

func parseIntEnv(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return parsed
}

func parseDurationEnv(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	parsed, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return parsed
}
