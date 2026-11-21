package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
)

func TestAuthMiddleware(t *testing.T) {
	tokens := NewTokenManager([]byte("rahasia-untuk-test"), 24*3600*1000000000)
	raw, err := tokens.Issue(User{Name: "Kicau Mania", Identifier: "test@gmail.com"})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		header string
		status int
	}{
		{name: "token valid", header: "Bearer " + raw, status: http.StatusNoContent},
		{name: "tanpa header", status: http.StatusUnauthorized},
		{name: "bearer kosong", header: "Bearer ", status: http.StatusUnauthorized},
		{name: "skema salah", header: "Basic " + raw, status: http.StatusUnauthorized},
		{name: "token invalid", header: "Bearer bukan.jwt.valid", status: http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, api := humatest.New(t)
			api.UseMiddleware(AuthMiddleware(tokens))
			huma.Get(api, "/test", func(ctx context.Context, input *struct{}) (*struct{}, error) {
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
	if _, ok := UserFromContext(context.Background()); ok {
		t.Error("UserFromContext() ingin false untuk konteks kosong")
	}
}
