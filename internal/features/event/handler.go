package event

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	"github.com/Junx27/kicaugo-backend/internal/middlewares"
	"github.com/Junx27/kicaugo-backend/internal/utils/authctx"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	"github.com/Junx27/kicaugo-backend/internal/utils/format"
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

func toContestDetailData(detail entities.ContestDetail) response.ContestDetailData {
	remaining := detail.TotalSlots - detail.ActiveTickets
	if remaining < 0 {
		remaining = 0
	}
	classes := make([]response.ContestClassData, 0, len(detail.Classes))
	for _, c := range detail.Classes {
		timeText := ""
		if c.StartTime != nil {
			timeText = c.StartTime.Format("15:04") + " WIB"
		}
		classes = append(classes, response.ContestClassData{
			ID:             c.ID,
			Name:           c.Name,
			Session:        c.Session,
			Time:           timeText,
			RemainingSlots: c.RemainingSlots,
			Price:          format.Rupiah(c.Price),
			PrizeHighlight: c.PrizeHighlight,
		})
	}
	rules := make([]response.ContestRuleData, 0, len(detail.Rules))
	for _, r := range detail.Rules {
		rules = append(rules, toContestRuleData(r))
	}
	scheduleDate := ""
	if detail.ScheduleDate != nil {
		scheduleDate = formatIndonesianDate(*detail.ScheduleDate)
	}
	prizeTotal := "-"
	if detail.PrizePool != nil && *detail.PrizePool != "" {
		prizeTotal = format.RupiahString(*detail.PrizePool)
	}
	// starting price
	startingPrice := "Rp 0"
	if len(detail.Classes) > 0 {
		min := detail.Classes[0].Price
		for _, c := range detail.Classes[1:] {
			if c.Price < min {
				min = c.Price
			}
		}
		if min == 0 {
			startingPrice = "Gratis"
		} else {
			startingPrice = format.Rupiah(min)
		}
	}
	status := detail.Status
	if status == "published" || status == "ongoing" {
		status = "open"
	}
	address := ""
	if detail.Address != nil {
		address = *detail.Address
	}
	hero := detail.HeroImageURL
	banner := detail.BannerImageURL
	mapImage := detail.MapImageURL
	return response.ContestDetailData{
		ID:             detail.ID,
		Title:          detail.Title,
		Subtitle:       detail.Subtitle,
		OfficialBadge:  detail.OfficialBadge,
		FormatBadge:    detail.FormatBadge,
		EditionLabel:   detail.EditionLabel,
		HeroImageURL:   hero,
		BannerImageURL: banner,
		Status:         status,
		Quota: response.ContestQuotaData{
			Total:     detail.TotalSlots,
			Remaining: remaining,
			Label:     "Kuota Tersedia",
		},
		Schedule: response.ContestScheduleData{
			Date:     scheduleDate,
			Time:     detail.ScheduleTime,
			GateOpen: detail.GateOpen,
		},
		Location: response.ContestLocationData{
			Name:        detail.Venue,
			Address:     address,
			Distance:    detail.DistanceText,
			MapImageURL: mapImage,
			Latitude:    detail.Latitude,
			Longitude:   detail.Longitude,
			Parking:     detail.Parking,
		},
		Prize: response.ContestPrizeData{
			Total:     prizeTotal,
			Trophy:    detail.PrizeTrophy,
			Highlight: detail.PrizeHighlight,
		},
		Stats: response.ContestStatsData{
			Format:     response.ContestStatItemData{Label: detail.StatFormatLabel, Note: detail.StatFormatNote},
			Judging:    response.ContestStatItemData{Label: detail.StatJudgingLabel, Note: detail.StatJudgingNote},
			Facilities: response.ContestStatItemData{Label: detail.StatFacilitiesLabel, Note: detail.StatFacilitiesNote},
		},
		Classes: classes,
		Rules:   rules,
		Ticket: response.ContestTicketData{
			StartingPrice: startingPrice,
			Unit:          "/gantangan",
			CTA:           "Beli Tiket",
		},
		Distance:     detail.DistanceText,
		ScheduleText: scheduleDate,
	}
}

