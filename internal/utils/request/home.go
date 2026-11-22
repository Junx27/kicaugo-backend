// Package request berisi tipe input request yang dipakai handler.
package request

// HomeProfileInput adalah input GET /api/v1/home/profile.
// Nilai 0 pada radius_km dan category_id memakai nilai dari profil.
type HomeProfileInput struct {
	RadiusKM   int   `query:"radius_km" default:"0" minimum:"0" maximum:"500" doc:"Radius km, 0 memakai radius profil"`
	CategoryID int64 `query:"category_id" default:"0" minimum:"0" doc:"Filter kategori burung; 0 berarti semua"`
	Limit      int   `query:"limit" default:"5" minimum:"1" maximum:"20" doc:"Jumlah maksimal event terdekat"`
}
