-- Tabel event_organizer_users: relasi organizer <-> user (role-based middleware)
CREATE TABLE IF NOT EXISTS event_organizer_users (
    id            BIGSERIAL PRIMARY KEY,
    organizer_id  BIGINT NOT NULL REFERENCES organizers(id) ON DELETE CASCADE,
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role          TEXT NOT NULL DEFAULT 'staff'
                  CHECK (role IN ('owner','admin','staff','verifier')),
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (organizer_id, user_id)
);