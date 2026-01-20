package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/anthropics/pickle-go/apps/api/internal/dto"
	"github.com/anthropics/pickle-go/apps/api/internal/middleware"
	"github.com/anthropics/pickle-go/apps/api/internal/model"
	"github.com/anthropics/pickle-go/apps/api/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegistrationHandler handles registration-related requests
type RegistrationHandler struct {
	registrationRepo *repository.RegistrationRepository
	eventRepo        *repository.EventRepository
	notificationRepo *repository.NotificationRepository
}

// NewRegistrationHandler creates a new RegistrationHandler
func NewRegistrationHandler(registrationRepo *repository.RegistrationRepository, eventRepo *repository.EventRepository, notificationRepo *repository.NotificationRepository) *RegistrationHandler {
	return &RegistrationHandler{
		registrationRepo: registrationRepo,
		eventRepo:        eventRepo,
		notificationRepo: notificationRepo,
	}
}

// RegisterEvent registers the current user for an event
// POST /api/v1/events/:id/register
func (h *RegistrationHandler) RegisterEvent(c *gin.Context) {
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

	eventIDStr := c.Param("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("VALIDATION_ERROR", "Invalid event ID"))
		return
	}

	// Check if event exists and is open
	event, err := h.eventRepo.FindByID(c.Request.Context(), eventID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse("NOT_FOUND", "Event not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to fetch event"))
		return
	}

	// Check event status
	if event.Status == model.EventStatusCancelled {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("EVENT_CANCELLED", "This event has been cancelled"))
		return
	}
	if event.Status == model.EventStatusCompleted {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("EVENT_COMPLETED", "This event has already ended"))
		return
	}

	// Check if user is already registered
	hasRegistered, err := h.registrationRepo.HasUserRegistered(c.Request.Context(), eventID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to check registration status"))
		return
	}
	if hasRegistered {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ALREADY_REGISTERED", "You are already registered for this event"))
		return
	}

	// Check if user is the host
	if event.HostID == userID {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("HOST_CANNOT_REGISTER", "You cannot register for your own event"))
		return
	}

	// Get current confirmed count
	confirmedCount, err := h.registrationRepo.CountConfirmed(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to get registration count"))
		return
	}

	// Determine registration status
	var status model.RegistrationStatus
	var waitlistPosition *int
	var message string

	if confirmedCount < event.Capacity {
		// Event has space, confirm registration
		status = model.RegistrationConfirmed
		message = "報名成功！"
	} else {
		// Event is full, add to waitlist
		status = model.RegistrationWaitlist
		pos, err := h.registrationRepo.GetNextWaitlistPosition(c.Request.Context(), eventID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to get waitlist position"))
			return
		}
		waitlistPosition = &pos
		message = fmt.Sprintf("已加入候補（第 %d 位）", pos)

		// Update event status to full if it wasn't already
		if event.Status != model.EventStatusFull {
			h.eventRepo.UpdateStatus(c.Request.Context(), eventID, model.EventStatusFull)
		}
	}

	// Create registration
	registration := &model.Registration{
		ID:               uuid.New(),
		EventID:          eventID,
		UserID:           userID,
		Status:           status,
		WaitlistPosition: waitlistPosition,
	}

	if err := h.registrationRepo.Create(c.Request.Context(), registration); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to create registration"))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(dto.RegistrationResponse{
		ID:               registration.ID.String(),
		EventID:          eventID.String(),
		Status:           string(status),
		WaitlistPosition: waitlistPosition,
		Message:          message,
	}))
}

// CancelRegistration cancels the current user's registration for an event
// DELETE /api/v1/events/:id/register
func (h *RegistrationHandler) CancelRegistration(c *gin.Context) {
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

	eventIDStr := c.Param("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("VALIDATION_ERROR", "Invalid event ID"))
		return
	}

	// Find user's registration for this event
	registration, err := h.registrationRepo.FindByEventAndUser(c.Request.Context(), eventID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse("NOT_FOUND", "You are not registered for this event"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to fetch registration"))
		return
	}

	// Check if already cancelled
	if registration.Status == model.RegistrationCancelled {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ALREADY_CANCELLED", "Registration is already cancelled"))
		return
	}

	wasConfirmed := registration.Status == model.RegistrationConfirmed

	// Cancel registration
	if err := h.registrationRepo.UpdateStatus(c.Request.Context(), registration.ID, model.RegistrationCancelled); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to cancel registration"))
		return
	}

	// If user was confirmed, promote first waitlist user
	if wasConfirmed {
		promoted, err := h.registrationRepo.PromoteFromWaitlist(c.Request.Context(), eventID)
		if err == nil && promoted != nil {
			// Send notification to promoted user
			event, eventErr := h.eventRepo.FindByID(c.Request.Context(), eventID)
			if eventErr == nil && h.notificationRepo != nil {
				eventTitle := event.LocationName
				if event.Title != nil && *event.Title != "" {
					eventTitle = *event.Title
				}
				eventTitle = fmt.Sprintf("%s @ %s", event.EventDate.Format("01/02"), eventTitle)
				h.notificationRepo.CreateWaitlistPromotedNotification(
					c.Request.Context(),
					promoted.UserID,
					eventID,
					eventTitle,
				)
			}
		}

		// Check if event should be set back to open
		event, err := h.eventRepo.FindByID(c.Request.Context(), eventID)
		if err == nil && event.Status == model.EventStatusFull {
			confirmedCount, _ := h.registrationRepo.CountConfirmed(c.Request.Context(), eventID)
			if confirmedCount < event.Capacity {
				h.eventRepo.UpdateStatus(c.Request.Context(), eventID, model.EventStatusOpen)
			}
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"message": "Registration cancelled successfully",
	}))
}

// GetEventRegistrations returns all registrations for an event
// GET /api/v1/events/:id/registrations
func (h *RegistrationHandler) GetEventRegistrations(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("VALIDATION_ERROR", "Invalid event ID"))
		return
	}

	// Check if event exists
	exists, err := h.eventRepo.Exists(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to check event"))
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("NOT_FOUND", "Event not found"))
		return
	}

	// Get registrations with user details
	registrations, err := h.registrationRepo.FindWithUsersByEventID(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to fetch registrations"))
		return
	}

	// Separate confirmed and waitlisted
	var confirmed []gin.H
	var waitlist []gin.H

	for _, reg := range registrations {
		item := gin.H{
			"id":            reg.ID.String(),
			"user":          reg.User,
			"registered_at": reg.RegisteredAt,
		}

		if reg.Status == model.RegistrationConfirmed {
			confirmed = append(confirmed, item)
		} else if reg.Status == model.RegistrationWaitlist {
			item["waitlist_position"] = reg.WaitlistPosition
			waitlist = append(waitlist, item)
		}
	}

	if confirmed == nil {
		confirmed = []gin.H{}
	}
	if waitlist == nil {
		waitlist = []gin.H{}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"confirmed":       confirmed,
		"waitlist":        waitlist,
		"confirmed_count": len(confirmed),
		"waitlist_count":  len(waitlist),
	}))
}

// Legacy handlers for backward compatibility

// RegisterEvent is the legacy handler
func RegisterEvent(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse("NOT_IMPLEMENTED", "Handler not configured"))
}

// CancelRegistration is the legacy handler
func CancelRegistration(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse("NOT_IMPLEMENTED", "Handler not configured"))
}

// GetEventRegistrations is the legacy handler
func GetEventRegistrations(c *gin.Context) {
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"confirmed": []interface{}{},
		"waitlist":  []interface{}{},
	}))
}
