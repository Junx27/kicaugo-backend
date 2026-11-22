-- Migration: create cities table
CREATE TABLE IF NOT EXISTS cities (
    id          BIGSERIAL PRIMARY KEY,
    province_id BIGINT           NOT NULL REFERENCES provinces(id),
    code        TEXT             NOT NULL UNIQUE,
    name        TEXT             NOT NULL,
    type        TEXT             NOT NULL,
    slug        TEXT             NOT NULL UNIQUE,
    latitude    DOUBLE PRECISION NOT NULL,
    longitude   DOUBLE PRECISION NOT NULL,
    created_at  TIMESTAMPTZ      NOT NULL DEFAULT now()
);
