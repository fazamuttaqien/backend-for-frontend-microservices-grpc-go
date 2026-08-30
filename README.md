# Backend for Frontend Microservices gRPC Go

Architecture:

```text
ReactJS -> BFF -> gRPC -> Microservices
```

## User Service authentication

The User Service provides Register and Login with bcrypt password hashing and short-lived HMAC-SHA256 JWT access tokens. Passwords are never returned or logged, and JWT secrets are loaded from environment configuration.

## Product Service caching

Product Service uses Redis only for cache-aside reads of individual products. PostgreSQL remains the source of truth.

- `GetProduct`: Redis read -> PostgreSQL on cache miss/error -> best-effort cache population.
- `UpdateProduct`: PostgreSQL write -> best-effort Redis invalidation.
- `ListProducts` is intentionally not cached.
- Cache entries use a configurable TTL and Redis operations have a short timeout.
- Redis failures fall back to PostgreSQL.

## OrderCreated asynchronous event

RabbitMQ is used as the message broker because this flow needs durable queues, acknowledgements, and simple producer/consumer semantics without replacing the existing synchronous gRPC calls.

```text
BFF
  |
  | synchronous gRPC
  v
Order Service
  |
  | persist order in PostgreSQL first
  |
  +--> RabbitMQ: order.events / order.created
                    |
                    v
          order.notification queue
                    |
                    v
          Notification Consumer
          (simulation/logging)
```

The Order Service continues to use synchronous gRPC to validate the user and retrieve product prices. Only the downstream notification work is asynchronous.

### Event contract

`OrderCreated` is published as JSON with:

- `type`: `OrderCreated`
- `version`: `v1`
- `event_id`: stable event identifier derived from the order ID
- `occurred_at`: event timestamp
- `data`: order ID, user ID, total, status, and creation time

The order is committed to PostgreSQL before the event is published. Publishing uses three attempts with a short backoff. If RabbitMQ is unavailable, order creation remains successful and the failure is logged; PostgreSQL remains the source of truth. A durable outbox can be introduced later if guaranteed event delivery becomes a requirement.

The notification consumer uses manual acknowledgements and retries processing up to three times before acknowledging the message and logging the failure. This keeps failure handling bounded without introducing an outbox or workflow framework for the MVP.

## Configuration

```bash
USER_DATABASE_URL=postgres://user:password@localhost:5432/user_service
PRODUCT_DATABASE_URL=postgres://user:password@localhost:5432/product_service
ORDER_DATABASE_URL=postgres://user:password@localhost:5432/order_service

REDIS_ADDRESS=localhost:6379
REDIS_TIMEOUT_MS=100
PRODUCT_CACHE_TTL_SECONDS=60

RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

## Database migration

Run the service migrations under `migrations/` against the corresponding PostgreSQL databases before starting the services.

## Build and test

```bash
make test
make build
```

## Docker Compose

The Docker Compose environment contains the BFF, User Service, Product Service, Order Service, Notification Consumer, PostgreSQL, Redis, and RabbitMQ.

```bash
cp .env.docker.example .env
docker compose up --build
```
