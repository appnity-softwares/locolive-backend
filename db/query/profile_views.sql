-- name: TrackProfileView :one
INSERT INTO profile_views (viewer_id, viewed_user_id, viewed_at)
VALUES ($1, $2, NOW())
ON CONFLICT (viewer_id, viewed_user_id) DO UPDATE
SET viewed_at = NOW()
RETURNING *;

-- name: GetRecentProfileVisitors :many
SELECT 
    u.id,
    u.username,
    u.full_name,
    u.avatar_url,
    pv.viewed_at
FROM profile_views pv
JOIN users u ON u.id = pv.viewer_id
WHERE pv.viewed_user_id = $1
  AND pv.viewed_at >= NOW() - INTERVAL '24 hours'
ORDER BY pv.viewed_at DESC
LIMIT 50;

-- name: GetProfileViewCount :one
SELECT COUNT(DISTINCT viewer_id) as total_views
FROM profile_views
WHERE viewed_user_id = $1
  AND viewed_at >= NOW() - INTERVAL '30 days';

-- name: GetMyProfileViews :many
SELECT 
    u.id,
    u.username,
    u.full_name,
    u.avatar_url,
    pv.viewed_at
FROM profile_views pv
JOIN users u ON u.id = pv.viewed_user_id
WHERE pv.viewer_id = $1
ORDER BY pv.viewed_at DESC
LIMIT 50;
