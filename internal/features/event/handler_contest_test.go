package event

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	"github.com/Junx27/kicaugo-backend/internal/testutil"
)

func sampleContestDetail() entities.ContestDetail {
	addr := "Jl. Test"
	lat, lng := -6.2, 106.8
	hero := "https://cdn/hero.jpg"
	banner := "https://cdn/banner.jpg"
	mapImg := "https://cdn/map.jpg"
	prizePool := "50000000"
	session := "Sesi 04"
	start := time.Date(2025, 5, 18, 8, 30, 0, 0, time.UTC)
	remaining := 4
	return entities.ContestDetail{
		ID: 101, Title: "Piala Raja", Subtitle: "Subtitle", OfficialBadge: "Resmi", FormatBadge: "24G", EditionLabel: "Edisi",
		HeroImageURL: &hero, BannerImageURL: &banner, Status: "published", TotalSlots: 36, ActiveTickets: 24,
		ScheduleDate: &start, ScheduleTime: "08:30 WIB – Selesai", GateOpen: "07:00 WIB",
		Venue: "Batavia Arena", Address: &addr, DistanceText: "5 Menit", MapImageURL: &mapImg,
		Latitude: &lat, Longitude: &lng, Parking: "Parkir", PrizePool: &prizePool,
		PrizeTrophy: "Trofi", PrizeHighlight: "Highlight",
		StatFormatLabel: "36 Gantangan", StatFormatNote: "Tertib", StatJudgingLabel: "PBI", StatJudgingNote: "Tanpa Koordinasi",
		StatFacilitiesLabel: "VIP", StatFacilitiesNote: "LED",
		Classes: []entities.ContestClass{{ID: 1, Name: "Murai", Session: &session, StartTime: &start, RemainingSlots: &remaining, Price: 500000, PrizeHighlight: "Juara 1"}},
		Rules:   []entities.ContestRule{{ID: 1, Title: "Aturan", Description: strPtr("Desc"), Severity: "info", SortOrder: 1}},
	}
}

func strPtr(v string) *string { return &v }

func timeNow() *time.Time {
	t := time.Date(2025, 5, 18, 8, 30, 0, 0, time.UTC)
	return &t
}

func TestGetContestDetailHandler(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		svc := &fakeService{contestDetail: sampleContestDetail()}
		api, tokens := setupHandler(t, svc)
		resp := api.Get("/api/v1/contests/101", testutil.BearerHeader(t, tokens))
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
		}
		var body struct {
			Message string `json:"message"`
			Data    struct {
				ID     int64  `json:"id"`
				Title  string `json:"title"`
				Status string `json:"status"`
				Quota  struct {
					Total     int `json:"total"`
					Remaining int `json:"remaining"`
				} `json:"quota"`
				Classes []struct {
					ID   int64  `json:"id"`
					Time string `json:"time"`
				} `json:"classes"`
				Rules []struct {
					Title string `json:"title"`
				} `json:"rules"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("json: %v", err)
		}
		if body.Message != "Detail lomba berhasil dimuat" || body.Data.ID != 101 || body.Data.Status != "open" {
			t.Errorf("body = %+v", body)
		}
		if body.Data.Quota.Total != 36 || body.Data.Quota.Remaining != 12 {
			t.Errorf("quota = %+v", body.Data.Quota)
		}
		if len(body.Data.Classes) != 1 || body.Data.Classes[0].Time == "" {
			t.Errorf("classes = %+v", body.Data.Classes)
		}
		if len(body.Data.Rules) != 1 {
			t.Errorf("rules = %+v", body.Data.Rules)
		}
	})
	t.Run("ditolak tanpa token", func(t *testing.T) {
		api, _ := setupHandler(t, &fakeService{})
		if resp := api.Get("/api/v1/contests/101"); resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})
	t.Run("error service", func(t *testing.T) {
		svc := &fakeService{contestErr: errors.New("db down")}
		api, tokens := setupHandler(t, svc)
		resp := api.Get("/api/v1/contests/101", testutil.BearerHeader(t, tokens))
		if resp.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
		}
	})
}
