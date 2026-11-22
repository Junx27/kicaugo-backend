// Package jwt mengelola penerbitan dan validasi token JWT HS256.
package jwt

import (
	"errors"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"

	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

// Claims adalah klaim token yang dipakai lintas handler.
type Claims struct {
	jwtlib.RegisteredClaims
	Name string `json:"name"`
}

// Manager menerbitkan dan memvalidasi JWT HS256.
type Manager struct {
	secret []byte
	ttl    time.Duration
}

// NewManager membuat pengelola token dengan secret dan masa berlaku tertentu.
func NewManager(secret []byte, ttl time.Duration) *Manager {
	return &Manager{secret: secret, ttl: ttl}
}

// Issue membuat token untuk subject dan name.
func (m *Manager) Issue(subject, name string) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(m.ttl)),
		},
		Name: name,
	}
	return jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse memverifikasi token dan mengembalikan klaim.
func (m *Manager) Parse(raw string) (Claims, error) {
	claims := &Claims{}
	token, err := jwtlib.ParseWithClaims(raw, claims, func(token *jwtlib.Token) (any, error) {
		if token.Method != jwtlib.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid || claims.Subject == "" {
		return Claims{}, apperrors.ErrInvalidToken
	}
	return *claims, nil
}
