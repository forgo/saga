package model

import "time"

// UnifiedRSVP represents a polymorphic RSVP that works across events, adventures, hangouts, and pool matches.
// This replaces the fragmented EventRSVP, event_participant, and rsvp relation tables.
type UnifiedRSVP struct {
	ID         string `json:"id"`
	TargetType string `json:"target_type"` // event, adventure, hangout, pool_match
	TargetID   string `json:"target_id"`   // Record ID of the target
	UserID     string `json:"user_id"`

	// Status and role
	Status string `json:"status"` // pending, approved, waitlisted, declined, cancelled, attended, no_show
	Role   string `json:"role"`   // organizer, host, co_host, participant, helper, driver

	// Timestamps
	CreatedOn time.Time `json:"created_on"`
	UpdatedOn time.Time `json:"updated_on"`

	// Values alignment (computed, internal)
	ValuesAligned  *bool    `json:"-"`
	AlignmentScore *float64 `json:"-"`
	YikesCount     *int     `json:"-"`

	// Plus ones
	PlusOnes     *int     `json:"plus_ones,omitempty"`
	PlusOneNames []string `json:"plus_one_names,omitempty"`

	// Notes
	Note     *string `json:"note,omitempty"`      // User's note to host
	HostNote *string `json:"host_note,omitempty"` // Host's response

	// Confirmation & verification
	CheckinTime         *time.Time `json:"checkin_time,omitempty"`
	CompletionConfirmed *time.Time `json:"completion_confirmed,omitempty"`
	EarlyConfirmed      *bool      `json:"early_confirmed,omitempty"`

	// Feedback (for support sessions / hangouts)
	HelpfulnessRating *string  `json:"helpfulness_rating,omitempty"` // yes, somewhat, no
	HelpfulnessTags   []string `json:"helpfulness_tags,omitempty"`
}

// Target type constants
const (
	RSVPTargetEvent     = "event"
	RSVPTargetAdventure = "adventure"
	RSVPTargetHangout   = "hangout"
	RSVPTargetPoolMatch = "pool_match"
)

// Unified RSVP status constants
const (
	UnifiedRSVPStatusPending    = "pending"
	UnifiedRSVPStatusApproved   = "approved"
	UnifiedRSVPStatusWaitlisted = "waitlisted"
	UnifiedRSVPStatusDeclined   = "declined"
	UnifiedRSVPStatusCancelled  = "cancelled"
	UnifiedRSVPStatusAttended   = "attended"
	UnifiedRSVPStatusNoShow     = "no_show"
)

// Unified RSVP role constants
const (
	RSVPRoleOrganizer   = "organizer"
	RSVPRoleHost        = "host"
	RSVPRoleCoHost      = "co_host"
	RSVPRoleParticipant = "participant"
	RSVPRoleHelper      = "helper"
	RSVPRoleDriver      = "driver"
)

// Note: HelpfulnessRating is defined in resonance.go

// CreateRSVPRequest represents a request to create/update an RSVP
type CreateRSVPRequest struct {
	TargetType   string   `json:"target_type"`
	TargetID     string   `json:"target_id"`
	Role         string   `json:"role,omitempty"` // Defaults to participant
	PlusOnes     *int     `json:"plus_ones,omitempty"`
	PlusOneNames []string `json:"plus_one_names,omitempty"`
	Note         *string  `json:"note,omitempty"`
}

// UpdateRSVPRequest represents a request to update an existing RSVP
type UpdateRSVPRequest struct {
	Status       *string  `json:"status,omitempty"`
	Role         *string  `json:"role,omitempty"`
	PlusOnes     *int     `json:"plus_ones,omitempty"`
	PlusOneNames []string `json:"plus_one_names,omitempty"`
	Note         *string  `json:"note,omitempty"`
	HostNote     *string  `json:"host_note,omitempty"`
}

// RSVPConfirmationRequest represents a user confirming event completion
type RSVPConfirmationRequest struct {
	Completed bool `json:"completed"`
}

// RSVPCheckinRequest represents a user checking in to an event
type RSVPCheckinRequest struct {
	// Could include location verification in future
}

