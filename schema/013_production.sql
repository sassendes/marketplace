ALTER TABLE products ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT '';
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS title TEXT NOT NULL DEFAULT '';
ALTER TABLE sagas ADD COLUMN IF NOT EXISTS buyer_id BIGINT NOT NULL DEFAULT 0;
ALTER TABLE sagas ADD COLUMN IF NOT EXISTS idempotency_key TEXT;
ALTER TABLE sagas ADD COLUMN IF NOT EXISTS payment_id BIGINT;
ALTER TABLE sagas ADD COLUMN IF NOT EXISTS error TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_sagas_buyer_idempotency ON sagas (buyer_id, idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_sagas_status_updated ON sagas (status, updated_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_variants_product_sku ON product_variants (product_id, sku);
CREATE INDEX IF NOT EXISTS idx_payments_order ON payments (order_id);
CREATE INDEX IF NOT EXISTS idx_reservations_order_status ON reservations (order_id, status);
CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);
CREATE INDEX IF NOT EXISTS idx_products_created ON products (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_order_items_product ON order_items (product_id);

CREATE TABLE IF NOT EXISTS review_votes (
    review_id  BIGINT NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (review_id, user_id)
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_role_check') THEN
        ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('buyer', 'seller', 'admin')) NOT VALID;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'variants_price_positive') THEN
        ALTER TABLE product_variants ADD CONSTRAINT variants_price_positive CHECK (price > 0) NOT VALID;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'cart_items_quantity_positive') THEN
        ALTER TABLE cart_items ADD CONSTRAINT cart_items_quantity_positive CHECK (quantity > 0 AND quantity <= 999) NOT VALID;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'order_items_quantity_positive') THEN
        ALTER TABLE order_items ADD CONSTRAINT order_items_quantity_positive CHECK (quantity > 0) NOT VALID;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'reservations_quantity_positive') THEN
        ALTER TABLE reservations ADD CONSTRAINT reservations_quantity_positive CHECK (quantity > 0) NOT VALID;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'products_status_check') THEN
        ALTER TABLE products ADD CONSTRAINT products_status_check CHECK (status IN ('draft', 'active', 'inactive', 'removed')) NOT VALID;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'sellers_commission_range') THEN
        ALTER TABLE sellers ADD CONSTRAINT sellers_commission_range CHECK (commission_rate BETWEEN 0 AND 10000) NOT VALID;
    END IF;
END
$$;

UPDATE users SET email = lower(email) WHERE email <> lower(email);
