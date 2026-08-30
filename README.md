# Backend for Frontend Microservices gRPC Go

Architecture:

```text
ReactJS -> BFF -> gRPC -> Microservices
```

## User Service authentication

The User Service provides Register and Login with bcrypt password hashing and short-lived HMAC-SHA256 JWT access tokens. Passwords are never returned or logged, and JWT secrets are loaded from environment configuration.

## Product Service caching

Product Service uses Redis only for cache-aside reads of individual products. PostgreSQL remains the source of truth.

## OrderCreated asynchronous event

RabbitMQ is used for the durable notification flow. RabbitMQ credentials must be supplied through environment configuration; there is no default credential fallback.

## Build and test

```bash
make test
make build
```

CI also runs the race detector, `go vet`, and `govulncheck`.

## Docker Compose

```bash
cp .env.docker.example .env
# Replace all example credentials/secrets before use.
docker compose up --build
```

See `docs/final-backend-review.md` for the final security/resilience audit and `docs/observability.md` for tracing and metrics.
