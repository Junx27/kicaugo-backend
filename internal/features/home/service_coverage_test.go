package home

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

func TestServiceGetSchedulesCoverage(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{nearbySchedErr: errors.New("db down")}, &fakeBirdCategoryReader{}, &fakeOrganizerReader{})
		if _, err := svc.GetSchedules(context.Background(), "user@example.com"); err == nil {
			t.Fatal("ingin error")
		}
	})
	t.Run("mapping tanggal dan kategori kosong", func(t *testing.T) {
		now := time.Date(2026, 11, 24, 0, 0, 0, 0, time.UTC)
		svc := NewService(&fakeProfileReader{}, &fakeEventReader{nearbySchedules: []entities.EventSchedule{{
			ID: 1, Title: "Latpres", Venue: "Arena", Date: &now, Price: 100000, Status: "available",
		}}}, &fakeBirdCategoryReader{}, &fakeOrganizerReader{})
		got, err := svc.GetSchedules(context.Background(), "user@example.com")
		if err != nil || len(got) != 1 {
			t.Fatalf("GetSchedules() = %v, %v", got, err)
		}
		if got[0].Date == "-" {
			t.Fatal("tanggal tidak terformat")
		}
		if got[0].Category.Name != "-" {
			t.Fatalf("kategori = %q", got[0].Category.Name)
		}
	})
}
