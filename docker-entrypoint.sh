#!/bin/sh
set -e

echo "==> Running database migrations..."
migrate -path=/app/migrations \
  -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_SSLMODE}" \
  up
echo "==> Migrations done!"

echo "==> Running seed (initial data)..."
/app/seed || echo "==> Seed skipped (already seeded or failed — continuing)"

echo "==> Starting API server..."
exec /app/api
