package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
	"github.com/Junx27/kicaugo-backend/internal/utils/response"
)

type fakeAuthService struct {
	result     entities.AuthResult
	err        error
	profile    entities.User
	profileErr error
}

func (f *fakeAuthService) Register(ctx context.Context, name, identifier, password string) (entities.AuthResult, error) {
	return f.result, f.err
}

func (f *fakeAuthService) Login(ctx context.Context, identifier, password string) (entities.AuthResult, error) {
	return f.result, f.err
}

func (f *fakeAuthService) Profile(ctx context.Context, identifier string) (entities.User, error) {
	return f.profile, f.profileErr
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		svc    *fakeAuthService
		status int
	}{
		{
			name:   "berhasil",
			body:   `{"name":"Kicau Mania","identifier":"test@gmail.com","password":"test123@"}`,
			svc:    &fakeAuthService{result: entities.AuthResult{User: entities.User{Name: "Kicau Mania", Identifier: "test@gmail.com"}, Token: "jwt"}},
			status: http.StatusCreated,
		},
		{
			name:   "payload invalid (missing fields)",
			body:   `{"name":"Kicau Mania"}`,
			svc:    &fakeAuthService{},
			status: http.StatusUnprocessableEntity,
		},
		{
			name:   "json rusak",
			body:   `{`,
			svc:    &fakeAuthService{},
			status: http.StatusBadRequest,
		},
		{
			name:   "identifier dipakai",
			body:   `{"name":"Kicau Mania","identifier":"test@gmail.com","password":"test123@"}`,
			svc:    &fakeAuthService{err: apperrors.ErrIdentifierTaken},
			status: http.StatusConflict,
		},
		{
			name:   "kesalahan server",
			body:   `{"name":"Kicau Mania","identifier":"test@gmail.com","password":"test123@"}`,
			svc:    &fakeAuthService{err: errors.New("db down")},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, api := humatest.New(t)
			tokens := jwt.NewManager([]byte("rahasia-untuk-test"), 24*3600*1000000000)
			RegisterRoutes(api, tc.svc, tokens)

			resp := api.Post("/api/v1/auth/register", strings.NewReader(tc.body))
			if resp.Code != tc.status {
				t.Fatalf("status = %d, ingin %d; body = %s", resp.Code, tc.status, resp.Body.String())
			}
			if tc.status == http.StatusCreated {
				var body struct {
					Status  int    `json:"status"`
					Message string `json:"message"`
					Data    struct {
						User  response.UserData `json:"user"`
						Token string            `json:"token"`
					} `json:"data"`
				}
				if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
					t.Fatalf("body bukan JSON: %v", err)
				}
				if body.Status != http.StatusCreated {
					t.Errorf("status = %d, ingin %d", body.Status, http.StatusCreated)
				}
				if body.Data.Token != "jwt" || body.Data.User.Identifier != "test@gmail.com" {
					t.Errorf("response = %+v", body)
				}
			}
		})
	}
}

func TestLogin(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		_, api := humatest.New(t)
		svc := &fakeAuthService{result: entities.AuthResult{User: entities.User{Name: "Kicau Mania", Identifier: "test@gmail.com"}, Token: "jwt"}}
		tokens := jwt.NewManager([]byte("rahasia-untuk-test"), 24*3600*1000000000)
		RegisterRoutes(api, svc, tokens)

		resp := api.Post("/api/v1/auth/login", strings.NewReader(`{"identifier":"test@gmail.com","password":"test123@"}`))
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, ingin 200", resp.Code)
		}
	})

	t.Run("kredensial salah", func(t *testing.T) {
		_, api := humatest.New(t)
		svc := &fakeAuthService{err: apperrors.ErrInvalidCredentials}
		tokens := jwt.NewManager([]byte("rahasia-untuk-test"), 24*3600*1000000000)
		RegisterRoutes(api, svc, tokens)

		resp := api.Post("/api/v1/auth/login", strings.NewReader(`{"identifier":"test@gmail.com","password":"salah"}`))
		if resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})

	t.Run("json rusak", func(t *testing.T) {
		_, api := humatest.New(t)
		svc := &fakeAuthService{}
		tokens := jwt.NewManager([]byte("rahasia-untuk-test"), 24*3600*1000000000)
		RegisterRoutes(api, svc, tokens)

		resp := api.Post("/api/v1/auth/login", strings.NewReader(`{`))
		if resp.Code != http.StatusBadRequest {
			t.Errorf("status = %d, ingin 400", resp.Code)
		}
	})
}

func TestLogout(t *testing.T) {
	_, api := humatest.New(t)
	svc := &fakeAuthService{}
	tokens := jwt.NewManager([]byte("rahasia-untuk-test"), 24*3600*1000000000)
	RegisterRoutes(api, svc, tokens)

	token, _ := tokens.Issue("test@gmail.com", "Kicau Mania")
	resp := api.Post("/api/v1/auth/logout", "Authorization: Bearer "+token)
	if resp.Code != http.StatusNoContent {
		t.Errorf("status = %d, ingin 204; body=%s", resp.Code, resp.Body.String())
	}
}

func TestMe(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		_, api := humatest.New(t)
		svc := &fakeAuthService{profile: entities.User{Name: "Kicau Mania", Identifier: "test@gmail.com"}}
		tokens := jwt.NewManager([]byte("rahasia-untuk-test"), 24*3600*1000000000)
		RegisterRoutes(api, svc, tokens)

		token, _ := tokens.Issue("test@gmail.com", "Kicau Mania")
		resp := api.Get("/api/v1/auth/me", "Authorization: Bearer "+token)
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, ingin 200; body=%s", resp.Code, resp.Body.String())
		}
		var body struct {
			Status  int               `json:"status"`
			Message string            `json:"message"`
			Data    response.UserData `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Status != http.StatusOK {
			t.Errorf("status = %d, ingin %d", body.Status, http.StatusOK)
		}
		if body.Data.Name != "Kicau Mania" {
			t.Errorf("response = %+v", body)
		}
	})

	t.Run("tanpa token", func(t *testing.T) {
		_, api := humatest.New(t)
		svc := &fakeAuthService{}
		tokens := jwt.NewManager([]byte("rahasia-untuk-test"), 24*3600*1000000000)
		RegisterRoutes(api, svc, tokens)

		resp := api.Get("/api/v1/auth/me")
		if resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})

	t.Run("profil gagal", func(t *testing.T) {
		_, api := humatest.New(t)
		svc := &fakeAuthService{profileErr: errors.New("db down")}
		tokens := jwt.NewManager([]byte("rahasia-untuk-test"), 24*3600*1000000000)
		RegisterRoutes(api, svc, tokens)

		token, _ := tokens.Issue("test@gmail.com", "Kicau Mania")
		resp := api.Get("/api/v1/auth/me", "Authorization: Bearer "+token)
		if resp.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, ingin 500", resp.Code)
		}
	})
}
