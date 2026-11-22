package response

// CityData adalah representasi ringkas kabupaten/kota untuk dropdown bertingkat.
type CityData struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Slug string `json:"slug"`
}

// CityListOutput adalah output untuk GET list kabupaten/kota per provinsi.
type CityListOutput struct {
	Status int `status:"200"`
	Body   Success[[]CityData]
}

// scaffold:response-types
