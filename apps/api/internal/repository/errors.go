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
)
