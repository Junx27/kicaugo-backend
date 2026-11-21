package auth

import (
	"context"
	"errors"
	"testing"
)

// mockRepository adalah implementasi Repository untuk testing.
type mockRepository struct {
	createFunc func(ctx context.Context, name, identifier, passwordHash string) (User, error)
	findFunc   func(ctx context.Context, identifier string) (User, error)
}

func (m *mockRepository) Create(ctx context.Context, name, identifier, passwordHash string) (User, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, name, identifier, passwordHash)
	}
	return User{}, errors.New("not implemented")
}

func (m *mockRepository) FindByIdentifier(ctx context.Context, identifier string) (User, error) {
	if m.findFunc != nil {
		return m.findFunc(ctx, identifier)
	}
	return User{}, errors.New("not implemented")
}

func TestSQLRepositoryCreate(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		repo := &mockRepository{
			createFunc: func(ctx context.Context, name, identifier, passwordHash string) (User, error) {
				return User{ID: 1, Name: name, Identifier: identifier, PasswordHash: passwordHash}, nil
			},
		}
		user, err := repo.Create(context.Background(), "Kicau Mania", "test@gmail.com", "hash")
		if err != nil || user.ID != 1 {
			t.Errorf("Create() = %+v, %v", user, err)
		}
	})

	t.Run("identifier duplikat", func(t *testing.T) {
		repo := &mockRepository{
			createFunc: func(ctx context.Context, name, identifier, passwordHash string) (User, error) {
				return User{}, ErrIdentifierTaken
			},
		}
		_, err := repo.Create(context.Background(), "Kicau Mania", "test@gmail.com", "hash")
		if !errors.Is(err, ErrIdentifierTaken) {
			t.Errorf("Create() error = %v, ingin ErrIdentifierTaken", err)
		}
	})

	t.Run("kesalahan database", func(t *testing.T) {
		repo := &mockRepository{
			createFunc: func(ctx context.Context, name, identifier, passwordHash string) (User, error) {
				return User{}, errors.New("db down")
			},
		}
		if _, err := repo.Create(context.Background(), "Kicau Mania", "test@gmail.com", "hash"); err == nil {
			t.Error("Create() ingin error database")
		}
	})
}

func TestSQLRepositoryFindByIdentifier(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		repo := &mockRepository{
			findFunc: func(ctx context.Context, identifier string) (User, error) {
				return User{ID: 1, Name: "Kicau Mania", Identifier: identifier, PasswordHash: "hash"}, nil
			},
		}
		user, err := repo.FindByIdentifier(context.Background(), "test@gmail.com")
		if err != nil || user.Identifier != "test@gmail.com" {
			t.Errorf("FindByIdentifier() = %+v, %v", user, err)
		}
	})

	t.Run("tidak ditemukan", func(t *testing.T) {
		repo := &mockRepository{
			findFunc: func(ctx context.Context, identifier string) (User, error) {
				return User{}, ErrNotFound
			},
		}
		_, err := repo.FindByIdentifier(context.Background(), "test@gmail.com")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("FindByIdentifier() error = %v, ingin ErrNotFound", err)
		}
	})

	t.Run("kesalahan database", func(t *testing.T) {
		repo := &mockRepository{
			findFunc: func(ctx context.Context, identifier string) (User, error) {
				return User{}, errors.New("db down")
			},
		}
		if _, err := repo.FindByIdentifier(context.Background(), "test@gmail.com"); err == nil {
			t.Error("FindByIdentifier() ingin error database")
		}
	})
}
