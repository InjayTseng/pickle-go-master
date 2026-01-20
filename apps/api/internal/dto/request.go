package dto

// LineCallbackRequest represents the request body for Line Login callback
type LineCallbackRequest struct {
	Code        string `json:"code" binding:"required"`
	State       string `json:"state"`
	RedirectURI string `json:"redirect_uri"`
}

// RefreshTokenRequest represents the request body for refreshing tokens
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// CreateEventRequest represents the request body for creating an event
type CreateEventRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	EventDate   string            `json:"event_date" binding:"required"`
	StartTime   string            `json:"start_time" binding:"required"`
	EndTime     string            `json:"end_time"`
	Location    LocationRequest   `json:"location" binding:"required"`
	Capacity    int               `json:"capacity" binding:"required,min=4,max=20"`
	SkillLevel  string            `json:"skill_level" binding:"required,oneof=beginner intermediate advanced expert any"`
	Fee         int               `json:"fee" binding:"min=0,max=9999"`
}

// LocationRequest represents location data in requests
type LocationRequest struct {
	Name          string  `json:"name" binding:"required"`
	Address       string  `json:"address"`
	Lat           float64 `json:"lat" binding:"required"`
	Lng           float64 `json:"lng" binding:"required"`
	GooglePlaceID string  `json:"google_place_id"`
}

// UpdateEventRequest represents the request body for updating an event
type UpdateEventRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	EventDate   *string `json:"event_date"`
	StartTime   *string `json:"start_time"`
	EndTime     *string `json:"end_time"`
	Capacity    *int    `json:"capacity" binding:"omitempty,min=4,max=20"`
	SkillLevel  *string `json:"skill_level" binding:"omitempty,oneof=beginner intermediate advanced expert any"`
	Fee         *int    `json:"fee" binding:"omitempty,min=0,max=9999"`
	Status      *string `json:"status" binding:"omitempty,oneof=open full cancelled"`
}

// ListEventsQuery represents query parameters for listing events
type ListEventsQuery struct {
	Lat        float64 `form:"lat"`
	Lng        float64 `form:"lng"`
	Radius     int     `form:"radius" binding:"max=50000"`
	SkillLevel string  `form:"skill_level"`
	Status     string  `form:"status"`
	Limit      int     `form:"limit" binding:"max=100"`
	Offset     int     `form:"offset"`
}
