FROM golang:1.24.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/users-service ./main.go

RUN mkdir -p /config
COPY config/ /config/
RUN ls -la /config

# Downlouding variables from .env
RUN go mod tidy
RUN go build -o main

# Installing migrate tool during build
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Final stage
FROM alpine:latest

WORKDIR /app/users-service

COPY --from=builder /app /app/users-service
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

RUN chmod +x /app/users-service

# Добавляем PostgreSQL клиент в образ
RUN apk update && apk add postgresql-client
RUN apk add --no-cache git

EXPOSE ${USER_SERVICE_PORT}

CMD ["./users-service"]
