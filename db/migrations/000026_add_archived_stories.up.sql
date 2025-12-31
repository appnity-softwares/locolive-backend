CREATE TABLE archived_stories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    story_id UUID NOT NULL,
    media_url TEXT NOT NULL,
    media_type TEXT NOT NULL,
    caption TEXT,
    geohash TEXT NOT NULL,
    geom GEOGRAPHY(POINT, 4326) NOT NULL,
    is_anonymous BOOLEAN DEFAULT false,
    show_location BOOLEAN DEFAULT true,
    original_created_at TIMESTAMPTZ NOT NULL,
    archived_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, story_id)
);

CREATE INDEX idx_archived_stories_user_id ON archived_stories(user_id);
CREATE INDEX idx_archived_stories_archived_at ON archived_stories(archived_at DESC);
