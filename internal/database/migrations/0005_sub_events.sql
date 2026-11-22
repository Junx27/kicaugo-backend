-- Tabel sub_events: sub-lomba + juara 1/2/3 + prize pool
CREATE TABLE IF NOT EXISTS sub_events (
    id               BIGSERIAL PRIMARY KEY,
    event_id         BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    bird_category_id BIGINT REFERENCES bird_categories(id) ON DELETE SET NULL,
    title            TEXT NOT NULL,
    description      TEXT,
    session          TEXT,
    gantangan_count  INT NOT NULL DEFAULT 0,
    max_slots        INT NOT NULL DEFAULT 60 CHECK (max_slots <= 60 AND max_slots > 0),
    price            BIGINT NOT NULL DEFAULT 0,
    winner_1_prize   BIGINT NOT NULL DEFAULT 0,
    winner_2_prize   BIGINT NOT NULL DEFAULT 0,
    winner_3_prize   BIGINT NOT NULL DEFAULT 0,
    status           TEXT NOT NULL DEFAULT 'available'
                     CHECK (status IN ('available','almostGone','soldOut','waiting')),
    start_time       TIMESTAMPTZ,
    end_time         TIMESTAMPTZ,
    sort_order       INT NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);