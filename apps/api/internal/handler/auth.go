package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// LineCallbackRequest represents the request body for Line Login callback
type LineCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state"`
}

// LineCallback handles Line Login OAuth callback
func LineCallback(c *gin.Context) {
	var req LineCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body",
			},
		})
		return
	}

	// TODO: Implement Line Login flow
	// 1. Exchange authorization code for access token
	// 2. Get user profile from Line
	// 3. Create or update user in database
	// 4. Generate JWT token

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Line callback received",
		},
	})
}

// RefreshToken handles JWT token refresh
func RefreshToken(c *gin.Context) {
	// TODO: Implement token refresh logic

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Token refreshed",
		},
	})
}

// Logout handles user logout
func Logout(c *gin.Context) {
	// TODO: Implement logout logic (invalidate token if needed)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Logged out successfully",
		},
	})
}
