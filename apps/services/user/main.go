package main

import (
  "context"
  "log"
  "net"
  "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/user/v1"
  "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/config"
  "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/application"
  "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/infrastructure/postgres"
  usergrpc "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/transport/grpc"
  "google.golang.org/grpc"
  "github.com/jackc/pgx/v5/pgxpool"
)
func main(){
 cfg:=config.Load();if cfg.DatabaseURL==""{log.Fatal("DATABASE_URL is required")}
 db,err:=pgxpool.New(context.Background(),cfg.DatabaseURL);if err!=nil{log.Fatal(err)};defer db.Close()
 if err=db.Ping(context.Background());err!=nil{log.Fatal(err)}
 repo:=postgres.NewUserRepository(db);app:=application.NewUserService(repo);h:=usergrpc.NewHandler(app)
 lis,err:=net.Listen("tcp",":"+cfg.UserGRPCPort);if err!=nil{log.Fatal(err)}
 s:=grpc.NewServer();userv1.RegisterUserServiceServer(s,h);log.Printf("user service listening on :%s",cfg.UserGRPCPort)
 if err:=s.Serve(lis);err!=nil{log.Fatal(err)}
}
