package event

import (
	"context"
	"errors"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

func TestServiceGetContestDetailCoverage(t *testing.T) {
	t.Run("sukses", func(t *testing.T) {
		want := entities.ContestDetail{ID: 1, Title: "Contest"}
		got, err := NewService(&fakeRepository{contestDetail: want}).GetContestDetail(context.Background(), 1)
		if err != nil || got.ID != 1 || got.Title != "Contest" {
			t.Fatalf("GetContestDetail() = %+v, %v", got, err)
		}
	})
	t.Run("id invalid", func(t *testing.T) {
		if _, err := NewService(&fakeRepository{}).GetContestDetail(context.Background(), 0); err == nil {
			t.Fatal("ingin error")
		}
	})
	t.Run("repo error", func(t *testing.T) {
		if _, err := NewService(&fakeRepository{contestErr: errors.New("db down")}).GetContestDetail(context.Background(), 1); err == nil {
			t.Fatal("ingin error")
		}
	})
}
