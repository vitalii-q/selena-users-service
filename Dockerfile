FROM golang:1.24.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN go build -o user-service ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/user-service .

EXPOSE 8080

CMD ["./user-service"]