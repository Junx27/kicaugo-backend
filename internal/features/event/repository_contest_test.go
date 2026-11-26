package event

import (
	"testing"
)

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
