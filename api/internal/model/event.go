package model

import "time"

// Event represents a scheduled gathering (can be standalone or nested in Adventure)
type Event struct {
	ID               string  `json:"id"`
	GuildID          *string `json:"guild_id,omitempty"`          // nil = public event
	AdventureID      *string `json:"adventure_id,omitempty"`      // If part of an adventure
	OrderInAdventure *int    `json:"order_in_adventure,omitempty"` // Sequence within adventure
	Title            string  `json:"title"`
	Description *string    `json:"description,omitempty"`
	Location    *EventLocation `json:"location,omitempty"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	// Event configuration
	Template         string `json:"template"`                  // casual, dinner_party, activity, etc.
	Visibility       string `json:"visibility"`                // public, circle, invite_only
	MaxAttendees     *int   `json:"max_attendees,omitempty"`
	WaitlistEnabled  bool   `json:"waitlist_enabled"`
	RequiresApproval bool   `json:"requires_approval"`         // Host must approve all RSVPs
	AllowPlusOnes    bool   `json:"allow_plus_ones"`           // Guests can bring +1
	MaxPlusOnes      int    `json:"max_plus_ones"`             // Per guest (default 1)
	// Styling
	CoverImage  *string `json:"cover_image,omitempty"`
	ThemeColor  *string `json:"theme_color,omitempty"`
	// Values alignment settings
	ValuesRequired     bool     `json:"values_required"`      // Must pass values check
	ValuesQuestions    []string `json:"values_questions,omitempty"` // Specific question IDs
	AutoApproveAligned bool     `json:"auto_approve_aligned"` // Auto-RSVP if aligned
	YikesThreshold     int      `json:"yikes_threshold"`      // Max yikes before waiting room (0=any)
	// Support/listening event
	IsSupportEvent bool `json:"is_support_event"` // For Mana scoring

	// Completion verification (for Resonance scoring)
	// 1:1 events: BOTH parties must confirm within deadline
	// Group events: host + ≥2 attendees must confirm within deadline
	ConfirmationDeadline *time.Time `json:"confirmation_deadline,omitempty"`
	ConfirmedCount       int        `json:"confirmed_count"`
	RequiresConfirmation bool       `json:"requires_confirmation"`
	CompletionVerified   bool       `json:"completion_verified"`
	CompletionVerifiedOn *time.Time `json:"completion_verified_on,omitempty"`

	// Denormalized count for performance
	AttendeeCount int `json:"attendee_count"`

	// Status
	Status    string    `json:"status"` // draft, published, cancelled, completed
	CreatedBy string    `json:"created_by"`
	CreatedOn time.Time `json:"created_on"`
	UpdatedOn time.Time `json:"updated_on"`
}

// EventLocation represents where an event takes place
type EventLocation struct {
	Name         string  `json:"name"`                   // e.g., "Central Park"
	Address      *string `json:"address,omitempty"`      // Full address (only for confirmed)
	Neighborhood *string `json:"neighborhood,omitempty"` // General area
	City         string  `json:"city"`
	// Internal coordinates - never exposed to non-attendees
	Lat float64 `json:"-"`
	Lng float64 `json:"-"`
	// Virtual event
	IsVirtual bool    `json:"is_virtual"`
	MeetLink  *string `json:"meet_link,omitempty"` // Only shown to confirmed attendees
}

// EventTemplate constants
const (
	EventTemplateCasual      = "casual"       // Informal hangout
	EventTemplateDinnerParty = "dinner_party" // Hosted meal
	EventTemplateActivity    = "activity"     // Specific activity (hiking, games, etc.)
	EventTemplateBirthday    = "birthday"     // Celebration
	EventTemplateSupport     = "support"      // Listening/support session
	EventTemplateWorkshop    = "workshop"     // Learning/teaching
	EventTemplateTrip        = "trip"         // Travel/outing
)

// EventVisibility constants
const (
	EventVisibilityPublic     = "public"      // Anyone can discover
	EventVisibilityGuilds     = "guilds"      // Guild members only
	EventVisibilityInviteOnly = "invite_only" // Must be invited
	EventVisibilityPrivate    = "private"     // Only organizers (draft mode)
)

// EventStatus constants
const (
	EventStatusDraft     = "draft"
	EventStatusPublished = "published"
	EventStatusCancelled = "cancelled"
	EventStatusCompleted = "completed"
)

// Confirmation deadline constants
const (
	// ConfirmationDeadlineHours is how long after event end users have to confirm
	// Note: Same as ConfirmWindowHours in resonance.go
	ConfirmationDeadlineHours = 48

	// MinConfirmationsFor1on1 is confirmations needed for 1:1 event
	MinConfirmationsFor1on1 = 2

	// MinConfirmationsForGroup is confirmations needed for group event
	MinConfirmationsForGroup = 3
)

// Note: EarlyConfirmHours and CheckinWindowMinutes are defined in resonance.go

// IsCompletionVerifiable checks if this event has enough confirmations to be verified
func (e *Event) IsCompletionVerifiable() bool {
	// 1:1 event (max 2 attendees): both must confirm
	if e.MaxAttendees != nil && *e.MaxAttendees <= 2 {
		return e.ConfirmedCount >= MinConfirmationsFor1on1
	}
	// Group event: host + 2 attendees (3 total) must confirm
	return e.ConfirmedCount >= MinConfirmationsForGroup
}

// IsWithinConfirmationDeadline checks if the event can still accept confirmations
func (e *Event) IsWithinConfirmationDeadline() bool {
	if e.ConfirmationDeadline == nil {
		return false
	}
	return time.Now().Before(*e.ConfirmationDeadline)
}

// EventRSVP represents a user's response to an event
type EventRSVP struct {
	ID             string     `json:"id"`
	EventID        string     `json:"event_id"`
	UserID         string     `json:"user_id"`
	Status         string     `json:"status"`         // pending, approved, waitlisted, declined, cancelled
	RSVPType       string     `json:"rsvp_type"`      // going, maybe, not_going
	// Values alignment results (internal, not exposed)
	ValuesAligned  bool       `json:"-"`
	AlignmentScore float64    `json:"-"`
	YikesCount     int        `json:"-"`
	// Waiting room info (only for pending)
	WaitingReason  *string    `json:"waiting_reason,omitempty"` // "values_review", "capacity", "host_approval"
	// Host response
	HostNote       *string    `json:"host_note,omitempty"`      // Private message to RSVP'er
	RespondedBy    *string    `json:"responded_by,omitempty"`   // Host who responded
	RespondedOn    *time.Time `json:"responded_on,omitempty"`
	// Timestamps
	RequestedOn    time.Time  `json:"requested_on"`
	UpdatedOn      time.Time  `json:"updated_on"`
	// Plus ones
	PlusOnes       int        `json:"plus_ones"`
	PlusOneNames   []string   `json:"plus_one_names,omitempty"`
	// Resonance tracking (completion verification, checkin, support feedback)
	CompletionConfirmed *time.Time `json:"completion_confirmed,omitempty"`
	CheckinTime         *time.Time `json:"checkin_time,omitempty"`
	HelpfulnessRating   *string    `json:"helpfulness_rating,omitempty"` // YES, SOMEWHAT, NOT_REALLY, SKIP
	HelpfulnessTags     []string   `json:"helpfulness_tags,omitempty"`
}

// RSVPStatus constants
const (
	RSVPStatusPending    = "pending"    // In waiting room
	RSVPStatusApproved   = "approved"   // Confirmed attendance
	RSVPStatusWaitlisted = "waitlisted" // Event full, on waitlist
	RSVPStatusDeclined   = "declined"   // Host declined
	RSVPStatusCancelled  = "cancelled"  // User cancelled
)

// RSVPType constants
const (
	RSVPTypeGoing    = "going"
	RSVPTypeMaybe    = "maybe"
	RSVPTypeNotGoing = "not_going"
)

// WaitingReason constants
const (
	WaitingReasonValuesReview = "values_review" // Values alignment under review
	WaitingReasonCapacity     = "capacity"      // Event at capacity
	WaitingReasonHostApproval = "host_approval" // Host must approve all RSVPs
	WaitingReasonYikes        = "yikes"         // Yikes flags triggered
)

// EventHost represents a host/organizer of an event
type EventHost struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`      // primary, co_host
	AddedOn   time.Time `json:"added_on"`
	AddedBy   string    `json:"added_by"`
}

