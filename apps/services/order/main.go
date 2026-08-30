package main

import (
	"context"
	"log"
	"net"
	"time"

	orderv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/order/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/config"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/application"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/broker"
	grpcclient "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/infrastructure/grpc"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/infrastructure/postgres"
	ordergrpc "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/transport/grpc"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()
	if cfg.OrderDatabaseURL == "" { log.Fatal("ORDER_DATABASE_URL is required") }
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.OrderDatabaseURL)
	if err != nil { log.Fatal(err) }
	defer db.Close()
	if err = db.Ping(ctx); err != nil { log.Fatal(err) }

	userConn, err := grpcclient.DialUser(ctx, cfg.UserServiceAddress, 5*time.Second)
	if err != nil { log.Fatalf("connect user service: %v", err) }
	defer userConn.Close()
	productConn, err := grpcclient.DialProduct(ctx, cfg.ProductServiceAddress, 5*time.Second)
	if err != nil { log.Fatalf("connect product service: %v", err) }
	defer productConn.Close()

	repo := postgres.NewOrderRepository(db)
	users := grpcclient.NewUserClient(userConn, 2*time.Second)
	products := grpcclient.NewProductClient(productConn, 2*time.Second)

	var publisher application.EventPublisher
	rabbitPublisher, err := broker.NewPublisher(cfg.RabbitMQURL)
	if err != nil {
		log.Printf("RabbitMQ unavailable; OrderCreated events disabled: %v", err)
	} else {
		defer rabbitPublisher.Close()
		publisher = broker.NewOrderEventPublisher(rabbitPublisher)
	}

	app := application.NewOrderService(repo, users, products, publisher)
	h := ordergrpc.NewHandler(app)
	lis, err := net.Listen("tcp", ":"+cfg.OrderGRPCPort)
	if err != nil { log.Fatal(err) }
	s := grpc.NewServer()
	orderv1.RegisterOrderServiceServer(s, h)
	log.Printf("order service listening on :%s", cfg.OrderGRPCPort)
	if err := s.Serve(lis); err != nil { log.Fatal(err) }
}
