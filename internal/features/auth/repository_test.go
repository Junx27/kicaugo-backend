package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

// mockRepository adalah implementasi Repository untuk testing.
type mockRepository struct {
	createFunc func(ctx context.Context, name, identifier, passwordHash string) (entities.User, error)
	findFunc   func(ctx context.Context, identifier string) (entities.User, error)
}

func (m *mockRepository) Create(ctx context.Context, name, identifier, passwordHash string) (entities.User, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, name, identifier, passwordHash)
	}
	return entities.User{}, errors.New("not implemented")
}

func (m *mockRepository) FindByIdentifier(ctx context.Context, identifier string) (entities.User, error) {
	if m.findFunc != nil {
		return m.findFunc(ctx, identifier)
	}
	return entities.User{}, errors.New("not implemented")
}

func TestRepositoryCreate(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		repo := &mockRepository{
			createFunc: func(ctx context.Context, name, identifier, passwordHash string) (entities.User, error) {
				return entities.User{ID: 1, Name: name, Identifier: identifier, PasswordHash: passwordHash}, nil
			},
		}
		user, err := repo.Create(context.Background(), "Kicau Mania", "test@gmail.com", "hash")
		if err != nil || user.ID != 1 {
			t.Errorf("Create() = %+v, %v", user, err)
		}
	})

	t.Run("identifier duplikat", func(t *testing.T) {
		repo := &mockRepository{
			createFunc: func(ctx context.Context, name, identifier, passwordHash string) (entities.User, error) {
				return entities.User{}, apperrors.ErrIdentifierTaken
			},
		}
		_, err := repo.Create(context.Background(), "Kicau Mania", "test@gmail.com", "hash")
		if !errors.Is(err, apperrors.ErrIdentifierTaken) {
			t.Errorf("Create() error = %v, ingin ErrIdentifierTaken", err)
		}
	})

	t.Run("kesalahan database", func(t *testing.T) {
		repo := &mockRepository{
			createFunc: func(ctx context.Context, name, identifier, passwordHash string) (entities.User, error) {
				return entities.User{}, errors.New("db down")
			},
		}
		if _, err := repo.Create(context.Background(), "Kicau Mania", "test@gmail.com", "hash"); err == nil {
			t.Error("Create() ingin error database")
		}
	})
}

func TestRepositoryFindByIdentifier(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		repo := &mockRepository{
			findFunc: func(ctx context.Context, identifier string) (entities.User, error) {
				return entities.User{ID: 1, Name: "Kicau Mania", Identifier: identifier, PasswordHash: "hash"}, nil
			},
		}
		user, err := repo.FindByIdentifier(context.Background(), "test@gmail.com")
		if err != nil || user.Identifier != "test@gmail.com" {
			t.Errorf("FindByIdentifier() = %+v, %v", user, err)
		}
	})

	t.Run("tidak ditemukan", func(t *testing.T) {
		repo := &mockRepository{
			findFunc: func(ctx context.Context, identifier string) (entities.User, error) {
				return entities.User{}, apperrors.ErrNotFound
			},
		}
		_, err := repo.FindByIdentifier(context.Background(), "test@gmail.com")
		if !errors.Is(err, apperrors.ErrNotFound) {
			t.Errorf("FindByIdentifier() error = %v, ingin ErrNotFound", err)
		}
	})

	t.Run("kesalahan database", func(t *testing.T) {
		repo := &mockRepository{
			findFunc: func(ctx context.Context, identifier string) (entities.User, error) {
				return entities.User{}, errors.New("db down")
			},
		}
		if _, err := repo.FindByIdentifier(context.Background(), "test@gmail.com"); err == nil {
			t.Error("FindByIdentifier() ingin error database")
		}
	})
}
