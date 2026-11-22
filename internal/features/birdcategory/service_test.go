package birdcategory

import (
	"context"
	"errors"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

type fakeRepository struct {
	listResult []entities.BirdCategory
	listErr    error
	findResult entities.BirdCategory
	findErr    error
}

func (f *fakeRepository) List(ctx context.Context) ([]entities.BirdCategory, error) {
	return f.listResult, f.listErr
}

func (f *fakeRepository) FindByID(ctx context.Context, id int64) (entities.BirdCategory, error) {
	return f.findResult, f.findErr
}

func TestServiceList(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		svc := NewService(&fakeRepository{listResult: []entities.BirdCategory{{ID: 1, Name: "Murai Batu"}}})
		got, err := svc.List(context.Background())
		if err != nil || len(got) != 1 {
			t.Errorf("List() = %+v, %v", got, err)
		}
	})

	t.Run("gagal", func(t *testing.T) {
		svc := NewService(&fakeRepository{listErr: errors.New("db down")})
		if _, err := svc.List(context.Background()); err == nil {
			t.Error("List() ingin error")
		}
	})
}

func TestServiceGetByID(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		svc := NewService(&fakeRepository{findResult: entities.BirdCategory{ID: 1, Name: "Murai Batu"}})
		got, err := svc.GetByID(context.Background(), 1)
		if err != nil || got.ID != 1 {
			t.Errorf("GetByID() = %+v, %v", got, err)
		}
	})

	t.Run("tidak ditemukan", func(t *testing.T) {
		svc := NewService(&fakeRepository{findErr: errors.New("not found")})
		if _, err := svc.GetByID(context.Background(), 99); err == nil {
			t.Error("GetByID() ingin error")
		}
	})
}
