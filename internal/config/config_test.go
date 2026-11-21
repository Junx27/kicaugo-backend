package config

import (
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Run("konfigurasi lengkap", func(t *testing.T) {
		t.Setenv(envDatabaseURL, "postgres://user:pass@localhost:5432/kicaugo")
		t.Setenv(envJWTSecret, "rahasia-panjang")
		t.Setenv(envPort, "9090")
		t.Setenv(envDBMaxConns, "20")
		t.Setenv(envDBMinConns, "5")
		t.Setenv(envDBMaxConnLife, "2h")
		t.Setenv(envDBMaxConnIdle, "1h")
		t.Setenv(envDBHealthCheck, "30s")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Port != "9090" {
			t.Errorf("Port = %q, ingin 9090", cfg.Port)
		}
		if cfg.DatabaseURL == "" || cfg.JWTSecret != "rahasia-panjang" {
			t.Errorf("konfigurasi tidak terbaca: %+v", cfg)
		}
		if cfg.TokenTTL != tokenTTL {
			t.Errorf("TokenTTL = %v, ingin %v", cfg.TokenTTL, tokenTTL)
		}
		if cfg.Pool.MaxConns != 20 {
			t.Errorf("Pool.MaxConns = %d, ingin 20", cfg.Pool.MaxConns)
		}
		if cfg.Pool.MinConns != 5 {
			t.Errorf("Pool.MinConns = %d, ingin 5", cfg.Pool.MinConns)
		}
		if cfg.Pool.MaxConnLifetime != 2*time.Hour {
			t.Errorf("Pool.MaxConnLifetime = %v, ingin 2h", cfg.Pool.MaxConnLifetime)
		}
		if cfg.Pool.MaxConnIdleTime != 1*time.Hour {
			t.Errorf("Pool.MaxConnIdleTime = %v, ingin 1h", cfg.Pool.MaxConnIdleTime)
		}
		if cfg.Pool.HealthCheckPeriod != 30*time.Second {
			t.Errorf("Pool.HealthCheckPeriod = %v, ingin 30s", cfg.Pool.HealthCheckPeriod)
		}
	})

	t.Run("port memakai default", func(t *testing.T) {
		t.Setenv(envDatabaseURL, "postgres://localhost/kicaugo")
		t.Setenv(envJWTSecret, "rahasia-panjang")
		t.Setenv(envPort, "")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Port != defaultPort {
			t.Errorf("Port = %q, ingin %q", cfg.Port, defaultPort)
		}
		if cfg.Pool.MaxConns != defaultMaxConns {
			t.Errorf("Pool.MaxConns = %d, ingin %d", cfg.Pool.MaxConns, defaultMaxConns)
		}
		if cfg.Pool.MinConns != defaultMinConns {
			t.Errorf("Pool.MinConns = %d, ingin %d", cfg.Pool.MinConns, defaultMinConns)
		}
	})

	t.Run("database url wajib", func(t *testing.T) {
		t.Setenv(envDatabaseURL, "")
		t.Setenv(envJWTSecret, "rahasia-panjang")

		if _, err := Load(); err == nil {
			t.Error("Load() ingin error saat DATABASE_URL kosong")
		}
	})

	t.Run("jwt secret wajib", func(t *testing.T) {
		t.Setenv(envDatabaseURL, "postgres://localhost/kicaugo")
		t.Setenv(envJWTSecret, "")

		if _, err := Load(); err == nil {
			t.Error("Load() ingin error saat JWT_SECRET kosong")
		}
	})
}
