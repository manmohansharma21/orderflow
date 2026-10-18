BEGIN;

-- Make the payment schema compatible with both the corrected clean-install
-- schema and the earlier local schema that used provider_reference.
ALTER TABLE payments
    ADD COLUMN IF NOT EXISTS provider_payment_id TEXT;

ALTER TABLE payments
    ADD COLUMN IF NOT EXISTS idempotency_key TEXT;

ALTER TABLE payments
    ADD COLUMN IF NOT EXISTS callback_event_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_provider_payment_id_unique
    ON payments(provider, provider_payment_id)
    WHERE provider_payment_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_idempotency_key_unique
    ON payments(idempotency_key)
    WHERE idempotency_key IS NOT NULL AND btrim(idempotency_key) <> '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_callback_event_id_unique
    ON payments(callback_event_id)
    WHERE callback_event_id IS NOT NULL AND btrim(callback_event_id) <> '';

CREATE INDEX IF NOT EXISTS idx_payments_order_status_created
    ON payments(order_id, status, created_at DESC);

COMMIT;
