package observability

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Metrics struct {
	Requests metric.Int64Counter
	Duration metric.Float64Histogram
}

func NewMetrics(serviceName string) *Metrics {
	meter := otel.Meter("github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/" + serviceName)
	requests, err := meter.Int64Counter("http_server_requests_total", metric.WithDescription("Total HTTP requests"))
	if err != nil {
		slog.Default().Warn("initialize request metric", "error", err)
	}
	duration, err := meter.Float64Histogram("http_server_request_duration_seconds", metric.WithDescription("HTTP request duration in seconds"))
	if err != nil {
		slog.Default().Warn("initialize request duration metric", "error", err)
	}
	return &Metrics{Requests: requests, Duration: duration}
}

func RecordHTTP(ctx context.Context, metrics *Metrics, serviceName, method, route string, status int, started time.Time) {
	if metrics == nil {
		return
	}
	attrs := []attribute.KeyValue{
		attribute.String("service.name", serviceName),
		attribute.String("http.request.method", method),
		attribute.String("http.route", route),
		attribute.Int("http.response.status_code", status),
	}
	metrics.Requests.Add(ctx, 1, metric.WithAttributes(attrs...))
	metrics.Duration.Record(ctx, time.Since(started).Seconds(), metric.WithAttributes(attrs...))
}
