-- Tabel events: lomba utama (support draft & reusable template)
CREATE TABLE IF NOT EXISTS events (
    id                  BIGSERIAL PRIMARY KEY,
    organizer_id        BIGINT NOT NULL REFERENCES organizers(id) ON DELETE CASCADE,
    title               TEXT NOT NULL,
    slug                TEXT NOT NULL UNIQUE,
    description         TEXT,
    venue               TEXT NOT NULL,
    address             TEXT,
    city                TEXT NOT NULL,
    province            TEXT,
    latitude            DOUBLE PRECISION,
    longitude           DOUBLE PRECISION,
    start_date          DATE,
    end_date            DATE,
    registration_open   TIMESTAMPTZ,
    registration_close  TIMESTAMPTZ,
    status              TEXT NOT NULL DEFAULT 'draft'
                        CHECK (status IN ('draft','published','ongoing','completed','cancelled')),
    is_reusable         BOOLEAN NOT NULL DEFAULT false,
    banner_url          TEXT,
    prize_pool          TEXT,
    total_slots         INT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);