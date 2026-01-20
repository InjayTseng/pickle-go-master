package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig holds configuration for the rate limiter
type RateLimitConfig struct {
	// Requests is the maximum number of requests allowed within the window
	Requests int

	// Window is the time window for rate limiting
	Window time.Duration

	// KeyFunc is a function that returns the key for rate limiting (e.g., IP address, user ID)
	KeyFunc func(*gin.Context) string

	// ExceededHandler is called when rate limit is exceeded
	ExceededHandler gin.HandlerFunc

	// SkipFunc is a function that returns true if the request should skip rate limiting
	SkipFunc func(*gin.Context) bool
}

// visitor tracks request counts for a single key
type visitor struct {
	requests  int
	expiresAt time.Time
}

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	config   RateLimitConfig
	visitors map[string]*visitor
	mu       sync.RWMutex
}

// NewRateLimiter creates a new rate limiter with the given configuration
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		config:   config,
		visitors: make(map[string]*visitor),
	}

	// Start cleanup goroutine
	go rl.cleanupVisitors()

	return rl
}

// cleanupVisitors periodically removes expired visitors
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, v := range rl.visitors {
			if now.After(v.expiresAt) {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(key string) (allowed bool, remaining int, resetAt time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	v, exists := rl.visitors[key]
	if !exists || now.After(v.expiresAt) {
		// Create new visitor or reset expired one
		rl.visitors[key] = &visitor{
			requests:  1,
			expiresAt: now.Add(rl.config.Window),
		}
		return true, rl.config.Requests - 1, now.Add(rl.config.Window)
	}

	// Check if limit exceeded
	if v.requests >= rl.config.Requests {
		return false, 0, v.expiresAt
	}

	// Increment requests
	v.requests++
	return true, rl.config.Requests - v.requests, v.expiresAt
}

// Middleware returns a gin middleware for rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if we should skip rate limiting
		if rl.config.SkipFunc != nil && rl.config.SkipFunc(c) {
			c.Next()
			return
		}

		// Get the key for this request
		key := rl.config.KeyFunc(c)

		// Check rate limit
		allowed, remaining, resetAt := rl.Allow(key)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", formatInt(rl.config.Requests))
		c.Header("X-RateLimit-Remaining", formatInt(remaining))
		c.Header("X-RateLimit-Reset", formatInt64(resetAt.Unix()))

		if !allowed {
			// Rate limit exceeded
			if rl.config.ExceededHandler != nil {
				rl.config.ExceededHandler(c)
			} else {
				c.Header("Retry-After", formatInt64(resetAt.Unix()-time.Now().Unix()))
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "RATE_LIMIT_EXCEEDED",
						"message": "Too many requests. Please try again later.",
					},
				})
			}
			return
		}

		c.Next()
	}
}

// DefaultRateLimitConfig returns a default rate limit configuration
// Default: 100 requests per minute per IP
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Requests: 100,
		Window:   time.Minute,
		KeyFunc:  GetClientIP,
	}
}

// StrictRateLimitConfig returns a stricter rate limit configuration
// Suitable for sensitive endpoints like login, registration
// Default: 10 requests per minute per IP
func StrictRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
		KeyFunc:  GetClientIP,
	}
}

// APIRateLimitConfig returns a rate limit configuration suitable for API endpoints
// Default: 60 requests per minute per IP
func APIRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Requests: 60,
		Window:   time.Minute,
		KeyFunc:  GetClientIP,
	}
}

// GetClientIP returns the client's IP address, handling proxies
func GetClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header (for proxies like Cloudflare)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// Take the first IP (original client)
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// Check CF-Connecting-IP header (Cloudflare specific)
	if cfIP := c.GetHeader("CF-Connecting-IP"); cfIP != "" {
		return cfIP
	}

	// Fall back to RemoteAddr
	return c.ClientIP()
}

// GetUserID returns the user ID from context (for authenticated rate limiting)
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("userID"); exists {
		if id, ok := userID.(string); ok {
			return "user:" + id
		}
	}
	// Fall back to IP if no user ID
	return "ip:" + GetClientIP(c)
}

// RateLimit returns a rate limiting middleware with default configuration
func RateLimit() gin.HandlerFunc {
	limiter := NewRateLimiter(DefaultRateLimitConfig())
	return limiter.Middleware()
}

// RateLimitStrict returns a strict rate limiting middleware
func RateLimitStrict() gin.HandlerFunc {
	limiter := NewRateLimiter(StrictRateLimitConfig())
	return limiter.Middleware()
}

// RateLimitAPI returns a rate limiting middleware for API endpoints
func RateLimitAPI() gin.HandlerFunc {
	limiter := NewRateLimiter(APIRateLimitConfig())
	return limiter.Middleware()
}

// RateLimitWithConfig returns a rate limiting middleware with custom configuration
func RateLimitWithConfig(config RateLimitConfig) gin.HandlerFunc {
	limiter := NewRateLimiter(config)
	return limiter.Middleware()
}

// formatInt64 converts an int64 to a string
func formatInt64(n int64) string {
	if n == 0 {
		return "0"
	}

	var result []byte
	negative := n < 0
	if negative {
		n = -n
	}

	for n > 0 {
		result = append([]byte{byte(n%10 + '0')}, result...)
		n /= 10
	}

	if negative {
		result = append([]byte{'-'}, result...)
	}

	return string(result)
}
