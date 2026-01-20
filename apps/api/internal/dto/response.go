package dto

import "github.com/anthropics/pickle-go/apps/api/internal/model"

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents an error response
type APIError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse creates a success response
func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
	}
}

// ErrorResponse creates an error response
func ErrorResponse(code, message string) APIResponse {
	return APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	}
}

// AuthResponse represents the response for authentication
type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token,omitempty"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	Email       *string `json:"email,omitempty"`
}

// FromUser converts a model.User to UserResponse
func FromUser(user *model.User) UserResponse {
	return UserResponse{
		ID:          user.ID.String(),
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Email:       user.Email,
	}
}

// EventResponse represents an event in API responses
type EventResponse struct {
	ID             string           `json:"id"`
	Host           UserResponse     `json:"host"`
	Title          *string          `json:"title,omitempty"`
	Description    *string          `json:"description,omitempty"`
	EventDate      string           `json:"event_date"`
	StartTime      string           `json:"start_time"`
	EndTime        *string          `json:"end_time,omitempty"`
	Location       LocationResponse `json:"location"`
	Capacity       int              `json:"capacity"`
	ConfirmedCount int              `json:"confirmed_count"`
	WaitlistCount  int              `json:"waitlist_count"`
	SkillLevel     string           `json:"skill_level"`
	SkillLevelLabel string          `json:"skill_level_label"`
	Fee            int              `json:"fee"`
	Status         string           `json:"status"`
}

// LocationResponse represents location data in responses
type LocationResponse struct {
	Name          string  `json:"name"`
	Address       *string `json:"address,omitempty"`
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	GooglePlaceID *string `json:"google_place_id,omitempty"`
}

// EventListResponse represents a list of events
type EventListResponse struct {
	Events  []EventResponse `json:"events"`
	Total   int             `json:"total"`
	HasMore bool            `json:"has_more"`
}

// RegistrationResponse represents a registration in API responses
type RegistrationResponse struct {
	ID               string  `json:"id"`
	EventID          string  `json:"event_id"`
	Status           string  `json:"status"`
	WaitlistPosition *int    `json:"waitlist_position,omitempty"`
	Message          string  `json:"message"`
}

// CreateEventResponse represents the response for creating an event
type CreateEventResponse struct {
	ID       string `json:"id"`
	ShareURL string `json:"share_url"`
}
