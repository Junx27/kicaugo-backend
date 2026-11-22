package event

import (
	"context"
	"encoding/json"
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
	StartStreaming(ctx context.Context, identifier string, in entities.StartEventStreamingInput) (entities.EventStreaming, error)
	UpdateStreamingJuara(ctx context.Context, identifier string, in entities.UpdateEventStreamingJuaraInput) (entities.EventStreaming, error)
	StopStreaming(ctx context.Context, identifier string, eventID int64) (entities.EventStreaming, error)
	GetActiveStreaming(ctx context.Context) (*entities.ActiveEventStreaming, error)
	GetRandomFeatured(ctx context.Context) (*entities.Event, error)
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

// StartStreaming memulai sesi streaming untuk event milik organizer aktif.
func (s *service) StartStreaming(ctx context.Context, identifier string, in entities.StartEventStreamingInput) (entities.EventStreaming, error) {
	normalizeStartStreamingInput(&in)
	if err := validateStartStreamingInput(in); err != nil {
		return entities.EventStreaming{}, err
	}
	ev, err := s.repo.EventByID(ctx, in.EventID)
	if err != nil {
		return entities.EventStreaming{}, err
	}
	if err := ensureEventStreamable(ev.Status); err != nil {
		return entities.EventStreaming{}, err
	}
	active, err := s.repo.IsActiveOrganizerMember(ctx, identifier, ev.OrganizerID)
	if err != nil {
		return entities.EventStreaming{}, err
	}
	if !active {
		return entities.EventStreaming{}, apperrors.ErrForbidden
	}
	hasActive, err := s.repo.HasActiveStreaming(ctx, in.EventID)
	if err != nil {
		return entities.EventStreaming{}, err
	}
	if hasActive {
		return entities.EventStreaming{}, apperrors.ErrStreamingAlreadyActive
	}
	return s.repo.CreateStreaming(ctx, in)
}

// UpdateStreamingJuara memperbarui juara pada streaming aktif event.
func (s *service) UpdateStreamingJuara(ctx context.Context, identifier string, in entities.UpdateEventStreamingJuaraInput) (entities.EventStreaming, error) {
	if err := validateUpdateStreamingJuaraInput(in); err != nil {
		return entities.EventStreaming{}, err
	}
	ev, err := s.repo.EventByID(ctx, in.EventID)
	if err != nil {
		return entities.EventStreaming{}, err
	}
	active, err := s.repo.IsActiveOrganizerMember(ctx, identifier, ev.OrganizerID)
	if err != nil {
		return entities.EventStreaming{}, err
	}
	if !active {
		return entities.EventStreaming{}, apperrors.ErrForbidden
	}
	return s.repo.UpdateStreamingJuara(ctx, in)
}

// StopStreaming menghentikan streaming aktif milik event organizer.
func (s *service) StopStreaming(ctx context.Context, identifier string, eventID int64) (entities.EventStreaming, error) {
	if eventID <= 0 {
		return entities.EventStreaming{}, apperrors.ErrInvalidPayload
	}
	ev, err := s.repo.EventByID(ctx, eventID)
	if err != nil {
		return entities.EventStreaming{}, err
	}
	active, err := s.repo.IsActiveOrganizerMember(ctx, identifier, ev.OrganizerID)
	if err != nil {
		return entities.EventStreaming{}, err
	}
	if !active {
		return entities.EventStreaming{}, apperrors.ErrForbidden
	}
	return s.repo.StopStreaming(ctx, eventID)
}

// GetActiveStreaming mengambil satu streaming aktif terbaru.
func (s *service) GetActiveStreaming(ctx context.Context) (*entities.ActiveEventStreaming, error) {
	return s.repo.GetActiveStreaming(ctx)
}

// GetRandomFeatured mengambil satu event acak untuk featured home.
func (s *service) GetRandomFeatured(ctx context.Context) (*entities.Event, error) {
	return s.repo.GetRandomFeatured(ctx)
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

func normalizeStartStreamingInput(in *entities.StartEventStreamingInput) {
	in.LinkStream = strings.TrimSpace(in.LinkStream)
}

func validateStartStreamingInput(in entities.StartEventStreamingInput) error {
	if in.EventID <= 0 || in.LinkStream == "" {
		return apperrors.ErrInvalidPayload
	}
	return nil
}

func validateUpdateStreamingJuaraInput(in entities.UpdateEventStreamingJuaraInput) error {
	if in.EventID <= 0 || len(in.Juara) == 0 || !json.Valid(in.Juara) {
		return apperrors.ErrInvalidPayload
	}
	return nil
}

func ensureEventStreamable(status string) error {
	switch status {
	case "published", "ongoing":
		return nil
	default:
		return apperrors.ErrEventNotStreamable
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
