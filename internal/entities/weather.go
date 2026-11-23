package entities

import "time"

// WeatherCurrent adalah kondisi cuaca saat ini.
type WeatherCurrent struct {
	TempC       float64   `json:"temp_c"`
	FeelsLikeC  float64   `json:"feelslike_c"`
	Condition   string    `json:"condition"`
	Icon        string    `json:"icon"`
	Humidity    int       `json:"humidity"`
	WindKph     float64   `json:"wind_kph"`
	PrecipMm    float64   `json:"precip_mm"`
	LastUpdated time.Time `json:"last_updated"`
}

// WeatherForecastDay adalah prakiraan cuaca per hari.
type WeatherForecastDay struct {
	Date       string  `json:"date"`
	MaxTempC   float64 `json:"maxtemp_c"`
	MinTempC   float64 `json:"mintemp_c"`
	Condition  string  `json:"condition"`
	Icon       string  `json:"icon"`
	ChanceRain int     `json:"chance_of_rain"`
}

// Weather adalah hasil deteksi cuaca untuk sebuah lokasi.
type Weather struct {
	Location  string               `json:"location"`
	Latitude  float64              `json:"latitude"`
	Longitude float64              `json:"longitude"`
	Current   WeatherCurrent       `json:"current"`
	Forecast  []WeatherForecastDay `json:"forecast"`
}
