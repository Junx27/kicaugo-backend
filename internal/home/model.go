// Package home menyajikan konten Beranda sesuai entity Flutter.
package home

// Summary adalah ringkasan kecil di atas Beranda dan badge navigasi.
type Summary struct {
	Location          string `json:"location"`
	LiveCount         int    `json:"liveCount"`
	ActiveTicketCount int    `json:"activeTicketCount"`
}

// Attendees adalah inisial peserta dan jumlah sisanya.
type Attendees struct {
	Initials []string `json:"initials"`
	Others   int      `json:"others"`
}

// Featured adalah lomba unggulan terdekat.
type Featured struct {
	Title     string    `json:"title"`
	Schedule  string    `json:"schedule"`
	Distance  string    `json:"distance"`
	Prize     string    `json:"prize"`
	Slots     string    `json:"slots"`
	Attendees Attendees `json:"attendees"`
}

// Upcoming adalah lomba dengan sisa tiket terbatas.
type Upcoming struct {
	Type             string `json:"type"`
	Title            string `json:"title"`
	Time             string `json:"time"`
	Price            string `json:"price"`
	RemainingTickets int    `json:"remainingTickets"`
}

// Stats adalah ringkasan statistik komunitas.
type Stats struct {
	Badge        string `json:"badge"`
	Headline     string `json:"headline"`
	Note         string `json:"note"`
	SecurityNote string `json:"securityNote"`
}

// ActiveTicket adalah tiket aktif terdekat milik pengguna.
type ActiveTicket struct {
	Event    string `json:"event"`
	Schedule string `json:"schedule"`
	Number   string `json:"number"`
	Session  string `json:"session"`
}

// ScheduleItem adalah satu jadwal gantangan populer.
type ScheduleItem struct {
	Title    string `json:"title"`
	Venue    string `json:"venue"`
	Date     string `json:"date"`
	Price    string `json:"price"`
	Status   string `json:"status"`
	Category string `json:"category"`
}

// Content adalah seluruh data yang ditampilkan di Beranda.
type Content struct {
	Summary      Summary        `json:"summary"`
	Featured     Featured       `json:"featured"`
	Upcoming     Upcoming       `json:"upcoming"`
	Stats        Stats          `json:"stats"`
	ActiveTicket ActiveTicket   `json:"activeTicket"`
	Schedules    []ScheduleItem `json:"schedules"`
}
