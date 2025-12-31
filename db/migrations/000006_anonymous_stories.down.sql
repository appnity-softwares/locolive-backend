DROP INDEX IF EXISTS idx_stories_anonymous;
ALTER TABLE stories DROP COLUMN IF EXISTS is_anonymous;
