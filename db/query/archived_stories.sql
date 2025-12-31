-- name: ArchiveStory :one
INSERT INTO archived_stories (
    user_id, story_id, media_url, media_type, caption,
    geohash, geom, is_anonymous, show_location, original_created_at
)
SELECT 
    s.user_id, s.id, s.media_url, s.media_type, s.caption,
    s.geohash, s.geom, s.is_anonymous, s.show_location, s.created_at
FROM stories s
WHERE s.id = $1 AND s.user_id = $2
ON CONFLICT (user_id, story_id) DO NOTHING
RETURNING *;

-- name: GetArchivedStories :many
SELECT * FROM archived_stories
WHERE user_id = $1
ORDER BY archived_at DESC
LIMIT $2 OFFSET $3;

-- name: GetArchivedStory :one
SELECT * FROM archived_stories
WHERE id = $1 AND user_id = $2;

-- name: DeleteArchivedStory :exec
DELETE FROM archived_stories
WHERE id = $1 AND user_id = $2;

-- name: CountArchivedStories :one
SELECT COUNT(*) FROM archived_stories
WHERE user_id = $1;
