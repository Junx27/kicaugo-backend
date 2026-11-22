package response

// ProfileProvinceData adalah ringkasan provinsi di dalam profil.
type ProfileProvinceData struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// ProfileCityData adalah kota domisili beserta provinsinya.
type ProfileCityData struct {
	ID       int64               `json:"id"`
	Name     string              `json:"name"`
	Type     string              `json:"type"`
	Slug     string              `json:"slug"`
	Province ProfileProvinceData `json:"province"`
}

// FavoriteCategoryData adalah kategori burung favorit.
type FavoriteCategoryData struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// CurrentLocationData adalah lokasi terkini milik user itu sendiri.
type CurrentLocationData struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	AccuracyM *int    `json:"accuracy_m"`
	UpdatedAt string  `json:"updated_at"`
}

// ProfileData adalah representasi profil lengkap untuk pemiliknya.
type ProfileData struct {
	FullName           string                 `json:"full_name"`
	Phone              *string                `json:"phone"`
	AvatarURL          *string                `json:"avatar_url"`
	Bio                *string                `json:"bio"`
	City               *ProfileCityData       `json:"city"`
	CurrentLocation    *CurrentLocationData   `json:"current_location"`
	SearchRadiusKM     int                    `json:"search_radius_km"`
	FavoriteCategories []FavoriteCategoryData `json:"favorite_categories"`
	IsComplete         bool                   `json:"is_complete"`
	MissingFields      []string               `json:"missing_fields"`
}

// ProfileOutput adalah output GET /me/profile.
type ProfileOutput struct {
	Status int `status:"200"`
	Body   Success[ProfileData]
}

// ProfileMutationOutput adalah output untuk PATCH dan PUT profil.
type ProfileMutationOutput struct {
	Status int `status:"200"`
	Body   Success[struct{}]
}
