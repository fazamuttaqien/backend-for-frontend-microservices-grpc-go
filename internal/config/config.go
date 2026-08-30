package config

import (
	"os"
	"strconv"
)

type Config struct { HTTPPort string; UserGRPCPort string; ProductGRPCPort string; OrderGRPCPort string; UserServiceAddress string; ProductServiceAddress string; DatabaseURL string; ProductDatabaseURL string; OrderDatabaseURL string; JWTSecret string; JWTIssuer string; JWTTTLMinutes int }
func Load() Config { ttl,_:=strconv.Atoi(env("JWT_TTL_MINUTES","60")); if ttl<=0{ttl=60}; return Config{HTTPPort:env("BFF_HTTP_PORT","8080"),UserGRPCPort:env("USER_GRPC_PORT","50051"),ProductGRPCPort:env("PRODUCT_GRPC_PORT","50052"),OrderGRPCPort:env("ORDER_GRPC_PORT","50053"),UserServiceAddress:env("USER_SERVICE_ADDRESS","localhost:50051"),ProductServiceAddress:env("PRODUCT_SERVICE_ADDRESS","localhost:50052"),DatabaseURL:os.Getenv("DATABASE_URL"),ProductDatabaseURL:os.Getenv("PRODUCT_DATABASE_URL"),OrderDatabaseURL:os.Getenv("ORDER_DATABASE_URL"),JWTSecret:os.Getenv("JWT_SECRET"),JWTIssuer:env("JWT_ISSUER","user-service"),JWTTTLMinutes:ttl} }
func env(k,def string)string{if v:=os.Getenv(k);v!=""{return v};return def}
