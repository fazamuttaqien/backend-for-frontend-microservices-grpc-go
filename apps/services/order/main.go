package main

import (
	"context"
	"log/slog"
	"net"
	"time"

	orderv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/order/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/config"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/observability"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/application"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/broker"
	grpcclient "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/infrastructure/grpc"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/infrastructure/postgres"
	ordergrpc "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/transport/grpc"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func main() {
	logger := observability.NewLogger()
	slog.SetDefault(logger)
	cfg := config.Load()
	if cfg.OrderDatabaseURL == "" {
		logger.Error("ORDER_DATABASE_URL is required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	telemetry := observability.Setup(ctx, "order-service")
	defer telemetry.Shutdown(context.Background())

	db, err := pgxpool.New(ctx, cfg.OrderDatabaseURL)
	if err != nil {
		logger.Error("initialize order database", "error", err)
		return
	}
	defer db.Close()
	if err = db.Ping(ctx); err != nil {
		logger.Error("ping order database", "error", err)
		return
	}

	userConn, err := grpcclient.DialUser(ctx, cfg.UserServiceAddress, 5*time.Second)
	if err != nil {
		logger.Error("connect user service", "error", err)
		return
	}
	defer userConn.Close()
	productConn, err := grpcclient.DialProduct(ctx, cfg.ProductServiceAddress, 5*time.Second)
	if err != nil {
		logger.Error("connect product service", "error", err)
		return
	}
	defer productConn.Close()

	repo := postgres.NewOrderRepository(db)
	users := grpcclient.NewUserClient(userConn, 2*time.Second)
	products := grpcclient.NewProductClient(productConn, 2*time.Second)

	var publisher application.EventPublisher
	rabbitPublisher, err := broker.NewPublisher(cfg.RabbitMQURL)
	if err != nil {
		logger.Warn("RabbitMQ unavailable; OrderCreated events disabled", "error", err)
	} else {
		defer rabbitPublisher.Close()
		publisher = broker.NewOrderEventPublisher(rabbitPublisher)
	}

	app := application.NewOrderService(repo, users, products, publisher)
	h := ordergrpc.NewHandler(app)
	lis, err := net.Listen("tcp", ":"+cfg.OrderGRPCPort)
	if err != nil {
		logger.Error("listen order service", "error", err)
		return
	}
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(observability.UnaryServerInterceptor("order-service")),
	)
	orderv1.RegisterOrderServiceServer(s, h)
	logger.Info("order service listening", "port", cfg.OrderGRPCPort)
	if err := s.Serve(lis); err != nil {
		logger.Error("serve order service", "error", err)
	}
}
