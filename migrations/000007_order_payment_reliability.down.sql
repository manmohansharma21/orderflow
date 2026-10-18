-- The EXPIRED state is retained in the status constraint intentionally so a
-- rollback cannot make existing EXPIRED rows violate the schema constraint.
DROP INDEX IF EXISTS idx_payments_currency_status_created;
DROP INDEX IF EXISTS idx_orders_pending_expiry;
DROP INDEX IF EXISTS idx_orders_user_idempotency_key_unique;
ALTER TABLE payments DROP COLUMN IF EXISTS failure_reason;
ALTER TABLE payments DROP COLUMN IF EXISTS currency;
ALTER TABLE orders DROP COLUMN IF EXISTS idempotency_request_hash;
ALTER TABLE orders DROP COLUMN IF EXISTS idempotency_key;
