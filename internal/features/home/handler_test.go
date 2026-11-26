package home

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2/humatest"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	"github.com/Junx27/kicaugo-backend/internal/features/event"
	"github.com/Junx27/kicaugo-backend/internal/features/profile"
	"github.com/Junx27/kicaugo-backend/internal/testutil"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
	"github.com/Junx27/kicaugo-backend/internal/utils/request"
)

type fakeService struct {
	view           View
	err            error
	identifier     string
	opts           Options
	featured       Featured
	upcoming       Upcoming
	stats          Stats
	activeTicket   ActiveTicket
	schedules      []Schedule
	scheduleErr    error
	birdCategories []entities.BirdCategory
	birdErr        error
}

func (f *fakeService) Get(ctx context.Context, identifier string, opts Options) (View, error) {
	f.identifier = identifier
	f.opts = opts
	return f.view, f.err
}

func (f *fakeService) GetFeatured(ctx context.Context, identifier string) (Featured, error) {
	return f.featured, f.err
}
func (f *fakeService) GetUpcoming(ctx context.Context) (Upcoming, error) { return f.upcoming, f.err }
func (f *fakeService) GetStats() Stats                                   { return f.stats }
func (f *fakeService) GetActiveTicket() ActiveTicket                     { return f.activeTicket }
func (f *fakeService) GetSchedules(ctx context.Context, identifier string) ([]Schedule, error) {
	f.identifier = identifier
	if f.scheduleErr != nil {
		return f.schedules, f.scheduleErr
	}
	return f.schedules, f.err
}
func (f *fakeService) GetBirdCategories(ctx context.Context) ([]entities.BirdCategory, error) {
	return f.birdCategories, f.birdErr
}

func setupHandler(t *testing.T, svc Service) (humatest.TestAPI, *jwt.Manager) {
	t.Helper()
	_, api := humatest.New(t)
	tokens := testutil.NewTokenManager()
	RegisterRoutes(api, svc, tokens)
	return api, tokens
}

func sampleView() View {
	phone := "+6281234567890"
	avatar := "https://cdn.example.com/avatar.png"
	latitude, longitude := -6.2, 106.8
	start := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	streamStart := time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)
	streamEnd := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	return View{
		Profile: profile.ProfileView{
			Profile: entities.UserProfile{
				FullName:         "Kicau Mania",
				Phone:            &phone,
				AvatarURL:        &avatar,
				CurrentLatitude:  &latitude,
				CurrentLongitude: &longitude,
				SearchRadiusKM:   30,
				City: &entities.City{
					ID: 19, Name: "Banda Aceh", Type: "kota", Slug: "kota-banda-aceh", ProvinceID: 11,
				},
				Province: &entities.Province{ID: 11, Code: "11", Name: "Aceh"},
			},
			Favorites:     []entities.BirdCategory{{ID: 1, Name: "Murai"}},
			MissingFields: []string{"city_id"},
		},
		Nearby: event.NearbyResult{
			Origin:   &entities.NearbyOrigin{Latitude: latitude, Longitude: longitude, Source: event.OriginCurrentLocation},
			RadiusKM: 30,
			Events: []entities.Event{{
				ID: 7, Title: "Kontes Murai", Slug: "kontes-murai", Venue: "GOR Kota",
				City: "Banda Aceh", Province: "Aceh", StartDate: &start,
				Status: "published", DistanceKM: 3.14159,
			}},
		},
		Greeting:           greetingNoon,
		OngoingEventsCount: 3,
		IsOrganizer:        true,
		ActiveStreaming: &entities.ActiveEventStreaming{
			ID: 9, EventID: 68, EventTitle: "Lomba Kicau", EventSlug: "lomba-kicau", Venue: "GOR", City: "Sleman",
			LinkStream: "https://youtube.com/live/abc", StartTime: streamStart, EndTime: &streamEnd, TotalViews: 12, Juara: []byte(`{"rank1":"A"}`),
		},
	}
}

