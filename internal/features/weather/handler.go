package weather

import (
	"context"
	"net/http"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	"github.com/Junx27/kicaugo-backend/internal/middlewares"
	"github.com/Junx27/kicaugo-backend/internal/utils/authctx"
	apphttp "github.com/Junx27/kicaugo-backend/internal/utils/http"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
	"github.com/Junx27/kicaugo-backend/internal/utils/response"
)

const (
	weatherTag             = "weather"
	bearerAuthSecurityName = "bearerAuth"
)

// ForecastInput adalah input GET /api/v1/weather/forecast.
type ForecastInput struct {
	Latitude  float64 `query:"lat" required:"true" minimum:"-90" maximum:"90" doc:"Latitude lokasi"`
	Longitude float64 `query:"lon" required:"true" minimum:"-180" maximum:"180" doc:"Longitude lokasi"`
	Days      int     `query:"days" default:"1" minimum:"1" maximum:"10" doc:"Jumlah hari prakiraan"`
}

// newGetForecastHandler membentuk handler agar dapat diuji langsung.
func newGetForecastHandler(service Service) func(context.Context, *ForecastInput) (*response.WeatherOutput, error) {
	return func(ctx context.Context, input *ForecastInput) (*response.WeatherOutput, error) {
		if _, err := authctx.IdentifierFromContext(ctx); err != nil {
			return nil, err
		}
		w, err := service.GetForecast(ctx, input.Latitude, input.Longitude, input.Days)
		if err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.WeatherOutput{}
		resp.Status = http.StatusOK
		resp.Body = response.NewSuccess(http.StatusOK, "Weather forecast loaded", toWeatherData(w))
		return resp, nil
	}
}

func toWeatherData(w entities.Weather) response.WeatherData {
	current := response.WeatherCurrentData{
		TempC:       w.Current.TempC,
		FeelsLikeC:  w.Current.FeelsLikeC,
		Condition:   w.Current.Condition,
		Icon:        w.Current.Icon,
		Humidity:    w.Current.Humidity,
		WindKph:     w.Current.WindKph,
		PrecipMm:    w.Current.PrecipMm,
		LastUpdated: w.Current.LastUpdated.Format(time.RFC3339),
	}
	forecast := make([]response.WeatherForecastDayData, 0, len(w.Forecast))
	for _, d := range w.Forecast {
		forecast = append(forecast, response.WeatherForecastDayData{
			Date:       d.Date,
			MaxTempC:   d.MaxTempC,
			MinTempC:   d.MinTempC,
			Condition:  d.Condition,
			Icon:       d.Icon,
			ChanceRain: d.ChanceRain,
		})
	}
	return response.WeatherData{
		Location:  w.Location,
		Latitude:  w.Latitude,
		Longitude: w.Longitude,
		Current:   current,
		Forecast:  forecast,
	}
}

// RegisterRoutes mendaftarkan route weather ke Huma API.
func RegisterRoutes(api apphttp.API, service Service, tokens *jwt.Manager) {
	auth := []apphttp.Middleware{middlewares.AuthMiddleware[jwt.Claims](tokens)}
	security := []map[string][]string{{bearerAuthSecurityName: {}}}

	apphttp.Register(api, apphttp.Operation{
		ID:          "getWeatherForecast",
		Method:      http.MethodGet,
		Path:        "/api/v1/weather/forecast",
		Summary:     "Get Weather Forecast",
		Tags:        []string{weatherTag},
		Security:    security,
		Middlewares: auth,
	}, newGetForecastHandler(service))
}

func mapServiceError(err error) error {
	if err == ErrNotConfigured {
		return apphttp.Error(503, "Weather service not configured")
	}
	return apphttp.Error(502, "Weather service unavailable")
}
