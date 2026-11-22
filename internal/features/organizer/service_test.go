package organizer

import (
	"context"
	"errors"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

type fakeRepository struct {
	userID        int64
	resolveErr    error
	created       entities.Organizer
	createErr     error
	member        entities.EventOrganizerUser
	addMemberErr  error
	active        bool
	activeErr     error
	hasActive     bool
	hasActiveErr  error
	addMemberRole string
}

func (f *fakeRepository) ResolveUserID(ctx context.Context, identifier string) (int64, error) {
	return f.userID, f.resolveErr
}

func (f *fakeRepository) Create(ctx context.Context, in entities.Organizer) (entities.Organizer, error) {
	if f.createErr != nil {
		return entities.Organizer{}, f.createErr
	}
	if f.created.ID == 0 {
		f.created = in
		f.created.ID = 1
	}
	return f.created, nil
}

func (f *fakeRepository) AddMember(ctx context.Context, organizerID, userID int64, role string) (entities.EventOrganizerUser, error) {
	f.addMemberRole = role
	if f.addMemberErr != nil {
		return entities.EventOrganizerUser{}, f.addMemberErr
	}
	if f.member.ID == 0 {
		f.member = entities.EventOrganizerUser{ID: 10, OrganizerID: organizerID, UserID: userID, Role: role, IsActive: true}
	}
	return f.member, nil
}

func (f *fakeRepository) IsActiveMember(ctx context.Context, identifier string, organizerID int64) (bool, error) {
	return f.active, f.activeErr
}

func (f *fakeRepository) HasActiveMembership(ctx context.Context, identifier string) (bool, error) {
	return f.hasActive, f.hasActiveErr
}

func TestServiceCreate(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		repo := &fakeRepository{userID: 7}
		svc := NewService(repo)
		got, err := svc.Create(context.Background(), "user@example.com", entities.Organizer{Name: " EO ", Slug: " eo-jaya ", Email: "EO@EXAMPLE.COM"})
		if err != nil || got.ID == 0 || repo.addMemberRole != "owner" {
			t.Fatalf("Create() = %+v, %v, role=%s", got, err, repo.addMemberRole)
		}
	})

	t.Run("user tidak ditemukan", func(t *testing.T) {
		svc := NewService(&fakeRepository{resolveErr: apperrors.ErrNotFound})
		if _, err := svc.Create(context.Background(), "x", entities.Organizer{Name: "EO", Slug: "eo", Email: "eo@example.com"}); !errors.Is(err, apperrors.ErrNotFound) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("payload invalid", func(t *testing.T) {
		svc := NewService(&fakeRepository{userID: 1})
		if _, err := svc.Create(context.Background(), "x", entities.Organizer{Name: "", Slug: "eo", Email: "eo@example.com"}); !errors.Is(err, apperrors.ErrInvalidPayload) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("create repo error", func(t *testing.T) {
		svc := NewService(&fakeRepository{userID: 1, createErr: errors.New("db down")})
		if _, err := svc.Create(context.Background(), "x", entities.Organizer{Name: "EO", Slug: "eo", Email: "eo@example.com"}); err == nil {
			t.Fatal("ingin error")
		}
	})

	t.Run("add owner member error", func(t *testing.T) {
		svc := NewService(&fakeRepository{userID: 1, addMemberErr: errors.New("db down")})
		if _, err := svc.Create(context.Background(), "x", entities.Organizer{Name: "EO", Slug: "eo", Email: "eo@example.com"}); err == nil {
			t.Fatal("ingin error")
		}
	})
}

func TestServiceRegisterMember(t *testing.T) {
	t.Run("berhasil default staff", func(t *testing.T) {
		repo := &fakeRepository{userID: 9}
		svc := NewService(repo)
		got, err := svc.RegisterMember(context.Background(), "user@example.com", 2, "")
		if err != nil || got.Role != "staff" {
			t.Fatalf("RegisterMember() = %+v, %v", got, err)
		}
	})

	t.Run("role invalid", func(t *testing.T) {
		svc := NewService(&fakeRepository{userID: 1})
		if _, err := svc.RegisterMember(context.Background(), "x", 2, "boss"); !errors.Is(err, apperrors.ErrInvalidPayload) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("organizer id invalid", func(t *testing.T) {
		svc := NewService(&fakeRepository{userID: 1})
		if _, err := svc.RegisterMember(context.Background(), "x", 0, "staff"); !errors.Is(err, apperrors.ErrInvalidPayload) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("resolve error", func(t *testing.T) {
		svc := NewService(&fakeRepository{resolveErr: apperrors.ErrNotFound})
		if _, err := svc.RegisterMember(context.Background(), "x", 1, "staff"); !errors.Is(err, apperrors.ErrNotFound) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestServiceMembershipChecks(t *testing.T) {
	svc := NewService(&fakeRepository{hasActive: true, active: true})
	has, err := svc.HasActiveMembership(context.Background(), "user@example.com")
	if err != nil || !has {
		t.Fatalf("HasActiveMembership() = %v, %v", has, err)
	}
	active, err := svc.IsActiveMember(context.Background(), "user@example.com", 1)
	if err != nil || !active {
		t.Fatalf("IsActiveMember() = %v, %v", active, err)
	}
	if _, err := svc.IsActiveMember(context.Background(), "user@example.com", 0); !errors.Is(err, apperrors.ErrInvalidPayload) {
		t.Fatalf("err = %v", err)
	}
}
