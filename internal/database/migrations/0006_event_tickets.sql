-- Tabel event_tickets: tiket + QR code + verifikasi oleh organizer
CREATE TABLE IF NOT EXISTS event_tickets (
    id              BIGSERIAL PRIMARY KEY,
    event_id        BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    sub_event_id    BIGINT REFERENCES sub_events(id) ON DELETE SET NULL,
    user_id         BIGINT REFERENCES users(id) ON DELETE SET NULL,
    ticket_number   TEXT NOT NULL UNIQUE,
    qr_code         TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active','used','cancelled','expired')),
    price_paid      BIGINT NOT NULL DEFAULT 0,
    purchased_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    verified_at     TIMESTAMPTZ,
    verified_by     BIGINT REFERENCES organizers(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);