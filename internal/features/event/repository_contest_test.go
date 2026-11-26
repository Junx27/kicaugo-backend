package event

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type contestQueryResult struct {
	rows *fakeRows
	err  error
}

type contestQuerier struct {
	fakeQuerier
	results []contestQueryResult
	calls   int
}

func (q *contestQuerier) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if q.calls >= len(q.results) {
		return &fakeRows{}, nil
	}
	result := q.results[q.calls]
	q.calls++
	if result.err != nil {
		return nil, result.err
	}
	return result.rows, nil
}

func contestClassRows(maxSlots, active int) *fakeRows {
	return &fakeRows{nextMax: 1, scanFunc: func(dest ...any) error {
		*dest[0].(*int64) = 1
		*dest[1].(*string) = "Class"
		*dest[2].(**string) = nil
		*dest[3].(**time.Time) = nil
		*dest[4].(*int) = maxSlots
		*dest[5].(*int) = active
		*dest[6].(*int64) = 100000
		*dest[7].(*string) = "Juara 1"
		return nil
	}}
}

func TestRepositoryGetContestDetailCoverage(t *testing.T) {
	repo, ctx, conn := newTestRepository(t)
	organizerID := createOrganizer(t, conn, ctx)
	suffix := testSuffix(t)

	var eventID int64
	if err := conn.QueryRow(ctx, `INSERT INTO events (organizer_id, title, slug, venue, city, status, total_slots, subtitle, hero_image_url, banner_url, distance_text, map_image_url, latitude, longitude, parking, prize_pool, prize_trophy, prize_highlight, stat_format_label, stat_format_note, stat_judging_label, stat_judging_note, stat_facilities_label, stat_facilities_note, schedule_time, gate_open, address) VALUES ($1, 'Contest', $2, 'Venue', 'City', 'published', 36, 'Sub', 'hero', 'banner', '5 Menit', 'map', -6.2, 106.8, 'Parkir', '50000000', 'Trofi', 'Highlight', '36 Gantangan', 'Tertib', 'PBI', 'Tanpa Koordinasi', 'VIP', 'LED', '08:30 WIB – Selesai', '07:00 WIB', 'Alamat') RETURNING id`, organizerID, "event-"+suffix).Scan(&eventID); err != nil {
		t.Fatalf("insert event: %v", err)
	}
	if _, err := conn.Exec(ctx, `INSERT INTO sub_events (event_id, title, session, max_slots, price, status, start_time, prize_highlight) VALUES ($1, 'Class', 'Sesi 04', 10, 500000, 'available', now(), 'Juara 1')`, eventID); err != nil {
		t.Fatalf("insert sub event: %v", err)
	}
	if _, err := conn.Exec(ctx, `INSERT INTO contest_rules (event_id, title, description, severity, sort_order) VALUES ($1, 'Aturan', 'Desc', 'info', 1)`, eventID); err != nil {
		t.Fatalf("insert rule: %v", err)
	}

	detail, err := repo.GetContestDetail(ctx, eventID)
	if err != nil {
		t.Fatalf("GetContestDetail() error = %v", err)
	}
	if len(detail.Classes) != 1 || len(detail.Rules) != 1 {
		t.Fatalf("classes=%d rules=%d", len(detail.Classes), len(detail.Rules))
	}
	if detail.Classes[0].RemainingSlots == nil {
		t.Fatal("remaining nil")
	}

	// empty classes/rules path
	var emptyEventID int64
	if err := conn.QueryRow(ctx, `INSERT INTO events (organizer_id, title, slug, venue, city, status, total_slots) VALUES ($1, 'Empty', $2, 'Venue', 'City', 'published', 5) RETURNING id`, organizerID, "empty-"+suffix).Scan(&emptyEventID); err != nil {
		t.Fatalf("insert empty event: %v", err)
	}
	emptyDetail, err := repo.GetContestDetail(ctx, emptyEventID)
	if err != nil {
		t.Fatalf("GetContestDetail() empty error = %v", err)
	}
	if len(emptyDetail.Classes) != 0 || len(emptyDetail.Rules) != 0 {
		t.Fatalf("empty classes=%d rules=%d", len(emptyDetail.Classes), len(emptyDetail.Rules))
	}
}

func TestRepositoryGetContestDetailNotFoundCoverage(t *testing.T) {
	repo, ctx, _ := newTestRepository(t)
	if _, err := repo.GetContestDetail(ctx, 999999); err == nil {
		t.Fatal("ingin error")
	}
}

func TestRepositoryGetContestDetailErrorBranchesCoverage(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db error")

	t.Run("detail scan error", func(t *testing.T) {
		repo := &repository{db: &contestQuerier{fakeQuerier: fakeQuerier{rowErr: dbErr}}}
		if _, err := repo.GetContestDetail(ctx, 1); err == nil {
			t.Fatal("ingin error")
		}
	})

	t.Run("class query error", func(t *testing.T) {
		repo := &repository{db: &contestQuerier{results: []contestQueryResult{{err: dbErr}}}}
		if _, err := repo.GetContestDetail(ctx, 1); err == nil {
			t.Fatal("ingin error")
		}
	})

	t.Run("class scan error", func(t *testing.T) {
		repo := &repository{db: &contestQuerier{results: []contestQueryResult{{rows: &fakeRows{nextMax: 1, scanErr: dbErr}}}}}
		if _, err := repo.GetContestDetail(ctx, 1); err == nil {
			t.Fatal("ingin error")
		}
	})

	t.Run("remaining slots clamped", func(t *testing.T) {
		repo := &repository{db: &contestQuerier{results: []contestQueryResult{
			{rows: contestClassRows(0, 5)},
			{rows: &fakeRows{}},
		}}}
		detail, err := repo.GetContestDetail(ctx, 1)
		if err != nil {
			t.Fatalf("GetContestDetail() error = %v", err)
		}
		if len(detail.Classes) != 1 || detail.Classes[0].RemainingSlots == nil || *detail.Classes[0].RemainingSlots != 0 {
			t.Fatalf("classes = %+v", detail.Classes)
		}
	})

	t.Run("class rows error", func(t *testing.T) {
		repo := &repository{db: &contestQuerier{results: []contestQueryResult{{rows: &fakeRows{err: dbErr}}}}}
		if _, err := repo.GetContestDetail(ctx, 1); err == nil {
			t.Fatal("ingin error")
		}
	})

	t.Run("rule query error", func(t *testing.T) {
		repo := &repository{db: &contestQuerier{results: []contestQueryResult{
			{rows: &fakeRows{}},
			{err: dbErr},
		}}}
		if _, err := repo.GetContestDetail(ctx, 1); err == nil {
			t.Fatal("ingin error")
		}
	})

	t.Run("rule scan error", func(t *testing.T) {
		repo := &repository{db: &contestQuerier{results: []contestQueryResult{
			{rows: &fakeRows{}},
			{rows: &fakeRows{nextMax: 1, scanErr: dbErr}},
		}}}
		if _, err := repo.GetContestDetail(ctx, 1); err == nil {
			t.Fatal("ingin error")
		}
	})

	t.Run("rule rows error", func(t *testing.T) {
		repo := &repository{db: &contestQuerier{results: []contestQueryResult{
			{rows: &fakeRows{}},
			{rows: &fakeRows{err: dbErr}},
		}}}
		if _, err := repo.GetContestDetail(ctx, 1); err == nil {
			t.Fatal("ingin error")
		}
	})
}
