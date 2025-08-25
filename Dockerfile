FROM golang:1.24.0-alpine AS builder

WORKDIR /app/users-service

COPY go.mod go.sum ./
RUN go mod download

# Устанавливаем uuid ДО сборки проекта
RUN go get github.com/google/uuid
RUN go mod tidy

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/main ./main.go

# Installing migrate tool during build
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Устанавливаем air для горячей перезагрузки
RUN go install github.com/air-verse/air@latest

# Stage 2: Final image
FROM golang:1.24.0-alpine AS final

WORKDIR /app/users-service

# Копируем бинарник и необходимые файлы из билд-образа
COPY --from=builder /app/bin/main /app/bin/main
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /go/bin/air /usr/local/bin/air
COPY --from=builder /app/users-service/db /app/users-service/db

# Копируем скрипты entrypoint
COPY ./_docker /app/users-service/_docker

# Устанавливаем Go в финальный контейнер (для air)
RUN apk add --no-cache go

# Добавляем права на выполнение
RUN chmod +x /app/bin/main

# Добавляем PostgreSQL клиент в образ
RUN apk update && apk add postgresql-client
RUN apk add --no-cache git

# Устанавливаем curl для отладки внутри контейнера
RUN apk add --no-cache curl

# Устанавливаем переменную окружения для конфиг-файла
ENV CONFIG_PATH="/app/users-service/config/config.yaml"

EXPOSE ${USERS_SERVICE_PORT}

ENTRYPOINT ["/app/users-service/_docker/entrypoint.sh"]