package basemap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/Junx27/kicaugo-backend/internal/middlewares"
	"github.com/Junx27/kicaugo-backend/internal/utils/authctx"
	apphttp "github.com/Junx27/kicaugo-backend/internal/utils/http"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
)

const (
	basemapTag             = "basemap"
	bearerAuthSecurityName = "bearerAuth"
	tileContentType        = "image/png"
	cacheControlFormat     = "private, max-age=%d"
)

// TileInput adalah input GET /api/v1/basemap/tiles/{z}/{x}/{y}.
type TileInput struct {
	Z      int    `path:"z" minimum:"0" maximum:"20" doc:"Zoom level (0-20)"`
	X      int    `path:"x" doc:"Tile X coordinate"`
	Y      int    `path:"y" doc:"Tile Y coordinate"`
	Retina bool   `query:"retina" default:"false" doc:"Gunakan tile retina (@2x)"`
	Style  string `query:"style" default:"voyager" doc:"Style basemap"`
}

// TileOutput adalah respons tile PNG.
type TileOutput struct {
	ContentType  string `header:"Content-Type"`
	CacheControl string `header:"Cache-Control"`
	Body         []byte
}

// newGetTileHandler membentuk handler agar dapat diuji langsung.
func newGetTileHandler(service Service) func(context.Context, *TileInput) (*TileOutput, error) {
	return func(ctx context.Context, input *TileInput) (*TileOutput, error) {
		if _, err := authctx.IdentifierFromContext(ctx); err != nil {
			return nil, err
		}
		if err := validateTile(input.Z, input.X, input.Y); err != nil {
			return nil, apphttp.BadRequest("Invalid tile coordinates")
		}
		data, err := service.Get(ctx, input.Z, input.X, input.Y, input.Retina, input.Style)
		if err != nil {
			return nil, mapServiceError(err)
		}
		return &TileOutput{
			ContentType:  tileContentType,
			CacheControl: fmt.Sprintf(cacheControlFormat, int64(service.CacheTTL()/time.Second)),
			Body:         data,
		}, nil
	}
}

// RegisterRoutes mendaftarkan route basemap ke Huma API.
func RegisterRoutes(api apphttp.API, service Service, tokens *jwt.Manager, limiter apphttp.Middleware) {
	auth := []apphttp.Middleware{middlewares.AuthMiddleware[jwt.Claims](tokens)}
	security := []map[string][]string{{bearerAuthSecurityName: {}}}

	apphttp.Register(api, apphttp.Operation{
		ID:          "getBasemapTile",
		Method:      http.MethodGet,
		Path:        "/api/v1/basemap/tiles/{z}/{x}/{y}",
		Summary:     "Get Basemap Tile",
		Tags:        []string{basemapTag},
		Security:    security,
		Middlewares: append(auth, limiter),
		Responses: map[string]*huma.Response{
			"200": {
				Description: "Tile basemap PNG",
				Content: map[string]*huma.MediaType{
					tileContentType: {
						Schema: &huma.Schema{Type: "string", Format: "binary"},
					},
				},
			},
		},
	}, newGetTileHandler(service))
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, ErrNotConfigured):
		return apphttp.Error(http.StatusServiceUnavailable, "Basemap service not configured")
	case errors.Is(err, ErrInvalidTile), errors.Is(err, ErrUnsupportedStyle):
		return apphttp.BadRequest("Invalid basemap request")
	default:
		return apphttp.Error(http.StatusBadGateway, "Basemap service unavailable")
	}
}
