#!/bin/sh
set -e # Script crash on any error

# The environment variables are expected to be already set by Docker
# Just print which file we are using for debug purposes
echo "📄 Environment variables loaded:"
env | grep USERS_ || true

MAX_RETRIES=90
RETRY_COUNT=0
SLEEP_SECONDS=15

echo "⏳ Waiting for PostgreSQL at ${USERS_POSTGRES_DB_HOST}:${USERS_POSTGRES_DB_PORT_INNER}..."
until nc -z "$USERS_POSTGRES_DB_HOST" "$USERS_POSTGRES_DB_PORT_INNER"; do
  RETRY_COUNT=$((RETRY_COUNT + 1))
  echo "⏳ Attempt ${RETRY_COUNT}/${MAX_RETRIES} — PostgreSQL not ready yet"
  if [ "$RETRY_COUNT" -ge "$MAX_RETRIES" ]; then
    echo "❌ Failed to connect to PostgreSQL after ${MAX_RETRIES} attempts. Exiting."
    exit 1
  fi
  sleep "$SLEEP_SECONDS"
done
echo "✅ PostgreSQL is available!"

# Determine SSL mode based on environment
if [ "$PROJECT_SUFFIX" = "dev" ]; then
  SSLMODE="disable"
else
  SSLMODE="require"
fi
echo "🔐 Using SSL mode: $SSLMODE"

# Connection check
echo "🔐 Verifying connection to PostgreSQL..."
PSQL_DSN="postgresql://$USERS_POSTGRES_DB_USER:$USERS_POSTGRES_DB_PASS@$USERS_POSTGRES_DB_HOST:$USERS_POSTGRES_DB_PORT_INNER/$USERS_POSTGRES_DB_NAME?sslmode=$SSLMODE"
PGPASSWORD=$USERS_POSTGRES_DB_PASS psql "$PSQL_DSN" -c "SELECT 1;" >/dev/null || {
  echo "❌ Unable to connect to PostgreSQL."
  exit 1
}

# Checking and creating a database
echo "🔍 Checking if database '${USERS_POSTGRES_DB_NAME}' exists..."
DB_EXISTS=$(PGPASSWORD=$USERS_POSTGRES_DB_PASS psql "postgresql://$USERS_POSTGRES_DB_USER:$USERS_POSTGRES_DB_PASS@$USERS_POSTGRES_DB_HOST:$USERS_POSTGRES_DB_PORT_INNER/postgres?sslmode=$SSLMODE" -tAc "SELECT 1 FROM pg_database WHERE datname='${USERS_POSTGRES_DB_NAME}';")
if [ "$DB_EXISTS" != "1" ]; then
  echo "🛠 Creating database '${USERS_POSTGRES_DB_NAME}'..."
  PGPASSWORD=$USERS_POSTGRES_DB_PASS psql "postgresql://$USERS_POSTGRES_DB_USER:$USERS_POSTGRES_DB_PASS@$USERS_POSTGRES_DB_HOST:$USERS_POSTGRES_DB_PORT_INNER/postgres?sslmode=$SSLMODE" -c "CREATE DATABASE ${USERS_POSTGRES_DB_NAME};"
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
  if [ -f "${USERS_SERVICE_ROOT}/cmd/seed/main.go" ]; then
    echo "🌱 Running seed script with go run for Docker..."
    go run "${USERS_SERVICE_ROOT}/cmd/seed/main.go"
  else
    echo "🌱 Running compiled seed binary (no Go source found)..."
    /app/bin/seed
  fi
fi

# Launching the application depending on the mode
if [ "$PROJECT_SUFFIX" = "dev" ]; then
  echo "🚀 Starting users-service with Air (development mode)..."
  cd /app/users-service
  exec air -c .air.toml
else
  echo "🚀 Starting users-service with compiled binary (production mode)..."
  exec /app/bin/main
fi
