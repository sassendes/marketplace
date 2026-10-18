CREATE TABLE IF NOT EXISTS reviews (
    id                BIGSERIAL PRIMARY KEY,
    product_id        BIGINT NOT NULL,
    user_id           BIGINT NOT NULL,
    rating            INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title             TEXT NOT NULL DEFAULT '',
    body              TEXT NOT NULL DEFAULT '',
    verified_purchase BOOLEAN NOT NULL DEFAULT false,
    helpful           INT NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (product_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_reviews_product ON reviews(product_id);
