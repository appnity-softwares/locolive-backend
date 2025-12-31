-- Add read_at column to messages for read receipts
ALTER TABLE messages ADD COLUMN read_at timestamptz;

-- Create message_reactions table
CREATE TABLE message_reactions (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id uuid NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    emoji varchar(10) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT (now()),
    UNIQUE(message_id, user_id, emoji)
);

CREATE INDEX idx_message_reactions_message ON message_reactions(message_id);
