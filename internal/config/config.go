package config

import "os"

type Config struct { HTTPPort string; UserGRPCPort string; DatabaseURL string }
func Load() Config { return Config{HTTPPort:env("BFF_HTTP_PORT","8080"),UserGRPCPort:env("USER_GRPC_PORT","50051"),DatabaseURL:os.Getenv("DATABASE_URL")} }
func env(k,def string)string{if v:=os.Getenv(k);v!=""{return v};return def}
