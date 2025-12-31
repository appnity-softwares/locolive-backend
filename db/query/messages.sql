-- name: CreateMessage :one
INSERT INTO messages (
  sender_id,
  receiver_id,
  group_id,
  content,
  media_url,
  media_type,
  expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: ListMessages :many
SELECT m.*,
       COALESCE(
           (SELECT json_agg(json_build_object(
               'id', mr.id,
               'emoji', mr.emoji,
               'user_id', mr.user_id,
               'username', u.username,
               'avatar_url', u.avatar_url,
               'created_at', mr.created_at
           ) ORDER BY mr.created_at ASC)
            FROM message_reactions mr
            JOIN users u ON mr.user_id = u.id
            WHERE mr.message_id = m.id),
           '[]'::json
       ) as reactions
FROM messages m
WHERE ((m.sender_id = $1 AND m.receiver_id = $2)
   OR (m.sender_id = $2 AND m.receiver_id = $1))
   AND m.group_id IS NULL
   AND (m.expires_at IS NULL OR m.expires_at > NOW())
ORDER BY m.created_at ASC;

-- name: GetGroupMessages :many
SELECT m.*, 
       u.username, 
       u.avatar_url,
       COALESCE(
           (SELECT json_agg(json_build_object(
               'id', mr.id,
               'emoji', mr.emoji,
               'user_id', mr.user_id,
               'username', reaction_user.username,
               'avatar_url', reaction_user.avatar_url,
               'created_at', mr.created_at
           ) ORDER BY mr.created_at ASC)
            FROM message_reactions mr
            JOIN users reaction_user ON mr.user_id = reaction_user.id
            WHERE mr.message_id = m.id),
           '[]'::json
       ) as reactions
FROM messages m
JOIN users u ON m.sender_id = u.id
WHERE m.group_id = $1
ORDER BY m.created_at ASC;


-- name: DeleteOldMessages :exec
-- Delete messages older than specified days (default: 30 days)
DELETE FROM messages
WHERE created_at < NOW() - INTERVAL '30 days';

-- name: DeleteExpiredMessages :exec
DELETE FROM messages
WHERE expires_at IS NOT NULL AND expires_at < NOW();

-- name: DeleteMessage :exec
DELETE FROM messages
WHERE id = $1 AND sender_id = $2;

-- name: UpdateMessage :one
UPDATE messages
SET content = $3, media_url = $4, media_type = $5
WHERE id = $1 AND sender_id = $2
RETURNING *;

-- name: SaveMessage :one
UPDATE messages
SET expires_at = NULL
WHERE id = $1
RETURNING *;

-- name: GetMessage :one
SELECT * FROM messages WHERE id = $1;

-- name: MarkMessageRead :one
UPDATE messages
SET read_at = NOW()
WHERE id = $1 AND receiver_id = $2 AND read_at IS NULL
RETURNING *;

-- name: MarkConversationRead :exec
UPDATE messages
SET read_at = NOW()
WHERE receiver_id = $1 AND sender_id = $2 AND read_at IS NULL;

-- name: CreateMessageReaction :one
INSERT INTO message_reactions (message_id, user_id, emoji)
VALUES ($1, $2, $3)
ON CONFLICT (message_id, user_id, emoji) DO NOTHING
RETURNING *;

-- name: DeleteMessageReaction :exec
DELETE FROM message_reactions
WHERE message_id = $1 AND user_id = $2 AND emoji = $3;

-- name: GetMessageReactions :many
SELECT mr.*, u.username, u.avatar_url
FROM message_reactions mr
JOIN users u ON mr.user_id = u.id
WHERE mr.message_id = $1

ORDER BY mr.created_at ASC;

-- name: GetUnreadMessageCount :one
SELECT COUNT(*) FROM messages
WHERE receiver_id = $1 AND read_at IS NULL;

-- name: GetConversationList :many
WITH conversation_partners AS (
  SELECT DISTINCT
    CASE 
      WHEN sender_id = $1 THEN receiver_id
      ELSE sender_id
    END as partner_id
  FROM messages
  WHERE sender_id = $1 OR receiver_id = $1
),
latest_messages AS (
  SELECT DISTINCT ON (
    CASE 
      WHEN m.sender_id = $1 THEN m.receiver_id
      ELSE m.sender_id
    END
  )
    CASE 
      WHEN m.sender_id = $1 THEN m.receiver_id
      ELSE m.sender_id
    END as partner_id,
    m.id as message_id,
    m.content as last_message,
    m.created_at as last_message_at,
    m.sender_id as last_sender_id
  FROM messages m
  WHERE (m.sender_id = $1 OR m.receiver_id = $1)
    AND (m.expires_at IS NULL OR m.expires_at > NOW())
  ORDER BY 
    CASE 
      WHEN m.sender_id = $1 THEN m.receiver_id
      ELSE m.sender_id
    END,
    m.created_at DESC
)
SELECT 
  u.id,
  u.username,
  u.full_name,
  u.avatar_url,
  lm.last_message,
  lm.last_message_at,
  lm.last_sender_id,
  COALESCE(
    (SELECT COUNT(*) 
     FROM messages m2
     WHERE m2.sender_id = u.id 
       AND m2.receiver_id = $1 
       AND m2.read_at IS NULL
       AND (m2.expires_at IS NULL OR m2.expires_at > NOW())
    ), 0
  ) as unread_count
FROM conversation_partners cp
JOIN users u ON u.id = cp.partner_id
JOIN latest_messages lm ON lm.partner_id = cp.partner_id
ORDER BY lm.last_message_at DESC;

-- name: DeleteConversation :exec
DELETE FROM messages
WHERE (sender_id = $1 AND receiver_id = $2)
   OR (sender_id = $2 AND receiver_id = $1);
