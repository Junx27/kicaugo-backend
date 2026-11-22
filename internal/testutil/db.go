// Package testutil menyediakan helper test yang memakai database Postgres lokal.
package testutil

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const pingTimeout = 3 * time.Second

// OpenDB membuka pool ke database test dari TEST_DATABASE_URL.
// Test di-skip bila variabel belum diset atau database tidak dapat dijangkau.
func OpenDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL belum diset")
	}
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("database test tidak tersedia: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("database test tidak tersedia: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}
