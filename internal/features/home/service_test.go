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
	gantanganCount  int
	gantanganErr    error
	activeStreaming *entities.ActiveEventStreaming
	activeErr       error
	featured        *entities.FeaturedEvent
	featuredErr     error
	upcoming        *entities.UpcomingEvent
	upcomingErr     error
	nearbySchedules []entities.EventSchedule
	nearbySchedErr  error
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

func (f *fakeEventReader) CountGantangan(ctx context.Context) (int, error) {
	if f.gantanganErr != nil {
		return 0, f.gantanganErr
	}
	return f.gantanganCount, nil
}

func (f *fakeEventReader) GetActiveStreaming(ctx context.Context) (*entities.ActiveEventStreaming, error) {
	return f.activeStreaming, f.activeErr
}

func (f *fakeEventReader) GetRandomFeatured(ctx context.Context, identifier string) (*entities.FeaturedEvent, error) {
	return f.featured, f.featuredErr
}

func (f *fakeEventReader) GetRandomUpcoming(ctx context.Context) (*entities.UpcomingEvent, error) {
	return f.upcoming, f.upcomingErr
}

func (f *fakeEventReader) ListNearbySchedules(ctx context.Context, identifier string, radiusKM int, categoryID int64, limit int) ([]entities.EventSchedule, error) {
	f.identifier = identifier
	f.opts = Options{RadiusKM: radiusKM, CategoryID: categoryID, Limit: limit}
	return f.nearbySchedules, f.nearbySchedErr
}

type fakeBirdCategoryReader struct {
	items []entities.BirdCategory
	err   error
}

func (f *fakeBirdCategoryReader) List(ctx context.Context) ([]entities.BirdCategory, error) {
	return f.items, f.err
}

type fakeOrganizerReader struct {
	active     bool
	err        error
	identifier string
	calls      int
}

func (f *fakeOrganizerReader) HasActiveMembership(ctx context.Context, identifier string) (bool, error) {
	f.identifier = identifier
	f.calls++
	return f.active, f.err
}

func fixedClock() func() time.Time {
	return func() time.Time { return time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC) }
}

func TestServiceGet(t *testing.T) {
	t.Run("berhasil menggabungkan profil, status organizer, dan event", func(t *testing.T) {
		view := profile.ProfileView{Profile: entities.UserProfile{FullName: "Kicau Mania"}}
		nearby := event.NearbyResult{RadiusKM: 30}
		active := &entities.ActiveEventStreaming{ID: 1, EventID: 68, EventTitle: "Lomba", LinkStream: "https://x"}
		profiles := &fakeProfileReader{view: view}
		events := &fakeEventReader{result: nearby, count: 4, activeStreaming: active}
		organizers := &fakeOrganizerReader{active: true}
		svc := NewServiceWithClock(profiles, events, &fakeBirdCategoryReader{}, organizers, fixedClock())

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
		if !got.IsOrganizer {
			t.Error("IsOrganizer = false, ingin true")
		}
		if profiles.identifier != "user@example.com" || events.identifier != "user@example.com" || organizers.identifier != "user@example.com" {
			t.Errorf("identifier: profil = %q, event = %q, organizer = %q", profiles.identifier, events.identifier, organizers.identifier)
		}
		if events.opts != (Options{RadiusKM: 25, CategoryID: 3, Limit: 7}) {
			t.Errorf("opsi event = %+v", events.opts)
		}
	})

	t.Run("bukan organizer", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{}, &fakeBirdCategoryReader{}, &fakeOrganizerReader{})
		got, err := svc.Get(context.Background(), "user@example.com", Options{})
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.IsOrganizer {
			t.Error("IsOrganizer = true, ingin false")
		}
	})

	t.Run("gagal membaca profil", func(t *testing.T) {
		events := &fakeEventReader{}
		organizers := &fakeOrganizerReader{}
		svc := NewService(&fakeProfileReader{err: errors.New("db down")}, events, &fakeBirdCategoryReader{}, organizers)
		if _, err := svc.Get(context.Background(), "user@example.com", Options{}); err == nil {
			t.Error("Get() ingin error")
		}
		if events.identifier != "" || organizers.calls != 0 {
			t.Error("event dan organizer tidak boleh dipanggil saat profil gagal")
		}
	})

	t.Run("gagal membaca status organizer", func(t *testing.T) {
		events := &fakeEventReader{}
		svc := NewService(&fakeProfileReader{}, events, &fakeBirdCategoryReader{}, &fakeOrganizerReader{err: errors.New("db down")})
		if _, err := svc.Get(context.Background(), "user@example.com", Options{}); err == nil {
			t.Error("Get() ingin error")
		}
		if events.identifier != "" {
			t.Error("event tidak boleh dipanggil saat cek organizer gagal")
		}
	})

	t.Run("gagal membaca event", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{err: errors.New("db down")}, &fakeBirdCategoryReader{}, &fakeOrganizerReader{})
		if _, err := svc.Get(context.Background(), "user@example.com", Options{}); err == nil {
			t.Error("Get() ingin error")
		}
	})

	t.Run("gagal menghitung event berlangsung", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{countErr: errors.New("db down")}, &fakeBirdCategoryReader{}, &fakeOrganizerReader{})
		if _, err := svc.Get(context.Background(), "user@example.com", Options{}); err == nil {
			t.Error("Get() ingin error")
		}
	})

	t.Run("gagal membaca streaming aktif", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{activeErr: errors.New("db down")}, &fakeBirdCategoryReader{}, &fakeOrganizerReader{})
		if _, err := svc.Get(context.Background(), "user@example.com", Options{}); err == nil {
			t.Error("Get() ingin error")
		}
	})
}

