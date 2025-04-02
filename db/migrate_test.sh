#!/bin/bash

DB_HOST="localhost"
DB_PORT="5432"
DB_USER="test_user"
DB_NAME="testdb"

echo "Applying migrations from db/migrations/..."

for file in db/migrations/*.up.sql; do
    echo "Applying migration: $file"
    PGPASSWORD="postgres" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$file"
done

echo "Migrations applied successfully!"
