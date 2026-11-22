package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
)

type fakeRepository struct {
	users     map[string]entities.User
	createErr error
	findErr   error
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{users: make(map[string]entities.User)}
}

func (f *fakeRepository) Create(_ context.Context, name, identifier, passwordHash string) (entities.User, error) {
	if f.createErr != nil {
		return entities.User{}, f.createErr
	}
	if _, exists := f.users[identifier]; exists {
		return entities.User{}, apperrors.ErrIdentifierTaken
	}
	user := entities.User{
		ID:           int64(len(f.users) + 1),
		Name:         name,
		Identifier:   identifier,
		PasswordHash: passwordHash,
	}
	f.users[identifier] = user
	return user, nil
}

func (f *fakeRepository) FindByIdentifier(_ context.Context, identifier string) (entities.User, error) {
	if f.findErr != nil {
		return entities.User{}, f.findErr
	}
	user, exists := f.users[identifier]
	if !exists {
		return entities.User{}, apperrors.ErrNotFound
	}
	return user, nil
}

func newTestTokenManager() *jwt.Manager {
	return jwt.NewManager([]byte("rahasia-untuk-test"), time.Hour)
}

func newTestService(repo Repository) (*Service, *jwt.Manager) {
	tokens := newTestTokenManager()
	return NewService(repo, tokens), tokens
}

type failingIssuer struct{}

func (failingIssuer) Issue(string, string) (string, error) {
	return "", errors.New("issuer down")
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	return string(hash)
}
