package home

import (
	"context"
	"errors"
	"net/http"
	"strings"
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

const (
	homeTag                = "home"
	bearerAuthSecurityName = "bearerAuth"
)

func firstName(fullName string) string {
	name := strings.TrimSpace(fullName)
	if name == "" {
		return ""
	}
	parts := strings.Fields(name)
	return parts[0]
}

// toData memilih field ringkas untuk response home profile.
func toData(view View) response.HomeProfileData {
	data := response.HomeProfileData{
		Name:        firstName(view.Profile.Profile.FullName),
		Avatar:      view.Profile.Profile.AvatarURL,
		Message:     view.Greeting,
		LiveCount:   view.OngoingEventsCount,
		IsOrganizer: view.IsOrganizer,
	}
	if view.ActiveStreaming != nil {
		streaming := toActiveStreamingData(*view.ActiveStreaming)
		data.ActiveStreaming = &streaming
	}
	return data
}

func toActiveStreamingData(s entities.ActiveEventStreaming) response.HomeActiveStreamingData {
	data := response.HomeActiveStreamingData{
		ID: s.ID, EventID: s.EventID, EventTitle: s.EventTitle, EventSlug: s.EventSlug,
		Venue: s.Venue, City: s.City, LinkStream: s.LinkStream,
		StartTime: s.StartTime.Format(time.RFC3339), TotalViews: s.TotalViews, Juara: s.Juara,
	}
	if s.EndTime != nil {
		v := s.EndTime.Format(time.RFC3339)
		data.EndTime = &v
	}
	return data
}

func toFeaturedData(data Featured) response.HomeFeaturedData {
	return response.HomeFeaturedData{
		Title:     data.Title,
		Schedule:  data.Schedule,
		Distance:  data.Distance,
		Prize:     data.Prize,
		Slots:     data.Slots,
		Address:   data.Address,
		Latitude:  data.Latitude,
		Longitude: data.Longitude,
		Attendees: response.FeaturedAttendeesData{
			Initials: data.Attendees.Initials,
			Others:   data.Attendees.Others,
		},
	}
}

func toUpcomingData(data Upcoming) response.HomeUpcomingData {
	return response.HomeUpcomingData{
		Type:             data.Type,
		Title:            data.Title,
		Time:             data.Time,
		Price:            data.Price,
		RemainingTickets: data.RemainingTickets,
	}
}

func toStatsData(data Stats) response.HomeStatsData {
	return response.HomeStatsData{
		Badge:        data.Badge,
		Headline:     data.Headline,
		Note:         data.Note,
		SecurityNote: data.SecurityNote,
	}
}

func toActiveTicketData(data ActiveTicket) response.HomeActiveTicketData {
	return response.HomeActiveTicketData{
		Event:     data.Event,
		Schedule:  data.Schedule,
		Number:    data.Number,
		Session:   data.Session,
		Address:   data.Address,
		Latitude:  data.Latitude,
		Longitude: data.Longitude,
	}
}

func toSchedulesData(items []Schedule) []response.HomeScheduleData {
	data := make([]response.HomeScheduleData, 0, len(items))
	for _, item := range items {
		data = append(data, response.HomeScheduleData{
			Title:  item.Title,
			Venue:  item.Venue,
			Date:   item.Date,
			Price:  item.Price,
			Status: item.Status,
			Category: response.HomeScheduleCategoryData{
				ID:   item.Category.ID,
				Name: item.Category.Name,
			},
		})
	}
	return data
}

func toBirdCategoriesData(items []entities.BirdCategory) []response.BirdCategoryData {
	data := make([]response.BirdCategoryData, 0, len(items))
	for _, item := range items {
		data = append(data, response.BirdCategoryData{
			ID:   item.ID,
			Name: item.Name,
			Slug: item.Slug,
			Icon: item.Icon,
		})
	}
	return data
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

func newGetFeaturedHandler(service Service) func(context.Context, *struct{}) (*response.HomeFeaturedOutput, error) {
	return func(ctx context.Context, _ *struct{}) (*response.HomeFeaturedOutput, error) {
		identifier, err := authctx.IdentifierFromContext(ctx)
		if err != nil {
			return nil, err
		}
		featured, err := service.GetFeatured(ctx, identifier)
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.HomeFeaturedOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Home featured loaded", toFeaturedData(featured))
		return resp, nil
	}
}

func newGetUpcomingHandler(service Service) func(context.Context, *struct{}) (*response.HomeUpcomingOutput, error) {
	return func(ctx context.Context, _ *struct{}) (*response.HomeUpcomingOutput, error) {
		if _, err := authctx.IdentifierFromContext(ctx); err != nil {
			return nil, err
		}
		upcoming, err := service.GetUpcoming(ctx)
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.HomeUpcomingOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Home upcoming loaded", toUpcomingData(upcoming))
		return resp, nil
	}
}

func newGetStatsHandler(service Service) func(context.Context, *struct{}) (*response.HomeStatsOutput, error) {
	return func(ctx context.Context, _ *struct{}) (*response.HomeStatsOutput, error) {
		if _, err := authctx.IdentifierFromContext(ctx); err != nil {
			return nil, err
		}
		resp := &response.HomeStatsOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Home stats loaded", toStatsData(service.GetStats()))
		return resp, nil
	}
}

func newGetActiveTicketHandler(service Service) func(context.Context, *struct{}) (*response.HomeActiveTicketOutput, error) {
	return func(ctx context.Context, _ *struct{}) (*response.HomeActiveTicketOutput, error) {
		if _, err := authctx.IdentifierFromContext(ctx); err != nil {
			return nil, err
		}
		resp := &response.HomeActiveTicketOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Home active ticket loaded", toActiveTicketData(service.GetActiveTicket()))
		return resp, nil
	}
}

func newGetSchedulesHandler(service Service) func(context.Context, *struct{}) (*response.HomeSchedulesOutput, error) {
	return func(ctx context.Context, _ *struct{}) (*response.HomeSchedulesOutput, error) {
		if _, err := authctx.IdentifierFromContext(ctx); err != nil {
			return nil, err
		}
		resp := &response.HomeSchedulesOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Home schedules loaded", toSchedulesData(service.GetSchedules()))
		return resp, nil
	}
}

func newGetBirdCategoriesHandler(service Service) func(context.Context, *struct{}) (*response.HomeBirdCategoriesOutput, error) {
	return func(ctx context.Context, _ *struct{}) (*response.HomeBirdCategoriesOutput, error) {
		if _, err := authctx.IdentifierFromContext(ctx); err != nil {
			return nil, err
		}
		categories, err := service.GetBirdCategories(ctx)
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.HomeBirdCategoriesOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Home bird categories loaded", toBirdCategoriesData(categories))
		return resp, nil
	}
}

// RegisterRoutes mendaftarkan route home ke Huma API.
func RegisterRoutes(api apphttp.API, service Service, tokens *jwt.Manager) {
	auth := []apphttp.Middleware{middlewares.AuthMiddleware[jwt.Claims](tokens)}
	security := []map[string][]string{{bearerAuthSecurityName: {}}}

	apphttp.Register(api, apphttp.Operation{
		ID:          "getHomeProfile",
		Method:      http.MethodGet,
		Path:        "/api/v1/home/profile",
		Summary:     "Get Home Profile",
		Tags:        []string{homeTag},
		Security:    security,
		Middlewares: auth,
	}, newGetProfileHandler(service))

	apphttp.Register(api, apphttp.Operation{
		ID:          "getHomeFeatured",
		Method:      http.MethodGet,
		Path:        "/api/v1/home/featured",
		Summary:     "Get Home Featured",
		Tags:        []string{homeTag},
		Security:    security,
		Middlewares: auth,
	}, newGetFeaturedHandler(service))

	apphttp.Register(api, apphttp.Operation{
		ID:          "getHomeUpcoming",
		Method:      http.MethodGet,
		Path:        "/api/v1/home/upcoming",
		Summary:     "Get Home Upcoming",
		Tags:        []string{homeTag},
		Security:    security,
		Middlewares: auth,
	}, newGetUpcomingHandler(service))

	apphttp.Register(api, apphttp.Operation{
		ID:          "getHomeStats",
		Method:      http.MethodGet,
		Path:        "/api/v1/home/stats",
		Summary:     "Get Home Stats",
		Tags:        []string{homeTag},
		Security:    security,
		Middlewares: auth,
	}, newGetStatsHandler(service))

	apphttp.Register(api, apphttp.Operation{
		ID:          "getHomeActiveTicket",
		Method:      http.MethodGet,
		Path:        "/api/v1/home/active-ticket",
		Summary:     "Get Home Active Ticket",
		Tags:        []string{homeTag},
		Security:    security,
		Middlewares: auth,
	}, newGetActiveTicketHandler(service))

	apphttp.Register(api, apphttp.Operation{
		ID:          "getHomeSchedules",
		Method:      http.MethodGet,
		Path:        "/api/v1/home/schedules",
		Summary:     "Get Home Schedules",
		Tags:        []string{homeTag},
		Security:    security,
		Middlewares: auth,
	}, newGetSchedulesHandler(service))

	apphttp.Register(api, apphttp.Operation{
		ID:          "getHomeBirdCategories",
		Method:      http.MethodGet,
		Path:        "/api/v1/home/bird-categories",
		Summary:     "Get Home Bird Categories",
		Tags:        []string{homeTag},
		Security:    security,
		Middlewares: auth,
	}, newGetBirdCategoriesHandler(service))
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
