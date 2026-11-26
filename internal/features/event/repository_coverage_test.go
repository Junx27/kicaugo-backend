package event

import (
	"context"
	"testing"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

func TestRepositoryCountGantanganCoverage(t *testing.T) {
	repo, ctx, conn := newTestRepository(t)
	organizerID := createOrganizer(t, conn, ctx)
	suffix := testSuffix(t)
	if _, err := conn.Exec(ctx, `INSERT INTO events (organizer_id, title, slug, venue, city, status, total_slots) VALUES ($1, 'E', $2, 'V', 'C', 'published', 7)`, organizerID, "event-"+suffix); err != nil {
		t.Fatalf("insert event: %v", err)
	}
	got, err := repo.CountGantangan(ctx)
	if err != nil {
		t.Fatalf("CountGantangan() error = %v", err)
	}
	if got < 7 {
		t.Fatalf("CountGantangan() = %d, ingin minimal 7", got)
	}
}

func TestRepositoryCountGantanganCanceledCoverage(t *testing.T) {
	repo, ctx, conn := newTestRepository(t)
	_ = conn
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := repo.CountGantangan(canceled); err == nil {
		t.Fatal("ingin error")
	}
}

func TestRepositoryListNearbySchedulesCoverage(t *testing.T) {
	repo, ctx, conn := newTestRepository(t)
	organizerID := createOrganizer(t, conn, ctx)
	suffix := testSuffix(t)

	var cityID int64
	var cityLat, cityLng float64
	if err := conn.QueryRow(ctx, `SELECT id, latitude, longitude FROM cities ORDER BY id LIMIT 1`).Scan(&cityID, &cityLat, &cityLng); err != nil {
		t.Fatalf("mengambil city: %v", err)
	}
	var categoryID int64
	if err := conn.QueryRow(ctx, `SELECT id FROM bird_categories ORDER BY id LIMIT 1`).Scan(&categoryID); err != nil {
		t.Fatalf("mengambil kategori: %v", err)
	}

	var eventID int64
	if err := conn.QueryRow(ctx, `INSERT INTO events (organizer_id, title, slug, venue, city, status, latitude, longitude, end_date, total_slots) VALUES ($1, 'Event', $2, 'GOR', 'Bandung', 'published', $3, $4, NULL, 10) RETURNING id`, organizerID, "event-"+suffix, cityLat, cityLng).Scan(&eventID); err != nil {
		t.Fatalf("insert event: %v", err)
	}
	start := time.Now()
	if _, err := conn.Exec(ctx, `INSERT INTO sub_events (event_id, bird_category_id, title, price, status, start_time) VALUES ($1, $2, 'Sesi', 100000, 'available', $3)`, eventID, categoryID, start); err != nil {
		t.Fatalf("insert sub event: %v", err)
	}

	origin := entities.NearbyOrigin{Latitude: cityLat, Longitude: cityLng, Source: OriginCurrentLocation, CityID: &cityID}
	items, err := repo.ListNearbySchedules(ctx, origin, 50, 0, 10)
	if err != nil {
		t.Fatalf("ListNearbySchedules() error = %v", err)
	}
	if len(items) == 0 {
		t.Fatal("ListNearbySchedules() ingin minimal 1 item")
	}
	if items[0].Date == nil {
		t.Fatal("Date nil")
	}
}

func TestRepositoryListNearbySchedulesEmptyCoverage(t *testing.T) {
	repo, ctx, conn := newTestRepository(t)
	var cityID int64
	var cityLat, cityLng float64
	if err := conn.QueryRow(ctx, `SELECT id, latitude, longitude FROM cities ORDER BY id LIMIT 1`).Scan(&cityID, &cityLat, &cityLng); err != nil {
		t.Fatalf("mengambil city: %v", err)
	}
	origin := entities.NearbyOrigin{Latitude: cityLat + 100, Longitude: cityLng + 100, Source: OriginCurrentLocation, CityID: &cityID}
	items, err := repo.ListNearbySchedules(ctx, origin, 1, 0, 10)
	if err != nil {
		t.Fatalf("ListNearbySchedules() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("len = %d, ingin 0", len(items))
	}
}

func TestRepositoryListNearbySchedulesCanceledQueryCoverage(t *testing.T) {
	repo, ctx, conn := newTestRepository(t)
	_ = conn
	var cityID int64
	var cityLat, cityLng float64
	if err := conn.QueryRow(ctx, `SELECT id, latitude, longitude FROM cities ORDER BY id LIMIT 1`).Scan(&cityID, &cityLat, &cityLng); err != nil {
		t.Fatalf("mengambil city: %v", err)
	}
	origin := entities.NearbyOrigin{Latitude: cityLat, Longitude: cityLng, Source: OriginCurrentLocation, CityID: &cityID}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := repo.ListNearbySchedules(canceled, origin, 50, 0, 10); err == nil {
		t.Fatal("ingin error")
	}
}
