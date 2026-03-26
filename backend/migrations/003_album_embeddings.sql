CREATE TABLE IF NOT EXISTS album_embeddings (
    album_id BIGINT PRIMARY KEY,
    album_name TEXT,
    embedding vector(32) NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_album_embeddings_cosine
    ON album_embeddings USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 800);
