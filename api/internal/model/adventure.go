package model

import "time"

// AdventureStatus represents the lifecycle stage of an adventure
type AdventureStatus string

const (
	AdventureStatusIdea      AdventureStatus = "idea"      // Just brainstorming
	AdventureStatusPlanning  AdventureStatus = "planning"  // Actively planning
	AdventureStatusConfirmed AdventureStatus = "confirmed" // Adventure is happening
	AdventureStatusActive    AdventureStatus = "active"    // Currently in progress
	AdventureStatusCompleted AdventureStatus = "completed" // Adventure finished
	AdventureStatusCancelled AdventureStatus = "cancelled" // Adventure cancelled
	AdventureStatusFrozen    AdventureStatus = "frozen"    // Organizer lost guild membership
)

// AdventureOrganizerType represents whether adventure is guild or user organized
type AdventureOrganizerType string

const (
	AdventureOrganizerGuild AdventureOrganizerType = "guild"
	AdventureOrganizerUser  AdventureOrganizerType = "user"
)

// AdventureVisibility represents who can see the adventure
type AdventureVisibility string

const (
	AdventureVisibilityPublic     AdventureVisibility = "public"      // Anyone can discover
	AdventureVisibilityGuilds     AdventureVisibility = "guilds"      // Guild members only
	AdventureVisibilityInviteOnly AdventureVisibility = "invite_only" // Must be invited
	AdventureVisibilityPrivate    AdventureVisibility = "private"     // Only organizers (draft)
)

// Adventure represents a multi-day, multi-location coordination (formerly Trip)
type Adventure struct {
	ID          string              `json:"id"`
	GuildID     *string             `json:"guild_id,omitempty"` // Deprecated: use OrganizerType/OrganizerID
	Title       string              `json:"title"`
	Description *string             `json:"description,omitempty"`
	StartDate   time.Time           `json:"start_date"`
	EndDate     time.Time           `json:"end_date"`
	Status      AdventureStatus     `json:"status"`
	Visibility  AdventureVisibility `json:"visibility"`
	CreatedByID string              `json:"created_by_id"` // User ID who created
	// Organizer fields (v2)
	OrganizerType   AdventureOrganizerType `json:"organizer_type"`    // guild or user
	OrganizerID     string                 `json:"organizer_id"`      // "guild:<id>" or "user:<id>"
	OrganizerUserID string                 `json:"organizer_user_id"` // User ID of the organizer
	// Frozen state (when organizer loses guild membership)
	FreezeReason *string    `json:"freeze_reason,omitempty"`
	FrozenOn     *time.Time `json:"frozen_on,omitempty"`
	// Values alignment settings
	ValuesRequired  bool     `json:"values_required"`
	ValuesQuestions []string `json:"values_questions,omitempty"`
	// Budget range (optional)
	BudgetMin *float64 `json:"budget_min,omitempty"`
	BudgetMax *float64 `json:"budget_max,omitempty"`
	Currency  *string  `json:"currency,omitempty"` // e.g., "USD", "EUR"
	// Styling
	CoverImage *string `json:"cover_image,omitempty"`
	// Voting settings
	VotingOpen     bool       `json:"voting_open"`
	VotingDeadline *time.Time `json:"voting_deadline,omitempty"`
	// Computed fields
	ParticipantCount int `json:"participant_count,omitempty"`
	DestinationCount int `json:"destination_count,omitempty"`
	EventCount       int `json:"event_count,omitempty"`
	AdmittedCount    int `json:"admitted_count,omitempty"` // Number of admitted users
	// Timestamps
	CreatedOn time.Time `json:"created_on"`
	UpdatedOn time.Time `json:"updated_on"`
}

// IsFrozen returns whether the adventure is frozen
func (a *Adventure) IsFrozen() bool {
	return a.Status == AdventureStatusFrozen
}

// IsGuildOrganized returns whether the adventure is organized by a guild
func (a *Adventure) IsGuildOrganized() bool {
	return a.OrganizerType == AdventureOrganizerGuild
}

// IsUserOrganized returns whether the adventure is organized by an individual user
func (a *Adventure) IsUserOrganized() bool {
	return a.OrganizerType == AdventureOrganizerUser
}

