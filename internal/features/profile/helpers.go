package profile

import (
	"strings"
	"unicode"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

const (
	indonesiaMinLatitude  = -11.5
	indonesiaMaxLatitude  = 6.5
	indonesiaMinLongitude = 94.0
	indonesiaMaxLongitude = 141.5
	fieldFullName         = "full_name"
	fieldCityID           = "city_id"
)

// normalizePhone mengubah nomor Indonesia menjadi format +62.
func normalizePhone(raw string) string {
	digits := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, raw)
	switch {
	case digits == "":
		return ""
	case strings.HasPrefix(digits, "62"):
		return "+" + digits
	case strings.HasPrefix(digits, "0"):
		return "+62" + digits[1:]
	default:
		return "+62" + digits
	}
}

// validateLocation memastikan pasangan koordinat lengkap dan berada di kotak Indonesia.
func validateLocation(latitude, longitude *float64) error {
	if (latitude == nil) != (longitude == nil) {
		return apperrors.ErrInvalidCoordinates
	}
	if latitude == nil {
		return nil
	}
	if *latitude < indonesiaMinLatitude || *latitude > indonesiaMaxLatitude ||
		*longitude < indonesiaMinLongitude || *longitude > indonesiaMaxLongitude {
		return apperrors.ErrInvalidCoordinates
	}
	return nil
}

// buildView menyusun profil beserta status kelengkapannya.
func buildView(p entities.UserProfile, favorites []entities.BirdCategory) ProfileView {
	missing := make([]string, 0, 2)
	if strings.TrimSpace(p.FullName) == "" {
		missing = append(missing, fieldFullName)
	}
	if p.CityID == nil {
		missing = append(missing, fieldCityID)
	}
	return ProfileView{
		Profile:       p,
		Favorites:     favorites,
		IsComplete:    len(missing) == 0,
		MissingFields: missing,
	}
}

// uniqueIDs menghapus ID duplikat sehingga jumlah valid bisa dibandingkan.
func uniqueIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
