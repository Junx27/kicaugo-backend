package birdcategory

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

// Repository menyimpan dan mencari kategori burung.
type Repository interface {
	List(ctx context.Context) ([]entities.BirdCategory, error)
	FindByID(ctx context.Context, id int64) (entities.BirdCategory, error)
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

// FindByID mencari kategori burung berdasarkan id.
func (r *repository) FindByID(ctx context.Context, id int64) (entities.BirdCategory, error) {
	const query = `SELECT id, name, slug, icon, created_at FROM bird_categories WHERE id = $1`
	var c entities.BirdCategory
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&c.ID, &c.Name, &c.Slug, &c.Icon, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return entities.BirdCategory{}, apperrors.ErrNotFound
	}
	if err != nil {
		return entities.BirdCategory{}, fmt.Errorf("find bird category %d: %w", id, err)
	}
	return c, nil
}
