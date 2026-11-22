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
}

func (f *fakeRepository) List(ctx context.Context) ([]entities.BirdCategory, error) {
	return f.listResult, f.listErr
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

func (f *fakeRepository) Delete(ctx context.Context, id int64) error {
	return nil
}
