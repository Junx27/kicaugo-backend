// Command api menjalankan server KicauGo.
package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Junx27/kicaugo-backend/internal/auth"
	"github.com/Junx27/kicaugo-backend/internal/config"
	"github.com/Junx27/kicaugo-backend/internal/docs"
	"github.com/Junx27/kicaugo-backend/internal/home"
	"github.com/Junx27/kicaugo-backend/internal/migrations"
	"github.com/Junx27/kicaugo-backend/internal/response"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 60 * time.Second
	pingTimeout       = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
	migrationTimeout  = 30 * time.Second
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	runServer(&cfg)
}

func runServer(cfg *config.Config) {
	// Override Huma's default error format to match {status, message, data}
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
		return &response.StandardError{
			Status:  status,
			Message: msg,
			Data:    nil,
			Errors:  details,
		}
	}
	maxConns, minConns, maxConnLife, maxConnIdle, healthCheck, tokenTTL := cfg.ParsedPoolConfig()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("parse database url: %v", err)
	}
	poolConfig.MaxConns = maxConns
	poolConfig.MinConns = minConns
	poolConfig.MaxConnLifetime = maxConnLife
	poolConfig.MaxConnIdleTime = maxConnIdle
	poolConfig.HealthCheckPeriod = healthCheck

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("buat connection pool: %v", err)
	}
	defer pool.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pingCtx, pingCancel := context.WithTimeout(ctx, pingTimeout)
	defer pingCancel()
	if err := pool.Ping(pingCtx); err != nil {
		log.Fatalf("database connection: %v", err)
	}

	// Run migrations on startup
	if err := runMigrations(pool); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	tokens := auth.NewTokenManager([]byte(cfg.JWTSecret), tokenTTL)
	authService := auth.NewService(auth.NewSQLRepository(pool), tokens)
	homeService := home.NewService(home.NewSQLRepository(pool))

	// Create Chi router
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	// Docs login handler (if enabled)
	if cfg.DocsEnabled {
		loginHandler := docs.NewLoginHandler(cfg.DocsUser, cfg.DocsPass)

		// Cookie-based auth middleware for /docs & /openapi.json (MUST be before routes)
		router.Use(docs.DocsAuthMiddleware)

		// Register login/logout routes (outside Huma, after middleware)
		router.Get("/docs/login", loginHandler.ServeLogin)
		router.Post("/docs/login", loginHandler.ServeLogin)
		router.Get("/docs/logout", loginHandler.ServeLogout)
	}

	// Create Huma API config
	humaConfig := huma.DefaultConfig("KicauGo API", "1.0.0")

	// Remove the default link transformer that adds $schema field
	humaConfig.CreateHooks = nil
	humaConfig.Transformers = nil

	// Security scheme: Bearer JWT
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearerAuth": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "JWT Bearer token (dapatkan dari /api/v1/auth/login)",
		},
	}

	// Docs configuration: disable if DOCS_ENABLED=false
	if !cfg.DocsEnabled {
		humaConfig.DocsPath = ""
		humaConfig.OpenAPIPath = ""
	}

	// Create Huma API
	api := humachi.New(router, humaConfig)

	// Register routes
	auth.RegisterRoutes(api, authService, tokens)
	home.RegisterRoutes(api, homeService, tokens)

	// Health endpoint (outside Huma for simplicity)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	go shutdownOnSignal(ctx, httpServer)

	slog.Info("server ready", "port", cfg.Port)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

// runMigrations mengeksekusi file SQL di migrations/ secara berurutan.
func runMigrations(pool *pgxpool.Pool) error {
	entries, err := migrations.MigrationFiles.ReadDir(".")
	if err != nil {
		return err
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, f := range files {
		content, err := migrations.MigrationFiles.ReadFile(f)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), migrationTimeout)
		defer cancel()
		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return err
		}
		slog.Info("migration applied", "file", f)
	}
	return nil
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
