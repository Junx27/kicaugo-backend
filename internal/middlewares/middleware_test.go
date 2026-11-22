package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
)

type fakeTokenParser struct {
	user string
	err  error
}

func (f fakeTokenParser) Parse(string) (string, error) {
	return f.user, f.err
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name   string
		header string
		parser fakeTokenParser
		status int
	}{
		{name: "token valid", header: "Bearer token", parser: fakeTokenParser{user: "test@gmail.com"}, status: http.StatusNoContent},
		{name: "tanpa header", status: http.StatusUnauthorized},
		{name: "bearer kosong", header: "Bearer ", status: http.StatusUnauthorized},
		{name: "skema salah", header: "Basic token", status: http.StatusUnauthorized},
		{name: "token invalid", header: "Bearer invalid", parser: fakeTokenParser{err: errors.New("invalid")}, status: http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, api := humatest.New(t)
			api.UseMiddleware(AuthMiddleware[string](tc.parser))
			huma.Get(api, "/test", func(ctx context.Context, input *struct{}) (*struct{}, error) {
				if _, ok := UserFromContext[string](ctx); !ok {
					t.Error("UserFromContext() tidak menemukan user")
				}
				return nil, nil
			})

			var resp *httptest.ResponseRecorder
			if tc.header != "" {
				resp = api.Get("/test", "Authorization: "+tc.header)
			} else {
				resp = api.Get("/test")
			}

			if resp.Code != tc.status {
				t.Errorf("status = %d, ingin %d", resp.Code, tc.status)
			}
		})
	}
}

func TestUserFromContextTanpaNilai(t *testing.T) {
	if _, ok := UserFromContext[string](context.Background()); ok {
		t.Error("UserFromContext() ingin false untuk konteks kosong")
	}
}

func TestUserFromContextTipeSalah(t *testing.T) {
	ctx := context.WithValue(context.Background(), userKey, 123)
	if _, ok := UserFromContext[string](ctx); ok {
		t.Error("UserFromContext() ingin false untuk tipe yang salah")
	}
}

type fakeOrganizerChecker struct {
	active bool
	err    error
}

func (f fakeOrganizerChecker) HasActiveMembership(ctx context.Context, identifier string) (bool, error) {
	return f.active, f.err
}

func TestOrganizerOnly(t *testing.T) {
	t.Run("aktif", func(t *testing.T) {
		_, api := humatest.New(t)
		api.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) {
			next(huma.WithValue(ctx, userKey, "user@example.com"))
		})
		api.UseMiddleware(OrganizerOnly[string](fakeOrganizerChecker{active: true}, func(s string) string { return s }))
		huma.Get(api, "/ok", func(context.Context, *struct{}) (*struct{}, error) { return nil, nil })
		if resp := api.Get("/ok"); resp.Code != http.StatusNoContent {
			t.Fatalf("status=%d", resp.Code)
		}
	})

	t.Run("bukan organizer", func(t *testing.T) {
		_, api := humatest.New(t)
		api.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) {
			next(huma.WithValue(ctx, userKey, "user@example.com"))
		})
		api.UseMiddleware(OrganizerOnly[string](fakeOrganizerChecker{active: false}, func(s string) string { return s }))
		huma.Get(api, "/ok", func(context.Context, *struct{}) (*struct{}, error) { return nil, nil })
		if resp := api.Get("/ok"); resp.Code != http.StatusForbidden {
			t.Fatalf("status=%d", resp.Code)
		}
	})

	t.Run("tanpa user", func(t *testing.T) {
		_, api := humatest.New(t)
		api.UseMiddleware(OrganizerOnly[string](fakeOrganizerChecker{active: true}, func(s string) string { return s }))
		huma.Get(api, "/ok", func(context.Context, *struct{}) (*struct{}, error) { return nil, nil })
		if resp := api.Get("/ok"); resp.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d", resp.Code)
		}
	})

	t.Run("subject kosong", func(t *testing.T) {
		_, api := humatest.New(t)
		api.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) { next(huma.WithValue(ctx, userKey, "x")) })
		api.UseMiddleware(OrganizerOnly[string](fakeOrganizerChecker{active: true}, func(string) string { return "" }))
		huma.Get(api, "/ok", func(context.Context, *struct{}) (*struct{}, error) { return nil, nil })
		if resp := api.Get("/ok"); resp.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d", resp.Code)
		}
	})

	t.Run("checker error", func(t *testing.T) {
		_, api := humatest.New(t)
		api.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) { next(huma.WithValue(ctx, userKey, "x")) })
		api.UseMiddleware(OrganizerOnly[string](fakeOrganizerChecker{err: errors.New("db")}, func(s string) string { return s }))
		huma.Get(api, "/ok", func(context.Context, *struct{}) (*struct{}, error) { return nil, nil })
		if resp := api.Get("/ok"); resp.Code != http.StatusForbidden {
			t.Fatalf("status=%d", resp.Code)
		}
	})
}
