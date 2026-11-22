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
	view       View
	err        error
	identifier string
	opts       Options
}

func (f *fakeService) Get(ctx context.Context, identifier string, opts Options) (View, error) {
	f.identifier = identifier
	f.opts = opts
	return f.view, f.err
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
	}
}

func TestGetHomeProfileHandler(t *testing.T) {
	t.Run("berhasil dengan ringkasan profil dan event", func(t *testing.T) {
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
				Profile struct {
					Greeting           string  `json:"greeting"`
					FullName           string  `json:"full_name"`
					AvatarURL          *string `json:"avatar_url"`
					SearchRadiusKM     int     `json:"search_radius_km"`
					FavoriteCategories []struct {
						ID   int64  `json:"id"`
						Name string `json:"name"`
					} `json:"favorite_categories"`
					IsComplete    bool     `json:"is_complete"`
					MissingFields []string `json:"missing_fields"`
					City          *struct {
						Name     string `json:"name"`
						Province struct {
							Name string `json:"name"`
						} `json:"province"`
					} `json:"city"`
				} `json:"profile"`
				Nearby struct {
					OriginSource       string `json:"origin_source"`
					RadiusKM           int    `json:"radius_km"`
					OngoingEventsCount int    `json:"ongoing_events_count"`
					Events             []struct {
						ID         int64   `json:"id"`
						DistanceKM float64 `json:"distance_km"`
						StartDate  *string `json:"start_date"`
					} `json:"events"`
				} `json:"nearby"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("body bukan JSON: %v", err)
		}
		if body.Status != http.StatusOK || body.Message != "Home profile loaded" {
			t.Errorf("status = %d, message = %q", body.Status, body.Message)
		}
		p := body.Data.Profile
		if p.Greeting != greetingNoon {
			t.Errorf("greeting = %q, ingin %q", p.Greeting, greetingNoon)
		}
		if p.FullName != "Kicau Mania" || p.AvatarURL == nil || *p.AvatarURL != "https://cdn.example.com/avatar.png" ||
			p.SearchRadiusKM != 30 || p.IsComplete || len(p.MissingFields) != 1 || p.MissingFields[0] != "city_id" {
			t.Errorf("profile = %+v", p)
		}
		if p.City == nil || p.City.Name != "Banda Aceh" || p.City.Province.Name != "Aceh" ||
			len(p.FavoriteCategories) != 1 || p.FavoriteCategories[0].Name != "Murai" {
			t.Errorf("city/favorites = %+v", p)
		}
		n := body.Data.Nearby
		if n.OngoingEventsCount != 3 {
			t.Errorf("ongoing_events_count = %d, ingin 3", n.OngoingEventsCount)
		}
		if n.OriginSource != event.OriginCurrentLocation || n.RadiusKM != 30 || len(n.Events) != 1 {
			t.Fatalf("nearby = %+v", n)
		}
		if n.Events[0].ID != 7 || n.Events[0].DistanceKM != 3.1 ||
			n.Events[0].StartDate == nil || *n.Events[0].StartDate != "2026-12-01" {
			t.Errorf("event = %+v", n.Events[0])
		}
		if svc.identifier != testutil.TestIdentifier || svc.opts != (Options{RadiusKM: 25, CategoryID: 2, Limit: 3}) {
			t.Errorf("identifier = %q, opts = %+v", svc.identifier, svc.opts)
		}

		var raw map[string]any
		if err := json.Unmarshal(resp.Body.Bytes(), &raw); err != nil {
			t.Fatalf("body bukan JSON: %v", err)
		}
		profileData := raw["data"].(map[string]any)["profile"].(map[string]any)
		if _, ok := profileData["phone"]; ok {
			t.Error("phone tidak boleh diekspos di home")
		}
		if _, ok := profileData["current_location"]; ok {
			t.Error("current_location tidak boleh diekspos di home")
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
				Nearby struct {
					OriginSource string  `json:"origin_source"`
					Hint         *string `json:"hint"`
					Events       []any   `json:"events"`
				} `json:"nearby"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("body bukan JSON: %v", err)
		}
		if body.Data.Nearby.OriginSource != "none" || body.Data.Nearby.Hint == nil ||
			*body.Data.Nearby.Hint != event.HintLocationNotSet || len(body.Data.Nearby.Events) != 0 {
			t.Errorf("nearby = %+v", body.Data.Nearby)
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
