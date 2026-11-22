package organizer

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
)

type fakeService struct {
	createResult entities.Organizer
	createErr    error
	memberResult entities.EventOrganizerUser
	memberErr    error
}

func (f *fakeService) Create(ctx context.Context, identifier string, in entities.Organizer) (entities.Organizer, error) {
	return f.createResult, f.createErr
}

func (f *fakeService) RegisterMember(ctx context.Context, identifier string, organizerID int64, role string) (entities.EventOrganizerUser, error) {
	return f.memberResult, f.memberErr
}

func (f *fakeService) HasActiveMembership(ctx context.Context, identifier string) (bool, error) {
	return false, nil
}
func (f *fakeService) IsActiveMember(ctx context.Context, identifier string, organizerID int64) (bool, error) {
	return false, nil
}

func setup(t *testing.T, svc Service) (humatest.TestAPI, *jwt.Manager) {
	t.Helper()
	_, api := humatest.New(t)
	tokens := jwt.NewManager([]byte("rahasia-untuk-test"), time.Hour)
	RegisterRoutes(api, svc, tokens)
	return api, tokens
}

func authHeader(t *testing.T, tokens *jwt.Manager) string {
	t.Helper()
	token, err := tokens.Issue("user@example.com", "Tester")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	return "Authorization: Bearer " + token
}

func TestCreateOrganizerHandler(t *testing.T) {
	body := `{"name":"EO Kicau","slug":"eo-kicau","email":"eo@example.com"}`

	t.Run("berhasil", func(t *testing.T) {
		api, tokens := setup(t, &fakeService{createResult: entities.Organizer{ID: 1, Name: "EO Kicau", Slug: "eo-kicau", Email: "eo@example.com"}})
		resp := api.Post("/api/v1/organizers", "Content-Type: application/json", authHeader(t, tokens), strings.NewReader(body))
		if resp.Code != http.StatusCreated {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
		var out struct {
			Data struct {
				ID int64 `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil || out.Data.ID != 1 {
			t.Fatalf("body=%s err=%v", resp.Body.String(), err)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		api, _ := setup(t, &fakeService{})
		resp := api.Post("/api/v1/organizers", "Content-Type: application/json", strings.NewReader(body))
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d", resp.Code)
		}
	})

	t.Run("conflict", func(t *testing.T) {
		api, tokens := setup(t, &fakeService{createErr: apperrors.ErrOrganizerSlugTaken})
		resp := api.Post("/api/v1/organizers", "Content-Type: application/json", authHeader(t, tokens), strings.NewReader(body))
		if resp.Code != http.StatusConflict {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
	})
}

func TestRegisterOrganizerMemberHandler(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		api, tokens := setup(t, &fakeService{memberResult: entities.EventOrganizerUser{ID: 2, OrganizerID: 1, UserID: 7, Role: "staff", IsActive: true}})
		resp := api.Post("/api/v1/organizers/1/members", "Content-Type: application/json", authHeader(t, tokens), strings.NewReader(`{"role":"staff"}`))
		if resp.Code != http.StatusCreated {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		api, tokens := setup(t, &fakeService{memberErr: apperrors.ErrOrganizerNotFound})
		resp := api.Post("/api/v1/organizers/99/members", "Content-Type: application/json", authHeader(t, tokens), strings.NewReader(`{"role":"staff"}`))
		if resp.Code != http.StatusNotFound {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
	})

	t.Run("invalid payload", func(t *testing.T) {
		api, tokens := setup(t, &fakeService{memberErr: apperrors.ErrInvalidPayload})
		resp := api.Post("/api/v1/organizers/1/members", "Content-Type: application/json", authHeader(t, tokens), strings.NewReader(`{"role":"boss"}`))
		if resp.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
	})

	t.Run("internal error", func(t *testing.T) {
		api, tokens := setup(t, &fakeService{memberErr: errors.New("db down")})
		resp := api.Post("/api/v1/organizers/1/members", "Content-Type: application/json", authHeader(t, tokens), strings.NewReader(`{"role":"staff"}`))
		if resp.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
		}
	})
}
