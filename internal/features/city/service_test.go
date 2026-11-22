package city

import (
	"context"
	"errors"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

type fakeRepository struct {
	listResult []entities.City
	listErr    error
}

func (f *fakeRepository) List(ctx context.Context) ([]entities.City, error) {
	return f.listResult, f.listErr
}

func (f *fakeRepository) FindByID(ctx context.Context, id int64) (entities.City, error) {
	return entities.City{}, errors.New("not implemented")
}

func (f *fakeRepository) Create(ctx context.Context, in entities.City) (entities.City, error) {
	return entities.City{}, errors.New("not implemented")
}

func (f *fakeRepository) Update(ctx context.Context, id int64, in entities.City) (entities.City, error) {
	return entities.City{}, errors.New("not implemented")
}

func (f *fakeRepository) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func TestServiceList(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		svc := NewService(&fakeRepository{listResult: []entities.City{entities.City{ID: 1}}})
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