func TestGetHomeProfileHandler(t *testing.T) {
	t.Run("berhasil dengan ringkasan sederhana", func(t *testing.T) {
		svc := &fakeService{view: sampleView()}
		api, tokens := setupHandler(t, svc)

		resp := api.Get("/api/v1/home/profile?radius_km=25&category_id=2&limit=3", testutil.BearerHeader(t, tokens))
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, ingin 200; body = %s", resp.Code, resp.Body.String())
		}

		var body struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
			Data    struct {
				Name            string  `json:"name"`
				Avatar          *string `json:"avatar"`
				Message         string  `json:"message"`
				LiveCount       int     `json:"live_count"`
				IsOrganizer     bool    `json:"is_organizer"`
				ActiveStreaming *struct {
					ID         int64   `json:"id"`
					EventID    int64   `json:"event_id"`
					EventTitle string  `json:"event_title"`
					LinkStream string  `json:"link_stream"`
					EndTime    *string `json:"end_time"`
				} `json:"active_streaming"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("body bukan JSON: %v", err)
		}
		if body.Status != http.StatusOK || body.Message != "Home profile loaded" {
			t.Errorf("status = %d, message = %q", body.Status, body.Message)
		}
		if body.Data.Name != "Kicau" || body.Data.Avatar == nil || *body.Data.Avatar != "https://cdn.example.com/avatar.png" {
			t.Errorf("data = %+v", body.Data)
		}
		if body.Data.Message != greetingNoon || body.Data.LiveCount != 3 {
			t.Errorf("message/live_count = %q/%d", body.Data.Message, body.Data.LiveCount)
		}
		if !body.Data.IsOrganizer {
			t.Error("is_organizer = false, ingin true")
		}
		if body.Data.ActiveStreaming == nil || body.Data.ActiveStreaming.EventID != 68 || body.Data.ActiveStreaming.LinkStream == "" {
			t.Fatalf("active_streaming=%+v", body.Data.ActiveStreaming)
		}
		if body.Data.ActiveStreaming.EndTime == nil || *body.Data.ActiveStreaming.EndTime != time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC).Format(time.RFC3339) {
			t.Errorf("end_time = %v", body.Data.ActiveStreaming.EndTime)
		}
		if svc.identifier != testutil.TestIdentifier || svc.opts != (Options{RadiusKM: 25, CategoryID: 2, Limit: 3}) {
			t.Errorf("identifier = %q, opts = %+v", svc.identifier, svc.opts)
		}

		var raw map[string]any
		if err := json.Unmarshal(resp.Body.Bytes(), &raw); err != nil {
			t.Fatalf("body bukan JSON: %v", err)
		}
		data := raw["data"].(map[string]any)
		for _, key := range []string{"profile", "nearby", "phone", "current_location"} {
			if _, ok := data[key]; ok {
				t.Errorf("field %q tidak boleh ada di response sederhana", key)
			}
		}
	})

	t.Run("memakai default radius profil dan limit", func(t *testing.T) {
		hint := event.HintLocationNotSet
		svc := &fakeService{view: View{Nearby: event.NearbyResult{RadiusKM: 50, Hint: &hint}}}
		api, tokens := setupHandler(t, svc)

		resp := api.Get("/api/v1/home/profile", testutil.BearerHeader(t, tokens))
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, ingin 200", resp.Code)
		}
		var body struct {
			Data struct {
				Name      string `json:"name"`
				Message   string `json:"message"`
				LiveCount int    `json:"live_count"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("body bukan JSON: %v", err)
		}
		if body.Data.Name != "" || body.Data.LiveCount != 0 {
			t.Errorf("data = %+v", body.Data)
		}
		var raw map[string]any
		if err := json.Unmarshal(resp.Body.Bytes(), &raw); err != nil {
			t.Fatalf("json: %v", err)
		}
		data := raw["data"].(map[string]any)
		if _, ok := data["active_streaming"]; !ok {
			t.Error("active_streaming harus ada (null)")
		}
		if v, ok := data["is_organizer"]; !ok || v != false {
			t.Errorf("is_organizer = %v (ada=%v), ingin false", v, ok)
		}
		if svc.opts != (Options{Limit: 5}) {
			t.Errorf("opts = %+v, ingin limit 5", svc.opts)
		}
	})

	t.Run("ditolak tanpa token", func(t *testing.T) {
		api, _ := setupHandler(t, &fakeService{})
		resp := api.Get("/api/v1/home/profile")
		if resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})

	t.Run("radius melebihi batas ditolak", func(t *testing.T) {
		api, tokens := setupHandler(t, &fakeService{})
		resp := api.Get("/api/v1/home/profile?radius_km=501", testutil.BearerHeader(t, tokens))
		if resp.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, ingin 422", resp.Code)
		}
	})
}

func TestMapServiceError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{name: "radius tidak valid", err: apperrors.ErrInvalidRadius, want: http.StatusUnprocessableEntity},
		{name: "user tidak ditemukan", err: apperrors.ErrNotFound, want: http.StatusNotFound},
		{name: "kesalahan server", err: errors.New("db down"), want: http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			api, tokens := setupHandler(t, &fakeService{err: tc.err})
			resp := api.Get("/api/v1/home/profile", testutil.BearerHeader(t, tokens))
			if resp.Code != tc.want {
				t.Errorf("status = %d, ingin %d; body = %s", resp.Code, tc.want, resp.Body.String())
			}
		})
	}
}

