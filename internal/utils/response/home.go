package response

// HomeProfileData adalah ringkasan sederhana profil untuk halaman home.
type HomeProfileData struct {
	Name      string  `json:"name"`
	Avatar    *string `json:"avatar"`
	Message   string  `json:"message"`
	LiveCount int     `json:"live_count"`
}

// HomeOutput adalah output GET /api/v1/home/profile.
type HomeOutput struct {
	Status int `status:"200"`
	Body   Success[HomeProfileData]
}

// FeaturedAttendeesData adalah ringkasan peserta pada featured contest.
type FeaturedAttendeesData struct {
	Initials []string `json:"initials"`
	Others   int      `json:"others"`
}

// HomeFeaturedData adalah data demo featured contest Beranda.
type HomeFeaturedData struct {
	Title     string                `json:"title"`
	Schedule  string                `json:"schedule"`
	Distance  string                `json:"distance"`
	Prize     string                `json:"prize"`
	Slots     string                `json:"slots"`
	Attendees FeaturedAttendeesData `json:"attendees"`
}

// HomeFeaturedOutput adalah output GET /api/v1/home/featured.
type HomeFeaturedOutput struct {
	Status int `status:"200"`
	Body   Success[HomeFeaturedData]
}

// HomeUpcomingData adalah data demo upcoming contest Beranda.
type HomeUpcomingData struct {
	Type             string `json:"type"`
	Title            string `json:"title"`
	Time             string `json:"time"`
	Price            string `json:"price"`
	RemainingTickets int    `json:"remaining_tickets"`
}

// HomeUpcomingOutput adalah output GET /api/v1/home/upcoming.
type HomeUpcomingOutput struct {
	Status int `status:"200"`
	Body   Success[HomeUpcomingData]
}

// HomeStatsData adalah data demo statistik Beranda.
type HomeStatsData struct {
	Badge        string `json:"badge"`
	Headline     string `json:"headline"`
	Note         string `json:"note"`
	SecurityNote string `json:"security_note"`
}

// HomeStatsOutput adalah output GET /api/v1/home/stats.
type HomeStatsOutput struct {
	Status int `status:"200"`
	Body   Success[HomeStatsData]
}

// HomeActiveTicketData adalah data demo tiket aktif Beranda.
type HomeActiveTicketData struct {
	Event    string `json:"event"`
	Schedule string `json:"schedule"`
	Number   string `json:"number"`
	Session  string `json:"session"`
}

// HomeActiveTicketOutput adalah output GET /api/v1/home/active-ticket.
type HomeActiveTicketOutput struct {
	Status int `status:"200"`
	Body   Success[HomeActiveTicketData]
}

// HomeScheduleCategoryData adalah kategori burung pada item jadwal.
type HomeScheduleCategoryData struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// HomeScheduleData adalah item demo jadwal Beranda.
type HomeScheduleData struct {
	Title    string                   `json:"title"`
	Venue    string                   `json:"venue"`
	Date     string                   `json:"date"`
	Price    string                   `json:"price"`
	Status   string                   `json:"status"`
	Category HomeScheduleCategoryData `json:"category"`
}

// HomeSchedulesOutput adalah output GET /api/v1/home/schedules.
type HomeSchedulesOutput struct {
	Status int `status:"200"`
	Body   Success[[]HomeScheduleData]
}
