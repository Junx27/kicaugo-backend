package weather

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientFetch(t *testing.T) {
	t.Run("sukses", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"location": map[string]any{"name": "Yogyakarta", "lat": -7.7, "lon": 110.3},
				"current": map[string]any{
					"temp_c": 28.5, "feelslike_c": 30.1, "humidity": 70, "wind_kph": 12.3,
					"precip_mm": 0.0, "last_updated": "2026-11-23 14:00",
					"condition": map[string]any{"text": "Partly cloudy", "icon": "//cdn.weatherapi.com/icon.png"},
				},
				"forecast": map[string]any{
					"forecastday": []map[string]any{{
						"date": "2026-11-23",
						"day": map[string]any{
							"maxtemp_c": 30.0, "mintemp_c": 24.0, "daily_chance_of_rain": 20,
							"condition": map[string]any{"text": "Sunny", "icon": "//cdn.weatherapi.com/sun.png"},
						},
					}},
				},
			})
		}))
		defer srv.Close()

		c := NewClient("key", srv.URL)
		got, err := c.Fetch(context.Background(), -7.7, 110.3, 1)
		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}
		if got.Location != "Yogyakarta" || got.Current.TempC != 28.5 {
			t.Errorf("got = %+v", got)
		}
		if got.Current.Icon != "https://cdn.weatherapi.com/icon.png" {
			t.Errorf("icon = %q", got.Current.Icon)
		}
		if len(got.Forecast) != 1 || got.Forecast[0].ChanceRain != 20 {
			t.Errorf("forecast = %+v", got.Forecast)
		}
	})

	t.Run("tidak terkonfigurasi", func(t *testing.T) {
		c := NewClient("", "https://example.com")
		if _, err := c.Fetch(context.Background(), 0, 0, 1); err != ErrNotConfigured {
			t.Errorf("err = %v, ingin ErrNotConfigured", err)
		}
	})

	t.Run("status bukan 200", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer srv.Close()
		c := NewClient("key", srv.URL)
		if _, err := c.Fetch(context.Background(), 0, 0, 1); err == nil {
			t.Error("Fetch() ingin error")
		}
	})

	t.Run("body tidak valid", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("{invalid"))
		}))
		defer srv.Close()
		c := NewClient("key", srv.URL)
		if _, err := c.Fetch(context.Background(), 0, 0, 1); err == nil {
			t.Error("Fetch() ingin error")
		}
	})

	t.Run("base kosong memakai default", func(t *testing.T) {
		c := NewClient("key", "")
		if c.base != "https://api.weatherapi.com/v1" {
			t.Errorf("base = %q", c.base)
		}
	})

	t.Run("base tidak valid", func(t *testing.T) {
		c := NewClient("key", "http://[::1")
		if _, err := c.Fetch(context.Background(), 0, 0, 1); err == nil {
			t.Error("Fetch() ingin error")
		}
	})

	t.Run("days kurang dari 1", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("days") != "1" {
				t.Errorf("days = %q, ingin 1", r.URL.Query().Get("days"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"location": map[string]any{"name": "X", "lat": 0.0, "lon": 0.0},
				"current":  map[string]any{"temp_c": 1.0, "condition": map[string]any{"text": "Clear", "icon": "i.png"}},
				"forecast": map[string]any{"forecastday": []any{}},
			})
		}))
		defer srv.Close()
		c := NewClient("key", srv.URL)
		if _, err := c.Fetch(context.Background(), 0, 0, 0); err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}
	})

	t.Run("days lebih dari 10", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("days") != "10" {
				t.Errorf("days = %q, ingin 10", r.URL.Query().Get("days"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"location": map[string]any{"name": "X", "lat": 0.0, "lon": 0.0},
				"current":  map[string]any{"temp_c": 1.0, "condition": map[string]any{"text": "Clear", "icon": "i.png"}},
				"forecast": map[string]any{"forecastday": []any{}},
			})
		}))
		defer srv.Close()
		c := NewClient("key", srv.URL)
		if _, err := c.Fetch(context.Background(), 0, 0, 99); err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}
	})

	t.Run("context dibatalkan", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		c := NewClient("key", srv.URL)
		if _, err := c.Fetch(ctx, 0, 0, 1); err == nil {
			t.Error("Fetch() ingin error")
		}
	})
}
