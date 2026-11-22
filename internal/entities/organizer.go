package entities

import "time"

// Organizer adalah entitas penyelenggara event.
type Organizer struct {
	ID          int64
	Name        string
	Slug        string
	Email       string
	Phone       *string
	LogoURL     *string
	Description *string
	IsVerified  bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// EventOrganizerUser adalah relasi user dengan organizer.
type EventOrganizerUser struct {
	ID          int64
	OrganizerID int64
	UserID      int64
	Role        string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
