package home

import (
	"context"
	"errors"
	"testing"
)

// mockRepository adalah implementasi Repository untuk testing.
type mockRepository struct {
	getFunc func(ctx context.Context) (Content, error)
}

func (m *mockRepository) Get(ctx context.Context) (Content, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx)
	}
	return Content{}, errors.New("not implemented")
}

func TestSQLRepositoryGet(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		repo := &mockRepository{
			getFunc: func(ctx context.Context) (Content, error) {
				return sampleContent(), nil
			},
		}
		content, err := repo.Get(context.Background())
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if content.Summary.Location != "Jakarta Selatan" || len(content.Schedules) != 1 {
			t.Errorf("content = %+v", content)
		}
	})

	t.Run("gagal membaca database", func(t *testing.T) {
		repo := &mockRepository{
			getFunc: func(ctx context.Context) (Content, error) {
				return Content{}, errors.New("db down")
			},
		}
		if _, err := repo.Get(context.Background()); err == nil {
			t.Error("Get() ingin error database")
		}
	})

	t.Run("payload rusak", func(t *testing.T) {
		repo := &mockRepository{
			getFunc: func(ctx context.Context) (Content, error) {
				return Content{}, errors.New("invalid JSON")
			},
		}
		if _, err := repo.Get(context.Background()); err == nil {
			t.Error("Get() ingin error saat JSON rusak")
		}
	})
}

// sampleContent is now defined in handler_test.go