func TestGetProfileHandlerTanpaIdentitas(t *testing.T) {
	_, err := newGetProfileHandler(&fakeService{})(context.Background(), &request.HomeProfileInput{})
	statusErr, ok := err.(interface{ GetStatus() int })
	if err == nil || !ok || statusErr.GetStatus() != http.StatusUnauthorized {
		t.Errorf("handler tanpa identitas = %v, ingin 401", err)
	}
}

func sampleFeatured() Featured {
	return Featured{
		Title:     "Piala Raja Nusantara 2026",
		Schedule:  "Minggu, 6 Des • Batavia Arena, Jaksel",
		Distance:  "3.2 KM",
		Prize:     "Rp 50 Jt",
		Slots:     "12 Gantangan",
		Attendees: FeaturedAttendees{Initials: []string{"MK", "BN"}, Others: 48},
	}
}

func sampleUpcoming() Upcoming {
	return Upcoming{Type: "Latber Presisi", Title: "Latber Presisi Kacer", Time: "Besok, 14:00 WIB", Price: "Rp 75.000", RemainingTickets: 3}
}

func sampleStats() Stats {
	return Stats{Badge: "100% Resmi", Headline: "520+ Gantangan", Note: "Terverifikasi panitia & EO resmi.", SecurityNote: "98% Transaksi Tiket Instan & Aman"}
}

func sampleActiveTicket() ActiveTicket {
	return ActiveTicket{Event: "Grand Prix Murai Batu VIP #18", Schedule: "Sabtu, 12 Des • Masterpiece Arena", Number: "#G-18", Session: "Sesi A - 16G"}
}

func sampleSchedules() []Schedule {
	return []Schedule{{Title: "Latpres Exclusive Cucak Ijo", Venue: "Arena Master Cilandak, Jaksel", Date: "Selasa, 24 Nov", Price: "Rp 100rb", Status: "available", Category: ScheduleCategory{ID: 1, Name: "Cucak Hijau"}}}
}

func ptrString(v string) *string { return &v }

func TestGetHomeFeaturedHandler(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		svc := &fakeService{featured: sampleFeatured()}
		api, tokens := setupHandler(t, svc)
		resp := api.Get("/api/v1/home/featured", testutil.BearerHeader(t, tokens))
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
		}
		var body struct {
			Message string `json:"message"`
			Data    struct {
				Title     string `json:"title"`
				Distance  string `json:"distance"`
				Prize     string `json:"prize"`
				Attendees struct {
					Initials []string `json:"initials"`
					Others   int      `json:"others"`
				} `json:"attendees"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("json: %v", err)
		}
		if body.Message != "Home featured loaded" || body.Data.Title != "Piala Raja Nusantara 2026" || body.Data.Attendees.Others != 48 || len(body.Data.Attendees.Initials) != 2 {
			t.Errorf("body = %+v", body)
		}
	})
	t.Run("ditolak tanpa token", func(t *testing.T) {
		api, _ := setupHandler(t, &fakeService{})
		if resp := api.Get("/api/v1/home/featured"); resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})
	t.Run("error service", func(t *testing.T) {
		api, tokens := setupHandler(t, &fakeService{err: errors.New("db down")})
		if resp := api.Get("/api/v1/home/featured", testutil.BearerHeader(t, tokens)); resp.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
	})
}

func TestGetHomeUpcomingHandler(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		svc := &fakeService{upcoming: sampleUpcoming()}
		api, tokens := setupHandler(t, svc)
		resp := api.Get("/api/v1/home/upcoming", testutil.BearerHeader(t, tokens))
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
		}
		var body struct {
			Message string `json:"message"`
			Data    struct {
				RemainingTickets int    `json:"remaining_tickets"`
				Title            string `json:"title"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("json: %v", err)
		}
		if body.Message != "Home upcoming loaded" || body.Data.RemainingTickets != 3 || body.Data.Title != "Latber Presisi Kacer" {
			t.Errorf("body = %+v", body)
		}
	})
	t.Run("ditolak tanpa token", func(t *testing.T) {
		api, _ := setupHandler(t, &fakeService{})
		if resp := api.Get("/api/v1/home/upcoming"); resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})
	t.Run("error service", func(t *testing.T) {
		api, tokens := setupHandler(t, &fakeService{err: errors.New("db down")})
		if resp := api.Get("/api/v1/home/upcoming", testutil.BearerHeader(t, tokens)); resp.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
	})
}

