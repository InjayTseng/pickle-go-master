package handler

import (
	"net/http"

	"github.com/anthropics/pickle-go/apps/api/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterEvent registers the current user for an event
func RegisterEvent(c *gin.Context) {
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

	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Event ID is required",
			},
		})
		return
	}

	// TODO: Implement registration logic
	// 1. Check if event exists and is open
	// 2. Check if user is already registered
	// 3. Check if event is full
	// 4. If full, add to waitlist
	// 5. Create registration record
	_ = claims

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"status":  "confirmed",
			"message": "Registration successful!",
		},
	})
}

// CancelRegistration cancels the current user's registration for an event
func CancelRegistration(c *gin.Context) {
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

	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Event ID is required",
			},
		})
		return
	}

	// TODO: Implement cancel registration logic
	// 1. Find user's registration for this event
	// 2. Check if cancellation is allowed
	// 3. Update registration status to cancelled
	// 4. If user was confirmed, promote first waitlist user
	_ = claims

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Registration cancelled successfully",
		},
	})
}

// GetEventRegistrations returns all registrations for an event
func GetEventRegistrations(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Event ID is required",
			},
		})
		return
	}

	// TODO: Implement fetch registrations from database

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"confirmed": []interface{}{},
			"waitlist":  []interface{}{},
		},
	})
}
