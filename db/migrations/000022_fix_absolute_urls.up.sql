UPDATE stories 
SET media_url = REPLACE(media_url, 'http://localhost:8080/uploads/', '/api/uploads/')
WHERE media_url LIKE 'http://localhost:8080/uploads/%';

UPDATE stories 
SET thumbnail_url = REPLACE(thumbnail_url, 'http://localhost:8080/uploads/', '/api/uploads/')
WHERE thumbnail_url LIKE 'http://localhost:8080/uploads/%';

UPDATE users 
SET avatar_url = REPLACE(avatar_url, 'http://localhost:8080/uploads/', '/api/uploads/')
WHERE avatar_url LIKE 'http://localhost:8080/uploads/%';

UPDATE messages
SET media_url = REPLACE(media_url, 'http://localhost:8080/uploads/', '/api/uploads/')
WHERE media_url LIKE 'http://localhost:8080/uploads/%';
