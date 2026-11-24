// Package db menyediakan kontrak querier minimal yang dipenuhi pgxpool.Pool maupun pgx.Tx.
package db

import (
	"context"
	"fmt"

	"github.com/Junx27/kicaugo-backend/internal/entities"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ScanBirdCategories membaca semua baris kategori burung dari rows.
func ScanBirdCategories(rows pgx.Rows) ([]entities.BirdCategory, error) {
	var out []entities.BirdCategory
	for rows.Next() {
		var c entities.BirdCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Icon, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan bird category: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bird categories: %w", err)
	}
	return out, nil
}

// Querier adalah kontrak minimal untuk eksekusi query.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}
