package basemap

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2/humatest"

	"github.com/Junx27/kicaugo-backend/internal/middlewares"
	"github.com/Junx27/kicaugo-backend/internal/testutil"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
)

const (
	tileBasePath    = "/api/v1/basemap/tiles/"
	openAPITilePath = "/api/v1/basemap/tiles/{z}/{x}/{y}"
	tileOne         = "1/0/0"
	want400Status   = "status = %d, ingin 400"
)

type fakeTileService struct {
	data   []byte
	err    error
	ttl    time.Duration
	z      int
	x      int
	y      int
	retina bool
	style  string
}

func (f *fakeTileService) Get(_ context.Context, z, x, y int, retina bool, style string) ([]byte, error) {
	f.z = z
	f.x = x
	f.y = y
	f.retina = retina
	f.style = style
	return f.data, f.err
}

func (f *fakeTileService) CacheTTL() time.Duration {
	return f.ttl
}

func setupTileAPI(t *testing.T, svc Service, perMin int) (humatest.TestAPI, *jwt.Manager) {
	t.Helper()
	_, api := humatest.New(t)
	tokens := testutil.NewTokenManager()
	limiter := middlewares.NewRateLimiter(perMin).Middleware(api)
	RegisterRoutes(api, svc, tokens, limiter)
	return api, tokens
}

func TestGetBasemapTileHandler(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		svc := &fakeTileService{data: []byte("\x89PNG\r\n"), ttl: time.Hour}
		api, tokens := setupTileAPI(t, svc, 1000)

		resp := api.Get(tileBasePath+"1/0/0?retina=true", testutil.BearerHeader(t, tokens))
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
		}
		if resp.Body.String() != "\x89PNG\r\n" {
			t.Errorf("body = %q", resp.Body.String())
		}
		if got := resp.Header().Get("Content-Type"); got != tileContentType {
			t.Errorf("Content-Type = %q, ingin %q", got, tileContentType)
		}
		if got := resp.Header().Get("Cache-Control"); got != "private, max-age=3600" {
			t.Errorf("Cache-Control = %q", got)
		}
		if svc.z != 1 || svc.x != 0 || svc.y != 0 || !svc.retina || svc.style != defaultStyle {
			t.Errorf("argumen service = %+v", svc)
		}
	})

	t.Run("tanpa token", func(t *testing.T) {
		api, _ := setupTileAPI(t, &fakeTileService{}, 1000)
		if resp := api.Get(tileBasePath + tileOne); resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})

	t.Run("x di luar batas", func(t *testing.T) {
		api, tokens := setupTileAPI(t, &fakeTileService{}, 1000)
		if resp := api.Get(tileBasePath+"1/2/0", testutil.BearerHeader(t, tokens)); resp.Code != http.StatusBadRequest {
			t.Errorf(want400Status, resp.Code)
		}
	})

	t.Run("y di luar batas", func(t *testing.T) {
		api, tokens := setupTileAPI(t, &fakeTileService{}, 1000)
		if resp := api.Get(tileBasePath+"1/0/2", testutil.BearerHeader(t, tokens)); resp.Code != http.StatusBadRequest {
			t.Errorf(want400Status, resp.Code)
		}
	})

	t.Run("y tidak kurang dari 2^z", func(t *testing.T) {
		api, tokens := setupTileAPI(t, &fakeTileService{}, 1000)
		if resp := api.Get(tileBasePath+"2/0/4", testutil.BearerHeader(t, tokens)); resp.Code != http.StatusBadRequest {
			t.Errorf(want400Status, resp.Code)
		}
	})

	t.Run("zoom di luar tag Huma", func(t *testing.T) {
		api, tokens := setupTileAPI(t, &fakeTileService{}, 1000)
		if resp := api.Get(tileBasePath+"21/0/0", testutil.BearerHeader(t, tokens)); resp.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, ingin 422", resp.Code)
		}
	})

	t.Run("style tidak dikenal", func(t *testing.T) {
		api, tokens := setupTileAPI(t, &fakeTileService{err: ErrUnsupportedStyle}, 1000)
		if resp := api.Get(tileBasePath+"1/0/0?style=dark", testutil.BearerHeader(t, tokens)); resp.Code != http.StatusBadRequest {
			t.Errorf(want400Status, resp.Code)
		}
	})

	t.Run("key kosong", func(t *testing.T) {
		api, tokens := setupTileAPI(t, &fakeTileService{err: ErrNotConfigured}, 1000)
		if resp := api.Get(tileBasePath+tileOne, testutil.BearerHeader(t, tokens)); resp.Code != http.StatusServiceUnavailable {
			t.Errorf("status = %d, ingin 503", resp.Code)
		}
	})

	t.Run("upstream error", func(t *testing.T) {
		api, tokens := setupTileAPI(t, &fakeTileService{err: errors.New("boom")}, 1000)
		if resp := api.Get(tileBasePath+tileOne, testutil.BearerHeader(t, tokens)); resp.Code != http.StatusBadGateway {
			t.Errorf("status = %d, ingin 502", resp.Code)
		}
	})

	t.Run("rate limit", func(t *testing.T) {
		svc := &fakeTileService{data: []byte("png"), ttl: time.Hour}
		api, tokens := setupTileAPI(t, svc, 1)
		header := testutil.BearerHeader(t, tokens)

		if resp := api.Get(tileBasePath+tileOne, header); resp.Code != http.StatusOK {
			t.Fatalf("request pertama status = %d", resp.Code)
		}
		resp := api.Get(tileBasePath+tileOne, header)
		if resp.Code != http.StatusTooManyRequests {
			t.Errorf("status = %d, ingin 429", resp.Code)
		}
		if resp.Header().Get("Retry-After") == "" {
			t.Error("Retry-After harus diisi")
		}
	})
}

