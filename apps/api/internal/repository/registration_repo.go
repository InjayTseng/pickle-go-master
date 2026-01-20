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

	// Reorder waitlist positions after promotion
	_, err = r.db.ExecContext(ctx, `
		UPDATE registrations
		SET waitlist_position = waitlist_position - 1
		WHERE event_id = $1 AND status = 'waitlist' AND waitlist_position > 1
	`, eventID)
	if err != nil {
		return nil, err
	}

	return reg, nil
}

// Delete deletes a registration by ID
func (r *RegistrationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM registrations WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// FindWithUsersByEventID finds all registrations for an event with user details
func (r *RegistrationRepository) FindWithUsersByEventID(ctx context.Context, eventID uuid.UUID) ([]model.RegistrationWithUser, error) {
	query := `
		SELECT
			r.id, r.event_id, r.user_id, r.status, r.waitlist_position,
			r.registered_at, r.confirmed_at, r.cancelled_at,
			u.id as "user.id", u.display_name as "user.display_name", u.avatar_url as "user.avatar_url"
		FROM registrations r
		JOIN users u ON r.user_id = u.id
		WHERE r.event_id = $1 AND r.status != 'cancelled'
		ORDER BY
			CASE r.status
				WHEN 'confirmed' THEN 0
				WHEN 'waitlist' THEN 1
			END,
			r.waitlist_position NULLS LAST,
			r.registered_at ASC`

	rows, err := r.db.QueryxContext(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.RegistrationWithUser
	for rows.Next() {
		var reg model.RegistrationWithUser
		var userID uuid.UUID
		var displayName string
		var avatarURL *string

		err := rows.Scan(
			&reg.ID, &reg.EventID, &reg.UserID, &reg.Status, &reg.WaitlistPosition,
			&reg.RegisteredAt, &reg.ConfirmedAt, &reg.CancelledAt,
			&userID, &displayName, &avatarURL,
		)
		if err != nil {
			return nil, err
		}

		reg.User = model.UserProfile{
			ID:          userID,
			DisplayName: displayName,
			AvatarURL:   avatarURL,
		}
		results = append(results, reg)
	}

	return results, nil
}

// CountWaitlist counts the number of waitlisted registrations for an event
func (r *RegistrationRepository) CountWaitlist(ctx context.Context, eventID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM registrations WHERE event_id = $1 AND status = 'waitlist'`
	err := r.db.GetContext(ctx, &count, query, eventID)
	return count, err
}

// Exists checks if a registration exists by ID
func (r *RegistrationRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM registrations WHERE id = $1)`
	err := r.db.GetContext(ctx, &exists, query, id)
	return exists, err
}

// HasUserRegistered checks if a user has already registered for an event (including cancelled)
func (r *RegistrationRepository) HasUserRegistered(ctx context.Context, eventID, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM registrations WHERE event_id = $1 AND user_id = $2 AND status != 'cancelled')`
	err := r.db.GetContext(ctx, &exists, query, eventID, userID)
	return exists, err
}

// FindEventsByUserID finds all events a user has registered for
func (r *RegistrationRepository) FindEventsByUserID(ctx context.Context, userID uuid.UUID, includeWaitlist bool) ([]uuid.UUID, error) {
	var eventIDs []uuid.UUID
	var query string
	if includeWaitlist {
		query = `SELECT event_id FROM registrations WHERE user_id = $1 AND status IN ('confirmed', 'waitlist') ORDER BY registered_at DESC`
	} else {
		query = `SELECT event_id FROM registrations WHERE user_id = $1 AND status = 'confirmed' ORDER BY registered_at DESC`
	}
	err := r.db.SelectContext(ctx, &eventIDs, query, userID)
	return eventIDs, err
}

// CancelAllByEventID cancels all registrations for an event
func (r *RegistrationRepository) CancelAllByEventID(ctx context.Context, eventID uuid.UUID) error {
	query := `UPDATE registrations SET status = 'cancelled', cancelled_at = NOW() WHERE event_id = $1 AND status != 'cancelled'`
	_, err := r.db.ExecContext(ctx, query, eventID)
	return err
}

// GetRegistrationStats gets registration statistics for an event
type RegistrationStats struct {
	ConfirmedCount int `db:"confirmed_count"`
	WaitlistCount  int `db:"waitlist_count"`
}

func (r *RegistrationRepository) GetRegistrationStats(ctx context.Context, eventID uuid.UUID) (*RegistrationStats, error) {
	var stats RegistrationStats
	query := `
		SELECT
			COUNT(CASE WHEN status = 'confirmed' THEN 1 END) as confirmed_count,
			COUNT(CASE WHEN status = 'waitlist' THEN 1 END) as waitlist_count
		FROM registrations
		WHERE event_id = $1 AND status != 'cancelled'`
	err := r.db.GetContext(ctx, &stats, query, eventID)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}