// AdventureParticipant represents a user participating in an adventure
type AdventureParticipant struct {
	ID          string    `json:"id"`
	AdventureID string    `json:"adventure_id"`
	UserID      string    `json:"user_id"`
	MemberID    *string   `json:"member_id,omitempty"` // If via guild membership
	Role        string    `json:"role"`                // organizer, participant
	Status      string    `json:"status"`              // interested, committed, maybe, out
	JoinedOn    time.Time `json:"joined_on"`
	UpdatedOn   time.Time `json:"updated_on"`
}

// AdventureParticipantRole constants
const (
	AdventureRoleOrganizer   = "organizer"
	AdventureRoleParticipant = "participant"
)

// AdventureParticipantStatus constants
const (
	AdventureParticipantInterested = "interested"
	AdventureParticipantCommitted  = "committed"
	AdventureParticipantMaybe      = "maybe"
	AdventureParticipantOut        = "out"
)

// Destination represents a proposed destination for an adventure
type Destination struct {
	ID           string  `json:"id"`
	AdventureID  string  `json:"adventure_id"`
	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	ProposedByID string  `json:"proposed_by_id"` // User ID
	OrderIndex   int     `json:"order_index"`    // Sequence in itinerary
	// Location info (privacy-safe)
	City    *string `json:"city,omitempty"`
	Country *string `json:"country,omitempty"`
	Region  *string `json:"region,omitempty"` // State/province
	Lat     float64 `json:"-"`                // Internal only
	Lng     float64 `json:"-"`                // Internal only
	// Estimated costs
	EstimatedCostMin *float64 `json:"estimated_cost_min,omitempty"`
	EstimatedCostMax *float64 `json:"estimated_cost_max,omitempty"`
	// Optional links
	InfoURL  *string `json:"info_url,omitempty"` // Tourist site, Airbnb, etc.
	ImageURL *string `json:"image_url,omitempty"`
	// Voting results (computed)
	VoteCount   int     `json:"vote_count,omitempty"`
	AverageRank float64 `json:"average_rank,omitempty"`
	VetoCount   int     `json:"veto_count,omitempty"`
	// Timestamps
	CreatedOn time.Time `json:"created_on"`
}

// DestinationVote represents a user's vote on a destination
type DestinationVote struct {
	ID            string `json:"id"`
	DestinationID string `json:"destination_id"`
	UserID        string `json:"user_id"`
	// Ranked choice: 1 = first choice, 2 = second, etc.
	Rank int `json:"rank"`
	// Veto means hard no (if allowed)
	Veto   bool    `json:"veto"`
	Reason *string `json:"reason,omitempty"` // Optional explanation for veto
	// Timestamps
	CreatedOn time.Time `json:"created_on"`
	UpdatedOn time.Time `json:"updated_on"`
}

// AdventureActivity represents a planned activity during the adventure
type AdventureActivity struct {
	ID            string     `json:"id"`
	AdventureID   string     `json:"adventure_id"`
	DestinationID *string    `json:"destination_id,omitempty"` // Optional link to destination
	Title         string     `json:"title"`
	Description   *string    `json:"description,omitempty"`
	ScheduledDate *time.Time `json:"scheduled_date,omitempty"`
	ScheduledTime *string    `json:"scheduled_time,omitempty"` // HH:MM format
	Duration      *int       `json:"duration_minutes,omitempty"`
	Location      *string    `json:"location,omitempty"` // Venue name/address
	EstimatedCost *float64   `json:"estimated_cost,omitempty"`
	BookingURL    *string    `json:"booking_url,omitempty"`
	ProposedByID  string     `json:"proposed_by_id"` // User ID
	// Voting
	VoteCount   int  `json:"vote_count,omitempty"`
	IsConfirmed bool `json:"is_confirmed"`
	// Timestamps
	CreatedOn time.Time `json:"created_on"`
	UpdatedOn time.Time `json:"updated_on"`
}

// ActivityVote represents a user's vote on an activity
type ActivityVote struct {
	ID         string    `json:"id"`
	ActivityID string    `json:"activity_id"`
	UserID     string    `json:"user_id"`
	Interested bool      `json:"interested"` // Simple yes/no for activities
	CreatedOn  time.Time `json:"created_on"`
}

// Constraints
const (
	MaxDestinationsPerAdventure     = 20
	MaxActivitiesPerAdventure       = 50
	MaxEventsPerAdventure           = 20
	MaxRidesharesPerAdventure       = 20
	MaxAdventureTitleLength         = 100
	MaxAdventureDescLength          = 1000
	MaxDestNameLength               = 100
	MaxDestDescLength               = 500
	MaxAdventureActivityTitleLength = 100
	MaxAdventureActivityDescLength  = 500
)

