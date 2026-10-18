-- Link orders to registered users.
-- Keep the column nullable for historical orders that cannot be matched safely.
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS user_id UUID
    REFERENCES users(id) ON DELETE RESTRICT;

-- Backfill only emails that identify exactly one user, ignoring case/whitespace.
UPDATE orders AS o
SET user_id = matched.id
FROM (
    SELECT
        LOWER(BTRIM(email)) AS normalized_email,
        (ARRAY_AGG(id))[1] AS id
    FROM users
    GROUP BY LOWER(BTRIM(email))
    HAVING COUNT(*) = 1
) AS matched
WHERE o.user_id IS NULL
  AND LOWER(BTRIM(o.customer_email)) = matched.normalized_email;

CREATE INDEX IF NOT EXISTS idx_orders_user_created_at
    ON orders(user_id, created_at DESC);
