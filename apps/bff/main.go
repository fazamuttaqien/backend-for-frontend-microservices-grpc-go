package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/auth"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/bff"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/config"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/observability"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func dial(ctx context.Context, address string) (*grpc.ClientConn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithStatsHandler(otelgrpc.NewClientHandler()), grpc.WithUnaryInterceptor(observability.UnaryClientInterceptor()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(4<<20), grpc.MaxCallSendMsgSize(4<<20)))
}

func main() {
	logger := observability.NewLogger()
	slog.SetDefault(logger)
	cfg := config.Load()
	if cfg.JWTSecret == "" {
		logger.Error("JWT_SECRET is required")
		return
	}
	jwt, err := auth.NewJWT(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTTTL)
	if err != nil {
		logger.Error("initialize JWT", "error", err)
		return
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	telemetry := observability.Setup(ctx, "bff")
	defer telemetry.Shutdown(context.Background())

	userConn, err := dial(ctx, cfg.UserServiceAddress)
	if err != nil {
		logger.Error("connect user service", "error", err)
		return
	}
	defer userConn.Close()
	productConn, err := dial(ctx, cfg.ProductServiceAddress)
	if err != nil {
		logger.Error("connect product service", "error", err)
		return
	}
	defer productConn.Close()
	orderConn, err := dial(ctx, cfg.OrderServiceAddress)
	if err != nil {
		logger.Error("connect order service", "error", err)
		return
	}
	defer orderConn.Close()

	clients := bff.Clients{User: bff.NewUserClient(userConn), Product: bff.NewProductClient(productConn), Order: bff.NewOrderClient(orderConn)}
	server := bff.NewServer(jwt, clients)
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/", server.Handler())

	cookieHandler := bff.CookieAuthHandler(mux, bff.CookieConfig{
		Name: cfg.AuthCookieName, Path: cfg.AuthCookiePath, Secure: cfg.AuthCookieSecure,
		SameSite: bff.ParseSameSite(cfg.AuthCookieSameSite), MaxAge: int(cfg.JWTTTL / time.Second), AllowedOrigins: cfg.CORSAllowedOrigins,
	})
	corsHandler := bff.CORSMiddleware(cfg.CORSAllowedOrigins, cfg.CORSCredentials)(cookieHandler)
	handler := observability.RateLimitMiddleware(observability.NewRateLimiter(120, time.Minute), observability.HTTPMiddleware("bff", corsHandler))

	httpServer := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		logger.Info("BFF listening", "port", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("serve BFF", "error", err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown BFF", "error", err)
	}
}
