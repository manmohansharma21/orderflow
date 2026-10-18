BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_email TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING'
        CHECK (status IN (
            'PENDING', 'CONFIRMED', 'CANCELLED', 'FAILED'
        )),
    total_amount_paise BIGINT NOT NULL DEFAULT 0
        CHECK (total_amount_paise >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_customer_email
    ON orders(customer_email);

CREATE INDEX IF NOT EXISTS idx_orders_created_at
    ON orders(created_at DESC);

CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    product_name TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0)    quantity INTEGER NOT NULL CHECK (quantity > 0)  rice_paise     quantity INTEGER NOT NULL CHECK (q NUL    quantity INTEGER NOT NULL CHEC      quantity INTEGER NOT NULL CHECK (quantity >);


   quantity INTEGER NOT NULL CHnt   quantity INTEGER NOT NULL CHnt   qen_ran   quantity INTEGER NOTd UU   quantity INTEGER NOT NULL CHnt   quantity INTEGER r TEX   quanULL,
   quantity INTEGER NOT NULL CHnt   quantit         quantity INTEGER NOT NULL CHnt   quING', 'SUCCESS', 'FAILED', 'REFUNDED'
        )),
    amount_paise BIGINT NOT NULL CHECK (amount_paise >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;
