-- name: GetStreakRetentionStats :one
SELECT 
    COUNT(*) FILTER (WHERE activity_streak >= 3) as retained_users_count,
    COUNT(*) as total_users_count,
    (COUNT(*) FILTER (WHERE activity_streak >= 3)::float / NULLIF(COUNT(*), 0)::float) * 100 as retention_rate
FROM users;

-- name: GetEngagementStats :one
SELECT 
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '7 days') as stories_last_7d,
    (COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '7 days')::float / NULLIF((SELECT COUNT(*) FROM users), 0)::float) as avg_stories_per_user_weekly
FROM stories;

-- name: GetConversionStats :one
WITH crossing_stats AS (
    SELECT COUNT(*) as total_crossings FROM crossings
),
connection_stats AS (
    SELECT COUNT(*) as total_connections FROM connections WHERE status = 'accepted'
)
SELECT 
    (SELECT total_connections FROM connection_stats) as total_connections,
    (SELECT total_crossings FROM crossing_stats) as total_crossings,
    -- Simple Ratio for now (Crossings -> Connections)
    ((SELECT total_connections FROM connection_stats)::float / NULLIF((SELECT total_crossings FROM crossing_stats), 0)::float) * 100 as crossing_conversion_rate
FROM crossing_stats, connection_stats;
