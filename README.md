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
