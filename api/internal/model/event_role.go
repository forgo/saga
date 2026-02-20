package model

import "time"

// EventRole represents a role that attendees can fill at an event
type EventRole struct {
	ID          string    `json:"id"`
	EventID     string    `json:"event_id"`
	Name        string    `json:"name"`                  // e.g., "Dessert-bringer", "DJ", "Guest"
	Description *string   `json:"description,omitempty"` // e.g., "Bring something sweet to share"
	MaxSlots    int       `json:"max_slots"`             // Default 1, 0 = unlimited (for default Guest role)
	FilledSlots int       `json:"filled_slots"`          // Computed from assignments
	IsDefault   bool      `json:"is_default"`            // True for the default "Guest" role
	SortOrder   int       `json:"sort_order"`            // Display ordering
	CreatedBy   string    `json:"created_by"`            // Host who created this role
	CreatedOn   time.Time `json:"created_on"`
	UpdatedOn   time.Time `json:"updated_on"`
	// Optional: suggest this role to users with matching interests
	SuggestedInterests []string `json:"suggested_interests,omitempty"`
}

// DefaultMaxSlotsPerRole is the default number of slots for a role (1 person per role)
const DefaultMaxSlotsPerRole = 1

// EventRoleAssignment represents a user's assignment to a role
// A user can have multiple role assignments per event (e.g., DJ + bring lasagna + wash dishes)
type EventRoleAssignment struct {
	ID         string    `json:"id"`
	EventID    string    `json:"event_id"`
	RoleID     string    `json:"role_id"`
	UserID     string    `json:"user_id"`
	Note       *string   `json:"note,omitempty"` // User's note about their contribution
	Status     string    `json:"status"`         // pending, confirmed, cancelled
	AssignedOn time.Time `json:"assigned_on"`
	UpdatedOn  time.Time `json:"updated_on"`
	// Populated by joins
	RoleName *string `json:"role_name,omitempty"`
}

// UserEventRoles represents all roles a user has taken on for an event
type UserEventRoles struct {
	UserID      string                `json:"user_id"`
	EventID     string                `json:"event_id"`
	Assignments []EventRoleAssignment `json:"assignments"`
}

// EventRoleAssignmentStatus constants
const (
	RoleAssignmentStatusPending   = "pending"   // User expressed interest, awaiting confirmation
	RoleAssignmentStatusConfirmed = "confirmed" // User is confirmed in this role
	RoleAssignmentStatusCancelled = "cancelled" // User cancelled their assignment
)

// Note: EventAttendanceConfig fields are now part of Event struct in event.go

// EventRoleWithAssignments includes the role and its current assignments
type EventRoleWithAssignments struct {
	Role        EventRole             `json:"role"`
	Assignments []EventRoleAssignment `json:"assignments"`
	IsFull      bool                  `json:"is_full"`
	SpotsLeft   int                   `json:"spots_left"` // -1 if unlimited
}

// EventRolesOverview provides a summary of all roles for an event
type EventRolesOverview struct {
	EventID        string                     `json:"event_id"`
	TotalSlots     int                        `json:"total_slots"`     // Total role slots available (-1 if unlimited)
	FilledSlots    int                        `json:"filled_slots"`    // Currently filled slots
	TotalAttendees int                        `json:"total_attendees"` // Same as filled_slots (for compatibility)
	MaxAttendees   *int                       `json:"max_attendees,omitempty"`
	IsFull         bool                       `json:"is_full"`
	Roles          []EventRoleWithAssignments `json:"roles"`
}

// Role name constants
const (
	DefaultRoleName        = "Guest"
	DefaultRoleDescription = "Join us and have a great time!"
)

// Constraints
const (
	MaxRolesPerEvent     = 20
	MaxRoleNameLength    = 50
	MaxRoleDescLength    = 500
	MaxAssignmentNoteLen = 200
)

// CreateEventRoleRequest represents a request to create a role
type CreateEventRoleRequest struct {
	Name               string   `json:"name"`
	Description        *string  `json:"description,omitempty"`
	MaxSlots           int      `json:"max_slots,omitempty"` // 0 = unlimited
	SuggestedInterests []string `json:"suggested_interests,omitempty"`
}

// UpdateEventRoleRequest represents a request to update a role
type UpdateEventRoleRequest struct {
	Name               *string  `json:"name,omitempty"`
	Description        *string  `json:"description,omitempty"`
	MaxSlots           *int     `json:"max_slots,omitempty"`
	SuggestedInterests []string `json:"suggested_interests,omitempty"`
}

// AssignRoleRequest represents a request to assign oneself to a role
type AssignRoleRequest struct {
	RoleID string  `json:"role_id"`
	Note   *string `json:"note,omitempty"` // e.g., "I'll bring vegan brownies"
}

// UpdateAssignmentRequest represents a request to update an assignment
type UpdateAssignmentRequest struct {
	Note *string `json:"note,omitempty"`
}

// RoleSuggestion represents a suggested role based on user interests
type RoleSuggestion struct {
	Role            EventRole `json:"role"`
	MatchedInterest string    `json:"matched_interest"`
	Reason          string    `json:"reason"` // e.g., "You're interested in baking"
}
