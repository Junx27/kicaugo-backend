package event

import (
	"context"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

// NearbyResult adalah hasil pencarian event terdekat beserta asalnya.
type NearbyResult struct {
	Origin   *entities.NearbyOrigin
	RadiusKM int
	Events   []entities.Event
	Hint     *string
}

// Service berisi aturan pencarian event terdekat.
type Service interface {
	ListNearby(ctx context.Context, identifier string, radiusKM int, categoryID int64, limit int) (NearbyResult, error)
	CountOngoing(ctx context.Context) (int, error)
}

// service adalah implementasi Service.
type service struct {
	repo Repository
	now  func() time.Time
}

// NewService membuat service dengan repository.
func NewService(repo Repository) Service {
	return &service{repo: repo, now: time.Now}
}

// CountOngoing meneruskan penghitungan event yang sedang berlangsung ke repository.
func (s *service) CountOngoing(ctx context.Context) (int, error) {
	return s.repo.CountOngoing(ctx)
}

// ListNearby menentukan asal pencarian, lalu mengambil event dalam radius.
func (s *service) ListNearby(ctx context.Context, identifier string, radiusKM int, categoryID int64, limit int) (NearbyResult, error) {
	candidate, err := s.repo.Origin(ctx, identifier)
	if err != nil {
		return NearbyResult{}, err
	}
	radius := effectiveRadius(radiusKM, candidate.RadiusKM)
	if radius < 1 || radius > maxRadiusKM {
		return NearbyResult{}, apperrors.ErrInvalidRadius
	}
	if limit < 1 || limit > maxLimit {
		return NearbyResult{}, apperrors.ErrInvalidPayload
	}

	origin, hint := resolveOrigin(candidate, s.now())
	result := NearbyResult{RadiusKM: radius, Hint: hint, Events: []entities.Event{}}
	if origin == nil {
		return result, nil
	}
	result.Origin = origin
	events, err := s.repo.ListNearby(ctx, *origin, radius, categoryID, limit)
	if err != nil {
		return NearbyResult{}, err
	}
	result.Events = events
	return result, nil
}
