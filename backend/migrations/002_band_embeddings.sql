CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS band_embeddings (
    band_id   BIGINT PRIMARY KEY,
    band_name TEXT,
    embedding vector(32) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_band_embeddings_cosine
    ON band_embeddings USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);