// toContestRuleData memetakan aturan ke response; deskripsi null dikirim sebagai string kosong.
func toContestRuleData(rule entities.ContestRule) response.ContestRuleData {
	description := ""
	if rule.Description != nil {
		description = *rule.Description
	}
	return response.ContestRuleData{
		Title:       rule.Title,
		Description: &description,
		Severity:    rule.Severity,
	}
}

func toEventRuleInput(body request.EventRuleBody) entities.EventRuleInput {
	return entities.EventRuleInput{
		Title:       body.Title,
		Description: body.Description,
		Severity:    body.Severity,
		SortOrder:   body.SortOrder,
	}
}

func newGetContestDetailHandler(service Service) func(context.Context, *request.ContestDetailInput) (*response.ContestDetailOutput, error) {
	return func(ctx context.Context, input *request.ContestDetailInput) (*response.ContestDetailOutput, error) {
		identifier, err := authctx.IdentifierFromContext(ctx)
		if err != nil {
			return nil, err
		}
		_ = identifier
		detail, err := service.GetContestDetail(ctx, input.ID)
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.ContestDetailOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Detail lomba berhasil dimuat", toContestDetailData(detail))
		return resp, nil
	}
}

func newListEventRulesHandler(service Service) func(context.Context, *request.ListEventRulesInput) (*response.ListEventRulesOutput, error) {
	return func(ctx context.Context, input *request.ListEventRulesInput) (*response.ListEventRulesOutput, error) {
		rules, err := service.ListEventRules(ctx, input.ID)
		if err != nil {
			return nil, mapServiceError(err)
		}
		data := make([]response.ContestRuleData, 0, len(rules))
		for _, rule := range rules {
			data = append(data, toContestRuleData(rule))
		}
		resp := &response.ListEventRulesOutput{Status: http.StatusOK}
		resp.Body = response.NewSuccess(http.StatusOK, "Event rules loaded", data)
		return resp, nil
	}
}

func newCreateEventRuleHandler(service Service) func(context.Context, *request.CreateEventRuleInput) (*response.CreateEventRuleOutput, error) {
	return func(ctx context.Context, input *request.CreateEventRuleInput) (*response.CreateEventRuleOutput, error) {
		identifier, err := authctx.IdentifierFromContext(ctx)
		if err != nil {
			return nil, err
		}
		rule, err := service.CreateEventRule(ctx, identifier, entities.EventRuleCreate{
			EventID: input.ID, EventRuleInput: toEventRuleInput(input.Body),
		})
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.CreateEventRuleOutput{Status: http.StatusCreated}
		resp.Body = response.NewSuccess(http.StatusCreated, "Event rule created", toContestRuleData(rule))
		return resp, nil
	}
}

func newUpdateEventRuleHandler(service Service) func(context.Context, *request.UpdateEventRuleInput) (*response.UpdateEventRuleOutput, error) {
	return func(ctx context.Context, input *request.UpdateEventRuleInput) (*response.UpdateEventRuleOutput, error) {
		identifier, err := authctx.IdentifierFromContext(ctx)
		if err != nil {
			return nil, err
		}
		rule, err := service.UpdateEventRule(ctx, identifier, entities.EventRuleUpdate{
			EventID: input.ID, RuleID: input.RuleID, EventRuleInput: toEventRuleInput(input.Body),
		})
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.UpdateEventRuleOutput{Status: http.StatusOK}
		resp.Body = response.NewSuccess(http.StatusOK, "Event rule updated", toContestRuleData(rule))
		return resp, nil
	}
}

