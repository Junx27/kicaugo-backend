package organizer

import (
	"context"
	"strings"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

// Service berisi aturan bisnis organizer.
type Service interface {
	Create(ctx context.Context, identifier string, in entities.Organizer) (entities.Organizer, error)
	RegisterMember(ctx context.Context, identifier string, organizerID int64, role string) (entities.EventOrganizerUser, error)
	HasActiveMembership(ctx context.Context, identifier string) (bool, error)
	IsActiveMember(ctx context.Context, identifier string, organizerID int64) (bool, error)
}

type service struct {
	repo Repository
}

// NewService membuat service dengan repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, identifier string, in entities.Organizer) (entities.Organizer, error) {
	userID, err := s.repo.ResolveUserID(ctx, identifier)
	if err != nil {
		return entities.Organizer{}, err
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Slug = strings.TrimSpace(in.Slug)
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	if in.Name == "" || in.Slug == "" || in.Email == "" {
		return entities.Organizer{}, apperrors.ErrInvalidPayload
	}
	created, err := s.repo.Create(ctx, in)
	if err != nil {
		return entities.Organizer{}, err
	}
	if _, err := s.repo.AddMember(ctx, created.ID, userID, "owner"); err != nil {
		return entities.Organizer{}, err
	}
	return created, nil
}

func (s *service) RegisterMember(ctx context.Context, identifier string, organizerID int64, role string) (entities.EventOrganizerUser, error) {
	userID, err := s.repo.ResolveUserID(ctx, identifier)
	if err != nil {
		return entities.EventOrganizerUser{}, err
	}
	role = strings.TrimSpace(role)
	if role == "" {
		role = "staff"
	}
	if !isValidRole(role) {
		return entities.EventOrganizerUser{}, apperrors.ErrInvalidPayload
	}
	if organizerID <= 0 {
		return entities.EventOrganizerUser{}, apperrors.ErrInvalidPayload
	}
	return s.repo.AddMember(ctx, organizerID, userID, role)
}

func (s *service) HasActiveMembership(ctx context.Context, identifier string) (bool, error) {
	return s.repo.HasActiveMembership(ctx, identifier)
}

func (s *service) IsActiveMember(ctx context.Context, identifier string, organizerID int64) (bool, error) {
	if organizerID <= 0 {
		return false, apperrors.ErrInvalidPayload
	}
	return s.repo.IsActiveMember(ctx, identifier, organizerID)
}

func isValidRole(role string) bool {
	switch role {
	case "owner", "admin", "staff", "verifier":
		return true
	default:
		return false
	}
}
