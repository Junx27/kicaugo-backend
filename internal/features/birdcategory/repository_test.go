package birdcategory

import (
	"context"
	"errors"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

type mockRepository struct {
	listFunc func(ctx context.Context) ([]entities.BirdCategory, error)
}

func (m *mockRepository) List(ctx context.Context) ([]entities.BirdCategory, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func TestRepositoryList(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		repo := &mockRepository{
			listFunc: func(ctx context.Context) ([]entities.BirdCategory, error) {
				return []entities.BirdCategory{{ID: 1, Name: "Murai Batu"}}, nil
			},
		}
		got, err := repo.List(context.Background())
		if err != nil || len(got) != 1 {
			t.Errorf("List() = %+v, %v", got, err)
		}
	})

	t.Run("kesalahan database", func(t *testing.T) {
		repo := &mockRepository{
			listFunc: func(ctx context.Context) ([]entities.BirdCategory, error) {
				return nil, errors.New("db down")
			},
		}
		if _, err := repo.List(context.Background()); err == nil {
			t.Error("List() ingin error database")
		}
	})
}

func (m *mockRepository) Delete(ctx context.Context, id int64) error {
	return nil
}