func TestGetHomeStatsHandler(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		svc := &fakeService{stats: sampleStats()}
		api, tokens := setupHandler(t, svc)
		resp := api.Get("/api/v1/home/stats", testutil.BearerHeader(t, tokens))
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
		}
		var body struct {
			Message string `json:"message"`
			Data    struct {
				Badge        string `json:"badge"`
				SecurityNote string `json:"security_note"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("json: %v", err)
		}
		if body.Message != "Home stats loaded" || body.Data.Badge != "100% Resmi" || body.Data.SecurityNote == "" {
			t.Errorf("body = %+v", body)
		}
	})
	t.Run("ditolak tanpa token", func(t *testing.T) {
		api, _ := setupHandler(t, &fakeService{})
		if resp := api.Get("/api/v1/home/stats"); resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})
}

func TestGetHomeActiveTicketHandler(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		svc := &fakeService{activeTicket: sampleActiveTicket()}
		api, tokens := setupHandler(t, svc)
		resp := api.Get("/api/v1/home/active-ticket", testutil.BearerHeader(t, tokens))
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
		}
		var body struct {
			Message string `json:"message"`
			Data    struct {
				Number  string `json:"number"`
				Session string `json:"session"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("json: %v", err)
		}
		if body.Message != "Home active ticket loaded" || body.Data.Number != "#G-18" || body.Data.Session != "Sesi A - 16G" {
			t.Errorf("body = %+v", body)
		}
	})
	t.Run("ditolak tanpa token", func(t *testing.T) {
		api, _ := setupHandler(t, &fakeService{})
		if resp := api.Get("/api/v1/home/active-ticket"); resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})
}

func TestGetHomeSchedulesHandler(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		svc := &fakeService{schedules: sampleSchedules()}
		api, tokens := setupHandler(t, svc)
		resp := api.Get("/api/v1/home/schedules", testutil.BearerHeader(t, tokens))
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
		}
		var body struct {
			Message string `json:"message"`
			Data    []struct {
				Title    string `json:"title"`
				Status   string `json:"status"`
				Category struct {
					ID   int64  `json:"id"`
					Name string `json:"name"`
				} `json:"category"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("json: %v", err)
		}
		if body.Message != "Home schedules loaded" || len(body.Data) != 1 || body.Data[0].Status != "available" || body.Data[0].Category.Name != "Cucak Hijau" {
			t.Errorf("body = %+v", body)
		}
	})
	t.Run("ditolak tanpa token", func(t *testing.T) {
		api, _ := setupHandler(t, &fakeService{})
		if resp := api.Get("/api/v1/home/schedules"); resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})
}

func TestGetHomeBirdCategoriesHandler(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		svc := &fakeService{birdCategories: []entities.BirdCategory{{ID: 1, Name: "Murai Batu", Slug: "murai_batu", Icon: "murai_batu"}}}
		api, tokens := setupHandler(t, svc)
		resp := api.Get("/api/v1/home/bird-categories", testutil.BearerHeader(t, tokens))
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
		}
		var body struct {
			Message string `json:"message"`
			Data    []struct {
				ID   int64  `json:"id"`
				Name string `json:"name"`
				Slug string `json:"slug"`
				Icon string `json:"icon"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("json: %v", err)
		}
		if body.Message != "Home bird categories loaded" || len(body.Data) != 1 || body.Data[0].Slug != "murai_batu" || body.Data[0].Icon != "murai_batu" {
			t.Errorf("body = %+v", body)
		}
	})

	t.Run("service error", func(t *testing.T) {
		svc := &fakeService{birdErr: errors.New("db down")}
		api, tokens := setupHandler(t, svc)
		resp := api.Get("/api/v1/home/bird-categories", testutil.BearerHeader(t, tokens))
		if resp.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, ingin 500", resp.Code)
		}
	})

	t.Run("ditolak tanpa token", func(t *testing.T) {
		api, _ := setupHandler(t, &fakeService{})
		if resp := api.Get("/api/v1/home/bird-categories"); resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})
}

func TestDummyHandlersTanpaIdentitas(t *testing.T) {
	cases := []struct {
		name string
		fn   func() error
	}{
		{name: "featured", fn: func() error {
			_, err := newGetFeaturedHandler(&fakeService{})(context.Background(), &struct{}{})
			return err
		}},
		{name: "upcoming", fn: func() error {
			_, err := newGetUpcomingHandler(&fakeService{})(context.Background(), &struct{}{})
			return err
		}},
		{name: "stats", fn: func() error {
			_, err := newGetStatsHandler(&fakeService{})(context.Background(), &struct{}{})
			return err
		}},
		{name: "active ticket", fn: func() error {
			_, err := newGetActiveTicketHandler(&fakeService{})(context.Background(), &struct{}{})
			return err
		}},
		{name: "schedules", fn: func() error {
			_, err := newGetSchedulesHandler(&fakeService{})(context.Background(), &struct{}{})
			return err
		}},
		{name: "bird categories", fn: func() error {
			_, err := newGetBirdCategoriesHandler(&fakeService{})(context.Background(), &struct{}{})
			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn()
			statusErr, ok := err.(interface{ GetStatus() int })
			if err == nil || !ok || statusErr.GetStatus() != http.StatusUnauthorized {
				t.Errorf("err = %v, ingin 401", err)
			}
		})
	}
}
