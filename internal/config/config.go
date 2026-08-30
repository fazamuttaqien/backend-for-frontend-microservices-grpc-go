package config

import "os"

type Config struct {
	HTTPPort string
}

func Load() Config {
	port := os.Getenv("BFF_HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	return Config{HTTPPort: port}
}
