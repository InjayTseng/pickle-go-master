package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns a gin middleware for logging requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Log format
		if raw != "" {
			path = path + "?" + raw
		}

		log.Printf("[%s] %d | %13v | %15s | %-7s %s",
			time.Now().Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			c.ClientIP(),
			c.Request.Method,
			path,
		)
	}
}
