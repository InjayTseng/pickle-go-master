package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/anthropics/pickle-go/apps/api/internal/database"
	"github.com/anthropics/pickle-go/apps/api/internal/dto"
	"github.com/anthropics/pickle-go/apps/api/internal/middleware"
	"github.com/anthropics/pickle-go/apps/api/internal/model"
	"github.com/anthropics/pickle-go/apps/api/internal/repository"
	"github.com/anthropics/pickle-go/apps/api/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// =============================================================================
// Test Setup Helpers
// =============================================================================

type testContext struct {
	router       *gin.Engine
	handler      *RegistrationHandler
	mock         sqlmock.Sqlmock
	db           *sqlx.DB
	regRepo      *repository.RegistrationRepository
	eventRepo    *repository.EventRepository
	notifRepo    *repository.NotificationRepository
	txManager    *database.TxManager
}

func setupTestContext(t *testing.T) *testContext {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}

	db := sqlx.NewDb(mockDB, "postgres")
	regRepo := repository.NewRegistrationRepository(db)
	eventRepo := repository.NewEventRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	txManager := database.NewTxManager(db)

	handler := NewRegistrationHandler(regRepo, eventRepo, notifRepo, txManager)

	router := gin.New()

	return &testContext{
		router:    router,
		handler:   handler,
		mock:      mock,
		db:        db,
		regRepo:   regRepo,
		eventRepo: eventRepo,
		notifRepo: notifRepo,
		txManager: txManager,
	}
}

func (tc *testContext) cleanup() {
	tc.db.Close()
}

// createAuthContext creates a gin context with authenticated user
func createAuthContext(userID, displayName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := &jwt.Claims{
			UserID:      userID,
			DisplayName: displayName,
		}
		c.Set(middleware.AuthUserKey, claims)
		c.Next()
	}
}

// parseResponse parses the JSON response body
func parseResponse(t *testing.T, recorder *httptest.ResponseRecorder) dto.APIResponse {
	var response dto.APIResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	return response
}

// =============================================================================
// RegisterEvent Handler Tests
// =============================================================================

