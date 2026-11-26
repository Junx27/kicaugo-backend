package event

import (
	"context"
	"errors"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

func TestServiceCountGantanganCoverage(t *testing.T) {
	got, err := NewService(&fakeRepository{gantanganCount: 7}).CountGantangan(context.Background())
	if err != nil || got != 7 {
		t.Fatalf("CountGantangan() = %d, %v", got, err)
	}
}

func TestServiceListNearbySchedulesCoverage(t *testing.T) {
	t.Run("sukses", func(t *testing.T) {
		repo := &fakeRepository{candidate: freshCandidate(), schedules: []entities.EventSchedule{{ID: 1, Title: "S"}}}
		got, err := NewService(repo).ListNearbySchedules(context.Background(), "user@example.com", 0, 0, 10)
		if err != nil || len(got) != 1 {
			t.Fatalf("ListNearbySchedules() = %v, %v", got, err)
		}
	})
	t.Run("tanpa origin", func(t *testing.T) {
		repo := &fakeRepository{candidate: entities.OriginCandidate{RadiusKM: 50}}
		got, err := NewService(repo).ListNearbySchedules(context.Background(), "user@example.com", 0, 0, 10)
		if err != nil || len(got) != 0 {
			t.Fatalf("ListNearbySchedules() = %v, %v", got, err)
		}
	})
	t.Run("radius invalid", func(t *testing.T) {
		repo := &fakeRepository{candidate: freshCandidate()}
		if _, err := NewService(repo).ListNearbySchedules(context.Background(), "user@example.com", maxRadiusKM+1, 0, 10); err == nil {
			t.Fatal("ingin error")
		}
	})
	t.Run("limit invalid", func(t *testing.T) {
		repo := &fakeRepository{candidate: freshCandidate()}
		if _, err := NewService(repo).ListNearbySchedules(context.Background(), "user@example.com", 0, 0, 0); err == nil {
			t.Fatal("ingin error")
		}
	})
	t.Run("origin error", func(t *testing.T) {
		repo := &fakeRepository{originErr: errors.New("db down")}
		if _, err := NewService(repo).ListNearbySchedules(context.Background(), "user@example.com", 0, 0, 10); err == nil {
			t.Fatal("ingin error")
		}
	})
	t.Run("list error", func(t *testing.T) {
		repo := &fakeRepository{candidate: freshCandidate(), listErr: errors.New("db down")}
		if _, err := NewService(repo).ListNearbySchedules(context.Background(), "user@example.com", 0, 0, 10); err == nil {
			t.Fatal("ingin error")
		}
	})
}
