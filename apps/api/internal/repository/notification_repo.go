package repository

import (
	"context"
	"time"

	"github.com/anthropics/pickle-go/apps/api/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// NotificationRepository handles notification data access
type NotificationRepository struct {
	db *sqlx.DB
}

// NewNotificationRepository creates a new NotificationRepository
func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create creates a new notification
func (r *NotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	query := `
		INSERT INTO notifications (id, user_id, event_id, type, title, message, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`

	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}

	return r.db.QueryRowxContext(ctx, query,
		notification.ID, notification.UserID, notification.EventID,
		notification.Type, notification.Title, notification.Message,
		notification.IsRead, notification.CreatedAt,
	).Scan(&notification.CreatedAt)
}

// FindByUserID finds notifications for a user
func (r *NotificationRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Notification, error) {
	var notifications []model.Notification
	query := `
		SELECT id, user_id, event_id, type, title, message, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	err := r.db.SelectContext(ctx, &notifications, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// CountUnread counts unread notifications for a user
func (r *NotificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`
	err := r.db.GetContext(ctx, &count, query, userID)
	return count, err
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE notifications SET is_read = true WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// MarkAllAsRead marks all notifications for a user as read
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE notifications SET is_read = true WHERE user_id = $1 AND is_read = false`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// Delete deletes a notification by ID
func (r *NotificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`
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

// CreateWaitlistPromotedNotification creates a notification for when a user is promoted from waitlist
func (r *NotificationRepository) CreateWaitlistPromotedNotification(ctx context.Context, userID uuid.UUID, eventID uuid.UUID, eventTitle string) error {
	title := "You have been promoted from the waitlist!"
	message := "A spot has opened up. You are now confirmed for: " + eventTitle

	notification := &model.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		EventID:   &eventID,
		Type:      model.NotificationWaitlistPromoted,
		Title:     title,
		Message:   &message,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	return r.Create(ctx, notification)
}

// CreateEventCancelledNotification creates a notification for when an event is cancelled
func (r *NotificationRepository) CreateEventCancelledNotification(ctx context.Context, userID uuid.UUID, eventID uuid.UUID, eventTitle string) error {
	title := "Event has been cancelled"
	message := "The event you registered for has been cancelled: " + eventTitle

	notification := &model.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		EventID:   &eventID,
		Type:      model.NotificationEventCancelled,
		Title:     title,
		Message:   &message,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	return r.Create(ctx, notification)
}
