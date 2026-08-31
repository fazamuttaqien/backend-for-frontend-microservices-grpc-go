# Observability

The backend emits structured logs plus OpenTelemetry traces and metrics. Docker Compose uses Grafana Alloy as the telemetry gateway.

## Signal architecture

```text
BFF / User / Product / Order
        |
        | OTLP/gRPC
        v
Grafana Alloy
   |            \
   |             \
OTLP/HTTP       Loki HTTP
   |             |
   v             v
Grafana Cloud Metrics/Traces   Grafana Cloud Logs
```

## Traces

The Go applications use OpenTelemetry HTTP and gRPC instrumentation. W3C trace context is propagated through the BFF and downstream gRPC calls. The response also exposes request/trace identifiers for operational debugging.

Docker Compose sets a stable OpenTelemetry resource identity for each application:

```text
service.name=bff
service.name=user-service
service.name=product-service
service.name=order-service
```

and the shared environment metadata:

```text
deployment.environment.name=docker
```

This metadata is attached by the application OpenTelemetry resource configuration and is used consistently across metrics and traces.

## Metrics

Application metrics are exported through OTLP. Existing HTTP/gRPC instrumentation is preserved; no high-cardinality custom metrics are introduced. The stable service identity and deployment environment metadata allow Grafana Cloud metrics to be filtered alongside traces.

## Logs

The existing Go logging system writes structured JSON logs to stdout. It is intentionally not rewritten to an OpenTelemetry logging SDK for this migration.

Grafana Alloy reads Docker container logs through `loki.source.docker`, parses the existing Docker envelope and JSON payload, and sends them to Grafana Cloud Logs. `trace_id` and `span_id` are extracted as Loki structured metadata rather than labels. This preserves trace correlation without creating a Loki stream for every trace.

Only low-cardinality metadata such as service identity and log level should be labels. Trace IDs, span IDs, request IDs, user IDs, and similar high-cardinality values must not be Loki labels.

Application logs must not contain request bodies, authorization headers, JWTs, passwords, API keys, database credentials, or other secrets.

## Grafana Cloud correlation

Correlation is based on:

- `service.name` for traces and metrics.
- `deployment.environment.name=docker` for environment filtering.
- `trace_id` and `span_id` in structured log metadata.
- W3C trace context across HTTP and gRPC boundaries.

A log generated during an instrumented request therefore carries the trace identifier needed to navigate from the log to the corresponding distributed trace. Metrics and traces share the same service resource identity.

## Docker permissions

Alloy needs read-only access to `/var/run/docker.sock` to collect stdout/container logs without modifying application logging. The socket is not exposed to backend application containers.

## Environment variables

Applications:

```text
OTEL_EXPORTER_OTLP_ENDPOINT=alloy:4317
OTEL_EXPORTER_OTLP_PROTOCOL=grpc
OTEL_RESOURCE_ATTRIBUTES=service.name=<service>,deployment.environment.name=docker
```

Alloy only:

```text
GRAFANA_CLOUD_OTLP_ENDPOINT=<Grafana Cloud OTLP endpoint>
GRAFANA_CLOUD_LOKI_ENDPOINT=<Grafana Cloud Loki push endpoint>
GRAFANA_CLOUD_INSTANCE_ID=<Grafana Cloud instance/stack identifier>
GRAFANA_CLOUD_API_KEY=<Grafana Cloud API key>
```

No Grafana Cloud credential is injected into the BFF or microservice containers.

## Validation

Run:

```bash
docker compose config
docker compose up -d --build
docker compose ps
docker compose logs alloy
```

Then generate a request through the BFF and verify in Grafana Cloud that:

1. the trace contains BFF and downstream gRPC spans;
2. metrics use the expected `service.name` and environment metadata;
3. the corresponding application log contains the same `trace_id` and `span_id` as the trace;
4. no JWT, password, Authorization header, or other secret appears in logs.
