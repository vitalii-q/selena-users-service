FROM golang:1.24.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/users-service/main ./main.go

RUN mkdir -p /config
COPY users-service/config/config.yaml /config/config.yaml

# Downlouding variables from .env
RUN go mod tidy
#RUN go build -o main

# Installing migrate tool during build
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Устанавливаем air для горячей перезагрузки
RUN go install github.com/air-verse/air@latest

# Stage 2: Final image
FROM golang:1.24.0-alpine AS final

WORKDIR /app/users-service

COPY --from=builder /app/users-service /app/users-service
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /go/bin/air /usr/local/bin/air

# Устанавливаем Go в финальный контейнер (для air)
RUN apk add --no-cache go

# Проверяем, есть ли бинарник main
RUN ls -la /app/users-service

# Добавляем права на выполнение
RUN chmod +x /app/users-service/main

# Добавляем PostgreSQL клиент в образ
RUN apk update && apk add postgresql-client
RUN apk add --no-cache git

EXPOSE ${USER_SERVICE_PORT}

CMD ["/app/users-service/main"]