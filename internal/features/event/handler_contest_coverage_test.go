package event

import (
	"testing"
)

func TestToContestDetailDataBranchesCoverage(t *testing.T) {
	base := sampleContestDetail()

	t.Run("remaining negative", func(t *testing.T) {
		d := base
		d.TotalSlots = 1
		d.ActiveTickets = 5
		_ = toContestDetailData(d)
	})
	t.Run("empty classes", func(t *testing.T) {
		d := base
		d.Classes = nil
		out := toContestDetailData(d)
		if out.Ticket.StartingPrice != "Rp 0" {
			t.Fatalf("starting price = %s", out.Ticket.StartingPrice)
		}
	})
	t.Run("gratis price", func(t *testing.T) {
		d := base
		d.Classes[0].Price = 0
		out := toContestDetailData(d)
		if out.Ticket.StartingPrice != "Gratis" {
			t.Fatalf("starting price = %s", out.Ticket.StartingPrice)
		}
	})
	t.Run("nil start time", func(t *testing.T) {
		d := base
		d.Classes[0].StartTime = nil
		out := toContestDetailData(d)
		if out.Classes[0].Time != "" {
			t.Fatalf("time = %q", out.Classes[0].Time)
		}
	})
	t.Run("nil schedule date", func(t *testing.T) {
		d := base
		d.ScheduleDate = nil
		out := toContestDetailData(d)
		if out.Schedule.Date != "" || out.ScheduleText != "" {
			t.Fatalf("date = %q", out.Schedule.Date)
		}
	})
	t.Run("nil prize pool", func(t *testing.T) {
		d := base
		d.PrizePool = nil
		out := toContestDetailData(d)
		if out.Prize.Total != "-" {
			t.Fatalf("prize total = %s", out.Prize.Total)
		}
	})
	t.Run("nil address", func(t *testing.T) {
		d := base
		d.Address = nil
		out := toContestDetailData(d)
		if out.Location.Address != "" {
			t.Fatalf("address = %q", out.Location.Address)
		}
	})
	t.Run("status other", func(t *testing.T) {
		d := base
		d.Status = "completed"
		out := toContestDetailData(d)
		if out.Status != "completed" {
			t.Fatalf("status = %s", out.Status)
		}
	})
	t.Run("nil rule description", func(t *testing.T) {
		d := base
		d.Rules[0].Description = nil
		out := toContestDetailData(d)
		if out.Rules[0].Description != nil {
			t.Fatal("description not nil")
		}
	})
	t.Run("nil remaining slots", func(t *testing.T) {
		d := base
		d.Classes[0].RemainingSlots = nil
		_ = toContestDetailData(d)
	})
}
