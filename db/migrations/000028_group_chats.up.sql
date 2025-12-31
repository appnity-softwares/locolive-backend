CREATE TABLE groups (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR NOT NULL,
  description TEXT,
  created_by UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  image_url VARCHAR
);

CREATE TABLE group_members (
  group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role VARCHAR NOT NULL DEFAULT 'member',
  joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (group_id, user_id)
);

ALTER TABLE messages
ADD COLUMN group_id UUID REFERENCES groups(id) ON DELETE CASCADE;

ALTER TABLE messages
ALTER COLUMN receiver_id DROP NOT NULL;

ALTER TABLE messages
ADD CONSTRAINT chk_receiver_or_group CHECK (
  (receiver_id IS NOT NULL AND group_id IS NULL) OR
  (receiver_id IS NULL AND group_id IS NOT NULL)
);

CREATE INDEX idx_group_members_user_id ON group_members(user_id);
CREATE INDEX idx_messages_group_id ON messages(group_id);
