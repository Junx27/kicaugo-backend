package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenManagerIssueAndParse(t *testing.T) {
	tokens := newTestTokenManager()
	user := User{Name: "Kicau Mania", Identifier: "test@gmail.com"}

	raw, err := tokens.Issue(user)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	got, err := tokens.Parse(raw)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got.Name != user.Name || got.Identifier != user.Identifier {
		t.Errorf("Parse() = %+v, ingin %+v", got, user)
	}
}

func TestTokenManagerParseFailures(t *testing.T) {
	user := User{Name: "Kicau Mania", Identifier: "test@gmail.com"}

	t.Run("token rusak", func(t *testing.T) {
		if _, err := newTestTokenManager().Parse("bukan.jwt.valid"); err != ErrInvalidToken {
			t.Errorf("Parse() error = %v, ingin ErrInvalidToken", err)
		}
	})

	t.Run("secret berbeda", func(t *testing.T) {
		raw, err := NewTokenManager([]byte("secret-satu"), time.Hour).Issue(user)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := NewTokenManager([]byte("secret-dua"), time.Hour).Parse(raw); err != ErrInvalidToken {
			t.Errorf("Parse() error = %v, ingin ErrInvalidToken", err)
		}
	})

	t.Run("metode tanda tangan lain", func(t *testing.T) {
		claims := Claims{RegisteredClaims: jwt.RegisteredClaims{Subject: user.Identifier}, Name: user.Name}
		raw, err := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString([]byte("rahasia-untuk-test"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := newTestTokenManager().Parse(raw); err != ErrInvalidToken {
			t.Errorf("Parse() error = %v, ingin ErrInvalidToken", err)
		}
	})

	t.Run("kedaluwarsa", func(t *testing.T) {
		expired := NewTokenManager([]byte("rahasia-untuk-test"), -time.Hour)
		raw, err := expired.Issue(user)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := expired.Parse(raw); err != ErrInvalidToken {
			t.Errorf("Parse() error = %v, ingin ErrInvalidToken", err)
		}
	})

	t.Run("subjek kosong", func(t *testing.T) {
		claims := Claims{RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		}}
		raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("rahasia-untuk-test"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := newTestTokenManager().Parse(raw); err != ErrInvalidToken {
			t.Errorf("Parse() error = %v, ingin ErrInvalidToken", err)
		}
	})
}
