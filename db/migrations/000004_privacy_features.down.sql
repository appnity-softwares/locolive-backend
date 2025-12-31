-- Remove ghost mode index
DROP INDEX IF EXISTS idx_users_ghost_mode;

-- Remove ghost mode column
ALTER TABLE users DROP COLUMN IF EXISTS is_ghost_mode;