// HostRole constants
const (
	HostRolePrimary = "primary"
	HostRoleCoHost  = "co_host"
)

// Note: EventParticipant is defined in resonance.go with full Resonance tracking fields

// EventValuesCheck holds the result of checking a user's values against event requirements
type EventValuesCheck struct {
	UserID          string   `json:"user_id"`
	EventID         string   `json:"event_id"`
	IsAligned       bool     `json:"is_aligned"`
	AlignmentScore  float64  `json:"alignment_score"`  // 0-100
	YikesCount      int      `json:"yikes_count"`
	YikesCategories []string `json:"yikes_categories,omitempty"`
	MissingQuestions []string `json:"missing_questions,omitempty"` // Questions user hasn't answered
	Recommendation  string   `json:"recommendation"` // auto_approve, needs_review, decline_suggested
}

// ValuesRecommendation constants
const (
	ValuesRecommendAutoApprove    = "auto_approve"
	ValuesRecommendNeedsReview    = "needs_review"
	ValuesRecommendDeclineSuggested = "decline_suggested"
)

// EventWithDetails includes full event information
type EventWithDetails struct {
	Event        Event              `json:"event"`
	Hosts        []EventHost        `json:"hosts"`
	Roles        []EventRole        `json:"roles,omitempty"`
	AttendeesCount int              `json:"attendees_count"`
	WaitlistCount  int              `json:"waitlist_count"`
	UserRSVP     *EventRSVP         `json:"user_rsvp,omitempty"` // Current user's RSVP
	UserRole     *EventRoleAssignment `json:"user_role,omitempty"`
}

