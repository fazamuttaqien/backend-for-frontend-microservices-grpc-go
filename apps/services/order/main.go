package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os/signal"
	"syscall"
	"time"

	orderv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/order/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/auth"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/config"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/database"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/health"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/observability"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/application"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/broker"
	grpcclient "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/infrastructure/grpc"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/infrastructure/postgres"
	ordergrpc "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/transport/grpc"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func main() {
	logger := observability.NewLogger(); slog.SetDefault(logger); cfg := config.Load()
	if cfg.OrderDatabaseURL == "" || cfg.JWTSecret == "" || cfg.RabbitMQURL == "" { logger.Error("required configuration is missing"); return }
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM); defer cancel()
	startup, cancelStartup := context.WithTimeout(ctx, 5*time.Second); defer cancelStartup()
	db, err := database.OpenPostgres(startup, cfg.OrderDatabaseURL); if err != nil { logger.Error("initialize order database", "error", err); return }; defer db.Close()
	if err = db.Ping(startup); err != nil { logger.Error("ping order database", "error", err); return }
	telemetry := observability.Setup(ctx, "order-service"); defer telemetry.Shutdown(context.Background())
	jwt, err := auth.NewJWT(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTTTL); if err != nil { logger.Error("initialize JWT", "error", err); return }
	userConn, err := grpcclient.DialUser(ctx, cfg.UserServiceAddress, 5*time.Second); if err != nil { logger.Error("connect user service", "error", err); return }; defer userConn.Close()
	productConn, err := grpcclient.DialProduct(ctx, cfg.ProductServiceAddress, 5*time.Second); if err != nil { logger.Error("connect product service", "error", err); return }; defer productConn.Close()
	repo := postgres.NewOrderRepository(db); users := grpcclient.NewUserClient(userConn, 2*time.Second); products := grpcclient.NewProductClient(productConn, 2*time.Second)
	var publisher application.EventPublisher
	rabbitPublisher, err := broker.NewPublisher(cfg.RabbitMQURL); if err != nil { logger.Warn("RabbitMQ unavailable; OrderCreated events disabled", "error", err) } else { defer rabbitPublisher.Close(); publisher = broker.NewOrderEventPublisher(rabbitPublisher) }
	app := application.NewOrderService(repo, users, products, publisher); h := ordergrpc.NewHandler(app)
	lis, err := net.Listen("tcp", ":"+cfg.OrderGRPCPort); if err != nil { logger.Error("listen order service", "error", err); return }
	s := grpc.NewServer(grpc.MaxRecvMsgSize(4<<20), grpc.MaxSendMsgSize(4<<20), grpc.StatsHandler(otelgrpc.NewServerHandler()), grpc.ChainUnaryInterceptor(observability.UnaryServerInterceptor("order-service"), auth.UnaryServerInterceptor(jwt, map[string]bool{})))
	orderv1.RegisterOrderServiceServer(s, h); hs := health.Register(s, "order-service")
	logger.Info("order service listening", "port", cfg.OrderGRPCPort)
	go func(){ if err := s.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) { logger.Error("serve order service", "error", err); cancel() } }()
	<-ctx.Done(); hs.Shutdown(); gracefulStop(s, 10*time.Second)
}

func gracefulStop(s *grpc.Server, timeout time.Duration) { done := make(chan struct{}); go func(){ s.GracefulStop(); close(done) }(); select { case <-done: case <-time.After(timeout): s.Stop() } }
