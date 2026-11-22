package province

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

// Repository menyimpan dan mencari province.
type Repository interface {
	List(ctx context.Context) ([]entities.Province, error)
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

// List mengambil semua province.
func (r *repository) List(ctx context.Context) ([]entities.Province, error) {
	const query = `SELECT id, code, name, slug, created_at FROM provinces ORDER BY id`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list province: %w", err)
	}
	defer rows.Close()

	var out []entities.Province
	for rows.Next() {
		var c entities.Province
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Slug, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan province: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate province: %w", err)
	}
	return out, nil
}

// scaffold:repo-methods
