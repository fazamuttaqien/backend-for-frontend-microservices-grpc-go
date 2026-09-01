package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"os"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Telemetry struct { TracerProvider *sdktrace.TracerProvider; MeterProvider *metric.MeterProvider }

func Setup(ctx context.Context, serviceName string) *Telemetry {
	if !observabilityEnabled() { return &Telemetry{} }
	logger := slog.Default()
	endpoint := env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	res, err := sdkresource.New(ctx, sdkresource.WithAttributes(attribute.String("service.name", serviceName)))
	if err != nil { logger.Warn("observability resource initialization failed", "error", err); res = sdkresource.Default() }
	traceCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	traceExporter, traceErr := otlptracegrpc.New(traceCtx, otlptracegrpc.WithEndpoint(endpoint), otlptracegrpc.WithInsecure()); cancel()
	metricCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	metricExporter, metricErr := otlpmetricgrpc.New(metricCtx, otlpmetricgrpc.WithEndpoint(endpoint), otlpmetricgrpc.WithInsecure()); cancel()
	if traceErr != nil || metricErr != nil {
		if traceExporter != nil { _ = traceExporter.Shutdown(context.Background()) }; if metricExporter != nil { _ = metricExporter.Shutdown(context.Background()) }
		if traceErr != nil { logger.Warn("OTLP trace exporter unavailable; tracing disabled", "error", traceErr) }; if metricErr != nil { logger.Warn("OTLP metric exporter unavailable; metrics disabled", "error", metricErr) }
		return &Telemetry{TracerProvider: sdktrace.NewTracerProvider(sdktrace.WithResource(res)), MeterProvider: metric.NewMeterProvider(metric.WithResource(res))}
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExporter), sdktrace.WithResource(res))
	mp := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(15*time.Second))), metric.WithResource(res))
	otel.SetTracerProvider(tp); otel.SetMeterProvider(mp)
	return &Telemetry{TracerProvider: tp, MeterProvider: mp}
}
func (t *Telemetry) Shutdown(ctx context.Context) { if t==nil{return}; if t.MeterProvider!=nil { _=t.MeterProvider.Shutdown(ctx) }; if t.TracerProvider!=nil { _=t.TracerProvider.Shutdown(ctx) } }
func NewLogger() *slog.Logger { return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level:slog.LevelInfo})) }
func RequestID() string { var b [16]byte; if _,err:=rand.Read(b[:]); err!=nil{return time.Now().UTC().Format("20060102T150405.000000000Z")}; return hex.EncodeToString(b[:]) }
func observabilityEnabled() bool { value:=os.Getenv("OBSERVABILITY_ENABLED"); if value=="" { return false }; enabled,err:=strconv.ParseBool(value); return err==nil && enabled }
func env(key,fallback string) string { if value:=os.Getenv(key); value!="" {return value}; return fallback }
