package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	"github.com/Junx27/kicaugo-backend/internal/utils/request"
)

func ruleBodyJSON(severity string) string {
	return fmt.Sprintf(`{"title":"Aturan Sangkar","description":"Deskripsi","severity":%q,"sort_order":3}`, severity)
}

func assertStatus(t *testing.T, resp *httptest.ResponseRecorder, want int) {
	t.Helper()
	if resp.Code != want {
		t.Fatalf("status = %d, ingin %d; body = %s", resp.Code, want, resp.Body.String())
	}
}

func TestListEventRulesHandler(t *testing.T) {
	t.Run("sukses dengan deskripsi null jadi string kosong", func(t *testing.T) {
		svc := &fakeService{rules: []entities.ContestRule{{
			ID: 1, Title: "Aturan Sangkar", Severity: entities.RuleSeverityInfo, SortOrder: 1,
		}}}
		api, tokens := setupHandler(t, svc)
		resp := api.Get("/api/v1/events/101/rules", bearerHeader(t, tokens))
		assertStatus(t, resp, http.StatusOK)

		var body struct {
			Data []struct {
				Title       string  `json:"title"`
				Description *string `json:"description"`
				Severity    string  `json:"severity"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("json: %v", err)
		}
		if len(body.Data) != 1 || body.Data[0].Title != "Aturan Sangkar" || body.Data[0].Severity != entities.RuleSeverityInfo {
			t.Fatalf("data = %+v", body.Data)
		}
		if body.Data[0].Description == nil || *body.Data[0].Description != "" {
			t.Fatalf("description = %v", body.Data[0].Description)
		}
	})

	t.Run("id invalid", func(t *testing.T) {
		api, tokens := setupHandler(t, &fakeService{})
		resp := api.Get("/api/v1/events/0/rules", bearerHeader(t, tokens))
		assertStatus(t, resp, http.StatusUnprocessableEntity)
	})

	t.Run("tanpa token", func(t *testing.T) {
		api, _ := setupHandler(t, &fakeService{})
		resp := api.Get("/api/v1/events/101/rules")
		assertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("error service", func(t *testing.T) {
		svc := &fakeService{ruleErr: apperrors.ErrNotFound}
		api, tokens := setupHandler(t, svc)
		resp := api.Get("/api/v1/events/101/rules", bearerHeader(t, tokens))
		assertStatus(t, resp, http.StatusNotFound)
	})
}

func TestCreateEventRuleHandler(t *testing.T) {
	path := "/api/v1/events/101/rules"

	t.Run("sukses untuk semua severity", func(t *testing.T) {
		for _, severity := range []string{entities.RuleSeverityInfo, entities.RuleSeverityError, entities.RuleSeveritySuccess} {
			t.Run(severity, func(t *testing.T) {
				svc := &fakeService{rule: entities.ContestRule{
					ID: 2, Title: "Aturan Sangkar", Description: strPtr("Deskripsi"), Severity: severity,
				}}
				api, tokens := setupHandler(t, svc)
				resp := api.Post(path, bearerHeader(t, tokens), "Content-Type: application/json", strings.NewReader(ruleBodyJSON(severity)))
				assertStatus(t, resp, http.StatusCreated)

				var body struct {
					Data struct {
						Title    string `json:"title"`
						Severity string `json:"severity"`
					} `json:"data"`
				}
				if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
					t.Fatalf("json: %v", err)
				}
				if body.Data.Title != "Aturan Sangkar" || body.Data.Severity != severity {
					t.Fatalf("data = %+v", body.Data)
				}
				if svc.ruleIn.EventID != 101 || svc.ruleIn.Title != "Aturan Sangkar" ||
					svc.ruleIn.Severity != severity || svc.ruleIn.SortOrder != 3 {
					t.Fatalf("ruleIn = %+v", svc.ruleIn)
				}
			})
		}
	})

	t.Run("severity di luar enum", func(t *testing.T) {
		api, tokens := setupHandler(t, &fakeService{})
		resp := api.Post(path, bearerHeader(t, tokens), "Content-Type: application/json", strings.NewReader(ruleBodyJSON("warning")))
		assertStatus(t, resp, http.StatusUnprocessableEntity)
	})

	t.Run("tanpa token", func(t *testing.T) {
		api, _ := setupHandler(t, &fakeService{})
		resp := api.Post(path, "Content-Type: application/json", strings.NewReader(ruleBodyJSON("info")))
		assertStatus(t, resp, http.StatusUnauthorized)
	})

	t.Run("bukan organizer aktif", func(t *testing.T) {
		api, tokens := setupHandlerWithOrganizer(t, &fakeService{}, fakeOrganizerChecker{active: false})
		resp := api.Post(path, bearerHeader(t, tokens), "Content-Type: application/json", strings.NewReader(ruleBodyJSON("info")))
		assertStatus(t, resp, http.StatusForbidden)
	})

	t.Run("error service", func(t *testing.T) {
		cases := []struct {
			name string
			err  error
			want int
		}{
			{"payload invalid", apperrors.ErrInvalidPayload, http.StatusUnprocessableEntity},
			{"tidak ditemukan", apperrors.ErrNotFound, http.StatusNotFound},
			{"forbidden", apperrors.ErrForbidden, http.StatusForbidden},
			{"db error", errors.New("db down"), http.StatusInternalServerError},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				api, tokens := setupHandler(t, &fakeService{ruleErr: tc.err})
				resp := api.Post(path, bearerHeader(t, tokens), "Content-Type: application/json", strings.NewReader(ruleBodyJSON("info")))
				assertStatus(t, resp, tc.want)
			})
		}
	})

	t.Run("konteks tanpa identitas", func(t *testing.T) {
		handler := newCreateEventRuleHandler(&fakeService{})
		if _, err := handler(context.Background(), &request.CreateEventRuleInput{}); err == nil {
			t.Fatal("ingin error identitas")
		}
	})
}

func TestUpdateEventRuleHandler(t *testing.T) {
	path := "/api/v1/events/101/rules/2"

	t.Run("sukses", func(t *testing.T) {
		svc := &fakeService{rule: entities.ContestRule{
			ID: 2, Title: "Aturan Diubah", Description: strPtr("Deskripsi baru"), Severity: entities.RuleSeverityError,
		}}
		api, tokens := setupHandler(t, svc)
		resp := api.Patch(path, bearerHeader(t, tokens), "Content-Type: application/json", strings.NewReader(ruleBodyJSON(entities.RuleSeverityError)))
		assertStatus(t, resp, http.StatusOK)

		var body struct {
			Data struct {
				Title string `json:"title"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("json: %v", err)
		}
		if body.Data.Title != "Aturan Diubah" {
			t.Fatalf("data = %+v", body.Data)
		}
		if svc.ruleUpdateIn.EventID != 101 || svc.ruleUpdateIn.RuleID != 2 || svc.ruleUpdateIn.Severity != entities.RuleSeverityError {
			t.Fatalf("ruleUpdateIn = %+v", svc.ruleUpdateIn)
		}
	})

	t.Run("id rule invalid", func(t *testing.T) {
		api, tokens := setupHandler(t, &fakeService{})
		resp := api.Patch("/api/v1/events/101/rules/0", bearerHeader(t, tokens), "Content-Type: application/json", strings.NewReader(ruleBodyJSON("info")))
		assertStatus(t, resp, http.StatusUnprocessableEntity)
	})

	t.Run("error service", func(t *testing.T) {
		svc := &fakeService{ruleErr: apperrors.ErrNotFound}
		api, tokens := setupHandler(t, svc)
		resp := api.Patch(path, bearerHeader(t, tokens), "Content-Type: application/json", strings.NewReader(ruleBodyJSON("info")))
		assertStatus(t, resp, http.StatusNotFound)
	})

	t.Run("konteks tanpa identitas", func(t *testing.T) {
		handler := newUpdateEventRuleHandler(&fakeService{})
		if _, err := handler(context.Background(), &request.UpdateEventRuleInput{}); err == nil {
			t.Fatal("ingin error identitas")
		}
	})
}

func TestDeleteEventRuleHandler(t *testing.T) {
	path := "/api/v1/events/101/rules/2"

	t.Run("sukses", func(t *testing.T) {
		svc := &fakeService{}
		api, tokens := setupHandler(t, svc)
		resp := api.Delete(path, bearerHeader(t, tokens))
		assertStatus(t, resp, http.StatusOK)
		if svc.deleteRuleIDs != [2]int64{101, 2} {
			t.Fatalf("deleteRuleIDs = %v", svc.deleteRuleIDs)
		}
	})

	t.Run("error service", func(t *testing.T) {
		svc := &fakeService{ruleErr: errors.New("db down")}
		api, tokens := setupHandler(t, svc)
		resp := api.Delete(path, bearerHeader(t, tokens))
		assertStatus(t, resp, http.StatusInternalServerError)
	})

	t.Run("konteks tanpa identitas", func(t *testing.T) {
		handler := newDeleteEventRuleHandler(&fakeService{})
		if _, err := handler(context.Background(), &request.DeleteEventRuleInput{}); err == nil {
			t.Fatal("ingin error identitas")
		}
	})
}
