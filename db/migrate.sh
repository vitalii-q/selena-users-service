#!/bin/bash

echo "Applying migrations from db/migrations/..."

for file in db/migrations/*.up.sql; do
    echo "Applying migration: $file"
    PGPASSWORD="postgres" psql -h "${LOCALHOST}" -p "${USERS_POSTGRES_DB_PORT}" -U "${USERS_POSTGRES_DB_USER}" -d "${USERS_POSTGRES_DB_NAME}" -f "$file"
done

echo "Migrations applied successfully!"
