package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type fakeRepository struct {
	users     map[string]User
	createErr error
	findErr   error
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{users: make(map[string]User)}
}

func (f *fakeRepository) Create(_ context.Context, name, identifier, passwordHash string) (User, error) {
	if f.createErr != nil {
		return User{}, f.createErr
	}
	if _, exists := f.users[identifier]; exists {
		return User{}, ErrIdentifierTaken
	}
	user := User{
		ID:           int64(len(f.users) + 1),
		Name:         name,
		Identifier:   identifier,
		PasswordHash: passwordHash,
	}
	f.users[identifier] = user
	return user, nil
}

func (f *fakeRepository) FindByIdentifier(_ context.Context, identifier string) (User, error) {
	if f.findErr != nil {
		return User{}, f.findErr
	}
	user, exists := f.users[identifier]
	if !exists {
		return User{}, ErrNotFound
	}
	return user, nil
}

func newTestTokenManager() *TokenManager {
	return NewTokenManager([]byte("rahasia-untuk-test"), time.Hour)
}

func newTestService(repo Repository) (*Service, *TokenManager) {
	tokens := newTestTokenManager()
	return NewService(repo, tokens), tokens
}

type failingIssuer struct{}

func (failingIssuer) Issue(User) (string, error) {
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