// CreateAdventureRequest represents a request to create an adventure
type CreateAdventureRequest struct {
	// Organizer type: "guild" or "user" (defaults to "guild" if GuildID provided)
	OrganizerType   *string  `json:"organizer_type,omitempty"` // guild or user
	GuildID         *string  `json:"guild_id,omitempty"`       // Required if organizer_type is guild
	Title           string   `json:"title"`
	Description     *string  `json:"description,omitempty"`
	StartDate       string   `json:"start_date"` // RFC3339 format
	EndDate         string   `json:"end_date"`   // RFC3339 format
	Visibility      string   `json:"visibility,omitempty"`
	ValuesRequired  bool     `json:"values_required,omitempty"`
	ValuesQuestions []string `json:"values_questions,omitempty"`
	BudgetMin       *float64 `json:"budget_min,omitempty"`
	BudgetMax       *float64 `json:"budget_max,omitempty"`
	Currency        *string  `json:"currency,omitempty"`
	CoverImage      *string  `json:"cover_image,omitempty"`
	VotingDeadline  *string  `json:"voting_deadline,omitempty"` // RFC3339 format
}

// GetOrganizerType returns the organizer type, defaulting based on GuildID
func (r *CreateAdventureRequest) GetOrganizerType() AdventureOrganizerType {
	if r.OrganizerType != nil {
		return AdventureOrganizerType(*r.OrganizerType)
	}
	if r.GuildID != nil && *r.GuildID != "" {
		return AdventureOrganizerGuild
	}
	return AdventureOrganizerUser
}

// Validate checks if the create request is valid
func (r *CreateAdventureRequest) Validate() []FieldError {
	var errors []FieldError

	if r.Title == "" {
		errors = append(errors, FieldError{Field: "title", Message: "title is required"})
	} else if len(r.Title) > MaxAdventureTitleLength {
		errors = append(errors, FieldError{Field: "title", Message: "title must be 100 characters or less"})
	}
	if r.Description != nil && len(*r.Description) > MaxAdventureDescLength {
		errors = append(errors, FieldError{Field: "description", Message: "description must be 1000 characters or less"})
	}
	if r.StartDate == "" {
		errors = append(errors, FieldError{Field: "start_date", Message: "start_date is required"})
	}
	if r.EndDate == "" {
		errors = append(errors, FieldError{Field: "end_date", Message: "end_date is required"})
	}
	// Validate organizer type
	if r.OrganizerType != nil {
		if *r.OrganizerType != string(AdventureOrganizerGuild) && *r.OrganizerType != string(AdventureOrganizerUser) {
			errors = append(errors, FieldError{Field: "organizer_type", Message: "organizer_type must be 'guild' or 'user'"})
		}
	}
	// Guild ID required for guild-organized adventures
	orgType := r.GetOrganizerType()
	if orgType == AdventureOrganizerGuild && (r.GuildID == nil || *r.GuildID == "") {
		errors = append(errors, FieldError{Field: "guild_id", Message: "guild_id is required for guild-organized adventures"})
	}

	return errors
}

// UpdateAdventureRequest represents a request to update an adventure
type UpdateAdventureRequest struct {
	Title           *string  `json:"title,omitempty"`
	Description     *string  `json:"description,omitempty"`
	StartDate       *string  `json:"start_date,omitempty"`
	EndDate         *string  `json:"end_date,omitempty"`
	Status          *string  `json:"status,omitempty"`
	Visibility      *string  `json:"visibility,omitempty"`
	ValuesRequired  *bool    `json:"values_required,omitempty"`
	ValuesQuestions []string `json:"values_questions,omitempty"`
	BudgetMin       *float64 `json:"budget_min,omitempty"`
	BudgetMax       *float64 `json:"budget_max,omitempty"`
	Currency        *string  `json:"currency,omitempty"`
	CoverImage      *string  `json:"cover_image,omitempty"`
	VotingOpen      *bool    `json:"voting_open,omitempty"`
	VotingDeadline  *string  `json:"voting_deadline,omitempty"`
}

// AddDestinationRequest represents a request to add a destination
type AddDestinationRequest struct {
	Name             string   `json:"name"`
	Description      *string  `json:"description,omitempty"`
	City             *string  `json:"city,omitempty"`
	Country          *string  `json:"country,omitempty"`
	Region           *string  `json:"region,omitempty"`
	EstimatedCostMin *float64 `json:"estimated_cost_min,omitempty"`
	EstimatedCostMax *float64 `json:"estimated_cost_max,omitempty"`
	InfoURL          *string  `json:"info_url,omitempty"`
	ImageURL         *string  `json:"image_url,omitempty"`
}

