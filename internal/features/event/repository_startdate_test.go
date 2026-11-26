package event

import (
	"testing"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

func TestRepositoryListNearbySchedulesStartDateCoverageDB(t *testing.T) {
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
	if err := conn.QueryRow(ctx, `INSERT INTO events (organizer_id, title, slug, venue, city, status, latitude, longitude, end_date, start_date, total_slots) VALUES ($1, 'Event', $2, 'GOR', 'Bandung', 'published', $3, $4, NULL, $5, 10) RETURNING id`, organizerID, "event-"+suffix, cityLat, cityLng, time.Now()).Scan(&eventID); err != nil {
		t.Fatalf("insert event: %v", err)
	}
	if _, err := conn.Exec(ctx, `INSERT INTO sub_events (event_id, bird_category_id, title, price, status) VALUES ($1, $2, 'Sesi', 100000, 'available')`, eventID, categoryID); err != nil {
		t.Fatalf("insert sub event: %v", err)
	}

	origin := entities.NearbyOrigin{Latitude: cityLat, Longitude: cityLng, Source: OriginCurrentLocation, CityID: &cityID}
	items, err := repo.ListNearbySchedules(ctx, origin, 50, 0, 10)
	if err != nil {
		t.Fatalf("ListNearbySchedules() error = %v", err)
	}
	if len(items) == 0 || items[0].Date == nil {
		t.Fatal("Date nil")
	}
	_ = categoryID
}
