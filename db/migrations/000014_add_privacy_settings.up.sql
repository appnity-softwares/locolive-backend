-- Create privacy settings table
CREATE TABLE IF NOT EXISTS privacy_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    who_can_message VARCHAR(20) DEFAULT 'connections' CHECK (who_can_message IN ('everyone', 'connections', 'nobody')),
    who_can_see_stories VARCHAR(20) DEFAULT 'connections' CHECK (who_can_see_stories IN ('everyone', 'connections', 'nobody')),
    show_location BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_privacy_settings_user ON privacy_settings(user_id);
