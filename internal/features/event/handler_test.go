package event

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
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
	createResult      entities.EventDetail
	createErr         error
	createIn          entities.EventCreate
	createID          string
	streamingResult   entities.EventStreaming
	streamingErr      error
	streamingIn       entities.StartEventStreamingInput
	streamingID       string
	updateJuaraResult entities.EventStreaming
	updateJuaraErr    error
	updateJuaraIn     entities.UpdateEventStreamingJuaraInput
	updateJuaraID     string
	stopResult        entities.EventStreaming
	stopErr           error
	stopEventID       int64
	stopID            string
}

func (f *fakeService) ListNearby(ctx context.Context, identifier string, radiusKM int, categoryID int64, limit int) (NearbyResult, error) {
	f.params.radius = radiusKM
	f.params.cat = categoryID
	f.params.limit = limit
	return f.result, f.err
}

func (f *fakeService) CountOngoing(ctx context.Context) (int, error) {
	return 0, nil
}

func (f *fakeService) Create(ctx context.Context, identifier string, in entities.EventCreate) (entities.EventDetail, error) {
	f.createIn = in
	f.createID = identifier
	return f.createResult, f.createErr
}

func (f *fakeService) StartStreaming(ctx context.Context, identifier string, in entities.StartEventStreamingInput) (entities.EventStreaming, error) {
	f.streamingIn = in
	f.streamingID = identifier
	return f.streamingResult, f.streamingErr
}

func (f *fakeService) UpdateStreamingJuara(ctx context.Context, identifier string, in entities.UpdateEventStreamingJuaraInput) (entities.EventStreaming, error) {
	f.updateJuaraIn = in
	f.updateJuaraID = identifier
	return f.updateJuaraResult, f.updateJuaraErr
}

func (f *fakeService) StopStreaming(ctx context.Context, identifier string, eventID int64) (entities.EventStreaming, error) {
	f.stopEventID = eventID
	f.stopID = identifier
	return f.stopResult, f.stopErr
}

func (f *fakeService) GetActiveStreaming(ctx context.Context) (*entities.ActiveEventStreaming, error) {
	return nil, nil
}

