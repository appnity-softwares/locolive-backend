-- Revert constraints
ALTER TABLE profile_views DROP CONSTRAINT IF EXISTS profile_views_viewer_id_viewed_user_id_key;

-- Add back old constraint
ALTER TABLE profile_views ADD CONSTRAINT profile_views_viewer_id_viewed_user_id_viewed_at_key UNIQUE (viewer_id, viewed_user_id, viewed_at);
