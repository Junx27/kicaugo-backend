package event

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

func validRuleInput() entities.EventRuleInput {
	return entities.EventRuleInput{
		Title:       "Aturan Sangkar",
		Description: strPtr("Wajib sangkar standar"),
		Severity:    entities.RuleSeverityInfo,
		SortOrder:   1,
	}
}

func assertRuleError(t *testing.T, err error, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, ingin %v", err, want)
	}
}

func newAuthorizedRuleService(repo *fakeRepository) Service {
	repo.eventByID = entities.Event{ID: 1, OrganizerID: 7}
	repo.activeMember = true
	return NewService(repo)
}

func TestServiceListEventRules(t *testing.T) {
	ctx := context.Background()

	t.Run("sukses", func(t *testing.T) {
		repo := &fakeRepository{rules: []entities.ContestRule{{ID: 1, Title: "Aturan"}}}
		svc := NewService(repo)
		rules, err := svc.ListEventRules(ctx, 1)
		if err != nil || len(rules) != 1 {
			t.Fatalf("rules = %+v, error = %v", rules, err)
		}
	})

	t.Run("id invalid", func(t *testing.T) {
		svc := NewService(&fakeRepository{})
		_, err := svc.ListEventRules(ctx, 0)
		assertRuleError(t, err, apperrors.ErrInvalidPayload)
	})

	t.Run("error repository", func(t *testing.T) {
		svc := NewService(&fakeRepository{ruleErr: errors.New("db down")})
		_, err := svc.ListEventRules(ctx, 1)
		if err == nil {
			t.Fatal("ingin error")
		}
	})
}

func TestServiceCreateEventRule(t *testing.T) {
	ctx := context.Background()
	validCreate := func() entities.EventRuleCreate {
		return entities.EventRuleCreate{EventID: 1, EventRuleInput: validRuleInput()}
	}

	t.Run("sukses", func(t *testing.T) {
		repo := &fakeRepository{rule: entities.ContestRule{ID: 2, Title: "Aturan"}}
		rule, err := newAuthorizedRuleService(repo).CreateEventRule(ctx, "user@example.com", validCreate())
		if err != nil || rule.ID != 2 {
			t.Fatalf("rule = %+v, error = %v", rule, err)
		}
	})

	t.Run("id event invalid", func(t *testing.T) {
		in := validCreate()
		in.EventID = 0
		_, err := newAuthorizedRuleService(&fakeRepository{}).CreateEventRule(ctx, "user", in)
		assertRuleError(t, err, apperrors.ErrInvalidPayload)
	})

	t.Run("payload invalid", func(t *testing.T) {
		cases := []struct {
			name   string
			mutate func(*entities.EventRuleInput)
		}{
			{"title kosong", func(in *entities.EventRuleInput) { in.Title = "  " }},
			{"title terlalu pendek", func(in *entities.EventRuleInput) { in.Title = "ab" }},
			{"title terlalu panjang", func(in *entities.EventRuleInput) { in.Title = strings.Repeat("a", 151) }},
			{"deskripsi terlalu panjang", func(in *entities.EventRuleInput) { in.Description = strPtr(strings.Repeat("a", 1001)) }},
			{"severity invalid", func(in *entities.EventRuleInput) { in.Severity = "warning" }},
			{"urutan negatif", func(in *entities.EventRuleInput) { in.SortOrder = -1 }},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				in := validCreate()
				tc.mutate(&in.EventRuleInput)
				_, err := newAuthorizedRuleService(&fakeRepository{}).CreateEventRule(ctx, "user", in)
				assertRuleError(t, err, apperrors.ErrInvalidPayload)
			})
		}
	})

	t.Run("event tidak ditemukan", func(t *testing.T) {
		repo := &fakeRepository{eventByIDErr: apperrors.ErrNotFound}
		_, err := NewService(repo).CreateEventRule(ctx, "user", validCreate())
		assertRuleError(t, err, apperrors.ErrNotFound)
	})

	t.Run("cek keanggotaan error", func(t *testing.T) {
		repo := &fakeRepository{
			eventByID: entities.Event{OrganizerID: 7},
			activeErr: errors.New("db down"),
		}
		_, err := NewService(repo).CreateEventRule(ctx, "user", validCreate())
		if err == nil {
			t.Fatal("ingin error")
		}
	})

	t.Run("bukan anggota organizer", func(t *testing.T) {
		repo := &fakeRepository{eventByID: entities.Event{OrganizerID: 7}}
		_, err := NewService(repo).CreateEventRule(ctx, "user", validCreate())
		assertRuleError(t, err, apperrors.ErrForbidden)
	})

	t.Run("error repository", func(t *testing.T) {
		repo := &fakeRepository{ruleErr: errors.New("db down")}
		_, err := newAuthorizedRuleService(repo).CreateEventRule(ctx, "user", validCreate())
		if err == nil {
			t.Fatal("ingin error")
		}
	})
}

