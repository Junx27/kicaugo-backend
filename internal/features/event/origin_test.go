package event

import (
	"testing"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

func float64Ptr(v float64) *float64  { return &v }
func int64Ptr(v int64) *int64        { return &v }
func timePtr(v time.Time) *time.Time { return &v }

func TestResolveOrigin(t *testing.T) {
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	cityID := int64(5)

	cases := []struct {
		name       string
		candidate  entities.OriginCandidate
		wantSource string
		wantLat    float64
		wantHint   bool
	}{
		{
			name: "lokasi terkini masih segar",
			candidate: entities.OriginCandidate{
				CurrentLatitude: float64Ptr(-6.2), CurrentLongitude: float64Ptr(106.8),
				CurrentUpdatedAt: timePtr(now.Add(-time.Hour)), CityID: &cityID,
			},
			wantSource: OriginCurrentLocation,
			wantLat:    -6.2,
		},
		{
			name: "lokasi terkini kedaluwarsa memakai pusat kota",
			candidate: entities.OriginCandidate{
				CurrentLatitude: float64Ptr(-6.2), CurrentLongitude: float64Ptr(106.8),
				CurrentUpdatedAt: timePtr(now.Add(-25 * time.Hour)),
				CityLatitude:     float64Ptr(-6.9), CityLongitude: float64Ptr(107.6), CityID: &cityID,
			},
			wantSource: OriginCityCentroid,
			wantLat:    -6.9,
		},
		{
			name: "batas usia tepat 24 jam memakai pusat kota",
			candidate: entities.OriginCandidate{
				CurrentLatitude: float64Ptr(-6.2), CurrentLongitude: float64Ptr(106.8),
				CurrentUpdatedAt: timePtr(now.Add(-LocationFreshWindow)),
				CityLatitude:     float64Ptr(-6.9), CityLongitude: float64Ptr(107.6),
			},
			wantSource: OriginCityCentroid,
			wantLat:    -6.9,
		},
		{
			name: "timestamp masa depan memakai pusat kota",
			candidate: entities.OriginCandidate{
				CurrentLatitude: float64Ptr(-6.2), CurrentLongitude: float64Ptr(106.8),
				CurrentUpdatedAt: timePtr(now.Add(time.Hour)),
				CityLatitude:     float64Ptr(-6.9), CityLongitude: float64Ptr(107.6),
			},
			wantSource: OriginCityCentroid,
			wantLat:    -6.9,
		},
		{
			name: "pasangan lokasi tidak lengkap",
			candidate: entities.OriginCandidate{
				CurrentLatitude: float64Ptr(-6.2),
				CityLatitude:    float64Ptr(-6.9), CityLongitude: float64Ptr(107.6),
			},
			wantSource: OriginCityCentroid,
			wantLat:    -6.9,
		},
		{
			name: "lokasi segar tanpa kota tetap dipakai",
			candidate: entities.OriginCandidate{
				CurrentLatitude: float64Ptr(-6.2), CurrentLongitude: float64Ptr(106.8),
				CurrentUpdatedAt: timePtr(now.Add(-time.Minute)),
			},
			wantSource: OriginCurrentLocation,
			wantLat:    -6.2,
		},
		{
			name:      "tanpa lokasi maupun kota",
			candidate: entities.OriginCandidate{},
			wantHint:  true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			origin, hint := resolveOrigin(tc.candidate, now)
			if tc.wantHint {
				if origin != nil || hint == nil || *hint != HintLocationNotSet {
					t.Errorf("origin = %+v, hint = %v", origin, hint)
				}
				return
			}
			if origin == nil || origin.Source != tc.wantSource || origin.Latitude != tc.wantLat || hint != nil {
				t.Errorf("origin = %+v, hint = %v", origin, hint)
			}
		})
	}
}

func TestEffectiveRadius(t *testing.T) {
	cases := []struct {
		requested int
		profile   int
		want      int
	}{
		{requested: 25, profile: 10, want: 25},
		{requested: 0, profile: 10, want: 10},
		{requested: 0, profile: 0, want: defaultRadiusKM},
	}
	for _, tc := range cases {
		if got := effectiveRadius(tc.requested, tc.profile); got != tc.want {
			t.Errorf("effectiveRadius(%d, %d) = %d, ingin %d", tc.requested, tc.profile, got, tc.want)
		}
	}
}

func TestRoundDistance(t *testing.T) {
	if got := roundDistance(3.14159); got != 3.1 {
		t.Errorf("roundDistance() = %v, ingin 3.1", got)
	}
}
