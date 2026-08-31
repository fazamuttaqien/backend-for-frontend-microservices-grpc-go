# Observability

The backend uses OpenTelemetry for traces and metrics and structured JSON logs on stdout. Grafana Alloy is the only telemetry gateway in Docker Compose; the local OpenTelemetry Collector, Jaeger, and Prometheus have been removed.

## Architecture

```text
Go services
(BFF / User / Product / Order)
        |
        | OTLP/gRPC
        v
Grafana Alloy
   |              \
   | OTLP/HTTP      \ Loki HTTP
   v                 \
Grafana Cloud         Grafana Cloud
Traces + Metrics       Logs
```

## Applications

The Go applications retain their OpenTelemetry SDK, HTTP/gRPC instrumentation, W3C trace propagation, and graceful telemetry shutdown. Docker Compose sends application telemetry to Alloy through Docker DNS:

```text
OTEL_EXPORTER_OTLP_ENDPOINT=alloy:4317
OTEL_EXPORTER_OTLP_PROTOCOL=grpc
```

Each application has a stable `service.name` and `deployment.environment.name=docker` resource identity.

## Grafana Alloy

Alloy receives OTLP on the Docker-internal `observability` network:

- gRPC: `alloy:4317`
- HTTP: `alloy:4318`

The receiver is not published to the host. Alloy applies memory protection and batching before exporting OTLP telemetry to Grafana Cloud with timeout, retry, and persistent sending queue settings.

Alloy also reads the four backend containers' Docker stdout logs through the read-only Docker socket. Existing application logging is not rewritten.

## Logs

Go services continue to emit structured JSON logs. Alloy collects Docker logs, parses the JSON payload, and sends them to Grafana Cloud Logs. `trace_id` and `span_id` are retained as structured metadata rather than Loki labels. High-cardinality identifiers such as trace IDs, span IDs, request IDs, and user IDs must not be labels.

Application logs must not contain JWTs, passwords, API keys, database credentials, Authorization headers, request bodies, or other secrets.

## Grafana Cloud credentials

Only Alloy receives:

```text
GRAFANA_CLOUD_OTLP_ENDPOINT
GRAFANA_CLOUD_LOKI_ENDPOINT
GRAFANA_CLOUD_INSTANCE_ID
GRAFANA_CLOUD_API_KEY
```

Application containers do not receive Grafana Cloud credentials. Real credentials must be supplied through the local environment and must never be committed or baked into an image.

## Removed local stack

The following obsolete components/configuration have been removed because Grafana Cloud is now the telemetry backend:

- OpenTelemetry Collector service
- Jaeger service
- Prometheus service
- `docker/observability/otel-collector.yaml`
- `docker/observability/prometheus.yml`
- legacy local collector/Jaeger/Prometheus ports and environment variables

The OpenTelemetry Go SDK and application instrumentation are intentionally retained.

## Makefile

No observability-specific Makefile target was present, so no Makefile change is required. Existing build, test, formatting, and protobuf targets remain unchanged.

## Validation

Run:

```bash
docker compose config
docker compose up -d --build
docker compose ps
docker compose logs alloy
```

Then generate traffic through the BFF and verify in Grafana Cloud that traces, metrics, and logs use the expected service identity and trace context. Also verify that no Grafana Cloud credential is present in application container environments and no JWT, password, Authorization header, or other secret appears in logs.
