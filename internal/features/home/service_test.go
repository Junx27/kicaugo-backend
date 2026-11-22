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
	result          event.NearbyResult
	err             error
	count           int
	countErr        error
	activeStreaming *entities.ActiveEventStreaming
	activeErr       error
	featured        *entities.Event
	featuredErr     error
	identifier      string
	opts            Options
}

func (f *fakeEventReader) ListNearby(ctx context.Context, identifier string, radiusKM int, categoryID int64, limit int) (event.NearbyResult, error) {
	f.identifier = identifier
	f.opts = Options{RadiusKM: radiusKM, CategoryID: categoryID, Limit: limit}
	return f.result, f.err
}

func (f *fakeEventReader) CountOngoing(ctx context.Context) (int, error) {
	return f.count, f.countErr
}

func (f *fakeEventReader) GetActiveStreaming(ctx context.Context) (*entities.ActiveEventStreaming, error) {
	return f.activeStreaming, f.activeErr
}

func (f *fakeEventReader) GetRandomFeatured(ctx context.Context) (*entities.Event, error) {
	return f.featured, f.featuredErr
}

type fakeBirdCategoryReader struct {
	items []entities.BirdCategory
	err   error
}

func (f *fakeBirdCategoryReader) List(ctx context.Context) ([]entities.BirdCategory, error) {
	return f.items, f.err
}

func fixedClock() func() time.Time {
	return func() time.Time { return time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC) }
}

func TestServiceGet(t *testing.T) {
	t.Run("berhasil menggabungkan profil dan event", func(t *testing.T) {
		view := profile.ProfileView{Profile: entities.UserProfile{FullName: "Kicau Mania"}}
		nearby := event.NearbyResult{RadiusKM: 30}
		active := &entities.ActiveEventStreaming{ID: 1, EventID: 68, EventTitle: "Lomba", LinkStream: "https://x"}
		profiles := &fakeProfileReader{view: view}
		events := &fakeEventReader{result: nearby, count: 4, activeStreaming: active}
		svc := NewServiceWithClock(profiles, events, &fakeBirdCategoryReader{}, fixedClock())

		got, err := svc.Get(context.Background(), "user@example.com", Options{
			RadiusKM: 25, CategoryID: 3, Limit: 7,
		})
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.Profile.Profile.FullName != "Kicau Mania" || got.Nearby.RadiusKM != 30 ||
			got.Greeting != greetingAfternoon || got.OngoingEventsCount != 4 || got.ActiveStreaming == nil || got.ActiveStreaming.EventID != 68 {
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
		svc := NewService(&fakeProfileReader{err: errors.New("db down")}, events, &fakeBirdCategoryReader{})
		if _, err := svc.Get(context.Background(), "user@example.com", Options{}); err == nil {
			t.Error("Get() ingin error")
		}
		if events.identifier != "" {
			t.Error("event tidak boleh dipanggil saat profil gagal")
		}
	})

	t.Run("gagal membaca event", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{err: errors.New("db down")}, &fakeBirdCategoryReader{})
		if _, err := svc.Get(context.Background(), "user@example.com", Options{}); err == nil {
			t.Error("Get() ingin error")
		}
	})

	t.Run("gagal menghitung event berlangsung", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{countErr: errors.New("db down")}, &fakeBirdCategoryReader{})
		if _, err := svc.Get(context.Background(), "user@example.com", Options{}); err == nil {
			t.Error("Get() ingin error")
		}
	})

	t.Run("gagal membaca streaming aktif", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{activeErr: errors.New("db down")}, &fakeBirdCategoryReader{})
		if _, err := svc.Get(context.Background(), "user@example.com", Options{}); err == nil {
			t.Error("Get() ingin error")
		}
	})
}

func TestServiceDummySections(t *testing.T) {
	start := time.Date(2026, 12, 6, 0, 0, 0, 0, time.UTC)
	svc := NewService(&fakeProfileReader{}, &fakeEventReader{featured: &entities.Event{Title: "Piala Raja Nusantara 2026", Venue: "Batavia Arena", City: "Jaksel", StartDate: &start}}, &fakeBirdCategoryReader{})

	featured, err := svc.GetFeatured(context.Background())
	if err != nil || featured.Title != "Piala Raja Nusantara 2026" || featured.Distance != "-" || featured.Prize != "-" || len(featured.Attendees.Initials) != 0 {
		t.Fatalf("featured=%+v err=%v", featured, err)
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

func TestServiceGetFeatured(t *testing.T) {
	t.Run("kosong", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{}, &fakeBirdCategoryReader{})
		got, err := svc.GetFeatured(context.Background())
		if err != nil || got.Title != "Belum ada event featured" {
			t.Fatalf("got=%+v err=%v", got, err)
		}
	})
	t.Run("error", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{featuredErr: errors.New("db down")}, &fakeBirdCategoryReader{})
		if _, err := svc.GetFeatured(context.Background()); err == nil {
			t.Fatal("ingin error")
		}
	})
}

func TestServiceGetBirdCategories(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		want := []entities.BirdCategory{{ID: 1, Name: "Murai Batu", Slug: "murai_batu", Icon: "murai_batu"}}
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{}, &fakeBirdCategoryReader{items: want})
		got, err := svc.GetBirdCategories(context.Background())
		if err != nil {
			t.Fatalf("GetBirdCategories() error = %v", err)
		}
		if len(got) != 1 || got[0].Slug != "murai_batu" {
			t.Errorf("GetBirdCategories() = %+v", got)
		}
	})

	t.Run("gagal", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{}, &fakeBirdCategoryReader{err: errors.New("db down")})
		if _, err := svc.GetBirdCategories(context.Background()); err == nil {
			t.Error("GetBirdCategories() ingin error")
		}
	})
}
