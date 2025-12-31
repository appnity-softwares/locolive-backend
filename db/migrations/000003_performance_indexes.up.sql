-- Phase 1: Database Indexes (Run this immediately)
-- Estimated time: 10 minutes
-- Impact: 40-60% query speed improvement

-- User indexes
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Story indexes
CREATE INDEX IF NOT EXISTS idx_stories_user_id ON stories(user_id);
CREATE INDEX IF NOT EXISTS idx_stories_created_at ON stories(created_at);
CREATE INDEX IF NOT EXISTS idx_stories_expires_at ON stories(expires_at);
CREATE INDEX IF NOT EXISTS idx_stories_geohash ON stories(geohash);

-- Location indexes
CREATE INDEX IF NOT EXISTS idx_locations_user_id ON locations(user_id);
CREATE INDEX IF NOT EXISTS idx_locations_expires_at ON locations(expires_at);
CREATE INDEX IF NOT EXISTS idx_locations_time_bucket ON locations(time_bucket);
CREATE INDEX IF NOT EXISTS idx_locations_geohash ON locations(geohash);

-- Connection indexes
CREATE INDEX IF NOT EXISTS idx_connections_requester_id ON connections(requester_id);
CREATE INDEX IF NOT EXISTS idx_connections_target_id ON connections(target_id);
CREATE INDEX IF NOT EXISTS idx_connections_status ON connections(status);

-- Session indexes
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- Message indexes
CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_receiver_id ON messages(receiver_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);

-- Report indexes
CREATE INDEX IF NOT EXISTS idx_reports_reporter_id ON reports(reporter_id);
CREATE INDEX IF NOT EXISTS idx_reports_is_resolved ON reports(is_resolved);
CREATE INDEX IF NOT EXISTS idx_reports_created_at ON reports(created_at);

-- Crossing indexes
CREATE INDEX IF NOT EXISTS idx_crossings_user_id_1 ON crossings(user_id_1);
CREATE INDEX IF NOT EXISTS idx_crossings_user_id_2 ON crossings(user_id_2);
CREATE INDEX IF NOT EXISTS idx_crossings_occurred_at ON crossings(occurred_at);

-- Analyze tables after index creation
ANALYZE users;
ANALYZE stories;
ANALYZE locations;
ANALYZE connections;
ANALYZE sessions;
ANALYZE messages;
ANALYZE reports;
ANALYZE crossings;