// VoteDestinationRequest represents a vote on a destination
type VoteDestinationRequest struct {
	Rank   int     `json:"rank"`             // 1-based ranking
	Veto   bool    `json:"veto,omitempty"`   // Hard no
	Reason *string `json:"reason,omitempty"` // If vetoing
}

// AddAdventureActivityRequest represents a request to add an activity
type AddAdventureActivityRequest struct {
	DestinationID *string  `json:"destination_id,omitempty"`
	Title         string   `json:"title"`
	Description   *string  `json:"description,omitempty"`
	ScheduledDate *string  `json:"scheduled_date,omitempty"` // RFC3339 date
	ScheduledTime *string  `json:"scheduled_time,omitempty"` // HH:MM
	Duration      *int     `json:"duration_minutes,omitempty"`
	Location      *string  `json:"location,omitempty"`
	EstimatedCost *float64 `json:"estimated_cost,omitempty"`
	BookingURL    *string  `json:"booking_url,omitempty"`
}

// UpdateAdventureActivityRequest represents a request to update an adventure activity
type UpdateAdventureActivityRequest struct {
	Title         *string  `json:"title,omitempty"`
	Description   *string  `json:"description,omitempty"`
	ScheduledDate *string  `json:"scheduled_date,omitempty"`
	ScheduledTime *string  `json:"scheduled_time,omitempty"`
	Duration      *int     `json:"duration_minutes,omitempty"`
	Location      *string  `json:"location,omitempty"`
	EstimatedCost *float64 `json:"estimated_cost,omitempty"`
	BookingURL    *string  `json:"booking_url,omitempty"`
	IsConfirmed   *bool    `json:"is_confirmed,omitempty"`
}

// UpdateParticipantStatusRequest updates a participant's status
type UpdateParticipantStatusRequest struct {
	Status string `json:"status"` // interested, committed, maybe, out
}

// AdventureWithDetails includes full adventure information
type AdventureWithDetails struct {
	Adventure    Adventure              `json:"adventure"`
	Participants []AdventureParticipant `json:"participants"`
	Destinations []Destination          `json:"destinations"`
	Activities   []AdventureActivity    `json:"activities"`
	Events       []EventSummary         `json:"events,omitempty"`
	Forum        *Forum                 `json:"forum,omitempty"`
}

// AdventureSummary provides a lightweight view of an adventure
type AdventureSummary struct {
	ID               string              `json:"id"`
	Title            string              `json:"title"`
	StartDate        time.Time           `json:"start_date"`
	EndDate          time.Time           `json:"end_date"`
	Status           AdventureStatus     `json:"status"`
	Visibility       AdventureVisibility `json:"visibility"`
	ParticipantCount int                 `json:"participant_count"`
	DestinationCount int                 `json:"destination_count"`
	EventCount       int                 `json:"event_count"`
	VotingOpen       bool                `json:"voting_open"`
}

// VotingResults contains aggregated voting results for an adventure
type VotingResults struct {
	AdventureID  string                   `json:"adventure_id"`
	Destinations []DestinationVoteSummary `json:"destinations"`
	TotalVoters  int                      `json:"total_voters"`
	VotingOpen   bool                     `json:"voting_open"`
}

// DestinationVoteSummary contains vote totals for a destination
type DestinationVoteSummary struct {
	DestinationID string  `json:"destination_id"`
	Name          string  `json:"name"`
	VoteCount     int     `json:"vote_count"`
	AverageRank   float64 `json:"average_rank"`
	VetoCount     int     `json:"veto_count"`
	// Rank distribution
	FirstChoiceCount  int `json:"first_choice_count"`
	SecondChoiceCount int `json:"second_choice_count"`
	ThirdChoiceCount  int `json:"third_choice_count"`
}

// AdventureAdmissionStatus represents the status of an admission request
type AdventureAdmissionStatus string

const (
	AdmissionStatusRequested AdventureAdmissionStatus = "requested"
	AdmissionStatusAdmitted  AdventureAdmissionStatus = "admitted"
	AdmissionStatusRejected  AdventureAdmissionStatus = "rejected"
)

