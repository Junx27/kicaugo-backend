package birdcategory

import (
	"context"
	"errors"
	"net/http"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	apphttp "github.com/Junx27/kicaugo-backend/internal/utils/http"
	"github.com/Junx27/kicaugo-backend/internal/utils/response"
)

// BirdCategoryPayload adalah representasi kategori burung di response.
type BirdCategoryPayload struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Icon string `json:"icon"`
}

// ListOutput adalah output untuk GET /bird-categories.
type ListOutput struct {
	Status int `status:"200"`
	Body   response.Success[[]BirdCategoryPayload]
}

// GetOutput adalah output untuk GET /bird-categories/{id}.
type GetOutput struct {
	Status int `status:"200"`
	Body   response.Success[BirdCategoryPayload]
}

func toPayload(c entities.BirdCategory) BirdCategoryPayload {
	return BirdCategoryPayload{ID: c.ID, Name: c.Name, Slug: c.Slug, Icon: c.Icon}
}

// RegisterRoutes mendaftarkan route bird-category ke Huma API.
func RegisterRoutes(api apphttp.API, service Service) {
	// GET /api/v1/bird-categories
	apphttp.Register(api, apphttp.Operation{
		ID:      "listBirdCategories",
		Method:  http.MethodGet,
		Path:    "/api/v1/bird-categories",
		Summary: "List Bird Categories",
		Tags:    []string{"bird-categories"},
	}, func(ctx context.Context, input *struct{}) (*ListOutput, error) {
		cats, err := service.List(ctx)
		if err != nil {
			return nil, mapServiceError(err)
		}
		payloads := make([]BirdCategoryPayload, 0, len(cats))
		for _, c := range cats {
			payloads = append(payloads, toPayload(c))
		}
		resp := &ListOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Bird categories loaded", payloads)
		return resp, nil
	})

	// GET /api/v1/bird-categories/{id}
	apphttp.Register(api, apphttp.Operation{
		ID:      "getBirdCategory",
		Method:  http.MethodGet,
		Path:    "/api/v1/bird-categories/{id}",
		Summary: "Get Bird Category",
		Tags:    []string{"bird-categories"},
	}, func(ctx context.Context, input *struct {
		ID int64 `path:"id"`
	}) (*GetOutput, error) {
		cat, err := service.GetByID(ctx, input.ID)
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &GetOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Bird category loaded", toPayload(cat))
		return resp, nil
	})
}

func mapServiceError(err error) error {
	if errors.Is(err, apperrors.ErrNotFound) {
		return apphttp.NotFound("Bird category not found")
	}
	return apphttp.InternalServerError("An error occurred. Please try again.")
}
