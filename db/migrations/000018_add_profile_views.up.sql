-- Create profile_views table to track who viewed whose profile
CREATE TABLE IF NOT EXISTS profile_views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    viewer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    viewed_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    viewed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Prevent duplicate views within same minute
    UNIQUE(viewer_id, viewed_user_id, viewed_at)
);

-- Index for fast queries: "Who viewed my profile recently?"
CREATE INDEX idx_profile_views_viewed_user ON profile_views(viewed_user_id, viewed_at DESC);

-- Index for analytics: "Whose profiles did I view?"
CREATE INDEX idx_profile_views_viewer ON profile_views(viewer_id, viewed_at DESC);

-- Prevent self-views
ALTER TABLE profile_views ADD CONSTRAINT no_self_views CHECK (viewer_id != viewed_user_id);
