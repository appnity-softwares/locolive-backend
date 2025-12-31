-- name: CreateGroup :one
INSERT INTO groups (
  name,
  description,
  created_by
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: AddGroupMember :one
INSERT INTO group_members (
  group_id,
  user_id,
  role
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: RemoveGroupMember :exec
DELETE FROM group_members
WHERE group_id = $1 AND user_id = $2;

-- name: GetUserGroups :many
SELECT g.* FROM groups g
JOIN group_members gm ON g.id = gm.group_id
WHERE gm.user_id = $1
ORDER BY g.created_at DESC;

-- name: GetGroupMembers :many
SELECT gm.*, u.username, u.avatar_url FROM group_members gm
JOIN users u ON gm.user_id = u.id
WHERE gm.group_id = $1;

-- name: GetGroupByID :one
SELECT * FROM groups
WHERE id = $1 LIMIT 1;

-- name: CheckGroupMembership :one
SELECT EXISTS (
  SELECT 1 FROM group_members
  WHERE group_id = $1 AND user_id = $2
);
