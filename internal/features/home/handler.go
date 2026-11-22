package home

import (
	"context"
	"errors"
	"net/http"

	"github.com/Junx27/kicaugo-backend/internal/features/event"
	"github.com/Junx27/kicaugo-backend/internal/features/profile"
	"github.com/Junx27/kicaugo-backend/internal/middlewares"
	"github.com/Junx27/kicaugo-backend/internal/utils/authctx"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	apphttp "github.com/Junx27/kicaugo-backend/internal/utils/http"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
	"github.com/Junx27/kicaugo-backend/internal/utils/request"
	"github.com/Junx27/kicaugo-backend/internal/utils/response"
)

// toData memilih field profil yang aman untuk home dan menggabungkan data nearby.
func toData(view View) response.HomeData {
	profileData := profile.ToData(view.Profile)
	return response.HomeData{
		Profile: response.HomeProfileData{
			FullName:           profileData.FullName,
			AvatarURL:          profileData.AvatarURL,
			City:               profileData.City,
			SearchRadiusKM:     profileData.SearchRadiusKM,
			FavoriteCategories: profileData.FavoriteCategories,
			IsComplete:         profileData.IsComplete,
			MissingFields:      profileData.MissingFields,
		},
		Nearby: event.ToData(view.Nearby),
	}
}

// newGetProfileHandler membentuk handler agar dapat diuji langsung tanpa Huma.
func newGetProfileHandler(service Service) func(context.Context, *request.HomeProfileInput) (*response.HomeOutput, error) {
	return func(ctx context.Context, input *request.HomeProfileInput) (*response.HomeOutput, error) {
		identifier, err := authctx.IdentifierFromContext(ctx)
		if err != nil {
			return nil, err
		}
		opts := Options{RadiusKM: input.RadiusKM, CategoryID: input.CategoryID, Limit: input.Limit}
		view, err := service.Get(ctx, identifier, opts)
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.HomeOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Home profile loaded", toData(view))
		return resp, nil
	}
}

// RegisterRoutes mendaftarkan route home ke Huma API.
func RegisterRoutes(api apphttp.API, service Service, tokens *jwt.Manager) {
	apphttp.Register(api, apphttp.Operation{
		ID:      "getHomeProfile",
		Method:  http.MethodGet,
		Path:    "/api/v1/home/profile",
		Summary: "Get Home Profile",
		Tags:    []string{"home"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: []apphttp.Middleware{
			middlewares.AuthMiddleware[jwt.Claims](tokens),
		},
	}, newGetProfileHandler(service))
}

// mapServiceError memetakan error domain ke status HTTP.
func mapServiceError(err error) error {
	switch {
	case errors.Is(err, apperrors.ErrInvalidRadius),
		errors.Is(err, apperrors.ErrInvalidPayload),
		errors.Is(err, apperrors.ErrInvalidCity),
		errors.Is(err, apperrors.ErrProvinceMismatch),
		errors.Is(err, apperrors.ErrInvalidCoordinates),
		errors.Is(err, apperrors.ErrInvalidCategory):
		return apphttp.UnprocessableEntity("Please check your input data")
	case errors.Is(err, apperrors.ErrNotFound):
		return apphttp.NotFound("User not found")
	default:
		return apphttp.InternalServerError("An error occurred. Please try again.")
	}
}
