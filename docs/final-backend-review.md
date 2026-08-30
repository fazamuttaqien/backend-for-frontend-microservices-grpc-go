# Final Backend Review

## Security fixes

- Removed hardcoded RabbitMQ `guest:guest` fallback; RabbitMQ credentials are now required through configuration.
- Added JWT authentication at gRPC service boundaries for protected User, Product mutation, and Order RPCs.
- Added object-level authorization: users can only read/update themselves; order access is scoped to the authenticated user.
- Added password length validation compatible with bcrypt (8-72 bytes).
- Added request size limits to HTTP and gRPC paths.
- Added per-client BFF rate limiting with a bounded in-memory table.
- Kept request bodies, authorization headers, JWTs, passwords, and secrets out of structured logs.

## Resilience

- Added bounded timeouts for startup, HTTP requests, database connections, and downstream gRPC calls.
- Added retry only for idempotent read-only gRPC calls on `Unavailable`; order creation and other mutations are not retried.
- Added bounded graceful shutdown with a hard fallback to `Stop` after 10 seconds.
- Added gRPC health checking and container readiness probes.
- Tuned PostgreSQL pool limits and connection lifetimes.

## Dependency review

OpenTelemetry 1.36 was upgraded to 1.41 because versions 1.36.0-1.40.0 are affected by a reviewed OpenTelemetry SDK vulnerability (GO-2026-4394). CI also runs `govulncheck` so future dependency vulnerabilities fail the build.

## SQL / resources

PostgreSQL access uses parameterized queries. Query rows are closed, transactions roll back on failure, and database pools are explicitly closed during shutdown.

## Docker

Application containers already run as a non-root user. They now also use a read-only filesystem, drop all Linux capabilities, and enable `no-new-privileges`.

## Known limitation

The MVP has no role/permission model. Product write RPCs therefore require authentication but cannot distinguish an administrator from a normal authenticated user. A future RBAC feature would be required before exposing product administration to untrusted clients.
