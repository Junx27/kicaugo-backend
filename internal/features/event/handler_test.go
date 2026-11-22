package event

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2/humatest"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
	"github.com/Junx27/kicaugo-backend/internal/utils/response"
)

type fakeService struct {
	result NearbyResult
	err    error
	params struct {
		radius int
		cat    int64
		limit  int
	}
}

func (f *fakeService) ListNearby(ctx context.Context, identifier string, radiusKM int, categoryID int64, limit int) (NearbyResult, error) {
	f.params.radius = radiusKM
	f.params.cat = categoryID
	f.params.limit = limit
	return f.result, f.err
}

func newTestTokens() *jwt.Manager {
	return jwt.NewManager([]byte("rahasia-untuk-test"), time.Hour)
}

func bearerHeader(t *testing.T, tokens *jwt.Manager) string {
	t.Helper()
	token, err := tokens.Issue("user@example.com", "Tester")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	return "Authorization: Bearer " + token
}

func setupHandler(t *testing.T, svc Service) (humatest.TestAPI, *jwt.Manager) {
	t.Helper()
	_, api := humatest.New(t)
	tokens := newTestTokens()
	RegisterRoutes(api, svc, tokens)
	return api, tokens
}

func TestListNearbyHandler(t *testing.T) {
	t.Run("berhasil dengan asal lokasi terkini", func(t *testing.T) {
		start := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, 12, 2, 0, 0, 0, 0, time.UTC)
		svc := &fakeService{result: NearbyResult{
			Origin:   &entities.NearbyOrigin{Latitude: -6.2, Longitude: 106.8, Source: OriginCurrentLocation},
			RadiusKM: 25,
			Events: []entities.Event{{
				ID: 1, Title: "Kontes Murai", Slug: "kontes-murai", Venue: "GOR Kota",
				City: "Bandung", Province: "Jawa Barat", StartDate: &start, EndDate: &end,
				Status: "published", DistanceKM: 3.14159,
			}},
		}}
		api, tokens := setupHandler(t, svc)

		resp := api.Get("/api/v1/events/nearby?radius_km=25&category_id=2&limit=5", bearerHeader(t, tokens))
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, ingin 200; body = %s", resp.Code, resp.Body.String())
		}
		var body struct {
			Status int `json:"status"`
			Data   struct {
				OriginSource   string   `json:"origin_source"`
				OriginLatitude *float64 `json:"origin_latitude"`
				RadiusKM       int      `json:"radius_km"`
				Hint           *string  `json:"hint"`
				Events         []struct {
					ID         int64   `json:"id"`
					Title      string  `json:"title"`
					StartDate  *string `json:"start_date"`
					EndDate    *string `json:"end_date"`
					DistanceKM float64 `json:"distance_km"`
					Latitude   any     `json:"latitude"`
					Longitude  any     `json:"longitude"`
				} `json:"events"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("body bukan JSON: %v", err)
		}
		if body.Status != http.StatusOK || body.Data.OriginSource != OriginCurrentLocation ||
			body.Data.OriginLatitude == nil || *body.Data.OriginLatitude != -6.2 || body.Data.RadiusKM != 25 ||
			body.Data.Hint != nil || len(body.Data.Events) != 1 {
			t.Fatalf("body = %+v", body)
		}
		got := body.Data.Events[0]
		if got.ID != 1 || got.DistanceKM != 3.1 || got.StartDate == nil || *got.StartDate != "2026-12-01" ||
			got.EndDate == nil || *got.EndDate != "2026-12-02" || got.Latitude != nil || got.Longitude != nil {
			t.Errorf("event = %+v", got)
		}
		if svc.params.radius != 25 || svc.params.cat != 2 || svc.params.limit != 5 {
			t.Errorf("params = %+v", svc.params)
		}
	})

	t.Run("tanpa asal mengembalikan hint", func(t *testing.T) {
		hint := HintLocationNotSet
		svc := &fakeService{result: NearbyResult{RadiusKM: 50, Events: []entities.Event{}, Hint: &hint}}
		api, tokens := setupHandler(t, svc)
		resp := api.Get("/api/v1/events/nearby", bearerHeader(t, tokens))
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, ingin 200", resp.Code)
		}
		var body struct {
			Data response.NearbyEventsData `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("body bukan JSON: %v", err)
		}
		if body.Data.OriginSource != originNone || body.Data.OriginLatitude != nil ||
			body.Data.Hint == nil || *body.Data.Hint != HintLocationNotSet || len(body.Data.Events) != 0 {
			t.Errorf("data = %+v", body.Data)
		}
		if svc.params.radius != 0 || svc.params.limit != defaultLimit {
			t.Errorf("params = %+v", svc.params)
		}
	})

	t.Run("ditolak tanpa token", func(t *testing.T) {
		api, _ := setupHandler(t, &fakeService{})
		resp := api.Get("/api/v1/events/nearby")
		if resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})

	t.Run("radius melebihi batas ditolak", func(t *testing.T) {
		api, tokens := setupHandler(t, &fakeService{})
		resp := api.Get("/api/v1/events/nearby?radius_km=501", bearerHeader(t, tokens))
		if resp.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, ingin 422", resp.Code)
		}
	})

	t.Run("gagal", func(t *testing.T) {
		api, tokens := setupHandler(t, &fakeService{err: errors.New("db down")})
		resp := api.Get("/api/v1/events/nearby", bearerHeader(t, tokens))
		if resp.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, ingin 500", resp.Code)
		}
	})
}

func TestMapServiceError(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{err: apperrors.ErrInvalidRadius, want: http.StatusUnprocessableEntity},
		{err: apperrors.ErrInvalidPayload, want: http.StatusUnprocessableEntity},
		{err: apperrors.ErrNotFound, want: http.StatusNotFound},
		{err: errors.New("boom"), want: http.StatusInternalServerError},
	}
	for _, tc := range cases {
		err := mapServiceError(tc.err)
		statusErr, ok := err.(interface{ GetStatus() int })
		if !ok || statusErr.GetStatus() != tc.want {
			t.Errorf("mapServiceError(%v) = %v, ingin status %d", tc.err, err, tc.want)
		}
	}
}
