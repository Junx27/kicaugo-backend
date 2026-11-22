package birdcategory

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

// Repository menyimpan dan mencari kategori burung.
type Repository interface {
	List(ctx context.Context) ([]entities.BirdCategory, error)

	// scaffold:repo-interface
}

// repository adalah implementasi Repository berbasis pgxpool.
type repository struct {
	pool *pgxpool.Pool
}

// NewRepository membuat repository dengan connection pool yang sudah ada.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

// List mengambil semua kategori burung.
func (r *repository) List(ctx context.Context) ([]entities.BirdCategory, error) {
	const query = `SELECT id, name, slug, icon, created_at FROM bird_categories ORDER BY id`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list bird categories: %w", err)
	}
	defer rows.Close()

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

// scaffold:repo-methods
