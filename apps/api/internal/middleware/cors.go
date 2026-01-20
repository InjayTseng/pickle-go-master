package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS configuration options
type CORSConfig struct {
	// AllowedOrigins is a list of origins that may access the resource
	// "*" allows all origins (not recommended for production with credentials)
	AllowedOrigins []string

	// AllowedMethods is a list of methods the client is allowed to use
	AllowedMethods []string

	// AllowedHeaders is a list of headers the client is allowed to send
	AllowedHeaders []string

	// ExposedHeaders indicates which headers can be exposed to the browser
	ExposedHeaders []string

	// AllowCredentials indicates whether the request can include user credentials
	AllowCredentials bool

	// MaxAge indicates how long the results of a preflight request can be cached (in seconds)
	MaxAge int
}

// DefaultCORSConfig returns a CORSConfig with default settings
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"Content-Type",
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORS returns a gin middleware for handling CORS with default configuration
func CORS(allowedOrigins []string) gin.HandlerFunc {
	config := DefaultCORSConfig()
	config.AllowedOrigins = allowedOrigins
	return CORSWithConfig(config)
}

// CORSWithConfig returns a gin middleware for handling CORS with custom configuration
func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	// Pre-compute methods and headers strings for efficiency
	methodsStr := strings.Join(config.AllowedMethods, ", ")
	headersStr := strings.Join(config.AllowedHeaders, ", ")
	exposedStr := strings.Join(config.ExposedHeaders, ", ")

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Skip if no origin (same-origin request)
		if origin == "" {
			c.Next()
			return
		}

		// Check if origin is allowed
		allowed := isOriginAllowed(origin, config.AllowedOrigins)

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)

			// Vary header is required for proper caching
			c.Header("Vary", "Origin")

			if config.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			if exposedStr != "" {
				c.Header("Access-Control-Expose-Headers", exposedStr)
			}
		}

		// Handle preflight requests (OPTIONS)
		if c.Request.Method == http.MethodOptions {
			if allowed {
				c.Header("Access-Control-Allow-Methods", methodsStr)
				c.Header("Access-Control-Allow-Headers", headersStr)

				if config.MaxAge > 0 {
					c.Header("Access-Control-Max-Age", formatInt(config.MaxAge))
				}
			}

			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Reject non-preflight requests from disallowed origins
		if !allowed {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}

// isOriginAllowed checks if the given origin is in the list of allowed origins
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		// Wildcard allows all origins
		if allowed == "*" {
			return true
		}

		// Exact match
		if allowed == origin {
			return true
		}

		// Subdomain wildcard (e.g., "*.picklego.tw")
		if strings.HasPrefix(allowed, "*.") {
			suffix := allowed[1:] // Remove the "*"
			if strings.HasSuffix(origin, suffix) {
				// Make sure we're matching a subdomain, not a suffix
				// e.g., "evil.com.picklego.tw" should not match "*.picklego.tw"
				prefix := strings.TrimSuffix(origin, suffix)
				if prefix != "" && !strings.Contains(prefix[len(prefix)-1:], ".") {
					continue
				}
				// Check for https:// prefix
				if strings.HasPrefix(origin, "https://") || strings.HasPrefix(origin, "http://") {
					return true
				}
			}
		}
	}

	return false
}

// formatInt converts an integer to a string
func formatInt(n int) string {
	return strings.TrimSpace(strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(
				string(rune(n/10000+'0'))+
					string(rune((n%10000)/1000+'0'))+
					string(rune((n%1000)/100+'0'))+
					string(rune((n%100)/10+'0'))+
					string(rune(n%10+'0')),
				"00000", ""),
			"0000", ""),
		"000", ""))
}

// ProductionCORSConfig returns a CORS configuration suitable for production
// Recommended for use with specific allowed origins
func ProductionCORSConfig(allowedOrigins []string) CORSConfig {
	config := DefaultCORSConfig()
	config.AllowedOrigins = allowedOrigins
	config.AllowCredentials = true
	return config
}
