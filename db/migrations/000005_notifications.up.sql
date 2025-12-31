CREATE TYPE notification_type AS ENUM (
  'connection_request',
  'connection_accepted',
  'crossing_detected',
  'message_received',
  'story_reaction'
);

CREATE TABLE notifications (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type notification_type NOT NULL,
  title varchar NOT NULL,
  message text NOT NULL,
  related_user_id uuid REFERENCES users(id) ON DELETE CASCADE,
  related_story_id uuid REFERENCES stories(id) ON DELETE CASCADE,
  related_crossing_id uuid REFERENCES crossings(id) ON DELETE CASCADE,
  is_read boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX idx_notifications_user_id ON notifications (user_id, created_at DESC);
CREATE INDEX idx_notifications_unread ON notifications (user_id, is_read) WHERE is_read = false;
