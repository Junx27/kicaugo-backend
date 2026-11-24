package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

type fakeRows struct {
	pgx.Rows
	values  [][]any
	index   int
	scanErr error
	err     error
}

func (r *fakeRows) Next() bool {
	if r.index >= len(r.values) {
		return false
	}
	r.index++
	return true
}

func (r *fakeRows) Scan(dest ...any) error {
	if r.scanErr != nil {
		return r.scanErr
	}
	values := r.values[r.index-1]
	for i, target := range dest {
		switch out := target.(type) {
		case *int64:
			*out = values[i].(int64)
		case *string:
			*out = values[i].(string)
		case *time.Time:
			*out = values[i].(time.Time)
		}
	}
	return nil
}

func (r *fakeRows) Err() error { return r.err }

type fakeRow struct {
	pgx.Row
	value int64
	err   error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*(dest[0].(*int64)) = r.value
	return nil
}

type fakeQuerier struct {
	Querier
	row pgx.Row
}

func (q fakeQuerier) QueryRow(context.Context, string, ...any) pgx.Row { return q.row }

func TestScanBirdCategories(t *testing.T) {
	now := time.Now()
	rows := &fakeRows{values: [][]any{{int64(7), "Kutilang", "kutilang", "icon.png", now}}}
	got, err := ScanBirdCategories(rows)
	if err != nil {
		t.Fatalf("ScanBirdCategories() error = %v", err)
	}
	if len(got) != 1 || got[0] != (entities.BirdCategory{ID: 7, Name: "Kutilang", Slug: "kutilang", Icon: "icon.png", CreatedAt: now}) {
		t.Errorf("ScanBirdCategories() = %v", got)
	}
}

func TestScanBirdCategoriesScanError(t *testing.T) {
	rows := &fakeRows{values: [][]any{{int64(1), "a", "a", "", time.Now()}}, scanErr: errors.New("scan gagal")}
	if _, err := ScanBirdCategories(rows); err == nil {
		t.Error("ScanBirdCategories() ingin error saat Scan gagal")
	}
}

func TestScanBirdCategoriesIterateError(t *testing.T) {
	rows := &fakeRows{err: errors.New("iterasi gagal")}
	if _, err := ScanBirdCategories(rows); err == nil {
		t.Error("ScanBirdCategories() ingin error saat Err gagal")
	}
}

func TestResolveUserID(t *testing.T) {
	q := fakeQuerier{row: fakeRow{value: 42}}
	id, err := ResolveUserID(context.Background(), q, "user@example.com")
	if err != nil {
		t.Fatalf("ResolveUserID() error = %v", err)
	}
	if id != 42 {
		t.Errorf("ResolveUserID() = %d, ingin 42", id)
	}
}

func TestResolveUserIDNotFound(t *testing.T) {
	q := fakeQuerier{row: fakeRow{err: pgx.ErrNoRows}}
	if _, err := ResolveUserID(context.Background(), q, "missing@example.com"); !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("ResolveUserID() error = %v, ingin ErrNotFound", err)
	}
}

func TestResolveUserIDError(t *testing.T) {
	q := fakeQuerier{row: fakeRow{err: errors.New("db mati")}}
	if _, err := ResolveUserID(context.Background(), q, "user@example.com"); err == nil {
		t.Error("ResolveUserID() ingin error saat query gagal")
	}
}
