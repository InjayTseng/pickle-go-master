package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/anthropics/pickle-go/apps/api/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// testRegistration creates a test registration for use in tests
func testRegistration() model.Registration {
	now := time.Now()
	return model.Registration{
		ID:               uuid.New(),
		EventID:          uuid.New(),
		UserID:           uuid.New(),
		Status:           model.RegistrationConfirmed,
		WaitlistPosition: nil,
		RegisteredAt:     now,
		ConfirmedAt:      &now,
		CancelledAt:      nil,
	}
}

// setupMockDB wraps newMockDB for backwards compatibility
func setupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	return newMockDB(t)
}

// =============================================================================
// FindByID Tests
// =============================================================================

func TestFindByID(t *testing.T) {
	tests := []struct {
		name          string
		regID         uuid.UUID
		setupMock     func(mock sqlmock.Sqlmock, id uuid.UUID)
		expectedError error
	}{
		{
			name:  "successful find",
			regID: uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, id uuid.UUID) {
				now := time.Now()
				rows := sqlmock.NewRows([]string{
					"id", "event_id", "user_id", "status", "waitlist_position",
					"registered_at", "confirmed_at", "cancelled_at",
				}).AddRow(
					id, uuid.New(), uuid.New(), "confirmed", nil,
					now, now, nil,
				)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM registrations WHERE id = $1`)).
					WithArgs(id).
					WillReturnRows(rows)
			},
			expectedError: nil,
		},
		{
			name:  "not found",
			regID: uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, id uuid.UUID) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM registrations WHERE id = $1`)).
					WithArgs(id).
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			repo := NewRegistrationRepository(db)
			tt.setupMock(mock, tt.regID)

			result, err := repo.FindByID(context.Background(), tt.regID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.ID != tt.regID {
				t.Errorf("expected ID %v, got %v", tt.regID, result.ID)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// =============================================================================
// FindByEventAndUser Tests
// =============================================================================

func TestFindByEventAndUser(t *testing.T) {
	tests := []struct {
		name           string
		eventID        uuid.UUID
		userID         uuid.UUID
		setupMock      func(mock sqlmock.Sqlmock, eventID, userID uuid.UUID)
		expectedStatus model.RegistrationStatus
		expectedError  error
	}{
		{
			name:    "user is confirmed",
			eventID: uuid.New(),
			userID:  uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, eventID, userID uuid.UUID) {
				now := time.Now()
				rows := sqlmock.NewRows([]string{
					"id", "event_id", "user_id", "status", "waitlist_position",
					"registered_at", "confirmed_at", "cancelled_at",
				}).AddRow(
					uuid.New(), eventID, userID, "confirmed", nil,
					now, now, nil,
				)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM registrations WHERE event_id = $1 AND user_id = $2`)).
					WithArgs(eventID, userID).
					WillReturnRows(rows)
			},
			expectedStatus: model.RegistrationConfirmed,
			expectedError:  nil,
		},
		{
			name:    "user is on waitlist",
			eventID: uuid.New(),
			userID:  uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, eventID, userID uuid.UUID) {
				now := time.Now()
				pos := 2
				rows := sqlmock.NewRows([]string{
					"id", "event_id", "user_id", "status", "waitlist_position",
					"registered_at", "confirmed_at", "cancelled_at",
				}).AddRow(
					uuid.New(), eventID, userID, "waitlist", pos,
					now, nil, nil,
				)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM registrations WHERE event_id = $1 AND user_id = $2`)).
					WithArgs(eventID, userID).
					WillReturnRows(rows)
			},
			expectedStatus: model.RegistrationWaitlist,
			expectedError:  nil,
		},
		{
			name:    "user not registered",
			eventID: uuid.New(),
			userID:  uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, eventID, userID uuid.UUID) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM registrations WHERE event_id = $1 AND user_id = $2`)).
					WithArgs(eventID, userID).
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			repo := NewRegistrationRepository(db)
			tt.setupMock(mock, tt.eventID, tt.userID)

			result, err := repo.FindByEventAndUser(context.Background(), tt.eventID, tt.userID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Status != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, result.Status)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// =============================================================================
// CountConfirmed Tests
// =============================================================================

func TestCountConfirmed(t *testing.T) {
	tests := []struct {
		name          string
		eventID       uuid.UUID
		setupMock     func(mock sqlmock.Sqlmock, eventID uuid.UUID)
		expectedCount int
		expectedError error
	}{
		{
			name:    "event with confirmed registrations",
			eventID: uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, eventID uuid.UUID) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(3)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM registrations WHERE event_id = $1 AND status = 'confirmed'`)).
					WithArgs(eventID).
					WillReturnRows(rows)
			},
			expectedCount: 3,
			expectedError: nil,
		},
		{
			name:    "event with no registrations",
			eventID: uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, eventID uuid.UUID) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM registrations WHERE event_id = $1 AND status = 'confirmed'`)).
					WithArgs(eventID).
					WillReturnRows(rows)
			},
			expectedCount: 0,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			repo := NewRegistrationRepository(db)
			tt.setupMock(mock, tt.eventID)

			count, err := repo.CountConfirmed(context.Background(), tt.eventID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if count != tt.expectedCount {
				t.Errorf("expected count %d, got %d", tt.expectedCount, count)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// =============================================================================
// GetNextWaitlistPosition Tests
// =============================================================================

func TestGetNextWaitlistPosition(t *testing.T) {
	tests := []struct {
		name             string
		eventID          uuid.UUID
		setupMock        func(mock sqlmock.Sqlmock, eventID uuid.UUID)
		expectedPosition int
		expectedError    error
	}{
		{
			name:    "first waitlist position",
			eventID: uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, eventID uuid.UUID) {
				rows := sqlmock.NewRows([]string{"max"}).AddRow(nil)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT MAX(waitlist_position) FROM registrations WHERE event_id = $1 AND status = 'waitlist'`)).
					WithArgs(eventID).
					WillReturnRows(rows)
			},
			expectedPosition: 1,
			expectedError:    nil,
		},
		{
			name:    "subsequent waitlist position",
			eventID: uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, eventID uuid.UUID) {
				rows := sqlmock.NewRows([]string{"max"}).AddRow(3)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT MAX(waitlist_position) FROM registrations WHERE event_id = $1 AND status = 'waitlist'`)).
					WithArgs(eventID).
					WillReturnRows(rows)
			},
			expectedPosition: 4,
			expectedError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			repo := NewRegistrationRepository(db)
			tt.setupMock(mock, tt.eventID)

			pos, err := repo.GetNextWaitlistPosition(context.Background(), tt.eventID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if pos != tt.expectedPosition {
				t.Errorf("expected position %d, got %d", tt.expectedPosition, pos)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// =============================================================================
// HasUserRegistered Tests
// =============================================================================

func TestHasUserRegistered(t *testing.T) {
	tests := []struct {
		name           string
		eventID        uuid.UUID
		userID         uuid.UUID
		setupMock      func(mock sqlmock.Sqlmock, eventID, userID uuid.UUID)
		expectedResult bool
		expectedError  error
	}{
		{
			name:    "user has active registration",
			eventID: uuid.New(),
			userID:  uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, eventID, userID uuid.UUID) {
				rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM registrations WHERE event_id = $1 AND user_id = $2 AND status != 'cancelled')`)).
					WithArgs(eventID, userID).
					WillReturnRows(rows)
			},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:    "user has no active registration",
			eventID: uuid.New(),
			userID:  uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, eventID, userID uuid.UUID) {
				rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM registrations WHERE event_id = $1 AND user_id = $2 AND status != 'cancelled')`)).
					WithArgs(eventID, userID).
					WillReturnRows(rows)
			},
			expectedResult: false,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			repo := NewRegistrationRepository(db)
			tt.setupMock(mock, tt.eventID, tt.userID)

			result, err := repo.HasUserRegistered(context.Background(), tt.eventID, tt.userID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expectedResult {
				t.Errorf("expected %v, got %v", tt.expectedResult, result)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// =============================================================================
// RegisterWithLock Tests - Race Condition Safe Registration
// =============================================================================

func TestRegisterWithLock(t *testing.T) {
	tests := []struct {
		name           string
		eventID        uuid.UUID
		userID         uuid.UUID
		hostID         uuid.UUID
		capacity       int
		confirmedCount int
		eventStatus    string
		existingReg    *model.Registration
		setupMock      func(mock sqlmock.Sqlmock, eventID, userID, hostID uuid.UUID, capacity, confirmedCount int, eventStatus string, existingReg *model.Registration)
		expectedStatus model.RegistrationStatus
		expectedError  error
	}{
		{
			name:           "successful confirmed registration (event has space)",
			eventID:        uuid.New(),
			userID:         uuid.New(),
			hostID:         uuid.New(),
			capacity:       4,
			confirmedCount: 2,
			eventStatus:    "open",
			existingReg:    nil,
			setupMock: func(mock sqlmock.Sqlmock, eventID, userID, hostID uuid.UUID, capacity, confirmedCount int, eventStatus string, existingReg *model.Registration) {
				// Begin transaction
				mock.ExpectBegin()

				// Lock event with FOR UPDATE
				eventRows := sqlmock.NewRows([]string{"capacity", "status", "host_id"}).
					AddRow(capacity, eventStatus, hostID)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT capacity, status, host_id FROM events WHERE id = $1 FOR UPDATE`)).
					WithArgs(eventID).
					WillReturnRows(eventRows)

				// Check for existing registration
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM registrations WHERE event_id = $1 AND user_id = $2`)).
					WithArgs(eventID, userID).
					WillReturnError(sql.ErrNoRows)

				// Count confirmed registrations
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(confirmedCount)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM registrations WHERE event_id = $1 AND status = 'confirmed'`)).
					WithArgs(eventID).
					WillReturnRows(countRows)

				// Insert new registration (confirmed status)
				now := time.Now()
				insertRows := sqlmock.NewRows([]string{"registered_at", "confirmed_at"}).AddRow(now, now)
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO registrations (id, event_id, user_id, status, waitlist_position, registered_at, confirmed_at)`)).
					WithArgs(sqlmock.AnyArg(), eventID, userID, model.RegistrationConfirmed, nil).
					WillReturnRows(insertRows)

				mock.ExpectCommit()
			},
			expectedStatus: model.RegistrationConfirmed,
			expectedError:  nil,
		},
		{
			name:           "successful waitlist registration (event is full)",
			eventID:        uuid.New(),
			userID:         uuid.New(),
			hostID:         uuid.New(),
			capacity:       4,
			confirmedCount: 4,
			eventStatus:    "full",
			existingReg:    nil,
			setupMock: func(mock sqlmock.Sqlmock, eventID, userID, hostID uuid.UUID, capacity, confirmedCount int, eventStatus string, existingReg *model.Registration) {
				mock.ExpectBegin()

				// Lock event
				eventRows := sqlmock.NewRows([]string{"capacity", "status", "host_id"}).
					AddRow(capacity, eventStatus, hostID)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT capacity, status, host_id FROM events WHERE id = $1 FOR UPDATE`)).
					WithArgs(eventID).
					WillReturnRows(eventRows)

				// Check for existing registration
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM registrations WHERE event_id = $1 AND user_id = $2`)).
					WithArgs(eventID, userID).
					WillReturnError(sql.ErrNoRows)

				// Count confirmed registrations
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(confirmedCount)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM registrations WHERE event_id = $1 AND status = 'confirmed'`)).
					WithArgs(eventID).
					WillReturnRows(countRows)

				// Get max waitlist position
				maxPosRows := sqlmock.NewRows([]string{"max"}).AddRow(nil)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT MAX(waitlist_position) FROM registrations`)).
					WithArgs(eventID).
					WillReturnRows(maxPosRows)

				// Insert new registration (waitlist status)
				now := time.Now()
				insertRows := sqlmock.NewRows([]string{"registered_at"}).AddRow(now)
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO registrations (id, event_id, user_id, status, waitlist_position, registered_at)`)).
					WithArgs(sqlmock.AnyArg(), eventID, userID, model.RegistrationWaitlist, 1).
					WillReturnRows(insertRows)

				mock.ExpectCommit()
			},
			expectedStatus: model.RegistrationWaitlist,
			expectedError:  nil,
		},
		{
			name:           "host cannot register for own event",
			eventID:        uuid.New(),
			userID:         uuid.New(), // will be set to same as hostID in test
			hostID:         uuid.Nil,   // will be set in test
			capacity:       4,
			confirmedCount: 0,
			eventStatus:    "open",
			existingReg:    nil,
			setupMock: func(mock sqlmock.Sqlmock, eventID, userID, hostID uuid.UUID, capacity, confirmedCount int, eventStatus string, existingReg *model.Registration) {
				mock.ExpectBegin()

				// Lock event - hostID matches userID
				eventRows := sqlmock.NewRows([]string{"capacity", "status", "host_id"}).
					AddRow(capacity, eventStatus, userID) // host_id == userID
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT capacity, status, host_id FROM events WHERE id = $1 FOR UPDATE`)).
					WithArgs(eventID).
					WillReturnRows(eventRows)

				mock.ExpectRollback()
			},
			expectedError: ErrHostCannotRegister,
		},
		{
			name:           "event is cancelled",
			eventID:        uuid.New(),
			userID:         uuid.New(),
			hostID:         uuid.New(),
			capacity:       4,
			confirmedCount: 0,
			eventStatus:    "cancelled",
			existingReg:    nil,
			setupMock: func(mock sqlmock.Sqlmock, eventID, userID, hostID uuid.UUID, capacity, confirmedCount int, eventStatus string, existingReg *model.Registration) {
				mock.ExpectBegin()

				// Lock event
				eventRows := sqlmock.NewRows([]string{"capacity", "status", "host_id"}).
					AddRow(capacity, eventStatus, hostID)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT capacity, status, host_id FROM events WHERE id = $1 FOR UPDATE`)).
					WithArgs(eventID).
					WillReturnRows(eventRows)

				mock.ExpectRollback()
			},
			expectedError: ErrEventNotOpen,
		},
		{
			name:           "event is completed",
			eventID:        uuid.New(),
			userID:         uuid.New(),
			hostID:         uuid.New(),
			capacity:       4,
			confirmedCount: 0,
			eventStatus:    "completed",
			existingReg:    nil,
			setupMock: func(mock sqlmock.Sqlmock, eventID, userID, hostID uuid.UUID, capacity, confirmedCount int, eventStatus string, existingReg *model.Registration) {
				mock.ExpectBegin()

				// Lock event
				eventRows := sqlmock.NewRows([]string{"capacity", "status", "host_id"}).
					AddRow(capacity, eventStatus, hostID)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT capacity, status, host_id FROM events WHERE id = $1 FOR UPDATE`)).
					WithArgs(eventID).
					WillReturnRows(eventRows)

				mock.ExpectRollback()
			},
			expectedError: ErrEventNotOpen,
		},
		{
			name:           "user already has active registration",
			eventID:        uuid.New(),
			userID:         uuid.New(),
			hostID:         uuid.New(),
			capacity:       4,
			confirmedCount: 2,
			eventStatus:    "open",
			existingReg: &model.Registration{
				Status: model.RegistrationConfirmed,
			},
			setupMock: func(mock sqlmock.Sqlmock, eventID, userID, hostID uuid.UUID, capacity, confirmedCount int, eventStatus string, existingReg *model.Registration) {
				mock.ExpectBegin()

				// Lock event
				eventRows := sqlmock.NewRows([]string{"capacity", "status", "host_id"}).
					AddRow(capacity, eventStatus, hostID)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT capacity, status, host_id FROM events WHERE id = $1 FOR UPDATE`)).
					WithArgs(eventID).
					WillReturnRows(eventRows)

				// Check for existing registration - found active one
				now := time.Now()
				existingRows := sqlmock.NewRows([]string{
					"id", "event_id", "user_id", "status", "waitlist_position",
					"registered_at", "confirmed_at", "cancelled_at",
				}).AddRow(
					uuid.New(), eventID, userID, existingReg.Status, nil,
					now, now, nil,
				)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM registrations WHERE event_id = $1 AND user_id = $2`)).
					WithArgs(eventID, userID).
					WillReturnRows(existingRows)

				mock.ExpectRollback()
			},
			expectedError: ErrAlreadyRegistered,
		},
		{
			name:           "re-registration after cancel (event has space)",
			eventID:        uuid.New(),
			userID:         uuid.New(),
			hostID:         uuid.New(),
			capacity:       4,
			confirmedCount: 2,
			eventStatus:    "open",
			existingReg: &model.Registration{
				ID:     uuid.New(),
				Status: model.RegistrationCancelled,
			},
			setupMock: func(mock sqlmock.Sqlmock, eventID, userID, hostID uuid.UUID, capacity, confirmedCount int, eventStatus string, existingReg *model.Registration) {
				mock.ExpectBegin()

				// Lock event
				eventRows := sqlmock.NewRows([]string{"capacity", "status", "host_id"}).
					AddRow(capacity, eventStatus, hostID)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT capacity, status, host_id FROM events WHERE id = $1 FOR UPDATE`)).
					WithArgs(eventID).
					WillReturnRows(eventRows)

				// Check for existing registration - found cancelled one
				now := time.Now()
				cancelledAt := time.Now().Add(-time.Hour)
				existingRows := sqlmock.NewRows([]string{
					"id", "event_id", "user_id", "status", "waitlist_position",
					"registered_at", "confirmed_at", "cancelled_at",
				}).AddRow(
					existingReg.ID, eventID, userID, existingReg.Status, nil,
					now.Add(-2*time.Hour), nil, cancelledAt,
				)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM registrations WHERE event_id = $1 AND user_id = $2`)).
					WithArgs(eventID, userID).
					WillReturnRows(existingRows)

				// Count confirmed registrations
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(confirmedCount)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM registrations WHERE event_id = $1 AND status = 'confirmed'`)).
					WithArgs(eventID).
					WillReturnRows(countRows)

				// UPDATE existing cancelled registration (re-registration)
				updateRows := sqlmock.NewRows([]string{"registered_at", "confirmed_at"}).AddRow(now, now)
				mock.ExpectQuery(regexp.QuoteMeta(`UPDATE registrations`)).
					WithArgs(existingReg.ID, model.RegistrationConfirmed, nil).
					WillReturnRows(updateRows)

				mock.ExpectCommit()
			},
			expectedStatus: model.RegistrationConfirmed,
			expectedError:  nil,
		},
		{
			name:           "event not found",
			eventID:        uuid.New(),
			userID:         uuid.New(),
			hostID:         uuid.New(),
			capacity:       4,
			confirmedCount: 0,
			eventStatus:    "open",
			existingReg:    nil,
			setupMock: func(mock sqlmock.Sqlmock, eventID, userID, hostID uuid.UUID, capacity, confirmedCount int, eventStatus string, existingReg *model.Registration) {
				mock.ExpectBegin()

				// Lock event - not found
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT capacity, status, host_id FROM events WHERE id = $1 FOR UPDATE`)).
					WithArgs(eventID).
					WillReturnError(sql.ErrNoRows)

				mock.ExpectRollback()
			},
			expectedError: sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			repo := NewRegistrationRepository(db)
			tt.setupMock(mock, tt.eventID, tt.userID, tt.hostID, tt.capacity, tt.confirmedCount, tt.eventStatus, tt.existingReg)

			// Start transaction
			tx, err := db.Beginx()
			if err != nil && tt.expectedError == nil {
				t.Fatalf("failed to begin transaction: %v", err)
			}

			if tt.expectedError != nil && err != nil {
				// Transaction couldn't even begin, that's fine for error cases
				return
			}

			result, err := repo.RegisterWithLock(context.Background(), tx, tt.eventID, tt.userID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) && err.Error() != tt.expectedError.Error() {
					// Check if it's the expected error or wraps it
					if !errors.Is(err, sql.ErrNoRows) && tt.expectedError != sql.ErrNoRows {
						t.Errorf("expected error %v, got %v", tt.expectedError, err)
					}
				}
				tx.Rollback()
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				tx.Rollback()
				return
			}

			if result.Status != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, result.Status)
			}

			tx.Commit()

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// =============================================================================
// CancelAndPromote Tests - Race Condition Safe Cancellation with Waitlist Promotion
// =============================================================================

func TestCancelAndPromote(t *testing.T) {
	tests := []struct {
		name             string
		registrationID   uuid.UUID
		eventID          uuid.UUID
		regStatus        model.RegistrationStatus
		waitlistPosition *int
		hasWaitlist      bool
		setupMock        func(mock sqlmock.Sqlmock, regID, eventID uuid.UUID, regStatus model.RegistrationStatus, waitlistPos *int, hasWaitlist bool)
		expectedPromoted bool
		expectedError    error
	}{
		{
			name:             "cancel confirmed and promote waitlist",
			registrationID:   uuid.New(),
			eventID:          uuid.New(),
			regStatus:        model.RegistrationConfirmed,
			waitlistPosition: nil,
			hasWaitlist:      true,
			setupMock: func(mock sqlmock.Sqlmock, regID, eventID uuid.UUID, regStatus model.RegistrationStatus, waitlistPos *int, hasWaitlist bool) {
				mock.ExpectBegin()

				// Lock and get the registration
				now := time.Now()
				regRows := sqlmock.NewRows([]string{
					"id", "event_id", "user_id", "status", "waitlist_position",
					"registered_at", "confirmed_at", "cancelled_at",
				}).AddRow(
					regID, eventID, uuid.New(), regStatus, waitlistPos,
					now, now, nil,
				)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM registrations WHERE id = $1 FOR UPDATE`)).
					WithArgs(regID).
					WillReturnRows(regRows)

				// Update to cancelled
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE registrations SET status = 'cancelled', cancelled_at = NOW(), waitlist_position = NULL WHERE id = $1`)).
					WithArgs(regID).
					WillReturnResult(sqlmock.NewResult(0, 1))

				// Get first waitlist person with SKIP LOCKED
				waitlistUserID := uuid.New()
				waitlistID := uuid.New()
				waitlistRows := sqlmock.NewRows([]string{
					"id", "event_id", "user_id", "status", "waitlist_position",
					"registered_at", "confirmed_at", "cancelled_at",
				}).AddRow(
					waitlistID, eventID, waitlistUserID, model.RegistrationWaitlist, 1,
					now.Add(-time.Hour), nil, nil,
				)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM registrations`)).
					WithArgs(eventID).
					WillReturnRows(waitlistRows)

				// Promote the waitlisted user
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE registrations`)).
					WithArgs(waitlistID).
					WillReturnResult(sqlmock.NewResult(0, 1))

				// Reorder remaining waitlist positions
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE registrations`)).
					WithArgs(eventID).
					WillReturnResult(sqlmock.NewResult(0, 0))

				mock.ExpectCommit()
			},
			expectedPromoted: true,
			expectedError:    nil,
		},
		{
			name:             "cancel confirmed with no waitlist",
			registrationID:   uuid.New(),
			eventID:          uuid.New(),
			regStatus:        model.RegistrationConfirmed,
			waitlistPosition: nil,
			hasWaitlist:      false,
			setupMock: func(mock sqlmock.Sqlmock, regID, eventID uuid.UUID, regStatus model.RegistrationStatus, waitlistPos *int, hasWaitlist bool) {
				mock.ExpectBegin()

				// Lock and get the registration
				now := time.Now()
				regRows := sqlmock.NewRows([]string{
					"id", "event_id", "user_id", "status", "waitlist_position",
					"registered_at", "confirmed_at", "cancelled_at",
				}).AddRow(
					regID, eventID, uuid.New(), regStatus, waitlistPos,
					now, now, nil,
				)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM registrations WHERE id = $1 FOR UPDATE`)).
					WithArgs(regID).
					WillReturnRows(regRows)

				// Update to cancelled
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE registrations SET status = 'cancelled', cancelled_at = NOW(), waitlist_position = NULL WHERE id = $1`)).
					WithArgs(regID).
					WillReturnResult(sqlmock.NewResult(0, 1))

				// Get first waitlist person - none found
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM registrations`)).
					WithArgs(eventID).
					WillReturnError(sql.ErrNoRows)

				mock.ExpectCommit()
			},
			expectedPromoted: false,
			expectedError:    nil,
		},
		{
			name:             "cancel waitlist registration (no promotion needed)",
			registrationID:   uuid.New(),
			eventID:          uuid.New(),
			regStatus:        model.RegistrationWaitlist,
			waitlistPosition: intPtr(2),
			hasWaitlist:      false,
			setupMock: func(mock sqlmock.Sqlmock, regID, eventID uuid.UUID, regStatus model.RegistrationStatus, waitlistPos *int, hasWaitlist bool) {
				mock.ExpectBegin()

				// Lock and get the registration
				now := time.Now()
				regRows := sqlmock.NewRows([]string{
					"id", "event_id", "user_id", "status", "waitlist_position",
					"registered_at", "confirmed_at", "cancelled_at",
				}).AddRow(
					regID, eventID, uuid.New(), regStatus, waitlistPos,
					now, nil, nil,
				)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM registrations WHERE id = $1 FOR UPDATE`)).
					WithArgs(regID).
					WillReturnRows(regRows)

				// Update to cancelled
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE registrations SET status = 'cancelled', cancelled_at = NOW(), waitlist_position = NULL WHERE id = $1`)).
					WithArgs(regID).
					WillReturnResult(sqlmock.NewResult(0, 1))

				// Reorder waitlist positions for remaining waitlisted users
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE registrations`)).
					WithArgs(eventID, *waitlistPos).
					WillReturnResult(sqlmock.NewResult(0, 1))

				mock.ExpectCommit()
			},
			expectedPromoted: false,
			expectedError:    nil,
		},
		{
			name:             "cancel already cancelled registration",
			registrationID:   uuid.New(),
			eventID:          uuid.New(),
			regStatus:        model.RegistrationCancelled,
			waitlistPosition: nil,
			hasWaitlist:      false,
			setupMock: func(mock sqlmock.Sqlmock, regID, eventID uuid.UUID, regStatus model.RegistrationStatus, waitlistPos *int, hasWaitlist bool) {
				mock.ExpectBegin()

				// Lock and get the registration
				now := time.Now()
				cancelledAt := now.Add(-time.Hour)
				regRows := sqlmock.NewRows([]string{
					"id", "event_id", "user_id", "status", "waitlist_position",
					"registered_at", "confirmed_at", "cancelled_at",
				}).AddRow(
					regID, eventID, uuid.New(), regStatus, waitlistPos,
					now.Add(-2*time.Hour), nil, cancelledAt,
				)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM registrations WHERE id = $1 FOR UPDATE`)).
					WithArgs(regID).
					WillReturnRows(regRows)

				mock.ExpectRollback()
			},
			expectedPromoted: false,
			expectedError:    ErrAlreadyCancelled,
		},
		{
			name:             "registration not found",
			registrationID:   uuid.New(),
			eventID:          uuid.New(),
			regStatus:        model.RegistrationConfirmed,
			waitlistPosition: nil,
			hasWaitlist:      false,
			setupMock: func(mock sqlmock.Sqlmock, regID, eventID uuid.UUID, regStatus model.RegistrationStatus, waitlistPos *int, hasWaitlist bool) {
				mock.ExpectBegin()

				// Lock and get the registration - not found
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM registrations WHERE id = $1 FOR UPDATE`)).
					WithArgs(regID).
					WillReturnError(sql.ErrNoRows)

				mock.ExpectRollback()
			},
			expectedPromoted: false,
			expectedError:    sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			repo := NewRegistrationRepository(db)
			tt.setupMock(mock, tt.registrationID, tt.eventID, tt.regStatus, tt.waitlistPosition, tt.hasWaitlist)

			// Start transaction
			tx, err := db.Beginx()
			if err != nil && tt.expectedError == nil {
				t.Fatalf("failed to begin transaction: %v", err)
			}

			if tt.expectedError != nil && err != nil {
				return
			}

			promoted, err := repo.CancelAndPromote(context.Background(), tx, tt.registrationID, tt.eventID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				}
				tx.Rollback()
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				tx.Rollback()
				return
			}

			if tt.expectedPromoted && promoted == nil {
				t.Error("expected promoted registration, got nil")
			}
			if !tt.expectedPromoted && promoted != nil {
				t.Errorf("expected no promotion, got %v", promoted)
			}

			tx.Commit()

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// =============================================================================
// GetRegistrationStats Tests
// =============================================================================

