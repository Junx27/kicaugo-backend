package config

import (
	"testing"
	"time"
)

const (
	testDatabaseURL = "postgres://localhost/kicaugo"
	testJWTSecret   = "rahasia-panjang"
	loadErrFormat   = "Load() error = %v"
	cartoTTLFormat  = "CartoTileCacheTTL = %v"
	cartoMaxFormat  = "CartoTileCacheMaxEntries = %d"
	cartoRateFormat = "CartoTileRatePerMin = %d"
)

func setRequiredConfigEnv(t *testing.T) {
	t.Helper()
	t.Setenv(envDatabaseURL, testDatabaseURL)
	t.Setenv(envJWTSecret, testJWTSecret)
}

func TestLoadBasemapConfig(t *testing.T) {
	t.Run("kosong memakai default", func(t *testing.T) {
		setRequiredConfigEnv(t)
		t.Setenv(envCartoBasemapKey, "")
		t.Setenv(envCartoBasemapURL, "")
		t.Setenv(envCartoTileCacheTTL, "")
		t.Setenv(envCartoTileCacheMaxEntries, "")
		t.Setenv(envCartoTileRatePerMin, "")

		cfg, err := Load()
		if err != nil {
			t.Fatalf(loadErrFormat, err)
		}
		if cfg.CartoBasemapKey != "" {
			t.Errorf("CartoBasemapKey = %q, ingin kosong", cfg.CartoBasemapKey)
		}
		if cfg.CartoBasemapURL != defaultCartoBasemapURL {
			t.Errorf("CartoBasemapURL = %q", cfg.CartoBasemapURL)
		}
		if cfg.CartoTileCacheTTL != defaultCartoTileCacheTTL {
			t.Errorf(cartoTTLFormat, cfg.CartoTileCacheTTL)
		}
		if cfg.CartoTileCacheMaxEntries != defaultCartoTileCacheMaxEntries {
			t.Errorf(cartoMaxFormat, cfg.CartoTileCacheMaxEntries)
		}
		if cfg.CartoTileRatePerMin != defaultCartoTileRatePerMin {
			t.Errorf(cartoRateFormat, cfg.CartoTileRatePerMin)
		}
	})

	t.Run("terisi", func(t *testing.T) {
		setRequiredConfigEnv(t)
		t.Setenv(envCartoBasemapKey, "fake-key")
		t.Setenv(envCartoBasemapURL, "https://tiles.example/{z}/{x}/{y}.png")
		t.Setenv(envCartoTileCacheTTL, "2h")
		t.Setenv(envCartoTileCacheMaxEntries, "10")
		t.Setenv(envCartoTileRatePerMin, "5")

		cfg, err := Load()
		if err != nil {
			t.Fatalf(loadErrFormat, err)
		}
		if cfg.CartoBasemapKey != "fake-key" {
			t.Errorf("CartoBasemapKey = %q", cfg.CartoBasemapKey)
		}
		if cfg.CartoBasemapURL != "https://tiles.example/{z}/{x}/{y}.png" {
			t.Errorf("CartoBasemapURL = %q", cfg.CartoBasemapURL)
		}
		if cfg.CartoTileCacheTTL != 2*time.Hour {
			t.Errorf(cartoTTLFormat, cfg.CartoTileCacheTTL)
		}
		if cfg.CartoTileCacheMaxEntries != 10 {
			t.Errorf(cartoMaxFormat, cfg.CartoTileCacheMaxEntries)
		}
		if cfg.CartoTileRatePerMin != 5 {
			t.Errorf(cartoRateFormat, cfg.CartoTileRatePerMin)
		}
	})

	t.Run("nilai tidak valid memakai default", func(t *testing.T) {
		setRequiredConfigEnv(t)
		t.Setenv(envCartoBasemapKey, "")
		t.Setenv(envCartoBasemapURL, "")
		t.Setenv(envCartoTileCacheTTL, "bukan-durasi")
		t.Setenv(envCartoTileCacheMaxEntries, "abc")
		t.Setenv(envCartoTileRatePerMin, "xyz")

		cfg, err := Load()
		if err != nil {
			t.Fatalf(loadErrFormat, err)
		}
		if cfg.CartoTileCacheTTL != defaultCartoTileCacheTTL {
			t.Errorf(cartoTTLFormat, cfg.CartoTileCacheTTL)
		}
		if cfg.CartoTileCacheMaxEntries != defaultCartoTileCacheMaxEntries {
			t.Errorf(cartoMaxFormat, cfg.CartoTileCacheMaxEntries)
		}
		if cfg.CartoTileRatePerMin != defaultCartoTileRatePerMin {
			t.Errorf(cartoRateFormat, cfg.CartoTileRatePerMin)
		}
	})
}

func TestLoad(t *testing.T) {
	t.Run("konfigurasi lengkap", func(t *testing.T) {
		t.Setenv(envDatabaseURL, "postgres://user:pass@localhost:5432/kicaugo")
		t.Setenv(envJWTSecret, testJWTSecret)
		t.Setenv(envPort, "9090")
		t.Setenv(envDBMaxConns, "20")
		t.Setenv(envDBMinConns, "5")
		t.Setenv(envDBMaxConnLife, "2h")
		t.Setenv(envDBMaxConnIdle, "1h")
		t.Setenv(envDBHealthCheck, "30s")

		cfg, err := Load()
		if err != nil {
			t.Fatalf(loadErrFormat, err)
		}
		if cfg.Port != "9090" {
			t.Errorf("Port = %q, ingin 9090", cfg.Port)
		}
		if cfg.DatabaseURL == "" || cfg.JWTSecret != testJWTSecret {
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
		t.Setenv(envDatabaseURL, testDatabaseURL)
		t.Setenv(envJWTSecret, testJWTSecret)
		t.Setenv(envPort, "")

		cfg, err := Load()
		if err != nil {
			t.Fatalf(loadErrFormat, err)
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
		t.Setenv(envJWTSecret, testJWTSecret)

		if _, err := Load(); err == nil {
			t.Error("Load() ingin error saat DATABASE_URL kosong")
		}
	})

	t.Run("jwt secret wajib", func(t *testing.T) {
		t.Setenv(envDatabaseURL, testDatabaseURL)
		t.Setenv(envJWTSecret, "")

		if _, err := Load(); err == nil {
			t.Error("Load() ingin error saat JWT_SECRET kosong")
		}
	})
}
