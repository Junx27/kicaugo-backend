// Package request berisi tipe input request yang dipakai handler.
package request

import "time"

// ListNearbyEventsInput adalah input GET /api/v1/events/nearby.
// Nilai 0 pada radius_km dan category_id berarti memakai nilai default.
type ListNearbyEventsInput struct {
	RadiusKM   int   `query:"radius_km" default:"0" minimum:"0" maximum:"500" doc:"Radius km, 0 memakai radius profil"`
	CategoryID int64 `query:"category_id" default:"0" minimum:"0" doc:"Filter kategori burung; 0 berarti semua"`
	Limit      int   `query:"limit" default:"20" minimum:"1" maximum:"50" doc:"Jumlah maksimal event"`
}

// CreateSubEventInput adalah input sub-event di dalam pembuatan event.
type CreateSubEventInput struct {
	BirdCategoryID *int64     `json:"bird_category_id,omitempty" minimum:"1" doc:"ID kategori burung"`
	Title          string     `json:"title" required:"true" minLength:"3" maxLength:"150" doc:"Judul sub-event"`
	Description    *string    `json:"description,omitempty" maxLength:"1000" doc:"Deskripsi sub-event"`
	Session        *string    `json:"session,omitempty" maxLength:"100" doc:"Sesi sub-event"`
	GantanganCount int        `json:"gantangan_count,omitempty" minimum:"0" doc:"Jumlah gantangan"`
	MaxSlots       int        `json:"max_slots,omitempty" default:"60" minimum:"1" maximum:"60" doc:"Maksimal slot"`
	Price          int64      `json:"price,omitempty" minimum:"0" doc:"Harga tiket"`
	Winner1Prize   int64      `json:"winner_1_prize,omitempty" minimum:"0" doc:"Hadiah juara 1"`
	Winner2Prize   int64      `json:"winner_2_prize,omitempty" minimum:"0" doc:"Hadiah juara 2"`
	Winner3Prize   int64      `json:"winner_3_prize,omitempty" minimum:"0" doc:"Hadiah juara 3"`
	Status         string     `json:"status,omitempty" enum:"available,almostGone,soldOut,waiting" default:"available" doc:"Status sub-event"`
	StartTime      *time.Time `json:"start_time,omitempty" doc:"Waktu mulai sub-event"`
	EndTime        *time.Time `json:"end_time,omitempty" doc:"Waktu selesai sub-event"`
	SortOrder      int        `json:"sort_order,omitempty" minimum:"0" doc:"Urutan tampil"`
}

// CreateEventInput adalah input POST /api/v1/events.
type CreateEventInput struct {
	Body struct {
		OrganizerID       int64                 `json:"organizer_id" required:"true" minimum:"1" doc:"ID organizer penyelenggara"`
		Title             string                `json:"title" required:"true" minLength:"3" maxLength:"150" doc:"Judul event"`
		Slug              string                `json:"slug" required:"true" minLength:"3" maxLength:"150" pattern:"^[a-z0-9]+(?:-[a-z0-9]+)*$" doc:"Slug unik event"`
		Description       *string               `json:"description,omitempty" maxLength:"2000" doc:"Deskripsi event"`
		Venue             string                `json:"venue" required:"true" minLength:"3" maxLength:"200" doc:"Nama venue"`
		Address           *string               `json:"address,omitempty" maxLength:"255" doc:"Alamat lengkap"`
		City              string                `json:"city" required:"true" minLength:"2" maxLength:"100" doc:"Kota"`
		Province          *string               `json:"province,omitempty" maxLength:"100" doc:"Provinsi"`
		CityID            *int64                `json:"city_id,omitempty" minimum:"1" doc:"ID kota referensi"`
		ProvinceID        *int64                `json:"province_id,omitempty" minimum:"1" doc:"ID provinsi referensi"`
		Latitude          *float64              `json:"latitude,omitempty" minimum:"-90" maximum:"90" doc:"Latitude venue"`
		Longitude         *float64              `json:"longitude,omitempty" minimum:"-180" maximum:"180" doc:"Longitude venue"`
		StartDate         *time.Time            `json:"start_date,omitempty" doc:"Tanggal mulai"`
		EndDate           *time.Time            `json:"end_date,omitempty" doc:"Tanggal selesai"`
		RegistrationOpen  *time.Time            `json:"registration_open,omitempty" doc:"Waktu buka registrasi"`
		RegistrationClose *time.Time            `json:"registration_close,omitempty" doc:"Waktu tutup registrasi"`
		Status            string                `json:"status,omitempty" enum:"draft,published,ongoing,completed,cancelled" default:"draft" doc:"Status event"`
		IsReusable        bool                  `json:"is_reusable,omitempty" doc:"Apakah event reusable"`
		BannerURL         *string               `json:"banner_url,omitempty" maxLength:"500" doc:"URL banner"`
		PrizePool         *string               `json:"prize_pool,omitempty" maxLength:"255" doc:"Total hadiah"`
		TotalSlots        int                   `json:"total_slots,omitempty" minimum:"0" doc:"Total slot"`
		SubEvents         []CreateSubEventInput `json:"sub_events,omitempty" doc:"Daftar sub-event"`
	}
}