func TestRegisterEvent_Success_Confirmed(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	userID := uuid.New()
	eventID := uuid.New()
	hostID := uuid.New()

	// Setup mock expectations for event lookup (outside transaction)
	now := time.Now()
	eventRows := sqlmock.NewRows([]string{
		"id", "host_id", "short_code", "title", "description", "event_date", "start_time", "end_time",
		"location_name", "location_address", "latitude", "longitude", "google_place_id",
		"capacity", "skill_level", "fee", "status", "created_at", "updated_at",
	}).AddRow(
		eventID, hostID, "abc123", nil, nil, now, "20:00", nil,
		"Test Location", nil, 25.033, 121.565, nil,
		4, "beginner", 200, "open", now, now,
	)
	tc.mock.ExpectQuery("SELECT .* FROM events WHERE id").
		WithArgs(eventID).
		WillReturnRows(eventRows)

	// Transaction expectations
	tc.mock.ExpectBegin()

	// Lock event
	lockEventRows := sqlmock.NewRows([]string{"capacity", "status", "host_id"}).
		AddRow(4, "open", hostID)
	tc.mock.ExpectQuery("SELECT capacity, status, host_id FROM events WHERE id = .* FOR UPDATE").
		WithArgs(eventID).
		WillReturnRows(lockEventRows)

	// Check existing registration
	tc.mock.ExpectQuery("SELECT .* FROM registrations WHERE event_id = .* AND user_id").
		WithArgs(eventID, userID).
		WillReturnError(sql.ErrNoRows)

	// Count confirmed
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	tc.mock.ExpectQuery("SELECT COUNT.* FROM registrations WHERE event_id = .* AND status = 'confirmed'").
		WithArgs(eventID).
		WillReturnRows(countRows)

	// Insert registration
	insertRows := sqlmock.NewRows([]string{"registered_at", "confirmed_at"}).AddRow(now, now)
	tc.mock.ExpectQuery("INSERT INTO registrations").
		WithArgs(sqlmock.AnyArg(), eventID, userID, model.RegistrationConfirmed, nil).
		WillReturnRows(insertRows)

	tc.mock.ExpectCommit()

	// Setup router
	tc.router.POST("/events/:id/register", createAuthContext(userID.String(), "Test User"), tc.handler.RegisterEvent)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/events/"+eventID.String()+"/register", nil)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, recorder.Code)
	}

	response := parseResponse(t, recorder)
	if !response.Success {
		t.Errorf("expected success, got error: %v", response.Error)
	}

	if err := tc.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRegisterEvent_Success_Waitlist(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	userID := uuid.New()
	eventID := uuid.New()
	hostID := uuid.New()

	now := time.Now()

	// Event lookup
	eventRows := sqlmock.NewRows([]string{
		"id", "host_id", "short_code", "title", "description", "event_date", "start_time", "end_time",
		"location_name", "location_address", "latitude", "longitude", "google_place_id",
		"capacity", "skill_level", "fee", "status", "created_at", "updated_at",
	}).AddRow(
		eventID, hostID, "abc123", nil, nil, now, "20:00", nil,
		"Test Location", nil, 25.033, 121.565, nil,
		4, "beginner", 200, "full", now, now,
	)
	tc.mock.ExpectQuery("SELECT .* FROM events WHERE id").
		WithArgs(eventID).
		WillReturnRows(eventRows)

	// Transaction
	tc.mock.ExpectBegin()

	// Lock event
	lockEventRows := sqlmock.NewRows([]string{"capacity", "status", "host_id"}).
		AddRow(4, "full", hostID)
	tc.mock.ExpectQuery("SELECT capacity, status, host_id FROM events WHERE id = .* FOR UPDATE").
		WithArgs(eventID).
		WillReturnRows(lockEventRows)

	// Check existing registration
	tc.mock.ExpectQuery("SELECT .* FROM registrations WHERE event_id = .* AND user_id").
		WithArgs(eventID, userID).
		WillReturnError(sql.ErrNoRows)

	// Count confirmed (event is full)
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(4)
	tc.mock.ExpectQuery("SELECT COUNT.* FROM registrations WHERE event_id = .* AND status = 'confirmed'").
		WithArgs(eventID).
		WillReturnRows(countRows)

	// Get max waitlist position
	maxPosRows := sqlmock.NewRows([]string{"max"}).AddRow(1)
	tc.mock.ExpectQuery("SELECT MAX.*waitlist_position.* FROM registrations").
		WithArgs(eventID).
		WillReturnRows(maxPosRows)

	// Insert waitlist registration
	insertRows := sqlmock.NewRows([]string{"registered_at"}).AddRow(now)
	tc.mock.ExpectQuery("INSERT INTO registrations").
		WithArgs(sqlmock.AnyArg(), eventID, userID, model.RegistrationWaitlist, 2).
		WillReturnRows(insertRows)

	tc.mock.ExpectCommit()

	// Setup router
	tc.router.POST("/events/:id/register", createAuthContext(userID.String(), "Test User"), tc.handler.RegisterEvent)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/events/"+eventID.String()+"/register", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, recorder.Code)
	}

	response := parseResponse(t, recorder)
	if !response.Success {
		t.Errorf("expected success, got error: %v", response.Error)
	}

	// Check the response includes waitlist position
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("response data is not a map")
	}
	if data["status"] != "waitlist" {
		t.Errorf("expected status 'waitlist', got %v", data["status"])
	}
	if data["waitlist_position"] != float64(2) {
		t.Errorf("expected waitlist_position 2, got %v", data["waitlist_position"])
	}

	if err := tc.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRegisterEvent_Unauthorized(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	eventID := uuid.New()

	// Setup router WITHOUT auth middleware
	tc.router.POST("/events/:id/register", tc.handler.RegisterEvent)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/events/"+eventID.String()+"/register", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestRegisterEvent_InvalidEventID(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	userID := uuid.New()

	// Setup router
	tc.router.POST("/events/:id/register", createAuthContext(userID.String(), "Test User"), tc.handler.RegisterEvent)

	// Make request with invalid event ID
	req := httptest.NewRequest(http.MethodPost, "/events/invalid-uuid/register", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	response := parseResponse(t, recorder)
	if response.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected error code VALIDATION_ERROR, got %s", response.Error.Code)
	}
}

