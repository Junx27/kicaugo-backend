package city

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

// Repository menyimpan dan mencari city.
type Repository interface {
	List(ctx context.Context) ([]entities.City, error)
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

// List mengambil semua city.
func (r *repository) List(ctx context.Context) ([]entities.City, error) {
	const query = `SELECT id, province_id, code, name, type, slug, latitude, longitude, created_at FROM cities ORDER BY id`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list city: %w", err)
	}
	defer rows.Close()

	var out []entities.City
	for rows.Next() {
		var c entities.City
		if err := rows.Scan(&c.ID, &c.ProvinceID, &c.Code, &c.Name, &c.Type, &c.Slug, &c.Latitude, &c.Longitude, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan city: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate city: %w", err)
	}
	return out, nil
}

// scaffold:repo-methods
