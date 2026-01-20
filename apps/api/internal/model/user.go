package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	LineUserID  string     `db:"line_user_id" json:"line_user_id"`
	DisplayName string     `db:"display_name" json:"display_name"`
	AvatarURL   *string    `db:"avatar_url" json:"avatar_url,omitempty"`
	Email       *string    `db:"email" json:"email,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

// UserProfile represents a simplified user profile for display
type UserProfile struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
}

// ToProfile converts a User to a UserProfile
func (u *User) ToProfile() UserProfile {
	return UserProfile{
		ID:          u.ID,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
	}
}
