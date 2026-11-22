// Package request berisi tipe input request yang dipakai handler.
package request

// UpdateProfileInput adalah input PATCH /me/profile.
// Semua field opsional memakai omitempty agar Huma tidak mewajibkannya.
type UpdateProfileInput struct {
	Body struct {
		FullName       *string `json:"full_name,omitempty" minLength:"1" maxLength:"100" doc:"Nama lengkap"`
		Phone          *string `json:"phone,omitempty" maxLength:"30" doc:"Nomor telepon Indonesia"`
		AvatarURL      *string `json:"avatar_url,omitempty" maxLength:"500" doc:"URL foto profil"`
		Bio            *string `json:"bio,omitempty" maxLength:"500" doc:"Biografi singkat"`
		CityID         *int64  `json:"city_id,omitempty" minimum:"1" doc:"ID kabupaten/kota domisili"`
		ProvinceID     *int64  `json:"province_id,omitempty" minimum:"1" doc:"Opsional, divalidasi terhadap city_id"`
		SearchRadiusKM *int    `json:"search_radius_km,omitempty" minimum:"1" maximum:"500" doc:"Radius pencarian default dalam km"`
	}
}

// CurrentLocationInput adalah input PUT /me/profile/current-location.
// Badan kosong berarti menghapus lokasi terkini.
type CurrentLocationInput struct {
	Body struct {
		Latitude  *float64 `json:"latitude,omitempty" minimum:"-90" maximum:"90" doc:"Latitude, wajib berpasangan dengan longitude"`
		Longitude *float64 `json:"longitude,omitempty" minimum:"-180" maximum:"180" doc:"Longitude, wajib berpasangan dengan latitude"`
		AccuracyM *int     `json:"accuracy_m,omitempty" minimum:"0" doc:"Akurasi lokasi dalam meter"`
	}
}

// FavoriteCategoriesInput adalah input PUT /me/profile/favorite-categories.
type FavoriteCategoriesInput struct {
	Body struct {
		CategoryIDs []int64 `json:"category_ids" required:"true" doc:"Daftar ID kategori burung favorit"`
	}
}
