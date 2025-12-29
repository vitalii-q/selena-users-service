# users-service/Dockerfile

# --- Start microservice
# docker build --no-cache --platform linux/amd64 -t selena-users-service:amd64 .
#
# docker run -d --name users-service --env-file .env -p 9065:9065 --network selena-dev_app_network -v $(pwd):/app/users-service selena-users-service:amd64
# -v $(pwd):/app/users-service — mount the local sources into the container

# --- Start DB for microservice
# Launch command in the root directory /users-service
# docker run -d --name users-db -p 9265:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=users_db -v $(pwd)/_docker/users-db-data:/var/lib/postgresql/data --network selena-dev_app_network postgres:15

# The sequence of launching microservices: hotels-service -> users-service -> bookings-service

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

# Stage 2: Final image
FROM golang:1.24.0-alpine AS final

WORKDIR /app/users-service

# Install all necessary packages in one layer
RUN apk update && apk add --no-cache curl git postgresql-client go

# Install AIR hot reload (prebuilt binary)
RUN curl -L https://github.com/air-verse/air/releases/download/v1.62.0/air_1.62.0_linux_amd64.tar.gz \
    | tar -xz \
    && mv air /usr/local/bin/air \
    && chmod +x /usr/local/bin/air

# Copy the binary and necessary files from the build image
COPY --from=builder /app/bin/main /app/bin/main
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/users-service/db /app/users-service/db
# Copy the seed binary
COPY --from=builder /app/bin/seed /app/bin/seed

# Copy the entrypoint scripts
COPY ./_docker /app/users-service/_docker

# Add execution rights
RUN chmod +x /app/bin/main

RUN chmod +x /app/bin/main /app/bin/seed

# Set the environment variable for the config file
ENV CONFIG_PATH="/app/users-service/config/config.yaml"

EXPOSE ${USERS_SERVICE_PORT}

ENTRYPOINT ["/app/users-service/_docker/entrypoint.sh"]