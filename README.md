# Backend for Frontend Microservices gRPC Go

Backend foundation for a system with the following architecture:

```text
ReactJS
  ↓
BFF
  ↓
gRPC
  ↓
Microservices
```

## Scope

This stage only establishes the project foundation. No business logic or concrete User, Product, or Order service is included.

## Structure

```text
.
├── apps/
│   ├── bff/              # BFF application entrypoint
│   └── services/         # Microservice application entrypoints (future)
├── gen/                  # Generated protobuf Go code (future)
├── internal/
│   └── config/           # Shared application configuration
├── proto/                # Protobuf definitions (future)
├── go.mod
├── Makefile
└── README.md
```

## Go module strategy

The repository currently uses one Go module at the repository root. This keeps the foundation simple and allows BFF, generated protobuf code, and future microservices to evolve together without premature multi-module complexity.

A multi-module workspace can be introduced later if independent module versioning or release boundaries become necessary.

## Requirements

- Go 1.24+

## Run

Build everything:

```bash
make build
```

Run the BFF foundation:

```bash
make run
```

Run tests:

```bash
make test
```

The BFF currently only provides a startup placeholder; transport, gRPC clients, protobuf contracts, and business services will be added in later stages.