// RSVPFeedbackRequest represents feedback after a hangout/support session
type RSVPFeedbackRequest struct {
	HelpfulnessRating string   `json:"helpfulness_rating"` // yes, somewhat, no
	Tags              []string `json:"tags,omitempty"`
}

// RSVPHostResponseRequest represents a host's response to a pending RSVP
type RSVPHostResponseRequest struct {
	Approved bool    `json:"approved"`
	Note     *string `json:"note,omitempty"`
}

// RSVPWithTarget includes the RSVP plus details about what it's for
type RSVPWithTarget struct {
	RSVP       UnifiedRSVP `json:"rsvp"`
	TargetName string      `json:"target_name"`
	StartTime  *time.Time  `json:"start_time,omitempty"`
	Location   *string     `json:"location,omitempty"`
}

// RSVPListResponse is a paginated list of RSVPs
type RSVPListResponse struct {
	RSVPs      []UnifiedRSVP `json:"rsvps"`
	TotalCount int           `json:"total_count"`
	HasMore    bool          `json:"has_more"`
}

// RSVPFilters for querying RSVPs
type RSVPFilters struct {
	TargetType *string `json:"target_type,omitempty"`
	TargetID   *string `json:"target_id,omitempty"`
	UserID     *string `json:"user_id,omitempty"`
	Status     *string `json:"status,omitempty"`
	Role       *string `json:"role,omitempty"`
}

// RSVPStats provides aggregate stats for a target
type RSVPStats struct {
	TargetType    string `json:"target_type"`
	TargetID      string `json:"target_id"`
	ApprovedCount int    `json:"approved_count"`
	PendingCount  int    `json:"pending_count"`
	WaitlistCount int    `json:"waitlist_count"`
	AttendedCount int    `json:"attended_count"`
	TotalPlusOnes int    `json:"total_plus_ones"`
}

// Constraints
const (
	MaxRSVPNote    = 500
	MaxHostNote    = 500
	MaxPlusOnes    = 5
	MaxHelpfulTags = 5
	MaxTagLength   = 50
)

// Validate validates a CreateRSVPRequest
func (r *CreateRSVPRequest) Validate() []FieldError {
	var errors []FieldError

	if r.TargetType == "" {
		errors = append(errors, FieldError{Field: "target_type", Message: "target_type is required"})
	} else if r.TargetType != RSVPTargetEvent && r.TargetType != RSVPTargetAdventure &&
		r.TargetType != RSVPTargetHangout && r.TargetType != RSVPTargetPoolMatch {
		errors = append(errors, FieldError{Field: "target_type", Message: "invalid target_type"})
	}

	if r.TargetID == "" {
		errors = append(errors, FieldError{Field: "target_id", Message: "target_id is required"})
	}

	if r.PlusOnes != nil && *r.PlusOnes > MaxPlusOnes {
		errors = append(errors, FieldError{Field: "plus_ones", Message: "maximum 5 plus ones allowed"})
	}

	if r.Note != nil && len(*r.Note) > MaxRSVPNote {
		errors = append(errors, FieldError{Field: "note", Message: "note too long"})
	}

	return errors
}

// Validate validates an RSVPFeedbackRequest
func (r *RSVPFeedbackRequest) Validate() []FieldError {
	var errors []FieldError

	rating := HelpfulnessRating(r.HelpfulnessRating)
	if rating != HelpfulnessYes && rating != HelpfulnessSomewhat && rating != HelpfulnessNotReally && rating != HelpfulnessSkip {
		errors = append(errors, FieldError{Field: "helpfulness_rating", Message: "must be YES, SOMEWHAT, NOT_REALLY, or SKIP"})
	}

	if len(r.Tags) > MaxHelpfulTags {
		errors = append(errors, FieldError{Field: "tags", Message: "maximum 5 tags allowed"})
	}

	for i, tag := range r.Tags {
		if len(tag) > MaxTagLength {
			errors = append(errors, FieldError{
				Field:   "tags",
				Message: "tag " + string(rune(i)) + " is too long",
			})
		}
	}

	return errors
}
