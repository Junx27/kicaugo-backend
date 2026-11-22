-- Tabel organizers: penyelenggara lomba (EO, panitia)
CREATE TABLE IF NOT EXISTS organizers (
    id            BIGSERIAL PRIMARY KEY,
    name          TEXT NOT NULL,
    slug          TEXT NOT NULL UNIQUE,
    email         TEXT NOT NULL UNIQUE,
    phone         TEXT,
    logo_url      TEXT,
    description   TEXT,
    is_verified   BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);