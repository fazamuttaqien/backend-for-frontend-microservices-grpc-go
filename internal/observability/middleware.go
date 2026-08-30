package observability

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type requestIDKey struct{}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey{}).(string)
	return value
}

func HTTPMiddleware(serviceName string, next http.Handler) http.Handler {
	metrics := NewMetrics(serviceName)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = RequestID()
		}
		w.Header().Set("X-Request-ID", requestID)

		ctx := WithRequestID(r.Context(), requestID)
		r = r.WithContext(ctx)
		rw := &statusWriter{ResponseWriter: w}
		started := time.Now()

		instrumented := otelhttp.NewMiddleware(serviceName, otelhttp.WithServerName(serviceName))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if sc := trace.SpanContextFromContext(r.Context()); sc.IsValid() {
				w.Header().Set("X-Trace-ID", sc.TraceID().String())
			}
			next.ServeHTTP(w, r)
			statusCode := rw.status
			if statusCode == 0 {
				statusCode = http.StatusOK
			}
			RecordHTTP(r.Context(), metrics, serviceName, r.Method, r.Pattern, statusCode, started)
			Logger(r.Context()).Info("http request completed",
				"service", serviceName,
				"method", r.Method,
				"path", r.URL.Path,
				"status", statusCode,
				"duration_ms", time.Since(started).Milliseconds(),
			)
		}))

		instrumented.ServeHTTP(rw, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(body)
}

func Logger(ctx context.Context) *slog.Logger {
	logger := slog.Default()
	attrs := make([]any, 0, 4)
	if requestID := RequestIDFromContext(ctx); requestID != "" {
		attrs = append(attrs, "request_id", requestID)
	}
	if traceID := traceID(ctx); traceID != "" {
		attrs = append(attrs, "trace_id", traceID)
	}
	if spanID := spanID(ctx); spanID != "" {
		attrs = append(attrs, "span_id", spanID)
	}
	return logger.With(attrs...)
}

func traceID(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return ""
	}
	return sc.TraceID().String()
}

func spanID(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return ""
	}
	return sc.SpanID().String()
}

func UnaryServerInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("x-request-id"); len(values) > 0 && values[0] != "" {
				ctx = WithRequestID(ctx, values[0])
			}
		}
		started := time.Now()
		resp, err := handler(ctx, req)
		code := status.Code(err)
		Logger(ctx).Info("grpc request completed",
			"service", serviceName,
			"rpc_method", info.FullMethod,
			"status", code.String(),
			"duration_ms", time.Since(started).Milliseconds(),
		)
		return resp, err
	}
}

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		requestID := RequestIDFromContext(ctx)
		if requestID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", requestID)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
