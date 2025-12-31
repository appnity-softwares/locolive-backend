-- Story engagement tables

CREATE TABLE story_views (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  story_id uuid NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  viewed_at timestamptz NOT NULL DEFAULT (now()),
  UNIQUE(story_id, user_id)
);

CREATE INDEX idx_story_views_story ON story_views (story_id, viewed_at DESC);
CREATE INDEX idx_story_views_user ON story_views (user_id, viewed_at DESC);

CREATE TABLE story_reactions (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  story_id uuid NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  emoji varchar(10) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT (now()),
  UNIQUE(story_id, user_id)
);

CREATE INDEX idx_story_reactions_story ON story_reactions (story_id, created_at DESC);
CREATE INDEX idx_story_reactions_user ON story_reactions (user_id, created_at DESC);
