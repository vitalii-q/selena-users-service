#!/bin/bash

DB_HOST="localhost"
DB_PORT="9265"
DB_USER="postgres"
DB_NAME="users_db"

echo "Applying migrations from db/migrations/..."

for file in migrations/*.up.sql; do
    echo "Applying migration: $file"
    PGPASSWORD="postgres" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$file"
done

echo "Migrations applied successfully!"
