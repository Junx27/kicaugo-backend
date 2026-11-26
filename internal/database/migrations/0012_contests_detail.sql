-- Tabel detail contest: metadata tambahan untuk halaman detail lomba.
ALTER TABLE events
	ADD COLUMN IF NOT EXISTS subtitle TEXT,
	ADD COLUMN IF NOT EXISTS official_badge TEXT,
	ADD COLUMN IF NOT EXISTS format_badge TEXT,
	ADD COLUMN IF NOT EXISTS edition_label TEXT,
	ADD COLUMN IF NOT EXISTS hero_image_url TEXT,
	ADD COLUMN IF NOT EXISTS gate_open TEXT,
	ADD COLUMN IF NOT EXISTS schedule_time TEXT,
	ADD COLUMN IF NOT EXISTS distance_text TEXT,
	ADD COLUMN IF NOT EXISTS map_image_url TEXT,
	ADD COLUMN IF NOT EXISTS parking TEXT,
	ADD COLUMN IF NOT EXISTS prize_trophy TEXT,
	ADD COLUMN IF NOT EXISTS prize_highlight TEXT,
	ADD COLUMN IF NOT EXISTS stat_format_label TEXT,
	ADD COLUMN IF NOT EXISTS stat_format_note TEXT,
	ADD COLUMN IF NOT EXISTS stat_judging_label TEXT,
	ADD COLUMN IF NOT EXISTS stat_judging_note TEXT,
	ADD COLUMN IF NOT EXISTS stat_facilities_label TEXT,
	ADD COLUMN IF NOT EXISTS stat_facilities_note TEXT;

ALTER TABLE sub_events
	ADD COLUMN IF NOT EXISTS prize_highlight TEXT;

CREATE TABLE IF NOT EXISTS contest_rules (
	id          BIGSERIAL PRIMARY KEY,
	event_id    BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
	title       TEXT NOT NULL,
	description TEXT,
	severity    TEXT NOT NULL DEFAULT 'info',
	sort_order  INT NOT NULL DEFAULT 0,
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);