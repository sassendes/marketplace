CREATE TABLE IF NOT EXISTS stock (
    variant_id  BIGINT PRIMARY KEY,
    on_hand     INT NOT NULL DEFAULT 0,
    reserved    INT NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (reserved >= 0),
    CHECK (on_hand >= 0),
    CHECK (reserved <= on_hand)
);

CREATE TABLE IF NOT EXISTS reservations (
    id          BIGSERIAL PRIMARY KEY,
    variant_id  BIGINT NOT NULL,
    order_id    BIGINT NOT NULL,
    quantity    INT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'held',
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_reservations_order ON reservations(order_id);
CREATE INDEX IF NOT EXISTS idx_reservations_status ON reservations(status, expires_at);
