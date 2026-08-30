package main

import (
	"context"
	"log/slog"
	"net"

	productv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/product/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/config"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/observability"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/application"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/infrastructure/postgres"
	productredis "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/infrastructure/redis"
	productgrpc "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/transport/grpc"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func main() {
	logger := observability.NewLogger()
	slog.SetDefault(logger)
	cfg := config.Load()
	if cfg.ProductDatabaseURL == "" {
		logger.Error("PRODUCT_DATABASE_URL is required")
		return
	}

	ctx := context.Background()
	telemetry := observability.Setup(ctx, "product-service")
	defer telemetry.Shutdown(context.Background())

	db, err := pgxpool.New(ctx, cfg.ProductDatabaseURL)
	if err != nil {
		logger.Error("initialize product database", "error", err)
		return
	}
	defer db.Close()
	if err = db.Ping(ctx); err != nil {
		logger.Error("ping product database", "error", err)
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
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(observability.UnaryServerInterceptor("product-service")),
	)
	productv1.RegisterProductServiceServer(s, h)
	logger.Info("product service listening", "port", cfg.ProductGRPCPort)
	if err := s.Serve(lis); err != nil {
		logger.Error("serve product service", "error", err)
	}
}
