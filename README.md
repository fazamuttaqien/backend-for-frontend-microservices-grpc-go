# Local Development with Docker Compose

This repository includes a Docker Compose environment for the complete backend:

- BFF (HTTP :8080)
- User Service (gRPC :50051, internal network)
- Product Service (gRPC :50052, internal network)
- Order Service (gRPC :50053, internal network)
- PostgreSQL (host :5432)

The Go services use a multi-stage Docker build. Runtime containers run as a non-root `app` user. No application secret is copied into the image.

## Prerequisites

- Docker Engine with Docker Compose v2
- A free local port 8080 and, if you want host access to PostgreSQL, 5432

## Start

1. Copy the Docker environment template:

   ```bash
   cp .env.docker.example .env
   ```

2. Set a local JWT secret in `.env` (at least 32 random bytes). Do not commit `.env`.

3. Start the stack:

   ```bash
   docker compose up --build
   ```

4. The BFF is available at:

   ```text
   http://localhost:8080
   ```

   PostgreSQL is available from the host at `localhost:5432`.

The Compose network uses Docker service names for gRPC communication (`user-service`, `product-service`, and `order-service`). Do not use `localhost` for service-to-service addresses inside containers.

## Verify

Check container status:

```bash
docker compose ps
```

The PostgreSQL and application containers have health checks. Order Service waits for User and Product Service to become healthy, and the BFF waits for all three services.

View logs for one service:

```bash
docker compose logs -f bff
docker compose logs -f order-service
```

## Stop

```bash
docker compose down
```

To remove the PostgreSQL development data as well:

```bash
docker compose down -v
```

The first PostgreSQL startup creates the three service databases and the tables required by the current services. The initialization script runs only when the PostgreSQL volume is created for the first time. If the schema changes during development, remove the volume and start the stack again.

## Configuration

`compose.yaml` passes configuration through environment variables. Database credentials and JWT secrets are supplied at runtime; they are not baked into Docker images.

For a local Docker setup, the important values are:

```text
USER_DATABASE_URL=postgres://...@postgres:5432/user_service
PRODUCT_DATABASE_URL=postgres://...@postgres:5432/product_service
ORDER_DATABASE_URL=postgres://...@postgres:5432/order_service
USER_SERVICE_ADDRESS=user-service:50051
PRODUCT_SERVICE_ADDRESS=product-service:50052
ORDER_SERVICE_ADDRESS=order-service:50053
```

Do not copy the `localhost` service addresses from the normal `.env.example` into container configuration: `localhost` inside a container refers to that same container.
