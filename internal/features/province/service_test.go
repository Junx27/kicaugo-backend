package province

import (
	"context"
	"errors"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

type fakeRepository struct {
	listResult []entities.Province
	listErr    error
}

func (f *fakeRepository) List(ctx context.Context) ([]entities.Province, error) {
	return f.listResult, f.listErr
}

func (f *fakeRepository) FindByID(ctx context.Context, id int64) (entities.Province, error) {
	return entities.Province{}, errors.New("not implemented")
}

func (f *fakeRepository) Create(ctx context.Context, in entities.Province) (entities.Province, error) {
	return entities.Province{}, errors.New("not implemented")
}

func (f *fakeRepository) Update(ctx context.Context, id int64, in entities.Province) (entities.Province, error) {
	return entities.Province{}, errors.New("not implemented")
}

func (f *fakeRepository) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func TestServiceList(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		svc := NewService(&fakeRepository{listResult: []entities.Province{entities.Province{ID: 1}}})
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
