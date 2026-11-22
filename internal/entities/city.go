package entities

import "time"

// City adalah entitas kabupaten/kota.
type City struct {
	ID         int64
	ProvinceID int64
	Code       string
	Name       string
	Type       string
	Slug       string
	Latitude   float64
	Longitude  float64
	CreatedAt  time.Time
}
