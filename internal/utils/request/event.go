// Package request berisi tipe input request yang dipakai handler.
package request

// ListNearbyEventsInput adalah input GET /api/v1/events/nearby.
// Nilai 0 pada radius_km dan category_id berarti memakai nilai default.
type ListNearbyEventsInput struct {
	RadiusKM   int   `query:"radius_km" default:"0" minimum:"0" maximum:"500" doc:"Radius km, 0 memakai radius profil"`
	CategoryID int64 `query:"category_id" default:"0" minimum:"0" doc:"Filter kategori burung; 0 berarti semua"`
	Limit      int   `query:"limit" default:"20" minimum:"1" maximum:"50" doc:"Jumlah maksimal event"`
}
