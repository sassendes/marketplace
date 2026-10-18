-- seller service schema
CREATE TABLE IF NOT EXISTS sellers (
    id                BIGSERIAL PRIMARY KEY,
    user_id           BIGINT NOT NULL UNIQUE,
    store_name        TEXT NOT NULL,
    slug              TEXT NOT NULL UNIQUE,
    description       TEXT NOT NULL DEFAULT '',
    status            TEXT NOT NULL DEFAULT 'pending',
    commission_rate   INT NOT NULL DEFAULT 1000,
    stripe_account_id TEXT NOT NULL DEFAULT '',
    payout_enabled    BOOLEAN NOT NULL DEFAULT false,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS payouts (
    id          BIGSERIAL PRIMARY KEY,
    seller_id   BIGINT NOT NULL REFERENCES sellers(id) ON DELETE CASCADE,
    amount      BIGINT NOT NULL,
    currency    TEXT NOT NULL DEFAULT 'USD',
    status      TEXT NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_payouts_seller ON payouts(seller_id);
