package repository

import "errors"

// Common repository errors
var (
	// ErrNotFound is returned when a record is not found
	ErrNotFound = errors.New("record not found")

	// ErrDuplicateKey is returned when a unique constraint is violated
	ErrDuplicateKey = errors.New("duplicate key")

	// ErrForeignKeyViolation is returned when a foreign key constraint is violated
	ErrForeignKeyViolation = errors.New("foreign key violation")

	// ErrInvalidInput is returned when the input is invalid
	ErrInvalidInput = errors.New("invalid input")

	// ErrEventNotOpen is returned when trying to register for a non-open event
	ErrEventNotOpen = errors.New("event is not open for registration")

	// ErrHostCannotRegister is returned when the host tries to register for their own event
	ErrHostCannotRegister = errors.New("host cannot register for their own event")

	// ErrAlreadyRegistered is returned when a user is already registered for an event
	ErrAlreadyRegistered = errors.New("user is already registered")

	// ErrAlreadyCancelled is returned when a registration is already cancelled
	ErrAlreadyCancelled = errors.New("registration is already cancelled")

	// ErrNoWaitlist is returned when there's no one in the waitlist to promote
	ErrNoWaitlist = errors.New("no one in waitlist")
)
