package event

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	dbutil "github.com/Junx27/kicaugo-backend/internal/utils/db"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

const uniqueViolation = "23505"

// Repository menyediakan operasi event.
type Repository interface {
	Origin(ctx context.Context, identifier string) (entities.OriginCandidate, error)
	ListNearby(ctx context.Context, origin entities.NearbyOrigin, radiusKM int, categoryID int64, limit int) ([]entities.Event, error)
	ListNearbySchedules(ctx context.Context, origin entities.NearbyOrigin, radiusKM int, categoryID int64, limit int) ([]entities.EventSchedule, error)
	CountOngoing(ctx context.Context) (int, error)
	CountGantangan(ctx context.Context) (int, error)
	IsActiveOrganizerMember(ctx context.Context, identifier string, organizerID int64) (bool, error)
	Create(ctx context.Context, in entities.EventCreate) (entities.EventDetail, error)
	EventByID(ctx context.Context, eventID int64) (entities.Event, error)
	GetContestDetail(ctx context.Context, eventID int64) (entities.ContestDetail, error)
	HasActiveStreaming(ctx context.Context, eventID int64) (bool, error)
	CreateStreaming(ctx context.Context, in entities.StartEventStreamingInput) (entities.EventStreaming, error)
	UpdateStreamingJuara(ctx context.Context, in entities.UpdateEventStreamingJuaraInput) (entities.EventStreaming, error)
	StopStreaming(ctx context.Context, eventID int64) (entities.EventStreaming, error)
	GetActiveStreaming(ctx context.Context) (*entities.ActiveEventStreaming, error)
	GetRandomFeatured(ctx context.Context, identifier string) (*entities.FeaturedEvent, error)
	GetRandomUpcoming(ctx context.Context) (*entities.UpcomingEvent, error)
}

// repository adalah implementasi Repository berbasis pgx.
type repository struct {
	db dbutil.Querier
}

// NewRepository membuat repository dengan connection pool yang sudah ada.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{db: pool}
}

// Origin membaca asal pencarian dari profil, membuat profil bila belum ada.
func (r *repository) Origin(ctx context.Context, identifier string) (entities.OriginCandidate, error) {
	const ensure = `INSERT INTO user_profiles (user_id, full_name)
		SELECT u.id, u.name FROM users u WHERE u.identifier = $1
		ON CONFLICT (user_id) DO NOTHING`
	if _, err := r.db.Exec(ctx, ensure, identifier); err != nil {
		return entities.OriginCandidate{}, fmt.Errorf("ensure profile for nearby: %w", err)
	}

	const query = `SELECT p.current_latitude, p.current_longitude, p.current_location_updated_at,
		c.latitude, c.longitude, c.id, p.search_radius_km
		FROM users u
		JOIN user_profiles p ON p.user_id = u.id
		LEFT JOIN cities c ON c.id = p.city_id
		WHERE u.identifier = $1`
	var candidate entities.OriginCandidate
	err := r.db.QueryRow(ctx, query, identifier).Scan(
		&candidate.CurrentLatitude, &candidate.CurrentLongitude, &candidate.CurrentUpdatedAt,
		&candidate.CityLatitude, &candidate.CityLongitude, &candidate.CityID, &candidate.RadiusKM,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return entities.OriginCandidate{}, apperrors.ErrNotFound
	}
	if err != nil {
		return entities.OriginCandidate{}, fmt.Errorf("get nearby origin: %w", err)
	}
	return candidate, nil
}

