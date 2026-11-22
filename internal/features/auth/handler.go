package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	"github.com/Junx27/kicaugo-backend/internal/middlewares"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	apphttp "github.com/Junx27/kicaugo-backend/internal/utils/http"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
	"github.com/Junx27/kicaugo-backend/internal/utils/request"
	"github.com/Junx27/kicaugo-backend/internal/utils/response"
)

// RegisterRoutes mendaftarkan route auth ke Huma API
func RegisterRoutes(api apphttp.API, service AuthService, tokens *jwt.Manager) {
	// POST /api/v1/auth/register
	apphttp.Register(api, apphttp.Operation{
		ID:      "register",
		Method:  http.MethodPost,
		Path:    "/api/v1/auth/register",
		Summary: "Register",
		Tags:    []string{"auth"},
	}, func(ctx context.Context, input *request.RegisterInput) (*response.RegisterOutput, error) {
		result, err := service.Register(ctx, input.Body.Name, input.Body.Identifier, input.Body.Password)
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.RegisterOutput{}
		resp.Status = http.StatusCreated
		resp.Body = response.NewSuccess(http.StatusCreated, "Registration successful", response.AuthData{
			User:  toPayload(result.User),
			Token: result.Token,
		})
		return resp, nil
	})

	// POST /api/v1/auth/login
	apphttp.Register(api, apphttp.Operation{
		ID:      "login",
		Method:  http.MethodPost,
		Path:    "/api/v1/auth/login",
		Summary: "Login",
		Tags:    []string{"auth"},
	}, func(ctx context.Context, input *request.LoginInput) (*response.LoginOutput, error) {
		result, err := service.Login(ctx, input.Body.Identifier, input.Body.Password)
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.LoginOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Login successful", response.AuthData{
			User:  toPayload(result.User),
			Token: result.Token,
		})
		return resp, nil
	})

	// POST /api/v1/auth/logout
	apphttp.Register(api, apphttp.Operation{
		ID:          "logout",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/logout",
		Summary:     "Logout",
		Tags:        []string{"auth"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: []apphttp.Middleware{middlewares.AuthMiddleware[jwt.Claims](tokens)},
	}, func(ctx context.Context, input *struct{}) (*response.LogoutOutput, error) {
		// Stateless logout - token dihapus di client
		resp := &response.LogoutOutput{}
		resp.Status = http.StatusNoContent
		return resp, nil
	})

	// GET /api/v1/auth/me
	apphttp.Register(api, apphttp.Operation{
		ID:          "me",
		Method:      http.MethodGet,
		Path:        "/api/v1/auth/me",
		Summary:     "Get Profile",
		Tags:        []string{"auth"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: []apphttp.Middleware{middlewares.AuthMiddleware[jwt.Claims](tokens)},
	}, func(ctx context.Context, input *struct{}) (*response.MeOutput, error) {
		claims, ok := middlewares.UserFromContext[jwt.Claims](ctx)
		if !ok {
			return nil, apphttp.Unauthorized("Authentication required")
		}
		result, err := service.Profile(ctx, claims.Subject)
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.MeOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Profile loaded successfully", toPayload(result))
		return resp, nil
	})
}

func toPayload(user entities.User) response.UserPayload {
	return response.UserPayload{Name: user.Name, Identifier: user.Identifier}
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, apperrors.ErrInvalidCredentials):
		return apphttp.Unauthorized("Invalid email/WhatsApp or password")
	case errors.Is(err, apperrors.ErrIdentifierTaken):
		return apphttp.Conflict("Email or WhatsApp number already registered")
	case errors.Is(err, apperrors.ErrInvalidPayload):
		return apphttp.BadRequest("Please check your input data")
	case errors.Is(err, apperrors.ErrNotFound):
		return apphttp.NotFound("User not found")
	default:
		return apphttp.InternalServerError("An error occurred. Please try again.")
	}
}

// AuthService adalah bagian Service yang dibutuhkan handler.
type AuthService interface {
	Register(ctx context.Context, name, identifier, password string) (entities.AuthResult, error)
	Login(ctx context.Context, identifier, password string) (entities.AuthResult, error)
	Profile(ctx context.Context, identifier string) (entities.User, error)
}
