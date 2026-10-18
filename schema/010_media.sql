CREATE TABLE IF NOT EXISTS media_assets (
    id           BIGSERIAL PRIMARY KEY,
    owner_id     BIGINT NOT NULL,
    kind         TEXT NOT NULL,
    object_key   TEXT NOT NULL UNIQUE,
    content_type TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_media_owner ON media_assets(owner_id);
