package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestHTTPMiddlewareCreatesServerTraceAndServerTimingWithoutIncomingTraceparent(t *testing.T) {
	previous := otel.GetTracerProvider()
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() {
		otel.SetTracerProvider(previous)
		_ = tp.Shutdown(t.Context())
	}()

	handler := HTTPMiddleware("bff", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !trace.SpanContextFromContext(r.Context()).IsValid() {
			t.Fatal("expected valid server span")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	serverTiming := res.Header().Get("Server-Timing")
	if !strings.HasPrefix(serverTiming, "traceparent;desc=\"00-") {
		t.Fatalf("unexpected Server-Timing: %q", serverTiming)
	}
	if !strings.HasSuffix(serverTiming, "-01\"") {
		t.Fatalf("unexpected Server-Timing flags: %q", serverTiming)
	}
}

func TestHTTPMiddlewareContinuesIncomingTraceparentAndReturnsServerTiming(t *testing.T) {
	previous := otel.GetTracerProvider()
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() {
		otel.SetTracerProvider(previous)
		_ = tp.Shutdown(t.Context())
	}()

	const traceID = "4bf92f3577b34da6a3ce929d0e0e4736"
	const parentSpanID = "00f067aa0ba902b7"
	traceparent := "00-" + traceID + "-" + parentSpanID + "-01"

	var observedTraceID string
	handler := HTTPMiddleware("bff", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc := trace.SpanContextFromContext(r.Context())
		if !sc.IsValid() {
			t.Fatal("expected valid continued server span")
		}
		observedTraceID = sc.TraceID().String()
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("traceparent", traceparent)
	req.Header.Set("tracestate", "vendor=value")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if observedTraceID != traceID {
		t.Fatalf("expected trace ID %s, got %s", traceID, observedTraceID)
	}
	serverTiming := res.Header().Get("Server-Timing")
	if !strings.Contains(serverTiming, "00-"+traceID+"-") {
		t.Fatalf("Server-Timing does not contain continued trace ID: %q", serverTiming)
	}
}
