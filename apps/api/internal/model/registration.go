package model

import (
	"time"

	"github.com/google/uuid"
)

// RegistrationStatus represents the status of a registration
type RegistrationStatus string

const (
	RegistrationConfirmed RegistrationStatus = "confirmed"
	RegistrationWaitlist  RegistrationStatus = "waitlist"
	RegistrationCancelled RegistrationStatus = "cancelled"
)

// Registration represents a user's registration for an event
type Registration struct {
	ID               uuid.UUID          `db:"id" json:"id"`
	EventID          uuid.UUID          `db:"event_id" json:"event_id"`
	UserID           uuid.UUID          `db:"user_id" json:"user_id"`
	Status           RegistrationStatus `db:"status" json:"status"`
	WaitlistPosition *int               `db:"waitlist_position" json:"waitlist_position,omitempty"`
	RegisteredAt     time.Time          `db:"registered_at" json:"registered_at"`
	ConfirmedAt      *time.Time         `db:"confirmed_at" json:"confirmed_at,omitempty"`
	CancelledAt      *time.Time         `db:"cancelled_at" json:"cancelled_at,omitempty"`
}

// RegistrationWithUser represents a registration with user details
type RegistrationWithUser struct {
	Registration
	User UserProfile `json:"user"`
}

// Notification represents a notification for a user
type Notification struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	UserID    uuid.UUID  `db:"user_id" json:"user_id"`
	EventID   *uuid.UUID `db:"event_id" json:"event_id,omitempty"`
	Type      string     `db:"type" json:"type"`
	Title     string     `db:"title" json:"title"`
	Message   *string    `db:"message" json:"message,omitempty"`
	IsRead    bool       `db:"is_read" json:"is_read"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

// NotificationType constants
const (
	NotificationWaitlistPromoted = "waitlist_promoted"
	NotificationEventCancelled   = "event_cancelled"
	NotificationEventUpdated     = "event_updated"
	NotificationEventReminder    = "event_reminder"
)
