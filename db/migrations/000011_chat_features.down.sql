-- Remove read_at column from messages
ALTER TABLE messages DROP COLUMN IF EXISTS read_at;

-- Drop message_reactions table
DROP TABLE IF EXISTS message_reactions;
