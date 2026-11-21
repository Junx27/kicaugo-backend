package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const minPasswordLength = 8

// Result adalah hasil autentikasi yang dikirim ke klien.
type Result struct {
	User  User
	Token string
}

// TokenIssuer menerbitkan token untuk pengguna yang sudah tervalidasi.
type TokenIssuer interface {
	Issue(user User) (string, error)
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
func (s *Service) Register(ctx context.Context, name, identifier, password string) (Result, error) {
	name = strings.TrimSpace(name)
	identifier = normalizeIdentifier(identifier)
	if name == "" || identifier == "" || len(password) < minPasswordLength {
		return Result{}, ErrInvalidPayload
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Result{}, fmt.Errorf("hash password: %w", err)
	}
	user, err := s.repo.Create(ctx, name, identifier, string(hash))
	if err != nil {
		return Result{}, fmt.Errorf("register %s: %w", identifier, err)
	}
	return s.issue(user)
}

// Login memverifikasi kredensial, lalu menerbitkan token.
func (s *Service) Login(ctx context.Context, identifier, password string) (Result, error) {
	user, err := s.repo.FindByIdentifier(ctx, normalizeIdentifier(identifier))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Result{}, ErrInvalidCredentials
		}
		return Result{}, fmt.Errorf("login %s: %w", identifier, err)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return Result{}, ErrInvalidCredentials
	}
	return s.issue(user)
}

// Profile mengambil pengguna berdasarkan identifier dari token.
func (s *Service) Profile(ctx context.Context, identifier string) (User, error) {
	return s.repo.FindByIdentifier(ctx, identifier)
}

func (s *Service) issue(user User) (Result, error) {
	token, err := s.tokens.Issue(user)
	if err != nil {
		return Result{}, fmt.Errorf("issue token: %w", err)
	}
	return Result{User: user, Token: token}, nil
}

func normalizeIdentifier(identifier string) string {
	return strings.ToLower(strings.TrimSpace(identifier))
}
