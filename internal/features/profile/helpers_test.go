package profile

import (
	"errors"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

func TestNormalizePhone(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{name: "kosong", raw: "", want: ""},
		{name: "awalan 0", raw: "0812-3456-7890", want: "+6281234567890"},
		{name: "awalan 62", raw: "6281234567890", want: "+6281234567890"},
		{name: "sudah plus dan spasi", raw: "+62 812 3456 7890", want: "+6281234567890"},
		{name: "tanpa awalan", raw: "81234567890", want: "+6281234567890"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizePhone(tc.raw); got != tc.want {
				t.Errorf("normalizePhone(%q) = %q, ingin %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestValidateLocation(t *testing.T) {
	lat, lng := -6.2, 106.8
	outsideLat, outsideLng := -20.0, 106.8
	cases := []struct {
		name      string
		latitude  *float64
		longitude *float64
		wantErr   bool
	}{
		{name: "lokasi belum diisi", latitude: nil, longitude: nil},
		{name: "pasangan lengkap", latitude: &lat, longitude: &lng},
		{name: "latitude saja", latitude: &lat, wantErr: true},
		{name: "longitude saja", longitude: &lng, wantErr: true},
		{name: "di luar bbox", latitude: &outsideLat, longitude: &outsideLng, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateLocation(tc.latitude, tc.longitude)
			if tc.wantErr && !errors.Is(err, apperrors.ErrInvalidCoordinates) {
				t.Errorf("validateLocation() = %v, ingin ErrInvalidCoordinates", err)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("validateLocation() = %v, ingin nil", err)
			}
		})
	}
}

func TestBuildView(t *testing.T) {
	t.Run("profil belum lengkap", func(t *testing.T) {
		view := buildView(entities.UserProfile{}, nil)
		if view.IsComplete {
			t.Error("IsComplete ingin false")
		}
		if len(view.MissingFields) != 2 || view.MissingFields[0] != fieldFullName || view.MissingFields[1] != fieldCityID {
			t.Errorf("MissingFields = %v", view.MissingFields)
		}
	})

	t.Run("profil lengkap", func(t *testing.T) {
		cityID := int64(1)
		view := buildView(entities.UserProfile{FullName: "Budi", CityID: &cityID}, nil)
		if !view.IsComplete || len(view.MissingFields) != 0 {
			t.Errorf("view = %+v", view)
		}
	})
}

func TestUniqueIDs(t *testing.T) {
	got := uniqueIDs([]int64{3, 1, 3, 2, 1})
	if len(got) != 3 || got[0] != 3 || got[1] != 1 || got[2] != 2 {
		t.Errorf("uniqueIDs() = %v", got)
	}
}
