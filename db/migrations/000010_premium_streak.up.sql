-- Up Migration
ALTER TABLE users 
ADD COLUMN streak_freezes_remaining INT NOT NULL DEFAULT 0,
ADD COLUMN boost_expires_at TIMESTAMPTZ;

-- Down Migration
-- ALTER TABLE users DROP COLUMN streak_freezes_remaining;
-- ALTER TABLE users DROP COLUMN boost_expires_at;
