package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os/signal"
	"syscall"
	"time"

	userv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/user/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/auth"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/config"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/database"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/health"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/observability"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/application"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/infrastructure/postgres"
	usergrpc "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/transport/grpc"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func main() {
	logger := observability.NewLogger(); slog.SetDefault(logger)
	cfg := config.Load()
	if cfg.UserDatabaseURL == "" || cfg.JWTSecret == "" { logger.Error("required configuration is missing"); return }
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM); defer cancel()
	startup, cancelStartup := context.WithTimeout(ctx, 5*time.Second); defer cancelStartup()
	db, err := database.OpenPostgres(startup, cfg.UserDatabaseURL); if err != nil { logger.Error("initialize user database", "error", err); return }; defer db.Close()
	if err = db.Ping(startup); err != nil { logger.Error("ping user database", "error", err); return }
	telemetry := observability.Setup(ctx, "user-service"); defer telemetry.Shutdown(context.Background())
	repo := postgres.NewUserRepository(db); app := application.NewUserService(repo)
	jwt, err := auth.NewJWT(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTTTL); if err != nil { logger.Error("initialize JWT", "error", err); return }
	h := usergrpc.NewHandler(app, jwt)
	lis, err := net.Listen("tcp", ":"+cfg.UserGRPCPort); if err != nil { logger.Error("listen user service", "error", err); return }
	s := grpc.NewServer(grpc.MaxRecvMsgSize(4<<20), grpc.MaxSendMsgSize(4<<20), grpc.StatsHandler(otelgrpc.NewServerHandler()), grpc.ChainUnaryInterceptor(observability.UnaryServerInterceptor("user-service"), auth.UnaryServerInterceptor(jwt, map[string]bool{"/user.v1.UserService/CreateUser": true, "/user.v1.UserService/Register": true, "/user.v1.UserService/Login": true})))
	userv1.RegisterUserServiceServer(s, h); hs := health.Register(s, "user-service")
	logger.Info("user service listening", "port", cfg.UserGRPCPort)
	go func() { if err := s.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) { logger.Error("serve user service", "error", err); cancel() } }()
	<-ctx.Done(); hs.Shutdown(); gracefulStop(s, 10*time.Second)
}

func gracefulStop(s *grpc.Server, timeout time.Duration) { done := make(chan struct{}); go func(){ s.GracefulStop(); close(done) }(); select { case <-done: case <-time.After(timeout): s.Stop() } }
