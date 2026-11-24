package profile

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	"github.com/Junx27/kicaugo-backend/internal/testutil"
	dbutil "github.com/Junx27/kicaugo-backend/internal/utils/db"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

// newTestRepository menyiapkan repository di dalam transaksi yang selalu di-rollback.
func newTestRepository(t *testing.T) (Repository, context.Context, dbutil.Querier) {
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

func newTestUser(t *testing.T, conn dbutil.Querier, ctx context.Context) (int64, string) {
	t.Helper()
	identifier := fmt.Sprintf("profile-test-%d@example.com", time.Now().UnixNano())
	var id int64
	if err := conn.QueryRow(ctx, `
		INSERT INTO users (name, identifier, password_hash)
		VALUES ('Profile Test', $1, 'rahasia') RETURNING id`, identifier).Scan(&id); err != nil {
		t.Fatalf("membuat user test: %v", err)
	}
	return id, identifier
}

func TestRepositoryProfileFlow(t *testing.T) {
	repo, ctx, conn := newTestRepository(t)
	userID, identifier := newTestUser(t, conn, ctx)

	resolved, err := repo.ResolveUserID(ctx, identifier)
	if err != nil || resolved != userID {
		t.Fatalf("ResolveUserID() = %d, %v; ingin %d, nil", resolved, err, userID)
	}
	if _, err := repo.ResolveUserID(ctx, "tidak-ada@example.com"); !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("ResolveUserID() = %v, ingin ErrNotFound", err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := repo.ResolveUserID(canceled, identifier); err == nil {
		t.Error("ResolveUserID() dengan konteks dibatalkan ingin error")
	}

	// Trigger membuat profil saat user dibuat; verifikasi lazy-create untuk user lama.
	if _, err := repo.EnsureAndGet(ctx, -1); !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("EnsureAndGet(-1) = %v, ingin ErrNotFound", err)
	}
	p, err := repo.EnsureAndGet(ctx, userID)
	if err != nil || p.UserID != userID || p.FullName != "Profile Test" {
		t.Fatalf("EnsureAndGet() = %+v, %v", p, err)
	}
	if _, err := conn.Exec(ctx, `DELETE FROM user_profiles WHERE user_id = $1`, userID); err != nil {
		t.Fatalf("menghapus profil: %v", err)
	}
	p, err = repo.EnsureAndGet(ctx, userID)
	if err != nil || p.ID == 0 || p.FullName != "Profile Test" {
		t.Fatalf("EnsureAndGet() lazy-create = %+v, %v", p, err)
	}

	var cityID int64
	if err := conn.QueryRow(ctx, `SELECT id FROM cities ORDER BY id LIMIT 1`).Scan(&cityID); err != nil {
		t.Fatalf("mengambil city: %v", err)
	}
	city, err := repo.City(ctx, cityID)
	if err != nil || city.ID != cityID {
		t.Fatalf("City() = %+v, %v", city, err)
	}
	if _, err := repo.City(ctx, -1); !errors.Is(err, apperrors.ErrInvalidCity) {
		t.Errorf("City(-1) = %v, ingin ErrInvalidCity", err)
	}

	phone := "+6281234567890"
	err = repo.Update(ctx, userID, entities.ProfileUpdate{
		FullName: stringPtr("Budi"), Phone: &phone, CityID: &cityID, SearchRadiusKM: intPtr(25),
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if err := repo.Update(ctx, userID, entities.ProfileUpdate{}); err != nil {
		t.Errorf("Update() kosong error = %v", err)
	}
	p, err = repo.EnsureAndGet(ctx, userID)
	if err != nil || p.FullName != "Budi" || p.City == nil || p.City.ID != cityID ||
		p.Province == nil || p.SearchRadiusKM != 25 || p.Phone == nil {
		t.Fatalf("profil setelah Update() = %+v, %v", p, err)
	}

	lat, lng := -6.2, 106.8
	acc := 20
	if err := repo.UpdateLocation(ctx, userID, &lat, &lng, &acc); err != nil {
		t.Fatalf("UpdateLocation() error = %v", err)
	}
	p, err = repo.EnsureAndGet(ctx, userID)
	if err != nil || p.CurrentLatitude == nil || *p.CurrentLatitude != lat ||
		p.CurrentLongitude == nil || p.CurrentLocationUpdatedAt == nil {
		t.Fatalf("profil setelah UpdateLocation() = %+v, %v", p, err)
	}
	if err := repo.UpdateLocation(ctx, userID, nil, nil, nil); err != nil {
		t.Fatalf("UpdateLocation() hapus error = %v", err)
	}
	p, err = repo.EnsureAndGet(ctx, userID)
	if err != nil || p.CurrentLatitude != nil || p.CurrentLongitude != nil || p.CurrentLocationUpdatedAt != nil {
		t.Fatalf("profil setelah hapus lokasi = %+v, %v", p, err)
	}

	var categoryID int64
	if err := conn.QueryRow(ctx, `SELECT id FROM bird_categories ORDER BY id LIMIT 1`).Scan(&categoryID); err != nil {
		t.Fatalf("mengambil kategori: %v", err)
	}
	count, err := repo.CountValidCategories(ctx, []int64{categoryID, -1})
	if err != nil || count != 1 {
		t.Fatalf("CountValidCategories() = %d, %v; ingin 1, nil", count, err)
	}
	if err := repo.ReplaceFavorites(ctx, userID, []int64{categoryID}); err != nil {
		t.Fatalf("ReplaceFavorites() error = %v", err)
	}
	favorites, err := repo.FavoriteCategories(ctx, userID)
	if err != nil || len(favorites) != 1 || favorites[0].ID != categoryID {
		t.Fatalf("FavoriteCategories() = %+v, %v", favorites, err)
	}
	if err := repo.ReplaceFavorites(ctx, userID, []int64{}); err != nil {
		t.Fatalf("ReplaceFavorites() kosong error = %v", err)
	}
	favorites, err = repo.FavoriteCategories(ctx, userID)
	if err != nil || len(favorites) != 0 {
		t.Fatalf("FavoriteCategories() setelah kosong = %+v, %v", favorites, err)
	}
}

func TestRepositoryErrorPaths(t *testing.T) {
	pool := testutil.OpenDB(t)
	ctx := context.Background()
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	repo := NewRepository(pool)

	if _, err := repo.EnsureAndGet(canceled, 1); err == nil {
		t.Error("EnsureAndGet() dengan konteks dibatalkan ingin error")
	}
	if _, err := repo.FavoriteCategories(canceled, 1); err == nil {
		t.Error("FavoriteCategories() dengan konteks dibatalkan ingin error")
	}
	if _, err := repo.City(canceled, 1); err == nil {
		t.Error("City() dengan konteks dibatalkan ingin error")
	}
	if _, err := repo.CountValidCategories(canceled, []int64{1}); err == nil {
		t.Error("CountValidCategories() dengan konteks dibatalkan ingin error")
	}
	if err := repo.Update(canceled, 1, entities.ProfileUpdate{Bio: stringPtr("x")}); err == nil {
		t.Error("Update() dengan konteks dibatalkan ingin error")
	}
	if err := repo.UpdateLocation(canceled, 1, nil, nil, nil); err == nil {
		t.Error("UpdateLocation() dengan konteks dibatalkan ingin error")
	}
	lat, lng := -6.2, 106.8
	if err := repo.UpdateLocation(canceled, 1, &lat, &lng, nil); err == nil {
		t.Error("UpdateLocation() set dengan konteks dibatalkan ingin error")
	}
	if err := repo.ReplaceFavorites(canceled, 1, []int64{1}); err == nil {
		t.Error("ReplaceFavorites() dengan konteks dibatalkan ingin error")
	}
}

func TestBuildUpdate(t *testing.T) {
	sets, args := buildUpdate(7, entities.ProfileUpdate{
		FullName: stringPtr("Budi"), Phone: stringPtr("+62812"),
		AvatarURL: stringPtr("url"), Bio: stringPtr("bio"),
		CityID: int64Ptr(5), SearchRadiusKM: intPtr(25),
	})
	if len(sets) != 6 || len(args) != 7 {
		t.Fatalf("sets = %v, args = %v", sets, args)
	}
	for i, want := range []string{
		"full_name = $2", "phone = $3", "avatar_url = $4",
		"bio = $5", "city_id = $6", "search_radius_km = $7",
	} {
		if sets[i] != want {
			t.Errorf("sets[%d] = %q, ingin %q", i, sets[i], want)
		}
	}
	if args[0] != int64(7) {
		t.Errorf("args[0] = %v, ingin 7", args[0])
	}
	if sets, args := buildUpdate(7, entities.ProfileUpdate{}); len(sets) != 0 || len(args) != 1 {
		t.Errorf("buildUpdate kosong = %v, %v", sets, args)
	}
}
