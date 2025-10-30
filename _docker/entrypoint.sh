#!/bin/sh
set -e # Script crash on any error

# Defining the environment: cloud or local
if [ -n "$AWS_EXECUTION_ENV" ]; then
  echo "☁️ Running in cloud environment (AWS)"
  ENV_FILE="/app/users-service/.env.cloud"
else
  echo "🏠 Running locally"
  ENV_FILE="/app/users-service/.env"
fi

# Load environment variables from .env file if it exists
if [ -f /app/users-service/.env ]; then
  echo "📄 Loading environment variables from .env file..."
  set -a  # automatically export all variables
  . /app/users-service/.env
  set +a
fi

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

# Connection check
echo "🔐 Verifying connection to PostgreSQL..."
PGPASSWORD=$USERS_POSTGRES_DB_PASS psql -h "$USERS_POSTGRES_DB_HOST" -U "$USERS_POSTGRES_DB_USER" -p "$USERS_POSTGRES_DB_PORT_INNER" -d postgres -c "SELECT 1;" >/dev/null
if [ $? -ne 0 ]; then
  echo "❌ Unable to connect to PostgreSQL."
  exit 1
fi

# Checking and creating a database
echo "🔍 Checking if database '${USERS_POSTGRES_DB_NAME}' exists..."
DB_EXISTS=$(PGPASSWORD=$USERS_POSTGRES_DB_PASS psql -h "$USERS_POSTGRES_DB_HOST" -U "$USERS_POSTGRES_DB_USER" -p "$USERS_POSTGRES_DB_PORT_INNER" -tAc "SELECT 1 FROM pg_database WHERE datname='${USERS_POSTGRES_DB_NAME}';")
if [ "$DB_EXISTS" != "1" ]; then
  echo "🛠 Creating database '${USERS_POSTGRES_DB_NAME}'..."
  PGPASSWORD=$USERS_POSTGRES_DB_PASS psql -h "$USERS_POSTGRES_DB_HOST" -U "$USERS_POSTGRES_DB_USER" -p "$USERS_POSTGRES_DB_PORT_INNER" -d postgres -c "CREATE DATABASE ${USERS_POSTGRES_DB_NAME};"
  echo "✅ Database created."
else
  echo "📦 Database '${USERS_POSTGRES_DB_NAME}' already exists."
fi

# The path to the root of microservices
USERS_SERVICE_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
echo "📁 USERS_SERVICE_ROOT=${USERS_SERVICE_ROOT}"

# Performing migrations
sh "${USERS_SERVICE_ROOT}/db/migrate.sh"

# Database seeding
if [ "$RUN_MODE" = "k8s" ]; then
  echo "🌱 Running seed binary for Kubernetes..."
  /app/bin/seed
else
  echo "🌱 Running seed script with go run for Docker..."
  go run "${USERS_SERVICE_ROOT}/cmd/seed/main.go"
fi

# Launching the application depending on the mode
if [ "$PROJECT_SUFFIX" = "dev" ]; then
  echo "🚀 Starting users-service with Air (development mode)..."
  exec air -c .air.toml
else
  echo "🚀 Starting users-service with compiled binary (production mode)..."
  exec /app/bin/main
fi
