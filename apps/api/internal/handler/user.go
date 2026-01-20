package handler

import (
	"net/http"

	"github.com/anthropics/pickle-go/apps/api/internal/middleware"
	"github.com/gin-gonic/gin"
)

// GetCurrentUser returns the current authenticated user
func GetCurrentUser(c *gin.Context) {
	claims, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	// TODO: Fetch full user data from database
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":           claims.UserID,
			"display_name": claims.DisplayName,
		},
	})
}

// GetMyEvents returns events hosted by the current user
func GetMyEvents(c *gin.Context) {
	claims, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	// TODO: Implement fetch user's events from database
	_ = claims

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"events": []interface{}{},
			"total":  0,
		},
	})
}

// GetMyRegistrations returns events the current user has registered for
func GetMyRegistrations(c *gin.Context) {
	claims, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	// TODO: Implement fetch user's registrations from database
	_ = claims

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"registrations": []interface{}{},
			"total":         0,
		},
	})
}

// GetMyNotifications returns notifications for the current user
func GetMyNotifications(c *gin.Context) {
	claims, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	// TODO: Implement fetch user's notifications from database
	_ = claims

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"notifications": []interface{}{},
			"total":         0,
		},
	})
}
