package event

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	"github.com/Junx27/kicaugo-backend/internal/testutil"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

// newTestRepository menyiapkan repository di dalam transaksi yang selalu di-rollback.
func newTestRepository(t *testing.T) (Repository, context.Context, db) {
	t.Helper()
	pool := testutil.OpenDB(t)
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin transaksi test: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })
	return &repository{db: tx}, ctx, tx
}

func testSuffix(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func createOrganizer(t *testing.T, conn db, ctx context.Context) int64 {
	t.Helper()
	suffix := testSuffix(t)
	var id int64
	err := conn.QueryRow(ctx, `
		INSERT INTO organizers (name, slug, email)
		VALUES ('Penyelenggara Test', $1, $2) RETURNING id`,
		"penyelenggara-"+suffix, "organizer-"+suffix+"@example.com",
	).Scan(&id)
	if err != nil {
		t.Fatalf("membuat organizer: %v", err)
	}
	return id
}

type eventFixture struct {
	slug       string
	latitude   *float64
	longitude  *float64
	cityID     *int64
	status     string
	endDate    *time.Time
	categoryID *int64
}

func insertEvent(t *testing.T, conn db, ctx context.Context, organizerID int64, f eventFixture) int64 {
	t.Helper()
	var id int64
	err := conn.QueryRow(ctx, `
		INSERT INTO events (organizer_id, title, slug, venue, city, province,
			latitude, longitude, city_id, status, end_date)
		VALUES ($1, 'Event ' || $2, $3, 'GOR Kota', 'Bandung', 'Jawa Barat',
			$4, $5, $6, $7, $8) RETURNING id`,
		organizerID, f.slug, "event-"+f.slug, f.latitude, f.longitude, f.cityID, f.status, f.endDate,
	).Scan(&id)
	if err != nil {
		t.Fatalf("membuat event %q: %v", f.slug, err)
	}
	if f.categoryID != nil {
		if _, err := conn.Exec(ctx, `
			INSERT INTO sub_events (event_id, bird_category_id, title)
			VALUES ($1, $2, 'Sesi ' || $3)`, id, *f.categoryID, f.slug); err != nil {
			t.Fatalf("membuat sub event: %v", err)
		}
	}
	return id
}

func TestRepositoryListNearby(t *testing.T) {
	repo, ctx, conn := newTestRepository(t)
	organizerID := createOrganizer(t, conn, ctx)

	var cityID int64
	var cityLat, cityLng float64
	if err := conn.QueryRow(ctx, `SELECT id, latitude, longitude FROM cities ORDER BY id LIMIT 1`).Scan(&cityID, &cityLat, &cityLng); err != nil {
		t.Fatalf("mengambil city: %v", err)
	}
	var categoryID int64
	if err := conn.QueryRow(ctx, `SELECT id FROM bird_categories ORDER BY id LIMIT 1`).Scan(&categoryID); err != nil {
		t.Fatalf("mengambil kategori: %v", err)
	}

	originLat, originLng := cityLat, cityLng-0.02
	expired := time.Now().AddDate(0, 0, -2)
	nearID := insertEvent(t, conn, ctx, organizerID, eventFixture{
		slug: "dekat-" + testSuffix(t), latitude: float64Ptr(originLat), longitude: float64Ptr(originLng + 0.01), status: "published",
	})
	fallbackID := insertEvent(t, conn, ctx, organizerID, eventFixture{
		slug: "kota-" + testSuffix(t), cityID: &cityID, status: "ongoing",
	})
	categoryEventID := insertEvent(t, conn, ctx, organizerID, eventFixture{
		slug: "kategori-" + testSuffix(t), latitude: float64Ptr(originLat), longitude: float64Ptr(originLng + 0.01),
		status: "published", categoryID: &categoryID,
	})
	insertEvent(t, conn, ctx, organizerID, eventFixture{
		slug: "draft-" + testSuffix(t), latitude: float64Ptr(originLat), longitude: float64Ptr(originLng), status: "draft",
	})
	insertEvent(t, conn, ctx, organizerID, eventFixture{
		slug: "kedaluwarsa-" + testSuffix(t), latitude: float64Ptr(originLat), longitude: float64Ptr(originLng),
		status: "published", endDate: &expired,
	})
	insertEvent(t, conn, ctx, organizerID, eventFixture{
		slug: "jauh-" + testSuffix(t), latitude: float64Ptr(originLat), longitude: float64Ptr(originLng + 15), status: "published",
	})

	origin := entities.NearbyOrigin{Latitude: originLat, Longitude: originLng, Source: OriginCurrentLocation, CityID: &cityID}
	events, err := repo.ListNearby(ctx, origin, 50, 0, 50)
	if err != nil {
		t.Fatalf("ListNearby() error = %v", err)
	}
	ids := make(map[int64]bool, len(events))
	for _, e := range events {
		ids[e.ID] = true
		if e.DistanceKM < 0 {
			t.Errorf("distance_km negatif: %v", e.DistanceKM)
		}
	}
	for _, want := range []int64{nearID, fallbackID, categoryEventID} {
		if !ids[want] {
			t.Errorf("event %d tidak ikut hasil; ids = %v", want, ids)
		}
	}
	if len(events) != 3 {
		t.Fatalf("jumlah event = %d, ingin 3", len(events))
	}
	for i := 1; i < len(events); i++ {
		if events[i-1].DistanceKM > events[i].DistanceKM {
			t.Errorf("urutan jarak salah: %v lalu %v", events[i-1].DistanceKM, events[i].DistanceKM)
		}
	}

	// Fallback koordinat memakai pusat kota yang tidak NULL.
	var fallback entities.Event
	for _, e := range events {
		if e.ID == fallbackID {
			fallback = e
		}
	}
	if fallback.Latitude != nil || fallback.CityID == nil || *fallback.CityID != cityID {
		t.Errorf("fallback event = %+v", fallback)
	}
	wantCityDistance := haversineKM(originLat, originLng, cityLat, cityLng)
	if diff := fallback.DistanceKM - wantCityDistance; diff > 0.1 || diff < -0.1 {
		t.Errorf("jarak fallback = %v, ingin sekitar %v", fallback.DistanceKM, wantCityDistance)
	}

	filtered, err := repo.ListNearby(ctx, origin, 50, categoryID, 50)
	if err != nil || len(filtered) != 1 || filtered[0].ID != categoryEventID {
		t.Fatalf("filter kategori = %+v, %v", filtered, err)
	}
	limited, err := repo.ListNearby(ctx, origin, 50, 0, 1)
	if err != nil || len(limited) != 1 || limited[0].ID != events[0].ID {
		t.Fatalf("limit = %+v, %v", limited, err)
	}
	none, err := repo.ListNearby(ctx, origin, 1, 0, 50)
	if err != nil || len(none) != 0 {
		t.Fatalf("radius kecil = %+v, %v", none, err)
	}
}

func TestRepositoryCountOngoing(t *testing.T) {
	repo, ctx, conn := newTestRepository(t)
	organizerID := createOrganizer(t, conn, ctx)

	before, err := repo.CountOngoing(ctx)
	if err != nil {
		t.Fatalf("CountOngoing() awal error = %v", err)
	}
	future := time.Now().AddDate(0, 0, 2)
	expired := time.Now().AddDate(0, 0, -2)
	insertEvent(t, conn, ctx, organizerID, eventFixture{slug: "berlangsung-" + testSuffix(t), status: "ongoing"})
	insertEvent(t, conn, ctx, organizerID, eventFixture{slug: "berlangsung-akhir-" + testSuffix(t), status: "ongoing", endDate: &future})
	insertEvent(t, conn, ctx, organizerID, eventFixture{slug: "selesai-" + testSuffix(t), status: "ongoing", endDate: &expired})
	insertEvent(t, conn, ctx, organizerID, eventFixture{slug: "diterbitkan-" + testSuffix(t), status: "published"})

	after, err := repo.CountOngoing(ctx)
	if err != nil {
		t.Fatalf("CountOngoing() error = %v", err)
	}
	if after-before != 2 {
		t.Errorf("jumlah event berlangsung bertambah %d, ingin 2", after-before)
	}
}

func haversineKM(lat1, lng1, lat2, lng2 float64) float64 {
	const toRad = math.Pi / 180
	dLat := (lat2 - lat1) * toRad
	dLng := (lng2 - lng1) * toRad
	a := math.Pow(math.Sin(dLat/2), 2) + math.Cos(lat1*toRad)*math.Cos(lat2*toRad)*math.Pow(math.Sin(dLng/2), 2)
	return 6371.0088 * 2 * math.Asin(math.Sqrt(math.Min(1, math.Max(-1, a))))
}

func TestRepositoryOrigin(t *testing.T) {
	repo, ctx, conn := newTestRepository(t)
	identifier := "nearby-test-" + testSuffix(t) + "@example.com"
	var userID int64
	if err := conn.QueryRow(ctx, `
		INSERT INTO users (name, identifier, password_hash)
		VALUES ('Nearby Test', $1, 'rahasia') RETURNING id`, identifier).Scan(&userID); err != nil {
		t.Fatalf("membuat user: %v", err)
	}

	candidate, err := repo.Origin(ctx, identifier)
	if err != nil || candidate.CityID != nil || candidate.RadiusKM != 50 || candidate.CurrentLatitude != nil {
		t.Fatalf("Origin() = %+v, %v", candidate, err)
	}

	var cityID int64
	if err := conn.QueryRow(ctx, `SELECT id FROM cities ORDER BY id LIMIT 1`).Scan(&cityID); err != nil {
		t.Fatalf("mengambil city: %v", err)
	}
	updated := time.Now()
	if _, err := conn.Exec(ctx, `
		UPDATE user_profiles
		SET city_id = $2, current_latitude = -6.2, current_longitude = 106.8,
			current_location_accuracy_m = 10, current_location_updated_at = $3,
			search_radius_km = 25
		WHERE user_id = $1`, userID, cityID, updated); err != nil {
		t.Fatalf("mengubah profil: %v", err)
	}
	candidate, err = repo.Origin(ctx, identifier)
	if err != nil || candidate.CurrentLatitude == nil || *candidate.CurrentLatitude != -6.2 ||
		candidate.CityLatitude == nil || candidate.CityID == nil || *candidate.CityID != cityID ||
		candidate.RadiusKM != 25 {
		t.Fatalf("Origin() setelah update = %+v, %v", candidate, err)
	}

	if _, err := conn.Exec(ctx, `DELETE FROM user_profiles WHERE user_id = $1`, userID); err != nil {
		t.Fatalf("menghapus profil: %v", err)
	}
	candidate, err = repo.Origin(ctx, identifier)
	if err != nil || candidate.RadiusKM != 50 {
		t.Fatalf("Origin() lazy-create = %+v, %v", candidate, err)
	}

	if _, err := repo.Origin(ctx, "tidak-ada@example.com"); !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("Origin() = %v, ingin ErrNotFound", err)
	}
}

func TestRepositoryErrorPaths(t *testing.T) {
	pool := testutil.OpenDB(t)
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	repo := NewRepository(pool)
	if _, err := repo.Origin(canceled, "x"); err == nil {
		t.Error("Origin() dengan konteks dibatalkan ingin error")
	}
	if _, err := repo.ListNearby(canceled, entities.NearbyOrigin{Latitude: -6.2, Longitude: 106.8}, 50, 0, 20); err == nil {
		t.Error("ListNearby() dengan konteks dibatalkan ingin error")
	}
	if _, err := repo.CountOngoing(canceled); err == nil {
		t.Error("CountOngoing() dengan konteks dibatalkan ingin error")
	}
}
