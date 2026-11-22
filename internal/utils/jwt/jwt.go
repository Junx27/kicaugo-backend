package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims adalah klaim token yang dipakai lintas handler.
type Claims struct {
	jwt.RegisteredClaims
	Name string `json:"name"`
}

// TokenManager menerbitkan dan memvalidasi JWT HS256.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

// NewTokenManager membuat pengelola token dengan secret dan masa berlaku tertentu.
func NewTokenManager(secret []byte, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: secret, ttl: ttl}
}

// Issue membuat token untuk pengguna.
func (m *TokenManager) Issue(user User) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Identifier,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
		Name: user.Name,
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse memverifikasi token dan mengembalikan identitas pengguna.
func (m *TokenManager) Parse(raw string) (User, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid || claims.Subject == "" {
		return User{}, ErrInvalidToken
	}
	return User{Identifier: claims.Subject, Name: claims.Name}, nil
}
