-- Activity tracking columns (using ALTER IF NOT EXISTS syntax)
ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS last_active_at timestamptz DEFAULT now();
ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS activity_streak integer DEFAULT 0;
ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS streak_updated_at timestamptz DEFAULT now();

-- Create index
CREATE INDEX IF NOT EXISTS idx_users_last_active ON users (last_active_at DESC);
