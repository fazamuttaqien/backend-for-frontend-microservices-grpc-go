package config

import "os"

type Config struct {
	HTTPPort             string
	UserGRPCPort         string
	ProductGRPCPort      string
	OrderGRPCPort        string
	UserServiceAddress   string
	ProductServiceAddress string
	DatabaseURL          string
	ProductDatabaseURL   string
	OrderDatabaseURL     string
}

func Load() Config {
	return Config{
		HTTPPort:              env("BFF_HTTP_PORT", "8080"),
		UserGRPCPort:          env("USER_GRPC_PORT", "50051"),
		ProductGRPCPort:       env("PRODUCT_GRPC_PORT", "50052"),
		OrderGRPCPort:         env("ORDER_GRPC_PORT", "50053"),
		UserServiceAddress:    env("USER_SERVICE_ADDRESS", "localhost:50051"),
		ProductServiceAddress: env("PRODUCT_SERVICE_ADDRESS", "localhost:50052"),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		ProductDatabaseURL:    os.Getenv("PRODUCT_DATABASE_URL"),
		OrderDatabaseURL:      os.Getenv("ORDER_DATABASE_URL"),
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
