package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/anthropics/pickle-go/apps/api/internal/dto"
	"github.com/anthropics/pickle-go/apps/api/internal/middleware"
	"github.com/anthropics/pickle-go/apps/api/internal/model"
	"github.com/anthropics/pickle-go/apps/api/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EventHandler handles event-related requests
type EventHandler struct {
	eventRepo        *repository.EventRepository
	userRepo         *repository.UserRepository
	registrationRepo *repository.RegistrationRepository
}

// NewEventHandler creates a new EventHandler
func NewEventHandler(eventRepo *repository.EventRepository, userRepo *repository.UserRepository, registrationRepo *repository.RegistrationRepository) *EventHandler {
	return &EventHandler{
		eventRepo:        eventRepo,
		userRepo:         userRepo,
		registrationRepo: registrationRepo,
	}
}

// ListEvents returns a list of events based on filters
// GET /api/v1/events
func (h *EventHandler) ListEvents(c *gin.Context) {
	var query dto.ListEventsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Set defaults
	if query.Limit == 0 {
		query.Limit = 20
	}
	if query.Radius == 0 {
		query.Radius = 10000 // 10km default
	}

	filter := repository.EventFilter{
		Lat:        query.Lat,
		Lng:        query.Lng,
		Radius:     query.Radius,
		SkillLevel: query.SkillLevel,
		Status:     query.Status,
		Limit:      query.Limit,
		Offset:     query.Offset,
	}

	events, err := h.eventRepo.FindNearby(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to fetch events"))
		return
	}

	// Convert to response format
	eventResponses := make([]dto.EventResponse, 0, len(events))
	for _, event := range events {
		// Get host information
		host, err := h.userRepo.FindByID(c.Request.Context(), event.HostID)
		hostResponse := dto.UserResponse{}
		if err == nil {
			hostResponse = dto.FromUser(host)
		}

		eventResponses = append(eventResponses, dto.EventResponse{
			ID:        event.ID.String(),
			Host:      hostResponse,
			Title:     event.Title,
			EventDate: event.EventDate.Format("2006-01-02"),
			StartTime: event.StartTime,
			EndTime:   event.EndTime,
			Location: dto.LocationResponse{
				Name:          event.LocationName,
				Address:       event.LocationAddress,
				Lat:           event.Latitude,
				Lng:           event.Longitude,
				GooglePlaceID: event.GooglePlaceID,
			},
			Capacity:        event.Capacity,
			ConfirmedCount:  event.ConfirmedCount,
			WaitlistCount:   event.WaitlistCount,
			SkillLevel:      string(event.SkillLevel),
			SkillLevelLabel: event.GetSkillLevelLabel(),
			Fee:             event.Fee,
			Status:          string(event.Status),
		})
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(dto.EventListResponse{
		Events:  eventResponses,
		Total:   len(eventResponses),
		HasMore: len(eventResponses) == query.Limit,
	}))
}

// GetEvent returns a single event by ID
// GET /api/v1/events/:id
func (h *EventHandler) GetEvent(c *gin.Context) {
	eventIDStr := c.Param("id")
	if eventIDStr == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("VALIDATION_ERROR", "Event ID is required"))
		return
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("VALIDATION_ERROR", "Invalid event ID"))
		return
	}

	event, err := h.eventRepo.FindWithHost(c.Request.Context(), eventID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse("NOT_FOUND", "Event not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to fetch event"))
		return
	}

	// Get host information
	host, err := h.userRepo.FindByID(c.Request.Context(), event.HostID)
	hostResponse := dto.UserResponse{}
	if err == nil {
		hostResponse = dto.FromUser(host)
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(dto.EventResponse{
		ID:        event.ID.String(),
		Host:      hostResponse,
		Title:     event.Title,
		EventDate: event.EventDate.Format("2006-01-02"),
		StartTime: event.StartTime,
		EndTime:   event.EndTime,
		Location: dto.LocationResponse{
			Name:          event.LocationName,
			Address:       event.LocationAddress,
			Lat:           event.Latitude,
			Lng:           event.Longitude,
			GooglePlaceID: event.GooglePlaceID,
		},
		Capacity:        event.Capacity,
		ConfirmedCount:  event.ConfirmedCount,
		WaitlistCount:   event.WaitlistCount,
		SkillLevel:      string(event.SkillLevel),
		SkillLevelLabel: event.GetSkillLevelLabel(),
		Fee:             event.Fee,
		Status:          string(event.Status),
	}))
}