func TestServiceDummySections(t *testing.T) {
	start := time.Date(2026, 12, 6, 0, 0, 0, 0, time.UTC)
	prizePool := "2000000"
	svc := NewService(&fakeProfileReader{}, &fakeEventReader{featured: &entities.FeaturedEvent{Title: "Piala Raja Nusantara 2026", StartDate: &start, Venue: "Batavia Arena", City: "Jaksel", PrizePool: &prizePool, TotalSlots: 24}}, &fakeBirdCategoryReader{}, &fakeOrganizerReader{})

	featured, err := svc.GetFeatured(context.Background(), "user@example.com")
	if err != nil || featured.Title != "Piala Raja Nusantara 2026" || featured.Distance != "-" || featured.Prize != "Rp 2.000.000" || featured.Slots != "24 Slot" {
		t.Fatalf("featured=%+v err=%v", featured, err)
	}

	upcoming, err := svc.GetUpcoming(context.Background())
	if err != nil || upcoming.Title != "Belum ada event upcoming" {
		t.Errorf("upcoming = %+v err=%v", upcoming, err)
	}

	stats := svc.GetStats()
	if stats.Badge != "100% Resmi" || stats.Headline == "" {
		t.Errorf("stats = %+v", stats)
	}

	ticket := svc.GetActiveTicket()
	if ticket.Number != "#G-18" || ticket.Session == "" {
		t.Errorf("active ticket = %+v", ticket)
	}

	categoryID := int64(1)
	eventSchedules := []entities.EventSchedule{{
		ID: 1, EventID: 2, Title: "Latpres Exclusive Cucak Ijo", Venue: "Arena Master Cilandak, Jaksel",
		Price: 100000, Status: "available", CategoryID: &categoryID, CategoryName: "Cucak Hijau",
	}}
	svc = NewService(&fakeProfileReader{}, &fakeEventReader{nearbySchedules: eventSchedules}, &fakeBirdCategoryReader{}, &fakeOrganizerReader{})
	schedules, err := svc.GetSchedules(context.Background(), "user@example.com")
	if err != nil {
		t.Fatalf("GetSchedules() error = %v", err)
	}
	if len(schedules) != 1 {
		t.Fatalf("len(schedules) = %d, ingin 1", len(schedules))
	}
	if schedules[0].Status != "available" || schedules[0].Category.Name != "Cucak Hijau" {
		t.Errorf("schedule[0] = %+v", schedules[0])
	}
}

