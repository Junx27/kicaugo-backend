package home

import "time"

const (
	wibOffsetSeconds  = 7 * 60 * 60
	greetingMorning   = "Selamat pagi"
	greetingNoon      = "Selamat siang"
	greetingAfternoon = "Selamat sore"
	greetingEvening   = "Selamat malam"
)

// wibLocation adalah zona tetap agar tidak bergantung tzdata pada image distroless.
var wibLocation = time.FixedZone("WIB", wibOffsetSeconds)

// greetingsByHour memetakan jam 0..23 dalam WIB ke sapaan yang sesuai.
var greetingsByHour = [24]string{
	greetingEvening, greetingEvening, greetingEvening, greetingEvening, greetingEvening,
	greetingMorning, greetingMorning, greetingMorning, greetingMorning, greetingMorning, greetingMorning,
	greetingNoon, greetingNoon, greetingNoon, greetingNoon,
	greetingAfternoon, greetingAfternoon, greetingAfternoon,
	greetingEvening, greetingEvening, greetingEvening, greetingEvening, greetingEvening, greetingEvening,
}

// greeting mengembalikan sapaan berdasarkan jam setempat WIB.
func greeting(now time.Time) string {
	return greetingsByHour[now.In(wibLocation).Hour()]
}
