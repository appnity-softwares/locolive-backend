-- Rollback performance indexes

DROP INDEX CONCURRENTLY IF EXISTS idx_users_phone;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_username;

DROP INDEX CONCURRENTLY IF EXISTS idx_stories_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_stories_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_stories_expires_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_stories_geohash;

DROP INDEX CONCURRENTLY IF EXISTS idx_locations_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_locations_expires_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_locations_time_bucket;
DROP INDEX CONCURRENTLY IF EXISTS idx_locations_geohash;

DROP INDEX CONCURRENTLY IF EXISTS idx_connections_requester_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_connections_target_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_connections_status;

DROP INDEX CONCURRENTLY IF EXISTS idx_sessions_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_sessions_expires_at;

DROP INDEX CONCURRENTLY IF EXISTS idx_messages_sender_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_messages_receiver_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_messages_created_at;

DROP INDEX CONCURRENTLY IF EXISTS idx_reports_reporter_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_reports_is_resolved;
DROP INDEX CONCURRENTLY IF EXISTS idx_reports_created_at;

DROP INDEX CONCURRENTLY IF EXISTS idx_crossings_user_id_1;
DROP INDEX CONCURRENTLY IF EXISTS idx_crossings_user_id_2;
DROP INDEX CONCURRENTLY IF EXISTS idx_crossings_occurred_at;