func TestRegisterEvent_EventNotFound(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	userID := uuid.New()
	eventID := uuid.New()

	// Event not found
	tc.mock.ExpectQuery("SELECT .* FROM events WHERE id").
		WithArgs(eventID).
		WillReturnError(sql.ErrNoRows)

	// Setup router
	tc.router.POST("/events/:id/register", createAuthContext(userID.String(), "Test User"), tc.handler.RegisterEvent)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/events/"+eventID.String()+"/register", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}

	response := parseResponse(t, recorder)
	if response.Error.Code != "NOT_FOUND" {
		t.Errorf("expected error code NOT_FOUND, got %s", response.Error.Code)
	}
}

func TestRegisterEvent_HostCannotRegister(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	userID := uuid.New()
	eventID := uuid.New()
	hostID := userID // Host is the same as user

	now := time.Now()

	// Event lookup
	eventRows := sqlmock.NewRows([]string{
		"id", "host_id", "short_code", "title", "description", "event_date", "start_time", "end_time",
		"location_name", "location_address", "latitude", "longitude", "google_place_id",
		"capacity", "skill_level", "fee", "status", "created_at", "updated_at",
	}).AddRow(
		eventID, hostID, "abc123", nil, nil, now, "20:00", nil,
		"Test Location", nil, 25.033, 121.565, nil,
		4, "beginner", 200, "open", now, now,
	)
	tc.mock.ExpectQuery("SELECT .* FROM events WHERE id").
		WithArgs(eventID).
		WillReturnRows(eventRows)

	// Transaction
	tc.mock.ExpectBegin()

	// Lock event - host_id matches userID
	lockEventRows := sqlmock.NewRows([]string{"capacity", "status", "host_id"}).
		AddRow(4, "open", hostID)
	tc.mock.ExpectQuery("SELECT capacity, status, host_id FROM events WHERE id = .* FOR UPDATE").
		WithArgs(eventID).
		WillReturnRows(lockEventRows)

	tc.mock.ExpectRollback()

	// Setup router
	tc.router.POST("/events/:id/register", createAuthContext(userID.String(), "Test User"), tc.handler.RegisterEvent)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/events/"+eventID.String()+"/register", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	response := parseResponse(t, recorder)
	if response.Error.Code != "HOST_CANNOT_REGISTER" {
		t.Errorf("expected error code HOST_CANNOT_REGISTER, got %s", response.Error.Code)
	}
}

