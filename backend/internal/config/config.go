package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
	Env         string
	LogLevel    string
	PublicURL   string // used to auto-register QuePasa webhook (e.g. http://backend:8080)
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://simplix:simplix@localhost:5432/simplix"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret"),
		Port:        getEnv("PORT", "8080"),
		Env:         getEnv("ENV", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		PublicURL:   getEnv("PUBLIC_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
