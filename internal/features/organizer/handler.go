package organizer

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

func toOrganizerData(o entities.Organizer) response.OrganizerData {
	return response.OrganizerData{
		ID: o.ID, Name: o.Name, Slug: o.Slug, Email: o.Email,
		Phone: o.Phone, LogoURL: o.LogoURL, Description: o.Description, IsVerified: o.IsVerified,
	}
}

func toMemberData(m entities.EventOrganizerUser) response.OrganizerMemberData {
	return response.OrganizerMemberData{
		ID: m.ID, OrganizerID: m.OrganizerID, UserID: m.UserID, Role: m.Role, IsActive: m.IsActive,
	}
}

// RegisterRoutes mendaftarkan route organizer ke Huma API.
func RegisterRoutes(api apphttp.API, service Service, tokens *jwt.Manager) {
	apphttp.Register(api, apphttp.Operation{
		ID:      "createOrganizer",
		Method:  http.MethodPost,
		Path:    "/api/v1/organizers",
		Summary: "Create Organizer",
		Tags:    []string{"organizer"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: []apphttp.Middleware{middlewares.AuthMiddleware[jwt.Claims](tokens)},
	}, func(ctx context.Context, input *request.CreateOrganizerInput) (*response.CreateOrganizerOutput, error) {
		identifier, err := authctx.IdentifierFromContext(ctx)
		if err != nil {
			return nil, err
		}
		created, err := service.Create(ctx, identifier, entities.Organizer{
			Name: input.Body.Name, Slug: input.Body.Slug, Email: input.Body.Email,
			Phone: input.Body.Phone, LogoURL: input.Body.LogoURL, Description: input.Body.Description,
		})
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.CreateOrganizerOutput{Status: http.StatusCreated}
		resp.Body = response.NewSuccess(http.StatusCreated, "Organizer created", toOrganizerData(created))
		return resp, nil
	})

	apphttp.Register(api, apphttp.Operation{
		ID:      "registerOrganizerMember",
		Method:  http.MethodPost,
		Path:    "/api/v1/organizers/{id}/members",
		Summary: "Register Organizer Member",
		Tags:    []string{"organizer"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: []apphttp.Middleware{middlewares.AuthMiddleware[jwt.Claims](tokens)},
	}, func(ctx context.Context, input *request.RegisterOrganizerMemberInput) (*response.RegisterOrganizerMemberOutput, error) {
		identifier, err := authctx.IdentifierFromContext(ctx)
		if err != nil {
			return nil, err
		}
		member, err := service.RegisterMember(ctx, identifier, input.ID, input.Body.Role)
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.RegisterOrganizerMemberOutput{Status: http.StatusCreated}
		resp.Body = response.NewSuccess(http.StatusCreated, "Organizer member registered", toMemberData(member))
		return resp, nil
	})
}

func mapServiceError(err error) error {
	if errors.Is(err, apperrors.ErrInvalidPayload) {
		return apphttp.UnprocessableEntity("Invalid payload")
	}
	if errors.Is(err, apperrors.ErrNotFound) {
		return apphttp.NotFound("User not found")
	}
	if errors.Is(err, apperrors.ErrOrganizerNotFound) {
		return apphttp.NotFound("Organizer not found")
	}
	if errors.Is(err, apperrors.ErrOrganizerSlugTaken) || errors.Is(err, apperrors.ErrOrganizerMemberTaken) || errors.Is(err, apperrors.ErrIdentifierTaken) {
		return apphttp.Conflict("Organizer data already exists")
	}
	if errors.Is(err, apperrors.ErrForbidden) {
		return apphttp.Forbidden("Organizer access required")
	}
	return apphttp.InternalServerError("An error occurred. Please try again.")
}
