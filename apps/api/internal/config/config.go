package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
// 應用程式的所有配置
type Config struct {
	// Server 伺服器設定
	Port        string
	Environment string

	// Database 資料庫設定
	DatabaseURL string

	// Redis 快取設定
	RedisURL string

	// JWT 認證設定
	JWTSecret string
	JWTExpiry string

	// Line Login Line 登入設定
	LineChannelID     string
	LineChannelSecret string
	LineRedirectURI   string

	// CORS 跨域設定
	CORSAllowedOrigins []string

	// Application 應用程式設定
	BaseURL string

	// Sentry 錯誤監控設定
	SentryDSN         string
	SentryEnvironment string
	SentryRelease     string
}

// Load loads configuration from environment variables
// 從環境變數載入配置
func Load() (*Config, error) {
	// Load .env file if it exists (for local development)
	// 如果 .env 檔案存在則載入（用於本地開發）
	_ = godotenv.Load()

	env := getEnv("ENVIRONMENT", "development")

	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		Environment:        env,
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/picklego?sslmode=disable"),
		RedisURL:           getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpiry:          getEnv("JWT_EXPIRY", "168h"),
		LineChannelID:      getEnv("LINE_CHANNEL_ID", ""),
		LineChannelSecret:  getEnv("LINE_CHANNEL_SECRET", ""),
		LineRedirectURI:    getEnv("LINE_REDIRECT_URI", "http://localhost:3000/auth/callback"),
		CORSAllowedOrigins: strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ","),
		BaseURL:            getEnv("BASE_URL", "http://localhost:3000"),
		// Sentry 設定
		SentryDSN:         getEnv("SENTRY_DSN", ""),
		SentryEnvironment: getEnv("SENTRY_ENVIRONMENT", env),
		SentryRelease:     getEnv("SENTRY_RELEASE", "1.0.0"),
	}

	return cfg, nil
}

// IsProduction returns true if running in production environment
// 判斷是否為生產環境
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if running in development environment
// 判斷是否為開發環境
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
