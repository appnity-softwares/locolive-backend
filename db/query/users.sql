-- name: CreateUser :one
INSERT INTO users (
  phone,
  password_hash,
  username,
  full_name
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetUserByPhone :one
SELECT * FROM users
WHERE phone = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: UpdateUserTrust :one
UPDATE users
SET trust_level = $2
WHERE id = $1
RETURNING *;

-- Privacy Features

-- name: ToggleGhostMode :one
UPDATE users
SET is_ghost_mode = $2,
    ghost_mode_expires_at = $3
WHERE id = $1
RETURNING *;

-- name: DeleteAllUserData :exec
-- Used for panic mode - deletes all user data
DELETE FROM users
WHERE id = $1;

-- Admin Queries

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: BanUser :one
UPDATE users
SET is_shadow_banned = $2
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: UpdateUserActivity :one
-- Updates last_active_at and calculates activity streak
UPDATE users
SET 
  last_active_at = now(),
  activity_streak = CASE
    -- If last active was yesterday, increment streak
    WHEN DATE(last_active_at) = CURRENT_DATE - INTERVAL '1 day' THEN activity_streak + 1
    -- If last active was today, keep streak
    WHEN DATE(last_active_at) = CURRENT_DATE THEN activity_streak
    -- If missed days but has freezes, keep streak
    WHEN streak_freezes_remaining > 0 THEN activity_streak
    -- Otherwise reset streak to 1
    ELSE 1
  END,
  streak_freezes_remaining = CASE
    -- Consume freeze only if missed days and has freezes
    WHEN DATE(last_active_at) < CURRENT_DATE - INTERVAL '1 day' AND streak_freezes_remaining > 0 THEN streak_freezes_remaining - 1
    ELSE streak_freezes_remaining
  END,
  streak_updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetUserActivityStatus :one
-- Get user's activity status and visibility
SELECT 
  id,
  username,
  last_active_at,
  CASE
    WHEN DATE(last_active_at) < CURRENT_DATE - INTERVAL '1 day' THEN 0
    ELSE activity_streak
  END as activity_streak,
  CASE
    WHEN DATE(last_active_at) >= CURRENT_DATE - INTERVAL '1 day' THEN 'active'
    ELSE 'hidden'
  END as visibility_status,
  CASE
    WHEN DATE(last_active_at) >= CURRENT_DATE - INTERVAL '1 day' THEN true
    ELSE false
  END as is_visible
FROM users
WHERE id = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET 
  full_name = COALESCE(sqlc.narg('full_name'), full_name),
  username = COALESCE(sqlc.narg('username'), username),
  avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
  bio = COALESCE(sqlc.narg('bio'), bio),
  banner_url = COALESCE(sqlc.narg('banner_url'), banner_url),
  theme = COALESCE(sqlc.narg('theme'), theme),
  profile_visibility = COALESCE(sqlc.narg('profile_visibility'), profile_visibility),
  website_url = COALESCE(sqlc.narg('website_url'), website_url),
  links = COALESCE(sqlc.narg('links'), links)
WHERE id = $1
RETURNING id, username, full_name, avatar_url, bio, banner_url, theme, profile_visibility, website_url, links, created_at;

-- name: GetUserProfile :one
SELECT 
  u.id, u.username, u.full_name, u.avatar_url, u.bio, u.banner_url, u.theme, u.profile_visibility, u.email, u.is_ghost_mode, u.website_url, u.links, u.created_at, u.is_premium, u.last_active_at,
  (SELECT COUNT(*) FROM stories WHERE stories.user_id = u.id) as story_count,
  (SELECT COUNT(*) FROM connections WHERE (connections.requester_id = u.id OR connections.target_id = u.id) AND status = 'accepted') as connection_count,
  CASE
    WHEN DATE(u.last_active_at) < CURRENT_DATE - INTERVAL '1 day' THEN 0
    ELSE u.activity_streak
  END as activity_streak,
  CASE
    WHEN DATE(u.last_active_at) >= CURRENT_DATE - INTERVAL '1 day' THEN 'active'
    ELSE 'hidden'
  END as visibility_status
FROM users u
WHERE u.id = $1;

-- name: GetUserEngagementStats :one
SELECT 
    (SELECT COUNT(*) FROM stories WHERE stories.user_id = $1) as story_count,
    (SELECT COUNT(*) FROM connections WHERE (connections.requester_id = $1 OR connections.target_id = $1) AND status = 'accepted') as connection_count,
    (SELECT COUNT(*) FROM story_views v JOIN stories s ON v.story_id = s.id WHERE s.user_id = $1) as total_views,
    (SELECT COUNT(*) FROM story_reactions r JOIN stories s ON r.story_id = s.id WHERE s.user_id = $1) as total_reactions;

-- name: GetSystemStats :one
SELECT 
  COUNT(*) as total_users,
  COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '24 hours') as new_users_24h,
  COUNT(*) FILTER (WHERE last_active_at > NOW() - INTERVAL '1 hour') as active_users_1h
FROM users;

-- name: BoostUser :one
UPDATE users
SET boost_expires_at = $2
WHERE id = $1
RETURNING *;

-- name: SearchUsers :many
SELECT 
  id,
  username,
  full_name,
  avatar_url,
  bio,
  is_verified,
  created_at
FROM users
WHERE 
  (username ILIKE '%' || sqlc.arg(query)::text || '%' OR full_name ILIKE '%' || sqlc.arg(query)::text || '%')
  AND is_shadow_banned = false
LIMIT 20;


-- name: UpdateUserEmail :one
UPDATE users
SET email = $2
WHERE id = $1
RETURNING id, username, email, full_name;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByGoogleID :one
SELECT * FROM users
WHERE google_id = $1 LIMIT 1;

-- name: UpdateUserGoogleID :one
UPDATE users
SET google_id = $2
WHERE id = $1
RETURNING *;

-- name: SetPasswordResetToken :one
UPDATE users
SET 
    password_reset_token = $2,
    password_reset_expires_at = $3
WHERE email = $1
RETURNING *;

-- name: GetUserByResetToken :one
SELECT * FROM users
WHERE password_reset_token = $1 
AND password_reset_expires_at > now()
LIMIT 1;

-- name: ClearPasswordResetToken :exec
UPDATE users
SET 
    password_reset_token = NULL,
    password_reset_expires_at = NULL
WHERE id = $1;
