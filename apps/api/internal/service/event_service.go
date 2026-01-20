package service

import (
	"context"
	"errors"

	"github.com/anthropics/pickle-go/apps/api/internal/model"
	"github.com/anthropics/pickle-go/apps/api/internal/repository"
	"github.com/google/uuid"
)

// EventService handles event business logic
type EventService struct {
	eventRepo *repository.EventRepository
	userRepo  *repository.UserRepository
}

// NewEventService creates a new EventService
func NewEventService(eventRepo *repository.EventRepository, userRepo *repository.UserRepository) *EventService {
	return &EventService{
		eventRepo: eventRepo,
		userRepo:  userRepo,
	}
}

// Common errors
var (
	ErrEventNotFound = errors.New("event not found")
	ErrNotAuthorized = errors.New("not authorized to perform this action")
)

// CreateEventInput represents the input for creating an event
type CreateEventInput struct {
	HostID          uuid.UUID
	Title           *string
	Description     *string
	EventDate       string
	StartTime       string
	EndTime         *string
	LocationName    string
	LocationAddress *string
	Latitude        float64
	Longitude       float64
	GooglePlaceID   *string
	Capacity        int
	SkillLevel      model.SkillLevel
	Fee             int
}

// CreateEvent creates a new event
func (s *EventService) CreateEvent(ctx context.Context, input CreateEventInput) (*model.Event, error) {
	event := &model.Event{
		ID:              uuid.New(),
		HostID:          input.HostID,
		Title:           input.Title,
		Description:     input.Description,
		LocationName:    input.LocationName,
		LocationAddress: input.LocationAddress,
		Latitude:        input.Latitude,
		Longitude:       input.Longitude,
		GooglePlaceID:   input.GooglePlaceID,
		Capacity:        input.Capacity,
		SkillLevel:      input.SkillLevel,
		Fee:             input.Fee,
		Status:          model.EventStatusOpen,
	}

	err := s.eventRepo.Create(ctx, event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

// GetEvent gets an event by ID
func (s *EventService) GetEvent(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	return s.eventRepo.FindByID(ctx, id)
}

// ListEventsInput represents the input for listing events
type ListEventsInput struct {
	Lat        float64
	Lng        float64
	Radius     int
	SkillLevel string
	Status     string
	Limit      int
	Offset     int
}

// ListEvents lists events based on filters
func (s *EventService) ListEvents(ctx context.Context, input ListEventsInput) ([]model.EventSummary, error) {
	filter := repository.EventFilter{
		Lat:        input.Lat,
		Lng:        input.Lng,
		Radius:     input.Radius,
		SkillLevel: input.SkillLevel,
		Status:     input.Status,
		Limit:      input.Limit,
		Offset:     input.Offset,
	}

	return s.eventRepo.FindNearby(ctx, filter)
}

// UpdateEventInput represents the input for updating an event
type UpdateEventInput struct {
	EventID     uuid.UUID
	UserID      uuid.UUID
	Title       *string
	Description *string
	EventDate   *string
	StartTime   *string
	EndTime     *string
	Capacity    *int
	SkillLevel  *model.SkillLevel
	Fee         *int
	Status      *model.EventStatus
}

// UpdateEvent updates an existing event
func (s *EventService) UpdateEvent(ctx context.Context, input UpdateEventInput) (*model.Event, error) {
	event, err := s.eventRepo.FindByID(ctx, input.EventID)
	if err != nil {
		return nil, ErrEventNotFound
	}

	// Check if user is the host
	if event.HostID != input.UserID {
		return nil, ErrNotAuthorized
	}

	// Apply updates
	if input.Title != nil {
		event.Title = input.Title
	}
	if input.Description != nil {
		event.Description = input.Description
	}
	if input.Capacity != nil {
		event.Capacity = *input.Capacity
	}
	if input.SkillLevel != nil {
		event.SkillLevel = *input.SkillLevel
	}
	if input.Fee != nil {
		event.Fee = *input.Fee
	}
	if input.Status != nil {
		event.Status = *input.Status
	}

	err = s.eventRepo.Update(ctx, event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

// CancelEvent cancels an event
func (s *EventService) CancelEvent(ctx context.Context, eventID, userID uuid.UUID) error {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return ErrEventNotFound
	}

	// Check if user is the host
	if event.HostID != userID {
		return ErrNotAuthorized
	}

	return s.eventRepo.UpdateStatus(ctx, eventID, model.EventStatusCancelled)
}

// GetUserEvents gets events hosted by a user
func (s *EventService) GetUserEvents(ctx context.Context, userID uuid.UUID) ([]model.Event, error) {
	return s.eventRepo.FindByHostID(ctx, userID)
}
