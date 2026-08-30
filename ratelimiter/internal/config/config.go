package config

import "os"

type Config struct {
	PostgresDSN string
	RedisAddr   string
	HTTPport    string
}

func Load() Config {
	return Config{
		PostgresDSN: getEnv("POSTGRES_DSN", "postgres://ratelimiter:ratelimiter@localhost:5432/ratelimiter?sslmode=disable"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		HTTPport:    getEnv("HTTP_PORT", "8000"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
