package entities

import "time"

// OriginCandidate adalah data asal pencarian yang dibaca dari profil pengguna.
type OriginCandidate struct {
	CurrentLatitude  *float64
	CurrentLongitude *float64
	CurrentUpdatedAt *time.Time
	CityLatitude     *float64
	CityLongitude    *float64
	CityID           *int64
	RadiusKM         int
}

// NearbyOrigin adalah titik asal yang sudah dipilih oleh service.
type NearbyOrigin struct {
	Latitude  float64
	Longitude float64
	Source    string
	CityID    *int64
}

// Event adalah ringkasan event hasil pencarian terdekat.
type Event struct {
	ID          int64
	OrganizerID int64
	Title       string
	Slug        string
	Venue       string
	City        string
	Province    string
	CityID      *int64
	ProvinceID  *int64
	Latitude    *float64
	Longitude   *float64
	StartDate   *time.Time
	EndDate     *time.Time
	Status      string
	BannerURL   *string
	DistanceKM  float64
}

// SubEvent adalah sub-lomba di dalam sebuah event.
type SubEvent struct {
	ID             int64
	EventID        int64
	BirdCategoryID *int64
	Title          string
	Description    *string
	Session        *string
	GantanganCount int
	MaxSlots       int
	Price          int64
	Winner1Prize   int64
	Winner2Prize   int64
	Winner3Prize   int64
	Status         string
	StartTime      *time.Time
	EndTime        *time.Time
	SortOrder      int
}

// EventCreate adalah payload pembuatan event beserta sub-event.
type EventCreate struct {
	OrganizerID       int64
	Title             string
	Slug              string
	Description       *string
	Venue             string
	Address           *string
	City              string
	Province          *string
	CityID            *int64
	ProvinceID        *int64
	Latitude          *float64
	Longitude         *float64
	StartDate         *time.Time
	EndDate           *time.Time
	RegistrationOpen  *time.Time
	RegistrationClose *time.Time
	Status            string
	IsReusable        bool
	BannerURL         *string
	PrizePool         *string
	TotalSlots        int
	SubEvents         []SubEventCreate
}

// SubEventCreate adalah payload pembuatan sub-event.
type SubEventCreate struct {
	BirdCategoryID *int64
	Title          string
	Description    *string
	Session        *string
	GantanganCount int
	MaxSlots       int
	Price          int64
	Winner1Prize   int64
	Winner2Prize   int64
	Winner3Prize   int64
	Status         string
	StartTime      *time.Time
	EndTime        *time.Time
	SortOrder      int
}

// EventDetail adalah event yang dibuat beserta sub-event-nya.
type EventDetail struct {
	Event     Event
	SubEvents []SubEvent
}

// UpcomingEvent adalah ringkasan event upcoming untuk halaman home.
type UpcomingEvent struct {
	ID               int64
	Title            string
	StartDate        *time.Time
	RegistrationOpen *time.Time
	MinPrice         int64
	RemainingTickets int
}

// FeaturedEvent adalah ringkasan event featured untuk halaman home.
type FeaturedEvent struct {
	ID         int64
	Title      string
	StartDate  *time.Time
	Venue      string
	City       string
	Address    *string
	Latitude   *float64
	Longitude  *float64
	PrizePool  *string
	TotalSlots int
	DistanceKM *float64
}
