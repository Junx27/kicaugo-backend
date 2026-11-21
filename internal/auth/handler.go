package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/Junx27/kicaugo-backend/internal/response"
	"github.com/danielgtaylor/huma/v2"
)

// RegisterInput input untuk registrasi
type RegisterInput struct {
	Body struct {
		Name       string `json:"name" required:"true" minLength:"3" maxLength:"100" doc:"Nama pengguna"`
		Identifier string `json:"identifier" required:"true" format:"email" doc:"Email atau nomor WhatsApp"`
		Password   string `json:"password" required:"true" minLength:"8" maxLength:"72" doc:"Kata sandi"`
	}
}

// AuthData data untuk response auth (register & login)
type AuthData struct {
	User  UserPayload `json:"user"`
	Token string      `json:"token"`
}

// RegisterOutput output untuk registrasi
type RegisterOutput struct {
	Status int `status:"201"`
	Body   response.Success[AuthData]
}

// LoginInput input untuk login
type LoginInput struct {
	Body struct {
		Identifier string `json:"identifier" required:"true" doc:"Email atau nomor WhatsApp"`
		Password   string `json:"password" required:"true" doc:"Kata sandi"`
	}
}

// LoginOutput output untuk login
type LoginOutput struct {
	Status int `status:"200"`
	Body   response.Success[AuthData]
}

// UserPayload representasi user di response
type UserPayload struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

// MeOutput output untuk GET /me
type MeOutput struct {
	Status int `status:"200"`
	Body   response.Success[UserPayload]
}

// LogoutOutput output untuk logout
type LogoutOutput struct {
	Status int `status:"204"`
}

// RegisterRoutes mendaftarkan route auth ke Huma API
func RegisterRoutes(api huma.API, service AuthService, tokens *TokenManager) {
	// POST /api/v1/auth/register
	huma.Register(api, huma.Operation{
		OperationID: "register",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/register",
		Summary:     "Register",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, input *RegisterInput) (*RegisterOutput, error) {
		result, err := service.Register(ctx, input.Body.Name, input.Body.Identifier, input.Body.Password)
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &RegisterOutput{}
		resp.Status = http.StatusCreated
		resp.Body = response.NewSuccess(http.StatusCreated, "Registration successful", AuthData{
			User:  toPayload(result.User),
			Token: result.Token,
		})
		return resp, nil
	})

	// POST /api/v1/auth/login
	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/login",
		Summary:     "Login",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
		result, err := service.Login(ctx, input.Body.Identifier, input.Body.Password)
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &LoginOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Login successful", AuthData{
			User:  toPayload(result.User),
			Token: result.Token,
		})
		return resp, nil
	})

	// POST /api/v1/auth/logout
	huma.Register(api, huma.Operation{
		OperationID: "logout",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/logout",
		Summary:     "Logout",
		Tags:        []string{"auth"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: []func(ctx huma.Context, next func(huma.Context)){AuthMiddleware(tokens)},
	}, func(ctx context.Context, input *struct{}) (*LogoutOutput, error) {
		// Stateless logout - token dihapus di client
		resp := &LogoutOutput{}
		resp.Status = http.StatusNoContent
		return resp, nil
	})

	// GET /api/v1/auth/me
	huma.Register(api, huma.Operation{
		OperationID: "me",
		Method:      http.MethodGet,
		Path:        "/api/v1/auth/me",
		Summary:     "Get Profile",
		Tags:        []string{"auth"},
		Security:    []map[string][]string{{"bearerAuth": {}}},
		Middlewares: []func(ctx huma.Context, next func(huma.Context)){AuthMiddleware(tokens)},
	}, func(ctx context.Context, input *struct{}) (*MeOutput, error) {
		user, ok := UserFromContext(ctx)
		if !ok {
			return nil, huma.Error401Unauthorized("Authentication required")
		}
		result, err := service.Profile(ctx, user.Identifier)
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &MeOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Profile loaded successfully", toPayload(result))
		return resp, nil
	})
}

func toPayload(user User) UserPayload {
	return UserPayload{Name: user.Name, Identifier: user.Identifier}
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		return huma.Error401Unauthorized("Invalid email/WhatsApp or password")
	case errors.Is(err, ErrIdentifierTaken):
		return huma.Error409Conflict("Email or WhatsApp number already registered")
	case errors.Is(err, ErrInvalidPayload):
		return huma.Error400BadRequest("Please check your input data")
	case errors.Is(err, ErrNotFound):
		return huma.Error404NotFound("User not found")
	default:
		return huma.Error500InternalServerError("An error occurred. Please try again.")
	}
}

// AuthService adalah bagian Service yang dibutuhkan handler.
type AuthService interface {
	Register(ctx context.Context, name, identifier, password string) (Result, error)
	Login(ctx context.Context, identifier, password string) (Result, error)
	Profile(ctx context.Context, identifier string) (User, error)
}
