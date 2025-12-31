-- name: CreateReport :one
INSERT INTO reports (
  reporter_id,
  target_user_id,
  target_story_id,
  reason,
  description
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- Admin: List all reports
-- name: ListReports :many
SELECT r.*, 
  u1.username as reporter_username,
  u2.username as target_username
FROM reports r
LEFT JOIN users u1 ON r.reporter_id = u1.id
LEFT JOIN users u2 ON r.target_user_id = u2.id
WHERE is_resolved = $1
ORDER BY r.created_at DESC
LIMIT $2 OFFSET $3;

-- Admin: Resolve report
-- name: ResolveReport :one
UPDATE reports
SET is_resolved = true
WHERE id = $1
RETURNING *;
