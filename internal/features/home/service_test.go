package home

import (
	"context"
	"errors"
	"testing"
	"time"

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
	count      int
	countErr   error
	identifier string
	opts       Options
}

func (f *fakeEventReader) ListNearby(ctx context.Context, identifier string, radiusKM int, categoryID int64, limit int) (event.NearbyResult, error) {
	f.identifier = identifier
	f.opts = Options{RadiusKM: radiusKM, CategoryID: categoryID, Limit: limit}
	return f.result, f.err
}

func (f *fakeEventReader) CountOngoing(ctx context.Context) (int, error) {
	return f.count, f.countErr
}

func fixedClock() func() time.Time {
	return func() time.Time { return time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC) }
}

func TestServiceGet(t *testing.T) {
	t.Run("berhasil menggabungkan profil dan event", func(t *testing.T) {
		view := profile.ProfileView{Profile: entities.UserProfile{FullName: "Kicau Mania"}}
		nearby := event.NearbyResult{RadiusKM: 30}
		profiles := &fakeProfileReader{view: view}
		events := &fakeEventReader{result: nearby, count: 4}
		svc := NewServiceWithClock(profiles, events, fixedClock())

		got, err := svc.Get(context.Background(), "user@example.com", Options{
			RadiusKM: 25, CategoryID: 3, Limit: 7,
		})
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.Profile.Profile.FullName != "Kicau Mania" || got.Nearby.RadiusKM != 30 ||
			got.Greeting != greetingAfternoon || got.OngoingEventsCount != 4 {
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

	t.Run("gagal menghitung event berlangsung", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{countErr: errors.New("db down")})
		if _, err := svc.Get(context.Background(), "user@example.com", Options{}); err == nil {
			t.Error("Get() ingin error")
		}
	})
}

func TestServiceDummySections(t *testing.T) {
	svc := NewService(&fakeProfileReader{}, &fakeEventReader{})

	featured := svc.GetFeatured()
	if featured.Title == "" || len(featured.Attendees.Initials) != 2 || featured.Attendees.Others != 48 {
		t.Errorf("featured = %+v", featured)
	}

	upcoming := svc.GetUpcoming()
	if upcoming.Title == "" || upcoming.RemainingTickets != 3 {
		t.Errorf("upcoming = %+v", upcoming)
	}

	stats := svc.GetStats()
	if stats.Badge != "100% Resmi" || stats.Headline == "" {
		t.Errorf("stats = %+v", stats)
	}

	ticket := svc.GetActiveTicket()
	if ticket.Number != "#G-18" || ticket.Session == "" {
		t.Errorf("active ticket = %+v", ticket)
	}

	schedules := svc.GetSchedules()
	if len(schedules) != 4 {
		t.Fatalf("len(schedules) = %d, ingin 4", len(schedules))
	}
	if schedules[0].Status != "available" || schedules[0].Category.Name != "Cucak Hijau" {
		t.Errorf("schedule[0] = %+v", schedules[0])
	}
}
