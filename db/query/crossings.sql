-- name: CreateCrossing :one
INSERT INTO crossings (
  user_id_1,
  user_id_2,
  location_center,
  occurred_at
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetCrossingsForUser :many
SELECT c.* FROM crossings c
JOIN users u1 ON c.user_id_1 = u1.id
JOIN users u2 ON c.user_id_2 = u2.id
WHERE 
  (c.user_id_1 = $1 OR c.user_id_2 = $1)
  -- Filter out ghost mode users (other user)
  AND (
    (c.user_id_1 = $1 AND u2.is_ghost_mode = false) OR
    (c.user_id_2 = $1 AND u1.is_ghost_mode = false)
  )
  -- strict streak visibility rule
  AND (
    (c.user_id_1 = $1 AND DATE(u2.last_active_at) >= CURRENT_DATE - INTERVAL '1 day') OR
    (c.user_id_2 = $1 AND DATE(u1.last_active_at) >= CURRENT_DATE - INTERVAL '1 day')
  )
  -- Shadow Ban Filter
  AND (
    (c.user_id_1 = $1 AND u2.is_shadow_banned = false) OR
    (c.user_id_2 = $1 AND u1.is_shadow_banned = false)
  )
  -- Block Logic
  AND NOT EXISTS (
    SELECT 1 FROM blocked_users bu 
    WHERE (bu.blocker_id = $1 AND bu.blocked_id = CASE WHEN c.user_id_1 = $1 THEN c.user_id_2 ELSE c.user_id_1 END)
       OR (bu.blocker_id = CASE WHEN c.user_id_1 = $1 THEN c.user_id_2 ELSE c.user_id_1 END AND bu.blocked_id = $1)
  )
ORDER BY c.occurred_at DESC;

-- name: CountCrossingsToday :one
SELECT COUNT(*) FROM crossings
WHERE (user_id_1 = $1 OR user_id_2 = $1)
AND occurred_at >= CURRENT_DATE;

-- name: FindPotentialCrossings :many
SELECT 
    l1.user_id AS user1, 
    l2.user_id AS user2, 
    l1.geohash AS location, 
    l1.time_bucket
FROM locations l1
JOIN locations l2 ON l1.geohash = l2.geohash AND l1.time_bucket = l2.time_bucket
JOIN users u1 ON l1.user_id = u1.id
JOIN users u2 ON l2.user_id = u2.id
WHERE l1.user_id < l2.user_id
AND l1.time_bucket >= @min_time::timestamptz
AND l1.time_bucket < @max_time::timestamptz
AND u1.is_ghost_mode = false
AND u2.is_ghost_mode = false
AND u1.is_shadow_banned = false
AND u2.is_shadow_banned = false
-- Block Logic
AND NOT EXISTS (
    SELECT 1 FROM blocked_users bu 
    WHERE (bu.blocker_id = l1.user_id AND bu.blocked_id = l2.user_id)
       OR (bu.blocker_id = l2.user_id AND bu.blocked_id = l1.user_id)
)
GROUP BY l1.user_id, l2.user_id, l1.geohash, l1.time_bucket;
