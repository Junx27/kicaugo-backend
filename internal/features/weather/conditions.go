package weather

// conditionID memetakan teks kondisi WeatherAPI (Bahasa Inggris) ke Bahasa Indonesia.
// Sumber: https://www.weatherapi.com/docs/weather_conditions.json
var conditionID = map[string]string{
	"Sunny":                            "Cerah",
	"Clear":                            "Cerah",
	"Partly cloudy":                    "Sebagian berawan",
	"Cloudy":                           "Berawan",
	"Overcast":                         "Mendung",
	"Mist":                             "Berkabut",
	"Patchy rain possible":             "Hujan ringan mungkin",
	"Patchy snow possible":             "Salju ringan mungkin",
	"Patchy sleet possible":            "Sleet ringan mungkin",
	"Patchy freezing drizzle possible": "Gerimis beku mungkin",
	"Thundery outbreaks possible":      "Kemungkinan petir",
	"Blowing snow":                     "Salju berangin",
	"Blizzard":                         "Badai salju",
	"Fog":                              "Kabut",
	"Freezing fog":                     "Kabut beku",
	"Patchy light drizzle":             "Gerimis ringan sporadis",
	"Light drizzle":                    "Gerimis ringan",
	"Freezing drizzle":                 "Gerimis beku",
	"Heavy freezing drizzle":           "Gerimis beku lebat",
	"Patchy light rain":                "Hujan ringan sporadis",
	"Light rain":                       "Hujan ringan",
	"Moderate rain at times":           "Hujan sedang kadang-kadang",
	"Moderate rain":                    "Hujan sedang",
	"Heavy rain at times":              "Hujan lebat kadang-kadang",
	"Heavy rain":                       "Hujan lebat",
	"Light freezing rain":              "Hujan beku ringan",
	"Moderate or heavy freezing rain":  "Hujan beku sedang/lebat",
	"Light sleet":                      "Sleet ringan",
	"Moderate or heavy sleet":          "Sleet sedang/lebat",
	"Patchy light snow":                "Salju ringan sporadis",
	"Light snow":                       "Salju ringan",
	"Patchy moderate snow":             "Salju sedang sporadis",
	"Moderate snow":                    "Salju sedang",
	"Patchy heavy snow":                "Salju lebat sporadis",
	"Heavy snow":                       "Salju lebat",
	"Ice pellets":                      "Butiran es",
	"Light rain shower":                "Hujan ringan",
	"Moderate or heavy rain shower":    "Hujan sedang/lebat",
	"Torrential rain shower":           "Hujan sangat lebat",
	"Light sleet showers":              "Hujan sleet ringan",
	"Moderate or heavy sleet showers":  "Hujan sleet sedang/lebat",
	"Light snow showers":               "Hujan salju ringan",
	"Moderate or heavy snow showers":   "Hujan salju sedang/lebat",
	"Light showers of ice pellets":     "Hujan butiran es ringan",
	"Moderate or heavy showers of ice pellets": "Hujan butiran es sedang/lebat",
	"Patchy light rain with thunder":           "Hujan ringan dengan petir",
	"Moderate or heavy rain with thunder":      "Hujan sedang/lebat dengan petir",
	"Patchy light snow with thunder":           "Salju ringan dengan petir",
	"Moderate or heavy snow with thunder":      "Salju sedang/lebat dengan petir",
}

// translateCondition mengembalikan teks kondisi dalam Bahasa Indonesia bila tersedia.
func translateCondition(text string) string {
	if v, ok := conditionID[text]; ok {
		return v
	}
	return text
}
