// Package home menyediakan kontrak ringkas untuk halaman utama aplikasi.
package home

import (
	"context"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	"github.com/Junx27/kicaugo-backend/internal/features/event"
	"github.com/Junx27/kicaugo-backend/internal/features/profile"
)

// Options adalah parameter opsional permintaan home.
type Options struct {
	RadiusKM   int
	CategoryID int64
	Limit      int
}

// ProfileReader adalah bagian profile.Service yang dipakai home.
type ProfileReader interface {
	Get(ctx context.Context, identifier string) (profile.ProfileView, error)
}

// EventReader adalah bagian event.Service yang dipakai home.
type EventReader interface {
	ListNearby(ctx context.Context, identifier string, radiusKM int, categoryID int64, limit int) (event.NearbyResult, error)
	CountOngoing(ctx context.Context) (int, error)
}

// BirdCategoryReader adalah bagian birdcategory.Service yang dipakai home.
type BirdCategoryReader interface {
	List(ctx context.Context) ([]entities.BirdCategory, error)
}

// View menggabungkan profil, event terdekat, sapaan, dan jumlah event berlangsung.
type View struct {
	Profile            profile.ProfileView
	Nearby             event.NearbyResult
	Greeting           string
	OngoingEventsCount int
}

// Featured adalah data demo featured contest Beranda.
type Featured struct {
	Title     string
	Schedule  string
	Distance  string
	Prize     string
	Slots     string
	Attendees FeaturedAttendees
}

// FeaturedAttendees adalah ringkasan peserta featured contest.
type FeaturedAttendees struct {
	Initials []string
	Others   int
}

// Upcoming adalah data demo upcoming contest Beranda.
type Upcoming struct {
	Type             string
	Title            string
	Time             string
	Price            string
	RemainingTickets int
}

// Stats adalah data demo statistik Beranda.
type Stats struct {
	Badge        string
	Headline     string
	Note         string
	SecurityNote string
}

// ActiveTicket adalah data demo tiket aktif Beranda.
type ActiveTicket struct {
	Event    string
	Schedule string
	Number   string
	Session  string
}

// ScheduleCategory adalah kategori burung pada jadwal Beranda.
type ScheduleCategory struct {
	ID   int64
	Name string
}

// Schedule adalah data demo item jadwal Beranda.
type Schedule struct {
	Title    string
	Venue    string
	Date     string
	Price    string
	Status   string
	Category ScheduleCategory
}

// Service menyediakan data halaman home.
type Service interface {
	Get(ctx context.Context, identifier string, opts Options) (View, error)
	GetFeatured() Featured
	GetUpcoming() Upcoming
	GetStats() Stats
	GetActiveTicket() ActiveTicket
	GetSchedules() []Schedule
	GetBirdCategories(ctx context.Context) ([]entities.BirdCategory, error)
}

// service adalah implementasi Service yang mengorkestrasi service existing.
type service struct {
	profiles       ProfileReader
	events         EventReader
	birdCategories BirdCategoryReader
	now            func() time.Time
}

// NewService membuat service dengan dependency home yang sudah ada.
func NewService(profiles ProfileReader, events EventReader, birdCategories BirdCategoryReader) Service {
	return NewServiceWithClock(profiles, events, birdCategories, time.Now)
}

// NewServiceWithClock membuat service dengan jam yang dapat ditentukan, antara lain untuk test.
func NewServiceWithClock(profiles ProfileReader, events EventReader, birdCategories BirdCategoryReader, now func() time.Time) Service {
	return &service{profiles: profiles, events: events, birdCategories: birdCategories, now: now}
}

// Get mengambil profil, event terdekat, dan jumlah event berlangsung untuk identifier tertentu.
func (s *service) Get(ctx context.Context, identifier string, opts Options) (View, error) {
	view, err := s.profiles.Get(ctx, identifier)
	if err != nil {
		return View{}, err
	}
	nearby, err := s.events.ListNearby(ctx, identifier, opts.RadiusKM, opts.CategoryID, opts.Limit)
	if err != nil {
		return View{}, err
	}
	count, err := s.events.CountOngoing(ctx)
	if err != nil {
		return View{}, err
	}
	return View{
		Profile:            view,
		Nearby:             nearby,
		Greeting:           greeting(s.now()),
		OngoingEventsCount: count,
	}, nil
}

func (s *service) GetFeatured() Featured {
	return demoFeatured()
}

func (s *service) GetUpcoming() Upcoming {
	return demoUpcoming()
}

func (s *service) GetStats() Stats {
	return demoStats()
}

func (s *service) GetActiveTicket() ActiveTicket {
	return demoActiveTicket()
}

func (s *service) GetSchedules() []Schedule {
	return demoSchedules()
}

func (s *service) GetBirdCategories(ctx context.Context) ([]entities.BirdCategory, error) {
	return s.birdCategories.List(ctx)
}

func demoFeatured() Featured {
	return Featured{
		Title:    "Piala Raja Nusantara 2026",
		Schedule: "Minggu, 6 Des • Batavia Arena, Jaksel",
		Distance: "3.2 KM",
		Prize:    "Rp 50 Jt",
		Slots:    "12 Gantangan",
		Attendees: FeaturedAttendees{
			Initials: []string{"MK", "BN"},
			Others:   48,
		},
	}
}

func demoUpcoming() Upcoming {
	return Upcoming{
		Type:             "Latber Presisi",
		Title:            "Latber Presisi Kacer",
		Time:             "Besok, 14:00 WIB",
		Price:            "Rp 75.000",
		RemainingTickets: 3,
	}
}

func demoStats() Stats {
	return Stats{
		Badge:        "100% Resmi",
		Headline:     "520+ Gantangan",
		Note:         "Terverifikasi panitia & EO resmi.",
		SecurityNote: "98% Transaksi Tiket Instan & Aman",
	}
}

func demoActiveTicket() ActiveTicket {
	return ActiveTicket{
		Event:    "Grand Prix Murai Batu VIP #18",
		Schedule: "Sabtu, 12 Des • Masterpiece Arena",
		Number:   "#G-18",
		Session:  "Sesi A - 16G",
	}
}

func demoSchedules() []Schedule {
	return []Schedule{
		{
			Title:  "Latpres Exclusive Cucak Ijo",
			Venue:  "Arena Master Cilandak, Jaksel",
			Date:   "Selasa, 24 Nov",
			Price:  "Rp 100rb",
			Status: "available",
			Category: ScheduleCategory{
				ID:   1,
				Name: "Cucak Hijau",
			},
		},
		{
			Title:  "Festival Kicau Batavia",
			Venue:  "Gantangan Taman Mini, Jaktim",
			Date:   "Kamis, 26 Nov",
			Price:  "Rp 150rb",
			Status: "almost_gone",
			Category: ScheduleCategory{
				ID:   2,
				Name: "Murai Batu",
			},
		},
		{
			Title:  "Latber Kacer Mingguan",
			Venue:  "Lapangan Kicau Pesanggrahan, Jaksel",
			Date:   "Sabtu, 28 Nov",
			Price:  "Rp 60rb",
			Status: "sold_out",
			Category: ScheduleCategory{
				ID:   3,
				Name: "Kacer",
			},
		},
		{
			Title:  "Kenari Cup Kemang",
			Venue:  "Kemang Bird Park, Jaksel",
			Date:   "Minggu, 29 Nov",
			Price:  "Rp 85rb",
			Status: "waiting",
			Category: ScheduleCategory{
				ID:   4,
				Name: "Kenari",
			},
		},
	}
}
