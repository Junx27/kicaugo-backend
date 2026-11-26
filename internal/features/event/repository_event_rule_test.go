package event

import (
	"context"
	"errors"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	dbutil "github.com/Junx27/kicaugo-backend/internal/utils/db"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	"github.com/jackc/pgx/v5/pgconn"
)

func createRuleEvent(t *testing.T, conn dbutil.Querier, ctx context.Context, organizerID int64, slug string) int64 {
	t.Helper()
	var id int64
	err := conn.QueryRow(ctx, `INSERT INTO events (organizer_id, title, slug, venue, city, status)
		VALUES ($1, 'Rule Event', $2, 'Venue', 'City', 'published') RETURNING id`, organizerID, slug).Scan(&id)
	if err != nil {
		t.Fatalf("insert event: %v", err)
	}
	return id
}

func TestRepositoryEventRuleCRUD(t *testing.T) {
	repo, ctx, conn := newTestRepository(t)
	organizerID := createOrganizer(t, conn, ctx)
	suffix := testSuffix(t)
	eventID := createRuleEvent(t, conn, ctx, organizerID, "rules-"+suffix)

	rulesToInsert := []entities.EventRuleInput{
		{Title: "Aturan Sangkar", Description: strPtr("Sangkar standar"), Severity: entities.RuleSeverityInfo, SortOrder: 2},
		{Title: "Larangan Tegas", Severity: entities.RuleSeverityError, SortOrder: 1},
		{Title: "Tim Juri", Severity: entities.RuleSeveritySuccess, SortOrder: 3},
	}
	for _, in := range rulesToInsert {
		if _, err := repo.CreateEventRule(ctx, entities.EventRuleCreate{EventID: eventID, EventRuleInput: in}); err != nil {
			t.Fatalf("CreateEventRule() error = %v", err)
		}
	}

	rules, err := repo.ListEventRules(ctx, eventID)
	if err != nil {
		t.Fatalf("ListEventRules() error = %v", err)
	}
	if len(rules) != 3 {
		t.Fatalf("rules = %d, ingin 3", len(rules))
	}
	wantOrder := []string{"Larangan Tegas", "Aturan Sangkar", "Tim Juri"}
	for i, title := range wantOrder {
		if rules[i].Title != title {
			t.Fatalf("rules[%d].Title = %q, ingin %q", i, rules[i].Title, title)
		}
	}
	if rules[1].Description == nil || *rules[1].Description != "Sangkar standar" {
		t.Fatalf("description = %v", rules[1].Description)
	}

	created, err := repo.CreateEventRule(ctx, entities.EventRuleCreate{
		EventID: eventID,
		EventRuleInput: entities.EventRuleInput{
			Title: "Aturan Baru", Severity: entities.RuleSeveritySuccess, SortOrder: 4,
		},
	})
	if err != nil || created.ID == 0 {
		t.Fatalf("CreateEventRule() = %+v, %v", created, err)
	}

	update := entities.EventRuleUpdate{EventID: eventID, RuleID: created.ID, EventRuleInput: entities.EventRuleInput{
		Title: "Aturan Diubah", Description: strPtr("Deskripsi baru"), Severity: entities.RuleSeverityError, SortOrder: 5,
	}}
	updated, err := repo.UpdateEventRule(ctx, update)
	if err != nil {
		t.Fatalf("UpdateEventRule() error = %v", err)
	}
	if updated.Title != "Aturan Diubah" || updated.Severity != entities.RuleSeverityError || updated.SortOrder != 5 {
		t.Fatalf("updated = %+v", updated)
	}

	if err := repo.DeleteEventRule(ctx, eventID, created.ID); err != nil {
		t.Fatalf("DeleteEventRule() error = %v", err)
	}
	rules, err = repo.ListEventRules(ctx, eventID)
	if err != nil || len(rules) != 3 {
		t.Fatalf("after delete rules = %d, error = %v", len(rules), err)
	}
}

