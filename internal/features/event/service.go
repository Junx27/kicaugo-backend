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
	ListNearbySchedules(ctx context.Context, identifier string, radiusKM int, categoryID int64, limit int) ([]entities.EventSchedule, error)
	CountOngoing(ctx context.Context) (int, error)
	CountGantangan(ctx context.Context) (int, error)
	Create(ctx context.Context, identifier string, in entities.EventCreate) (entities.EventDetail, error)
	StartStreaming(ctx context.Context, identifier string, in entities.StartEventStreamingInput) (entities.EventStreaming, error)
	UpdateStreamingJuara(ctx context.Context, identifier string, in entities.UpdateEventStreamingJuaraInput) (entities.EventStreaming, error)
	StopStreaming(ctx context.Context, identifier string, eventID int64) (entities.EventStreaming, error)
	GetActiveStreaming(ctx context.Context) (*entities.ActiveEventStreaming, error)
	GetRandomFeatured(ctx context.Context, identifier string) (*entities.FeaturedEvent, error)
	GetRandomUpcoming(ctx context.Context) (*entities.UpcomingEvent, error)
	GetContestDetail(ctx context.Context, eventID int64) (entities.ContestDetail, error)
	ListEventRules(ctx context.Context, eventID int64) ([]entities.ContestRule, error)
	CreateEventRule(ctx context.Context, identifier string, in entities.EventRuleCreate) (entities.ContestRule, error)
	UpdateEventRule(ctx context.Context, identifier string, in entities.EventRuleUpdate) (entities.ContestRule, error)
	DeleteEventRule(ctx context.Context, identifier string, eventID, ruleID int64) error
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

// CountGantangan meneruskan penjumlahan gantangan seluruh event ke repository.
func (s *service) CountGantangan(ctx context.Context) (int, error) {
	return s.repo.CountGantangan(ctx)
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
func (s *service) GetRandomFeatured(ctx context.Context, identifier string) (*entities.FeaturedEvent, error) {
	return s.repo.GetRandomFeatured(ctx, identifier)
}

// GetRandomUpcoming mengambil satu event acak yang belum release untuk home upcoming.
func (s *service) GetRandomUpcoming(ctx context.Context) (*entities.UpcomingEvent, error) {
	return s.repo.GetRandomUpcoming(ctx)
}

// GetContestDetail mengambil detail lomba berdasarkan ID.
func (s *service) GetContestDetail(ctx context.Context, eventID int64) (entities.ContestDetail, error) {
	if eventID <= 0 {
		return entities.ContestDetail{}, apperrors.ErrInvalidPayload
	}
	return s.repo.GetContestDetail(ctx, eventID)
}

// ListEventRules mengambil aturan event setelah memvalidasi ID.
func (s *service) ListEventRules(ctx context.Context, eventID int64) ([]entities.ContestRule, error) {
	if err := validateEventRuleIDs(eventID); err != nil {
		return nil, err
	}
	return s.repo.ListEventRules(ctx, eventID)
}

// CreateEventRule menyimpan aturan baru untuk event milik organizer aktif.
func (s *service) CreateEventRule(ctx context.Context, identifier string, in entities.EventRuleCreate) (entities.ContestRule, error) {
	if err := validateEventRuleIDs(in.EventID); err != nil {
		return entities.ContestRule{}, err
	}
	if err := validateEventRuleInput(in.EventRuleInput); err != nil {
		return entities.ContestRule{}, err
	}
	if err := s.authorizeRuleWriter(ctx, identifier, in.EventID); err != nil {
		return entities.ContestRule{}, err
	}
	return s.repo.CreateEventRule(ctx, in)
}

// UpdateEventRule memperbarui aturan event milik organizer aktif.
func (s *service) UpdateEventRule(ctx context.Context, identifier string, in entities.EventRuleUpdate) (entities.ContestRule, error) {
	if err := validateEventRuleIDs(in.EventID, in.RuleID); err != nil {
		return entities.ContestRule{}, err
	}
	if err := validateEventRuleInput(in.EventRuleInput); err != nil {
		return entities.ContestRule{}, err
	}
	if err := s.authorizeRuleWriter(ctx, identifier, in.EventID); err != nil {
		return entities.ContestRule{}, err
	}
	return s.repo.UpdateEventRule(ctx, in)
}

// DeleteEventRule menghapus aturan event milik organizer aktif.
func (s *service) DeleteEventRule(ctx context.Context, identifier string, eventID, ruleID int64) error {
	if err := validateEventRuleIDs(eventID, ruleID); err != nil {
		return err
	}
	if err := s.authorizeRuleWriter(ctx, identifier, eventID); err != nil {
		return err
	}
	return s.repo.DeleteEventRule(ctx, eventID, ruleID)
}

func validateEventRuleIDs(eventID int64, ruleIDs ...int64) error {
	if eventID <= 0 {
		return apperrors.ErrInvalidPayload
	}
	for _, id := range ruleIDs {
		if id <= 0 {
			return apperrors.ErrInvalidPayload
		}
	}
	return nil
}

func validateEventRuleInput(in entities.EventRuleInput) error {
	title := strings.TrimSpace(in.Title)
	if title == "" || len(title) < 3 || len(title) > 150 {
		return apperrors.ErrInvalidPayload
	}
	if in.Description != nil && len(*in.Description) > 1000 {
		return apperrors.ErrInvalidPayload
	}
	if !isRuleSeverity(in.Severity) || in.SortOrder < 0 {
		return apperrors.ErrInvalidPayload
	}
	return nil
}

func isRuleSeverity(value string) bool {
	switch value {
	case entities.RuleSeverityInfo, entities.RuleSeverityError, entities.RuleSeveritySuccess:
		return true
	default:
		return false
	}
}

func (s *service) authorizeRuleWriter(ctx context.Context, identifier string, eventID int64) error {
	ev, err := s.repo.EventByID(ctx, eventID)
	if err != nil {
		return err
	}
	return s.ensureOrganizerMember(ctx, identifier, ev.OrganizerID)
}

func (s *service) ensureOrganizerMember(ctx context.Context, identifier string, organizerID int64) error {
	active, err := s.repo.IsActiveOrganizerMember(ctx, identifier, organizerID)
	if err != nil {
		return err
	}
	if !active {
		return apperrors.ErrForbidden
	}
	return nil
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

// ListNearbySchedules menentukan asal pencarian, lalu mengambil jadwal pada event terdekat.
func (s *service) ListNearbySchedules(ctx context.Context, identifier string, radiusKM int, categoryID int64, limit int) ([]entities.EventSchedule, error) {
	candidate, err := s.repo.Origin(ctx, identifier)
	if err != nil {
		return nil, err
	}
	radius := effectiveRadius(radiusKM, candidate.RadiusKM)
	if radius < 1 || radius > maxRadiusKM {
		return nil, apperrors.ErrInvalidRadius
	}
	if limit < 1 || limit > maxLimit {
		return nil, apperrors.ErrInvalidPayload
	}
	origin, _ := resolveOrigin(candidate, s.now())
	if origin == nil {
		return []entities.EventSchedule{}, nil
	}
	items, err := s.repo.ListNearbySchedules(ctx, *origin, radius, categoryID, limit)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []entities.EventSchedule{}
	}
	return items, nil
}
