// Package database mengelola koneksi, migrasi, index, dan seeder PostgreSQL.
package database

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Junx27/kicaugo-backend/internal/config"
	"github.com/Junx27/kicaugo-backend/internal/database/indexes"
	"github.com/Junx27/kicaugo-backend/internal/database/migrations"
	"github.com/Junx27/kicaugo-backend/internal/database/seeders"
)

const (
	pingTimeout  = 5 * time.Second
	applyTimeout = 30 * time.Second
)

// Executor menjalankan SQL pada database.
type Executor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type sqlSource struct {
	label string
	files fs.FS
}

var sources = []sqlSource{
	{label: "migration", files: migrations.MigrationFiles},
	{label: "index", files: indexes.IndexFiles},
	{label: "seeder", files: seeders.SeedFiles},
}

// Connect membuat connection pool, menerapkan konfigurasi, dan memastikan koneksi hidup.
func Connect(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	poolConfig.MaxConns = cfg.Pool.MaxConns
	poolConfig.MinConns = cfg.Pool.MinConns
	poolConfig.MaxConnLifetime = cfg.Pool.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.Pool.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.Pool.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

// Setup menjalankan migrasi, index, dan seeder secara berurutan.
func Setup(ctx context.Context, db Executor) error {
	return setup(ctx, db)
}

func setup(ctx context.Context, db Executor) error {
	for _, source := range sources {
		if err := applyFiles(ctx, db, source); err != nil {
			return err
		}
	}
	return nil
}

func applyFiles(ctx context.Context, db Executor, source sqlSource) error {
	files, err := sqlFiles(source.files)
	if err != nil {
		return fmt.Errorf("read %s files: %w", source.label, err)
	}
	for _, name := range files {
		if err := applyFile(ctx, db, source, name); err != nil {
			return err
		}
	}
	return nil
}

func applyFile(ctx context.Context, db Executor, source sqlSource, name string) error {
	content, err := fs.ReadFile(source.files, name)
	if err != nil {
		return fmt.Errorf("read %s %s: %w", source.label, name, err)
	}

	execCtx, cancel := context.WithTimeout(ctx, applyTimeout)
	defer cancel()
	if _, err := db.Exec(execCtx, string(content)); err != nil {
		return fmt.Errorf("apply %s %s: %w", source.label, name, err)
	}
	slog.Info(source.label+" applied", "file", name)
	return nil
}

func sqlFiles(fsys fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}
