package birdcategory

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
	listResult []entities.BirdCategory
	listErr    error
}

func (f *fakeService) List(ctx context.Context) ([]entities.BirdCategory, error) {
	return f.listResult, f.listErr
}

func TestList(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		_, api := humatest.New(t)
		svc := &fakeService{listResult: []entities.BirdCategory{{ID: 1, Name: "Murai Batu", Slug: "murai_batu", Icon: "murai_batu"}}}
		RegisterRoutes(api, svc)

		resp := api.Get("/api/v1/bird-categories")
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, ingin 200; body = %s", resp.Code, resp.Body.String())
		}
		var body struct {
			Status int                   `json:"status"`
			Data   []BirdCategoryPayload `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("body bukan JSON: %v", err)
		}
		if body.Status != http.StatusOK {
			t.Errorf("status = %d, ingin %d", body.Status, http.StatusOK)
		}
		if len(body.Data) != 1 || body.Data[0].Name != "Murai Batu" {
			t.Errorf("data = %+v", body.Data)
		}
	})

	t.Run("gagal", func(t *testing.T) {
		_, api := humatest.New(t)
		svc := &fakeService{listErr: errors.New("db down")}
		RegisterRoutes(api, svc)

		resp := api.Get("/api/v1/bird-categories")
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
