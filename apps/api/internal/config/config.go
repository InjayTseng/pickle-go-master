package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server
	Port        string
	Environment string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTSecret string
	JWTExpiry string

	// Line Login
	LineChannelID     string
	LineChannelSecret string

	// CORS
	CORSAllowedOrigins []string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		Environment:        getEnv("ENVIRONMENT", "development"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/picklego?sslmode=disable"),
		RedisURL:           getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpiry:          getEnv("JWT_EXPIRY", "168h"),
		LineChannelID:      getEnv("LINE_CHANNEL_ID", ""),
		LineChannelSecret:  getEnv("LINE_CHANNEL_SECRET", ""),
		CORSAllowedOrigins: strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ","),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
