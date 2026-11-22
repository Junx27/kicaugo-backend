package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

const minPasswordLength = 8

// TokenIssuer menerbitkan token untuk pengguna yang sudah tervalidasi.
type TokenIssuer interface {
	Issue(subject, name string) (string, error)
}

// Service berisi aturan bisnis autentikasi.
type Service struct {
	repo   Repository
	tokens TokenIssuer
}

// NewService membuat service dengan repository dan pengelola token.
func NewService(repo Repository, tokens TokenIssuer) *Service {
	return &Service{repo: repo, tokens: tokens}
}

// Register membuat akun, lalu menerbitkan token.
func (s *Service) Register(ctx context.Context, name, identifier, password string) (entities.AuthResult, error) {
	name = strings.TrimSpace(name)
	identifier = normalizeIdentifier(identifier)
	if name == "" || identifier == "" || len(password) < minPasswordLength {
		return entities.AuthResult{}, apperrors.ErrInvalidPayload
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return entities.AuthResult{}, fmt.Errorf("hash password: %w", err)
	}
	user, err := s.repo.Create(ctx, name, identifier, string(hash))
	if err != nil {
		return entities.AuthResult{}, fmt.Errorf("register %s: %w", identifier, err)
	}
	return s.issue(user)
}

// Login memverifikasi kredensial, lalu menerbitkan token.
func (s *Service) Login(ctx context.Context, identifier, password string) (entities.AuthResult, error) {
	user, err := s.repo.FindByIdentifier(ctx, normalizeIdentifier(identifier))
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return entities.AuthResult{}, apperrors.ErrInvalidCredentials
		}
		return entities.AuthResult{}, fmt.Errorf("login %s: %w", identifier, err)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return entities.AuthResult{}, apperrors.ErrInvalidCredentials
	}
	return s.issue(user)
}

// Profile mengambil pengguna berdasarkan identifier dari token.
func (s *Service) Profile(ctx context.Context, identifier string) (entities.User, error) {
	return s.repo.FindByIdentifier(ctx, identifier)
}

func (s *Service) issue(user entities.User) (entities.AuthResult, error) {
	token, err := s.tokens.Issue(user.Identifier, user.Name)
	if err != nil {
		return entities.AuthResult{}, fmt.Errorf("issue token: %w", err)
	}
	return entities.AuthResult{User: user, Token: token}, nil
}

func normalizeIdentifier(identifier string) string {
	return strings.ToLower(strings.TrimSpace(identifier))
}

// scaffold:service-methods
