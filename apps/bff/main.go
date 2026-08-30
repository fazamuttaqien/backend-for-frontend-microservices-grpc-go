package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/auth"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/bff"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func dial(ctx context.Context, address string) (*grpc.ClientConn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func main() {
	cfg := config.Load()
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	jwt, err := auth.NewJWT(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTTTL)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	userConn, err := dial(ctx, cfg.UserServiceAddress)
	if err != nil {
		log.Fatalf("connect user service: %v", err)
	}
	defer userConn.Close()

	productConn, err := dial(ctx, cfg.ProductServiceAddress)
	if err != nil {
		log.Fatalf("connect product service: %v", err)
	}
	defer productConn.Close()

	orderConn, err := dial(ctx, cfg.OrderServiceAddress)
	if err != nil {
		log.Fatalf("connect order service: %v", err)
	}
	defer orderConn.Close()

	clients := bff.Clients{
		User:    bff.NewUserClient(userConn),
		Product: bff.NewProductClient(productConn),
		Order:   bff.NewOrderClient(orderConn),
	}

	server := bff.NewServer(jwt, clients)
	httpServer := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("BFF listening on :%s", cfg.HTTPPort)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
