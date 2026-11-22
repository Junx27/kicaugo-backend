package response

// HomeProfileData adalah ringkasan profil untuk halaman home.
// Nomor telepon dan koordinat lokasi sengaja tidak disertakan.
type HomeProfileData struct {
	Greeting           string                 `json:"greeting"`
	FullName           string                 `json:"full_name"`
	AvatarURL          *string                `json:"avatar_url"`
	City               *ProfileCityData       `json:"city"`
	SearchRadiusKM     int                    `json:"search_radius_km"`
	FavoriteCategories []FavoriteCategoryData `json:"favorite_categories"`
	IsComplete         bool                   `json:"is_complete"`
	MissingFields      []string               `json:"missing_fields"`
}

// HomeNearbyData adalah data event terdekat yang ditambah jumlah event berlangsung.
type HomeNearbyData struct {
	NearbyEventsData
	OngoingEventsCount int `json:"ongoing_events_count"`
}

// HomeData menggabungkan ringkasan profil dan event terdekat.
type HomeData struct {
	Profile HomeProfileData `json:"profile"`
	Nearby  HomeNearbyData  `json:"nearby"`
}

// HomeOutput adalah output GET /api/v1/home/profile.
type HomeOutput struct {
	Status int `status:"200"`
	Body   Success[HomeData]
}
