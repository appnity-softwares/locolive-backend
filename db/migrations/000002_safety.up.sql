CREATE TYPE report_reason AS ENUM ('spam', 'abuse', 'inappropriate', 'other');

CREATE TABLE reports (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  reporter_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  target_user_id uuid REFERENCES users(id) ON DELETE CASCADE,
  target_story_id uuid REFERENCES stories(id) ON DELETE CASCADE,
  reason report_reason NOT NULL,
  description text,
  is_resolved boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX idx_reports_unresolved ON reports (created_at) WHERE is_resolved = false;