// AdventureAdmissionRequestedBy indicates how the admission was initiated
type AdventureAdmissionRequestedBy string

const (
	AdmissionRequestedBySelf    AdventureAdmissionRequestedBy = "self"
	AdmissionRequestedByInvited AdventureAdmissionRequestedBy = "invited"
)

// AdventureAdmission represents a user's admission to view/RSVP to an adventure
// Note: This is separate from RSVP - admission grants permission to view and RSVP,
// but does not mean the user is committed to attend any events
type AdventureAdmission struct {
	ID              string                        `json:"id"`
	AdventureID     string                        `json:"adventure_id"`
	UserID          string                        `json:"user_id"`
	Status          AdventureAdmissionStatus      `json:"status"`       // requested, admitted, rejected
	RequestedBy     AdventureAdmissionRequestedBy `json:"requested_by"` // self or invited
	InvitedByID     *string                       `json:"invited_by_id,omitempty"`
	RejectionReason *string                       `json:"rejection_reason,omitempty"`
	RequestedOn     time.Time                     `json:"requested_on"`
	DecidedOn       *time.Time                    `json:"decided_on,omitempty"`
	// Populated by joins
	UserName *string `json:"user_name,omitempty"`
}

// Constraints for adventure admission
const (
	MaxAdmittedPerAdventure  = 500
	MaxRejectionReasonLength = 500
)

// RequestAdmissionRequest represents a request to join an adventure
type RequestAdmissionRequest struct {
	// Note is optional message with the request
	Note *string `json:"note,omitempty"`
}

// RespondToAdmissionRequest represents an organizer's response to admission request
type RespondToAdmissionRequest struct {
	Admit           bool    `json:"admit"`                      // true = admit, false = reject
	RejectionReason *string `json:"rejection_reason,omitempty"` // Required if rejecting
}

// Validate checks if the response is valid
func (r *RespondToAdmissionRequest) Validate() []FieldError {
	var errors []FieldError

	if !r.Admit && (r.RejectionReason == nil || *r.RejectionReason == "") {
		errors = append(errors, FieldError{Field: "rejection_reason", Message: "rejection_reason is required when rejecting"})
	}
	if r.RejectionReason != nil && len(*r.RejectionReason) > MaxRejectionReasonLength {
		errors = append(errors, FieldError{Field: "rejection_reason", Message: "rejection_reason must be 500 characters or less"})
	}

	return errors
}

// InviteToAdventureRequest represents a request to invite a user
type InviteToAdventureRequest struct {
	UserID string `json:"user_id"`
}

// Validate checks if the invite request is valid
func (r *InviteToAdventureRequest) Validate() []FieldError {
	var errors []FieldError

	if r.UserID == "" {
		errors = append(errors, FieldError{Field: "user_id", Message: "user_id is required"})
	}

	return errors
}

// TransferAdventureRequest represents a request to transfer organizer role
type TransferAdventureRequest struct {
	NewOrganizerUserID string `json:"new_organizer_user_id"` // Must be guild member for guild adventures
}

// Validate checks if the transfer request is valid
func (r *TransferAdventureRequest) Validate() []FieldError {
	var errors []FieldError

	if r.NewOrganizerUserID == "" {
		errors = append(errors, FieldError{Field: "new_organizer_user_id", Message: "new_organizer_user_id is required"})
	}

	return errors
}

// UnfreezeAdventureRequest represents a request to unfreeze an adventure
type UnfreezeAdventureRequest struct {
	NewOrganizerUserID string `json:"new_organizer_user_id"` // Must be current guild member
}

// Validate checks if the unfreeze request is valid
func (r *UnfreezeAdventureRequest) Validate() []FieldError {
	var errors []FieldError

	if r.NewOrganizerUserID == "" {
		errors = append(errors, FieldError{Field: "new_organizer_user_id", Message: "new_organizer_user_id is required"})
	}

	return errors
}

// AdventureAdmissionWithUser includes user details
type AdventureAdmissionWithUser struct {
	Admission   AdventureAdmission `json:"admission"`
	DisplayName string             `json:"display_name,omitempty"`
	Username    string             `json:"username,omitempty"`
}

// Trip is a backward compatibility alias for Adventure (deprecated).
type Trip = Adventure
type TripStatus = AdventureStatus
type TripParticipant = AdventureParticipant
type TripActivity = AdventureActivity
type TripWithDetails = AdventureWithDetails
type TripSummary = AdventureSummary
