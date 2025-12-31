-- Add links JSONB column to users table
ALTER TABLE users ADD COLUMN links JSONB DEFAULT '[]'::jsonb NOT NULL;

-- Migrate existing website_url to links array if it exists
UPDATE users 
SET links = jsonb_build_array(jsonb_build_object('label', 'Website', 'url', website_url))
WHERE website_url IS NOT NULL AND website_url != '';
