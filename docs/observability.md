# Local Observability

The backend now emits structured logs plus OpenTelemetry traces and metrics.

## Stack

- OpenTelemetry Collector: receives OTLP from the BFF and services.
- Jaeger: trace UI.
- Prometheus: metrics UI.

The application services use OTLP/gRPC and default to `localhost:4317`. Docker Compose overrides this to `otel-collector:4317`.

## Run locally

1. Copy the environment example and set the existing application secrets as usual.
2. Start the full stack:

```bash
docker compose up --build
```

3. Send a request through the BFF. For example:

```bash
curl -i http://localhost:8080/api/v1/products
```

The response contains `X-Request-ID` and `X-Trace-ID` headers.

4. Open the local UIs:

- Jaeger: http://localhost:16686
- Prometheus: http://localhost:9090
- Collector Prometheus endpoint: http://localhost:8889/metrics

## Distributed trace flow

The BFF accepts W3C `traceparent`/`tracestate` headers. OpenTelemetry HTTP instrumentation creates the BFF server span and the gRPC instrumentation propagates the same trace context to downstream services:

```text
ReactJS/browser
    |
    | W3C traceparent
    v
BFF HTTP span
    |
    | gRPC trace context
    v
Order Service
    |             \\
    | gRPC          \\ gRPC
    v               v
User Service    Product Service
```

A React application should initialize OpenTelemetry Web tracing and send W3C trace context on requests to the BFF. The backend does not require the React application to expose tokens or user data to tracing.

## Environment variables

`OTEL_EXPORTER_OTLP_ENDPOINT` controls the OTLP/gRPC collector endpoint. Example for Docker Compose:

```text
OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector:4317
```

Telemetry export is non-fatal: if the collector is unavailable during startup, the application continues with no-op exporters.

## Logging safety

Structured logs contain operational metadata such as service, method, status, request ID, trace ID, span ID, and duration. They do not log request bodies, authorization headers, JWTs, passwords, secrets, or sensitive user fields.
