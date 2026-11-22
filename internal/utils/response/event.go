package response

// NearbyEventData adalah representasi ringkas event terdekat.
// Koordinat venue sengaja tidak diekspos; hanya jarak dari titik asal.
type NearbyEventData struct {
	ID         int64   `json:"id"`
	Title      string  `json:"title"`
	Slug       string  `json:"slug"`
	Venue      string  `json:"venue"`
	City       string  `json:"city"`
	Province   string  `json:"province"`
	StartDate  *string `json:"start_date"`
	EndDate    *string `json:"end_date"`
	Status     string  `json:"status"`
	BannerURL  *string `json:"banner_url"`
	DistanceKM float64 `json:"distance_km"`
}

// NearbyEventsData membawa titik asal, radius, dan daftar event.
type NearbyEventsData struct {
	OriginSource    string            `json:"origin_source"`
	OriginLatitude  *float64          `json:"origin_latitude"`
	OriginLongitude *float64          `json:"origin_longitude"`
	RadiusKM        int               `json:"radius_km"`
	Events          []NearbyEventData `json:"events"`
	Hint            *string           `json:"hint"`
}

// NearbyEventsOutput adalah output GET /api/v1/events/nearby.
type NearbyEventsOutput struct {
	Status int `status:"200"`
	Body   Success[NearbyEventsData]
}