func (f *fakeService) GetRandomFeatured(ctx context.Context) (*entities.Event, error) {
	return nil, nil
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

type fakeOrganizerChecker struct {
	active bool
	err    error
}

func (f fakeOrganizerChecker) HasActiveMembership(ctx context.Context, identifier string) (bool, error) {
	return f.active, f.err
}

func setupHandler(t *testing.T, svc Service) (humatest.TestAPI, *jwt.Manager) {
	t.Helper()
	_, api := humatest.New(t)
	tokens := newTestTokens()
	RegisterRoutes(api, svc, tokens, fakeOrganizerChecker{active: true})
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

func TestCreateEventHandler(t *testing.T) {
	body := `{"organizer_id":1,"title":"Lomba","slug":"lomba","venue":"GOR","city":"Bandung","sub_events":[{"title":"Kelas A","max_slots":60}]}`
	t.Run("berhasil", func(t *testing.T) {
		start := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, 12, 2, 0, 0, 0, 0, time.UTC)
		st := time.Date(2026, 12, 1, 8, 0, 0, 0, time.UTC)
		et := time.Date(2026, 12, 1, 10, 0, 0, 0, time.UTC)
		province := "Jawa Barat"
		svc := &fakeService{createResult: entities.EventDetail{Event: entities.Event{ID: 11, OrganizerID: 1, Title: "Lomba", Slug: "lomba", Venue: "GOR", City: "Bandung", Province: province, Status: "draft", StartDate: &start, EndDate: &end}, SubEvents: []entities.SubEvent{{ID: 21, EventID: 11, Title: "Kelas A", MaxSlots: 60, Status: "available", StartTime: &st, EndTime: &et}}}}
		_, api := humatest.New(t)
		tokens := newTestTokens()
		RegisterRoutes(api, svc, tokens, fakeOrganizerChecker{active: true})
		resp := api.Post("/api/v1/events", bearerHeader(t, tokens), "Content-Type: application/json", strings.NewReader(body))
		if resp.Code != http.StatusCreated {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
		var out struct {
			Data struct {
				Province  *string `json:"province"`
				StartDate *string `json:"start_date"`
				EndDate   *string `json:"end_date"`
				SubEvents []struct {
					StartTime *string `json:"start_time"`
					EndTime   *string `json:"end_time"`
				} `json:"sub_events"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
			t.Fatalf("json: %v", err)
		}
		if out.Data.Province == nil || *out.Data.Province != "Jawa Barat" || out.Data.StartDate == nil || out.Data.EndDate == nil {
			t.Fatalf("data=%+v", out.Data)
		}
		if len(out.Data.SubEvents) != 1 || out.Data.SubEvents[0].StartTime == nil || out.Data.SubEvents[0].EndTime == nil {
			t.Fatalf("sub=%+v", out.Data.SubEvents)
		}
		if svc.createIn.OrganizerID != 1 || len(svc.createIn.SubEvents) != 1 {
			t.Fatalf("createIn=%+v", svc.createIn)
		}
	})

	t.Run("forbidden bukan organizer", func(t *testing.T) {
		_, api := humatest.New(t)
		tokens := newTestTokens()
		RegisterRoutes(api, &fakeService{}, tokens, fakeOrganizerChecker{active: false})
		resp := api.Post("/api/v1/events", bearerHeader(t, tokens), "Content-Type: application/json", strings.NewReader(body))
		if resp.Code != http.StatusForbidden {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		_, api := humatest.New(t)
		RegisterRoutes(api, &fakeService{}, newTestTokens(), fakeOrganizerChecker{active: true})
		resp := api.Post("/api/v1/events", "Content-Type: application/json", strings.NewReader(body))
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d", resp.Code)
		}
	})

	t.Run("service error", func(t *testing.T) {
		_, api := humatest.New(t)
		tokens := newTestTokens()
		RegisterRoutes(api, &fakeService{createErr: apperrors.ErrForbidden}, tokens, fakeOrganizerChecker{active: true})
		resp := api.Post("/api/v1/events", bearerHeader(t, tokens), "Content-Type: application/json", strings.NewReader(body))
		if resp.Code != http.StatusForbidden {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
	})
}

func TestStartEventStreamingHandler(t *testing.T) {
	body := `{"link_stream":"https://youtube.com/live/abc"}`
	t.Run("berhasil", func(t *testing.T) {
		start := time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)
		svc := &fakeService{streamingResult: entities.EventStreaming{ID: 3, EventID: 68, StartTime: start, TotalViews: 0, Juara: []byte(`{}`), LinkStream: "https://youtube.com/live/abc"}}
		_, api := humatest.New(t)
		tokens := newTestTokens()
		RegisterRoutes(api, svc, tokens, fakeOrganizerChecker{active: true})
		resp := api.Post("/api/v1/events/68/streamings/start", bearerHeader(t, tokens), "Content-Type: application/json", strings.NewReader(body))
		if resp.Code != http.StatusCreated {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
		var out struct {
			Data struct {
				ID         int64           `json:"id"`
				EventID    int64           `json:"event_id"`
				LinkStream string          `json:"link_stream"`
				Juara      json.RawMessage `json:"juara"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
			t.Fatalf("json: %v", err)
		}
		if out.Data.ID != 3 || out.Data.EventID != 68 || out.Data.LinkStream != "https://youtube.com/live/abc" || string(out.Data.Juara) != "{}" {
			t.Fatalf("data=%+v juara=%s", out.Data, string(out.Data.Juara))
		}
		if svc.streamingIn.EventID != 68 || svc.streamingIn.LinkStream != "https://youtube.com/live/abc" {
			t.Fatalf("in=%+v", svc.streamingIn)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		_, api := humatest.New(t)
		RegisterRoutes(api, &fakeService{}, newTestTokens(), fakeOrganizerChecker{active: true})
		resp := api.Post("/api/v1/events/68/streamings/start", "Content-Type: application/json", strings.NewReader(body))
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d", resp.Code)
		}
	})

	t.Run("forbidden bukan organizer", func(t *testing.T) {
		_, api := humatest.New(t)
		tokens := newTestTokens()
		RegisterRoutes(api, &fakeService{}, tokens, fakeOrganizerChecker{active: false})
		resp := api.Post("/api/v1/events/68/streamings/start", bearerHeader(t, tokens), "Content-Type: application/json", strings.NewReader(body))
		if resp.Code != http.StatusForbidden {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
	})

	t.Run("conflict streaming aktif", func(t *testing.T) {
		_, api := humatest.New(t)
		tokens := newTestTokens()
		RegisterRoutes(api, &fakeService{streamingErr: apperrors.ErrStreamingAlreadyActive}, tokens, fakeOrganizerChecker{active: true})
		resp := api.Post("/api/v1/events/68/streamings/start", bearerHeader(t, tokens), "Content-Type: application/json", strings.NewReader(body))
		if resp.Code != http.StatusConflict {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
	})

	t.Run("event not streamable", func(t *testing.T) {
		_, api := humatest.New(t)
		tokens := newTestTokens()
		RegisterRoutes(api, &fakeService{streamingErr: apperrors.ErrEventNotStreamable}, tokens, fakeOrganizerChecker{active: true})
		resp := api.Post("/api/v1/events/68/streamings/start", bearerHeader(t, tokens), "Content-Type: application/json", strings.NewReader(body))
		if resp.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
	})
}

func TestUpdateEventStreamingJuaraHandler(t *testing.T) {
	body := `{"juara":{"rank1":"A"}}`
	t.Run("berhasil", func(t *testing.T) {
		svc := &fakeService{updateJuaraResult: entities.EventStreaming{ID: 3, EventID: 68, Juara: []byte(`{"rank1":"A"}`), LinkStream: "https://youtube.com/live/abc"}}
		_, api := humatest.New(t)
		tokens := newTestTokens()
		RegisterRoutes(api, svc, tokens, fakeOrganizerChecker{active: true})
		resp := api.Patch("/api/v1/events/68/streamings/juara", bearerHeader(t, tokens), "Content-Type: application/json", strings.NewReader(body))
		if resp.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
		var out struct {
			Data struct {
				ID    int64           `json:"id"`
				Juara json.RawMessage `json:"juara"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
			t.Fatalf("json: %v", err)
		}
		if out.Data.ID != 3 || string(out.Data.Juara) != `{"rank1":"A"}` || svc.updateJuaraIn.EventID != 68 {
			t.Fatalf("out=%+v in=%+v", out, svc.updateJuaraIn)
		}
	})
	t.Run("unauthorized", func(t *testing.T) {
		_, api := humatest.New(t)
		RegisterRoutes(api, &fakeService{}, newTestTokens(), fakeOrganizerChecker{active: true})
		if resp := api.Patch("/api/v1/events/68/streamings/juara", "Content-Type: application/json", strings.NewReader(body)); resp.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d", resp.Code)
		}
	})
	t.Run("forbidden bukan organizer", func(t *testing.T) {
		_, api := humatest.New(t)
		tokens := newTestTokens()
		RegisterRoutes(api, &fakeService{}, tokens, fakeOrganizerChecker{active: false})
		if resp := api.Patch("/api/v1/events/68/streamings/juara", bearerHeader(t, tokens), "Content-Type: application/json", strings.NewReader(body)); resp.Code != http.StatusForbidden {
			t.Fatalf("status=%d", resp.Code)
		}
	})
	t.Run("no active streaming", func(t *testing.T) {
		_, api := humatest.New(t)
		tokens := newTestTokens()
		RegisterRoutes(api, &fakeService{updateJuaraErr: apperrors.ErrStreamingNotActive}, tokens, fakeOrganizerChecker{active: true})
		if resp := api.Patch("/api/v1/events/68/streamings/juara", bearerHeader(t, tokens), "Content-Type: application/json", strings.NewReader(body)); resp.Code != http.StatusConflict {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
	})
}

func TestStopEventStreamingHandler(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		end := time.Date(2026, 11, 22, 11, 0, 0, 0, time.UTC)
		svc := &fakeService{stopResult: entities.EventStreaming{ID: 3, EventID: 68, EndTime: &end, LinkStream: "https://youtube.com/live/abc"}}
		_, api := humatest.New(t)
		tokens := newTestTokens()
		RegisterRoutes(api, svc, tokens, fakeOrganizerChecker{active: true})
		resp := api.Post("/api/v1/events/68/streamings/stop", bearerHeader(t, tokens))
		if resp.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
		var out struct {
			Data struct {
				ID      int64   `json:"id"`
				EndTime *string `json:"end_time"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
			t.Fatalf("json: %v", err)
		}
		if out.Data.ID != 3 || out.Data.EndTime == nil || svc.stopEventID != 68 {
			t.Fatalf("out=%+v svc=%+v", out, svc)
		}
	})
	t.Run("unauthorized", func(t *testing.T) {
		_, api := humatest.New(t)
		RegisterRoutes(api, &fakeService{}, newTestTokens(), fakeOrganizerChecker{active: true})
		if resp := api.Post("/api/v1/events/68/streamings/stop"); resp.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d", resp.Code)
		}
	})
	t.Run("forbidden bukan organizer", func(t *testing.T) {
		_, api := humatest.New(t)
		tokens := newTestTokens()
		RegisterRoutes(api, &fakeService{}, tokens, fakeOrganizerChecker{active: false})
		if resp := api.Post("/api/v1/events/68/streamings/stop", bearerHeader(t, tokens)); resp.Code != http.StatusForbidden {
			t.Fatalf("status=%d", resp.Code)
		}
	})
	t.Run("no active streaming", func(t *testing.T) {
		_, api := humatest.New(t)
		tokens := newTestTokens()
		RegisterRoutes(api, &fakeService{stopErr: apperrors.ErrStreamingNotActive}, tokens, fakeOrganizerChecker{active: true})
		if resp := api.Post("/api/v1/events/68/streamings/stop", bearerHeader(t, tokens)); resp.Code != http.StatusConflict {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
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
		{err: apperrors.ErrInvalidCategory, want: http.StatusUnprocessableEntity},
		{err: apperrors.ErrEventNotStreamable, want: http.StatusUnprocessableEntity},
		{err: apperrors.ErrNotFound, want: http.StatusNotFound},
		{err: apperrors.ErrOrganizerNotFound, want: http.StatusNotFound},
		{err: apperrors.ErrForbidden, want: http.StatusForbidden},
		{err: apperrors.ErrIdentifierTaken, want: http.StatusConflict},
		{err: apperrors.ErrStreamingAlreadyActive, want: http.StatusConflict},
		{err: apperrors.ErrStreamingNotActive, want: http.StatusConflict},
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
