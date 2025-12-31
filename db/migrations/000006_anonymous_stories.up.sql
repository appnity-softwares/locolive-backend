-- Add anonymous flag to stories
ALTER TABLE stories ADD COLUMN is_anonymous BOOLEAN NOT NULL DEFAULT false;

-- Create index for filtering anonymous stories
CREATE INDEX idx_stories_anonymous ON stories (is_anonymous);
