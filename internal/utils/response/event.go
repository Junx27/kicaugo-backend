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

// SubEventData adalah representasi sub-event di response.
type SubEventData struct {
	ID             int64   `json:"id"`
	EventID        int64   `json:"event_id"`
	BirdCategoryID *int64  `json:"bird_category_id"`
	Title          string  `json:"title"`
	Description    *string `json:"description"`
	Session        *string `json:"session"`
	GantanganCount int     `json:"gantangan_count"`
	MaxSlots       int     `json:"max_slots"`
	Price          int64   `json:"price"`
	Winner1Prize   int64   `json:"winner_1_prize"`
	Winner2Prize   int64   `json:"winner_2_prize"`
	Winner3Prize   int64   `json:"winner_3_prize"`
	Status         string  `json:"status"`
	StartTime      *string `json:"start_time"`
	EndTime        *string `json:"end_time"`
	SortOrder      int     `json:"sort_order"`
}

// EventData adalah representasi event yang dibuat.
type EventData struct {
	ID                int64          `json:"id"`
	OrganizerID       int64          `json:"organizer_id"`
	Title             string         `json:"title"`
	Slug              string         `json:"slug"`
	Description       *string        `json:"description"`
	Venue             string         `json:"venue"`
	Address           *string        `json:"address"`
	City              string         `json:"city"`
	Province          *string        `json:"province"`
	CityID            *int64         `json:"city_id"`
	ProvinceID        *int64         `json:"province_id"`
	Latitude          *float64       `json:"latitude"`
	Longitude         *float64       `json:"longitude"`
	StartDate         *string        `json:"start_date"`
	EndDate           *string        `json:"end_date"`
	RegistrationOpen  *string        `json:"registration_open"`
	RegistrationClose *string        `json:"registration_close"`
	Status            string         `json:"status"`
	IsReusable        bool           `json:"is_reusable"`
	BannerURL         *string        `json:"banner_url"`
	PrizePool         *string        `json:"prize_pool"`
	TotalSlots        int            `json:"total_slots"`
	SubEvents         []SubEventData `json:"sub_events"`
}

// CreateEventOutput adalah output POST /api/v1/events.
type CreateEventOutput struct {
	Status int `status:"201"`
	Body   Success[EventData]
}
