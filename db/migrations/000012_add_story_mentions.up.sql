-- Create story_mentions table
CREATE TABLE IF NOT EXISTS story_mentions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  story_id UUID NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
  mentioned_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  UNIQUE(story_id, mentioned_user_id)
);

CREATE INDEX idx_story_mentions_story_id ON story_mentions(story_id);
CREATE INDEX idx_story_mentions_user_id ON story_mentions(mentioned_user_id);
