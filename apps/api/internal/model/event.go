package model

import (
	"time"

	"github.com/google/uuid"
)

// SkillLevel represents the skill level of an event
type SkillLevel string

const (
	SkillBeginner     SkillLevel = "beginner"
	SkillIntermediate SkillLevel = "intermediate"
	SkillAdvanced     SkillLevel = "advanced"
	SkillExpert       SkillLevel = "expert"
	SkillAny          SkillLevel = "any"
)

// SkillLevelLabels maps skill levels to display labels
var SkillLevelLabels = map[SkillLevel]string{
	SkillBeginner:     "新手友善 (2.0-2.5)",
	SkillIntermediate: "中階 (2.5-3.5)",
	SkillAdvanced:     "進階 (3.5-4.5)",
	SkillExpert:       "高階 (4.5+)",
	SkillAny:          "不限程度",
}

// EventStatus represents the status of an event
type EventStatus string

const (
	EventStatusOpen      EventStatus = "open"
	EventStatusFull      EventStatus = "full"
	EventStatusCancelled EventStatus = "cancelled"
	EventStatusCompleted EventStatus = "completed"
)

// Event represents an event in the system
type Event struct {
	ID              uuid.UUID   `db:"id" json:"id"`
	HostID          uuid.UUID   `db:"host_id" json:"host_id"`
	Title           *string     `db:"title" json:"title,omitempty"`
	Description     *string     `db:"description" json:"description,omitempty"`
	EventDate       time.Time   `db:"event_date" json:"event_date"`
	StartTime       string      `db:"start_time" json:"start_time"`
	EndTime         *string     `db:"end_time" json:"end_time,omitempty"`
	LocationName    string      `db:"location_name" json:"location_name"`
	LocationAddress *string     `db:"location_address" json:"location_address,omitempty"`
	Latitude        float64     `db:"latitude" json:"latitude"`
	Longitude       float64     `db:"longitude" json:"longitude"`
	GooglePlaceID   *string     `db:"google_place_id" json:"google_place_id,omitempty"`
	Capacity        int         `db:"capacity" json:"capacity"`
	SkillLevel      SkillLevel  `db:"skill_level" json:"skill_level"`
	Fee             int         `db:"fee" json:"fee"`
	Status          EventStatus `db:"status" json:"status"`
	CreatedAt       time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time   `db:"updated_at" json:"updated_at"`
}

// EventSummary represents an event with registration counts
type EventSummary struct {
	Event
	ConfirmedCount int         `db:"confirmed_count" json:"confirmed_count"`
	WaitlistCount  int         `db:"waitlist_count" json:"waitlist_count"`
	Host           UserProfile `json:"host"`
}

// EventLocation represents the location of an event for API responses
type EventLocation struct {
	Name          string  `json:"name"`
	Address       *string `json:"address,omitempty"`
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	GooglePlaceID *string `json:"google_place_id,omitempty"`
}

// GetLocation returns the event location as an EventLocation struct
func (e *Event) GetLocation() EventLocation {
	return EventLocation{
		Name:          e.LocationName,
		Address:       e.LocationAddress,
		Lat:           e.Latitude,
		Lng:           e.Longitude,
		GooglePlaceID: e.GooglePlaceID,
	}
}

// GetSkillLevelLabel returns the display label for the event's skill level
func (e *Event) GetSkillLevelLabel() string {
	if label, ok := SkillLevelLabels[e.SkillLevel]; ok {
		return label
	}
	return string(e.SkillLevel)
}
