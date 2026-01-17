package model

import "time"

// TrustLevel represents the trust assessment of a user
type TrustLevel string

const (
	TrustLevelTrust    TrustLevel = "trust"
	TrustLevelDistrust TrustLevel = "distrust"
)

// TrustAnchorType represents what interaction anchors the trust rating
type TrustAnchorType string

const (
	TrustAnchorEvent    TrustAnchorType = "event"
	TrustAnchorRideshare TrustAnchorType = "rideshare"
)

// ReviewVisibility determines who can see the trust review
type ReviewVisibility string

const (
	ReviewVisibilityPublic    ReviewVisibility = "public"
	ReviewVisibilityAdminOnly ReviewVisibility = "admin_only"
)

// TrustRating represents an event-anchored trust assessment between users
type TrustRating struct {
	ID               string           `json:"id"`
	RaterID          string           `json:"rater_id"`           // User who is rating
	RateeID          string           `json:"ratee_id"`           // User being rated
	AnchorType       TrustAnchorType  `json:"anchor_type"`        // event or rideshare
	AnchorID         string           `json:"anchor_id"`          // ID of the event/rideshare
	TrustLevel       TrustLevel       `json:"trust_level"`        // trust or distrust
	TrustReview      string           `json:"trust_review"`       // Required review (max 240 chars)
	ReviewVisibility ReviewVisibility `json:"review_visibility"`  // public or admin_only
	CreatedOn        time.Time        `json:"created_on"`
	UpdatedOn        time.Time        `json:"updated_on"`
	// Computed fields
	EndorsementCount int  `json:"endorsement_count,omitempty"`
	AgreeCount       int  `json:"agree_count,omitempty"`
	DisagreeCount    int  `json:"disagree_count,omitempty"`
	// Cooldown info
	CanEdit          bool       `json:"can_edit,omitempty"`
	NextEditableAt   *time.Time `json:"next_editable_at,omitempty"`
}

// TrustRatingHistory represents an audit entry for trust rating changes
type TrustRatingHistory struct {
	ID             string     `json:"id"`
	TrustRatingID  string     `json:"trust_rating_id"`
	PreviousLevel  *string    `json:"previous_level,omitempty"`
	NewLevel       string     `json:"new_level"`
	PreviousReview *string    `json:"previous_review,omitempty"`
	NewReview      string     `json:"new_review"`
	ChangedOn      time.Time  `json:"changed_on"`
}

// EndorsementType represents whether an endorser agrees or disagrees
type EndorsementType string

const (
	EndorsementAgree    EndorsementType = "agree"
	EndorsementDisagree EndorsementType = "disagree"
)

// TrustEndorsement represents a co-attendee's endorsement of a trust rating
type TrustEndorsement struct {
	ID              string          `json:"id"`
	TrustRatingID   string          `json:"trust_rating_id"`
	EndorserID      string          `json:"endorser_id"`
	EndorsementType EndorsementType `json:"endorsement_type"` // agree or disagree
	Note            *string         `json:"note,omitempty"`   // Optional note (max 240 chars)
	CreatedOn       time.Time       `json:"created_on"`
}

// TrustRatingDailyCount tracks daily rating count for throttling
type TrustRatingDailyCount struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Date   string `json:"date"` // YYYY-MM-DD format
	Count  int    `json:"count"`
}

// TrustAggregate provides aggregated trust stats for a user
type TrustAggregate struct {
	UserID           string `json:"user_id"`
	TrustCount       int    `json:"trust_count"`
	DistrustCount    int    `json:"distrust_count"`
	EndorsementCount int    `json:"endorsement_count"`
	NetTrust         int    `json:"net_trust"` // trust_count - distrust_count
}

// TrustRatingWithContext includes anchor context
type TrustRatingWithContext struct {
	TrustRating   TrustRating `json:"trust_rating"`
	AnchorTitle   string      `json:"anchor_title,omitempty"`   // Event/rideshare title
	AnchorDate    *time.Time  `json:"anchor_date,omitempty"`    // When it occurred
	RaterName     string      `json:"rater_name,omitempty"`     // Display name of rater
}

// Constraints
const (
	MaxTrustReviewLength     = 240
	MaxTrustRatingsPerDay    = 10
	TrustRatingCooldownDays  = 30
	MaxEndorsementNoteLength = 240
)

