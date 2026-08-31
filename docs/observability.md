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

The Go applications retain their OpenTelemetry SDK, HTTP/gRPC instrumentation, W3C trace propagation, structured logging, and graceful telemetry shutdown. Docker Compose sends application telemetry to Alloy through Docker DNS:

```text
OTEL_EXPORTER_OTLP_ENDPOINT=alloy:4317
OTEL_EXPORTER_OTLP_PROTOCOL=grpc
```

The current Go SDK implementation explicitly constructs OTLP/gRPC trace and metric exporters, so `grpc` is the effective protocol. The endpoint remains environment-driven and can be changed without modifying application code.

Each application has a stable `service.name` and `deployment.environment.name=docker` resource identity. Trace context is propagated with W3C TraceContext and Baggage; request IDs are propagated separately through HTTP/gRPC metadata.

## Telemetry reliability

Application telemetry is asynchronous where possible: traces use the OpenTelemetry batch span processor and metrics use a periodic metric reader. Exporter initialization is bounded by a short timeout and exporter creation failures disable the affected telemetry rather than terminating the application. Shutdown is explicit and uses the application's shutdown context.

Exporter failure after startup is handled by the OpenTelemetry exporter/SDK; telemetry is best-effort and must never be treated as an application dependency.

Alloy adds a second bounded protection layer:

- memory limiter: 80% limit with a 20% spike allowance
- batch size: 512
- batch timeout: 5s
- OTLP/HTTP timeout: 10s
- retry: 1s initial, 30s maximum interval, 5m maximum elapsed time
- sending queue: 1000 batches with 4 consumers
- persistent Alloy storage at `/var/lib/alloy/data`

These limits protect the process from unbounded telemetry buffering. During a prolonged Grafana Cloud outage or sustained overload, telemetry can eventually be dropped after bounded buffers are exhausted. This is intentional: losing telemetry is preferable to exhausting application memory or blocking business traffic indefinitely.

## Grafana Alloy

Alloy receives OTLP on the Docker-internal `observability` network:

- gRPC: `alloy:4317`
- HTTP: `alloy:4318`

Neither receiver is published to the host. The `observability` network is internal to Docker. Alloy has a separate `egress` network for outbound Grafana Cloud traffic.

Alloy runs with a read-only root filesystem, all Linux capabilities dropped, and `no-new-privileges`. Its persistent data volume is writable because Alloy requires storage for its local state/queue.

Alloy also reads the four backend containers' Docker stdout logs through the read-only Docker socket. The Docker socket is a privileged host interface even when mounted read-only, so this is the main remaining host-level security trade-off. If stronger isolation is required later, logs can be shipped through a dedicated Docker logging driver or external log collector instead of mounting the socket.

## Logs

Go services continue to emit structured JSON logs. Alloy collects Docker logs, parses the JSON payload, and sends them to Grafana Cloud Logs. `trace_id` and `span_id` are retained as structured metadata rather than Loki labels. High-cardinality identifiers such as trace IDs, span IDs, request IDs, and user IDs must not be labels.

Application logs must not contain JWTs, passwords, API keys, database credentials, Authorization headers, request bodies, or other secrets. The current observability middleware logs only request metadata, status, duration, request ID, trace ID, and span ID.

## Traces

Traces originate in the Go HTTP/gRPC instrumentation and are exported through OTLP/gRPC to Alloy. Resource attributes provide `service.name` and the Docker deployment environment. W3C trace context keeps a single distributed trace connected across the BFF and downstream gRPC services.

## Metrics

Metrics originate from the OpenTelemetry Go SDK and HTTP/gRPC instrumentation. They are exported through OTLP/gRPC to Alloy and then OTLP/HTTP to Grafana Cloud. Resource attributes provide service identity and environment. No high-cardinality custom dimensions are introduced by the observability pipeline.

## Grafana Cloud credentials

Only Alloy receives:

```text
GRAFANA_CLOUD_OTLP_ENDPOINT
GRAFANA_CLOUD_LOKI_ENDPOINT
GRAFANA_CLOUD_INSTANCE_ID
GRAFANA_CLOUD_API_KEY
```

Application containers do not receive Grafana Cloud credentials. Real credentials must be supplied through the local environment and must never be committed or baked into an image.

Grafana Cloud uses the OTLP HTTP exporter with Basic authentication for traces, metrics, and OTLP logs. Grafana Cloud's OTLP connection details must come from the Grafana Cloud OpenTelemetry/OTLP details page. Do not infer or hardcode tenant-specific endpoints.

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

For the Go code, the repository CI runs:

```bash
go test ./...
go test -race ./...
go vet ./...
go build ./...
```

Runtime Docker and Grafana Cloud validation requires a Docker daemon and valid Grafana Cloud credentials; those are intentionally not committed to the repository.
