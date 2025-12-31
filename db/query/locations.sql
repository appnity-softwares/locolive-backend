-- name: CreateLocation :one
INSERT INTO locations (
  user_id,
  geohash,
  geom,
  time_bucket,
  expires_at
) VALUES (
  @user_id, @geohash, ST_SetSRID(ST_MakePoint(@lng::float8, @lat::float8), 4326), @time_bucket, @expires_at
) RETURNING *;

-- name: DeleteExpiredLocations :exec
DELETE FROM locations
WHERE expires_at < now();

-- name: GetHeatmapData :many
SELECT 
  ST_X(ST_SnapToGrid(geom, 0.001)) as longitude,
  ST_Y(ST_SnapToGrid(geom, 0.001)) as latitude,
  COUNT(*) as weight
FROM locations
WHERE time_bucket > NOW() - INTERVAL '1 hour'
GROUP BY ST_SnapToGrid(geom, 0.001);
