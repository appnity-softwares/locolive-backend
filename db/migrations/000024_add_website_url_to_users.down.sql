-- Remove website_url column from users table
ALTER TABLE users DROP COLUMN IF EXISTS website_url;
