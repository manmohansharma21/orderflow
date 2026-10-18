-- Keep provider_payment_id and idempotency_key because they may have existed
-- in the pre-migration schema. Remove only the callback-specific addition.
DROP INDEX IF EXISTS idx_payments_order_status_created;
DROP INDEX IF EXISTS idx_payments_callback_event_id_unique;
DROP INDEX IF EXISTS idx_payments_idempotency_key_unique;
DROP INDEX IF EXISTS idx_payments_provider_payment_id_unique;
ALTER TABLE payments DROP COLUMN IF EXISTS callback_event_id;
