-- Add ghost mode support to users table
ALTER TABLE users ADD COLUMN is_ghost_mode BOOLEAN NOT NULL DEFAULT false;

-- Create index for efficient ghost mode filtering
CREATE INDEX idx_users_ghost_mode ON users (is_ghost_mode) WHERE is_ghost_mode = true;
