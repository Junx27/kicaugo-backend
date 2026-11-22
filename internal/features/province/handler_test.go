package province

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
)

type fakeService struct {
	listResult []entities.Province
	listErr    error
}

func (f *fakeService) List(ctx context.Context) ([]entities.Province, error) {
	return f.listResult, f.listErr
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

func TestList(t *testing.T) {
	t.Run("berhasil dengan token valid", func(t *testing.T) {
		_, api := humatest.New(t)
		svc := &fakeService{listResult: []entities.Province{{ID: 1, Code: "11", Name: "Aceh", Slug: "aceh"}}}
		tokens := newTestTokens()
		RegisterRoutes(api, svc, tokens)

		resp := api.Get("/api/v1/provinces", bearerHeader(t, tokens))
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, ingin 200; body = %s", resp.Code, resp.Body.String())
		}
		var body struct {
			Status int `json:"status"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("body bukan JSON: %v", err)
		}
		if body.Status != http.StatusOK {
			t.Errorf("status = %d, ingin %d", body.Status, http.StatusOK)
		}
	})

	t.Run("gagal dengan token valid", func(t *testing.T) {
		_, api := humatest.New(t)
		svc := &fakeService{listErr: errors.New("db down")}
		tokens := newTestTokens()
		RegisterRoutes(api, svc, tokens)

		resp := api.Get("/api/v1/provinces", bearerHeader(t, tokens))
		if resp.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, ingin 500", resp.Code)
		}
	})

	t.Run("ditolak tanpa token", func(t *testing.T) {
		_, api := humatest.New(t)
		RegisterRoutes(api, &fakeService{}, newTestTokens())

		resp := api.Get("/api/v1/provinces")
		if resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})

	t.Run("ditolak dengan token invalid", func(t *testing.T) {
		_, api := humatest.New(t)
		RegisterRoutes(api, &fakeService{}, newTestTokens())

		resp := api.Get("/api/v1/provinces", "Authorization: Bearer token-salah")
		if resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})
}

// TestMapServiceError covers kedua cabang, termasuk ErrNotFound.
func TestMapServiceError(t *testing.T) {
	for _, err := range []error{errors.New("boom"), apperrors.ErrNotFound} {
		if got := mapServiceError(err); got == nil {
			t.Errorf("mapServiceError(%v) ingin error", err)
		}
	}
}
