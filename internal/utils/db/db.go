// Package db menyediakan kontrak querier minimal yang dipenuhi pgxpool.Pool maupun pgx.Tx.
package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Querier adalah kontrak minimal untuk eksekusi query.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

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

// ResolveUserID menerjemahkan identifier dari token menjadi ID user.
func ResolveUserID(ctx context.Context, q Querier, identifier string) (int64, error) {
	var id int64
	err := q.QueryRow(ctx, `SELECT id FROM users WHERE identifier = $1`, identifier).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, apperrors.ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("resolve user: %w", err)
	}
	return id, nil
}
