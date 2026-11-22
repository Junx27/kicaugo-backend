-- Tabel bird_categories: jenis burung (general lookup, tidak tied ke event)
CREATE TABLE IF NOT EXISTS bird_categories (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    icon        TEXT NOT NULL,        -- icon identifier (e.g. "murai_batu", "kacer")
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
