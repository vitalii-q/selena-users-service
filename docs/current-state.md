# Current State: Users Service

## Local Runtime

- Technology: Go, Gin, pgx, and golang-migrate.
- Development HTTP port: `9065`.
- Development database: PostgreSQL, exposed locally on `9265` and reachable in
  Docker as `users-db:5432`.
- Local Redis is configured on port `9765`.
- Build command:

  ```bash
  docker build --no-cache --platform linux/amd64 -t selena-users-service:amd64 .
  ```

- The development container expects the shared Docker network
  `selena-dev_app_network` and a local `.env` file.
- The entrypoint waits for PostgreSQL, runs migrations, seeds data, and starts
  the service with Air in development mode.

## Health and Observability

- Liveness endpoint: `/health`.
- Database readiness endpoint: `/ready`.
- Prometheus endpoint: `/metrics`.
- Request ID and Prometheus middleware are configured.

## Database and Migrations

- The service owns a PostgreSQL database.
- SQL migrations are stored in `db/migrations`.
- Existing schema covers users, user profiles, indexes, roles, soft deletion,
  OAuth sessions, and authorization codes.

## Existing HTTP API

- `POST /api/v1/users`
- `GET /api/v1/users`
- `GET /api/v1/users/{id}`
- `PUT /api/v1/users/{id}`
- `DELETE /api/v1/users/{id}`
- `GET /api/v1/locations`
- `GET /users/{id}/hotels`
- OAuth-related routes also exist but are outside the dev MVP scope.

## Tests

- `internal/handlers/user_handler_test.go`.
- `internal/services/user_service_test.go`.
- Integration tests under `tests/integration`.
- Standard command: `go test ./...`.

## Known MVP Gaps

- The public list endpoint does not yet provide the planned standard pagination
  and sorting contract.
- Error response formats and HTTP status handling need standardization.
- Password/password-hash exposure must be audited across all responses.
- A minimal internal booking-eligibility endpoint is not implemented.
- OAuth and authentication paths are present but are not part of the dev MVP.
- Hotels enrichment depends on a hotels-service route that currently differs
  from the versioned hotel API.
- Kubernetes manifests exist under the root-level `k8s/users-service` path,
  not in this repository; they still need validation against the local MVP.

## Working Tree

There was an existing uncommitted modification when this document was created:

```text
M internal/services/auth_service.go
```

Future tasks must preserve or explicitly resolve this change before modifying
the same file.
