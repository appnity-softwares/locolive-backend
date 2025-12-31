-- name: CreateStoryMention :one
INSERT INTO story_mentions (
  story_id,
  mentioned_user_id
) VALUES (
  $1, $2
) ON CONFLICT (story_id, mentioned_user_id) DO NOTHING
RETURNING *;

-- name: GetStoryMentions :many
SELECT sm.*, u.username, u.avatar_url
FROM story_mentions sm
JOIN users u ON sm.mentioned_user_id = u.id
WHERE sm.story_id = $1
ORDER BY sm.created_at DESC;

-- name: GetUserMentions :many
SELECT sm.*, s.media_url, s.media_type, u.username as story_author
FROM story_mentions sm
JOIN stories s ON sm.story_id = s.id
JOIN users u ON s.user_id = u.id
WHERE sm.mentioned_user_id = $1
AND s.expires_at > now()
ORDER BY sm.created_at DESC
LIMIT $2 OFFSET $3;

-- name: DeleteStoryMentions :exec
DELETE FROM story_mentions
WHERE story_id = $1;
