package home

import (
	"testing"
	"time"
)

func TestGreeting(t *testing.T) {
	cases := []struct {
		hour int
		want string
	}{
		{hour: 0, want: greetingEvening},
		{hour: 4, want: greetingEvening},
		{hour: 5, want: greetingMorning},
		{hour: 10, want: greetingMorning},
		{hour: 11, want: greetingNoon},
		{hour: 14, want: greetingNoon},
		{hour: 15, want: greetingAfternoon},
		{hour: 17, want: greetingAfternoon},
		{hour: 18, want: greetingEvening},
		{hour: 23, want: greetingEvening},
	}
	for _, tc := range cases {
		now := time.Date(2026, 11, 22, tc.hour, 0, 0, 0, wibLocation)
		if got := greeting(now); got != tc.want {
			t.Errorf("greeting(jam %d WIB) = %q, ingin %q", tc.hour, got, tc.want)
		}
	}
}

func TestGreetingMengubahUTCKeWIB(t *testing.T) {
	now := time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)
	if got := greeting(now); got != greetingAfternoon {
		t.Errorf("greeting(10:00 UTC) = %q, ingin %q", got, greetingAfternoon)
	}
}
