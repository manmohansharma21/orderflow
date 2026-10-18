DROP INDEX IF EXISTS idx_orders_user_created_at;

ALTER TABLE orders
    DROP COLUMN IF EXISTS user_id;
