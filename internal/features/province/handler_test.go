package province

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

type fakeService struct {
	listResult []entities.Province
	listErr    error
}

func (f *fakeService) List(ctx context.Context) ([]entities.Province, error) {
	return f.listResult, f.listErr
}

func (f *fakeService) GetByID(ctx context.Context, id int64) (entities.Province, error) {
	return entities.Province{}, errors.New("not implemented")
}

func (f *fakeService) Create(ctx context.Context, in entities.Province) (entities.Province, error) {
	return entities.Province{}, errors.New("not implemented")
}

func (f *fakeService) Update(ctx context.Context, id int64, in entities.Province) (entities.Province, error) {
	return entities.Province{}, errors.New("not implemented")
}

func (f *fakeService) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func TestList(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		_, api := humatest.New(t)
		svc := &fakeService{listResult: []entities.Province{entities.Province{ID: 1}}}
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
	})

	t.Run("gagal", func(t *testing.T) {
		_, api := humatest.New(t)
		svc := &fakeService{listErr: errors.New("db down")}
		RegisterRoutes(api, svc)

		resp := api.Get("/api/v1/provinces")
		if resp.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, ingin 500", resp.Code)
		}
	})
}

// TestMapServiceError covers the default branch.
func TestMapServiceError(t *testing.T) {
	if got := mapServiceError(errors.New("boom")); got == nil {
		t.Error("mapServiceError() ingin error")
	}
}
