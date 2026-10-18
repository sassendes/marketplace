CREATE TABLE IF NOT EXISTS moderation_actions (
    id          BIGSERIAL PRIMARY KEY,
    admin_id    BIGINT NOT NULL,
    target_type TEXT NOT NULL,
    target_id   BIGINT NOT NULL,
    action      TEXT NOT NULL,
    reason      TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_moderation_target ON moderation_actions(target_type, target_id);
