DROP INDEX IF EXISTS idx_users_last_active;
ALTER TABLE users DROP COLUMN IF EXISTS streak_updated_at;
ALTER TABLE users DROP COLUMN IF EXISTS activity_streak;
ALTER TABLE users DROP COLUMN IF EXISTS last_active_at;
