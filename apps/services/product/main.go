package main

import (
	"context"
	"log"
	"net"

	productv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/product/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/config"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/application"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/infrastructure/postgres"
	productredis "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/infrastructure/redis"
	productgrpc "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/transport/grpc"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()
	if cfg.ProductDatabaseURL == "" { log.Fatal("PRODUCT_DATABASE_URL is required") }
	db, err := pgxpool.New(context.Background(), cfg.ProductDatabaseURL)
	if err != nil { log.Fatal(err) }
	defer db.Close()
	if err = db.Ping(context.Background()); err != nil { log.Fatal(err) }

	repo := postgres.NewProductRepository(db)
	cache := productredis.NewProductCache(cfg.RedisAddress, cfg.RedisPassword, cfg.RedisDB, cfg.RedisTimeout, cfg.ProductCacheTTL)
	app := application.NewProductService(repo, cache)
	h := productgrpc.NewHandler(app)
	lis, err := net.Listen("tcp", ":"+cfg.ProductGRPCPort)
	if err != nil { log.Fatal(err) }
	s := grpc.NewServer()
	productv1.RegisterProductServiceServer(s, h)
	log.Printf("product service listening on :%s", cfg.ProductGRPCPort)
	if err := s.Serve(lis); err != nil { log.Fatal(err) }
}
