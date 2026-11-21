package home

import (
	"context"
	"net/http"

	"github.com/Junx27/kicaugo-backend/internal/auth"
	"github.com/Junx27/kicaugo-backend/internal/response"
	"github.com/danielgtaylor/huma/v2"
)

// ContentService adalah bagian Service yang dibutuhkan handler.
type ContentService interface {
	Get(ctx context.Context) (Content, error)
}

// HomeOutput output untuk GET /home
type HomeOutput struct {
	Status int `status:"200"`
	Body   response.Success[Content]
}

// RegisterRoutes mendaftarkan route home ke Huma API
func RegisterRoutes(api huma.API, service ContentService, tokens *auth.TokenManager) {
	// GET /api/v1/home
	huma.Register(api, huma.Operation{
		OperationID: "getHome",
		Method:      http.MethodGet,
		Path:        "/api/v1/home",
		Summary:     "Get Home",
		Tags:        []string{"home"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: []func(ctx huma.Context, next func(huma.Context)){auth.AuthMiddleware(tokens)},
	}, func(ctx context.Context, input *struct{}) (*HomeOutput, error) {
		content, err := service.Get(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to load data. Please try again.")
		}
		resp := &HomeOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Home content loaded successfully", content)
		return resp, nil
	})
}
