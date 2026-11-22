package birdcategory

import (
	"context"
	"errors"
	"net/http"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	"github.com/Junx27/kicaugo-backend/internal/middlewares"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	apphttp "github.com/Junx27/kicaugo-backend/internal/utils/http"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
	"github.com/Junx27/kicaugo-backend/internal/utils/response"
)

func toData(c entities.BirdCategory) response.BirdCategoryData {
	return response.BirdCategoryData{ID: c.ID, Name: c.Name, Slug: c.Slug, Icon: c.Icon}
}

// RegisterRoutes mendaftarkan route bird-category ke Huma API.
func RegisterRoutes(api apphttp.API, service Service, tokens *jwt.Manager) {
	// GET /api/v1/bird-categories
	apphttp.Register(api, apphttp.Operation{
		ID:      "listBirdCategories",
		Method:  http.MethodGet,
		Path:    "/api/v1/bird-categories",
		Summary: "List Bird Categories",
		Tags:    []string{"bird-categories"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: []apphttp.Middleware{
			middlewares.AuthMiddleware[jwt.Claims](tokens),
		},
	}, func(ctx context.Context, input *struct{}) (*response.BirdCategoryListOutput, error) {
		cats, err := service.List(ctx)
		if err != nil {
			return nil, mapServiceError(err)
		}
		data := make([]response.BirdCategoryData, 0, len(cats))
		for _, c := range cats {
			data = append(data, toData(c))
		}
		resp := &response.BirdCategoryListOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Bird categories loaded", data)
		return resp, nil
	})

	// scaffold:routes
}

func mapServiceError(err error) error {
	if errors.Is(err, apperrors.ErrNotFound) {
		return apphttp.NotFound("Bird category not found")
	}
	return apphttp.InternalServerError("An error occurred. Please try again.")
}
