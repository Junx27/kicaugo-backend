package response

// CityData adalah representasi city di response.
type CityData struct {
	ID         int64   `json:"id"`
	ProvinceID int64   `json:"province_id"`
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Slug       string  `json:"slug"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
}

// CityListOutput adalah output untuk GET list.
type CityListOutput struct {
	Status int `status:"200"`
	Body   Success[[]CityData]
}

// scaffold:response-types
