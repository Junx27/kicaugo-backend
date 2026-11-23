package weather

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2/humatest"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	"github.com/Junx27/kicaugo-backend/internal/testutil"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
	"github.com/Junx27/kicaugo-backend/internal/utils/response"
)

type fakeService struct {
	weather entities.Weather
	err     error
}

func (f *fakeService) GetForecast(ctx context.Context, lat, lon float64, days int) (entities.Weather, error) {
	return f.weather, f.err
}

func setupHandler(t *testing.T, svc Service) (humatest.TestAPI, *jwt.Manager) {
	t.Helper()
	_, api := humatest.New(t)
	tokens := testutil.NewTokenManager()
	RegisterRoutes(api, svc, tokens)
	return api, tokens
}

func sampleWeather() entities.Weather {
	return entities.Weather{
		Location:  "Yogyakarta",
		Latitude:  -7.7,
		Longitude: 110.3,
		Current: entities.WeatherCurrent{
			TempC: 28.5, FeelsLikeC: 30.1, Condition: "Partly cloudy", Icon: "https://icon.png",
			Humidity: 70, WindKph: 12.3, PrecipMm: 0.1,
			LastUpdated: time.Date(2026, 11, 23, 14, 0, 0, 0, time.UTC),
		},
		Forecast: []entities.WeatherForecastDay{
			{Date: "2026-11-23", MaxTempC: 30, MinTempC: 24, Condition: "Sunny", Icon: "https://sun.png", ChanceRain: 20},
		},
	}
}

func TestGetWeatherForecastHandler(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		svc := &fakeService{weather: sampleWeather()}
		api, tokens := setupHandler(t, svc)
		resp := api.Get("/api/v1/weather/forecast?lat=-7.7&lon=110.3&days=1", testutil.BearerHeader(t, tokens))
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
		}
		var body struct {
			Status  int                  `json:"status"`
			Message string               `json:"message"`
			Data    response.WeatherData `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("json: %v", err)
		}
		if body.Status != http.StatusOK || body.Message != "Weather forecast loaded" {
			t.Errorf("status/message = %d/%q", body.Status, body.Message)
		}
		if body.Data.Location != "Yogyakarta" || body.Data.Current.TempC != 28.5 {
			t.Errorf("data = %+v", body.Data)
		}
		if len(body.Data.Forecast) != 1 || body.Data.Forecast[0].ChanceRain != 20 {
			t.Errorf("forecast = %+v", body.Data.Forecast)
		}
	})

	t.Run("tanpa token", func(t *testing.T) {
		api, _ := setupHandler(t, &fakeService{})
		if resp := api.Get("/api/v1/weather/forecast?lat=1&lon=2"); resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})

	t.Run("service error", func(t *testing.T) {
		api, tokens := setupHandler(t, &fakeService{err: errors.New("boom")})
		if resp := api.Get("/api/v1/weather/forecast?lat=1&lon=2", testutil.BearerHeader(t, tokens)); resp.Code != http.StatusBadGateway {
			t.Errorf("status = %d, ingin 502", resp.Code)
		}
	})

	t.Run("tidak terkonfigurasi", func(t *testing.T) {
		api, tokens := setupHandler(t, &fakeService{err: ErrNotConfigured})
		if resp := api.Get("/api/v1/weather/forecast?lat=1&lon=2", testutil.BearerHeader(t, tokens)); resp.Code != http.StatusServiceUnavailable {
			t.Errorf("status = %d, ingin 503", resp.Code)
		}
	})

	t.Run("param wajib", func(t *testing.T) {
		api, tokens := setupHandler(t, &fakeService{})
		if resp := api.Get("/api/v1/weather/forecast", testutil.BearerHeader(t, tokens)); resp.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, ingin 422", resp.Code)
		}
	})
}

func TestGetWeatherForecastHandlerTanpaIdentitas(t *testing.T) {
	_, err := newGetForecastHandler(&fakeService{})(context.Background(), &ForecastInput{})
	statusErr, ok := err.(interface{ GetStatus() int })
	if err == nil || !ok || statusErr.GetStatus() != http.StatusUnauthorized {
		t.Errorf("handler tanpa identitas = %v, ingin 401", err)
	}
}
func TestMapServiceErrorWeather(t *testing.T) {
	if got := mapServiceError(ErrNotConfigured); got == nil {
		t.Error("ingin error")
	}
	if got := mapServiceError(errors.New("boom")); got == nil {
		t.Error("ingin error")
	}
}
