package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/anthropics/pickle-go/apps/api/internal/database"
	"github.com/anthropics/pickle-go/apps/api/internal/dto"
	"github.com/anthropics/pickle-go/apps/api/internal/middleware"
	"github.com/anthropics/pickle-go/apps/api/internal/model"
	"github.com/anthropics/pickle-go/apps/api/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// RegistrationHandler handles registration-related requests
type RegistrationHandler struct {
	registrationRepo *repository.RegistrationRepository
	eventRepo        *repository.EventRepository
	notificationRepo *repository.NotificationRepository
	txManager        *database.TxManager
}

// NewRegistrationHandler creates a new RegistrationHandler
func NewRegistrationHandler(registrationRepo *repository.RegistrationRepository, eventRepo *repository.EventRepository, notificationRepo *repository.NotificationRepository, txManager *database.TxManager) *RegistrationHandler {
	return &RegistrationHandler{
		registrationRepo: registrationRepo,
		eventRepo:        eventRepo,
		notificationRepo: notificationRepo,
		txManager:        txManager,
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

	// Check if event exists first (outside transaction for fast fail)
	event, err := h.eventRepo.FindByID(c.Request.Context(), eventID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse("NOT_FOUND", "Event not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to fetch event"))
		return
	}

	// Use transactional registration to prevent race conditions
	var registration *model.Registration
	err = h.txManager.WithTx(c.Request.Context(), func(tx *sqlx.Tx) error {
		var txErr error
		registration, txErr = h.registrationRepo.RegisterWithLock(c.Request.Context(), tx, eventID, userID)
		return txErr
	})

	if err != nil {
		switch {
		case errors.Is(err, repository.ErrEventNotOpen):
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("EVENT_CLOSED", "This event is not open for registration"))
		case errors.Is(err, repository.ErrHostCannotRegister):
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("HOST_CANNOT_REGISTER", "You cannot register for your own event"))
		case errors.Is(err, repository.ErrAlreadyRegistered):
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("ALREADY_REGISTERED", "You are already registered for this event"))
		case errors.Is(err, sql.ErrNoRows):
			c.JSON(http.StatusNotFound, dto.ErrorResponse("NOT_FOUND", "Event not found"))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to create registration"))
		}
		return
	}

	// Build response message
	var message string
	if registration.Status == model.RegistrationConfirmed {
		message = "報名成功！"
	} else {
		message = fmt.Sprintf("已加入候補（第 %d 位）", *registration.WaitlistPosition)
	}

	// Update event status to full if needed (outside transaction, non-critical)
	if registration.Status == model.RegistrationWaitlist && event.Status != model.EventStatusFull {
		h.eventRepo.UpdateStatus(c.Request.Context(), eventID, model.EventStatusFull)
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(dto.RegistrationResponse{
		ID:               registration.ID.String(),
		EventID:          eventID.String(),
		Status:           string(registration.Status),
		WaitlistPosition: registration.WaitlistPosition,
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

	// Find user's registration for this event (outside transaction for fast fail)
	registration, err := h.registrationRepo.FindByEventAndUser(c.Request.Context(), eventID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse("NOT_FOUND", "You are not registered for this event"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to fetch registration"))
		return
	}

	// Use transactional cancel and promote to prevent race conditions
	var promoted *model.Registration
	err = h.txManager.WithTx(c.Request.Context(), func(tx *sqlx.Tx) error {
		var txErr error
		promoted, txErr = h.registrationRepo.CancelAndPromote(c.Request.Context(), tx, registration.ID, eventID)
		return txErr
	})

	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAlreadyCancelled):
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("ALREADY_CANCELLED", "Registration is already cancelled"))
		case errors.Is(err, sql.ErrNoRows):
			c.JSON(http.StatusNotFound, dto.ErrorResponse("NOT_FOUND", "Registration not found"))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to cancel registration"))
		}
		return
	}

	// Send notification to promoted user (outside transaction, async-friendly)
	if promoted != nil && h.notificationRepo != nil {
		event, eventErr := h.eventRepo.FindByID(c.Request.Context(), eventID)
		if eventErr == nil {
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

	// Update event status if needed (outside transaction, non-critical)
	event, err := h.eventRepo.FindByID(c.Request.Context(), eventID)
	if err == nil && event.Status == model.EventStatusFull {
		confirmedCount, _ := h.registrationRepo.CountConfirmed(c.Request.Context(), eventID)
		if confirmedCount < event.Capacity {
			h.eventRepo.UpdateStatus(c.Request.Context(), eventID, model.EventStatusOpen)
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
