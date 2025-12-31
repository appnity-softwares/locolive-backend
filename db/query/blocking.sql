-- name: BlockUser :one
INSERT INTO blocked_users (blocker_id, blocked_id)
VALUES ($1, $2)
RETURNING *;

-- name: UnblockUser :exec
DELETE FROM blocked_users
WHERE blocker_id = $1 AND blocked_id = $2;

-- name: GetBlockedUsers :many
SELECT u.id, u.username, u.full_name, u.avatar_url, b.created_at as blocked_at
FROM blocked_users b
JOIN users u ON b.blocked_id = u.id
WHERE b.blocker_id = $1
ORDER BY b.created_at DESC;

-- name: IsUserBlocked :one
SELECT EXISTS (
    SELECT 1 FROM blocked_users
    WHERE blocker_id = $1 AND blocked_id = $2
);