// CreateEvent creates a new event
// POST /api/v1/events
func (h *EventHandler) CreateEvent(c *gin.Context) {
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

	var req dto.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Parse event date
	eventDate, err := time.Parse("2006-01-02", req.EventDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("VALIDATION_ERROR", "Invalid event date format"))
		return
	}

	// Create event
	event := &model.Event{
		ID:              uuid.New(),
		HostID:          userID,
		EventDate:       eventDate,
		StartTime:       req.StartTime,
		LocationName:    req.Location.Name,
		Latitude:        req.Location.Lat,
		Longitude:       req.Location.Lng,
		Capacity:        req.Capacity,
		SkillLevel:      model.SkillLevel(req.SkillLevel),
		Fee:             req.Fee,
		Status:          model.EventStatusOpen,
	}

	if req.Title != "" {
		event.Title = &req.Title
	}
	if req.Description != "" {
		event.Description = &req.Description
	}
	if req.EndTime != "" {
		event.EndTime = &req.EndTime
	}
	if req.Location.Address != "" {
		event.LocationAddress = &req.Location.Address
	}
	if req.Location.GooglePlaceID != "" {
		event.GooglePlaceID = &req.Location.GooglePlaceID
	}

	if err := h.eventRepo.Create(c.Request.Context(), event); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to create event"))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(dto.CreateEventResponse{
		ID:       event.ID.String(),
		ShareURL: "https://picklego.tw/events/" + event.ID.String(),
	}))
}

// UpdateEvent updates an existing event
// PUT /api/v1/events/:id
func (h *EventHandler) UpdateEvent(c *gin.Context) {
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

	// Check if user is the host
	isHost, err := h.eventRepo.IsHost(c.Request.Context(), eventID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to verify ownership"))
		return
	}
	if !isHost {
		c.JSON(http.StatusForbidden, dto.ErrorResponse("FORBIDDEN", "You are not the host of this event"))
		return
	}

	var req dto.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Get existing event
	event, err := h.eventRepo.FindByID(c.Request.Context(), eventID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse("NOT_FOUND", "Event not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to fetch event"))
		return
	}

	// Update fields
	if req.Title != nil {
		event.Title = req.Title
	}
	if req.Description != nil {
		event.Description = req.Description
	}
	if req.EventDate != nil {
		eventDate, err := time.Parse("2006-01-02", *req.EventDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("VALIDATION_ERROR", "Invalid event date format"))
			return
		}
		event.EventDate = eventDate
	}
	if req.StartTime != nil {
		event.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		event.EndTime = req.EndTime
	}
	if req.Capacity != nil {
		event.Capacity = *req.Capacity
	}
	if req.SkillLevel != nil {
		event.SkillLevel = model.SkillLevel(*req.SkillLevel)
	}
	if req.Fee != nil {
		event.Fee = *req.Fee
	}
	if req.Status != nil {
		event.Status = model.EventStatus(*req.Status)
	}

	if err := h.eventRepo.Update(c.Request.Context(), event); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to update event"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"id":      event.ID.String(),
		"message": "Event updated successfully",
	}))
}

// DeleteEvent cancels an event
// DELETE /api/v1/events/:id
func (h *EventHandler) DeleteEvent(c *gin.Context) {
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

	// Check if user is the host
	isHost, err := h.eventRepo.IsHost(c.Request.Context(), eventID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to verify ownership"))
		return
	}
	if !isHost {
		c.JSON(http.StatusForbidden, dto.ErrorResponse("FORBIDDEN", "You are not the host of this event"))
		return
	}

	// Update event status to cancelled
	if err := h.eventRepo.UpdateStatus(c.Request.Context(), eventID, model.EventStatusCancelled); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("INTERNAL_ERROR", "Failed to cancel event"))
		return
	}

	// Cancel all registrations
	if err := h.registrationRepo.CancelAllByEventID(c.Request.Context(), eventID); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"message": "Event cancelled successfully",
	}))
}

// Legacy handlers for backward compatibility

// ListEvents is the legacy handler
func ListEvents(c *gin.Context) {
	c.JSON(http.StatusOK, dto.SuccessResponse(dto.EventListResponse{
		Events:  []dto.EventResponse{},
		Total:   0,
		HasMore: false,
	}))
}

// GetEvent is the legacy handler
func GetEvent(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse("NOT_IMPLEMENTED", "Handler not configured"))
}

// CreateEvent is the legacy handler
func CreateEvent(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse("NOT_IMPLEMENTED", "Handler not configured"))
}

// UpdateEvent is the legacy handler
func UpdateEvent(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse("NOT_IMPLEMENTED", "Handler not configured"))
}

// DeleteEvent is the legacy handler
func DeleteEvent(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse("NOT_IMPLEMENTED", "Handler not configured"))
}
