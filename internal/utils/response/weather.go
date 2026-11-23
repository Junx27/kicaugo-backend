package response

// WeatherCurrentData adalah kondisi cuaca saat ini pada response.
type WeatherCurrentData struct {
	TempC       float64 `json:"temp_c"`
	FeelsLikeC  float64 `json:"feelslike_c"`
	Condition   string  `json:"condition"`
	Icon        string  `json:"icon"`
	Humidity    int     `json:"humidity"`
	WindKph     float64 `json:"wind_kph"`
	PrecipMm    float64 `json:"precip_mm"`
	LastUpdated string  `json:"last_updated"`
}

// WeatherForecastDayData adalah prakiraan cuaca per hari pada response.
type WeatherForecastDayData struct {
	Date       string  `json:"date"`
	MaxTempC   float64 `json:"maxtemp_c"`
	MinTempC   float64 `json:"mintemp_c"`
	Condition  string  `json:"condition"`
	Icon       string  `json:"icon"`
	ChanceRain int     `json:"chance_of_rain"`
}

// WeatherData adalah payload cuaca yang dikembalikan API.
type WeatherData struct {
	Location  string                   `json:"location"`
	Latitude  float64                  `json:"latitude"`
	Longitude float64                  `json:"longitude"`
	Current   WeatherCurrentData       `json:"current"`
	Forecast  []WeatherForecastDayData `json:"forecast"`
}

// WeatherOutput adalah output GET /api/v1/weather/forecast.
type WeatherOutput struct {
	Status int `status:"200"`
	Body   Success[WeatherData]
}
