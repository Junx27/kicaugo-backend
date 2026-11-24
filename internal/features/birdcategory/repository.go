package birdcategory

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	dbutil "github.com/Junx27/kicaugo-backend/internal/utils/db"
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

	return dbutil.ScanBirdCategories(rows)
}

// scaffold:repo-methods
