package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

const uniqueViolation = "23505"

// Repository menyimpan dan mencari pengguna.
type Repository interface {
	Create(ctx context.Context, name, identifier, passwordHash string) (entities.User, error)
	FindByIdentifier(ctx context.Context, identifier string) (entities.User, error)
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

// Create menyimpan pengguna baru dan menerjemahkan pelanggaran unik.
func (r *repository) Create(ctx context.Context, name, identifier, passwordHash string) (entities.User, error) {
	const query = `INSERT INTO users (name, identifier, password_hash) VALUES ($1, $2, $3) RETURNING id, name, identifier, password_hash`
	var user entities.User
	err := r.pool.QueryRow(ctx, query, name, identifier, passwordHash).
		Scan(&user.ID, &user.Name, &user.Identifier, &user.PasswordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return entities.User{}, apperrors.ErrIdentifierTaken
		}
		return entities.User{}, fmt.Errorf("insert user %s: %w", identifier, err)
	}
	return user, nil
}

// FindByIdentifier mencari pengguna berdasarkan email atau nomor WhatsApp.
func (r *repository) FindByIdentifier(ctx context.Context, identifier string) (entities.User, error) {
	const query = `SELECT id, name, identifier, password_hash FROM users WHERE identifier = $1`
	var user entities.User
	err := r.pool.QueryRow(ctx, query, identifier).
		Scan(&user.ID, &user.Name, &user.Identifier, &user.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return entities.User{}, apperrors.ErrNotFound
	}
	if err != nil {
		return entities.User{}, fmt.Errorf("find user %s: %w", identifier, err)
	}
	return user, nil
}

// scaffold:repo-methods
