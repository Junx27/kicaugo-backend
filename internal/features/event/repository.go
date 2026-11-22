package event

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

// Repository menyediakan operasi event.
type Repository interface {
	Origin(ctx context.Context, identifier string) (entities.OriginCandidate, error)
	ListNearby(ctx context.Context, origin entities.NearbyOrigin, radiusKM int, categoryID int64, limit int) ([]entities.Event, error)
	CountOngoing(ctx context.Context) (int, error)
	IsActiveOrganizerMember(ctx context.Context, identifier string, organizerID int64) (bool, error)
	Create(ctx context.Context, in entities.EventCreate) (entities.EventDetail, error)
}

// db adalah kontrak minimal yang dipenuhi pgxpool.Pool maupun pgx.Tx.
type db interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

// repository adalah implementasi Repository berbasis pgx.
type repository struct {
	db db
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

// CountOngoing menghitung event berstatus ongoing yang belum terlewat tanggal akhirnya.
func (r *repository) CountOngoing(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM events
		WHERE status = 'ongoing' AND (end_date IS NULL OR end_date >= CURRENT_DATE)`
	var count int
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("count ongoing events: %w", err)
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
