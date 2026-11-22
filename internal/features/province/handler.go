package province

import (
	"context"
	"errors"
	"net/http"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	apphttp "github.com/Junx27/kicaugo-backend/internal/utils/http"
	"github.com/Junx27/kicaugo-backend/internal/utils/response"
)

func toData(c entities.Province) response.ProvinceData {
	p := response.ProvinceData{ID: c.ID}
	p.Code = c.Code
	p.Name = c.Name
	p.Slug = c.Slug
	return p
}

// RegisterRoutes mendaftarkan route province ke Huma API.
func RegisterRoutes(api apphttp.API, service Service) {
	// GET /api/v1/provinces
	apphttp.Register(api, apphttp.Operation{
		ID:      "listProvinces",
		Method:  http.MethodGet,
		Path:    "/api/v1/provinces",
		Summary: "List Provinces",
		Tags:    []string{"province"},
	}, func(ctx context.Context, input *struct{}) (*response.ProvinceListOutput, error) {
		items, err := service.List(ctx)
		if err != nil {
			return nil, mapServiceError(err)
		}
		data := make([]response.ProvinceData, 0, len(items))
		for _, it := range items {
			data = append(data, toData(it))
		}
		resp := &response.ProvinceListOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Provinces loaded", data)
		return resp, nil
	})

	// scaffold:routes
}

func mapServiceError(err error) error {
	if errors.Is(err, apperrors.ErrNotFound) {
		return apphttp.NotFound("Province not found")
	}
	return apphttp.InternalServerError("An error occurred. Please try again.")
}