// EventSummary provides minimal event info for lists
type EventSummary struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	StartTime     time.Time  `json:"start_time"`
	Location      *EventLocation `json:"location,omitempty"`
	Template      string     `json:"template"`
	AttendeesCount int       `json:"attendees_count"`
	IsFull        bool       `json:"is_full"`
	UserRSVP      *string    `json:"user_rsvp,omitempty"` // User's current status
}

// Constraints
const (
	MaxEventTitleLength       = 100
	MaxEventDescriptionLength = 2000
	MaxEventHosts             = 5
	MaxValuesQuestions        = 10
	MaxPlusOnesPerRSVP        = 5
	DefaultYikesThreshold     = 2
)

// CreateEventRequest represents a request to create an event
type CreateEventRequest struct {
	GuildID            *string        `json:"guild_id,omitempty"`
	AdventureID        *string        `json:"adventure_id,omitempty"`
	Title              string         `json:"title"`
	Description        *string        `json:"description,omitempty"`
	Location           *EventLocation `json:"location,omitempty"`
	StartTime          time.Time      `json:"start_time"`
	EndTime            *time.Time     `json:"end_time,omitempty"`
	Template           string         `json:"template"`
	Visibility         string         `json:"visibility"`
	MaxAttendees       *int           `json:"max_attendees,omitempty"`
	WaitlistEnabled    bool           `json:"waitlist_enabled"`
	RequiresApproval   bool           `json:"requires_approval"`
	AllowPlusOnes      bool           `json:"allow_plus_ones"`
	MaxPlusOnes        int            `json:"max_plus_ones,omitempty"`
	CoverImage         *string        `json:"cover_image,omitempty"`
	ThemeColor         *string        `json:"theme_color,omitempty"`
	ValuesRequired     bool           `json:"values_required"`
	ValuesQuestions    []string       `json:"values_questions,omitempty"`
	AutoApproveAligned bool           `json:"auto_approve_aligned"`
	YikesThreshold     int            `json:"yikes_threshold"`
	IsSupportEvent     bool           `json:"is_support_event"`
}

// UpdateEventRequest represents a request to update an event
type UpdateEventRequest struct {
	Title              *string        `json:"title,omitempty"`
	Description        *string        `json:"description,omitempty"`
	Location           *EventLocation `json:"location,omitempty"`
	StartTime          *time.Time     `json:"start_time,omitempty"`
	EndTime            *time.Time     `json:"end_time,omitempty"`
	MaxAttendees       *int           `json:"max_attendees,omitempty"`
	WaitlistEnabled    *bool          `json:"waitlist_enabled,omitempty"`
	RequiresApproval   *bool          `json:"requires_approval,omitempty"`
	AllowPlusOnes      *bool          `json:"allow_plus_ones,omitempty"`
	MaxPlusOnes        *int           `json:"max_plus_ones,omitempty"`
	CoverImage         *string        `json:"cover_image,omitempty"`
	ThemeColor         *string        `json:"theme_color,omitempty"`
	ValuesRequired     *bool          `json:"values_required,omitempty"`
	ValuesQuestions    []string       `json:"values_questions,omitempty"`
	AutoApproveAligned *bool          `json:"auto_approve_aligned,omitempty"`
	YikesThreshold     *int           `json:"yikes_threshold,omitempty"`
	Status             *string        `json:"status,omitempty"`
}

// RSVPRequest represents a request to RSVP to an event
type RSVPRequest struct {
	RSVPType     string   `json:"rsvp_type"` // going, maybe, not_going
	PlusOnes     int      `json:"plus_ones,omitempty"`
	PlusOneNames []string `json:"plus_one_names,omitempty"`
	Note         *string  `json:"note,omitempty"` // Message to host
}

// RespondToRSVPRequest represents host's response to an RSVP
type RespondToRSVPRequest struct {
	Approved bool    `json:"approved"`
	Note     *string `json:"note,omitempty"` // Private message to user
}

// ConfirmEventCompletionRequest for resonance scoring
type ConfirmEventCompletionRequest struct {
	Completed bool `json:"completed"`
}

// EventCheckinRequest for on-time bonus
type EventCheckinRequest struct {
	// Location can be verified but we don't require it
}

// EventFeedbackRequest for support event helpfulness
type EventFeedbackRequest struct {
	HelpfulnessRating string   `json:"helpfulness_rating"` // yes, somewhat, not_really
	Tags              []string `json:"tags,omitempty"`
}

// EventSearchFilters for discovering events
type EventSearchFilters struct {
	GuildID     *string    `json:"guild_id,omitempty"`
	AdventureID *string    `json:"adventure_id,omitempty"`
	Template    *string    `json:"template,omitempty"`
	StartAfter *time.Time `json:"start_after,omitempty"`
	StartBefore *time.Time `json:"start_before,omitempty"`
	City       *string    `json:"city,omitempty"`
	Visibility *string    `json:"visibility,omitempty"`
	HostID     *string    `json:"host_id,omitempty"`
}
