package model

import "time"

// HangoutType represents the type of hangout being sought
type HangoutType string

const (
	HangoutTypeTalkItOut        HangoutType = "talk_it_out"       // Need a listening ear
	HangoutTypeHereToListen     HangoutType = "here_to_listen"    // Ready to support someone
	HangoutTypeConcreteActivity HangoutType = "concrete_activity" // Specific activity/venue
	HangoutTypeMutualInterest   HangoutType = "mutual_interest"   // Connect over shared interest
	HangoutTypeMeetAnyone       HangoutType = "meet_anyone"       // Open to anyone
)

// HangoutTypeInfo provides display information for hangout types
type HangoutTypeInfo struct {
	Type        HangoutType `json:"type"`
	Label       string      `json:"label"`
	Description string      `json:"description"`
	Icon        string      `json:"icon"`
}

// GetHangoutTypeInfo returns display info for each type
func GetHangoutTypeInfo() []HangoutTypeInfo {
	return []HangoutTypeInfo{
		{
			Type:        HangoutTypeTalkItOut,
			Label:       "Talk It Out",
			Description: "I have something on my mind I'd like to talk through",
			Icon:        "bubble.left.fill",
		},
		{
			Type:        HangoutTypeHereToListen,
			Label:       "Here to Listen",
			Description: "I'm in a good headspace and happy to listen",
			Icon:        "ear.fill",
		},
		{
			Type:        HangoutTypeConcreteActivity,
			Label:       "Concrete Activity",
			Description: "I want to do a specific thing with someone",
			Icon:        "figure.walk",
		},
		{
			Type:        HangoutTypeMutualInterest,
			Label:       "Mutual Interest",
			Description: "I want to connect with someone who shares an interest",
			Icon:        "heart.fill",
		},
		{
			Type:        HangoutTypeMeetAnyone,
			Label:       "Meet Anyone",
			Description: "I'm open to meeting new people, surprise me",
			Icon:        "sparkles",
		},
	}
}

// AvailabilityStatus constants
type AvailabilityStatus string

const (
	AvailabilityStatusAvailable AvailabilityStatus = "available"
	AvailabilityStatusMaybe     AvailabilityStatus = "maybe"
	AvailabilityStatusBusy      AvailabilityStatus = "busy"
)

// AvailabilityLocation represents internal location with radius for matching
type AvailabilityLocation struct {
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	Radius float64 `json:"radius"` // in km
}

// HangoutLocation represents a hangout location
type HangoutLocation struct {
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
	Venue *string `json:"venue,omitempty"`
}

// Availability represents a user's availability window
type Availability struct {
	ID                  string                `json:"id"`
	UserID              string                `json:"user_id"`
	Status              AvailabilityStatus    `json:"status"`
	StartTime           time.Time             `json:"start_time"`
	EndTime             time.Time             `json:"end_time"`
	Location            *AvailabilityLocation `json:"-"` // Never expose exact location
	HangoutType         HangoutType           `json:"hangout_type"`
	ActivityDescription *string               `json:"activity_description,omitempty"` // For concrete_activity
	ActivityVenue       *string               `json:"activity_venue,omitempty"`       // Specific venue
	InterestID          *string               `json:"interest_id,omitempty"`          // For mutual_interest
	MaxPeople           int                   `json:"max_people"`
	Note                *string               `json:"note,omitempty"`
	Visibility          string                `json:"visibility"` // circles, public
	ExpiresAt           time.Time             `json:"expires_at"`
	CreatedOn           time.Time             `json:"created_on"`
	UpdatedOn           time.Time             `json:"updated_on"`
}

// AvailabilityPublic is what other users see (with distance bucket, not coordinates)
type AvailabilityPublic struct {
	ID                  string         `json:"id"`
	UserID              string         `json:"user_id"`
	Status              string         `json:"status"`
	StartTime           time.Time      `json:"start_time"`
	EndTime             time.Time      `json:"end_time"`
	Distance            DistanceBucket `json:"distance,omitempty"`
	City                string         `json:"city,omitempty"`
	HangoutType         HangoutType    `json:"hangout_type"`
	ActivityDescription *string        `json:"activity_description,omitempty"`
	ActivityVenue       *string        `json:"activity_venue,omitempty"`
	InterestID          *string        `json:"interest_id,omitempty"`
	MaxPeople           int            `json:"max_people"`
	Note                *string        `json:"note,omitempty"`
	// User info
	UserProfile *PublicProfile `json:"user_profile,omitempty"`
	// Interest details if mutual_interest type
	Interest *Interest `json:"interest,omitempty"`
}

