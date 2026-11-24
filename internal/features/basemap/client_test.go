package basemap

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

const (
	fakeCartoKey     = "fake-carto-key"
	keyLeakFormat    = "error memuat key: %v"
	fetchErrFormat   = "Fetch() error = %v"
	tilePathTemplate = "/{z}/{x}/{y}.png"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func clientResponse(status int, body []byte) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestClientFetch(t *testing.T) {
	t.Run("sukses", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/rastertiles/voyager/3/2/1@2x.png" {
				t.Errorf("path = %q", r.URL.Path)
			}
			if r.URL.Query().Get(apiKeyParam) != fakeCartoKey {
				t.Errorf("query key = %q", r.URL.Query().Get(apiKeyParam))
			}
			_, _ = w.Write([]byte("png"))
		}))
		defer srv.Close()

		client := NewClient(fakeCartoKey, srv.URL+"/rastertiles/voyager/{z}/{x}/{y}{r}.png")
		got, err := client.Fetch(context.Background(), 3, 2, 1, true, defaultStyle)
		if err != nil {
			t.Fatalf(fetchErrFormat, err)
		}
		if string(got) != "png" {
			t.Errorf("body = %q", got)
		}
	})

	t.Run("subdomain dan retina", func(t *testing.T) {
		var captured *url.URL
		client := NewClient(fakeCartoKey, "https://{s}.example.com/tiles/{z}/{x}/{y}{r}.png")
		client.http.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			captured = req.URL
			return clientResponse(http.StatusOK, []byte("ok")), nil
		})

		if _, err := client.Fetch(context.Background(), 3, 1, 2, true, defaultStyle); err != nil {
			t.Fatalf(fetchErrFormat, err)
		}
		if captured.Host != "d.example.com" {
			t.Errorf("host = %q, ingin d.example.com", captured.Host)
		}
		if captured.Path != "/tiles/3/1/2@2x.png" {
			t.Errorf("path = %q", captured.Path)
		}
		if captured.Query().Get(apiKeyParam) != fakeCartoKey {
			t.Errorf("key = %q", captured.Query().Get(apiKeyParam))
		}
	})

	t.Run("tanpa retina", func(t *testing.T) {
		var captured *url.URL
		client := NewClient(fakeCartoKey, "https://{s}.example.com/tiles/{z}/{x}/{y}{r}.png")
		client.http.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			captured = req.URL
			return clientResponse(http.StatusOK, []byte("ok")), nil
		})

		if _, err := client.Fetch(context.Background(), 0, 0, 0, false, defaultStyle); err != nil {
			t.Fatalf(fetchErrFormat, err)
		}
		if captured.Host != "a.example.com" || captured.Path != "/tiles/0/0/0.png" {
			t.Errorf("url = %s", captured)
		}
	})

	t.Run("template kosong memakai default", func(t *testing.T) {
		client := NewClient(fakeCartoKey, "")
		if client.styles[defaultStyle] != defaultTemplate {
			t.Errorf("template = %q", client.styles[defaultStyle])
		}
	})

	t.Run("key kosong", func(t *testing.T) {
		client := NewClient("", defaultTemplate)
		if _, err := client.Fetch(context.Background(), 0, 0, 0, false, defaultStyle); !errors.Is(err, ErrNotConfigured) {
			t.Errorf("error = %v, ingin ErrNotConfigured", err)
		}
	})

	t.Run("style tidak dikenal", func(t *testing.T) {
		client := NewClient(fakeCartoKey, defaultTemplate)
		if _, err := client.Fetch(context.Background(), 0, 0, 0, false, "dark"); !errors.Is(err, ErrUnsupportedStyle) {
			t.Errorf("error = %v, ingin ErrUnsupportedStyle", err)
		}
	})

	t.Run("tile tidak valid", func(t *testing.T) {
		client := NewClient(fakeCartoKey, defaultTemplate)
		tests := []struct {
			name string
			z    int
			x    int
			y    int
		}{
			{name: "zoom negatif", z: -1},
			{name: "zoom terlalu besar", z: maxZoom + 1},
			{name: "x negatif", z: 1, x: -1},
			{name: "y negatif", z: 1, y: -1},
			{name: "x di luar batas", z: 1, x: 2},
			{name: "y di luar batas", z: 1, y: 2},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				if _, err := client.Fetch(context.Background(), tc.z, tc.x, tc.y, false, defaultStyle); !errors.Is(err, ErrInvalidTile) {
					t.Errorf("error = %v, ingin ErrInvalidTile", err)
				}
			})
		}
	})

	t.Run("upstream 404 dan 500", func(t *testing.T) {
		for _, status := range []int{http.StatusNotFound, http.StatusInternalServerError} {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
			}))
			client := NewClient(fakeCartoKey, srv.URL+tilePathTemplate)
			_, err := client.Fetch(context.Background(), 0, 0, 0, false, defaultStyle)
			srv.Close()
			if err == nil {
				t.Errorf("status %d: ingin error", status)
			}
			if strings.Contains(err.Error(), fakeCartoKey) {
				t.Errorf(keyLeakFormat, err)
			}
		}
	})

	t.Run("timeout", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		client := NewClient(fakeCartoKey, srv.URL+tilePathTemplate)
		client.http.Timeout = 10 * time.Millisecond
		_, err := client.Fetch(context.Background(), 0, 0, 0, false, defaultStyle)
		if err == nil {
			t.Error("Fetch() ingin error timeout")
		}
		if strings.Contains(err.Error(), fakeCartoKey) {
			t.Errorf(keyLeakFormat, err)
		}
	})

	t.Run("body terlalu besar", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write(make([]byte, maxTileBytes+1))
		}))
		defer srv.Close()

		client := NewClient(fakeCartoKey, srv.URL+tilePathTemplate)
		if _, err := client.Fetch(context.Background(), 0, 0, 0, false, defaultStyle); err == nil {
			t.Error("Fetch() ingin error body terlalu besar")
		}
	})

	t.Run("gagal membaca body", func(t *testing.T) {
		if _, err := readLimited(errorReader{}); err == nil {
			t.Error("readLimited() ingin error")
		}
	})

	t.Run("template tidak valid", func(t *testing.T) {
		client := NewClient(fakeCartoKey, "http://[::1")
		_, err := client.Fetch(context.Background(), 0, 0, 0, false, defaultStyle)
		if err == nil {
			t.Error("Fetch() ingin error template")
		}
		if strings.Contains(err.Error(), fakeCartoKey) {
			t.Errorf(keyLeakFormat, err)
		}
	})

	t.Run("context nil", func(t *testing.T) {
		client := NewClient(fakeCartoKey, defaultTemplate)
		_, err := client.Fetch(nil, 0, 0, 0, false, defaultStyle)
		if err == nil {
			t.Error("Fetch() ingin error context nil")
		}
		if strings.Contains(err.Error(), fakeCartoKey) {
			t.Errorf(keyLeakFormat, err)
		}
	})

	t.Run("context dibatalkan", func(t *testing.T) {
		client := NewClient(fakeCartoKey, defaultTemplate)
		client.http.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, req.Context().Err()
		})
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := client.Fetch(ctx, 0, 0, 0, false, defaultStyle); err == nil {
			t.Error("Fetch() ingin error context dibatalkan")
		}
	})

	t.Run("key tidak bocor dari error transport", func(t *testing.T) {
		client := NewClient(fakeCartoKey, defaultTemplate)
		client.http.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, &url.Error{Op: "Get", URL: "https://example.com/?key=" + fakeCartoKey, Err: errors.New("boom")}
		})
		_, err := client.Fetch(context.Background(), 0, 0, 0, false, defaultStyle)
		if err == nil {
			t.Fatal("Fetch() ingin error")
		}
		if strings.Contains(err.Error(), fakeCartoKey) {
			t.Errorf(keyLeakFormat, err)
		}
	})

	t.Run("sanitize error non-url", func(t *testing.T) {
		plain := errors.New("plain")
		if got := sanitizeRequestError(plain); got != plain {
			t.Errorf("sanitize = %v, ingin %v", got, plain)
		}
	})

	t.Run("sanitize url error kosong", func(t *testing.T) {
		if err := sanitizeRequestError(&url.Error{}); err == nil {
			t.Error("sanitizeRequestError() ingin error")
		}
	})
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}
