#!/bin/bash

DB_HOST="localhost"
DB_PORT="5432"
DB_USER="test_user"
DB_NAME="testdb"

echo "Creating role if it doesn't exist..."
PGPASSWORD="postgres" psql -h "$DB_HOST" -p "$DB_PORT" -U "postgres" -d "$DB_NAME" -c "DO \$\$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '$DB_USER') THEN CREATE ROLE $DB_USER LOGIN PASSWORD 'test_password'; END IF; END \$\$;"


echo "Applying migrations from db/migrations/..."

# Применяем миграции и проверяем статус выполнения каждой
all_migrations_successful=true

for file in db/migrations/*.up.sql; do
    echo "Applying migration: $file"
    PGPASSWORD="postgres" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$file"
    if [ $? -ne 0 ]; then
        echo "Migration $file failed!"
        all_migrations_successful=false
    fi
done

if $all_migrations_successful; then
    echo "Migrations applied successfully!"
else
    echo "Some migrations failed."
fi
