package auth

import (
	"context"
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

func TestServiceRegister(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		repo := newFakeRepository()
		service, _ := newTestService(repo)

		result, err := service.Register(context.Background(), " Kicau Mania ", " Test@Gmail.com ", "test123@")
		if err != nil {
			t.Fatalf("Register() error = %v", err)
		}
		if result.Token == "" || result.User.Identifier != "test@gmail.com" {
			t.Errorf("result = %+v", result)
		}
		stored := repo.users["test@gmail.com"]
		if stored.Name != "Kicau Mania" {
			t.Errorf("nama tersimpan = %q", stored.Name)
		}
		if bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("test123@")) != nil {
			t.Error("password tidak tersimpan sebagai hash bcrypt")
		}
	})

	t.Run("payload invalid", func(t *testing.T) {
		service, _ := newTestService(newFakeRepository())
		if _, err := service.Register(context.Background(), "Kicau Mania", "test@gmail.com", "pendek"); err != apperrors.ErrInvalidPayload {
			t.Errorf("Register() error = %v, ingin ErrInvalidPayload", err)
		}
	})

	t.Run("identifier sudah dipakai", func(t *testing.T) {
		repo := newFakeRepository()
		service, _ := newTestService(repo)
		if _, err := service.Register(context.Background(), "Kicau Mania", "test@gmail.com", "test123@"); err != nil {
			t.Fatal(err)
		}

		_, err := service.Register(context.Background(), "Kicau Lain", "test@gmail.com", "test123@")
		if !errors.Is(err, apperrors.ErrIdentifierTaken) {
			t.Errorf("Register() error = %v, ingin ErrIdentifierTaken", err)
		}
	})

	t.Run("password terlalu panjang untuk bcrypt", func(t *testing.T) {
		service, _ := newTestService(newFakeRepository())
		if _, err := service.Register(context.Background(), "Kicau Mania", "test@gmail.com", strings.Repeat("a", 73)); err == nil {
			t.Error("Register() ingin error untuk password > 72 byte")
		}
	})

	t.Run("gagal menerbitkan token", func(t *testing.T) {
		service := NewService(newFakeRepository(), failingIssuer{})
		if _, err := service.Register(context.Background(), "Kicau Mania", "test@gmail.com", "test123@"); err == nil {
			t.Error("Register() ingin error saat token gagal diterbitkan")
		}
	})
}

func TestServiceLogin(t *testing.T) {
	repo := newFakeRepository()
	repo.users["test@gmail.com"] = entities.User{
		Name:         "Kicau Mania",
		Identifier:   "test@gmail.com",
		PasswordHash: hashPassword(t, "test123@"),
	}
	service, _ := newTestService(repo)

	t.Run("berhasil", func(t *testing.T) {
		result, err := service.Login(context.Background(), " TEST@GMAIL.COM ", "test123@")
		if err != nil {
			t.Fatalf("Login() error = %v", err)
		}
		if result.Token == "" || result.User.Name != "Kicau Mania" {
			t.Errorf("result = %+v", result)
		}
	})

	t.Run("password salah", func(t *testing.T) {
		if _, err := service.Login(context.Background(), "test@gmail.com", "salah123@"); err != apperrors.ErrInvalidCredentials {
			t.Errorf("Login() error = %v, ingin ErrInvalidCredentials", err)
		}
	})

	t.Run("pengguna tidak ada", func(t *testing.T) {
		if _, err := service.Login(context.Background(), "tidak@gmail.com", "test123@"); err != apperrors.ErrInvalidCredentials {
			t.Errorf("Login() error = %v, ingin ErrInvalidCredentials", err)
		}
	})

	t.Run("gagal membaca database", func(t *testing.T) {
		repo.findErr = errors.New("db down")
		t.Cleanup(func() { repo.findErr = nil })
		if _, err := service.Login(context.Background(), "test@gmail.com", "test123@"); err == nil {
			t.Error("Login() ingin error saat database gagal")
		}
	})

	t.Run("gagal menerbitkan token", func(t *testing.T) {
		failing := NewService(repo, failingIssuer{})
		if _, err := failing.Login(context.Background(), "test@gmail.com", "test123@"); err == nil {
			t.Error("Login() ingin error saat token gagal diterbitkan")
		}
	})
}

func TestServiceProfile(t *testing.T) {
	repo := newFakeRepository()
	repo.users["test@gmail.com"] = entities.User{Name: "Kicau Mania", Identifier: "test@gmail.com"}
	service, _ := newTestService(repo)

	user, err := service.Profile(context.Background(), "test@gmail.com")
	if err != nil || user.Name != "Kicau Mania" {
		t.Errorf("Profile() = %+v, %v", user, err)
	}
}
