#!/bin/sh
set -e # Падение скрипта при любой ошибке

MAX_RETRIES=10
RETRY_COUNT=0

echo "⏳ Waiting for PostgreSQL at ${USERS_POSTGRES_DB_HOST}:${USERS_POSTGRES_DB_PORT_INNER}..."
until nc -z "$USERS_POSTGRES_DB_HOST" "$USERS_POSTGRES_DB_PORT_INNER"; do
  RETRY_COUNT=$((RETRY_COUNT + 1))
  echo "✅ Attempt $RETRY_COUNT"
  if [ "$RETRY_COUNT" -ge "$MAX_RETRIES" ]; then
    echo "❌ Failed to connect to PostgreSQL after ${MAX_RETRIES} attempts. Exiting."
    exit 1
  fi
  sleep 1
done
echo "✅ PostgreSQL is available!"

# Проверка соединения
echo "🔐 Verifying connection to PostgreSQL..."
PGPASSWORD=$USERS_POSTGRES_DB_PASS psql -h "$USERS_POSTGRES_DB_HOST" -U "$USERS_POSTGRES_DB_USER" -p "$USERS_POSTGRES_DB_PORT_INNER" -d postgres -c "SELECT 1;" >/dev/null
if [ $? -ne 0 ]; then
  echo "❌ Unable to connect to PostgreSQL."
  exit 1
fi

# Проверка и создание базы данных
echo "🔍 Checking if database '${USERS_POSTGRES_DB_NAME}' exists..."
DB_EXISTS=$(PGPASSWORD=$USERS_POSTGRES_DB_PASS psql -h "$USERS_POSTGRES_DB_HOST" -U "$USERS_POSTGRES_DB_USER" -p "$USERS_POSTGRES_DB_PORT_INNER" -tAc "SELECT 1 FROM pg_database WHERE datname='${USERS_POSTGRES_DB_NAME}';")
if [ "$DB_EXISTS" != "1" ]; then
  echo "🛠 Creating database '${USERS_POSTGRES_DB_NAME}'..."
  PGPASSWORD=$USERS_POSTGRES_DB_PASS psql -h "$USERS_POSTGRES_DB_HOST" -U "$USERS_POSTGRES_DB_USER" -p "$USERS_POSTGRES_DB_PORT_INNER" -d postgres -c "CREATE DATABASE ${USERS_POSTGRES_DB_NAME};"
  echo "✅ Database created."
else
  echo "📦 Database '${USERS_POSTGRES_DB_NAME}' already exists."
fi

# Путь к корню микросервиса
USERS_SERVICE_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
echo "📁 USERS_SERVICE_ROOT=${USERS_SERVICE_ROOT}"

# Выполняем миграции
sh "${USERS_SERVICE_ROOT}/db/migrate.sh"

# --- New: запускаем сиды после миграций ---
echo "🌱 Seeding database..."
go run "${USERS_SERVICE_ROOT}/cmd/seed/main.go"
echo "✅ Seeding finished!"

# Запуск приложения в зависимости от режима
if [ "$PROJECT_SUFFIX" = "dev" ]; then
  echo "🚀 Starting users-service with Air (development mode)..."
  exec air -c .air.toml
else
  echo "🚀 Starting users-service with compiled binary (production mode)..."
  exec /app/bin/main
fi
