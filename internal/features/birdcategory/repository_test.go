package birdcategory

import (
	"context"
	"errors"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

type mockRepository struct {
	listFunc func(ctx context.Context) ([]entities.BirdCategory, error)
	findFunc func(ctx context.Context, id int64) (entities.BirdCategory, error)
}

func (m *mockRepository) List(ctx context.Context) ([]entities.BirdCategory, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) FindByID(ctx context.Context, id int64) (entities.BirdCategory, error) {
	if m.findFunc != nil {
		return m.findFunc(ctx, id)
	}
	return entities.BirdCategory{}, errors.New("not implemented")
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

func TestRepositoryFindByID(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		repo := &mockRepository{
			findFunc: func(ctx context.Context, id int64) (entities.BirdCategory, error) {
				return entities.BirdCategory{ID: id, Name: "Murai Batu"}, nil
			},
		}
		got, err := repo.FindByID(context.Background(), 1)
		if err != nil || got.ID != 1 {
			t.Errorf("FindByID() = %+v, %v", got, err)
		}
	})

	t.Run("tidak ditemukan", func(t *testing.T) {
		repo := &mockRepository{
			findFunc: func(ctx context.Context, id int64) (entities.BirdCategory, error) {
				return entities.BirdCategory{}, apperrors.ErrNotFound
			},
		}
		_, err := repo.FindByID(context.Background(), 99)
		if err == nil {
			t.Error("FindByID() ingin error not found")
		}
	})

	t.Run("kesalahan database", func(t *testing.T) {
		repo := &mockRepository{
			findFunc: func(ctx context.Context, id int64) (entities.BirdCategory, error) {
				return entities.BirdCategory{}, errors.New("db down")
			},
		}
		if _, err := repo.FindByID(context.Background(), 1); err == nil {
			t.Error("FindByID() ingin error database")
		}
	})
}
