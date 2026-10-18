CREATE TABLE IF NOT EXISTS sagas (
    id           BIGSERIAL PRIMARY KEY,
    order_id     BIGINT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'running',
    current_step TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS saga_steps (
    id         BIGSERIAL PRIMARY KEY,
    saga_id    BIGINT NOT NULL REFERENCES sagas(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    status     TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_saga_steps_saga ON saga_steps(saga_id);
