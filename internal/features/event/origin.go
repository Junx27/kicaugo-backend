package event

import (
	"math"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

const (
	// LocationFreshWindow adalah batas usia lokasi terkini yang masih dipakai sebagai asal.
	LocationFreshWindow = 24 * time.Hour
	// OriginCurrentLocation menandakan asal dari lokasi terkini pengguna.
	OriginCurrentLocation = "current_location"
	// OriginCityCentroid menandakan asal dari titik pusat kota domisili.
	OriginCityCentroid = "city_centroid"
	// originNone dipakai saat tidak ada titik asal yang bisa dipakai.
	originNone = "none"
	// HintLocationNotSet diberikan ketika pengguna belum memiliki lokasi maupun kota.
	HintLocationNotSet = "location_not_set"
	// dateLayout adalah format tanggal pada response event.
	dateLayout      = "2006-01-02"
	defaultRadiusKM = 50
	maxRadiusKM     = 500
	defaultLimit    = 20
	maxLimit        = 50
)

// resolveOrigin memilih lokasi terkini yang masih segar, lalu titik pusat kota.
func resolveOrigin(candidate entities.OriginCandidate, now time.Time) (*entities.NearbyOrigin, *string) {
	if candidate.CurrentLatitude != nil && candidate.CurrentLongitude != nil && candidate.CurrentUpdatedAt != nil {
		age := now.Sub(*candidate.CurrentUpdatedAt)
		if age >= 0 && age < LocationFreshWindow {
			return &entities.NearbyOrigin{
				Latitude:  *candidate.CurrentLatitude,
				Longitude: *candidate.CurrentLongitude,
				Source:    OriginCurrentLocation,
				CityID:    candidate.CityID,
			}, nil
		}
	}
	if candidate.CityLatitude != nil && candidate.CityLongitude != nil {
		return &entities.NearbyOrigin{
			Latitude:  *candidate.CityLatitude,
			Longitude: *candidate.CityLongitude,
			Source:    OriginCityCentroid,
			CityID:    candidate.CityID,
		}, nil
	}
	hint := HintLocationNotSet
	return nil, &hint
}

// effectiveRadius memakai radius permintaan, lalu radius profil, lalu nilai default.
func effectiveRadius(requested, profile int) int {
	if requested > 0 {
		return requested
	}
	if profile > 0 {
		return profile
	}
	return defaultRadiusKM
}

// roundDistance membulatkan jarak ke satu angka desimal.
func roundDistance(distance float64) float64 {
	return math.Round(distance*10) / 10
}
