-- name: CreateConnectionRequest :one
INSERT INTO connections (
  requester_id,
  target_id,
  status
) VALUES (
  $1, $2, 'pending'
) RETURNING *;

-- name: UpdateConnectionStatus :one
UPDATE connections
SET status = $3, updated_at = now()
WHERE requester_id = $1 AND target_id = $2
RETURNING *;

-- name: CountConnectionRequestsToday :one
SELECT COUNT(*) FROM connections
WHERE requester_id = $1
AND created_at > NOW() - INTERVAL '24 hours'
AND status = 'pending';

-- name: GetConnection :one
SELECT * FROM connections
WHERE (requester_id = $1 AND target_id = $2)
   OR (requester_id = $2 AND target_id = $1)
LIMIT 1;

-- name: ListConnections :many
SELECT 
    u.id, 
    u.username, 
    u.full_name, 
    u.avatar_url,
    u.last_active_at
FROM connections c
JOIN users u ON (u.id = c.requester_id OR u.id = c.target_id)
WHERE (c.requester_id = $1 OR c.target_id = $1)
  AND u.id != $1
  AND c.status = 'accepted';

-- name: ListPendingRequests :many
SELECT 
    c.requester_id, 
    c.target_id, 
    c.status, 
    c.created_at,
    u.username,
    u.full_name,
    u.avatar_url
FROM connections c
JOIN users u ON c.requester_id = u.id
WHERE c.target_id = $1 
  AND c.status = 'pending'
ORDER BY c.created_at DESC;

-- name: ListSentConnectionRequests :many
SELECT 
    c.requester_id, 
    c.target_id, 
    c.status, 
    c.created_at,
    u.username,
    u.full_name,
    u.avatar_url
FROM connections c
JOIN users u ON c.target_id = u.id
WHERE c.requester_id = $1 
  AND c.status = 'pending'
ORDER BY c.created_at DESC;

-- name: DeleteConnection :exec
DELETE FROM connections
WHERE (requester_id = $1 AND target_id = $2)
   OR (requester_id = $2 AND target_id = $1);

-- name: GetSuggestedConnections :many
WITH my_connections AS (
    SELECT c1.target_id as friend_id FROM connections c1 WHERE c1.requester_id = $1 AND c1.status = 'accepted'
    UNION
    SELECT c2.requester_id as friend_id FROM connections c2 WHERE c2.target_id = $1 AND c2.status = 'accepted'
),
excluded_users AS (
    SELECT c3.target_id as id FROM connections c3 WHERE c3.requester_id = $1
    UNION
    SELECT c4.requester_id as id FROM connections c4 WHERE c4.target_id = $1
    UNION
    SELECT $1::uuid as id
)
SELECT 
    u.id, 
    u.username, 
    u.full_name, 
    u.avatar_url,
    COALESCE((
        SELECT COUNT(*)
        FROM connections c
        WHERE 
            c.status = 'accepted' AND (
                (c.requester_id = u.id AND c.target_id IN (SELECT friend_id FROM my_connections)) OR
                (c.target_id = u.id AND c.requester_id IN (SELECT friend_id FROM my_connections))
            )
    ), 0)::bigint as mutual_count
FROM users u
WHERE u.id NOT IN (SELECT id FROM excluded_users)
AND u.is_shadow_banned = false
ORDER BY mutual_count DESC, u.created_at DESC
LIMIT $2;
