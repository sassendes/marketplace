CREATE TABLE IF NOT EXISTS orders (
    id          BIGSERIAL PRIMARY KEY,
    buyer_id    BIGINT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending',
    total_cents BIGINT NOT NULL,
    currency    TEXT NOT NULL DEFAULT 'USD',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sub_orders (
    id               BIGSERIAL PRIMARY KEY,
    order_id         BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    seller_id        BIGINT NOT NULL,
    status           TEXT NOT NULL DEFAULT 'pending',
    subtotal_cents   BIGINT NOT NULL,
    commission_cents BIGINT NOT NULL DEFAULT 0,
    currency         TEXT NOT NULL DEFAULT 'USD'
);

CREATE TABLE IF NOT EXISTS order_items (
    id               BIGSERIAL PRIMARY KEY,
    sub_order_id     BIGINT NOT NULL REFERENCES sub_orders(id) ON DELETE CASCADE,
    variant_id       BIGINT NOT NULL,
    product_id       BIGINT NOT NULL,
    quantity         INT NOT NULL,
    unit_price_cents BIGINT NOT NULL,
    currency         TEXT NOT NULL DEFAULT 'USD'
);
CREATE INDEX IF NOT EXISTS idx_orders_buyer ON orders(buyer_id);
CREATE INDEX IF NOT EXISTS idx_sub_orders_order ON sub_orders(order_id);
CREATE INDEX IF NOT EXISTS idx_sub_orders_seller ON sub_orders(seller_id);
CREATE INDEX IF NOT EXISTS idx_order_items_sub ON order_items(sub_order_id);
