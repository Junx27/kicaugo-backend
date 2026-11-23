package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

// ErrNotConfigured dikembalikan bila API key WeatherAPI belum di-set.
var ErrNotConfigured = fmt.Errorf("weather api key not configured")

type apiCondition struct {
	Text string `json:"text"`
	Icon string `json:"icon"`
}

type apiCurrent struct {
	TempC       float64      `json:"temp_c"`
	FeelsLikeC  float64      `json:"feelslike_c"`
	Humidity    int          `json:"humidity"`
	WindKph     float64      `json:"wind_kph"`
	PrecipMm    float64      `json:"precip_mm"`
	LastUpdated string       `json:"last_updated"`
	Condition   apiCondition `json:"condition"`
}

type apiForecastDay struct {
	Date string `json:"date"`
	Day  struct {
		MaxTempC     float64      `json:"maxtemp_c"`
		MinTempC     float64      `json:"mintemp_c"`
		ChanceOfRain int          `json:"daily_chance_of_rain"`
		Condition    apiCondition `json:"condition"`
	} `json:"day"`
}

type apiResponse struct {
	Location struct {
		Name string  `json:"name"`
		Lat  float64 `json:"lat"`
		Lon  float64 `json:"lon"`
	} `json:"location"`
	Current  apiCurrent `json:"current"`
	Forecast struct {
		ForecastDay []apiForecastDay `json:"forecastday"`
	} `json:"forecast"`
}

// Client memanggil WeatherAPI.com.
type Client struct {
	apiKey string
	base   string
	lang   string
	http   *http.Client
}

// NewClient membuat client dengan timeout 5 detik.
func NewClient(apiKey, base, lang string) *Client {
	if base == "" {
		base = "https://api.weatherapi.com/v1"
	}
	if lang == "" {
		lang = "id"
	}
	return &Client{apiKey: apiKey, base: base, lang: lang, http: &http.Client{Timeout: 5 * time.Second}}
}

// Fetch mengambil prakiraan cuaca untuk koordinat dan jumlah hari.
func (c *Client) Fetch(ctx context.Context, lat, lon float64, days int) (entities.Weather, error) {
	if c.apiKey == "" {
		return entities.Weather{}, ErrNotConfigured
	}
	if days < 1 {
		days = 1
	}
	if days > 10 {
		days = 10
	}
	u, err := url.Parse(c.base + "/forecast.json")
	if err != nil {
		return entities.Weather{}, fmt.Errorf("parse url: %w", err)
	}
	q := u.Query()
	q.Set("key", c.apiKey)
	q.Set("q", fmt.Sprintf("%.6f,%.6f", lat, lon))
	q.Set("days", strconv.Itoa(days))
	q.Set("aqi", "no")
	q.Set("alerts", "no")
	if c.lang != "" {
		q.Set("lang", c.lang)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return entities.Weather{}, fmt.Errorf("build request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return entities.Weather{}, fmt.Errorf("weather request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return entities.Weather{}, fmt.Errorf("weather api status %d", resp.StatusCode)
	}
	var api apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&api); err != nil {
		return entities.Weather{}, fmt.Errorf("decode weather: %w", err)
	}
	return mapWeather(api, lat, lon), nil
}

func mapWeather(api apiResponse, lat, lon float64) entities.Weather {
	lastUpdated, _ := time.Parse("2006-01-02 15:04", api.Current.LastUpdated)
	forecast := make([]entities.WeatherForecastDay, 0, len(api.Forecast.ForecastDay))
	for _, d := range api.Forecast.ForecastDay {
		forecast = append(forecast, entities.WeatherForecastDay{
			Date:       d.Date,
			MaxTempC:   d.Day.MaxTempC,
			MinTempC:   d.Day.MinTempC,
			Condition:  translateCondition(d.Day.Condition.Text),
			Icon:       normalizeIcon(d.Day.Condition.Icon),
			ChanceRain: d.Day.ChanceOfRain,
		})
	}
	return entities.Weather{
		Location:  api.Location.Name,
		Latitude:  lat,
		Longitude: lon,
		Current: entities.WeatherCurrent{
			TempC:       api.Current.TempC,
			FeelsLikeC:  api.Current.FeelsLikeC,
			Condition:   translateCondition(api.Current.Condition.Text),
			Icon:        normalizeIcon(api.Current.Condition.Icon),
			Humidity:    api.Current.Humidity,
			WindKph:     api.Current.WindKph,
			PrecipMm:    api.Current.PrecipMm,
			LastUpdated: lastUpdated,
		},
		Forecast: forecast,
	}
}

func normalizeIcon(icon string) string {
	if strings.HasPrefix(icon, "//") {
		return "https:" + icon
	}
	return icon
}
