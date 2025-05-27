#!/bin/bash

set -e # Падение скрипта при любой ошибке

# Подключаем переменные окружения из .env
set -o allexport
source ".env"
set +o allexport

# Определяем, где запускаемся: в контейнере или на хосте
if grep -q docker /proc/1/cgroup || [ -f /.dockerenv ]; then
  echo "🧱 Running inside Docker container"
  DB_HOST=${USERS_POSTGRES_DB_HOST}
  DB_PORT=${USERS_POSTGRES_DB_PORT_INNER}
else
  echo "💻 Running on host machine"
  DB_HOST=${LOCALHOST}
  DB_PORT=${USERS_POSTGRES_DB_PORT}
fi

DB_USER="${USERS_POSTGRES_DB_USER}"
DB_NAME="${USERS_POSTGRES_DB_NAME}"
DB_PASS="${USERS_POSTGRES_DB_PASS}"
MIGRATIONS_DIR="db/migrations"

echo "Applying migrations from $MIGRATIONS_DIR..."

FOUND_FILES=false

for file in $MIGRATIONS_DIR/*.up.sql; do
  if [ -f "$file" ]; then
    FOUND_FILES=true
    echo "Applying migration file: $file"
    PGPASSWORD="$DB_PASS" psql \
      -h "$DB_HOST" \
      -p "$DB_PORT" \
      -U "$DB_USER" \
      -d "$DB_NAME" \
      -f "$file"
  fi
done

if [ "$FOUND_FILES" = false ]; then
  echo "No migration files found."
  exit 1
fi

echo "Migrations applied successfully!"
