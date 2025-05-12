#!/bin/bash

# Получаем параметры
DB_USER=$1
DB_PASSWORD=$2
DB_HOST=$3
DB_PORT=$4
DB_NAME=$5

ROOT_DIR=$7


#echo "psql credentionals> host:$DB_HOST port:$DB_PORT user:$DB_USER name:$DB_NAME"

echo "Applying migrations from db/migrations/..."

#echo "Current directory: $(pwd)"
#ls -l $ROOT_DIR/db/migrations/

# Применяем миграции и проверяем статус выполнения каждой
all_migrations_successful=true

for file in $ROOT_DIR/db/migrations/*.up.sql; do
    echo "Applying migration: $file"
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$file"
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