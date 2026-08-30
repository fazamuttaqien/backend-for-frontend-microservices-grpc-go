# Backend for Frontend Microservices gRPC Go

Architecture:

```text
ReactJS -> BFF -> gRPC -> Microservices
```

## User Service authentication

The User Service now provides Register and Login with bcrypt password hashing and short-lived HMAC-SHA256 JWT access tokens. Passwords are never returned or logged, and JWT secrets are loaded from environment configuration.

Required configuration:

```bash
USER_DATABASE_URL=postgres://user:password@localhost:5432/user_service
```

Optional:

```bash
USER_GRPC_PORT=50051
JWT_ISSUER=user-service
JWT_TTL_MINUTES=60
```

`Register` and `Login` are public RPCs. `GetUser` and `UpdateUser` require `authorization: Bearer <JWT>` metadata. The same JWT can be forwarded by the BFF to downstream protected gRPC calls.

## Product Service caching

Product Service uses Redis only for cache-aside reads of individual products. PostgreSQL remains the source of truth.

- `GetProduct`: Redis read -> PostgreSQL on cache miss/error -> best-effort cache population.
- `UpdateProduct`: PostgreSQL write -> best-effort Redis invalidation.
- `ListProducts` is intentionally not cached to avoid unnecessary stale pagination data.
- Cache entries use a configurable TTL and Redis operations are bounded by a short timeout.
- Redis failures do not make Product Service unavailable; reads fall back to PostgreSQL.

Configuration:

```bash
REDIS_ADDRESS=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_TIMEOUT_MS=100
PRODUCT_CACHE_TTL_SECONDS=60
```

## Database migration

Run `migrations/user/001_create_users.sql` against the User Service PostgreSQL database before starting the service.

## Run User Service

```bash
make run-user
```

## Build and test

```bash
make test
make build
```

`make test` and `make build` regenerate protobuf code before compiling.

## Docker Compose

For a complete local environment containing the BFF, User Service, Product Service, Order Service, PostgreSQL, and Redis, see [`README.docker.md`](README.docker.md).

Quick start:

```bash
cp .env.docker.example .env
docker compose up --build
```

The Docker Compose setup uses Docker service names for gRPC communication, initializes the three PostgreSQL service databases, runs Redis for Product Service caching, and runs application containers as a non-root user.
