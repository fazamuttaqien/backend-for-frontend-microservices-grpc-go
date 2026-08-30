# Backend for Frontend Microservices gRPC Go

Architecture:

```text
ReactJS -> BFF -> gRPC -> Microservices
```

## Current stage

The User Service MVP is implemented with Go, gRPC, PostgreSQL, and a simple Clean Architecture. Product and Order services are intentionally not implemented.

## User Service structure

```text
apps/services/user/
└── main.go

internal/user/
├── domain/
├── application/
├── repository/
├── infrastructure/postgres/
└── transport/grpc/

migrations/user/
└── 001_create_users.sql
```

The User Service owns its `users` table and only accesses its own PostgreSQL database.

## Configuration

Required:

```bash
DATABASE_URL=postgres://user:password@localhost:5432/user_service
```

Optional:

```bash
USER_GRPC_PORT=50051
```

## Database migration

Run `migrations/user/001_create_users.sql` against the User Service PostgreSQL database before starting the service. No migration framework is required at this stage.

## Run User Service

```bash
make run-user
```

The service listens on `:50051` by default.

## Build and test

```bash
make test
make build
```

## Protobuf

Contracts are versioned under `proto/user/v1`, `proto/product/v1`, and `proto/order/v1`. Generated Go code is under `gen/`.

Regenerate contracts with:

```bash
make proto-generate
```

## Security

Passwords are hashed with bcrypt and are never returned by the gRPC response. Authentication/authorization is intentionally outside this MVP.