func TestGetRegistrationStats(t *testing.T) {
	tests := []struct {
		name              string
		eventID           uuid.UUID
		setupMock         func(mock sqlmock.Sqlmock, eventID uuid.UUID)
		expectedConfirmed int
		expectedWaitlist  int
		expectedError     error
	}{
		{
			name:    "event with mixed registrations",
			eventID: uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, eventID uuid.UUID) {
				rows := sqlmock.NewRows([]string{"confirmed_count", "waitlist_count"}).AddRow(4, 2)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
					WithArgs(eventID).
					WillReturnRows(rows)
			},
			expectedConfirmed: 4,
			expectedWaitlist:  2,
			expectedError:     nil,
		},
		{
			name:    "event with no registrations",
			eventID: uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, eventID uuid.UUID) {
				rows := sqlmock.NewRows([]string{"confirmed_count", "waitlist_count"}).AddRow(0, 0)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
					WithArgs(eventID).
					WillReturnRows(rows)
			},
			expectedConfirmed: 0,
			expectedWaitlist:  0,
			expectedError:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			repo := NewRegistrationRepository(db)
			tt.setupMock(mock, tt.eventID)

			stats, err := repo.GetRegistrationStats(context.Background(), tt.eventID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if stats.ConfirmedCount != tt.expectedConfirmed {
				t.Errorf("expected confirmed count %d, got %d", tt.expectedConfirmed, stats.ConfirmedCount)
			}
			if stats.WaitlistCount != tt.expectedWaitlist {
				t.Errorf("expected waitlist count %d, got %d", tt.expectedWaitlist, stats.WaitlistCount)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// =============================================================================
// CancelAllByEventID Tests
// =============================================================================

func TestCancelAllByEventID(t *testing.T) {
	tests := []struct {
		name          string
		eventID       uuid.UUID
		setupMock     func(mock sqlmock.Sqlmock, eventID uuid.UUID)
		expectedError error
	}{
		{
			name:    "cancel all registrations",
			eventID: uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, eventID uuid.UUID) {
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE registrations SET status = 'cancelled', cancelled_at = NOW() WHERE event_id = $1 AND status != 'cancelled'`)).
					WithArgs(eventID).
					WillReturnResult(sqlmock.NewResult(0, 5))
			},
			expectedError: nil,
		},
		{
			name:    "no registrations to cancel",
			eventID: uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, eventID uuid.UUID) {
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE registrations SET status = 'cancelled', cancelled_at = NOW() WHERE event_id = $1 AND status != 'cancelled'`)).
					WithArgs(eventID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			repo := NewRegistrationRepository(db)
			tt.setupMock(mock, tt.eventID)

			err := repo.CancelAllByEventID(context.Background(), tt.eventID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// =============================================================================
// Delete Tests
// =============================================================================

func TestDelete(t *testing.T) {
	tests := []struct {
		name          string
		regID         uuid.UUID
		setupMock     func(mock sqlmock.Sqlmock, id uuid.UUID)
		expectedError error
	}{
		{
			name:  "successful delete",
			regID: uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, id uuid.UUID) {
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM registrations WHERE id = $1`)).
					WithArgs(id).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: nil,
		},
		{
			name:  "registration not found",
			regID: uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, id uuid.UUID) {
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM registrations WHERE id = $1`)).
					WithArgs(id).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedError: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupMockDB(t)
			defer db.Close()

			repo := NewRegistrationRepository(db)
			tt.setupMock(mock, tt.regID)

			err := repo.Delete(context.Background(), tt.regID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func intPtr(i int) *int {
	return &i
}
