-- Remove duplicate views, keeping only the latest one per viewer-viewed pair
DELETE FROM profile_views
WHERE id NOT IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER (
            PARTITION BY viewer_id, viewed_user_id 
            ORDER BY viewed_at DESC
        ) as rn
        FROM profile_views
    ) t WHERE t.rn = 1
);

-- Drop old ineffective constraint
ALTER TABLE profile_views DROP CONSTRAINT IF EXISTS profile_views_viewer_id_viewed_user_id_viewed_at_key;

-- Add new constraint for true uniqueness per pair (allowing updates on conflict)
ALTER TABLE profile_views ADD CONSTRAINT profile_views_viewer_id_viewed_user_id_key UNIQUE (viewer_id, viewed_user_id);
