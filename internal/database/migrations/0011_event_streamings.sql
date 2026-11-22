-- Tabel event_streamings: sesi streaming/lomba yang sedang berlangsung.
CREATE TABLE IF NOT EXISTS event_streamings (
    id          BIGSERIAL PRIMARY KEY,
    event_id    BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    start_time  TIMESTAMPTZ NOT NULL DEFAULT now(),
    end_time    TIMESTAMPTZ,
    total_views BIGINT NOT NULL DEFAULT 0 CHECK (total_views >= 0),
    juara       JSONB NOT NULL DEFAULT '{}'::jsonb,
    link_stream TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_event_streamings_event_id ON event_streamings(event_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_event_streamings_one_active_per_event
    ON event_streamings(event_id)
    WHERE end_time IS NULL;

DROP TRIGGER IF EXISTS trg_event_streamings_updated_at ON event_streamings;
CREATE TRIGGER trg_event_streamings_updated_at
    BEFORE UPDATE ON event_streamings
    FOR EACH ROW EXECUTE FUNCTION kicaugo_set_updated_at();
