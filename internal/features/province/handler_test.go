package province

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

type fakeService struct {
	listResult []entities.Province
	listErr    error
}

func (f *fakeService) List(ctx context.Context) ([]entities.Province, error) {
	return f.listResult, f.listErr
}

func TestList(t *testing.T) {
	t.Run("berhasil tanpa token", func(t *testing.T) {
		_, api := humatest.New(t)
		svc := &fakeService{listResult: []entities.Province{{ID: 1, Code: "11", Name: "Aceh", Slug: "aceh"}}}
		RegisterRoutes(api, svc)

		resp := api.Get("/api/v1/provinces")
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
		if got := resp.Header().Get("Cache-Control"); got == "" {
			t.Error("header Cache-Control tidak di-set")
		}
	})

	t.Run("gagal", func(t *testing.T) {
		_, api := humatest.New(t)
		RegisterRoutes(api, &fakeService{listErr: errors.New("db down")})

		resp := api.Get("/api/v1/provinces")
		if resp.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, ingin 500", resp.Code)
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
