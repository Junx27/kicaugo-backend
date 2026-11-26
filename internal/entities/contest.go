package entities

import "time"

// ContestDetail adalah data detail lomba untuk halaman contest.
type ContestDetail struct {
	ID                  int64
	Title               string
	Subtitle            string
	OfficialBadge       string
	FormatBadge         string
	EditionLabel        string
	HeroImageURL        *string
	BannerImageURL      *string
	Status              string
	TotalSlots          int
	ActiveTickets       int
	ScheduleDate        *time.Time
	ScheduleTime        string
	GateOpen            string
	Venue               string
	Address             *string
	DistanceText        string
	MapImageURL         *string
	Latitude            *float64
	Longitude           *float64
	Parking             string
	PrizePool           *string
	PrizeTrophy         string
	PrizeHighlight      string
	StatFormatLabel     string
	StatFormatNote      string
	StatJudgingLabel    string
	StatJudgingNote     string
	StatFacilitiesLabel string
	StatFacilitiesNote  string
	Classes             []ContestClass
	Rules               []ContestRule
}

// ContestClass adalah kelas/sub-lomba pada detail contest.
type ContestClass struct {
	ID             int64
	Name           string
	Session        *string
	StartTime      *time.Time
	RemainingSlots *int
	Price          int64
	PrizeHighlight string
}

// ContestRule adalah aturan lomba pada detail contest.
type ContestRule struct {
	ID          int64
	Title       string
	Description *string
	Severity    string
	SortOrder   int
}