func TestRepositoryEventRuleDomainErrors(t *testing.T) {
	validInput := entities.EventRuleInput{Title: "Aturan", Severity: entities.RuleSeverityInfo}

	t.Run("create event tidak ada", func(t *testing.T) {
		repo, ctx, _ := newTestRepository(t)
		_, err := repo.CreateEventRule(ctx, entities.EventRuleCreate{
			EventID: 999999999, EventRuleInput: validInput,
		})
		if !errors.Is(err, apperrors.ErrNotFound) {
			t.Fatalf("error = %v, ingin ErrNotFound", err)
		}
	})

	t.Run("severity di luar enum", func(t *testing.T) {
		repo, ctx, conn := newTestRepository(t)
		organizerID := createOrganizer(t, conn, ctx)
		eventID := createRuleEvent(t, conn, ctx, organizerID, "rules-invalid-"+testSuffix(t))
		invalid := validInput
		invalid.Severity = "warning"
		_, err := repo.CreateEventRule(ctx, entities.EventRuleCreate{EventID: eventID, EventRuleInput: invalid})
		if !errors.Is(err, apperrors.ErrInvalidPayload) {
			t.Fatalf("error = %v, ingin ErrInvalidPayload", err)
		}
	})

	t.Run("update rule tidak ada", func(t *testing.T) {
		repo, ctx, conn := newTestRepository(t)
		organizerID := createOrganizer(t, conn, ctx)
		eventID := createRuleEvent(t, conn, ctx, organizerID, "rules-update-"+testSuffix(t))
		_, err := repo.UpdateEventRule(ctx, entities.EventRuleUpdate{
			EventID: eventID, RuleID: 999999999, EventRuleInput: validInput,
		})
		if !errors.Is(err, apperrors.ErrNotFound) {
			t.Fatalf("error = %v, ingin ErrNotFound", err)
		}
	})

	t.Run("delete rule tidak ada", func(t *testing.T) {
		repo, ctx, conn := newTestRepository(t)
		organizerID := createOrganizer(t, conn, ctx)
		eventID := createRuleEvent(t, conn, ctx, organizerID, "rules-delete-"+testSuffix(t))
		err := repo.DeleteEventRule(ctx, eventID, 999999999)
		if !errors.Is(err, apperrors.ErrNotFound) {
			t.Fatalf("error = %v, ingin ErrNotFound", err)
		}
	})
}

func TestRepositoryEventRuleQueryErrors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db error")
	pgErr := &pgconn.PgError{Code: "42601"}

	t.Run("list query error", func(t *testing.T) {
		repo := &repository{db: fakeQuerier{queryErr: dbErr}}
		if _, err := repo.ListEventRules(ctx, 1); err == nil {
			t.Fatal("ingin error")
		}
	})

	t.Run("list scan error", func(t *testing.T) {
		repo := &repository{db: fakeQuerier{rows: &fakeRows{nextMax: 1, scanErr: dbErr}}}
		if _, err := repo.ListEventRules(ctx, 1); err == nil {
			t.Fatal("ingin error")
		}
	})

	t.Run("list rows error", func(t *testing.T) {
		repo := &repository{db: fakeQuerier{rows: &fakeRows{err: dbErr}}}
		if _, err := repo.ListEventRules(ctx, 1); err == nil {
			t.Fatal("ingin error")
		}
	})

	t.Run("create pg error", func(t *testing.T) {
		repo := &repository{db: fakeQuerier{rowPgErr: pgErr}}
		if _, err := repo.CreateEventRule(ctx, entities.EventRuleCreate{}); err == nil {
			t.Fatal("ingin error")
		}
	})

	t.Run("update query error", func(t *testing.T) {
		repo := &repository{db: fakeQuerier{rowErr: dbErr}}
		if _, err := repo.UpdateEventRule(ctx, entities.EventRuleUpdate{}); err == nil {
			t.Fatal("ingin error")
		}
	})

	t.Run("delete query error", func(t *testing.T) {
		repo := &repository{db: fakeQuerier{execErr: dbErr}}
		if err := repo.DeleteEventRule(ctx, 1, 1); err == nil {
			t.Fatal("ingin error")
		}
	})
}
