package service

import (
	"context"
	"errors"

	"github.com/anthropics/pickle-go/apps/api/internal/model"
	"github.com/anthropics/pickle-go/apps/api/internal/repository"
	"github.com/google/uuid"
)

// RegistrationService handles registration business logic
type RegistrationService struct {
	regRepo   *repository.RegistrationRepository
	eventRepo *repository.EventRepository
}

// NewRegistrationService creates a new RegistrationService
func NewRegistrationService(regRepo *repository.RegistrationRepository, eventRepo *repository.EventRepository) *RegistrationService {
	return &RegistrationService{
		regRepo:   regRepo,
		eventRepo: eventRepo,
	}
}

// Registration errors
var (
	ErrAlreadyRegistered = errors.New("user is already registered for this event")
	ErrNotRegistered     = errors.New("user is not registered for this event")
	ErrEventFull         = errors.New("event is full")
	ErrEventCancelled    = errors.New("event has been cancelled")
	ErrCannotCancel      = errors.New("cannot cancel registration")
)

// RegisterResult represents the result of a registration
type RegisterResult struct {
	Registration *model.Registration
	Status       model.RegistrationStatus
	Position     *int // waitlist position if applicable
}

// Register registers a user for an event
func (s *RegistrationService) Register(ctx context.Context, eventID, userID uuid.UUID) (*RegisterResult, error) {
	// Check if event exists and is open
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, ErrEventNotFound
	}

	if event.Status == model.EventStatusCancelled {
		return nil, ErrEventCancelled
	}

	// Check if user is already registered
	existingReg, err := s.regRepo.FindByEventAndUser(ctx, eventID, userID)
	if err == nil && existingReg != nil && existingReg.Status != model.RegistrationCancelled {
		return nil, ErrAlreadyRegistered
	}

	// Check if event is full
	confirmedCount, err := s.regRepo.CountConfirmed(ctx, eventID)
	if err != nil {
		return nil, err
	}

	var status model.RegistrationStatus
	var waitlistPos *int

	if confirmedCount >= event.Capacity {
		// Add to waitlist
		status = model.RegistrationWaitlist
		pos, err := s.regRepo.GetNextWaitlistPosition(ctx, eventID)
		if err != nil {
			return nil, err
		}
		waitlistPos = &pos
	} else {
		// Confirmed
		status = model.RegistrationConfirmed
	}

	reg := &model.Registration{
		ID:               uuid.New(),
		EventID:          eventID,
		UserID:           userID,
		Status:           status,
		WaitlistPosition: waitlistPos,
	}

	err = s.regRepo.Create(ctx, reg)
	if err != nil {
		return nil, err
	}

	// Update event status if full
	if status == model.RegistrationConfirmed && confirmedCount+1 >= event.Capacity {
		_ = s.eventRepo.UpdateStatus(ctx, eventID, model.EventStatusFull)
	}

	return &RegisterResult{
		Registration: reg,
		Status:       status,
		Position:     waitlistPos,
	}, nil
}

// CancelRegistration cancels a user's registration
func (s *RegistrationService) CancelRegistration(ctx context.Context, eventID, userID uuid.UUID) error {
	// Find user's registration
	reg, err := s.regRepo.FindByEventAndUser(ctx, eventID, userID)
	if err != nil {
		return ErrNotRegistered
	}

	if reg.Status == model.RegistrationCancelled {
		return ErrCannotCancel
	}

	wasConfirmed := reg.Status == model.RegistrationConfirmed

	// Cancel the registration
	err = s.regRepo.UpdateStatus(ctx, reg.ID, model.RegistrationCancelled)
	if err != nil {
		return err
	}

	// If user was confirmed, promote first waitlist person
	if wasConfirmed {
		promotedReg, _ := s.regRepo.PromoteFromWaitlist(ctx, eventID)
		if promotedReg != nil {
			// TODO: Send notification to promoted user
		}

		// Update event status back to open
		_ = s.eventRepo.UpdateStatus(ctx, eventID, model.EventStatusOpen)
	}

	return nil
}

// GetEventRegistrations gets all registrations for an event
func (s *RegistrationService) GetEventRegistrations(ctx context.Context, eventID uuid.UUID) ([]model.Registration, error) {
	return s.regRepo.FindByEventID(ctx, eventID)
}

// GetUserRegistrations gets all registrations for a user
func (s *RegistrationService) GetUserRegistrations(ctx context.Context, userID uuid.UUID) ([]model.Registration, error) {
	return s.regRepo.FindByUserID(ctx, userID)
}
