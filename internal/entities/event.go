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
	ID         int64
	Title      string
	Slug       string
	Venue      string
	City       string
	Province   string
	CityID     *int64
	ProvinceID *int64
	Latitude   *float64
	Longitude  *float64
	StartDate  *time.Time
	EndDate    *time.Time
	Status     string
	BannerURL  *string
	DistanceKM float64
}
