-- Rollback profile fields
ALTER TABLE users DROP COLUMN IF EXISTS bio;
ALTER TABLE users DROP COLUMN IF EXISTS banner_url;
ALTER TABLE users DROP COLUMN IF EXISTS theme;
ALTER TABLE users DROP COLUMN IF EXISTS profile_visibility;
