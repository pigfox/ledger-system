#!/bin/bash
set -euo pipefail

set -x
set -e
clear
echo "This script will drop and recreate the database. All data will be lost. Required for first time"
read -p "Do you want to continue? (y/N): " confirm
if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
  echo "❌ Operation cancelled."
  exit 1
fi

# Load environment variables
if [[ ! -f ./.env ]]; then
  echo "❌ .env file not found"
  exit 1
fi
. ./.env
CONTAINER_NAME="$CONTAINER_NAME"
POSTGRES_USER="$POSTGRES_USER"
POSTGRES_DB="$POSTGRES_DB"
POSTGRES_PASSWORD="$POSTGRES_PASSWORD"
SCHEMA_FILE="migrations/001_init.sql"
echo "🧨 Dropping and recreating $POSTGRES_DB..."

# Drop and recreate the database inside the container
docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" -i "$CONTAINER_NAME" psql -U "$POSTGRES_USER" -d postgres <<EOF
REVOKE CONNECT ON DATABASE $POSTGRES_DB FROM public;
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$POSTGRES_DB';
DROP DATABASE IF EXISTS $POSTGRES_DB;
CREATE DATABASE $POSTGRES_DB;
EOF

echo "Database recreated."

# Re-apply schema
echo "Applying migrations from $SCHEMA_FILE..."
docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" -i "$CONTAINER_NAME" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" < "$SCHEMA_FILE"
# echo "Seeding data..."
# docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" < migrations/002_seed.sql

echo "Reset complete."