package city

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	"github.com/Junx27/kicaugo-backend/internal/utils/response"
)

type fakeService struct {
	result     []entities.City
	err        error
	provinceID int64
	query      string
}

func (f *fakeService) ListByProvince(ctx context.Context, provinceID int64, q string) ([]entities.City, error) {
	f.provinceID = provinceID
	f.query = q
	return f.result, f.err
}

func TestListCities(t *testing.T) {
	t.Run("berhasil tanpa token", func(t *testing.T) {
		_, api := humatest.New(t)
		svc := &fakeService{result: []entities.City{
			{ID: 1, Name: "Banda Aceh", Type: "kota", Slug: "kota-banda-aceh"},
		}}
		RegisterRoutes(api, svc)

		resp := api.Get("/api/v1/cities?province_id=11")
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, ingin 200; body = %s", resp.Code, resp.Body.String())
		}
		var body struct {
			Status int                 `json:"status"`
			Data   []response.CityData `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("body bukan JSON: %v", err)
		}
		if body.Status != http.StatusOK || len(body.Data) != 1 || body.Data[0].Slug != "kota-banda-aceh" {
			t.Errorf("body = %+v", body)
		}
		if svc.provinceID != 11 {
			t.Errorf("provinceID = %d, ingin 11", svc.provinceID)
		}
		if got := resp.Header().Get("Cache-Control"); got == "" {
			t.Error("header Cache-Control tidak di-set")
		}
	})

	t.Run("meneruskan pencarian nama", func(t *testing.T) {
		_, api := humatest.New(t)
		svc := &fakeService{}
		RegisterRoutes(api, svc)

		resp := api.Get("/api/v1/cities?province_id=11&name=banda")
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, ingin 200", resp.Code)
		}
		if svc.query != "banda" {
			t.Errorf("name = %q, ingin banda", svc.query)
		}
	})

	t.Run("province_id wajib", func(t *testing.T) {
		_, api := humatest.New(t)
		RegisterRoutes(api, &fakeService{})

		resp := api.Get("/api/v1/cities")
		if resp.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, ingin 422", resp.Code)
		}
	})

	t.Run("gagal", func(t *testing.T) {
		_, api := humatest.New(t)
		RegisterRoutes(api, &fakeService{err: errors.New("db down")})

		resp := api.Get("/api/v1/cities?province_id=11")
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
