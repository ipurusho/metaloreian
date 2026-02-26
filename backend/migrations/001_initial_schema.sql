CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS bands (
    ma_id       BIGINT PRIMARY KEY,
    name        TEXT NOT NULL,
    genre       TEXT NOT NULL DEFAULT '',
    country     TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT '',
    themes      TEXT NOT NULL DEFAULT '',
    formed_in   TEXT NOT NULL DEFAULT '',
    years_active TEXT NOT NULL DEFAULT '',
    logo_url    TEXT NOT NULL DEFAULT '',
    photo_url   TEXT NOT NULL DEFAULT '',
    scraped_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bands_name_trgm ON bands USING gin (name gin_trgm_ops);

CREATE TABLE IF NOT EXISTS albums (
    album_id     BIGINT PRIMARY KEY,
    band_id      BIGINT NOT NULL REFERENCES bands(ma_id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    type         TEXT NOT NULL DEFAULT '',
    release_date TEXT NOT NULL DEFAULT '',
    label        TEXT NOT NULL DEFAULT '',
    format       TEXT NOT NULL DEFAULT '',
    cover_url    TEXT NOT NULL DEFAULT '',
    scraped_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_albums_band_id ON albums(band_id);

CREATE TABLE IF NOT EXISTS tracks (
    id           BIGSERIAL PRIMARY KEY,
    album_id     BIGINT NOT NULL REFERENCES albums(album_id) ON DELETE CASCADE,
    track_number INTEGER NOT NULL,
    title        TEXT NOT NULL,
    duration     TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_tracks_album_id ON tracks(album_id);

CREATE TABLE IF NOT EXISTS members (
    member_id BIGINT PRIMARY KEY,
    name      TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS band_lineup (
    id          BIGSERIAL PRIMARY KEY,
    band_id     BIGINT NOT NULL REFERENCES bands(ma_id) ON DELETE CASCADE,
    member_id   BIGINT NOT NULL REFERENCES members(member_id) ON DELETE CASCADE,
    instrument  TEXT NOT NULL DEFAULT '',
    lineup_type TEXT NOT NULL DEFAULT 'current',
    years       TEXT NOT NULL DEFAULT '',
    UNIQUE(band_id, member_id, lineup_type)
);

CREATE INDEX IF NOT EXISTS idx_band_lineup_band_id ON band_lineup(band_id);

CREATE TABLE IF NOT EXISTS album_lineup (
    id          BIGSERIAL PRIMARY KEY,
    album_id    BIGINT NOT NULL REFERENCES albums(album_id) ON DELETE CASCADE,
    member_id   BIGINT NOT NULL REFERENCES members(member_id) ON DELETE CASCADE,
    instrument  TEXT NOT NULL DEFAULT '',
    UNIQUE(album_id, member_id)
);

CREATE INDEX IF NOT EXISTS idx_album_lineup_album_id ON album_lineup(album_id);

CREATE TABLE IF NOT EXISTS member_bands (
    id          BIGSERIAL PRIMARY KEY,
    member_id   BIGINT NOT NULL REFERENCES members(member_id) ON DELETE CASCADE,
    band_id     BIGINT NOT NULL,
    band_name   TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_member_bands_member_id ON member_bands(member_id);
