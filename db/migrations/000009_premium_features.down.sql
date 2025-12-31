DROP INDEX IF EXISTS idx_users_premium;
ALTER TABLE stories DROP COLUMN IF EXISTS is_premium;
ALTER TABLE users DROP COLUMN IF EXISTS is_premium;