// HangoutRequest represents a request to join someone's availability
type HangoutRequest struct {
	ID             string     `json:"id"`
	AvailabilityID string     `json:"availability_id"`
	RequesterID    string     `json:"requester_id"`
	Note           string     `json:"note"`   // Required, min 20 chars
	Status         string     `json:"status"` // pending, accepted, declined, cancelled
	RespondedOn    *time.Time `json:"responded_on,omitempty"`
	CreatedOn      time.Time  `json:"created_on"`
}

// Hangout represents a confirmed meetup between users
type Hangout struct {
	ID                  string           `json:"id"`
	Participants        []string         `json:"participants"` // User IDs
	AvailabilityID      *string          `json:"availability_id,omitempty"`
	HangoutType         HangoutType      `json:"hangout_type"`
	ActivityDescription *string          `json:"activity_description,omitempty"`
	ScheduledTime       time.Time        `json:"scheduled_time"`
	Location            *HangoutLocation `json:"-"` // Internal only
	IsSupportSession    bool             `json:"is_support_session"`
	Status              string           `json:"status"` // scheduled, completed, cancelled, no_show
	CreatedOn           time.Time        `json:"created_on"`
	UpdatedOn           time.Time        `json:"updated_on"`
}

// HangoutStatus constants
const (
	HangoutStatusScheduled = "scheduled"
	HangoutStatusCompleted = "completed"
	HangoutStatusCancelled = "cancelled"
	HangoutStatusNoShow    = "no_show"
)

// RequestStatus constants
const (
	RequestStatusPending   = "pending"
	RequestStatusAccepted  = "accepted"
	RequestStatusDeclined  = "declined"
	RequestStatusCancelled = "cancelled"
)

// Availability constraints
const (
	MinNoteLength          = 20 // Minimum note length for friction
	MinHangoutNoteLength   = 20 // Alias for clarity
	MaxNoteLength          = 500
	MaxActivityDescLength  = 200
	MaxVenueLength         = 200
	DefaultAvailabilityTTL = 24 * time.Hour // Expire after 24 hours
)

// HangoutRequestStatus constants
const (
	HangoutRequestStatusPending   = "pending"
	HangoutRequestStatusAccepted  = "accepted"
	HangoutRequestStatusDeclined  = "declined"
	HangoutRequestStatusCancelled = "cancelled"
)

// AvailabilityLocationRequest is used when creating availability with location/radius
type AvailabilityLocationRequest struct {
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	Radius float64 `json:"radius,omitempty"` // Default: 5km
}

// CreateAvailabilityRequest represents a request to post availability
type CreateAvailabilityRequest struct {
	Status              string                       `json:"status,omitempty"` // Default: available
	StartTime           string                       `json:"start_time"`
	EndTime             string                       `json:"end_time"`
	Location            *AvailabilityLocationRequest `json:"location,omitempty"`
	HangoutType         string                       `json:"hangout_type"`
	ActivityDescription *string                      `json:"activity_description,omitempty"`
	ActivityVenue       *string                      `json:"activity_venue,omitempty"`
	InterestID          *string                      `json:"interest_id,omitempty"`
	MaxPeople           *int                         `json:"max_people,omitempty"` // Default: 1
	Note                *string                      `json:"note,omitempty"`
	Visibility          *string                      `json:"visibility,omitempty"` // Default: circles
}

// UpdateAvailabilityRequest represents a request to update availability
type UpdateAvailabilityRequest struct {
	Status              *string `json:"status,omitempty"`
	EndTime             *string `json:"end_time,omitempty"`
	ActivityDescription *string `json:"activity_description,omitempty"`
	ActivityVenue       *string `json:"activity_venue,omitempty"`
	MaxPeople           *int    `json:"max_people,omitempty"`
	Note                *string `json:"note,omitempty"`
}

// CreateHangoutRequestRequest represents a request to join someone's availability
type CreateHangoutRequestRequest struct {
	AvailabilityID string `json:"availability_id"`
	Note           string `json:"note"` // Required, min 20 chars
}

// RespondToRequestRequest represents accepting/declining a hangout request
type RespondToRequestRequest struct {
	Status      string  `json:"status"`                // accepted, declined
	Alternative *string `json:"alternative,omitempty"` // Suggest alternative if declining
}

// ActivitySuggestion represents a suggested concrete activity
type ActivitySuggestion struct {
	Category    string `json:"category"` // nearby_now, popular, your_interests
	Icon        string `json:"icon"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Venue       string `json:"venue,omitempty"`
	Time        string `json:"time,omitempty"` // "8pm", "open till 11"
}

// AvailabilityFilter for searching available people
type AvailabilityFilter struct {
	HangoutTypes []HangoutType `json:"hangout_types,omitempty"`
	RadiusKm     *float64      `json:"radius_km,omitempty"`
	StartAfter   *time.Time    `json:"start_after,omitempty"`
	EndBefore    *time.Time    `json:"end_before,omitempty"`
	InterestID   *string       `json:"interest_id,omitempty"`
}
