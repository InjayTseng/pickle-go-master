package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/anthropics/pickle-go/apps/api/internal/dto"
	"github.com/anthropics/pickle-go/apps/api/internal/middleware"
	"github.com/anthropics/pickle-go/apps/api/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userRepo         *repository.UserRepository
	eventRepo        *repository.EventRepository
	registrationRepo *repository.RegistrationRepository
	notificationRepo *repository.NotificationRepository
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userRepo *repository.UserRepository, eventRepo *repository.EventRepository, registrationRepo *repository.RegistrationRepository, notificationRepo *repository.NotificationRepository) *UserHandler {
	return &UserHandler{
		userRepo:         userRepo,
		eventRepo:        eventRepo,
		registrationRepo: registrationRepo,
		notificationRepo: notificationRepo,
	}
}

// GetMyEvents returns events hosted by the current user
// GET /api/v1/users/me/events
func (h *UserHandler) GetMyEvents(c *gin.Context) {
	claims, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("UNAUTHORIZED", "Not authenticated"))
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("INVALID_TOKEN", "Invalid user ID"))
		return
	}

	events, err := h.eventRepo.FindByHostID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to get events"))
		return
	}

	// Convert to response format
	eventResponses := make([]dto.EventResponse, 0, len(events))
	for _, event := range events {
		eventResponses = append(eventResponses, dto.EventResponse{
			ID:        event.ID.String(),
			Title:     event.Title,
			EventDate: event.EventDate.Format("2006-01-02"),
			StartTime: event.StartTime,
			EndTime:   event.EndTime,
			Location: dto.LocationResponse{
				Name:    event.LocationName,
				Address: event.LocationAddress,
				Lat:     event.Latitude,
				Lng:     event.Longitude,
			},
			Capacity:        event.Capacity,
			SkillLevel:      string(event.SkillLevel),
			SkillLevelLabel: event.GetSkillLevelLabel(),
			Fee:             event.Fee,
			Status:          string(event.Status),
		})
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"events": eventResponses,
		"total":  len(eventResponses),
	}))
}

// GetMyRegistrations returns events the current user has registered for
// GET /api/v1/users/me/registrations
func (h *UserHandler) GetMyRegistrations(c *gin.Context) {
	claims, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("UNAUTHORIZED", "Not authenticated"))
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("INVALID_TOKEN", "Invalid user ID"))
		return
	}

	registrations, err := h.registrationRepo.FindByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to get registrations"))
		return
	}

	// Get event details for each registration
	var responses []gin.H
	for _, reg := range registrations {
		event, err := h.eventRepo.FindByID(c.Request.Context(), reg.EventID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue // Skip if event not found
			}
			continue
		}

		responses = append(responses, gin.H{
			"id":                reg.ID.String(),
			"event_id":          reg.EventID.String(),
			"status":            string(reg.Status),
			"waitlist_position": reg.WaitlistPosition,
			"registered_at":     reg.RegisteredAt,
			"event": gin.H{
				"id":          event.ID.String(),
				"title":       event.Title,
				"event_date":  event.EventDate.Format("2006-01-02"),
				"start_time":  event.StartTime,
				"location":    event.LocationName,
				"skill_level": string(event.SkillLevel),
				"status":      string(event.Status),
			},
		})
	}

	if responses == nil {
		responses = []gin.H{}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"registrations": responses,
		"total":         len(responses),
	}))
}

// GetMyNotifications returns notifications for the current user
// GET /api/v1/users/me/notifications
func (h *UserHandler) GetMyNotifications(c *gin.Context) {
	claims, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("UNAUTHORIZED", "Not authenticated"))
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("INVALID_TOKEN", "Invalid user ID"))
		return
	}

	// Check if notification repo is available
	if h.notificationRepo == nil {
		c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
			"notifications": []interface{}{},
			"total":         0,
			"unread_count":  0,
		}))
		return
	}

	// Get notifications (limit 50, offset 0 by default)
	notifications, err := h.notificationRepo.FindByUserID(c.Request.Context(), userID, 50, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to get notifications"))
		return
	}

	// Get unread count
	unreadCount, err := h.notificationRepo.CountUnread(c.Request.Context(), userID)
	if err != nil {
		unreadCount = 0
	}

	// Convert to response format
	notificationResponses := make([]gin.H, 0, len(notifications))
	for _, n := range notifications {
		notificationResponses = append(notificationResponses, gin.H{
			"id":         n.ID.String(),
			"type":       n.Type,
			"title":      n.Title,
			"message":    n.Message,
			"event_id":   n.EventID,
			"is_read":    n.IsRead,
			"created_at": n.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"notifications": notificationResponses,
		"total":         len(notificationResponses),
		"unread_count":  unreadCount,
	}))
}

// Legacy handlers for backward compatibility

// GetCurrentUser is the legacy handler
func GetCurrentUser(c *gin.Context) {
	claims, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("UNAUTHORIZED", "User not authenticated"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"id":           claims.UserID,
		"display_name": claims.DisplayName,
	}))
}

// GetMyEvents is the legacy handler
func GetMyEvents(c *gin.Context) {
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"events": []interface{}{},
		"total":  0,
	}))
}

// GetMyRegistrations is the legacy handler
func GetMyRegistrations(c *gin.Context) {
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"registrations": []interface{}{},
		"total":         0,
	}))
}

// GetMyNotifications is the legacy handler
func GetMyNotifications(c *gin.Context) {
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"notifications": []interface{}{},
		"total":         0,
	}))
}
