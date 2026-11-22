package event

import (
	"context"
	"errors"
	"net/http"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	"github.com/Junx27/kicaugo-backend/internal/middlewares"
	"github.com/Junx27/kicaugo-backend/internal/utils/authctx"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	apphttp "github.com/Junx27/kicaugo-backend/internal/utils/http"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
	"github.com/Junx27/kicaugo-backend/internal/utils/request"
	"github.com/Junx27/kicaugo-backend/internal/utils/response"
)

// ToData mengubah NearbyResult menjadi data response event terdekat.
func ToData(result NearbyResult) response.NearbyEventsData {
	data := response.NearbyEventsData{
		OriginSource: originNone,
		RadiusKM:     result.RadiusKM,
		Hint:         result.Hint,
		Events:       make([]response.NearbyEventData, 0, len(result.Events)),
	}
	if result.Origin != nil {
		latitude, longitude := result.Origin.Latitude, result.Origin.Longitude
		data.OriginLatitude = &latitude
		data.OriginLongitude = &longitude
		data.OriginSource = result.Origin.Source
	}
	for _, item := range result.Events {
		data.Events = append(data.Events, toEventData(item))
	}
	return data
}

func toEventData(e entities.Event) response.NearbyEventData {
	data := response.NearbyEventData{
		ID:         e.ID,
		Title:      e.Title,
		Slug:       e.Slug,
		Venue:      e.Venue,
		City:       e.City,
		Province:   e.Province,
		Status:     e.Status,
		BannerURL:  e.BannerURL,
		DistanceKM: roundDistance(e.DistanceKM),
	}
	if e.StartDate != nil {
		value := e.StartDate.Format(dateLayout)
		data.StartDate = &value
	}
	if e.EndDate != nil {
		value := e.EndDate.Format(dateLayout)
		data.EndDate = &value
	}
	return data
}

// RegisterRoutes mendaftarkan route event ke Huma API.
func RegisterRoutes(api apphttp.API, service Service, tokens *jwt.Manager) {
	// GET /api/v1/events/nearby — butuh token karena asal berasal dari profil pengguna.
	apphttp.Register(api, apphttp.Operation{
		ID:      "listNearbyEvents",
		Method:  http.MethodGet,
		Path:    "/api/v1/events/nearby",
		Summary: "List Nearby Events",
		Tags:    []string{"event"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: []apphttp.Middleware{
			middlewares.AuthMiddleware[jwt.Claims](tokens),
		},
	}, func(ctx context.Context, input *request.ListNearbyEventsInput) (*response.NearbyEventsOutput, error) {
		identifier, err := authctx.IdentifierFromContext(ctx)
		if err != nil {
			return nil, err
		}
		result, err := service.ListNearby(ctx, identifier, input.RadiusKM, input.CategoryID, input.Limit)
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.NearbyEventsOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Nearby events loaded", ToData(result))
		return resp, nil
	})
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, apperrors.ErrInvalidRadius),
		errors.Is(err, apperrors.ErrInvalidPayload):
		return apphttp.UnprocessableEntity("Please check your input data")
	case errors.Is(err, apperrors.ErrNotFound):
		return apphttp.NotFound("User not found")
	default:
		return apphttp.InternalServerError("An error occurred. Please try again.")
	}
}
