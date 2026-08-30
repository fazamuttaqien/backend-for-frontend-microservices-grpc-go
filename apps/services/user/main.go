package main

import (
	"context"
	"log"
	"net"
	"time"
	userv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/user/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/auth"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/config"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/application"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/infrastructure/postgres"
	usergrpc "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/transport/grpc"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)
func main(){cfg:=config.Load();if cfg.DatabaseURL==""{log.Fatal("DATABASE_URL is required")};if cfg.JWTSecret==""{log.Fatal("JWT_SECRET is required")};jwt,err:=auth.NewJWT(cfg.JWTSecret,cfg.JWTIssuer,time.Duration(cfg.JWTTTLMinutes)*time.Minute);if err!=nil{log.Fatal(err)};db,err:=pgxpool.New(context.Background(),cfg.DatabaseURL);if err!=nil{log.Fatal(err)};defer db.Close();if err=db.Ping(context.Background());err!=nil{log.Fatal(err)};repo:=postgres.NewUserRepository(db);app:=application.NewUserService(repo);h:=usergrpc.NewHandler(app,jwt);lis,err:=net.Listen("tcp",":"+cfg.UserGRPCPort);if err!=nil{log.Fatal(err)};interceptor:=auth.NewInterceptor(jwt,"/user.v1.UserService/CreateUser","/user.v1.UserService/Register","/user.v1.UserService/Login");s:=grpc.NewServer(grpc.UnaryInterceptor(interceptor.Unary()));userv1.RegisterUserServiceServer(s,h);log.Printf("user service listening on :%s",cfg.UserGRPCPort);if err:=s.Serve(lis);err!=nil{log.Fatal(err)}}
