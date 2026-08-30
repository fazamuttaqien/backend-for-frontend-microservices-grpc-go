package observability

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)
type requestIDKey struct{}
func WithRequestID(ctx context.Context, requestID string)context.Context{return context.WithValue(ctx,requestIDKey{},requestID)}
func RequestIDFromContext(ctx context.Context)string{value,_:=ctx.Value(requestIDKey{}).(string);return value}
func HTTPMiddleware(serviceName string,next http.Handler)http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){requestID:=r.Header.Get("X-Request-ID");if requestID==""{requestID=RequestID()};w.Header().Set("X-Request-ID",requestID);r=r.WithContext(WithRequestID(r.Context(),requestID));rw:=&statusWriter{ResponseWriter:w};started:=time.Now();instrumented:=otelhttp.NewMiddleware(serviceName,otelhttp.WithServerName(serviceName))(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){if sc:=trace.SpanContextFromContext(r.Context());sc.IsValid(){w.Header().Set("X-Trace-ID",sc.TraceID().String())};next.ServeHTTP(w,r);Logger(r.Context()).Info("http request completed","service",serviceName,"method",r.Method,"path",r.URL.Path,"status",rw.status,"duration_ms",time.Since(started).Milliseconds())}));instrumented.ServeHTTP(rw,r)})}
type statusWriter struct{http.ResponseWriter;status int}
func(w *statusWriter)WriteHeader(code int){w.status=code;w.ResponseWriter.WriteHeader(code)}
func(w *statusWriter)Write(body []byte)(int,error){if w.status==0{w.status=http.StatusOK};return w.ResponseWriter.Write(body)}
type limiterEntry struct{window time.Time;count int}
type RateLimiter struct{mu sync.Mutex;entries map[string]limiterEntry;limit int;window time.Duration;maxEntries int}
func NewRateLimiter(limit int,window time.Duration)*RateLimiter{return &RateLimiter{entries:make(map[string]limiterEntry),limit:limit,window:window,maxEntries:10000}}
func(l *RateLimiter)Allow(key string,now time.Time)bool{l.mu.Lock();defer l.mu.Unlock();if len(l.entries)>=l.maxEntries{for k,v:=range l.entries{if now.Sub(v.window)>=l.window{delete(l.entries,k)}};if len(l.entries)>=l.maxEntries{return false}};entry,ok:=l.entries[key];if !ok||now.Sub(entry.window)>=l.window{l.entries[key]=limiterEntry{window:now,count:1};return true};if entry.count>=l.limit{return false};entry.count++;l.entries[key]=entry;return true}
func RateLimitMiddleware(l *RateLimiter,next http.Handler)http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){if r.URL.Path=="/healthz"||r.URL.Path=="/readyz"{next.ServeHTTP(w,r);return};key:=r.RemoteAddr;if host,_,err:=net.SplitHostPort(r.RemoteAddr);err==nil{key=host};if !l.Allow(key,time.Now()){w.Header().Set("Retry-After","60");w.Header().Set("Content-Type","application/json");w.WriteHeader(http.StatusTooManyRequests);_,_=w.Write([]byte(`{"error":"rate limit exceeded"}`));return};next.ServeHTTP(w,r)})}
func Logger(ctx context.Context)*slog.Logger{logger:=slog.Default();attrs:=make([]any,0,4);if requestID:=RequestIDFromContext(ctx);requestID!=""{attrs=append(attrs,"request_id",requestID)};if traceID:=traceID(ctx);traceID!=""{attrs=append(attrs,"trace_id",traceID)};if spanID:=spanID(ctx);spanID!=""{attrs=append(attrs,"span_id",spanID)};return logger.With(attrs...)}
func traceID(ctx context.Context)string{sc:=trace.SpanContextFromContext(ctx);if !sc.IsValid(){return ""};return sc.TraceID().String()}
func spanID(ctx context.Context)string{sc:=trace.SpanContextFromContext(ctx);if !sc.IsValid(){return ""};return sc.SpanID().String()}
func UnaryServerInterceptor(serviceName string)grpc.UnaryServerInterceptor{return func(ctx context.Context,req any,info *grpc.UnaryServerInfo,handler grpc.UnaryHandler)(any,error){if md,ok:=metadata.FromIncomingContext(ctx);ok{if values:=md.Get("x-request-id");len(values)>0&&values[0]!=""{ctx=WithRequestID(ctx,values[0])}};started:=time.Now();resp,err:=handler(ctx,req);Logger(ctx).Info("grpc request completed","service",serviceName,"rpc_method",info.FullMethod,"status",status.Code(err).String(),"duration_ms",time.Since(started).Milliseconds());return resp,err}}
func UnaryClientInterceptor()grpc.UnaryClientInterceptor{return func(ctx context.Context,method string,req,reply any,cc *grpc.ClientConn,invoker grpc.UnaryInvoker,opts ...grpc.CallOption)error{if requestID:=RequestIDFromContext(ctx);requestID!=""{ctx=metadata.AppendToOutgoingContext(ctx,"x-request-id",requestID)};if md,ok:=metadata.FromOutgoingContext(ctx);ok{if values:=md.Get("authorization");len(values)==1{ctx=metadata.AppendToOutgoingContext(ctx,"authorization",values[0])}};for attempt:=0;attempt<2;attempt++{err:=invoker(ctx,method,req,reply,cc,opts...);if err==nil||!retryableRPC(method,err)||attempt==1{return err};select{case<-time.After(50*time.Millisecond):case<-ctx.Done():return ctx.Err()}};return nil}}
func retryableRPC(method string,err error)bool{if status.Code(err)!=codes.Unavailable{return false};switch method{case "/user.v1.UserService/GetUser","/product.v1.ProductService/GetProduct","/product.v1.ProductService/ListProduct","/order.v1.OrderService/GetOrder","/order.v1.OrderService/ListOrders":return true};return false}
