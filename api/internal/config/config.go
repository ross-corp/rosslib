package config

import "os"

type Config struct {
	DatabaseURL string
	RedisURL    string
	Port        string
	JWTSecret   string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-in-production"
	}
	return Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),
		Port:        port,
		JWTSecret:   secret,
	}
}
