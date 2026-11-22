// Package server menyiapkan HTTP server, routing, dan middleware aplikasi.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Junx27/kicaugo-backend/internal/config"
	"github.com/Junx27/kicaugo-backend/internal/features/auth"
	"github.com/Junx27/kicaugo-backend/internal/features/birdcategory"
	"github.com/Junx27/kicaugo-backend/internal/features/province"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
	// scaffold:imports
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 60 * time.Second
	requestTimeout    = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
)

// Dependencies berisi konfigurasi dan service yang dibutuhkan server.
type Dependencies struct {
	Config   *config.Config
	Tokens   *jwt.Manager
	Province province.Service
	// scaffold:deps
	Auth           auth.AuthService
	BirdCategories birdcategory.Service
}

// Run menyiapkan router, mendaftarkan route, lalu menjalankan HTTP server.
func Run(ctx context.Context, deps Dependencies) error {
	configureHumaError()

	router := newRouter(deps.Config)
	api := humachi.New(router, newHumaConfig(deps.Config))
	auth.RegisterRoutes(api, deps.Auth, deps.Tokens)
	birdcategory.RegisterRoutes(api, deps.BirdCategories, deps.Tokens)
	province.RegisterRoutes(api, deps.Province, deps.Tokens)
	// scaffold:routes
	router.Get("/health", healthHandler)

	httpServer := &http.Server{
		Addr:              ":" + deps.Config.Port,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	go shutdownOnSignal(ctx, httpServer)

	slog.Info("server ready", "port", deps.Config.Port)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func newRouter(cfg *config.Config) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(requestTimeout))

	if cfg.DocsEnabled {
		router.Use(docsAuthMiddleware(cfg))
	}
	return router
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"healthy"}`))
}

func docsAuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/docs" || r.URL.Path == "/openapi.json" {
				user, pass, ok := r.BasicAuth()
				if !ok || user != cfg.DocsUser || pass != cfg.DocsPass {
					w.Header().Set("WWW-Authenticate", `Basic realm="API Documentation"`)
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte(`{"status":401,"message":"Authentication required","data":null}`))
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func newHumaConfig(cfg *config.Config) huma.Config {
	humaConfig := huma.DefaultConfig("KicauGo API", "1.0.0")
	humaConfig.CreateHooks = nil
	humaConfig.Transformers = nil
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearerAuth": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "JWT Bearer token (dapatkan dari /api/v1/auth/login)",
		},
	}
	if !cfg.DocsEnabled {
		humaConfig.DocsPath = ""
		humaConfig.OpenAPIPath = ""
	}
	return humaConfig
}

func configureHumaError() {
	huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
		details := make([]*huma.ErrorDetail, len(errs))
		for i := range errs {
			if converted, ok := errs[i].(huma.ErrorDetailer); ok {
				details[i] = converted.ErrorDetail()
			} else {
				if errs[i] == nil {
					continue
				}
				details[i] = &huma.ErrorDetail{Message: errs[i].Error()}
			}
		}
		return &apperrors.StandardError{
			Status:  status,
			Message: msg,
			Data:    nil,
			Errors:  details,
		}
	}
}

// shutdownOnSignal menunggu sinyal berhenti, lalu menutup server dengan anggun.
func shutdownOnSignal(ctx context.Context, httpServer *http.Server) {
	<-ctx.Done()
	shutdownCtx, stop := context.WithTimeout(context.Background(), shutdownTimeout)
	defer stop()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown server", "error", err)
	}
}
