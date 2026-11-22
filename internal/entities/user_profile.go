package entities

import "time"

// UserProfile adalah profil pengguna dengan lokasi opsional.
type UserProfile struct {
	ID                       int64
	UserID                   int64
	FullName                 string
	Phone                    *string
	AvatarURL                *string
	Bio                      *string
	CityID                   *int64
	CurrentLatitude          *float64
	CurrentLongitude         *float64
	CurrentLocationAccuracyM *int
	CurrentLocationUpdatedAt *time.Time
	SearchRadiusKM           int
	CreatedAt                time.Time
	UpdatedAt                time.Time
	City                     *City
	Province                 *Province
}

// ProfileUpdate adalah perubahan parsial pada profil.
type ProfileUpdate struct {
	FullName       *string
	Phone          *string
	AvatarURL      *string
	Bio            *string
	CityID         *int64
	SearchRadiusKM *int
}