func TestRegisterEvent_AlreadyRegistered(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	userID := uuid.New()
	eventID := uuid.New()
	hostID := uuid.New()

	now := time.Now()

	// Event lookup
	eventRows := sqlmock.NewRows([]string{
		"id", "host_id", "short_code", "title", "description", "event_date", "start_time", "end_time",
		"location_name", "location_address", "latitude", "longitude", "google_place_id",
		"capacity", "skill_level", "fee", "status", "created_at", "updated_at",
	}).AddRow(
		eventID, hostID, "abc123", nil, nil, now, "20:00", nil,
		"Test Location", nil, 25.033, 121.565, nil,
		4, "beginner", 200, "open", now, now,
	)
	tc.mock.ExpectQuery("SELECT .* FROM events WHERE id").
		WithArgs(eventID).
		WillReturnRows(eventRows)

	// Transaction
	tc.mock.ExpectBegin()

	// Lock event
	lockEventRows := sqlmock.NewRows([]string{"capacity", "status", "host_id"}).
		AddRow(4, "open", hostID)
	tc.mock.ExpectQuery("SELECT capacity, status, host_id FROM events WHERE id = .* FOR UPDATE").
		WithArgs(eventID).
		WillReturnRows(lockEventRows)

	// Check existing registration - already registered
	existingRows := sqlmock.NewRows([]string{
		"id", "event_id", "user_id", "status", "waitlist_position",
		"registered_at", "confirmed_at", "cancelled_at",
	}).AddRow(
		uuid.New(), eventID, userID, "confirmed", nil,
		now, now, nil,
	)
	tc.mock.ExpectQuery("SELECT .* FROM registrations WHERE event_id = .* AND user_id").
		WithArgs(eventID, userID).
		WillReturnRows(existingRows)

	tc.mock.ExpectRollback()

	// Setup router
	tc.router.POST("/events/:id/register", createAuthContext(userID.String(), "Test User"), tc.handler.RegisterEvent)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/events/"+eventID.String()+"/register", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	response := parseResponse(t, recorder)
	if response.Error.Code != "ALREADY_REGISTERED" {
		t.Errorf("expected error code ALREADY_REGISTERED, got %s", response.Error.Code)
	}
}

func TestRegisterEvent_EventClosed(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	userID := uuid.New()
	eventID := uuid.New()
	hostID := uuid.New()

	now := time.Now()

	// Event lookup
	eventRows := sqlmock.NewRows([]string{
		"id", "host_id", "short_code", "title", "description", "event_date", "start_time", "end_time",
		"location_name", "location_address", "latitude", "longitude", "google_place_id",
		"capacity", "skill_level", "fee", "status", "created_at", "updated_at",
	}).AddRow(
		eventID, hostID, "abc123", nil, nil, now, "20:00", nil,
		"Test Location", nil, 25.033, 121.565, nil,
		4, "beginner", 200, "cancelled", now, now,
	)
	tc.mock.ExpectQuery("SELECT .* FROM events WHERE id").
		WithArgs(eventID).
		WillReturnRows(eventRows)

	// Transaction
	tc.mock.ExpectBegin()

	// Lock event - status is cancelled
	lockEventRows := sqlmock.NewRows([]string{"capacity", "status", "host_id"}).
		AddRow(4, "cancelled", hostID)
	tc.mock.ExpectQuery("SELECT capacity, status, host_id FROM events WHERE id = .* FOR UPDATE").
		WithArgs(eventID).
		WillReturnRows(lockEventRows)

	tc.mock.ExpectRollback()

	// Setup router
	tc.router.POST("/events/:id/register", createAuthContext(userID.String(), "Test User"), tc.handler.RegisterEvent)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/events/"+eventID.String()+"/register", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	response := parseResponse(t, recorder)
	if response.Error.Code != "EVENT_CLOSED" {
		t.Errorf("expected error code EVENT_CLOSED, got %s", response.Error.Code)
	}
}

// =============================================================================
// CancelRegistration Handler Tests
// =============================================================================

