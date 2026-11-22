package home

import (
	"context"
	"errors"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	"github.com/Junx27/kicaugo-backend/internal/features/event"
	"github.com/Junx27/kicaugo-backend/internal/features/profile"
)

type fakeProfileReader struct {
	view       profile.ProfileView
	err        error
	identifier string
}

func (f *fakeProfileReader) Get(ctx context.Context, identifier string) (profile.ProfileView, error) {
	f.identifier = identifier
	return f.view, f.err
}

type fakeEventReader struct {
	result     event.NearbyResult
	err        error
	identifier string
	opts       Options
}

func (f *fakeEventReader) ListNearby(ctx context.Context, identifier string, radiusKM int, categoryID int64, limit int) (event.NearbyResult, error) {
	f.identifier = identifier
	f.opts = Options{RadiusKM: radiusKM, CategoryID: categoryID, Limit: limit}
	return f.result, f.err
}

func TestServiceGet(t *testing.T) {
	t.Run("berhasil menggabungkan profil dan event", func(t *testing.T) {
		view := profile.ProfileView{Profile: entities.UserProfile{FullName: "Kicau Mania"}}
		nearby := event.NearbyResult{RadiusKM: 30}
		profiles := &fakeProfileReader{view: view}
		events := &fakeEventReader{result: nearby}

		got, err := NewService(profiles, events).Get(context.Background(), "user@example.com", Options{
			RadiusKM: 25, CategoryID: 3, Limit: 7,
		})
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.Profile.Profile.FullName != "Kicau Mania" || got.Nearby.RadiusKM != 30 {
			t.Errorf("Get() = %+v", got)
		}
		if profiles.identifier != "user@example.com" || events.identifier != "user@example.com" {
			t.Errorf("identifier: profil = %q, event = %q", profiles.identifier, events.identifier)
		}
		if events.opts != (Options{RadiusKM: 25, CategoryID: 3, Limit: 7}) {
			t.Errorf("opsi event = %+v", events.opts)
		}
	})

	t.Run("gagal membaca profil", func(t *testing.T) {
		events := &fakeEventReader{}
		svc := NewService(&fakeProfileReader{err: errors.New("db down")}, events)
		if _, err := svc.Get(context.Background(), "user@example.com", Options{}); err == nil {
			t.Error("Get() ingin error")
		}
		if events.identifier != "" {
			t.Error("event tidak boleh dipanggil saat profil gagal")
		}
	})

	t.Run("gagal membaca event", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{err: errors.New("db down")})
		if _, err := svc.Get(context.Background(), "user@example.com", Options{}); err == nil {
			t.Error("Get() ingin error")
		}
	})
}
