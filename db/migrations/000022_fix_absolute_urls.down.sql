UPDATE stories 
SET media_url = REPLACE(media_url, '/api/uploads/', 'http://localhost:8080/uploads/')
WHERE media_url LIKE '/api/uploads/%';

UPDATE stories 
SET thumbnail_url = REPLACE(thumbnail_url, '/api/uploads/', 'http://localhost:8080/uploads/')
WHERE thumbnail_url LIKE '/api/uploads/%';

UPDATE users 
SET avatar_url = REPLACE(avatar_url, '/api/uploads/', 'http://localhost:8080/uploads/')
WHERE avatar_url LIKE '/api/uploads/%';

UPDATE messages
SET media_url = REPLACE(media_url, '/api/uploads/', 'http://localhost:8080/uploads/')
WHERE media_url LIKE '/api/uploads/%';
