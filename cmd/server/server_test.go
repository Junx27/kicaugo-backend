package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/Junx27/kicaugo-backend/internal/config"
	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
)

func TestShutdownOnSignal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	httpServer := &http.Server{}
	done := make(chan struct{})

	go func() {
		shutdownOnSignal(ctx, httpServer)
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("shutdownOnSignal() tidak selesai setelah konteks dibatalkan")
	}
}

func TestHealthHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	healthHandler(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, ingin %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, ingin application/json", got)
	}
	if got := rec.Body.String(); got != `{"status":"healthy"}` {
		t.Errorf("body = %q", got)
	}
}

func TestDocsAuthMiddleware(t *testing.T) {
	cfg := &config.Config{DocsUser: "admin", DocsPass: "secret"}
	tests := []struct {
		name string
		path string
		user string
		pass string
		want int
	}{
		{name: "docs tanpa auth", path: "/docs", want: http.StatusUnauthorized},
		{name: "docs auth salah", path: "/docs", user: "admin", pass: "salah", want: http.StatusUnauthorized},
		{name: "docs auth benar", path: "/docs", user: "admin", pass: "secret", want: http.StatusNoContent},
		{name: "openapi auth benar", path: "/openapi.json", user: "admin", pass: "secret", want: http.StatusNoContent},
		{name: "path lain tanpa auth", path: "/health", want: http.StatusNoContent},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := docsAuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.user != "" || tc.pass != "" {
				req.SetBasicAuth(tc.user, tc.pass)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.want {
				t.Errorf("status = %d, ingin %d", rec.Code, tc.want)
			}
		})
	}
}

func TestNewHumaConfig(t *testing.T) {
	t.Run("docs aktif", func(t *testing.T) {
		got := newHumaConfig(&config.Config{DocsEnabled: true})
		if got.DocsPath == "" || got.OpenAPIPath == "" {
			t.Errorf("DocsPath = %q, OpenAPIPath = %q; ingin keduanya terisi", got.DocsPath, got.OpenAPIPath)
		}
		if _, ok := got.Components.SecuritySchemes["bearerAuth"]; !ok {
			t.Error("security scheme bearerAuth tidak terdaftar")
		}
	})

	t.Run("docs nonaktif", func(t *testing.T) {
		got := newHumaConfig(&config.Config{DocsEnabled: false})
		if got.DocsPath != "" || got.OpenAPIPath != "" {
			t.Errorf("DocsPath = %q, OpenAPIPath = %q; ingin keduanya kosong", got.DocsPath, got.OpenAPIPath)
		}
	})
}

func TestConfigureHumaError(t *testing.T) {
	original := huma.NewError
	t.Cleanup(func() { huma.NewError = original })

	configureHumaError()
	got := huma.NewError(http.StatusBadRequest, "invalid",
		&huma.ErrorDetail{Message: "detail"},
		errors.New("plain"),
		nil,
	)
	stdErr, ok := got.(*apperrors.StandardError)
	if !ok {
		t.Fatalf("NewError() = %T, ingin *apperrors.StandardError", got)
	}
	if stdErr.Status != http.StatusBadRequest || stdErr.Message != "invalid" {
		t.Errorf("StandardError = %+v", stdErr)
	}
	if len(stdErr.Errors) != 3 {
		t.Fatalf("Errors = %+v, ingin 3 entri", stdErr.Errors)
	}
	if stdErr.Errors[0].Message != "detail" || stdErr.Errors[1].Message != "plain" {
		t.Errorf("Errors = %+v", stdErr.Errors)
	}
	if stdErr.Errors[2] != nil {
		t.Errorf("Errors[2] = %+v, ingin nil", stdErr.Errors[2])
	}
}

func TestNewRouter(t *testing.T) {
	t.Run("docs nonaktif", func(t *testing.T) {
		router := newRouter(&config.Config{DocsEnabled: false})
		router.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))
		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, ingin %d", rec.Code, http.StatusNoContent)
		}
	})

	t.Run("docs aktif melindungi docs", func(t *testing.T) {
		router := newRouter(&config.Config{DocsEnabled: true, DocsUser: "admin", DocsPass: "secret"})
		router.Get("/docs", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/docs", nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin %d", rec.Code, http.StatusUnauthorized)
		}
	})
}

func TestRunWithCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := &config.Config{Port: "0"}
	deps := Dependencies{
		Config: cfg,
		Tokens: jwt.NewManager([]byte("rahasia-untuk-test"), time.Hour),
		Auth:   fakeAuthService{},
	}
	if err := Run(ctx, deps); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunListenError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &config.Config{Port: "invalid"}
	deps := Dependencies{
		Config: cfg,
		Tokens: jwt.NewManager([]byte("rahasia-untuk-test"), time.Hour),
		Auth:   fakeAuthService{},
	}
	if err := Run(ctx, deps); err == nil {
		t.Error("Run() ingin error saat port tidak valid")
	}
}

type fakeAuthService struct{}

func (fakeAuthService) Register(context.Context, string, string, string) (entities.AuthResult, error) {
	return entities.AuthResult{}, nil
}

func (fakeAuthService) Login(context.Context, string, string) (entities.AuthResult, error) {
	return entities.AuthResult{}, nil
}

func (fakeAuthService) Profile(context.Context, string) (entities.User, error) {
	return entities.User{}, nil
}
