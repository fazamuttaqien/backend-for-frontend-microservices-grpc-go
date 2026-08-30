package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os/signal"
	"syscall"
	"time"

	productv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/product/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/auth"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/config"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/database"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/health"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/observability"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/application"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/infrastructure/postgres"
	productredis "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/infrastructure/redis"
	productgrpc "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/transport/grpc"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func main() {
	logger := observability.NewLogger()
	slog.SetDefault(logger)
	cfg := config.Load()
	if cfg.ProductDatabaseURL == "" || cfg.JWTSecret == "" {
		logger.Error("required configuration is missing")
		return
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	startup, cancelStartup := context.WithTimeout(ctx, 5*time.Second)
	defer cancelStartup()
	db, err := database.OpenPostgres(startup, cfg.ProductDatabaseURL)
	if err != nil {
		logger.Error("initialize product database", "error", err)
		return
	}
	defer db.Close()
	if err = db.Ping(startup); err != nil {
		logger.Error("ping product database", "error", err)
		return
	}
	telemetry := observability.Setup(ctx, "product-service")
	defer telemetry.Shutdown(context.Background())
	jwt, err := auth.NewJWT(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTTTL)
	if err != nil {
		logger.Error("initialize JWT", "error", err)
		return
	}
	repo := postgres.NewProductRepository(db)
	cache := productredis.NewProductCache(cfg.RedisAddress, cfg.RedisPassword, cfg.RedisDB, cfg.RedisTimeout, cfg.ProductCacheTTL)
	app := application.NewProductService(repo, cache)
	h := productgrpc.NewHandler(app)
	lis, err := net.Listen("tcp", ":"+cfg.ProductGRPCPort)
	if err != nil {
		logger.Error("listen product service", "error", err)
		return
	}
	s := grpc.NewServer(grpc.MaxRecvMsgSize(4<<20), grpc.MaxSendMsgSize(4<<20), grpc.StatsHandler(otelgrpc.NewServerHandler()), grpc.ChainUnaryInterceptor(observability.UnaryServerInterceptor("product-service"), auth.UnaryServerInterceptor(jwt, map[string]bool{"/product.v1.ProductService/GetProduct": true, "/product.v1.ProductService/ListProduct": true})))
	productv1.RegisterProductServiceServer(s, h)
	hs := health.Register(s, "product-service")
	logger.Info("product service listening", "port", cfg.ProductGRPCPort)
	go func() {
		if err := s.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logger.Error("serve product service", "error", err)
			cancel()
		}
	}()
	<-ctx.Done()
	hs.Shutdown()
	gracefulStop(s, 10*time.Second)
}

func gracefulStop(s *grpc.Server, timeout time.Duration) {
	done := make(chan struct{})
	go func() { s.GracefulStop(); close(done) }()
	select {
	case <-done:
	case <-time.After(timeout):
		s.Stop()
	}
}