func TestGetBasemapTileHandlerTanpaIdentitas(t *testing.T) {
	_, err := newGetTileHandler(&fakeTileService{})(context.Background(), &TileInput{})
	statusErr, ok := err.(interface{ GetStatus() int })
	if err == nil || !ok || statusErr.GetStatus() != http.StatusUnauthorized {
		t.Errorf("handler tanpa identitas = %v, ingin 401", err)
	}
}

func TestValidateTile(t *testing.T) {
	tests := []struct {
		name string
		z    int
		x    int
		y    int
		ok   bool
	}{
		{name: "valid zoom 0", z: 0, x: 0, y: 0, ok: true},
		{name: "valid zoom maksimum", z: maxZoom, x: (1 << maxZoom) - 1, y: (1 << maxZoom) - 1, ok: true},
		{name: "zoom negatif", z: -1, ok: false},
		{name: "zoom terlalu besar", z: maxZoom + 1, ok: false},
		{name: "x negatif", z: 1, x: -1, ok: false},
		{name: "y negatif", z: 1, y: -1, ok: false},
		{name: "x di luar batas", z: 1, x: 2, ok: false},
		{name: "y di luar batas", z: 1, y: 2, ok: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTile(tc.z, tc.x, tc.y)
			if tc.ok && err != nil {
				t.Errorf("validateTile() error = %v", err)
			}
			if !tc.ok && !errors.Is(err, ErrInvalidTile) {
				t.Errorf("validateTile() = %v, ingin ErrInvalidTile", err)
			}
		})
	}
}

func TestMapServiceError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "key kosong", err: ErrNotConfigured, status: http.StatusServiceUnavailable},
		{name: "style", err: ErrUnsupportedStyle, status: http.StatusBadRequest},
		{name: "tile", err: ErrInvalidTile, status: http.StatusBadRequest},
		{name: "upstream", err: errors.New("boom"), status: http.StatusBadGateway},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := mapServiceError(tc.err)
			statusErr, ok := err.(interface{ GetStatus() int })
			if !ok || statusErr.GetStatus() != tc.status {
				t.Errorf("mapServiceError() = %v, ingin status %d", err, tc.status)
			}
		})
	}
}

func TestBasemapOpenAPI(t *testing.T) {
	api, _ := setupTileAPI(t, &fakeTileService{}, 1000)
	content := api.OpenAPI().Paths[openAPITilePath].Get.Responses["200"].Content
	if content[tileContentType] == nil {
		t.Errorf("OpenAPI tidak memuat media type %q", tileContentType)
	}
}
