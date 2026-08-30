# Backend for Frontend Microservices gRPC Go

Backend foundation for:

```text
ReactJS
  ↓
BFF
  ↓
gRPC
  ↓
Microservices
```

## Current stage

This stage establishes the Protocol Buffers and gRPC contract foundation. It does not implement the BFF, User Service, database, or business logic.

## Structure

```text
.
├── apps/
│   ├── bff/                 # BFF entrypoint (future implementation)
│   └── services/            # Microservice entrypoints (future)
├── gen/                     # Generated Go protobuf/gRPC code
│   └── user/v1/
├── internal/
│   └── config/              # Shared application configuration
├── proto/
│   ├── user/v1/             # User API version 1
│   ├── product/v1/           # Reserved for Product API version 1
│   └── order/v1/             # Reserved for Order API version 1
├── buf.yaml                 # Protobuf module/lint/breaking-change config
├── buf.gen.yaml              # Protobuf and gRPC generation config
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Protobuf versioning

API contracts are versioned by directory and protobuf package:

```text
proto/user/v1/
proto/product/v1/
proto/order/v1/
```

A breaking API evolution should introduce a new version such as `user/v2` rather than changing the existing `user/v1` contract incompatibly.

## User Service contract

The first example contract is `proto/user/v1/user.proto`. It defines only the API shape:

- `GetUserRequest`
- `User`
- `GetUserResponse`
- `UserService.GetUser`

There is intentionally no User Service implementation.

## Generate protobuf and gRPC code

Requirements:

- Go 1.24+
- Buf CLI

Run:

```bash
make proto-generate
```

This runs `buf generate` and writes generated Go files under `gen/`.

Validate protobuf definitions:

```bash
make proto-lint
```

Check breaking changes against `main`:

```bash
make proto-breaking
```

## Build and test

```bash
make build
make test
```

The generated contract package is compiled as part of `go build ./...`.

## Go module strategy

The repository continues to use one Go module at the root. Protobuf contracts and generated Go code live in the same module so BFF and future microservices can consume the same versioned contracts without premature multi-module complexity.