func newDeleteEventRuleHandler(service Service) func(context.Context, *request.DeleteEventRuleInput) (*response.DeleteEventRuleOutput, error) {
	return func(ctx context.Context, input *request.DeleteEventRuleInput) (*response.DeleteEventRuleOutput, error) {
		identifier, err := authctx.IdentifierFromContext(ctx)
		if err != nil {
			return nil, err
		}
		if err := service.DeleteEventRule(ctx, identifier, input.ID, input.RuleID); err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.DeleteEventRuleOutput{Status: http.StatusOK}
		resp.Body = response.NewSuccess(http.StatusOK, "Event rule deleted", struct{}{})
		return resp, nil
	}
}

// RegisterRoutes mendaftarkan route event ke Huma API.
func RegisterRoutes(api apphttp.API, service Service, tokens *jwt.Manager, organizerChecker middlewares.OrganizerChecker) {
	organizerAuth := []apphttp.Middleware{
		middlewares.AuthMiddleware[jwt.Claims](tokens),
		middlewares.OrganizerOnly[jwt.Claims](organizerChecker, func(c jwt.Claims) string { return c.Subject }),
	}
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

	apphttp.Register(api, apphttp.Operation{
		ID:          "createEvent",
		Method:      http.MethodPost,
		Path:        "/api/v1/events",
		Summary:     "Create Event",
		Tags:        []string{"event"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: organizerAuth,
	}, func(ctx context.Context, input *request.CreateEventInput) (*response.CreateEventOutput, error) {
		identifier, err := authctx.IdentifierFromContext(ctx)
		if err != nil {
			return nil, err
		}
		created, err := service.Create(ctx, identifier, toCreateEntity(input))
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.CreateEventOutput{Status: http.StatusCreated}
		resp.Body = response.NewSuccess(http.StatusCreated, "Event created", toCreatedEventData(created))
		return resp, nil
	})

	apphttp.Register(api, apphttp.Operation{
		ID:          "startEventStreaming",
		Method:      http.MethodPost,
		Path:        "/api/v1/events/{id}/streamings/start",
		Summary:     "Start Event Streaming",
		Tags:        []string{"event"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: organizerAuth,
	}, func(ctx context.Context, input *request.StartEventStreamingInput) (*response.StartEventStreamingOutput, error) {
		identifier, err := authctx.IdentifierFromContext(ctx)
		if err != nil {
			return nil, err
		}
		created, err := service.StartStreaming(ctx, identifier, entities.StartEventStreamingInput{
			EventID:    input.ID,
			LinkStream: input.Body.LinkStream,
		})
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.StartEventStreamingOutput{Status: http.StatusCreated}
		resp.Body = response.NewSuccess(http.StatusCreated, "Event streaming started", toStreamingData(created))
		return resp, nil
	})

	apphttp.Register(api, apphttp.Operation{
		ID:          "updateEventStreamingJuara",
		Method:      http.MethodPatch,
		Path:        "/api/v1/events/{id}/streamings/juara",
		Summary:     "Update Event Streaming Juara",
		Tags:        []string{"event"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: organizerAuth,
	}, func(ctx context.Context, input *request.UpdateEventStreamingJuaraInput) (*response.UpdateEventStreamingJuaraOutput, error) {
		identifier, err := authctx.IdentifierFromContext(ctx)
		if err != nil {
			return nil, err
		}
		updated, err := service.UpdateStreamingJuara(ctx, identifier, entities.UpdateEventStreamingJuaraInput{
			EventID: input.ID,
			Juara:   input.Body.Juara,
		})
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.UpdateEventStreamingJuaraOutput{Status: http.StatusOK}
		resp.Body = response.NewSuccess(http.StatusOK, "Event streaming juara updated", toStreamingData(updated))
		return resp, nil
	})

	apphttp.Register(api, apphttp.Operation{
		ID:          "stopEventStreaming",
		Method:      http.MethodPost,
		Path:        "/api/v1/events/{id}/streamings/stop",
		Summary:     "Stop Event Streaming",
		Tags:        []string{"event"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: organizerAuth,
	}, func(ctx context.Context, input *request.StopEventStreamingInput) (*response.StopEventStreamingOutput, error) {
		identifier, err := authctx.IdentifierFromContext(ctx)
		if err != nil {
			return nil, err
		}
		stopped, err := service.StopStreaming(ctx, identifier, input.ID)
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.StopEventStreamingOutput{Status: http.StatusOK}
		resp.Body = response.NewSuccess(http.StatusOK, "Event streaming stopped", toStreamingData(stopped))
		return resp, nil
	})

	apphttp.Register(api, apphttp.Operation{
		ID:          "getContestDetail",
		Method:      http.MethodGet,
		Path:        "/api/v1/contests/{id}",
		Summary:     "Get Contest Detail",
		Tags:        []string{"event"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: []apphttp.Middleware{middlewares.AuthMiddleware[jwt.Claims](tokens)},
	}, newGetContestDetailHandler(service))

	apphttp.Register(api, apphttp.Operation{
		ID:          "listEventRules",
		Method:      http.MethodGet,
		Path:        "/api/v1/events/{id}/rules",
		Summary:     "List Event Rules",
		Tags:        []string{"event"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: []apphttp.Middleware{middlewares.AuthMiddleware[jwt.Claims](tokens)},
	}, newListEventRulesHandler(service))

	apphttp.Register(api, apphttp.Operation{
		ID:          "createEventRule",
		Method:      http.MethodPost,
		Path:        "/api/v1/events/{id}/rules",
		Summary:     "Create Event Rule",
		Tags:        []string{"event"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: organizerAuth,
	}, newCreateEventRuleHandler(service))

	apphttp.Register(api, apphttp.Operation{
		ID:          "updateEventRule",
		Method:      http.MethodPatch,
		Path:        "/api/v1/events/{id}/rules/{ruleID}",
		Summary:     "Update Event Rule",
		Tags:        []string{"event"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: organizerAuth,
	}, newUpdateEventRuleHandler(service))

	apphttp.Register(api, apphttp.Operation{
		ID:          "deleteEventRule",
		Method:      http.MethodDelete,
		Path:        "/api/v1/events/{id}/rules/{ruleID}",
		Summary:     "Delete Event Rule",
		Tags:        []string{"event"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: organizerAuth,
	}, newDeleteEventRuleHandler(service))
}

func toCreateEntity(input *request.CreateEventInput) entities.EventCreate {
	subEvents := make([]entities.SubEventCreate, 0, len(input.Body.SubEvents))
	for _, s := range input.Body.SubEvents {
		subEvents = append(subEvents, entities.SubEventCreate{
			BirdCategoryID: s.BirdCategoryID, Title: s.Title, Description: s.Description, Session: s.Session,
			GantanganCount: s.GantanganCount, MaxSlots: s.MaxSlots, Price: s.Price,
			Winner1Prize: s.Winner1Prize, Winner2Prize: s.Winner2Prize, Winner3Prize: s.Winner3Prize,
			Status: s.Status, StartTime: s.StartTime, EndTime: s.EndTime, SortOrder: s.SortOrder,
		})
	}
	return entities.EventCreate{
		OrganizerID: input.Body.OrganizerID, Title: input.Body.Title, Slug: input.Body.Slug, Description: input.Body.Description,
		Venue: input.Body.Venue, Address: input.Body.Address, City: input.Body.City, Province: input.Body.Province,
		CityID: input.Body.CityID, ProvinceID: input.Body.ProvinceID, Latitude: input.Body.Latitude, Longitude: input.Body.Longitude,
		StartDate: input.Body.StartDate, EndDate: input.Body.EndDate, RegistrationOpen: input.Body.RegistrationOpen, RegistrationClose: input.Body.RegistrationClose,
		Status: input.Body.Status, IsReusable: input.Body.IsReusable, BannerURL: input.Body.BannerURL, PrizePool: input.Body.PrizePool,
		TotalSlots: input.Body.TotalSlots, SubEvents: subEvents,
	}
}

func toCreatedEventData(detail entities.EventDetail) response.EventData {
	data := response.EventData{
		ID: detail.Event.ID, OrganizerID: detail.Event.OrganizerID, Title: detail.Event.Title, Slug: detail.Event.Slug,
		Description: nil, Venue: detail.Event.Venue, Address: nil, City: detail.Event.City, Province: nil,
		CityID: detail.Event.CityID, ProvinceID: detail.Event.ProvinceID, Latitude: detail.Event.Latitude, Longitude: detail.Event.Longitude,
		Status: detail.Event.Status, IsReusable: false, BannerURL: detail.Event.BannerURL, TotalSlots: 0,
		SubEvents: make([]response.SubEventData, 0, len(detail.SubEvents)),
	}
	if detail.Event.Province != "" {
		province := detail.Event.Province
		data.Province = &province
	}
	if detail.Event.StartDate != nil {
		v := detail.Event.StartDate.Format(dateLayout)
		data.StartDate = &v
	}
	if detail.Event.EndDate != nil {
		v := detail.Event.EndDate.Format(dateLayout)
		data.EndDate = &v
	}
	for _, s := range detail.SubEvents {
		item := response.SubEventData{
			ID: s.ID, EventID: s.EventID, BirdCategoryID: s.BirdCategoryID, Title: s.Title, Description: s.Description,
			Session: s.Session, GantanganCount: s.GantanganCount, MaxSlots: s.MaxSlots, Price: s.Price,
			Winner1Prize: s.Winner1Prize, Winner2Prize: s.Winner2Prize, Winner3Prize: s.Winner3Prize,
			Status: s.Status, SortOrder: s.SortOrder,
		}
		if s.StartTime != nil {
			v := s.StartTime.Format(time.RFC3339)
			item.StartTime = &v
		}
		if s.EndTime != nil {
			v := s.EndTime.Format(time.RFC3339)
			item.EndTime = &v
		}
		data.SubEvents = append(data.SubEvents, item)
	}
	return data
}

func toStreamingData(s entities.EventStreaming) response.EventStreamingData {
	data := response.EventStreamingData{
		ID: s.ID, EventID: s.EventID, StartTime: s.StartTime.Format(time.RFC3339),
		TotalViews: s.TotalViews, Juara: s.Juara, LinkStream: s.LinkStream,
	}
	if s.EndTime != nil {
		v := s.EndTime.Format(time.RFC3339)
		data.EndTime = &v
	}
	return data
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, apperrors.ErrInvalidRadius),
		errors.Is(err, apperrors.ErrInvalidPayload),
		errors.Is(err, apperrors.ErrInvalidCategory),
		errors.Is(err, apperrors.ErrEventNotStreamable):
		return apphttp.UnprocessableEntity("Please check your input data")
	case errors.Is(err, apperrors.ErrNotFound), errors.Is(err, apperrors.ErrOrganizerNotFound):
		return apphttp.NotFound("Resource not found")
	case errors.Is(err, apperrors.ErrForbidden):
		return apphttp.Forbidden("Organizer access required")
	case errors.Is(err, apperrors.ErrIdentifierTaken),
		errors.Is(err, apperrors.ErrStreamingAlreadyActive),
		errors.Is(err, apperrors.ErrStreamingNotActive):
		return apphttp.Conflict("Event data already exists")
	default:
		return apphttp.InternalServerError("An error occurred. Please try again.")
	}
}

// formatIndonesianDate memformat tanggal ke bentuk Indonesia, contoh: Minggu, 18 Mei 2025.
func formatIndonesianDate(t time.Time) string {
	days := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	months := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	return fmt.Sprintf("%s, %d %s %d", days[t.Weekday()], t.Day(), months[int(t.Month())-1], t.Year())
}
