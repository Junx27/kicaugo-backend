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
	ListNearbySchedules(ctx context.Context, identifier string, radiusKM int, categoryID int64, limit int) ([]entities.EventSchedule, error)
	CountOngoing(ctx context.Context) (int, error)
	CountGantangan(ctx context.Context) (int, error)
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
	Address   *string
	Latitude  *float64
	Longitude *float64
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
	Event     string
	Schedule  string
	Number    string
	Session   string
	Address   *string
	Latitude  *float64
	Longitude *float64
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
	GetSchedules(ctx context.Context, identifier string) ([]Schedule, error)
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

const defaultScheduleLimit = 20

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
	stats := defaultStats()
	count, err := s.events.CountGantangan(context.Background())
	if err == nil && count >= 0 {
		stats.Headline = fmt.Sprintf("%d+ Gantangan", count)
	}
	return stats
}

func (s *service) GetActiveTicket() ActiveTicket {
	return demoActiveTicket()
}

func (s *service) GetSchedules(ctx context.Context, identifier string) ([]Schedule, error) {
	items, err := s.events.ListNearbySchedules(ctx, identifier, 0, 0, defaultScheduleLimit)
	if err != nil {
		return nil, err
	}
	out := make([]Schedule, 0, len(items))
	for _, item := range items {
		out = append(out, scheduleFromEventSchedule(item))
	}
	return out, nil
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
		Address:   ev.Address,
		Latitude:  ev.Latitude,
		Longitude: ev.Longitude,
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

func defaultStats() Stats {
	return Stats{
		Badge:        "100% Resmi",
		Headline:     "520+ Gantangan",
		Note:         "Terverifikasi panitia & EO resmi.",
		SecurityNote: "98% Transaksi Tiket Instan & Aman",
	}
}

func demoActiveTicket() ActiveTicket {
	address := "GOR Kota, Banda Aceh"
	lat, lng := -6.2, 106.8
	return ActiveTicket{
		Event:     "Grand Prix Murai Batu VIP #18",
		Schedule:  "Sabtu, 12 Des • Masterpiece Arena",
		Number:    "#G-18",
		Session:   "Sesi A - 16G",
		Address:   &address,
		Latitude:  &lat,
		Longitude: &lng,
	}
}

// scheduleFromEventSchedule memetakan EventSchedule dari database ke data jadwal home.
func scheduleFromEventSchedule(item entities.EventSchedule) Schedule {
	date := "-"
	if item.Date != nil {
		date = item.Date.Format("Monday, 02 Jan")
	}
	price := "Gratis"
	if item.Price > 0 {
		price = format.Rupiah(item.Price)
	}
	category := ScheduleCategory{}
	if item.CategoryID != nil {
		category.ID = *item.CategoryID
	}
	if item.CategoryName != "" {
		category.Name = item.CategoryName
	} else {
		category.Name = "-"
	}
	return Schedule{
		Title:    item.Title,
		Venue:    item.Venue,
		Date:     date,
		Price:    price,
		Status:   item.Status,
		Category: category,
	}
}
