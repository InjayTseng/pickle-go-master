package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/anthropics/pickle-go/apps/api/internal/model"
	"github.com/google/uuid"
)

// TestEventRepository_Create tests the Create method
func TestEventRepository_Create(t *testing.T) {
	tests := []struct {
		name      string
		event     *model.Event
		mockSetup func(mock sqlmock.Sqlmock, event *model.Event)
		wantErr   bool
	}{
		{
			name: "successful event creation",
			event: &model.Event{
				ID:           uuid.New(),
				HostID:       uuid.New(),
				ShortCode:    "abc123",
				Title:        strPtr("Test Event"),
				Description:  strPtr("Test Description"),
				EventDate:    time.Now().Add(24 * time.Hour),
				StartTime:    "19:00",
				EndTime:      strPtr("21:00"),
				LocationName: "Test Location",
				LocationAddress: strPtr("123 Test St"),
				Latitude:     25.0330,
				Longitude:    121.5654,
				GooglePlaceID: strPtr("place123"),
				Capacity:     8,
				SkillLevel:   model.SkillBeginner,
				Fee:          200,
			},
			mockSetup: func(mock sqlmock.Sqlmock, event *model.Event) {
				mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO events (
			id, host_id, short_code, title, description, event_date, start_time, end_time,
			location_name, location_address, location_point, google_place_id,
			capacity, skill_level, fee, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			ST_SetSRID(ST_MakePoint($11, $12), 4326)::geography,
			$13, $14, $15, $16, 'open', NOW(), NOW()
		)
		RETURNING created_at, updated_at`)).
					WithArgs(
						event.ID, event.HostID, event.ShortCode, event.Title, event.Description,
						event.EventDate, event.StartTime, event.EndTime,
						event.LocationName, event.LocationAddress,
						event.Longitude, event.Latitude, event.GooglePlaceID,
						event.Capacity, event.SkillLevel, event.Fee,
					).
					WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).
						AddRow(time.Now(), time.Now()))
			},
			wantErr: false,
		},
		{
			name: "event creation with minimal fields",
			event: &model.Event{
				ID:           uuid.New(),
				HostID:       uuid.New(),
				ShortCode:    "xyz789",
				EventDate:    time.Now().Add(48 * time.Hour),
				StartTime:    "10:00",
				LocationName: "Minimal Location",
				Latitude:     25.1000,
				Longitude:    121.6000,
				Capacity:     4,
				SkillLevel:   model.SkillAny,
				Fee:          0,
			},
			mockSetup: func(mock sqlmock.Sqlmock, event *model.Event) {
				mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO events (
			id, host_id, short_code, title, description, event_date, start_time, end_time,
			location_name, location_address, location_point, google_place_id,
			capacity, skill_level, fee, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			ST_SetSRID(ST_MakePoint($11, $12), 4326)::geography,
			$13, $14, $15, $16, 'open', NOW(), NOW()
		)
		RETURNING created_at, updated_at`)).
					WithArgs(
						event.ID, event.HostID, event.ShortCode, event.Title, event.Description,
						event.EventDate, event.StartTime, event.EndTime,
						event.LocationName, event.LocationAddress,
						event.Longitude, event.Latitude, event.GooglePlaceID,
						event.Capacity, event.SkillLevel, event.Fee,
					).
					WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).
						AddRow(time.Now(), time.Now()))
			},
			wantErr: false,
		},
		{
			name: "database error on creation",
			event: &model.Event{
				ID:           uuid.New(),
				HostID:       uuid.New(),
				ShortCode:    "err123",
				EventDate:    time.Now().Add(24 * time.Hour),
				StartTime:    "12:00",
				LocationName: "Error Location",
				Latitude:     25.0000,
				Longitude:    121.5000,
				Capacity:     6,
				SkillLevel:   model.SkillIntermediate,
				Fee:          100,
			},
			mockSetup: func(mock sqlmock.Sqlmock, event *model.Event) {
				mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO events (
			id, host_id, short_code, title, description, event_date, start_time, end_time,
			location_name, location_address, location_point, google_place_id,
			capacity, skill_level, fee, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			ST_SetSRID(ST_MakePoint($11, $12), 4326)::geography,
			$13, $14, $15, $16, 'open', NOW(), NOW()
		)
		RETURNING created_at, updated_at`)).
					WithArgs(
						event.ID, event.HostID, event.ShortCode, event.Title, event.Description,
						event.EventDate, event.StartTime, event.EndTime,
						event.LocationName, event.LocationAddress,
						event.Longitude, event.Latitude, event.GooglePlaceID,
						event.Capacity, event.SkillLevel, event.Fee,
					).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			defer db.Close()

			repo := NewEventRepository(db)
			tt.mockSetup(mock, tt.event)

			err := repo.Create(context.Background(), tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// TestEventRepository_FindByID tests the FindByID method
func TestEventRepository_FindByID(t *testing.T) {
	eventID := uuid.New()
	hostID := uuid.New()
	now := time.Now()
	eventDate := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour)

	tests := []struct {
		name      string
		eventID   uuid.UUID
		mockSetup func(mock sqlmock.Sqlmock)
		wantEvent *model.Event
		wantErr   bool
	}{
		{
			name:    "event found",
			eventID: eventID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "host_id", "short_code", "title", "description",
					"event_date", "start_time", "end_time",
					"location_name", "location_address", "latitude", "longitude",
					"google_place_id", "capacity", "skill_level", "fee", "status",
					"created_at", "updated_at",
				}).AddRow(
					eventID, hostID, "abc123", "Test Event", "Test Description",
					eventDate, "19:00", "21:00",
					"Test Location", "123 Test St", 25.0330, 121.5654,
					"place123", 8, "beginner", 200, "open",
					now, now,
				)
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, host_id, COALESCE(short_code, '') as short_code, title, description, event_date, start_time, end_time,
			   location_name, location_address,
			   ST_Y(location_point::geometry) as latitude,
			   ST_X(location_point::geometry) as longitude,
			   google_place_id, capacity, skill_level, fee, status, created_at, updated_at
		FROM events WHERE id = $1`)).
					WithArgs(eventID).
					WillReturnRows(rows)
			},
			wantEvent: &model.Event{
				ID:        eventID,
				HostID:    hostID,
				ShortCode: "abc123",
			},
			wantErr: false,
		},
		{
			name:    "event not found",
			eventID: uuid.New(),
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, host_id, COALESCE(short_code, '') as short_code, title, description, event_date, start_time, end_time,
			   location_name, location_address,
			   ST_Y(location_point::geometry) as latitude,
			   ST_X(location_point::geometry) as longitude,
			   google_place_id, capacity, skill_level, fee, status, created_at, updated_at
		FROM events WHERE id = $1`)).
					WillReturnError(sql.ErrNoRows)
			},
			wantEvent: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			defer db.Close()

			repo := NewEventRepository(db)
			tt.mockSetup(mock)

			event, err := repo.FindByID(context.Background(), tt.eventID)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantEvent != nil && event != nil {
				if event.ID != tt.wantEvent.ID {
					t.Errorf("FindByID() got ID = %v, want %v", event.ID, tt.wantEvent.ID)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// TestEventRepository_FindNearby tests the FindNearby method with geo-spatial queries
func TestEventRepository_FindNearby(t *testing.T) {
	eventID1 := uuid.New()
	eventID2 := uuid.New()
	hostID := uuid.New()
	now := time.Now()
	eventDate := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour)

	tests := []struct {
		name       string
		filter     EventFilter
		mockSetup  func(mock sqlmock.Sqlmock)
		wantCount  int
		wantErr    bool
	}{
		{
			name: "find events within radius",
			filter: EventFilter{
				Lat:        25.0330,
				Lng:        121.5654,
				Radius:     10000, // 10km
				SkillLevel: "",
				Status:     "",
				Limit:      20,
				Offset:     0,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "host_id", "short_code", "title", "description",
					"event_date", "start_time", "end_time",
					"location_name", "location_address", "latitude", "longitude",
					"google_place_id", "capacity", "skill_level", "fee", "status",
					"created_at", "updated_at", "confirmed_count", "waitlist_count",
				}).
					AddRow(
						eventID1, hostID, "abc123", "Event 1", "Description 1",
						eventDate, "19:00", "21:00",
						"Location 1", "Address 1", 25.0330, 121.5654,
						"place1", 8, "beginner", 200, "open",
						now, now, 3, 1,
					).
					AddRow(
						eventID2, hostID, "xyz789", "Event 2", "Description 2",
						eventDate, "10:00", "12:00",
						"Location 2", "Address 2", 25.0350, 121.5700,
						"place2", 4, "intermediate", 150, "open",
						now, now, 2, 0,
					)
				mock.ExpectQuery(regexp.QuoteMeta(`
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
		LIMIT $6 OFFSET $7`)).
					WithArgs(121.5654, 25.0330, 10000, "", "", 20, 0).
					WillReturnRows(rows)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "find events filtered by skill level",
			filter: EventFilter{
				Lat:        25.0330,
				Lng:        121.5654,
				Radius:     5000,
				SkillLevel: "beginner",
				Status:     "",
				Limit:      10,
				Offset:     0,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "host_id", "short_code", "title", "description",
					"event_date", "start_time", "end_time",
					"location_name", "location_address", "latitude", "longitude",
					"google_place_id", "capacity", "skill_level", "fee", "status",
					"created_at", "updated_at", "confirmed_count", "waitlist_count",
				}).
					AddRow(
						eventID1, hostID, "abc123", "Beginner Event", "For beginners",
						eventDate, "19:00", "21:00",
						"Location 1", "Address 1", 25.0330, 121.5654,
						"place1", 8, "beginner", 200, "open",
						now, now, 2, 0,
					)
				mock.ExpectQuery(regexp.QuoteMeta(`
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
		LIMIT $6 OFFSET $7`)).
					WithArgs(121.5654, 25.0330, 5000, "beginner", "", 10, 0).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "find events filtered by status",
			filter: EventFilter{
				Lat:        25.0330,
				Lng:        121.5654,
				Radius:     10000,
				SkillLevel: "",
				Status:     "open",
				Limit:      20,
				Offset:     0,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "host_id", "short_code", "title", "description",
					"event_date", "start_time", "end_time",
					"location_name", "location_address", "latitude", "longitude",
					"google_place_id", "capacity", "skill_level", "fee", "status",
					"created_at", "updated_at", "confirmed_count", "waitlist_count",
				})
				mock.ExpectQuery(regexp.QuoteMeta(`
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
		LIMIT $6 OFFSET $7`)).
					WithArgs(121.5654, 25.0330, 10000, "", "open", 20, 0).
					WillReturnRows(rows)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "pagination with offset",
			filter: EventFilter{
				Lat:        25.0330,
				Lng:        121.5654,
				Radius:     10000,
				SkillLevel: "",
				Status:     "",
				Limit:      10,
				Offset:     10,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "host_id", "short_code", "title", "description",
					"event_date", "start_time", "end_time",
					"location_name", "location_address", "latitude", "longitude",
					"google_place_id", "capacity", "skill_level", "fee", "status",
					"created_at", "updated_at", "confirmed_count", "waitlist_count",
				})
				mock.ExpectQuery(regexp.QuoteMeta(`
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
		LIMIT $6 OFFSET $7`)).
					WithArgs(121.5654, 25.0330, 10000, "", "", 10, 10).
					WillReturnRows(rows)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "database error",
			filter: EventFilter{
				Lat:    25.0330,
				Lng:    121.5654,
				Radius: 10000,
				Limit:  20,
				Offset: 0,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`
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
		LIMIT $6 OFFSET $7`)).
					WillReturnError(sql.ErrConnDone)
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			defer db.Close()

			repo := NewEventRepository(db)
			tt.mockSetup(mock)

			events, err := repo.FindNearby(context.Background(), tt.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindNearby() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(events) != tt.wantCount {
				t.Errorf("FindNearby() got %d events, want %d", len(events), tt.wantCount)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// TestEventRepository_Update tests the Update method
func TestEventRepository_Update(t *testing.T) {
	eventID := uuid.New()
	hostID := uuid.New()
	now := time.Now()

	tests := []struct {
		name      string
		event     *model.Event
		mockSetup func(mock sqlmock.Sqlmock, event *model.Event)
		wantErr   bool
	}{
		{
			name: "successful update",
			event: &model.Event{
				ID:          eventID,
				HostID:      hostID,
				Title:       strPtr("Updated Title"),
				Description: strPtr("Updated Description"),
				EventDate:   time.Now().Add(48 * time.Hour),
				StartTime:   "20:00",
				EndTime:     strPtr("22:00"),
				Capacity:    10,
				SkillLevel:  model.SkillIntermediate,
				Fee:         300,
				Status:      model.EventStatusOpen,
			},
			mockSetup: func(mock sqlmock.Sqlmock, event *model.Event) {
				mock.ExpectQuery(regexp.QuoteMeta(`
		UPDATE events SET
			title = $2, description = $3, event_date = $4, start_time = $5, end_time = $6,
			capacity = $7, skill_level = $8, fee = $9, status = $10, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`)).
					WithArgs(
						event.ID, event.Title, event.Description, event.EventDate,
						event.StartTime, event.EndTime, event.Capacity,
						event.SkillLevel, event.Fee, event.Status,
					).
					WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow(now))
			},
			wantErr: false,
		},
		{
			name: "update non-existent event",
			event: &model.Event{
				ID:         uuid.New(),
				HostID:     hostID,
				Title:      strPtr("Non-existent"),
				EventDate:  time.Now().Add(24 * time.Hour),
				StartTime:  "19:00",
				Capacity:   8,
				SkillLevel: model.SkillBeginner,
				Fee:        200,
				Status:     model.EventStatusOpen,
			},
			mockSetup: func(mock sqlmock.Sqlmock, event *model.Event) {
				mock.ExpectQuery(regexp.QuoteMeta(`
		UPDATE events SET
			title = $2, description = $3, event_date = $4, start_time = $5, end_time = $6,
			capacity = $7, skill_level = $8, fee = $9, status = $10, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`)).
					WithArgs(
						event.ID, event.Title, event.Description, event.EventDate,
						event.StartTime, event.EndTime, event.Capacity,
						event.SkillLevel, event.Fee, event.Status,
					).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			defer db.Close()

			repo := NewEventRepository(db)
			tt.mockSetup(mock, tt.event)

			err := repo.Update(context.Background(), tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// TestEventRepository_UpdateStatus tests the UpdateStatus method
func TestEventRepository_UpdateStatus(t *testing.T) {
	eventID := uuid.New()

	tests := []struct {
		name      string
		eventID   uuid.UUID
		status    model.EventStatus
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:    "update to full status",
			eventID: eventID,
			status:  model.EventStatusFull,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE events SET status = $2, updated_at = NOW() WHERE id = $1`)).
					WithArgs(eventID, model.EventStatusFull).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:    "update to cancelled status",
			eventID: eventID,
			status:  model.EventStatusCancelled,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE events SET status = $2, updated_at = NOW() WHERE id = $1`)).
					WithArgs(eventID, model.EventStatusCancelled).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:    "update to completed status",
			eventID: eventID,
			status:  model.EventStatusCompleted,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE events SET status = $2, updated_at = NOW() WHERE id = $1`)).
					WithArgs(eventID, model.EventStatusCompleted).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:    "update to open status",
			eventID: eventID,
			status:  model.EventStatusOpen,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE events SET status = $2, updated_at = NOW() WHERE id = $1`)).
					WithArgs(eventID, model.EventStatusOpen).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:    "database error",
			eventID: eventID,
			status:  model.EventStatusFull,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE events SET status = $2, updated_at = NOW() WHERE id = $1`)).
					WithArgs(eventID, model.EventStatusFull).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			defer db.Close()

			repo := NewEventRepository(db)
			tt.mockSetup(mock)

			err := repo.UpdateStatus(context.Background(), tt.eventID, tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// TestEventRepository_Delete tests the Delete method
func TestEventRepository_Delete(t *testing.T) {
	eventID := uuid.New()

	tests := []struct {
		name      string
		eventID   uuid.UUID
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
		errType   error
	}{
		{
			name:    "successful deletion",
			eventID: eventID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM events WHERE id = $1`)).
					WithArgs(eventID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:    "event not found",
			eventID: uuid.New(),
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM events WHERE id = $1`)).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
			errType: ErrNotFound,
		},
		{
			name:    "database error",
			eventID: eventID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM events WHERE id = $1`)).
					WithArgs(eventID).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			defer db.Close()

			repo := NewEventRepository(db)
			tt.mockSetup(mock)

			err := repo.Delete(context.Background(), tt.eventID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.errType != nil && err != tt.errType {
				t.Errorf("Delete() error = %v, want %v", err, tt.errType)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// TestEventRepository_FindByShortCode tests the FindByShortCode method
func TestEventRepository_FindByShortCode(t *testing.T) {
	eventID := uuid.New()
	hostID := uuid.New()
	now := time.Now()
	eventDate := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour)

	tests := []struct {
		name      string
		shortCode string
		mockSetup func(mock sqlmock.Sqlmock)
		wantEvent bool
		wantErr   bool
	}{
		{
			name:      "event found by short code",
			shortCode: "abc123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "host_id", "short_code", "title", "description",
					"event_date", "start_time", "end_time",
					"location_name", "location_address", "latitude", "longitude",
					"google_place_id", "capacity", "skill_level", "fee", "status",
					"created_at", "updated_at",
				}).AddRow(
					eventID, hostID, "abc123", "Test Event", "Test Description",
					eventDate, "19:00", "21:00",
					"Test Location", "123 Test St", 25.0330, 121.5654,
					"place123", 8, "beginner", 200, "open",
					now, now,
				)
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, host_id, COALESCE(short_code, '') as short_code, title, description, event_date, start_time, end_time,
			   location_name, location_address,
			   ST_Y(location_point::geometry) as latitude,
			   ST_X(location_point::geometry) as longitude,
			   google_place_id, capacity, skill_level, fee, status, created_at, updated_at
		FROM events WHERE short_code = $1`)).
					WithArgs("abc123").
					WillReturnRows(rows)
			},
			wantEvent: true,
			wantErr:   false,
		},
		{
			name:      "event not found by short code",
			shortCode: "nonexistent",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, host_id, COALESCE(short_code, '') as short_code, title, description, event_date, start_time, end_time,
			   location_name, location_address,
			   ST_Y(location_point::geometry) as latitude,
			   ST_X(location_point::geometry) as longitude,
			   google_place_id, capacity, skill_level, fee, status, created_at, updated_at
		FROM events WHERE short_code = $1`)).
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			wantEvent: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			defer db.Close()

			repo := NewEventRepository(db)
			tt.mockSetup(mock)

			event, err := repo.FindByShortCode(context.Background(), tt.shortCode)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByShortCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantEvent && event == nil {
				t.Error("FindByShortCode() expected event, got nil")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// TestEventRepository_IsHost tests the IsHost method
func TestEventRepository_IsHost(t *testing.T) {
	eventID := uuid.New()
	hostID := uuid.New()
	otherUserID := uuid.New()

	tests := []struct {
		name      string
		eventID   uuid.UUID
		userID    uuid.UUID
		mockSetup func(mock sqlmock.Sqlmock)
		wantHost  bool
		wantErr   bool
	}{
		{
			name:    "user is host",
			eventID: eventID,
			userID:  hostID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM events WHERE id = $1 AND host_id = $2)`)).
					WithArgs(eventID, hostID).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			wantHost: true,
			wantErr:  false,
		},
		{
			name:    "user is not host",
			eventID: eventID,
			userID:  otherUserID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM events WHERE id = $1 AND host_id = $2)`)).
					WithArgs(eventID, otherUserID).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			},
			wantHost: false,
			wantErr:  false,
		},
		{
			name:    "database error",
			eventID: eventID,
			userID:  hostID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM events WHERE id = $1 AND host_id = $2)`)).
					WithArgs(eventID, hostID).
					WillReturnError(sql.ErrConnDone)
			},
			wantHost: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			defer db.Close()

			repo := NewEventRepository(db)
			tt.mockSetup(mock)

			isHost, err := repo.IsHost(context.Background(), tt.eventID, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && isHost != tt.wantHost {
				t.Errorf("IsHost() = %v, want %v", isHost, tt.wantHost)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// TestEventRepository_Exists tests the Exists method
func TestEventRepository_Exists(t *testing.T) {
	eventID := uuid.New()

	tests := []struct {
		name       string
		eventID    uuid.UUID
		mockSetup  func(mock sqlmock.Sqlmock)
		wantExists bool
		wantErr    bool
	}{
		{
			name:    "event exists",
			eventID: eventID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM events WHERE id = $1)`)).
					WithArgs(eventID).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			wantExists: true,
			wantErr:    false,
		},
		{
			name:    "event does not exist",
			eventID: uuid.New(),
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM events WHERE id = $1)`)).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			},
			wantExists: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			defer db.Close()

			repo := NewEventRepository(db)
			tt.mockSetup(mock)

			exists, err := repo.Exists(context.Background(), tt.eventID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && exists != tt.wantExists {
				t.Errorf("Exists() = %v, want %v", exists, tt.wantExists)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// TestEventRepository_FindByHostID tests the FindByHostID method
func TestEventRepository_FindByHostID(t *testing.T) {
	eventID1 := uuid.New()
	eventID2 := uuid.New()
	hostID := uuid.New()
	now := time.Now()
	eventDate := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour)

	tests := []struct {
		name      string
		hostID    uuid.UUID
		mockSetup func(mock sqlmock.Sqlmock)
		wantCount int
		wantErr   bool
	}{
		{
			name:   "find events by host",
			hostID: hostID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "host_id", "short_code", "title", "description",
					"event_date", "start_time", "end_time",
					"location_name", "location_address", "latitude", "longitude",
					"google_place_id", "capacity", "skill_level", "fee", "status",
					"created_at", "updated_at",
				}).
					AddRow(
						eventID1, hostID, "abc123", "Event 1", "Description 1",
						eventDate, "19:00", "21:00",
						"Location 1", "Address 1", 25.0330, 121.5654,
						"place1", 8, "beginner", 200, "open",
						now, now,
					).
					AddRow(
						eventID2, hostID, "xyz789", "Event 2", "Description 2",
						eventDate.Add(24*time.Hour), "10:00", "12:00",
						"Location 2", "Address 2", 25.0350, 121.5700,
						"place2", 4, "intermediate", 150, "open",
						now, now,
					)
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, host_id, COALESCE(short_code, '') as short_code, title, description, event_date, start_time, end_time,
			   location_name, location_address,
			   ST_Y(location_point::geometry) as latitude,
			   ST_X(location_point::geometry) as longitude,
			   google_place_id, capacity, skill_level, fee, status, created_at, updated_at
		FROM events
		WHERE host_id = $1
		ORDER BY event_date DESC, start_time DESC`)).
					WithArgs(hostID).
					WillReturnRows(rows)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:   "no events for host",
			hostID: uuid.New(),
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "host_id", "short_code", "title", "description",
					"event_date", "start_time", "end_time",
					"location_name", "location_address", "latitude", "longitude",
					"google_place_id", "capacity", "skill_level", "fee", "status",
					"created_at", "updated_at",
				})
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, host_id, COALESCE(short_code, '') as short_code, title, description, event_date, start_time, end_time,
			   location_name, location_address,
			   ST_Y(location_point::geometry) as latitude,
			   ST_X(location_point::geometry) as longitude,
			   google_place_id, capacity, skill_level, fee, status, created_at, updated_at
		FROM events
		WHERE host_id = $1
		ORDER BY event_date DESC, start_time DESC`)).
					WillReturnRows(rows)
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := newMockDB(t)
			defer db.Close()

			repo := NewEventRepository(db)
			tt.mockSetup(mock)

			events, err := repo.FindByHostID(context.Background(), tt.hostID)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByHostID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(events) != tt.wantCount {
				t.Errorf("FindByHostID() got %d events, want %d", len(events), tt.wantCount)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}
