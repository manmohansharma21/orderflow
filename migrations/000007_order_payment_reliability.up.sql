BEGIN;

-- Order request idempotency, verified payment callback metadata, and expiry state.
ALTER TABLE orders ADD COLUMN IF NOT EXISTS idempotency_key TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS idempotency_request_hash TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_user_idempotency_key_unique
    ON orders(user_id, idempotency_key)
    WHERE user_id IS NOT NULL AND idempotency_key IS NOT NULL AND btrim(idempotency_key) <> '';

ALTER TABLE payments ADD COLUMN IF NOT EXISTS currency CHAR(3) NOT NULL DEFAULT 'INR';
ALTER TABLE payments ADD COLUMN IF NOT EXISTS failure_reason TEXT;

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('PENDING', 'CONFIRMED', 'CANCELLED', 'FAILED', 'EXPIRED'));

CREATE INDEX IF NOT EXISTS idx_orders_pending_expiry
    ON orders(created_at, id) WHERE status = 'PENDING';
CREATE INDEX IF NOT EXISTS idx_payments_currency_status_created
    ON payments(currency, status, created_at DESC);

COMMIT;