func TestCancelRegistration_Success_WithPromotion(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	userID := uuid.New()
	eventID := uuid.New()
	regID := uuid.New()
	hostID := uuid.New()
	promotedUserID := uuid.New()
	promotedRegID := uuid.New()

	now := time.Now()

	// Find user's registration
	regRows := sqlmock.NewRows([]string{
		"id", "event_id", "user_id", "status", "waitlist_position",
		"registered_at", "confirmed_at", "cancelled_at",
	}).AddRow(
		regID, eventID, userID, "confirmed", nil,
		now, now, nil,
	)
	tc.mock.ExpectQuery("SELECT .* FROM registrations WHERE event_id = .* AND user_id").
		WithArgs(eventID, userID).
		WillReturnRows(regRows)

	// Transaction
	tc.mock.ExpectBegin()

	// Lock the registration
	lockRegRows := sqlmock.NewRows([]string{
		"id", "event_id", "user_id", "status", "waitlist_position",
		"registered_at", "confirmed_at", "cancelled_at",
	}).AddRow(
		regID, eventID, userID, "confirmed", nil,
		now, now, nil,
	)
	tc.mock.ExpectQuery("SELECT .* FROM registrations WHERE id = .* FOR UPDATE").
		WithArgs(regID).
		WillReturnRows(lockRegRows)

	// Update to cancelled
	tc.mock.ExpectExec("UPDATE registrations SET status = 'cancelled'").
		WithArgs(regID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Get first waitlist person
	waitlistRows := sqlmock.NewRows([]string{
		"id", "event_id", "user_id", "status", "waitlist_position",
		"registered_at", "confirmed_at", "cancelled_at",
	}).AddRow(
		promotedRegID, eventID, promotedUserID, "waitlist", 1,
		now.Add(-time.Hour), nil, nil,
	)
	tc.mock.ExpectQuery("SELECT .* FROM registrations").
		WithArgs(eventID).
		WillReturnRows(waitlistRows)

	// Promote the waitlisted user
	tc.mock.ExpectExec("UPDATE registrations").
		WithArgs(promotedRegID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Reorder waitlist
	tc.mock.ExpectExec("UPDATE registrations").
		WithArgs(eventID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	tc.mock.ExpectCommit()

	// Notification (look up event for notification)
	eventRows := sqlmock.NewRows([]string{
		"id", "host_id", "short_code", "title", "description", "event_date", "start_time", "end_time",
		"location_name", "location_address", "latitude", "longitude", "google_place_id",
		"capacity", "skill_level", "fee", "status", "created_at", "updated_at",
	}).AddRow(
		eventID, hostID, "abc123", nil, nil, now, "20:00", nil,
		"Test Location", nil, 25.033, 121.565, nil,
		4, "beginner", 200, "open", now, now,
	)
	tc.mock.ExpectQuery("SELECT .* FROM events WHERE id").
		WithArgs(eventID).
		WillReturnRows(eventRows)

	// Create notification
	tc.mock.ExpectQuery("INSERT INTO notifications").
		WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(now))

	// Update event status (after cancel) - event lookup
	eventRows2 := sqlmock.NewRows([]string{
		"id", "host_id", "short_code", "title", "description", "event_date", "start_time", "end_time",
		"location_name", "location_address", "latitude", "longitude", "google_place_id",
		"capacity", "skill_level", "fee", "status", "created_at", "updated_at",
	}).AddRow(
		eventID, hostID, "abc123", nil, nil, now, "20:00", nil,
		"Test Location", nil, 25.033, 121.565, nil,
		4, "beginner", 200, "full", now, now,
	)
	tc.mock.ExpectQuery("SELECT .* FROM events WHERE id").
		WithArgs(eventID).
		WillReturnRows(eventRows2)

	// Count confirmed
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
	tc.mock.ExpectQuery("SELECT COUNT.* FROM registrations").
		WithArgs(eventID).
		WillReturnRows(countRows)

	// Update event status to open
	tc.mock.ExpectExec("UPDATE events SET status").
		WithArgs(eventID, model.EventStatusOpen).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Setup router
	tc.router.DELETE("/events/:id/register", createAuthContext(userID.String(), "Test User"), tc.handler.CancelRegistration)

	// Make request
	req := httptest.NewRequest(http.MethodDelete, "/events/"+eventID.String()+"/register", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	response := parseResponse(t, recorder)
	if !response.Success {
		t.Errorf("expected success, got error: %v", response.Error)
	}

	if err := tc.mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCancelRegistration_Success_NoWaitlist(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	userID := uuid.New()
	eventID := uuid.New()
	regID := uuid.New()
	hostID := uuid.New()

	now := time.Now()

	// Find user's registration
	regRows := sqlmock.NewRows([]string{
		"id", "event_id", "user_id", "status", "waitlist_position",
		"registered_at", "confirmed_at", "cancelled_at",
	}).AddRow(
		regID, eventID, userID, "confirmed", nil,
		now, now, nil,
	)
	tc.mock.ExpectQuery("SELECT .* FROM registrations WHERE event_id = .* AND user_id").
		WithArgs(eventID, userID).
		WillReturnRows(regRows)

	// Transaction
	tc.mock.ExpectBegin()

	// Lock the registration
	lockRegRows := sqlmock.NewRows([]string{
		"id", "event_id", "user_id", "status", "waitlist_position",
		"registered_at", "confirmed_at", "cancelled_at",
	}).AddRow(
		regID, eventID, userID, "confirmed", nil,
		now, now, nil,
	)
	tc.mock.ExpectQuery("SELECT .* FROM registrations WHERE id = .* FOR UPDATE").
		WithArgs(regID).
		WillReturnRows(lockRegRows)

	// Update to cancelled
	tc.mock.ExpectExec("UPDATE registrations SET status = 'cancelled'").
		WithArgs(regID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Get first waitlist person - none
	tc.mock.ExpectQuery("SELECT .* FROM registrations").
		WithArgs(eventID).
		WillReturnError(sql.ErrNoRows)

	tc.mock.ExpectCommit()

	// Update event status (after cancel) - event lookup
	eventRows := sqlmock.NewRows([]string{
		"id", "host_id", "short_code", "title", "description", "event_date", "start_time", "end_time",
		"location_name", "location_address", "latitude", "longitude", "google_place_id",
		"capacity", "skill_level", "fee", "status", "created_at", "updated_at",
	}).AddRow(
		eventID, hostID, "abc123", nil, nil, now, "20:00", nil,
		"Test Location", nil, 25.033, 121.565, nil,
		4, "beginner", 200, "full", now, now,
	)
	tc.mock.ExpectQuery("SELECT .* FROM events WHERE id").
		WithArgs(eventID).
		WillReturnRows(eventRows)

	// Count confirmed
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
	tc.mock.ExpectQuery("SELECT COUNT.* FROM registrations").
		WithArgs(eventID).
		WillReturnRows(countRows)

	// Update event status to open
	tc.mock.ExpectExec("UPDATE events SET status").
		WithArgs(eventID, model.EventStatusOpen).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Setup router
	tc.router.DELETE("/events/:id/register", createAuthContext(userID.String(), "Test User"), tc.handler.CancelRegistration)

	// Make request
	req := httptest.NewRequest(http.MethodDelete, "/events/"+eventID.String()+"/register", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	response := parseResponse(t, recorder)
	if !response.Success {
		t.Errorf("expected success, got error: %v", response.Error)
	}
}

func TestCancelRegistration_Unauthorized(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	eventID := uuid.New()

	// Setup router WITHOUT auth middleware
	tc.router.DELETE("/events/:id/register", tc.handler.CancelRegistration)

	// Make request
	req := httptest.NewRequest(http.MethodDelete, "/events/"+eventID.String()+"/register", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestCancelRegistration_NotRegistered(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	userID := uuid.New()
	eventID := uuid.New()

	// Find user's registration - not found
	tc.mock.ExpectQuery("SELECT .* FROM registrations WHERE event_id = .* AND user_id").
		WithArgs(eventID, userID).
		WillReturnError(sql.ErrNoRows)

	// Setup router
	tc.router.DELETE("/events/:id/register", createAuthContext(userID.String(), "Test User"), tc.handler.CancelRegistration)

	// Make request
	req := httptest.NewRequest(http.MethodDelete, "/events/"+eventID.String()+"/register", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}

	response := parseResponse(t, recorder)
	if response.Error.Code != "NOT_FOUND" {
		t.Errorf("expected error code NOT_FOUND, got %s", response.Error.Code)
	}
}

func TestCancelRegistration_AlreadyCancelled(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	userID := uuid.New()
	eventID := uuid.New()
	regID := uuid.New()

	now := time.Now()
	cancelledAt := now.Add(-time.Hour)

	// Find user's registration
	regRows := sqlmock.NewRows([]string{
		"id", "event_id", "user_id", "status", "waitlist_position",
		"registered_at", "confirmed_at", "cancelled_at",
	}).AddRow(
		regID, eventID, userID, "cancelled", nil,
		now.Add(-2*time.Hour), nil, cancelledAt,
	)
	tc.mock.ExpectQuery("SELECT .* FROM registrations WHERE event_id = .* AND user_id").
		WithArgs(eventID, userID).
		WillReturnRows(regRows)

	// Transaction
	tc.mock.ExpectBegin()

	// Lock the registration
	lockRegRows := sqlmock.NewRows([]string{
		"id", "event_id", "user_id", "status", "waitlist_position",
		"registered_at", "confirmed_at", "cancelled_at",
	}).AddRow(
		regID, eventID, userID, "cancelled", nil,
		now.Add(-2*time.Hour), nil, cancelledAt,
	)
	tc.mock.ExpectQuery("SELECT .* FROM registrations WHERE id = .* FOR UPDATE").
		WithArgs(regID).
		WillReturnRows(lockRegRows)

	tc.mock.ExpectRollback()

	// Setup router
	tc.router.DELETE("/events/:id/register", createAuthContext(userID.String(), "Test User"), tc.handler.CancelRegistration)

	// Make request
	req := httptest.NewRequest(http.MethodDelete, "/events/"+eventID.String()+"/register", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	response := parseResponse(t, recorder)
	if response.Error.Code != "ALREADY_CANCELLED" {
		t.Errorf("expected error code ALREADY_CANCELLED, got %s", response.Error.Code)
	}
}

// =============================================================================
// GetEventRegistrations Handler Tests
// =============================================================================

func TestGetEventRegistrations_Success(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	eventID := uuid.New()
	user1ID := uuid.New()
	user2ID := uuid.New()
	user3ID := uuid.New()

	now := time.Now()

	// Check event exists
	existsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	tc.mock.ExpectQuery("SELECT EXISTS").
		WithArgs(eventID).
		WillReturnRows(existsRows)

	// Get registrations with users
	regRows := sqlmock.NewRows([]string{
		"id", "event_id", "user_id", "status", "waitlist_position",
		"registered_at", "confirmed_at", "cancelled_at",
		"user.id", "user.display_name", "user.avatar_url",
	}).
		AddRow(uuid.New(), eventID, user1ID, "confirmed", nil, now, now, nil, user1ID, "User 1", nil).
		AddRow(uuid.New(), eventID, user2ID, "confirmed", nil, now, now, nil, user2ID, "User 2", nil).
		AddRow(uuid.New(), eventID, user3ID, "waitlist", 1, now, nil, nil, user3ID, "User 3", nil)
	tc.mock.ExpectQuery("SELECT").
		WithArgs(eventID).
		WillReturnRows(regRows)

	// Setup router (no auth required for viewing registrations)
	tc.router.GET("/events/:id/registrations", tc.handler.GetEventRegistrations)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/events/"+eventID.String()+"/registrations", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	response := parseResponse(t, recorder)
	if !response.Success {
		t.Errorf("expected success, got error: %v", response.Error)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("response data is not a map")
	}

	confirmedCount := int(data["confirmed_count"].(float64))
	waitlistCount := int(data["waitlist_count"].(float64))

	if confirmedCount != 2 {
		t.Errorf("expected confirmed_count 2, got %d", confirmedCount)
	}
	if waitlistCount != 1 {
		t.Errorf("expected waitlist_count 1, got %d", waitlistCount)
	}
}

func TestGetEventRegistrations_EventNotFound(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	eventID := uuid.New()

	// Check event exists - not found
	existsRows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	tc.mock.ExpectQuery("SELECT EXISTS").
		WithArgs(eventID).
		WillReturnRows(existsRows)

	// Setup router
	tc.router.GET("/events/:id/registrations", tc.handler.GetEventRegistrations)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/events/"+eventID.String()+"/registrations", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}

	response := parseResponse(t, recorder)
	if response.Error.Code != "NOT_FOUND" {
		t.Errorf("expected error code NOT_FOUND, got %s", response.Error.Code)
	}
}

func TestGetEventRegistrations_InvalidEventID(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	// Setup router
	tc.router.GET("/events/:id/registrations", tc.handler.GetEventRegistrations)

	// Make request with invalid event ID
	req := httptest.NewRequest(http.MethodGet, "/events/not-a-uuid/registrations", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	response := parseResponse(t, recorder)
	if response.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected error code VALIDATION_ERROR, got %s", response.Error.Code)
	}
}

func TestGetEventRegistrations_EmptyList(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	eventID := uuid.New()

	// Check event exists
	existsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	tc.mock.ExpectQuery("SELECT EXISTS").
		WithArgs(eventID).
		WillReturnRows(existsRows)

	// Get registrations - empty
	regRows := sqlmock.NewRows([]string{
		"id", "event_id", "user_id", "status", "waitlist_position",
		"registered_at", "confirmed_at", "cancelled_at",
		"user.id", "user.display_name", "user.avatar_url",
	})
	tc.mock.ExpectQuery("SELECT").
		WithArgs(eventID).
		WillReturnRows(regRows)

	// Setup router
	tc.router.GET("/events/:id/registrations", tc.handler.GetEventRegistrations)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/events/"+eventID.String()+"/registrations", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	response := parseResponse(t, recorder)
	if !response.Success {
		t.Errorf("expected success, got error: %v", response.Error)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("response data is not a map")
	}

	confirmed, ok := data["confirmed"].([]interface{})
	if !ok || len(confirmed) != 0 {
		t.Errorf("expected empty confirmed list, got %v", data["confirmed"])
	}

	waitlist, ok := data["waitlist"].([]interface{})
	if !ok || len(waitlist) != 0 {
		t.Errorf("expected empty waitlist, got %v", data["waitlist"])
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestRegisterEvent_DatabaseError(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	userID := uuid.New()
	eventID := uuid.New()

	// Database error on event lookup
	tc.mock.ExpectQuery("SELECT .* FROM events WHERE id").
		WithArgs(eventID).
		WillReturnError(errors.New("database connection lost"))

	// Setup router
	tc.router.POST("/events/:id/register", createAuthContext(userID.String(), "Test User"), tc.handler.RegisterEvent)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/events/"+eventID.String()+"/register", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}

	response := parseResponse(t, recorder)
	if response.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected error code INTERNAL_ERROR, got %s", response.Error.Code)
	}
}

func TestCancelRegistration_DatabaseError(t *testing.T) {
	tc := setupTestContext(t)
	defer tc.cleanup()

	userID := uuid.New()
	eventID := uuid.New()

	// Database error on registration lookup
	tc.mock.ExpectQuery("SELECT .* FROM registrations WHERE event_id = .* AND user_id").
		WithArgs(eventID, userID).
		WillReturnError(errors.New("database connection lost"))

	// Setup router
	tc.router.DELETE("/events/:id/register", createAuthContext(userID.String(), "Test User"), tc.handler.CancelRegistration)

	// Make request
	req := httptest.NewRequest(http.MethodDelete, "/events/"+eventID.String()+"/register", nil)
	recorder := httptest.NewRecorder()

	tc.router.ServeHTTP(recorder, req)

	// Assert response
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}

	response := parseResponse(t, recorder)
	if response.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected error code INTERNAL_ERROR, got %s", response.Error.Code)
	}
}

// =============================================================================
// Helper for request body
// =============================================================================

func jsonBody(v interface{}) *bytes.Buffer {
	data, _ := json.Marshal(v)
	return bytes.NewBuffer(data)
}
