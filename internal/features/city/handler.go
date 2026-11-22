package city

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

func toData(c entities.City) response.CityData {
	p := response.CityData{ID: c.ID}
	p.ProvinceID = c.ProvinceID
	p.Code = c.Code
	p.Name = c.Name
	p.Type = c.Type
	p.Slug = c.Slug
	p.Latitude = c.Latitude
	p.Longitude = c.Longitude
	return p
}

// RegisterRoutes mendaftarkan route city ke Huma API.
func RegisterRoutes(api apphttp.API, service Service, tokens *jwt.Manager) {
	// GET /api/v1/cities
	apphttp.Register(api, apphttp.Operation{
		ID:      "listCities",
		Method:  http.MethodGet,
		Path:    "/api/v1/cities",
		Summary: "List Cities",
		Tags:    []string{"city"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: []apphttp.Middleware{
			middlewares.AuthMiddleware[jwt.Claims](tokens),
		},
	}, func(ctx context.Context, input *struct{}) (*response.CityListOutput, error) {
		items, err := service.List(ctx)
		if err != nil {
			return nil, mapServiceError(err)
		}
		data := make([]response.CityData, 0, len(items))
		for _, it := range items {
			data = append(data, toData(it))
		}
		resp := &response.CityListOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Cities loaded", data)
		return resp, nil
	})

	// scaffold:routes
}

func mapServiceError(err error) error {
	if errors.Is(err, apperrors.ErrNotFound) {
		return apphttp.NotFound("City not found")
	}
	return apphttp.InternalServerError("An error occurred. Please try again.")
}
