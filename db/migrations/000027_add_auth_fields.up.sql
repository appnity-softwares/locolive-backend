ALTER TABLE users ADD COLUMN google_id VARCHAR UNIQUE;
ALTER TABLE users ADD COLUMN password_reset_token VARCHAR;
ALTER TABLE users ADD COLUMN password_reset_expires_at TIMESTAMPTZ;

CREATE INDEX idx_users_google_id ON users(google_id);
CREATE INDEX idx_users_reset_token ON users(password_reset_token);
