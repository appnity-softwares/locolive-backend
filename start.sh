#!/bin/sh

set -e

echo "Running database migrations..."
/usr/bin/migrate -path /app/db/migrations -database "$DB_SOURCE" -verbose up

echo "Starting the application..."
exec "$@"
