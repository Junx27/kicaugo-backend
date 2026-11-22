package profile

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	"github.com/Junx27/kicaugo-backend/internal/middlewares"
	"github.com/Junx27/kicaugo-backend/internal/utils/authctx"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	apphttp "github.com/Junx27/kicaugo-backend/internal/utils/http"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
	"github.com/Junx27/kicaugo-backend/internal/utils/request"
	"github.com/Junx27/kicaugo-backend/internal/utils/response"
)

// ToData mengubah ProfileView menjadi data response profil.
func ToData(view ProfileView) response.ProfileData {
	p := view.Profile
	data := response.ProfileData{
		FullName:           p.FullName,
		Phone:              p.Phone,
		AvatarURL:          p.AvatarURL,
		Bio:                p.Bio,
		SearchRadiusKM:     p.SearchRadiusKM,
		FavoriteCategories: make([]response.FavoriteCategoryData, 0, len(view.Favorites)),
		IsComplete:         view.IsComplete,
		MissingFields:      view.MissingFields,
	}
	if p.City != nil {
		city := &response.ProfileCityData{ID: p.City.ID, Name: p.City.Name, Type: p.City.Type, Slug: p.City.Slug}
		if p.Province != nil {
			city.Province = response.ProfileProvinceData{ID: p.Province.ID, Code: p.Province.Code, Name: p.Province.Name}
		}
		data.City = city
	}
	if p.CurrentLatitude != nil && p.CurrentLongitude != nil {
		location := &response.CurrentLocationData{
			Latitude:  *p.CurrentLatitude,
			Longitude: *p.CurrentLongitude,
			AccuracyM: p.CurrentLocationAccuracyM,
		}
		if p.CurrentLocationUpdatedAt != nil {
			location.UpdatedAt = p.CurrentLocationUpdatedAt.Format(time.RFC3339)
		}
		data.CurrentLocation = location
	}
	for _, c := range view.Favorites {
		data.FavoriteCategories = append(data.FavoriteCategories, response.FavoriteCategoryData{ID: c.ID, Name: c.Name})
	}
	return data
}

// RegisterRoutes mendaftarkan route profil ke Huma API.
func RegisterRoutes(api apphttp.API, service Service, tokens *jwt.Manager) {
	security := []map[string][]string{{"bearerAuth": {}}}
	authMiddleware := []apphttp.Middleware{middlewares.AuthMiddleware[jwt.Claims](tokens)}

	// GET /api/v1/me/profile
	apphttp.Register(api, apphttp.Operation{
		ID: "getMyProfile", Method: http.MethodGet, Path: "/api/v1/me/profile",
		Summary: "Get My Profile", Tags: []string{"profile"},
		Security: security, Middlewares: authMiddleware,
	}, func(ctx context.Context, input *struct{}) (*response.ProfileOutput, error) {
		identifier, err := authctx.IdentifierFromContext(ctx)
		if err != nil {
			return nil, err
		}
		view, err := service.Get(ctx, identifier)
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.ProfileOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Profile loaded", ToData(view))
		return resp, nil
	})

	// PATCH /api/v1/me/profile
	apphttp.Register(api, apphttp.Operation{
		ID: "updateMyProfile", Method: http.MethodPatch, Path: "/api/v1/me/profile",
		Summary: "Update My Profile", Tags: []string{"profile"},
		Security: security, Middlewares: authMiddleware,
	}, func(ctx context.Context, input *request.UpdateProfileInput) (*response.ProfileMutationOutput, error) {
		identifier, err := authctx.IdentifierFromContext(ctx)
		if err != nil {
			return nil, err
		}
		in := entities.ProfileUpdate{
			FullName:       input.Body.FullName,
			Phone:          input.Body.Phone,
			AvatarURL:      input.Body.AvatarURL,
			Bio:            input.Body.Bio,
			CityID:         input.Body.CityID,
			SearchRadiusKM: input.Body.SearchRadiusKM,
		}
		if err := service.Update(ctx, identifier, in, input.Body.ProvinceID); err != nil {
			return nil, mapServiceError(err)
		}
		return newMutationResponse(), nil
	})

	// PUT /api/v1/me/profile/current-location
	apphttp.Register(api, apphttp.Operation{
		ID: "putMyCurrentLocation", Method: http.MethodPut, Path: "/api/v1/me/profile/current-location",
		Summary: "Update My Current Location", Tags: []string{"profile"},
		Security: security, Middlewares: authMiddleware,
	}, func(ctx context.Context, input *request.CurrentLocationInput) (*response.ProfileMutationOutput, error) {
		identifier, err := authctx.IdentifierFromContext(ctx)
		if err != nil {
			return nil, err
		}
		err = service.UpdateLocation(ctx, identifier, input.Body.Latitude, input.Body.Longitude, input.Body.AccuracyM)
		if err != nil {
			return nil, mapServiceError(err)
		}
		return newMutationResponse(), nil
	})

	// PUT /api/v1/me/profile/favorite-categories
	apphttp.Register(api, apphttp.Operation{
		ID: "putMyFavoriteCategories", Method: http.MethodPut, Path: "/api/v1/me/profile/favorite-categories",
		Summary: "Replace My Favorite Categories", Tags: []string{"profile"},
		Security: security, Middlewares: authMiddleware,
	}, func(ctx context.Context, input *request.FavoriteCategoriesInput) (*response.ProfileMutationOutput, error) {
		identifier, err := authctx.IdentifierFromContext(ctx)
		if err != nil {
			return nil, err
		}
		if err := service.ReplaceFavorites(ctx, identifier, input.Body.CategoryIDs); err != nil {
			return nil, mapServiceError(err)
		}
		return newMutationResponse(), nil
	})
}

func newMutationResponse() *response.ProfileMutationOutput {
	resp := &response.ProfileMutationOutput{}
	resp.Status = http.StatusOK
	resp.Body = response.NewSuccess(http.StatusOK, "Profile updated", struct{}{})
	return resp
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, apperrors.ErrInvalidPayload),
		errors.Is(err, apperrors.ErrInvalidCity),
		errors.Is(err, apperrors.ErrProvinceMismatch),
		errors.Is(err, apperrors.ErrInvalidCoordinates),
		errors.Is(err, apperrors.ErrInvalidCategory),
		errors.Is(err, apperrors.ErrInvalidRadius):
		return apphttp.UnprocessableEntity("Please check your input data")
	case errors.Is(err, apperrors.ErrNotFound):
		return apphttp.NotFound("User not found")
	default:
		return apphttp.InternalServerError("An error occurred. Please try again.")
	}
}
