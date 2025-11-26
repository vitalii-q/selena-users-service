FROM golang:1.24.0-alpine AS builder

WORKDIR /app/users-service

COPY go.mod go.sum ./
RUN go mod download

# Set the uuid BEFORE building the project
RUN go get github.com/google/uuid
RUN go mod tidy

COPY . ./

# Build main binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/main ./main.go
# Build seed binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/seed ./cmd/seed/main.go

# Installing migrate tool during build
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Installing git for building air
RUN apk add --no-cache git

# Installing Air for a hot reboot
RUN go install github.com/air-verse/air@v1.62.0

# Stage 2: Final image
FROM golang:1.24.0-alpine AS final

WORKDIR /app/users-service

# Copy the binary and necessary files from the build image
COPY --from=builder /app/bin/main /app/bin/main
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /go/bin/air /usr/local/bin/air
COPY --from=builder /app/users-service/db /app/users-service/db
# Copy the seed binary
COPY --from=builder /app/bin/seed /app/bin/seed

# Copy the entrypoint scripts
COPY ./_docker /app/users-service/_docker

# Copy .env file
COPY .env /app/users-service/.env

# Install Go in the final container (for air)
RUN apk add --no-cache go

# Add execution rights
RUN chmod +x /app/bin/main

# Add PostgreSQL client to image
RUN apk update && apk add postgresql-client
RUN apk add --no-cache git

# Install curl for debugging inside the container
RUN apk add --no-cache curl

RUN chmod +x /app/bin/main /app/bin/seed

# Set the environment variable for the config file
ENV CONFIG_PATH="/app/users-service/config/config.yaml"

EXPOSE ${USERS_SERVICE_PORT}

ENTRYPOINT ["/app/users-service/_docker/entrypoint.sh"]