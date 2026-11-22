package organizer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

const uniqueViolation = "23505"

// Repository menyimpan dan mencari organizer serta membership.
type Repository interface {
	ResolveUserID(ctx context.Context, identifier string) (int64, error)
	Create(ctx context.Context, in entities.Organizer) (entities.Organizer, error)
	AddMember(ctx context.Context, organizerID, userID int64, role string) (entities.EventOrganizerUser, error)
	IsActiveMember(ctx context.Context, identifier string, organizerID int64) (bool, error)
	HasActiveMembership(ctx context.Context, identifier string) (bool, error)
}

type db interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type repository struct {
	db db
}

// NewRepository membuat repository dengan connection pool yang sudah ada.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{db: pool}
}

func (r *repository) ResolveUserID(ctx context.Context, identifier string) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `SELECT id FROM users WHERE identifier = $1`, identifier).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, apperrors.ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("resolve user: %w", err)
	}
	return id, nil
}

func (r *repository) Create(ctx context.Context, in entities.Organizer) (entities.Organizer, error) {
	const query = `INSERT INTO organizers (name, slug, email, phone, logo_url, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, slug, email, phone, logo_url, description, is_verified, created_at, updated_at`
	var out entities.Organizer
	err := r.db.QueryRow(ctx, query, in.Name, in.Slug, in.Email, in.Phone, in.LogoURL, in.Description).
		Scan(&out.ID, &out.Name, &out.Slug, &out.Email, &out.Phone, &out.LogoURL, &out.Description, &out.IsVerified, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			if strings.Contains(pgErr.ConstraintName, "slug") {
				return entities.Organizer{}, apperrors.ErrOrganizerSlugTaken
			}
			return entities.Organizer{}, apperrors.ErrIdentifierTaken
		}
		return entities.Organizer{}, fmt.Errorf("insert organizer: %w", err)
	}
	return out, nil
}

func (r *repository) AddMember(ctx context.Context, organizerID, userID int64, role string) (entities.EventOrganizerUser, error) {
	const query = `INSERT INTO event_organizer_users (organizer_id, user_id, role, is_active)
		VALUES ($1, $2, $3, true)
		RETURNING id, organizer_id, user_id, role, is_active, created_at, updated_at`
	var out entities.EventOrganizerUser
	err := r.db.QueryRow(ctx, query, organizerID, userID, role).
		Scan(&out.ID, &out.OrganizerID, &out.UserID, &out.Role, &out.IsActive, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return entities.EventOrganizerUser{}, apperrors.ErrOrganizerMemberTaken
		}
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return entities.EventOrganizerUser{}, apperrors.ErrOrganizerNotFound
		}
		return entities.EventOrganizerUser{}, fmt.Errorf("insert organizer member: %w", err)
	}
	return out, nil
}

func (r *repository) IsActiveMember(ctx context.Context, identifier string, organizerID int64) (bool, error) {
	const query = `SELECT EXISTS (
		SELECT 1
		FROM event_organizer_users eou
		JOIN users u ON u.id = eou.user_id
		WHERE u.identifier = $1 AND eou.organizer_id = $2 AND eou.is_active = true
	)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, identifier, organizerID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check organizer member: %w", err)
	}
	return exists, nil
}

func (r *repository) HasActiveMembership(ctx context.Context, identifier string) (bool, error) {
	const query = `SELECT EXISTS (
		SELECT 1
		FROM event_organizer_users eou
		JOIN users u ON u.id = eou.user_id
		WHERE u.identifier = $1 AND eou.is_active = true
	)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, identifier).Scan(&exists); err != nil {
		return false, fmt.Errorf("check organizer membership: %w", err)
	}
	return exists, nil
}
