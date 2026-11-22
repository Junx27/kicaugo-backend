package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolation = "23505"

// User adalah baris pengguna di database.
type User struct {
	ID           int64
	Name         string
	Identifier   string
	PasswordHash string
}

// Repository menyimpan dan mencari pengguna.
type Repository interface {
	Create(ctx context.Context, name, identifier, passwordHash string) (User, error)
	FindByIdentifier(ctx context.Context, identifier string) (User, error)
}

// SQLRepository adalah implementasi Repository berbasis pgxpool.
type SQLRepository struct {
	pool *pgxpool.Pool
}

// NewSQLRepository membuat repository dengan connection pool yang sudah ada.
func NewSQLRepository(pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{pool: pool}
}

// Create menyimpan pengguna baru dan menerjemahkan pelanggaran unik.
func (r *SQLRepository) Create(ctx context.Context, name, identifier, passwordHash string) (User, error) {
	const query = `INSERT INTO users (name, identifier, password_hash) VALUES ($1, $2, $3) RETURNING id, name, identifier, password_hash`
	var user User
	err := r.pool.QueryRow(ctx, query, name, identifier, passwordHash).
		Scan(&user.ID, &user.Name, &user.Identifier, &user.PasswordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return User{}, ErrIdentifierTaken
		}
		return User{}, fmt.Errorf("insert user %s: %w", identifier, err)
	}
	return user, nil
}

// FindByIdentifier mencari pengguna berdasarkan email atau nomor WhatsApp.
func (r *SQLRepository) FindByIdentifier(ctx context.Context, identifier string) (User, error) {
	const query = `SELECT id, name, identifier, password_hash FROM users WHERE identifier = $1`
	var user User
	err := r.pool.QueryRow(ctx, query, identifier).
		Scan(&user.ID, &user.Name, &user.Identifier, &user.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("find user %s: %w", identifier, err)
	}
	return user, nil
}
