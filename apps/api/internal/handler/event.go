package handler

import (
	"net/http"

	"github.com/anthropics/pickle-go/apps/api/internal/middleware"
	"github.com/gin-gonic/gin"
)

// ListEventsQuery represents query parameters for listing events
type ListEventsQuery struct {
	Lat        float64 `form:"lat"`
	Lng        float64 `form:"lng"`
	Radius     int     `form:"radius" binding:"max=50000"`
	SkillLevel string  `form:"skill_level"`
	Status     string  `form:"status"`
	Limit      int     `form:"limit" binding:"max=100"`
	Offset     int     `form:"offset"`
}

// CreateEventRequest represents the request body for creating an event
type CreateEventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	EventDate   string `json:"event_date" binding:"required"`
	StartTime   string `json:"start_time" binding:"required"`
	EndTime     string `json:"end_time"`
	Location    struct {
		Name          string  `json:"name" binding:"required"`
		Address       string  `json:"address"`
		Lat           float64 `json:"lat" binding:"required"`
		Lng           float64 `json:"lng" binding:"required"`
		GooglePlaceID string  `json:"google_place_id"`
	} `json:"location" binding:"required"`
	Capacity   int    `json:"capacity" binding:"required,min=4,max=20"`
	SkillLevel string `json:"skill_level" binding:"required,oneof=beginner intermediate advanced expert any"`
	Fee        int    `json:"fee" binding:"min=0,max=9999"`
}

// UpdateEventRequest represents the request body for updating an event
type UpdateEventRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	EventDate   *string `json:"event_date"`
	StartTime   *string `json:"start_time"`
	EndTime     *string `json:"end_time"`
	Capacity    *int    `json:"capacity" binding:"omitempty,min=4,max=20"`
	SkillLevel  *string `json:"skill_level" binding:"omitempty,oneof=beginner intermediate advanced expert any"`
	Fee         *int    `json:"fee" binding:"omitempty,min=0,max=9999"`
	Status      *string `json:"status" binding:"omitempty,oneof=open full cancelled"`
}

// ListEvents returns a list of events based on filters
func ListEvents(c *gin.Context) {
	var query ListEventsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	// Set defaults
	if query.Limit == 0 {
		query.Limit = 20
	}
	if query.Radius == 0 {
		query.Radius = 10000 // 10km default
	}

	// TODO: Implement geo query with PostGIS
	// ST_DWithin(location_point, ST_MakePoint(lng, lat)::geography, radius)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"events":   []interface{}{},
			"total":    0,
			"has_more": false,
		},
	})
}

// GetEvent returns a single event by ID
func GetEvent(c *gin.Context) {
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

	// TODO: Implement fetch event from database

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id": eventID,
		},
	})
}

// CreateEvent creates a new event
func CreateEvent(c *gin.Context) {
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

	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	// TODO: Implement create event in database
	_ = claims

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":        "new-event-id",
			"share_url": "https://picklego.tw/g/abc123",
		},
	})
}

// UpdateEvent updates an existing event
func UpdateEvent(c *gin.Context) {
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

	var req UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	// TODO: Implement update event in database
	// 1. Check if user is the host
	// 2. Update event fields
	_ = claims

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id": eventID,
		},
	})
}

// DeleteEvent cancels an event
func DeleteEvent(c *gin.Context) {
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

	// TODO: Implement cancel event
	// 1. Check if user is the host
	// 2. Update event status to cancelled
	// 3. Notify all registered users
	_ = claims

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Event cancelled successfully",
		},
	})
}
