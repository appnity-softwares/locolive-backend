-- name: GetPrivacySettings :one
SELECT * FROM privacy_settings WHERE user_id = $1;

-- name: UpsertPrivacySettings :one
INSERT INTO privacy_settings (
    user_id, who_can_message, who_can_see_stories, show_location
) VALUES (
    $1, $2, $3, $4
) ON CONFLICT (user_id) DO UPDATE
SET 
    who_can_message = EXCLUDED.who_can_message,
    who_can_see_stories = EXCLUDED.who_can_see_stories,
    show_location = EXCLUDED.show_location,
    updated_at = NOW()
RETURNING *;
