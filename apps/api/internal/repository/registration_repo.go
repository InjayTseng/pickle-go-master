package repository

import (
	"context"

	"github.com/anthropics/pickle-go/apps/api/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// RegistrationRepository handles registration data access
type RegistrationRepository struct {
	db *sqlx.DB
}

// NewRegistrationRepository creates a new RegistrationRepository
func NewRegistrationRepository(db *sqlx.DB) *RegistrationRepository {
	return &RegistrationRepository{db: db}
}

// FindByID finds a registration by ID
func (r *RegistrationRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Registration, error) {
	var reg model.Registration
	query := `SELECT * FROM registrations WHERE id = $1`
	err := r.db.GetContext(ctx, &reg, query, id)
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

// FindByEventAndUser finds a registration by event and user
func (r *RegistrationRepository) FindByEventAndUser(ctx context.Context, eventID, userID uuid.UUID) (*model.Registration, error) {
	var reg model.Registration
	query := `SELECT * FROM registrations WHERE event_id = $1 AND user_id = $2`
	err := r.db.GetContext(ctx, &reg, query, eventID, userID)
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

// FindByEventID finds all registrations for an event
func (r *RegistrationRepository) FindByEventID(ctx context.Context, eventID uuid.UUID) ([]model.Registration, error) {
	var regs []model.Registration
	query := `
		SELECT * FROM registrations
		WHERE event_id = $1 AND status != 'cancelled'
		ORDER BY
			CASE status
				WHEN 'confirmed' THEN 0
				WHEN 'waitlist' THEN 1
			END,
			waitlist_position NULLS LAST,
			registered_at ASC`
	err := r.db.SelectContext(ctx, &regs, query, eventID)
	if err != nil {
		return nil, err
	}
	return regs, nil
}

// FindByUserID finds all registrations for a user
func (r *RegistrationRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.Registration, error) {
	var regs []model.Registration
	query := `
		SELECT * FROM registrations
		WHERE user_id = $1 AND status != 'cancelled'
		ORDER BY registered_at DESC`
	err := r.db.SelectContext(ctx, &regs, query, userID)
	if err != nil {
		return nil, err
	}
	return regs, nil
}

// CountConfirmed counts confirmed registrations for an event
func (r *RegistrationRepository) CountConfirmed(ctx context.Context, eventID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM registrations WHERE event_id = $1 AND status = 'confirmed'`
	err := r.db.GetContext(ctx, &count, query, eventID)
	return count, err
}

// GetNextWaitlistPosition gets the next waitlist position for an event
func (r *RegistrationRepository) GetNextWaitlistPosition(ctx context.Context, eventID uuid.UUID) (int, error) {
	var maxPos *int
	query := `SELECT MAX(waitlist_position) FROM registrations WHERE event_id = $1 AND status = 'waitlist'`
	err := r.db.GetContext(ctx, &maxPos, query, eventID)
	if err != nil || maxPos == nil {
		return 1, nil
	}
	return *maxPos + 1, nil
}

// Create creates a new registration
func (r *RegistrationRepository) Create(ctx context.Context, reg *model.Registration) error {
	query := `
		INSERT INTO registrations (id, event_id, user_id, status, waitlist_position, registered_at, confirmed_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), $6)
		RETURNING registered_at`
	var confirmedAt *interface{}
	if reg.Status == model.RegistrationConfirmed {
		// Use NOW() for confirmed_at
		query = `
			INSERT INTO registrations (id, event_id, user_id, status, waitlist_position, registered_at, confirmed_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			RETURNING registered_at, confirmed_at`
		return r.db.QueryRowxContext(ctx, query,
			reg.ID, reg.EventID, reg.UserID, reg.Status, reg.WaitlistPosition,
		).Scan(&reg.RegisteredAt, &reg.ConfirmedAt)
	}
	return r.db.QueryRowxContext(ctx, query,
		reg.ID, reg.EventID, reg.UserID, reg.Status, reg.WaitlistPosition, confirmedAt,
	).Scan(&reg.RegisteredAt)
}

// UpdateStatus updates the status of a registration
func (r *RegistrationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.RegistrationStatus) error {
	var query string
	switch status {
	case model.RegistrationConfirmed:
		query = `UPDATE registrations SET status = $2, confirmed_at = NOW(), waitlist_position = NULL WHERE id = $1`
	case model.RegistrationCancelled:
		query = `UPDATE registrations SET status = $2, cancelled_at = NOW() WHERE id = $1`
	default:
		query = `UPDATE registrations SET status = $2 WHERE id = $1`
	}
	_, err := r.db.ExecContext(ctx, query, id, status)
	return err
}

// GetFirstWaitlist gets the first person in the waitlist for an event
func (r *RegistrationRepository) GetFirstWaitlist(ctx context.Context, eventID uuid.UUID) (*model.Registration, error) {
	var reg model.Registration
	query := `
		SELECT * FROM registrations
		WHERE event_id = $1 AND status = 'waitlist'
		ORDER BY waitlist_position ASC
		LIMIT 1`
	err := r.db.GetContext(ctx, &reg, query, eventID)
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

// PromoteFromWaitlist promotes the first waitlist person to confirmed
func (r *RegistrationRepository) PromoteFromWaitlist(ctx context.Context, eventID uuid.UUID) (*model.Registration, error) {
	reg, err := r.GetFirstWaitlist(ctx, eventID)
	if err != nil {
		return nil, err
	}

	err = r.UpdateStatus(ctx, reg.ID, model.RegistrationConfirmed)
	if err != nil {
		return nil, err
	}

	return reg, nil
}
