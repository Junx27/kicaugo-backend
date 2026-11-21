package home

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const selectContent = `SELECT payload FROM home_contents WHERE id = true`

// Repository menyediakan konten Beranda.
type Repository interface {
	Get(ctx context.Context) (Content, error)
}

// SQLRepository mengambil payload JSONB dari database.
type SQLRepository struct {
	pool *pgxpool.Pool
}

// NewSQLRepository membuat repository dengan connection pool yang sudah ada.
func NewSQLRepository(pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{pool: pool}
}

// Get membaca dan mengurai payload konten Beranda.
func (r *SQLRepository) Get(ctx context.Context) (Content, error) {
	var raw []byte
	if err := r.pool.QueryRow(ctx, selectContent).Scan(&raw); err != nil {
		return Content{}, fmt.Errorf("load home content: %w", err)
	}
	var content Content
	if err := json.Unmarshal(raw, &content); err != nil {
		return Content{}, fmt.Errorf("decode home content: %w", err)
	}
	return content, nil
}
