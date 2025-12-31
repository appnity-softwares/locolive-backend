-- Add premium user flag
ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS is_premium BOOLEAN DEFAULT false;

-- Add premium flag to stories
ALTER TABLE IF EXISTS stories ADD COLUMN IF NOT EXISTS is_premium BOOLEAN DEFAULT false;

-- Create index for premium users
CREATE INDEX IF NOT EXISTS idx_users_premium ON users (is_premium) WHERE is_premium = true;
