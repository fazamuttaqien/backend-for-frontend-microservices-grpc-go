package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort              string
	UserGRPCPort          string
	ProductGRPCPort       string
	OrderGRPCPort         string
	UserServiceAddress    string
	ProductServiceAddress string
	OrderServiceAddress   string
	UserDatabaseURL       string
	ProductDatabaseURL    string
	OrderDatabaseURL      string
	JWTSecret             string
	JWTIssuer             string
	JWTTTL                time.Duration
}

func Load() Config {
	return Config{
		HTTPPort:              env("BFF_HTTP_PORT", "8080"),
		UserGRPCPort:          env("USER_GRPC_PORT", "50051"),
		ProductGRPCPort:       env("PRODUCT_GRPC_PORT", "50052"),
		OrderGRPCPort:         env("ORDER_GRPC_PORT", "50053"),
		UserServiceAddress:    env("USER_SERVICE_ADDRESS", "localhost:50051"),
		ProductServiceAddress: env("PRODUCT_SERVICE_ADDRESS", "localhost:50052"),
		OrderServiceAddress:   env("ORDER_SERVICE_ADDRESS", "localhost:50053"),
		UserDatabaseURL:       os.Getenv("USER_DATABASE_URL"),
		ProductDatabaseURL:    os.Getenv("PRODUCT_DATABASE_URL"),
		OrderDatabaseURL:      os.Getenv("ORDER_DATABASE_URL"),
		JWTSecret:             os.Getenv("JWT_SECRET"),
		JWTIssuer:             env("JWT_ISSUER", "user-service"),
		JWTTTL:                time.Duration(envInt("JWT_TTL_MINUTES", 60)) * time.Minute,
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
