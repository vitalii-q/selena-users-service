#!/bin/bash

DB_HOST="localhost"
DB_PORT="9265"
DB_USER="postgres"
DB_NAME="users_db"

echo "Rolling back migrations from db/migrations/..."

for file in $(ls -r db/migrations/*.down.sql); do
    echo "Rolling back: $file"
    PGPASSWORD="postgres" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$file"
done

echo "Rollback completed!"
