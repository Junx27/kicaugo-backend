-- Migration: profil user, kategori favorit, dan relasi lokasi event.
-- Dijalankan ulang setiap boot: semua DDL dan DML di sini harus idempotent.

CREATE TABLE IF NOT EXISTS user_profiles (
    id                          BIGSERIAL PRIMARY KEY,
    user_id                     BIGINT      NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    full_name                   TEXT        NOT NULL,
    phone                       TEXT,
    avatar_url                  TEXT,
    bio                         TEXT,
    city_id                     BIGINT      REFERENCES cities(id) ON DELETE SET NULL,
    current_latitude            DOUBLE PRECISION CHECK (current_latitude BETWEEN -90 AND 90),
    current_longitude           DOUBLE PRECISION CHECK (current_longitude BETWEEN -180 AND 180),
    current_location_accuracy_m INT         CHECK (current_location_accuracy_m >= 0),
    current_location_updated_at TIMESTAMPTZ,
    search_radius_km            INT         NOT NULL DEFAULT 50 CHECK (search_radius_km BETWEEN 1 AND 500),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_user_profiles_location_lengkap CHECK (
        (current_latitude IS NULL AND current_longitude IS NULL AND current_location_updated_at IS NULL)
        OR
        (current_latitude IS NOT NULL AND current_longitude IS NOT NULL AND current_location_updated_at IS NOT NULL)
    )
);

CREATE TABLE IF NOT EXISTS user_favorite_categories (
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id BIGINT NOT NULL REFERENCES bird_categories(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, category_id)
);

ALTER TABLE events
    ADD COLUMN IF NOT EXISTS city_id BIGINT REFERENCES cities(id),
    ADD COLUMN IF NOT EXISTS province_id BIGINT REFERENCES provinces(id);

-- Backfill best-effort dari kolom TEXT lama.
-- Nama kota yang cocok dengan lebih dari satu baris (mis. Bandung kabupaten/kota)
-- sengaja dibiarkan NULL agar tidak menebak.
WITH kandidat AS (
    SELECT e.id AS event_id,
           e.city_id AS lama_city_id,
           e.province_id AS lama_province_id,
           count(DISTINCT c.id) AS jumlah_kota,
           min(c.id) AS kota_id,
           count(DISTINCT p.id) AS jumlah_provinsi,
           min(p.id) AS provinsi_dari_teks
    FROM events e
    LEFT JOIN cities c ON lower(btrim(c.name)) = lower(btrim(e.city))
    LEFT JOIN provinces p ON lower(btrim(p.name)) = lower(btrim(e.province))
    WHERE e.city_id IS NULL OR e.province_id IS NULL
    GROUP BY e.id, e.city_id, e.province_id
)
UPDATE events e
SET city_id = CASE
        WHEN k.lama_city_id IS NULL AND k.jumlah_kota = 1 THEN k.kota_id
        ELSE k.lama_city_id
    END,
    province_id = CASE
        WHEN k.lama_province_id IS NOT NULL THEN k.lama_province_id
        WHEN k.jumlah_kota = 1 THEN (SELECT c.province_id FROM cities c WHERE c.id = k.kota_id)
        WHEN k.jumlah_provinsi = 1 THEN k.provinsi_dari_teks
        ELSE NULL
    END
FROM kandidat k
WHERE e.id = k.event_id;

-- updated_at otomatis untuk profil.
CREATE OR REPLACE FUNCTION kicaugo_set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_user_profiles_updated_at ON user_profiles;
CREATE TRIGGER trg_user_profiles_updated_at
    BEFORE UPDATE ON user_profiles
    FOR EACH ROW EXECUTE FUNCTION kicaugo_set_updated_at();

-- Profil dibuat otomatis saat user baru terdaftar.
CREATE OR REPLACE FUNCTION kicaugo_create_user_profile() RETURNS trigger AS $$
BEGIN
    INSERT INTO user_profiles (user_id, full_name)
    VALUES (NEW.id, NEW.name)
    ON CONFLICT (user_id) DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_users_create_profile ON users;
CREATE TRIGGER trg_users_create_profile
    AFTER INSERT ON users
    FOR EACH ROW EXECUTE FUNCTION kicaugo_create_user_profile();