func TestServiceGetFeatured(t *testing.T) {
	t.Run("kosong", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{}, &fakeBirdCategoryReader{}, &fakeOrganizerReader{})
		got, err := svc.GetFeatured(context.Background(), "user@example.com")
		if err != nil || got.Title != "Belum ada event featured" {
			t.Fatalf("got=%+v err=%v", got, err)
		}
	})
	t.Run("error", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{featuredErr: errors.New("db down")}, &fakeBirdCategoryReader{}, &fakeOrganizerReader{})
		if _, err := svc.GetFeatured(context.Background(), "user@example.com"); err == nil {
			t.Fatal("ingin error")
		}
	})
	t.Run("berhasil", func(t *testing.T) {
		prizePool := "2000000"
		start := time.Date(2026, 12, 6, 0, 0, 0, 0, time.UTC)
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{featured: &entities.FeaturedEvent{Title: "Piala Raja Nusantara 2026", StartDate: &start, Venue: "Batavia Arena", City: "Jaksel", PrizePool: &prizePool, TotalSlots: 24}}, &fakeBirdCategoryReader{}, &fakeOrganizerReader{})
		got, err := svc.GetFeatured(context.Background(), "user@example.com")
		if err != nil || got.Title != "Piala Raja Nusantara 2026" || got.Prize != "Rp 2.000.000" || got.Slots != "24 Slot" {
			t.Fatalf("got=%+v err=%v", got, err)
		}
	})
	t.Run("jarak dan lokasi", func(t *testing.T) {
		near := 0.05
		far := 3.14
		cases := []struct {
			name     string
			event    entities.FeaturedEvent
			distance string
			schedule string
		}{
			{name: "jarak sangat dekat", event: entities.FeaturedEvent{Title: "Dekat", DistanceKM: &near}, distance: "< 0.1 km", schedule: "-"},
			{name: "jarak normal", event: entities.FeaturedEvent{Title: "Jauh", DistanceKM: &far}, distance: "3.1 km", schedule: "-"},
			{name: "lokasi tanpa tanggal", event: entities.FeaturedEvent{Title: "Lokasi", Venue: "Arena", City: "Jaksel"}, distance: "-", schedule: "Arena, Jaksel"},
			{name: "lokasi kosong", event: entities.FeaturedEvent{Title: "Kosong", Venue: "  ", City: ""}, distance: "-", schedule: "-"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				featuredEvent := tc.event
				svc := NewService(&fakeProfileReader{}, &fakeEventReader{featured: &featuredEvent}, &fakeBirdCategoryReader{}, &fakeOrganizerReader{})
				got, err := svc.GetFeatured(context.Background(), "user@example.com")
				if err != nil || got.Distance != tc.distance || got.Schedule != tc.schedule {
					t.Fatalf("got=%+v err=%v", got, err)
				}
			})
		}
	})
}

func TestServiceGetUpcoming(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		start := time.Date(2026, 12, 10, 0, 0, 0, 0, time.UTC)
		reg := time.Date(2026, 12, 1, 8, 30, 0, 0, time.UTC)
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{upcoming: &entities.UpcomingEvent{Title: "Lomba Draft", StartDate: &start, RegistrationOpen: &reg, MinPrice: 50000, RemainingTickets: 24}}, &fakeBirdCategoryReader{}, &fakeOrganizerReader{})
		got, err := svc.GetUpcoming(context.Background())
		if err != nil || got.Title != "Lomba Draft" || got.Type != "Upcoming Event" || got.Price != "Rp 50.000" || got.RemainingTickets != 24 || got.Time != "01 Dec 2026 08:30" {
			t.Fatalf("got=%+v err=%v", got, err)
		}
	})
	t.Run("gratis tanpa registrasi", func(t *testing.T) {
		start := time.Date(2026, 12, 10, 0, 0, 0, 0, time.UTC)
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{upcoming: &entities.UpcomingEvent{Title: "Lomba Draft", StartDate: &start, MinPrice: 0, RemainingTickets: 10}}, &fakeBirdCategoryReader{}, &fakeOrganizerReader{})
		got, err := svc.GetUpcoming(context.Background())
		if err != nil || got.Price != "Gratis" || got.Time != "10 Dec 2026" {
			t.Fatalf("got=%+v err=%v", got, err)
		}
	})
	t.Run("kosong", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{}, &fakeBirdCategoryReader{}, &fakeOrganizerReader{})
		got, err := svc.GetUpcoming(context.Background())
		if err != nil || got.Title != "Belum ada event upcoming" {
			t.Fatalf("got=%+v err=%v", got, err)
		}
	})
	t.Run("error", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{upcomingErr: errors.New("db down")}, &fakeBirdCategoryReader{}, &fakeOrganizerReader{})
		if _, err := svc.GetUpcoming(context.Background()); err == nil {
			t.Fatal("ingin error")
		}
	})
}

func TestServiceGetBirdCategories(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		want := []entities.BirdCategory{{ID: 1, Name: "Murai Batu", Slug: "murai_batu", Icon: "murai_batu"}}
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{}, &fakeBirdCategoryReader{items: want}, &fakeOrganizerReader{})
		got, err := svc.GetBirdCategories(context.Background())
		if err != nil {
			t.Fatalf("GetBirdCategories() error = %v", err)
		}
		if len(got) != 1 || got[0].Slug != "murai_batu" {
			t.Errorf("GetBirdCategories() = %+v", got)
		}
	})

	t.Run("gagal", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{}, &fakeBirdCategoryReader{err: errors.New("db down")}, &fakeOrganizerReader{})
		if _, err := svc.GetBirdCategories(context.Background()); err == nil {
			t.Error("GetBirdCategories() ingin error")
		}
	})
}