// CountOngoing menghitung event yang sedang live/streaming aktif.
func (r *repository) CountOngoing(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM event_streamings WHERE end_time IS NULL`
	var count int
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("count ongoing events: %w", err)
	}
	return count, nil
}

// CountGantangan menjumlahkan total gantangan dari semua event terdaftar.
func (r *repository) CountGantangan(ctx context.Context) (int, error) {
	const query = `SELECT COALESCE(SUM(total_slots), 0) FROM events`
	var count int
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("count gantangan: %w", err)
	}
	return count, nil
}

// nearbyQuery menghitung jarak Haversine dalam satu helper SQL terparameterisasi.
// Koordinat venue diambil dari events, lalu jatuh ke pusat kota bila kosong.
const nearbyQuery = `WITH origin AS (
		SELECT $1::double precision AS lat, $2::double precision AS lng
	), candidates AS (
		SELECT e.id, e.title, e.slug, e.venue, e.city, COALESCE(e.province, '') AS province,
			e.city_id, e.province_id, e.latitude, e.longitude, e.start_date, e.end_date,
			e.status, e.banner_url,
			COALESCE(e.latitude, c.latitude) AS event_lat,
			COALESCE(e.longitude, c.longitude) AS event_lng
		FROM events e
		LEFT JOIN cities c ON c.id = e.city_id
		WHERE e.status IN ('published', 'ongoing')
		  AND (e.end_date IS NULL OR e.end_date >= CURRENT_DATE)
		  AND ($5::bigint = 0 OR EXISTS (
			SELECT 1 FROM sub_events se
			WHERE se.event_id = e.id AND se.bird_category_id = $5::bigint
		  ))
	), distances AS (
		SELECT c.*, 6371.0088 * 2 * asin(LEAST(1, GREATEST(-1, sqrt(
			power(sin(radians(c.event_lat - o.lat) / 2), 2) +
			cos(radians(o.lat)) * cos(radians(c.event_lat)) *
			power(sin(radians(c.event_lng - o.lng) / 2), 2)
		)))) AS distance_km
		FROM candidates c
		CROSS JOIN origin o
		WHERE c.event_lat IS NOT NULL AND c.event_lng IS NOT NULL
	)
	SELECT id, title, slug, venue, city, province, city_id, province_id,
		latitude, longitude, start_date, end_date, status, banner_url, distance_km
	FROM distances
	WHERE distance_km <= $3::double precision
	ORDER BY distance_km, id
	LIMIT $4::int`

// ListNearby mengambil event aktif dalam radius tertentu, diurutkan dari yang terdekat.
func (r *repository) ListNearby(ctx context.Context, origin entities.NearbyOrigin, radiusKM int, categoryID int64, limit int) ([]entities.Event, error) {
	rows, err := r.db.Query(ctx, nearbyQuery, origin.Latitude, origin.Longitude, radiusKM, limit, categoryID)
	if err != nil {
		return nil, fmt.Errorf("list nearby events: %w", err)
	}
	defer rows.Close()

	var out []entities.Event
	for rows.Next() {
		var e entities.Event
		if err := rows.Scan(
			&e.ID, &e.Title, &e.Slug, &e.Venue, &e.City, &e.Province,
			&e.CityID, &e.ProvinceID, &e.Latitude, &e.Longitude,
			&e.StartDate, &e.EndDate, &e.Status, &e.BannerURL, &e.DistanceKM,
		); err != nil {
			return nil, fmt.Errorf("scan nearby event: %w", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate nearby events: %w", err)
	}
	return out, nil
}

// nearbySchedulesQuery mengambil sub-event (jadwal) pada event terdekat.
const nearbySchedulesQuery = `WITH origin AS (
		SELECT $1::double precision AS lat, $2::double precision AS lng
	), candidates AS (
		SELECT e.id AS event_id, e.venue, e.city,
			COALESCE(e.latitude, c.latitude) AS event_lat,
			COALESCE(e.longitude, c.longitude) AS event_lng,
			se.id AS schedule_id, se.title, se.price, se.status, se.start_time,
			e.start_date, se.bird_category_id, bc.name AS category_name
		FROM events e
		JOIN sub_events se ON se.event_id = e.id
		LEFT JOIN cities c ON c.id = e.city_id
		LEFT JOIN bird_categories bc ON bc.id = se.bird_category_id
		WHERE e.status IN ('published', 'ongoing')
		  AND (e.end_date IS NULL OR e.end_date >= CURRENT_DATE)
		  AND ($5::bigint = 0 OR se.bird_category_id = $5::bigint)
	), distances AS (
		SELECT cand.*, 6371.0088 * 2 * asin(LEAST(1, GREATEST(-1, sqrt(
			power(sin(radians(cand.event_lat - o.lat) / 2), 2) +
			cos(radians(o.lat)) * cos(radians(cand.event_lat)) *
			power(sin(radians(cand.event_lng - o.lng) / 2), 2)
		)))) AS distance_km
		FROM candidates cand
		CROSS JOIN origin o
		WHERE cand.event_lat IS NOT NULL AND cand.event_lng IS NOT NULL
	)
	SELECT schedule_id, event_id, title, venue, price, status, start_time, start_date, bird_category_id, category_name, distance_km
	FROM distances
	WHERE distance_km <= $3::double precision
	ORDER BY distance_km, schedule_id
	LIMIT $4::int`

// ListNearbySchedules mengambil jadwal (sub-event) pada event aktif dalam radius tertentu.
func (r *repository) ListNearbySchedules(ctx context.Context, origin entities.NearbyOrigin, radiusKM int, categoryID int64, limit int) ([]entities.EventSchedule, error) {
	rows, err := r.db.Query(ctx, nearbySchedulesQuery, origin.Latitude, origin.Longitude, radiusKM, limit, categoryID)
	if err != nil {
		return nil, fmt.Errorf("list nearby schedules: %w", err)
	}
	defer rows.Close()

	var out []entities.EventSchedule
	for rows.Next() {
		var item entities.EventSchedule
		var startTime *time.Time
		var startDate *time.Time
		var categoryIDVal *int64
		var categoryName *string
		if err := rows.Scan(
			&item.ID, &item.EventID, &item.Title, &item.Venue, &item.Price, &item.Status,
			&startTime, &startDate, &categoryIDVal, &categoryName, &item.DistanceKM,
		); err != nil {
			return nil, fmt.Errorf("scan nearby schedule: %w", err)
		}
		if startTime != nil {
			item.Date = startTime
		} else {
			item.Date = startDate
		}
		item.CategoryID = categoryIDVal
		if categoryName != nil {
			item.CategoryName = *categoryName
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate nearby schedules: %w", err)
	}
	if out == nil {
		out = []entities.EventSchedule{}
	}
	return out, nil
}

// IsActiveOrganizerMember memeriksa keanggotaan organizer aktif milik identifier.
func (r *repository) IsActiveOrganizerMember(ctx context.Context, identifier string, organizerID int64) (bool, error) {
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

// Create menyimpan event dan sub-event dalam satu transaksi.
func (r *repository) Create(ctx context.Context, in entities.EventCreate) (entities.EventDetail, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return entities.EventDetail{}, fmt.Errorf("begin create event: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	event, err := insertEvent(ctx, tx, in)
	if err != nil {
		return entities.EventDetail{}, err
	}
	subEvents, err := insertSubEvents(ctx, tx, event.ID, in.SubEvents)
	if err != nil {
		return entities.EventDetail{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return entities.EventDetail{}, fmt.Errorf("commit create event: %w", err)
	}
	return entities.EventDetail{Event: event, SubEvents: subEvents}, nil
}

func insertEvent(ctx context.Context, tx pgx.Tx, in entities.EventCreate) (entities.Event, error) {
	const query = `INSERT INTO events (
		organizer_id, title, slug, description, venue, address, city, province,
		city_id, province_id, latitude, longitude, start_date, end_date,
		registration_open, registration_close, status, is_reusable, banner_url,
		prize_pool, total_slots
	) VALUES (
		$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21
	) RETURNING id, title, slug, venue, city, COALESCE(province,''), city_id, province_id, latitude, longitude, start_date, end_date, status, banner_url`
	var out entities.Event
	err := tx.QueryRow(ctx, query,
		in.OrganizerID, in.Title, in.Slug, in.Description, in.Venue, in.Address, in.City, in.Province,
		in.CityID, in.ProvinceID, in.Latitude, in.Longitude, in.StartDate, in.EndDate,
		in.RegistrationOpen, in.RegistrationClose, in.Status, in.IsReusable, in.BannerURL,
		in.PrizePool, in.TotalSlots,
	).Scan(&out.ID, &out.Title, &out.Slug, &out.Venue, &out.City, &out.Province, &out.CityID, &out.ProvinceID, &out.Latitude, &out.Longitude, &out.StartDate, &out.EndDate, &out.Status, &out.BannerURL)
	if err != nil {
		return entities.Event{}, mapCreateEventError(err)
	}
	out.OrganizerID = in.OrganizerID
	return out, nil
}

func insertSubEvents(ctx context.Context, tx pgx.Tx, eventID int64, items []entities.SubEventCreate) ([]entities.SubEvent, error) {
	out := make([]entities.SubEvent, 0, len(items))
	for _, item := range items {
		sub, err := insertSubEvent(ctx, tx, eventID, item)
		if err != nil {
			return nil, err
		}
		out = append(out, sub)
	}
	return out, nil
}

func insertSubEvent(ctx context.Context, tx pgx.Tx, eventID int64, in entities.SubEventCreate) (entities.SubEvent, error) {
	const query = `INSERT INTO sub_events (
		event_id, bird_category_id, title, description, session, gantangan_count,
		max_slots, price, winner_1_prize, winner_2_prize, winner_3_prize, status,
		start_time, end_time, sort_order
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
	RETURNING id, event_id, bird_category_id, title, description, session, gantangan_count, max_slots, price, winner_1_prize, winner_2_prize, winner_3_prize, status, start_time, end_time, sort_order`
	var out entities.SubEvent
	err := tx.QueryRow(ctx, query,
		eventID, in.BirdCategoryID, in.Title, in.Description, in.Session, in.GantanganCount,
		in.MaxSlots, in.Price, in.Winner1Prize, in.Winner2Prize, in.Winner3Prize, in.Status,
		in.StartTime, in.EndTime, in.SortOrder,
	).Scan(&out.ID, &out.EventID, &out.BirdCategoryID, &out.Title, &out.Description, &out.Session, &out.GantanganCount, &out.MaxSlots, &out.Price, &out.Winner1Prize, &out.Winner2Prize, &out.Winner3Prize, &out.Status, &out.StartTime, &out.EndTime, &out.SortOrder)
	if err != nil {
		return entities.SubEvent{}, mapCreateEventError(err)
	}
	return out, nil
}

func mapCreateEventError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return apperrors.ErrIdentifierTaken
	}
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		if strings.Contains(pgErr.ConstraintName, "bird_category") {
			return apperrors.ErrInvalidCategory
		}
		if strings.Contains(pgErr.ConstraintName, "organizer") {
			return apperrors.ErrOrganizerNotFound
		}
		return apperrors.ErrInvalidPayload
	}
	return fmt.Errorf("create event: %w", err)
}

// EventByID mengambil event ringkas untuk validasi streaming.
func (r *repository) EventByID(ctx context.Context, eventID int64) (entities.Event, error) {
	const query = `SELECT id, organizer_id, title, slug, venue, city, COALESCE(province, ''), city_id, province_id, latitude, longitude, start_date, end_date, status, banner_url
		FROM events WHERE id = $1`
	var out entities.Event
	err := r.db.QueryRow(ctx, query, eventID).Scan(
		&out.ID, &out.OrganizerID, &out.Title, &out.Slug, &out.Venue, &out.City, &out.Province,
		&out.CityID, &out.ProvinceID, &out.Latitude, &out.Longitude, &out.StartDate, &out.EndDate, &out.Status, &out.BannerURL,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return entities.Event{}, apperrors.ErrNotFound
	}
	if err != nil {
		return entities.Event{}, fmt.Errorf("get event by id: %w", err)
	}
	return out, nil
}

// GetContestDetail mengambil detail lomba beserta kelas dan aturan.
func (r *repository) GetContestDetail(ctx context.Context, eventID int64) (entities.ContestDetail, error) {
	const query = `SELECT id, title, COALESCE(subtitle, ''), COALESCE(official_badge, ''),
		COALESCE(format_badge, ''), COALESCE(edition_label, ''), hero_image_url, banner_url,
		status, total_slots,
		(SELECT COUNT(*) FROM event_tickets WHERE event_id = $1 AND status = 'active') AS active_tickets,
		start_date, COALESCE(schedule_time, ''), COALESCE(gate_open, ''),
		venue, address, COALESCE(distance_text, ''), map_image_url, latitude, longitude,
		COALESCE(parking, ''), prize_pool, COALESCE(prize_trophy, ''), COALESCE(prize_highlight, ''),
		COALESCE(stat_format_label, ''), COALESCE(stat_format_note, ''),
		COALESCE(stat_judging_label, ''), COALESCE(stat_judging_note, ''),
		COALESCE(stat_facilities_label, ''), COALESCE(stat_facilities_note, '')
		FROM events WHERE id = $1`
	var out entities.ContestDetail
	err := r.db.QueryRow(ctx, query, eventID).Scan(
		&out.ID, &out.Title, &out.Subtitle, &out.OfficialBadge, &out.FormatBadge, &out.EditionLabel,
		&out.HeroImageURL, &out.BannerImageURL, &out.Status, &out.TotalSlots, &out.ActiveTickets,
		&out.ScheduleDate, &out.ScheduleTime, &out.GateOpen, &out.Venue, &out.Address, &out.DistanceText,
		&out.MapImageURL, &out.Latitude, &out.Longitude, &out.Parking, &out.PrizePool, &out.PrizeTrophy,
		&out.PrizeHighlight, &out.StatFormatLabel, &out.StatFormatNote, &out.StatJudgingLabel, &out.StatJudgingNote,
		&out.StatFacilitiesLabel, &out.StatFacilitiesNote,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return entities.ContestDetail{}, apperrors.ErrNotFound
	}
	if err != nil {
		return entities.ContestDetail{}, fmt.Errorf("get contest detail: %w", err)
	}

	classRows, err := r.db.Query(ctx, `SELECT se.id, se.title, se.session, se.start_time, se.max_slots,
		(SELECT COUNT(*) FROM event_tickets et WHERE et.sub_event_id = se.id AND et.status = 'active') AS active_tickets,
		se.price, COALESCE(se.prize_highlight, '')
		FROM sub_events se WHERE se.event_id = $1 ORDER BY se.sort_order, se.id`, eventID)
	if err != nil {
		return entities.ContestDetail{}, fmt.Errorf("list contest classes: %w", err)
	}
	defer classRows.Close()
	for classRows.Next() {
		var c entities.ContestClass
		var maxSlots, active int
		if err := classRows.Scan(&c.ID, &c.Name, &c.Session, &c.StartTime, &maxSlots, &active, &c.Price, &c.PrizeHighlight); err != nil {
			return entities.ContestDetail{}, fmt.Errorf("scan contest class: %w", err)
		}
		remaining := maxSlots - active
		if remaining < 0 {
			remaining = 0
		}
		remainingPtr := remaining
		c.RemainingSlots = &remainingPtr
		out.Classes = append(out.Classes, c)
	}
	if err := classRows.Err(); err != nil {
		return entities.ContestDetail{}, fmt.Errorf("iterate contest classes: %w", err)
	}
	if out.Classes == nil {
		out.Classes = []entities.ContestClass{}
	}

	ruleRows, err := r.db.Query(ctx, `SELECT id, title, description, severity, sort_order
		FROM contest_rules WHERE event_id = $1 ORDER BY sort_order, id`, eventID)
	if err != nil {
		return entities.ContestDetail{}, fmt.Errorf("list contest rules: %w", err)
	}
	defer ruleRows.Close()
	for ruleRows.Next() {
		var rule entities.ContestRule
		if err := ruleRows.Scan(&rule.ID, &rule.Title, &rule.Description, &rule.Severity, &rule.SortOrder); err != nil {
			return entities.ContestDetail{}, fmt.Errorf("scan contest rule: %w", err)
		}
		out.Rules = append(out.Rules, rule)
	}
	if err := ruleRows.Err(); err != nil {
		return entities.ContestDetail{}, fmt.Errorf("iterate contest rules: %w", err)
	}
	if out.Rules == nil {
		out.Rules = []entities.ContestRule{}
	}
	return out, nil
}

// HasActiveStreaming memeriksa apakah event memiliki streaming yang masih aktif.
func (r *repository) HasActiveStreaming(ctx context.Context, eventID int64) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM event_streamings WHERE event_id = $1 AND end_time IS NULL)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, eventID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check active streaming: %w", err)
	}
	return exists, nil
}

// CreateStreaming menyimpan sesi streaming event.
func (r *repository) CreateStreaming(ctx context.Context, in entities.StartEventStreamingInput) (entities.EventStreaming, error) {
	const query = `INSERT INTO event_streamings (event_id, start_time, total_views, juara, link_stream)
		VALUES ($1, now(), 0, '{}'::jsonb, $2)
		RETURNING id, event_id, start_time, end_time, total_views, juara, link_stream, created_at, updated_at`
	var out entities.EventStreaming
	err := r.db.QueryRow(ctx, query, in.EventID, in.LinkStream).Scan(
		&out.ID, &out.EventID, &out.StartTime, &out.EndTime, &out.TotalViews, &out.Juara, &out.LinkStream, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return entities.EventStreaming{}, mapCreateStreamingError(err)
	}
	return out, nil
}

func mapCreateStreamingError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return apperrors.ErrStreamingAlreadyActive
	}
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return apperrors.ErrNotFound
	}
	return fmt.Errorf("create event streaming: %w", err)
}

// UpdateStreamingJuara memperbarui juara pada streaming aktif event.
func (r *repository) UpdateStreamingJuara(ctx context.Context, in entities.UpdateEventStreamingJuaraInput) (entities.EventStreaming, error) {
	const query = `UPDATE event_streamings
		SET juara = $2, updated_at = now()
		WHERE event_id = $1 AND end_time IS NULL
		RETURNING id, event_id, start_time, end_time, total_views, juara, link_stream, created_at, updated_at`
	var out entities.EventStreaming
	err := r.db.QueryRow(ctx, query, in.EventID, in.Juara).Scan(
		&out.ID, &out.EventID, &out.StartTime, &out.EndTime, &out.TotalViews, &out.Juara, &out.LinkStream, &out.CreatedAt, &out.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return entities.EventStreaming{}, apperrors.ErrStreamingNotActive
	}
	if err != nil {
		return entities.EventStreaming{}, fmt.Errorf("update event streaming juara: %w", err)
	}
	return out, nil
}

// StopStreaming menghentikan streaming aktif milik event.
func (r *repository) StopStreaming(ctx context.Context, eventID int64) (entities.EventStreaming, error) {
	const query = `UPDATE event_streamings
		SET end_time = now(), updated_at = now()
		WHERE event_id = $1 AND end_time IS NULL
		RETURNING id, event_id, start_time, end_time, total_views, juara, link_stream, created_at, updated_at`
	var out entities.EventStreaming
	err := r.db.QueryRow(ctx, query, eventID).Scan(
		&out.ID, &out.EventID, &out.StartTime, &out.EndTime, &out.TotalViews, &out.Juara, &out.LinkStream, &out.CreatedAt, &out.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return entities.EventStreaming{}, apperrors.ErrStreamingNotActive
	}
	if err != nil {
		return entities.EventStreaming{}, fmt.Errorf("stop event streaming: %w", err)
	}
	return out, nil
}

// GetActiveStreaming mengambil satu streaming aktif terbaru.
func (r *repository) GetActiveStreaming(ctx context.Context) (*entities.ActiveEventStreaming, error) {
	const query = `SELECT es.id, es.event_id, e.title, e.slug, e.venue, e.city, es.link_stream, es.start_time, es.end_time, es.total_views, es.juara
		FROM event_streamings es
		JOIN events e ON e.id = es.event_id
		WHERE es.end_time IS NULL
		ORDER BY es.start_time DESC, es.id DESC
		LIMIT 1`
	var out entities.ActiveEventStreaming
	err := r.db.QueryRow(ctx, query).Scan(
		&out.ID, &out.EventID, &out.EventTitle, &out.EventSlug, &out.Venue, &out.City,
		&out.LinkStream, &out.StartTime, &out.EndTime, &out.TotalViews, &out.Juara,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get active streaming: %w", err)
	}
	return &out, nil
}

// GetRandomFeatured mengambil satu event acak yang eligible untuk featured.
// Jika origin user tersedia, jarak dihitung dari origin; jika tidak, distance_km NULL.
func (r *repository) GetRandomFeatured(ctx context.Context, identifier string) (*entities.FeaturedEvent, error) {
	origin, originErr := r.Origin(ctx, identifier)
	var lat, lng *float64
	var hasOrigin bool
	if originErr == nil {
		if origin.CurrentLatitude != nil && origin.CurrentLongitude != nil {
			lat, lng, hasOrigin = origin.CurrentLatitude, origin.CurrentLongitude, true
		} else if origin.CityLatitude != nil && origin.CityLongitude != nil {
			lat, lng, hasOrigin = origin.CityLatitude, origin.CityLongitude, true
		}
	}
	if originErr != nil && !errors.Is(originErr, apperrors.ErrNotFound) {
		return nil, fmt.Errorf("get featured origin: %w", originErr)
	}
	// Tanpa origin (lokasi user/kota), jarak tidak dapat dihitung → tidak tampilkan featured.
	if !hasOrigin {
		return nil, nil
	}

	const query = `WITH candidates AS (
		SELECT e.id, e.title, e.start_date, e.venue, e.city, e.address, e.prize_pool, e.total_slots,
			COALESCE(e.latitude, c.latitude) AS event_lat,
			COALESCE(e.longitude, c.longitude) AS event_lng
		FROM events e
		LEFT JOIN cities c ON c.id = e.city_id
		WHERE e.status IN ('published', 'ongoing')
		  AND (e.end_date IS NULL OR e.end_date >= CURRENT_DATE)
	), distances AS (
		SELECT id, title, start_date, venue, city, address, prize_pool, total_slots, event_lat, event_lng,
			6371.0088 * 2 * asin(LEAST(1, GREATEST(-1, sqrt(
				power(sin(radians(event_lat - $1::double precision) / 2), 2) +
				cos(radians($1::double precision)) * cos(radians(event_lat)) *
				power(sin(radians(event_lng - $2::double precision) / 2), 2)
			)))) AS distance_km
		FROM candidates
		WHERE event_lat IS NOT NULL AND event_lng IS NOT NULL
	)
	SELECT id, title, start_date, venue, city, address, prize_pool, total_slots, event_lat, event_lng, distance_km
	FROM distances
	WHERE distance_km <= 50
	ORDER BY RANDOM()
	LIMIT 1`
	var latVal, lngVal float64
	latVal, lngVal = *lat, *lng
	var out entities.FeaturedEvent
	err := r.db.QueryRow(ctx, query, latVal, lngVal).Scan(
		&out.ID, &out.Title, &out.StartDate, &out.Venue, &out.City, &out.Address, &out.PrizePool, &out.TotalSlots, &out.Latitude, &out.Longitude, &out.DistanceKM,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get random featured event: %w", err)
	}
	return &out, nil
}

// GetRandomUpcoming mengambil satu event acak yang belum release (draft).
func (r *repository) GetRandomUpcoming(ctx context.Context) (*entities.UpcomingEvent, error) {
	const query = `SELECT e.id, e.title, e.start_date, e.registration_open,
		COALESCE((SELECT MIN(se.price) FROM sub_events se WHERE se.event_id = e.id), 0) AS min_price,
		GREATEST(
			e.total_slots - (
				SELECT COUNT(*) FROM event_tickets et
				WHERE et.event_id = e.id AND et.status = 'active'
			),
			0
		) AS remaining_tickets
		FROM events e
		WHERE e.status = 'draft'
		  AND (e.end_date IS NULL OR e.end_date >= CURRENT_DATE)
		ORDER BY RANDOM()
		LIMIT 1`
	var out entities.UpcomingEvent
	err := r.db.QueryRow(ctx, query).Scan(
		&out.ID, &out.Title, &out.StartDate, &out.RegistrationOpen, &out.MinPrice, &out.RemainingTickets,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get random upcoming event: %w", err)
	}
	return &out, nil
}
