#!/bin/bash

# Получаем параметры
DB_USER=$1
DB_PASSWORD=$2
DB_HOST=$3
DB_PORT=$4
DB_NAME=$5

echo "testtesttest"

# Пример применения миграций с использованием goose
goose -dir $6 postgres "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable" up