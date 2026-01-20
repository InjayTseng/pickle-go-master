package repository

import (
	"context"

	"github.com/anthropics/pickle-go/apps/api/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// EventRepository handles event data access
type EventRepository struct {
	db *sqlx.DB
}

// NewEventRepository creates a new EventRepository
func NewEventRepository(db *sqlx.DB) *EventRepository {
	return &EventRepository{db: db}
}

// FindByID finds an event by ID
func (r *EventRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	var event model.Event
	query := `
		SELECT id, host_id, COALESCE(short_code, '') as short_code, title, description, event_date, start_time, end_time,
			   location_name, location_address,
			   ST_Y(location_point::geometry) as latitude,
			   ST_X(location_point::geometry) as longitude,
			   google_place_id, capacity, skill_level, fee, status, created_at, updated_at
		FROM events WHERE id = $1`
	err := r.db.GetContext(ctx, &event, query, id)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// EventFilter represents filter options for listing events
type EventFilter struct {
	Lat        float64
	Lng        float64
	Radius     int // in meters
	SkillLevel string
	Status     string
	Limit      int
	Offset     int
}

// FindNearby finds events near a given location
func (r *EventRepository) FindNearby(ctx context.Context, filter EventFilter) ([]model.EventSummary, error) {
	var events []model.EventSummary

	query := `
		SELECT
			e.id, e.host_id, COALESCE(e.short_code, '') as short_code, e.title, e.description, e.event_date, e.start_time, e.end_time,
			e.location_name, e.location_address,
			ST_Y(e.location_point::geometry) as latitude,
			ST_X(e.location_point::geometry) as longitude,
			e.google_place_id, e.capacity, e.skill_level, e.fee, e.status, e.created_at, e.updated_at,
			COALESCE(COUNT(CASE WHEN r.status = 'confirmed' THEN 1 END), 0) as confirmed_count,
			COALESCE(COUNT(CASE WHEN r.status = 'waitlist' THEN 1 END), 0) as waitlist_count
		FROM events e
		LEFT JOIN registrations r ON e.id = r.event_id
		WHERE ST_DWithin(e.location_point, ST_MakePoint($1, $2)::geography, $3)
		AND ($4 = '' OR e.skill_level = $4)
		AND ($5 = '' OR e.status = $5)
		AND e.event_date >= CURRENT_DATE
		GROUP BY e.id
		ORDER BY e.event_date ASC, e.start_time ASC
		LIMIT $6 OFFSET $7`

	err := r.db.SelectContext(ctx, &events, query,
		filter.Lng, filter.Lat, filter.Radius,
		filter.SkillLevel, filter.Status,
		filter.Limit, filter.Offset,
	)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// FindByHostID finds events hosted by a user
func (r *EventRepository) FindByHostID(ctx context.Context, hostID uuid.UUID) ([]model.Event, error) {
	var events []model.Event
	query := `
		SELECT id, host_id, COALESCE(short_code, '') as short_code, title, description, event_date, start_time, end_time,
			   location_name, location_address,
			   ST_Y(location_point::geometry) as latitude,
			   ST_X(location_point::geometry) as longitude,
			   google_place_id, capacity, skill_level, fee, status, created_at, updated_at
		FROM events
		WHERE host_id = $1
		ORDER BY event_date DESC, start_time DESC`
	err := r.db.SelectContext(ctx, &events, query, hostID)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// Create creates a new event
func (r *EventRepository) Create(ctx context.Context, event *model.Event) error {
	query := `
		INSERT INTO events (
			id, host_id, short_code, title, description, event_date, start_time, end_time,
			location_name, location_address, location_point, google_place_id,
			capacity, skill_level, fee, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			ST_SetSRID(ST_MakePoint($11, $12), 4326)::geography,
			$13, $14, $15, $16, 'open', NOW(), NOW()
		)
		RETURNING created_at, updated_at`
	return r.db.QueryRowxContext(ctx, query,
		event.ID, event.HostID, event.ShortCode, event.Title, event.Description,
		event.EventDate, event.StartTime, event.EndTime,
		event.LocationName, event.LocationAddress,
		event.Longitude, event.Latitude, event.GooglePlaceID,
		event.Capacity, event.SkillLevel, event.Fee,
	).StructScan(event)
}

// Update updates an existing event
func (r *EventRepository) Update(ctx context.Context, event *model.Event) error {
	query := `
		UPDATE events SET
			title = $2, description = $3, event_date = $4, start_time = $5, end_time = $6,
			capacity = $7, skill_level = $8, fee = $9, status = $10, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`
	return r.db.QueryRowxContext(ctx, query,
		event.ID, event.Title, event.Description, event.EventDate,
		event.StartTime, event.EndTime, event.Capacity,
		event.SkillLevel, event.Fee, event.Status,
	).Scan(&event.UpdatedAt)
}

// UpdateStatus updates the status of an event
func (r *EventRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.EventStatus) error {
	query := `UPDATE events SET status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status)
	return err
}

// Delete deletes an event by ID
func (r *EventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM events WHERE id = $1`
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

// FindByShortCode finds an event by its short code
func (r *EventRepository) FindByShortCode(ctx context.Context, shortCode string) (*model.Event, error) {
	var event model.Event
	query := `
		SELECT id, host_id, COALESCE(short_code, '') as short_code, title, description, event_date, start_time, end_time,
			   location_name, location_address,
			   ST_Y(location_point::geometry) as latitude,
			   ST_X(location_point::geometry) as longitude,
			   google_place_id, capacity, skill_level, fee, status, created_at, updated_at
		FROM events WHERE short_code = $1`
	err := r.db.GetContext(ctx, &event, query, shortCode)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// FindWithHost finds an event with host information
func (r *EventRepository) FindWithHost(ctx context.Context, id uuid.UUID) (*model.EventSummary, error) {
	var event model.EventSummary
	query := `
		SELECT
			e.id, e.host_id, COALESCE(e.short_code, '') as short_code, e.title, e.description, e.event_date, e.start_time, e.end_time,
			e.location_name, e.location_address,
			ST_Y(e.location_point::geometry) as latitude,
			ST_X(e.location_point::geometry) as longitude,
			e.google_place_id, e.capacity, e.skill_level, e.fee, e.status, e.created_at, e.updated_at,
			COALESCE(COUNT(CASE WHEN r.status = 'confirmed' THEN 1 END), 0) as confirmed_count,
			COALESCE(COUNT(CASE WHEN r.status = 'waitlist' THEN 1 END), 0) as waitlist_count
		FROM events e
		LEFT JOIN registrations r ON e.id = r.event_id AND r.status != 'cancelled'
		WHERE e.id = $1
		GROUP BY e.id`
	err := r.db.GetContext(ctx, &event, query, id)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// CountByHostID counts events hosted by a user
func (r *EventRepository) CountByHostID(ctx context.Context, hostID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM events WHERE host_id = $1`
	err := r.db.GetContext(ctx, &count, query, hostID)
	return count, err
}

// FindUpcoming finds upcoming events (future events that are open)
func (r *EventRepository) FindUpcoming(ctx context.Context, limit, offset int) ([]model.EventSummary, error) {
	var events []model.EventSummary
	query := `
		SELECT
			e.id, e.host_id, COALESCE(e.short_code, '') as short_code, e.title, e.description, e.event_date, e.start_time, e.end_time,
			e.location_name, e.location_address,
			ST_Y(e.location_point::geometry) as latitude,
			ST_X(e.location_point::geometry) as longitude,
			e.google_place_id, e.capacity, e.skill_level, e.fee, e.status, e.created_at, e.updated_at,
			COALESCE(COUNT(CASE WHEN r.status = 'confirmed' THEN 1 END), 0) as confirmed_count,
			COALESCE(COUNT(CASE WHEN r.status = 'waitlist' THEN 1 END), 0) as waitlist_count
		FROM events e
		LEFT JOIN registrations r ON e.id = r.event_id AND r.status != 'cancelled'
		WHERE e.event_date >= CURRENT_DATE
		AND e.status IN ('open', 'full')
		GROUP BY e.id
		ORDER BY e.event_date ASC, e.start_time ASC
		LIMIT $1 OFFSET $2`
	err := r.db.SelectContext(ctx, &events, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// UpdateLocation updates the location of an event
func (r *EventRepository) UpdateLocation(ctx context.Context, id uuid.UUID, locationName string, locationAddress *string, lat, lng float64, googlePlaceID *string) error {
	query := `
		UPDATE events SET
			location_name = $2,
			location_address = $3,
			location_point = ST_SetSRID(ST_MakePoint($4, $5), 4326)::geography,
			google_place_id = $6,
			updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, locationName, locationAddress, lng, lat, googlePlaceID)
	return err
}

// Exists checks if an event exists by ID
func (r *EventRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM events WHERE id = $1)`
	err := r.db.GetContext(ctx, &exists, query, id)
	return exists, err
}

// IsHost checks if a user is the host of an event
func (r *EventRepository) IsHost(ctx context.Context, eventID, userID uuid.UUID) (bool, error) {
	var isHost bool
	query := `SELECT EXISTS(SELECT 1 FROM events WHERE id = $1 AND host_id = $2)`
	err := r.db.GetContext(ctx, &isHost, query, eventID, userID)
	return isHost, err
}
