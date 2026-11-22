package jwt

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"

	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

func TestManagerIssueAndParse(t *testing.T) {
	manager := newTestManager()

	raw, err := manager.Issue("test@gmail.com", "Kicau Mania")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	got, err := manager.Parse(raw)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got.Name != "Kicau Mania" || got.Subject != "test@gmail.com" {
		t.Errorf("Parse() = %+v", got)
	}
}

func TestManagerParseFailures(t *testing.T) {
	t.Run("token rusak", func(t *testing.T) {
		if _, err := newTestManager().Parse("bukan.jwt.valid"); err != apperrors.ErrInvalidToken {
			t.Errorf("Parse() error = %v, ingin ErrInvalidToken", err)
		}
	})

	t.Run("secret berbeda", func(t *testing.T) {
		raw, err := NewManager([]byte("secret-satu"), time.Hour).Issue("test@gmail.com", "Kicau Mania")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := NewManager([]byte("secret-dua"), time.Hour).Parse(raw); err != apperrors.ErrInvalidToken {
			t.Errorf("Parse() error = %v, ingin ErrInvalidToken", err)
		}
	})

	t.Run("metode tanda tangan lain", func(t *testing.T) {
		claims := Claims{RegisteredClaims: jwtlib.RegisteredClaims{Subject: "test@gmail.com"}, Name: "Kicau Mania"}
		raw, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS512, claims).SignedString([]byte("rahasia-untuk-test"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := newTestManager().Parse(raw); err != apperrors.ErrInvalidToken {
			t.Errorf("Parse() error = %v, ingin ErrInvalidToken", err)
		}
	})

	t.Run("kedaluwarsa", func(t *testing.T) {
		expired := NewManager([]byte("rahasia-untuk-test"), -time.Hour)
		raw, err := expired.Issue("test@gmail.com", "Kicau Mania")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := expired.Parse(raw); err != apperrors.ErrInvalidToken {
			t.Errorf("Parse() error = %v, ingin ErrInvalidToken", err)
		}
	})

	t.Run("subjek kosong", func(t *testing.T) {
		claims := Claims{RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   "",
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
		}}
		raw, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString([]byte("rahasia-untuk-test"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := newTestManager().Parse(raw); err != apperrors.ErrInvalidToken {
			t.Errorf("Parse() error = %v, ingin ErrInvalidToken", err)
		}
	})
}

func newTestManager() *Manager {
	return NewManager([]byte("rahasia-untuk-test"), time.Hour)
}
