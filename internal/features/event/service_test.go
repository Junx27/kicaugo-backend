package event

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

type fakeRepository struct {
	candidate    entities.OriginCandidate
	originErr    error
	events       []entities.Event
	listErr      error
	ongoingCount int
	countErr     error
	originID     string
	listOrigin   entities.NearbyOrigin
	listRadius   int
	listCat      int64
	listLimit    int
}

func (f *fakeRepository) Origin(ctx context.Context, identifier string) (entities.OriginCandidate, error) {
	f.originID = identifier
	return f.candidate, f.originErr
}

func (f *fakeRepository) ListNearby(ctx context.Context, origin entities.NearbyOrigin, radiusKM int, categoryID int64, limit int) ([]entities.Event, error) {
	f.listOrigin = origin
	f.listRadius = radiusKM
	f.listCat = categoryID
	f.listLimit = limit
	return f.events, f.listErr
}

func (f *fakeRepository) CountOngoing(ctx context.Context) (int, error) {
	return f.ongoingCount, f.countErr
}

func freshCandidate() entities.OriginCandidate {
	now := time.Now()
	return entities.OriginCandidate{
		CurrentLatitude: float64Ptr(-6.2), CurrentLongitude: float64Ptr(106.8),
		CurrentUpdatedAt: timePtr(now), RadiusKM: 30,
	}
}

func TestServiceListNearby(t *testing.T) {
	t.Run("berhasil memakai lokasi terkini dan radius profil", func(t *testing.T) {
		repo := &fakeRepository{
			candidate: freshCandidate(),
			events:    []entities.Event{{ID: 1, DistanceKM: 2.34}},
		}
		svc := NewService(repo)
		result, err := svc.ListNearby(context.Background(), "user@example.com", 0, 3, 10)
		if err != nil {
			t.Fatalf("ListNearby() error = %v", err)
		}
		if result.Origin == nil || result.Origin.Source != OriginCurrentLocation || result.RadiusKM != 30 {
			t.Errorf("result = %+v", result)
		}
		if len(result.Events) != 1 || repo.listRadius != 30 || repo.listCat != 3 || repo.listLimit != 10 {
			t.Errorf("events = %+v, query = %+v", result.Events, repo)
		}
		if repo.originID != "user@example.com" || repo.listOrigin.Latitude != -6.2 {
			t.Errorf("origin = %+v", repo.listOrigin)
		}
	})

	t.Run("radius permintaan mengalahkan radius profil", func(t *testing.T) {
		repo := &fakeRepository{candidate: freshCandidate()}
		result, err := NewService(repo).ListNearby(context.Background(), "x", 120, 0, 20)
		if err != nil || result.RadiusKM != 120 || repo.listRadius != 120 || repo.listCat != 0 {
			t.Errorf("result = %+v, err = %v", result, err)
		}
	})

	t.Run("tanpa asal mengembalikan hint tanpa query", func(t *testing.T) {
		repo := &fakeRepository{candidate: entities.OriginCandidate{RadiusKM: 50}}
		result, err := NewService(repo).ListNearby(context.Background(), "x", 0, 0, 20)
		if err != nil {
			t.Fatalf("ListNearby() error = %v", err)
		}
		if result.Origin != nil || result.Hint == nil || *result.Hint != HintLocationNotSet ||
			len(result.Events) != 0 || result.RadiusKM != 50 {
			t.Errorf("result = %+v", result)
		}
		if repo.listOrigin.Latitude != 0 || repo.listLimit != 0 {
			t.Error("ListNearby tidak boleh dipanggil tanpa asal")
		}
	})

	t.Run("radius tidak valid", func(t *testing.T) {
		repo := &fakeRepository{candidate: freshCandidate()}
		_, err := NewService(repo).ListNearby(context.Background(), "x", maxRadiusKM+1, 0, 20)
		if !errors.Is(err, apperrors.ErrInvalidRadius) {
			t.Errorf("ListNearby() = %v, ingin ErrInvalidRadius", err)
		}
	})

	t.Run("limit tidak valid", func(t *testing.T) {
		repo := &fakeRepository{candidate: freshCandidate()}
		_, err := NewService(repo).ListNearby(context.Background(), "x", 0, 0, maxLimit+1)
		if !errors.Is(err, apperrors.ErrInvalidPayload) {
			t.Errorf("ListNearby() = %v, ingin ErrInvalidPayload", err)
		}
	})

	t.Run("gagal membaca asal", func(t *testing.T) {
		svc := NewService(&fakeRepository{originErr: errors.New("db down")})
		if _, err := svc.ListNearby(context.Background(), "x", 0, 0, 20); err == nil {
			t.Error("ListNearby() ingin error")
		}
	})

	t.Run("gagal mengambil event", func(t *testing.T) {
		svc := NewService(&fakeRepository{candidate: freshCandidate(), listErr: errors.New("db down")})
		if _, err := svc.ListNearby(context.Background(), "x", 0, 0, 20); err == nil {
			t.Error("ListNearby() ingin error")
		}
	})
}

func TestServiceCountOngoing(t *testing.T) {
	t.Run("meneruskan jumlah dari repository", func(t *testing.T) {
		count, err := NewService(&fakeRepository{ongoingCount: 4}).CountOngoing(context.Background())
		if err != nil || count != 4 {
			t.Errorf("CountOngoing() = %d, %v; ingin 4, nil", count, err)
		}
	})

	t.Run("gagal menghitung dari repository", func(t *testing.T) {
		svc := NewService(&fakeRepository{countErr: errors.New("db down")})
		if _, err := svc.CountOngoing(context.Background()); err == nil {
			t.Error("CountOngoing() ingin error")
		}
	})
}
