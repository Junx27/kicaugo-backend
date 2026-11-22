package event

import (
	"context"
	"strings"
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

// Service berisi aturan event.
type Service interface {
	ListNearby(ctx context.Context, identifier string, radiusKM int, categoryID int64, limit int) (NearbyResult, error)
	CountOngoing(ctx context.Context) (int, error)
	Create(ctx context.Context, identifier string, in entities.EventCreate) (entities.EventDetail, error)
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

// Create membuat event beserta sub-event untuk organizer aktif.
func (s *service) Create(ctx context.Context, identifier string, in entities.EventCreate) (entities.EventDetail, error) {
	normalizeCreateInput(&in)
	if err := validateCreateInput(in); err != nil {
		return entities.EventDetail{}, err
	}
	active, err := s.repo.IsActiveOrganizerMember(ctx, identifier, in.OrganizerID)
	if err != nil {
		return entities.EventDetail{}, err
	}
	if !active {
		return entities.EventDetail{}, apperrors.ErrForbidden
	}
	return s.repo.Create(ctx, in)
}

func validateCreateInput(in entities.EventCreate) error {
	if in.OrganizerID <= 0 || strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Slug) == "" || strings.TrimSpace(in.Venue) == "" || strings.TrimSpace(in.City) == "" {
		return apperrors.ErrInvalidPayload
	}
	if !isValidEventStatus(in.Status) {
		return apperrors.ErrInvalidPayload
	}
	if in.TotalSlots < 0 {
		return apperrors.ErrInvalidPayload
	}
	for _, sub := range in.SubEvents {
		if err := validateSubEvent(sub); err != nil {
			return err
		}
	}
	return nil
}

func validateSubEvent(sub entities.SubEventCreate) error {
	if strings.TrimSpace(sub.Title) == "" || sub.MaxSlots < 1 || sub.MaxSlots > 60 || sub.Price < 0 || sub.GantanganCount < 0 || sub.SortOrder < 0 {
		return apperrors.ErrInvalidPayload
	}
	if sub.Winner1Prize < 0 || sub.Winner2Prize < 0 || sub.Winner3Prize < 0 {
		return apperrors.ErrInvalidPayload
	}
	switch sub.Status {
	case "available", "almostGone", "soldOut", "waiting":
		return nil
	default:
		return apperrors.ErrInvalidPayload
	}
}

func normalizeCreateInput(in *entities.EventCreate) {
	in.Title = strings.TrimSpace(in.Title)
	in.Slug = strings.TrimSpace(in.Slug)
	in.Venue = strings.TrimSpace(in.Venue)
	in.City = strings.TrimSpace(in.City)
	if in.Status == "" {
		in.Status = "draft"
	}
	for i := range in.SubEvents {
		in.SubEvents[i].Title = strings.TrimSpace(in.SubEvents[i].Title)
		if in.SubEvents[i].Status == "" {
			in.SubEvents[i].Status = "available"
		}
	}
}

func isValidEventStatus(status string) bool {
	switch status {
	case "draft", "published", "ongoing", "completed", "cancelled", "":
		return true
	default:
		return false
	}
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
