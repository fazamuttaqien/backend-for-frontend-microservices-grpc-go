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
	RedisAddress          string
	RedisPassword         string
	RedisDB               int
	RedisTimeout          time.Duration
	ProductCacheTTL       time.Duration
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
		RedisAddress:          env("REDIS_ADDRESS", "localhost:6379"),
		RedisPassword:         os.Getenv("REDIS_PASSWORD"),
		RedisDB:               envInt("REDIS_DB", 0),
		RedisTimeout:          time.Duration(envInt("REDIS_TIMEOUT_MS", 100)) * time.Millisecond,
		ProductCacheTTL:       time.Duration(envInt("PRODUCT_CACHE_TTL_SECONDS", 60)) * time.Second,
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" { return v }
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil { return n }
	}
	return def
}
