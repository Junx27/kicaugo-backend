// Package home menyediakan kontrak ringkas untuk halaman utama aplikasi.
package home

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	"github.com/Junx27/kicaugo-backend/internal/features/event"
	"github.com/Junx27/kicaugo-backend/internal/features/profile"
	"github.com/Junx27/kicaugo-backend/internal/utils/format"
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
	GetActiveStreaming(ctx context.Context) (*entities.ActiveEventStreaming, error)
	GetRandomFeatured(ctx context.Context, identifier string) (*entities.FeaturedEvent, error)
	GetRandomUpcoming(ctx context.Context) (*entities.UpcomingEvent, error)
}

// BirdCategoryReader adalah bagian birdcategory.Service yang dipakai home.
type BirdCategoryReader interface {
	List(ctx context.Context) ([]entities.BirdCategory, error)
}

// OrganizerReader adalah bagian organizer.Service yang dipakai home.
type OrganizerReader interface {
	HasActiveMembership(ctx context.Context, identifier string) (bool, error)
}

// View menggabungkan profil, event terdekat, sapaan, dan jumlah event berlangsung.
type View struct {
	Profile            profile.ProfileView
	Nearby             event.NearbyResult
	Greeting           string
	OngoingEventsCount int
	ActiveStreaming    *entities.ActiveEventStreaming
	IsOrganizer        bool
}

// eventSummary merangkum data event yang dibutuhkan halaman home.
type eventSummary struct {
	nearby          event.NearbyResult
	ongoingCount    int
	activeStreaming *entities.ActiveEventStreaming
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
	GetFeatured(ctx context.Context, identifier string) (Featured, error)
	GetUpcoming(ctx context.Context) (Upcoming, error)
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
	organizers     OrganizerReader
	now            func() time.Time
}

// NewService membuat service dengan dependency home yang sudah ada.
func NewService(profiles ProfileReader, events EventReader, birdCategories BirdCategoryReader, organizers OrganizerReader) Service {
	return NewServiceWithClock(profiles, events, birdCategories, organizers, time.Now)
}

// NewServiceWithClock membuat service dengan jam yang dapat ditentukan, antara lain untuk test.
func NewServiceWithClock(profiles ProfileReader, events EventReader, birdCategories BirdCategoryReader, organizers OrganizerReader, now func() time.Time) Service {
	return &service{profiles: profiles, events: events, birdCategories: birdCategories, organizers: organizers, now: now}
}

// Get mengambil profil, status organizer, event terdekat, dan jumlah event berlangsung untuk identifier tertentu.
func (s *service) Get(ctx context.Context, identifier string, opts Options) (View, error) {
	view, err := s.profiles.Get(ctx, identifier)
	if err != nil {
		return View{}, err
	}
	isOrganizer, err := s.organizers.HasActiveMembership(ctx, identifier)
	if err != nil {
		return View{}, err
	}
	summary, err := s.loadEventSummary(ctx, identifier, opts)
	if err != nil {
		return View{}, err
	}
	return View{
		Profile:            view,
		Nearby:             summary.nearby,
		Greeting:           greeting(s.now()),
		OngoingEventsCount: summary.ongoingCount,
		ActiveStreaming:    summary.activeStreaming,
		IsOrganizer:        isOrganizer,
	}, nil
}

// loadEventSummary mengambil event terdekat, jumlah event berlangsung, dan streaming aktif.
func (s *service) loadEventSummary(ctx context.Context, identifier string, opts Options) (eventSummary, error) {
	nearby, err := s.events.ListNearby(ctx, identifier, opts.RadiusKM, opts.CategoryID, opts.Limit)
	if err != nil {
		return eventSummary{}, err
	}
	count, err := s.events.CountOngoing(ctx)
	if err != nil {
		return eventSummary{}, err
	}
	activeStreaming, err := s.events.GetActiveStreaming(ctx)
	if err != nil {
		return eventSummary{}, err
	}
	return eventSummary{nearby: nearby, ongoingCount: count, activeStreaming: activeStreaming}, nil
}

func (s *service) GetFeatured(ctx context.Context, identifier string) (Featured, error) {
	ev, err := s.events.GetRandomFeatured(ctx, identifier)
	if err != nil {
		return Featured{}, err
	}
	if ev == nil {
		return emptyFeatured(), nil
	}
	return featuredFromFeaturedEvent(*ev), nil
}

func (s *service) GetUpcoming(ctx context.Context) (Upcoming, error) {
	ev, err := s.events.GetRandomUpcoming(ctx)
	if err != nil {
		return Upcoming{}, err
	}
	if ev == nil {
		return emptyUpcoming(), nil
	}
	return upcomingFromEvent(*ev), nil
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

func emptyFeatured() Featured {
	return Featured{
		Title:     "Belum ada event featured",
		Schedule:  "-",
		Distance:  "-",
		Prize:     "-",
		Slots:     "-",
		Attendees: FeaturedAttendees{Initials: []string{}, Others: 0},
	}
}

// featuredFromFeaturedEvent memetakan FeaturedEvent ke data ringkas home.
func featuredFromFeaturedEvent(ev entities.FeaturedEvent) Featured {
	schedule := "-"
	if ev.StartDate != nil {
		schedule = ev.StartDate.Format("02 Jan 2006")
	}
	location := stringsTrimJoin([]string{ev.Venue, ev.City})
	if location != "" && schedule != "-" {
		schedule = fmt.Sprintf("%s • %s", schedule, location)
	} else if location != "" {
		schedule = location
	}
	prize := "-"
	if ev.PrizePool != nil && *ev.PrizePool != "" {
		prize = format.RupiahString(*ev.PrizePool)
	}
	slots := "-"
	if ev.TotalSlots > 0 {
		slots = fmt.Sprintf("%d Slot", ev.TotalSlots)
	}
	distance := "-"
	if ev.DistanceKM != nil {
		d := *ev.DistanceKM
		if d < 0.1 {
			distance = "< 0.1 km"
		} else {
			distance = fmt.Sprintf("%.1f km", d)
		}
	}
	return Featured{
		Title:     ev.Title,
		Schedule:  schedule,
		Distance:  distance,
		Prize:     prize,
		Slots:     slots,
		Attendees: FeaturedAttendees{Initials: []string{}, Others: 0},
	}
}

func stringsTrimJoin(parts []string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, ", ")
}

func emptyUpcoming() Upcoming {
	return Upcoming{Type: "Upcoming Event", Title: "Belum ada event upcoming", Time: "-", Price: "-", RemainingTickets: 0}
}

func upcomingFromEvent(ev entities.UpcomingEvent) Upcoming {
	timeText := "-"
	if ev.RegistrationOpen != nil {
		timeText = ev.RegistrationOpen.Format("02 Jan 2006 15:04")
	} else if ev.StartDate != nil {
		timeText = ev.StartDate.Format("02 Jan 2006")
	}
	price := "Gratis"
	if ev.MinPrice > 0 {
		price = format.Rupiah(ev.MinPrice)
	}
	return Upcoming{Type: "Upcoming Event", Title: ev.Title, Time: timeText, Price: price, RemainingTickets: ev.RemainingTickets}
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
