DROP INDEX idx_users_reset_token;
DROP INDEX idx_users_google_id;

ALTER TABLE users DROP COLUMN password_reset_expires_at;
ALTER TABLE users DROP COLUMN password_reset_token;
ALTER TABLE users DROP COLUMN google_id;
