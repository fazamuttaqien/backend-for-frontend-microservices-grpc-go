package main

import (
	"context"
	"log/slog"
	"net"

	userv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/user/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/auth"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/config"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/observability"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/application"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/infrastructure/postgres"
	usergrpc "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/transport/grpc"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func main() {
	logger := observability.NewLogger()
	slog.SetDefault(logger)
	cfg := config.Load()
	if cfg.UserDatabaseURL == "" {
		logger.Error("USER_DATABASE_URL is required")
		return
	}
	if cfg.JWTSecret == "" {
		logger.Error("JWT_SECRET is required")
		return
	}

	ctx := context.Background()
	telemetry := observability.Setup(ctx, "user-service")
	defer telemetry.Shutdown(context.Background())

	db, err := pgxpool.New(ctx, cfg.UserDatabaseURL)
	if err != nil {
		logger.Error("initialize user database", "error", err)
		return
	}
	defer db.Close()
	if err = db.Ping(ctx); err != nil {
		logger.Error("ping user database", "error", err)
		return
	}
	repo := postgres.NewUserRepository(db)
	app := application.NewUserService(repo)
	jwt, err := auth.NewJWT(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTTTL)
	if err != nil {
		logger.Error("initialize JWT", "error", err)
		return
	}
	h := usergrpc.NewHandler(app, jwt)
	lis, err := net.Listen("tcp", ":"+cfg.UserGRPCPort)
	if err != nil {
		logger.Error("listen user service", "error", err)
		return
	}
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(observability.UnaryServerInterceptor("user-service")),
	)
	userv1.RegisterUserServiceServer(s, h)
	logger.Info("user service listening", "port", cfg.UserGRPCPort)
	if err := s.Serve(lis); err != nil {
		logger.Error("serve user service", "error", err)
	}
}
