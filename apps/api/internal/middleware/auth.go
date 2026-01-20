package middleware

import (
	"net/http"
	"strings"

	"github.com/anthropics/pickle-go/apps/api/pkg/jwt"
	"github.com/gin-gonic/gin"
)

const (
	// AuthUserKey is the key used to store user info in gin context
	AuthUserKey = "auth_user"
)

// AuthRequired returns a gin middleware that validates JWT tokens
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authorization header is required",
				},
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "Invalid authorization header format",
				},
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := jwt.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "Invalid or expired token",
				},
			})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set(AuthUserKey, claims)
		c.Next()
	}
}

// GetAuthUser returns the authenticated user from gin context
func GetAuthUser(c *gin.Context) (*jwt.Claims, bool) {
	value, exists := c.Get(AuthUserKey)
	if !exists {
		return nil, false
	}
	claims, ok := value.(*jwt.Claims)
	return claims, ok
}
