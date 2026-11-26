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

// ContestQuotaData adalah kuota contest.
type ContestQuotaData struct {
	Total     int    `json:"total"`
	Remaining int    `json:"remaining"`
	Label     string `json:"label"`
}

// ContestScheduleData adalah jadwal contest.
type ContestScheduleData struct {
	Date     string `json:"date"`
	Time     string `json:"time"`
	GateOpen string `json:"gate_open"`
}

// ContestLocationData adalah lokasi contest.
type ContestLocationData struct {
	Name        string   `json:"name"`
	Address     string   `json:"address"`
	Distance    string   `json:"distance"`
	MapImageURL *string  `json:"map_image_url"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	Parking     string   `json:"parking"`
}

// ContestPrizeData adalah hadiah contest.
type ContestPrizeData struct {
	Total     string `json:"total"`
	Trophy    string `json:"trophy"`
	Highlight string `json:"highlight"`
}

// ContestStatItemData adalah item statistik contest.
type ContestStatItemData struct {
	Label string `json:"label"`
	Note  string `json:"note"`
}

// ContestStatsData adalah statistik contest.
type ContestStatsData struct {
	Format     ContestStatItemData `json:"format"`
	Judging    ContestStatItemData `json:"judging"`
	Facilities ContestStatItemData `json:"facilities"`
}

// ContestClassData adalah kelas contest.
type ContestClassData struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Session        *string `json:"session"`
	Time           string  `json:"time"`
	RemainingSlots *int    `json:"remaining_slots"`
	Price          string  `json:"price"`
	PrizeHighlight string  `json:"prize_highlight"`
}

// ContestRuleData adalah aturan contest.
type ContestRuleData struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Severity    string  `json:"severity"`
}

// ContestTicketData adalah info tiket contest.
type ContestTicketData struct {
	StartingPrice string `json:"starting_price"`
	Unit          string `json:"unit"`
	CTA           string `json:"cta"`
}

// ContestDetailData adalah response detail contest.
type ContestDetailData struct {
	ID             int64               `json:"id"`
	Title          string              `json:"title"`
	Subtitle       string              `json:"subtitle"`
	OfficialBadge  string              `json:"official_badge"`
	FormatBadge    string              `json:"format_badge"`
	EditionLabel   string              `json:"edition_label"`
	HeroImageURL   *string             `json:"hero_image_url"`
	BannerImageURL *string             `json:"banner_image_url"`
	Status         string              `json:"status"`
	Quota          ContestQuotaData    `json:"quota"`
	Schedule       ContestScheduleData `json:"schedule"`
	Location       ContestLocationData `json:"location"`
	Prize          ContestPrizeData    `json:"prize"`
	Stats          ContestStatsData    `json:"stats"`
	Classes        []ContestClassData  `json:"classes"`
	Rules          []ContestRuleData   `json:"rules"`
	Ticket         ContestTicketData   `json:"ticket"`
	Distance       string              `json:"distance"`
	ScheduleText   string              `json:"schedule_text"`
}

// ContestDetailOutput adalah output GET /api/v1/contests/{id}.
type ContestDetailOutput struct {
	Status int `status:"200"`
	Body   Success[ContestDetailData]
}

// ListEventRulesOutput adalah output GET /api/v1/events/{id}/rules.
type ListEventRulesOutput struct {
	Status int `status:"200"`
	Body   Success[[]ContestRuleData]
}

// CreateEventRuleOutput adalah output POST /api/v1/events/{id}/rules.
type CreateEventRuleOutput struct {
	Status int `status:"201"`
	Body   Success[ContestRuleData]
}

// UpdateEventRuleOutput adalah output PATCH /api/v1/events/{id}/rules/{ruleID}.
type UpdateEventRuleOutput struct {
	Status int `status:"200"`
	Body   Success[ContestRuleData]
}

// DeleteEventRuleOutput adalah output DELETE /api/v1/events/{id}/rules/{ruleID}.
type DeleteEventRuleOutput struct {
	Status int `status:"200"`
	Body   Success[struct{}]
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