func TestServiceUpdateEventRule(t *testing.T) {
	ctx := context.Background()
	validUpdate := func() entities.EventRuleUpdate {
		return entities.EventRuleUpdate{EventID: 1, RuleID: 2, EventRuleInput: validRuleInput()}
	}

	t.Run("sukses", func(t *testing.T) {
		repo := &fakeRepository{rule: entities.ContestRule{ID: 2, Title: "Aturan"}}
		rule, err := newAuthorizedRuleService(repo).UpdateEventRule(ctx, "user@example.com", validUpdate())
		if err != nil || rule.ID != 2 {
			t.Fatalf("rule = %+v, error = %v", rule, err)
		}
	})

	t.Run("rule id invalid", func(t *testing.T) {
		in := validUpdate()
		in.RuleID = 0
		_, err := newAuthorizedRuleService(&fakeRepository{}).UpdateEventRule(ctx, "user", in)
		assertRuleError(t, err, apperrors.ErrInvalidPayload)
	})

	t.Run("payload invalid", func(t *testing.T) {
		in := validUpdate()
		in.Severity = "warning"
		_, err := newAuthorizedRuleService(&fakeRepository{}).UpdateEventRule(ctx, "user", in)
		assertRuleError(t, err, apperrors.ErrInvalidPayload)
	})

	t.Run("event tidak ditemukan", func(t *testing.T) {
		repo := &fakeRepository{eventByIDErr: apperrors.ErrNotFound}
		_, err := NewService(repo).UpdateEventRule(ctx, "user", validUpdate())
		assertRuleError(t, err, apperrors.ErrNotFound)
	})

	t.Run("cek keanggotaan error", func(t *testing.T) {
		repo := &fakeRepository{
			eventByID: entities.Event{OrganizerID: 7},
			activeErr: errors.New("db down"),
		}
		_, err := NewService(repo).UpdateEventRule(ctx, "user", validUpdate())
		if err == nil {
			t.Fatal("ingin error")
		}
	})

	t.Run("bukan anggota organizer", func(t *testing.T) {
		repo := &fakeRepository{eventByID: entities.Event{OrganizerID: 7}}
		_, err := NewService(repo).UpdateEventRule(ctx, "user", validUpdate())
		assertRuleError(t, err, apperrors.ErrForbidden)
	})

	t.Run("error repository", func(t *testing.T) {
		repo := &fakeRepository{ruleErr: apperrors.ErrNotFound}
		_, err := newAuthorizedRuleService(repo).UpdateEventRule(ctx, "user", validUpdate())
		assertRuleError(t, err, apperrors.ErrNotFound)
	})
}

func TestServiceDeleteEventRule(t *testing.T) {
	ctx := context.Background()

	t.Run("sukses", func(t *testing.T) {
		if err := newAuthorizedRuleService(&fakeRepository{}).DeleteEventRule(ctx, "user", 1, 2); err != nil {
			t.Fatalf("DeleteEventRule() error = %v", err)
		}
	})

	t.Run("id invalid", func(t *testing.T) {
		svc := newAuthorizedRuleService(&fakeRepository{})
		if err := svc.DeleteEventRule(ctx, "user", 0, 2); !errors.Is(err, apperrors.ErrInvalidPayload) {
			t.Fatalf("error = %v", err)
		}
		if err := svc.DeleteEventRule(ctx, "user", 1, 0); !errors.Is(err, apperrors.ErrInvalidPayload) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("event tidak ditemukan", func(t *testing.T) {
		repo := &fakeRepository{eventByIDErr: apperrors.ErrNotFound}
		err := NewService(repo).DeleteEventRule(ctx, "user", 1, 2)
		assertRuleError(t, err, apperrors.ErrNotFound)
	})

	t.Run("bukan anggota organizer", func(t *testing.T) {
		repo := &fakeRepository{eventByID: entities.Event{OrganizerID: 7}}
		err := NewService(repo).DeleteEventRule(ctx, "user", 1, 2)
		assertRuleError(t, err, apperrors.ErrForbidden)
	})

	t.Run("error repository", func(t *testing.T) {
		repo := &fakeRepository{ruleErr: errors.New("db down")}
		err := newAuthorizedRuleService(repo).DeleteEventRule(ctx, "user", 1, 2)
		if err == nil {
			t.Fatal("ingin error")
		}
	})
}
