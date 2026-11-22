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
	activeMember bool
	activeErr    error
	createResult entities.EventDetail
	createErr    error
	createIn     entities.EventCreate
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

func (f *fakeRepository) IsActiveOrganizerMember(ctx context.Context, identifier string, organizerID int64) (bool, error) {
	return f.activeMember, f.activeErr
}

func (f *fakeRepository) Create(ctx context.Context, in entities.EventCreate) (entities.EventDetail, error) {
	f.createIn = in
	return f.createResult, f.createErr
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

func TestServiceCreate(t *testing.T) {
	valid := entities.EventCreate{
		OrganizerID: 1, Title: " Lomba ", Slug: " lomba ", Venue: "GOR", City: "Bandung", Status: "draft",
		SubEvents: []entities.SubEventCreate{{Title: " Kelas A ", MaxSlots: 60, Status: "available"}},
	}

	t.Run("berhasil", func(t *testing.T) {
		repo := &fakeRepository{activeMember: true, createResult: entities.EventDetail{Event: entities.Event{ID: 9}}}
		got, err := NewService(repo).Create(context.Background(), "user@example.com", valid)
		if err != nil || got.Event.ID != 9 || repo.createIn.Title != "Lomba" || repo.createIn.SubEvents[0].Title != "Kelas A" {
			t.Fatalf("Create() = %+v, %v, in=%+v", got, err, repo.createIn)
		}
	})

	t.Run("payload invalid", func(t *testing.T) {
		_, err := NewService(&fakeRepository{activeMember: true}).Create(context.Background(), "x", entities.EventCreate{})
		if !errors.Is(err, apperrors.ErrInvalidPayload) {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("bukan organizer aktif", func(t *testing.T) {
		_, err := NewService(&fakeRepository{activeMember: false}).Create(context.Background(), "x", valid)
		if !errors.Is(err, apperrors.ErrForbidden) {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("cek membership error", func(t *testing.T) {
		_, err := NewService(&fakeRepository{activeErr: errors.New("db down")}).Create(context.Background(), "x", valid)
		if err == nil {
			t.Fatal("ingin error")
		}
	})

	t.Run("sub event invalid", func(t *testing.T) {
		bad := valid
		bad.SubEvents = []entities.SubEventCreate{{Title: "A", MaxSlots: 61, Status: "available"}}
		_, err := NewService(&fakeRepository{activeMember: true}).Create(context.Background(), "x", bad)
		if !errors.Is(err, apperrors.ErrInvalidPayload) {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("status event invalid", func(t *testing.T) {
		bad := valid
		bad.Status = "unknown"
		_, err := NewService(&fakeRepository{activeMember: true}).Create(context.Background(), "x", bad)
		if !errors.Is(err, apperrors.ErrInvalidPayload) {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("total slots invalid", func(t *testing.T) {
		bad := valid
		bad.TotalSlots = -1
		_, err := NewService(&fakeRepository{activeMember: true}).Create(context.Background(), "x", bad)
		if !errors.Is(err, apperrors.ErrInvalidPayload) {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("sub event status invalid", func(t *testing.T) {
		bad := valid
		bad.SubEvents = []entities.SubEventCreate{{Title: "A", MaxSlots: 10, Status: "bad"}}
		_, err := NewService(&fakeRepository{activeMember: true}).Create(context.Background(), "x", bad)
		if !errors.Is(err, apperrors.ErrInvalidPayload) {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("repo create error", func(t *testing.T) {
		_, err := NewService(&fakeRepository{activeMember: true, createErr: errors.New("db down")}).Create(context.Background(), "x", valid)
		if err == nil {
			t.Fatal("ingin error")
		}
	})

	t.Run("sub event prize invalid", func(t *testing.T) {
		bad := valid
		bad.SubEvents = []entities.SubEventCreate{{Title: "A", MaxSlots: 10, Status: "available", Winner1Prize: -1}}
		_, err := NewService(&fakeRepository{activeMember: true}).Create(context.Background(), "x", bad)
		if !errors.Is(err, apperrors.ErrInvalidPayload) {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("normalize default status", func(t *testing.T) {
		in := entities.EventCreate{
			OrganizerID: 1, Title: "Lomba", Slug: "lomba", Venue: "GOR", City: "Bandung",
			SubEvents: []entities.SubEventCreate{{Title: "A", MaxSlots: 10}},
		}
		repo := &fakeRepository{activeMember: true}
		_, err := NewService(repo).Create(context.Background(), "x", in)
		if err != nil || repo.createIn.Status != "draft" || repo.createIn.SubEvents[0].Status != "available" {
			t.Fatalf("err=%v in=%+v", err, repo.createIn)
		}
	})
}
