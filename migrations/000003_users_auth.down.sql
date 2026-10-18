-- No-op intentionally: users is created by 000001_init_schema.up.sql.
-- 000003_users_auth.up.sql is a redundant CREATE TABLE IF NOT EXISTS migration.
-- Rolling back 000003 must not drop the users table owned by 000001.
SELECT 1;