// CreateTrustRatingRequest represents a request to create a trust rating
type CreateTrustRatingRequest struct {
	RateeID     string `json:"ratee_id"`     // User being rated
	AnchorType  string `json:"anchor_type"`  // event or rideshare
	AnchorID    string `json:"anchor_id"`    // ID of the event/rideshare
	TrustLevel  string `json:"trust_level"`  // trust or distrust
	TrustReview string `json:"trust_review"` // Required review
}

// Validate checks if the create request is valid
func (r *CreateTrustRatingRequest) Validate() []FieldError {
	var errors []FieldError

	if r.RateeID == "" {
		errors = append(errors, FieldError{Field: "ratee_id", Message: "ratee_id is required"})
	}
	if r.AnchorType == "" {
		errors = append(errors, FieldError{Field: "anchor_type", Message: "anchor_type is required"})
	} else if r.AnchorType != string(TrustAnchorEvent) && r.AnchorType != string(TrustAnchorRideshare) {
		errors = append(errors, FieldError{Field: "anchor_type", Message: "anchor_type must be 'event' or 'rideshare'"})
	}
	if r.AnchorID == "" {
		errors = append(errors, FieldError{Field: "anchor_id", Message: "anchor_id is required"})
	}
	if r.TrustLevel == "" {
		errors = append(errors, FieldError{Field: "trust_level", Message: "trust_level is required"})
	} else if r.TrustLevel != string(TrustLevelTrust) && r.TrustLevel != string(TrustLevelDistrust) {
		errors = append(errors, FieldError{Field: "trust_level", Message: "trust_level must be 'trust' or 'distrust'"})
	}
	if r.TrustReview == "" {
		errors = append(errors, FieldError{Field: "trust_review", Message: "trust_review is required"})
	} else if len(r.TrustReview) > MaxTrustReviewLength {
		errors = append(errors, FieldError{Field: "trust_review", Message: "trust_review must be 240 characters or less"})
	}

	return errors
}

// UpdateTrustRatingRequest represents a request to update a trust rating
type UpdateTrustRatingRequest struct {
	TrustLevel  *string `json:"trust_level,omitempty"`
	TrustReview *string `json:"trust_review,omitempty"`
}

// Validate checks if the update request is valid
func (r *UpdateTrustRatingRequest) Validate() []FieldError {
	var errors []FieldError

	if r.TrustLevel != nil {
		if *r.TrustLevel != string(TrustLevelTrust) && *r.TrustLevel != string(TrustLevelDistrust) {
			errors = append(errors, FieldError{Field: "trust_level", Message: "trust_level must be 'trust' or 'distrust'"})
		}
	}
	if r.TrustReview != nil {
		if *r.TrustReview == "" {
			errors = append(errors, FieldError{Field: "trust_review", Message: "trust_review cannot be empty"})
		} else if len(*r.TrustReview) > MaxTrustReviewLength {
			errors = append(errors, FieldError{Field: "trust_review", Message: "trust_review must be 240 characters or less"})
		}
	}

	return errors
}

// CreateEndorsementRequest represents a request to endorse a trust rating
type CreateEndorsementRequest struct {
	EndorsementType string  `json:"endorsement_type"` // agree or disagree
	Note            *string `json:"note,omitempty"`
}

// Validate checks if the endorsement request is valid
func (r *CreateEndorsementRequest) Validate() []FieldError {
	var errors []FieldError

	if r.EndorsementType == "" {
		errors = append(errors, FieldError{Field: "endorsement_type", Message: "endorsement_type is required"})
	} else if r.EndorsementType != string(EndorsementAgree) && r.EndorsementType != string(EndorsementDisagree) {
		errors = append(errors, FieldError{Field: "endorsement_type", Message: "endorsement_type must be 'agree' or 'disagree'"})
	}
	if r.Note != nil && len(*r.Note) > MaxEndorsementNoteLength {
		errors = append(errors, FieldError{Field: "note", Message: "note must be 240 characters or less"})
	}

	return errors
}

// DistrustSignal represents a user with significant distrust for admin review
type DistrustSignal struct {
	UserID        string    `json:"user_id"`
	Username      string    `json:"username,omitempty"`
	DistrustCount int       `json:"distrust_count"`
	TrustCount    int       `json:"trust_count"`
	NetTrust      int       `json:"net_trust"`
	LatestReason  string    `json:"latest_reason,omitempty"` // Most recent distrust review
	LatestRatingOn time.Time `json:"latest_rating_on"`
}
