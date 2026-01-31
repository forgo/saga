package model

import "time"

// VoteScopeType determines if the vote is guild-scoped or global
type VoteScopeType string

const (
	VoteScopeGuild  VoteScopeType = "guild"
	VoteScopeGlobal VoteScopeType = "global"
)

// VoteType represents the voting method
type VoteType string

const (
	VoteTypeFPTP         VoteType = "fptp"          // First Past The Post (plurality)
	VoteTypeRankedChoice VoteType = "ranked_choice" // Instant runoff
	VoteTypeApproval     VoteType = "approval"      // Vote for all you approve
	VoteTypeMultiSelect  VoteType = "multi_select"  // Select up to N options
)

// VoteStatus represents the lifecycle stage of a vote
type VoteStatus string

const (
	VoteStatusDraft     VoteStatus = "draft"     // Not yet open
	VoteStatusOpen      VoteStatus = "open"      // Accepting ballots
	VoteStatusClosed    VoteStatus = "closed"    // Voting ended
	VoteStatusCancelled VoteStatus = "cancelled" // Vote was cancelled
)

// ResultsVisibility determines when results are visible
type ResultsVisibility string

const (
	ResultsVisibilityLive       ResultsVisibility = "live"        // Results visible during voting
	ResultsVisibilityAfterClose ResultsVisibility = "after_close" // Results visible after close
	ResultsVisibilityAdminOnly  ResultsVisibility = "admin_only"  // Only admins can see results
)

// Vote represents a voting poll
type Vote struct {
	ID                   string            `json:"id"`
	ScopeType            VoteScopeType     `json:"scope_type"`              // guild or global
	ScopeID              *string           `json:"scope_id,omitempty"`      // Guild ID for guild votes
	CreatedBy            string            `json:"created_by"`              // User ID
	Title                string            `json:"title"`
	Description          *string           `json:"description,omitempty"`
	VoteType             VoteType          `json:"vote_type"`               // fptp, ranked_choice, approval, multi_select
	OpensAt              time.Time         `json:"opens_at"`
	ClosesAt             time.Time         `json:"closes_at"`
	Status               VoteStatus        `json:"status"`
	ResultsVisibility    ResultsVisibility `json:"results_visibility"`
	MaxOptionsSelectable *int              `json:"max_options_selectable,omitempty"` // For multi_select
	AllowAbstain         bool              `json:"allow_abstain"`
	CreatedOn            time.Time         `json:"created_on"`
	UpdatedOn            time.Time         `json:"updated_on"`
	// Computed fields
	OptionCount  int `json:"option_count,omitempty"`
	BallotCount  int `json:"ballot_count,omitempty"`
	VoterCount   int `json:"voter_count,omitempty"`
}

// VoteOption represents a choice in a vote
type VoteOption struct {
	ID                string    `json:"id"`
	VoteID            string    `json:"vote_id"`
	OptionText        string    `json:"option_text"`
	OptionDescription *string   `json:"option_description,omitempty"`
	SortOrder         int       `json:"sort_order"`
	CreatedBy         string    `json:"created_by"` // User ID
	CreatedOn         time.Time `json:"created_on"`
	// Computed fields (for results)
	VoteCount int `json:"vote_count,omitempty"`
}

// VoterSnapshot captures voter identity at time of voting
type VoterSnapshot struct {
	Username    string `json:"username,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

// BallotData represents the actual vote cast (varies by vote type)
// FPTP: {"option_id": "vote_option:xxx"}
// Ranked Choice: {"rankings": ["vote_option:a", "vote_option:b", "vote_option:c"]}
// Approval: {"approved_options": ["vote_option:a", "vote_option:b"]}
// Multi-select: {"selected_options": ["vote_option:a", "vote_option:b"]}
type BallotData map[string]interface{}

// VoteBallot represents a voter's ballot (immutable after creation)
type VoteBallot struct {
	ID            string        `json:"id"`
	VoteID        string        `json:"vote_id"`
	VoterUserID   string        `json:"voter_user_id"`
	VoterSnapshot VoterSnapshot `json:"voter_snapshot"`
	BallotData    BallotData    `json:"ballot_data"`
	IsAbstain     bool          `json:"is_abstain"`
	CreatedOn     time.Time     `json:"created_on"`
}

// VoteWithOptions includes the vote and all its options
type VoteWithOptions struct {
	Vote    Vote         `json:"vote"`
	Options []VoteOption `json:"options"`
}

// VoteWithDetails includes full vote information including ballots
type VoteWithDetails struct {
	Vote       Vote         `json:"vote"`
	Options    []VoteOption `json:"options"`
	Ballots    []VoteBallot `json:"ballots,omitempty"`    // Only if results visible
	MyBallot   *VoteBallot  `json:"my_ballot,omitempty"`  // Current user's ballot
	HasVoted   bool         `json:"has_voted"`
	CanVote    bool         `json:"can_vote"`
}

// VoteResult represents the computed results of a vote
type VoteResult struct {
	VoteID        string              `json:"vote_id"`
	VoteType      VoteType            `json:"vote_type"`
	TotalBallots  int                 `json:"total_ballots"`
	TotalAbstains int                 `json:"total_abstains"`
	OptionResults []OptionResult      `json:"option_results"`
	Winner        *string             `json:"winner,omitempty"` // Option ID of winner (if any)
	RoundDetails  []RoundDetail       `json:"round_details,omitempty"` // For ranked choice
}

// OptionResult contains results for a single option
type OptionResult struct {
	OptionID    string  `json:"option_id"`
	OptionText  string  `json:"option_text"`
	VoteCount   int     `json:"vote_count"`
	Percentage  float64 `json:"percentage"`
	Rank        int     `json:"rank"` // 1 = first place
	IsWinner    bool    `json:"is_winner"`
	IsEliminated bool   `json:"is_eliminated,omitempty"` // For ranked choice
}

// RoundDetail contains details of each round in ranked choice voting
type RoundDetail struct {
	Round           int            `json:"round"`
	OptionCounts    map[string]int `json:"option_counts"` // option_id -> count
	EliminatedID    *string        `json:"eliminated_id,omitempty"`
	EliminatedCount int            `json:"eliminated_count,omitempty"`
}

// Constraints
const (
	MaxVoteTitleLength       = 200
	MaxVoteDescriptionLength = 2000
	MaxOptionsPerVote        = 20
	MaxOptionTextLength      = 200
	MaxOptionDescLength      = 500
	MaxActiveVotesPerGuild   = 50
)

// CreateVoteRequest represents a request to create a vote
type CreateVoteRequest struct {
	ScopeType            string  `json:"scope_type"`                        // guild or global
	ScopeID              *string `json:"scope_id,omitempty"`                // Guild ID for guild votes
	Title                string  `json:"title"`
	Description          *string `json:"description,omitempty"`
	VoteType             string  `json:"vote_type"`                         // fptp, ranked_choice, approval, multi_select
	OpensAt              string  `json:"opens_at"`                          // RFC3339 datetime
	ClosesAt             string  `json:"closes_at"`                         // RFC3339 datetime
	ResultsVisibility    *string `json:"results_visibility,omitempty"`      // live, after_close, admin_only
	MaxOptionsSelectable *int    `json:"max_options_selectable,omitempty"`  // For multi_select
	AllowAbstain         bool    `json:"allow_abstain,omitempty"`
}

// Validate checks if the create request is valid
func (r *CreateVoteRequest) Validate() []FieldError {
	var errors []FieldError

	if r.ScopeType == "" {
		errors = append(errors, FieldError{Field: "scope_type", Message: "scope_type is required"})
	} else if r.ScopeType != string(VoteScopeGuild) && r.ScopeType != string(VoteScopeGlobal) {
		errors = append(errors, FieldError{Field: "scope_type", Message: "scope_type must be 'guild' or 'global'"})
	}
	if r.ScopeType == string(VoteScopeGuild) && (r.ScopeID == nil || *r.ScopeID == "") {
		errors = append(errors, FieldError{Field: "scope_id", Message: "scope_id is required for guild votes"})
	}
	if r.Title == "" {
		errors = append(errors, FieldError{Field: "title", Message: "title is required"})
	} else if len(r.Title) > MaxVoteTitleLength {
		errors = append(errors, FieldError{Field: "title", Message: "title must be 200 characters or less"})
	}
	if r.Description != nil && len(*r.Description) > MaxVoteDescriptionLength {
		errors = append(errors, FieldError{Field: "description", Message: "description must be 2000 characters or less"})
	}
	if r.VoteType == "" {
		errors = append(errors, FieldError{Field: "vote_type", Message: "vote_type is required"})
	} else {
		validTypes := map[string]bool{
			string(VoteTypeFPTP): true, string(VoteTypeRankedChoice): true,
			string(VoteTypeApproval): true, string(VoteTypeMultiSelect): true,
		}
		if !validTypes[r.VoteType] {
			errors = append(errors, FieldError{Field: "vote_type", Message: "vote_type must be fptp, ranked_choice, approval, or multi_select"})
		}
	}
	if r.OpensAt == "" {
		errors = append(errors, FieldError{Field: "opens_at", Message: "opens_at is required"})
	}
	if r.ClosesAt == "" {
		errors = append(errors, FieldError{Field: "closes_at", Message: "closes_at is required"})
	}
	if r.ResultsVisibility != nil {
		validVisibility := map[string]bool{
			string(ResultsVisibilityLive): true, string(ResultsVisibilityAfterClose): true,
			string(ResultsVisibilityAdminOnly): true,
		}
		if !validVisibility[*r.ResultsVisibility] {
			errors = append(errors, FieldError{Field: "results_visibility", Message: "results_visibility must be live, after_close, or admin_only"})
		}
	}
	if r.VoteType == string(VoteTypeMultiSelect) && r.MaxOptionsSelectable != nil && *r.MaxOptionsSelectable < 1 {
		errors = append(errors, FieldError{Field: "max_options_selectable", Message: "max_options_selectable must be at least 1"})
	}

	return errors
}

// UpdateVoteRequest represents a request to update a vote (only when draft)
type UpdateVoteRequest struct {
	Title                *string `json:"title,omitempty"`
	Description          *string `json:"description,omitempty"`
	OpensAt              *string `json:"opens_at,omitempty"`
	ClosesAt             *string `json:"closes_at,omitempty"`
	ResultsVisibility    *string `json:"results_visibility,omitempty"`
	MaxOptionsSelectable *int    `json:"max_options_selectable,omitempty"`
	AllowAbstain         *bool   `json:"allow_abstain,omitempty"`
}

// Validate checks if the update request is valid
func (r *UpdateVoteRequest) Validate() []FieldError {
	var errors []FieldError

	if r.Title != nil {
		if *r.Title == "" {
			errors = append(errors, FieldError{Field: "title", Message: "title cannot be empty"})
		} else if len(*r.Title) > MaxVoteTitleLength {
			errors = append(errors, FieldError{Field: "title", Message: "title must be 200 characters or less"})
		}
	}
	if r.Description != nil && len(*r.Description) > MaxVoteDescriptionLength {
		errors = append(errors, FieldError{Field: "description", Message: "description must be 2000 characters or less"})
	}
	if r.ResultsVisibility != nil {
		validVisibility := map[string]bool{
			string(ResultsVisibilityLive): true, string(ResultsVisibilityAfterClose): true,
			string(ResultsVisibilityAdminOnly): true,
		}
		if !validVisibility[*r.ResultsVisibility] {
			errors = append(errors, FieldError{Field: "results_visibility", Message: "results_visibility must be live, after_close, or admin_only"})
		}
	}

	return errors
}

// CreateVoteOptionRequest represents a request to add an option
type CreateVoteOptionRequest struct {
	OptionText        string  `json:"option_text"`
	OptionDescription *string `json:"option_description,omitempty"`
	SortOrder         *int    `json:"sort_order,omitempty"`
}

// Validate checks if the request is valid
func (r *CreateVoteOptionRequest) Validate() []FieldError {
	var errors []FieldError

	if r.OptionText == "" {
		errors = append(errors, FieldError{Field: "option_text", Message: "option_text is required"})
	} else if len(r.OptionText) > MaxOptionTextLength {
		errors = append(errors, FieldError{Field: "option_text", Message: "option_text must be 200 characters or less"})
	}
	if r.OptionDescription != nil && len(*r.OptionDescription) > MaxOptionDescLength {
		errors = append(errors, FieldError{Field: "option_description", Message: "option_description must be 500 characters or less"})
	}

	return errors
}

// UpdateVoteOptionRequest represents a request to update an option
type UpdateVoteOptionRequest struct {
	OptionText        *string `json:"option_text,omitempty"`
	OptionDescription *string `json:"option_description,omitempty"`
	SortOrder         *int    `json:"sort_order,omitempty"`
}

// Validate checks if the update request is valid
func (r *UpdateVoteOptionRequest) Validate() []FieldError {
	var errors []FieldError

	if r.OptionText != nil {
		if *r.OptionText == "" {
			errors = append(errors, FieldError{Field: "option_text", Message: "option_text cannot be empty"})
		} else if len(*r.OptionText) > MaxOptionTextLength {
			errors = append(errors, FieldError{Field: "option_text", Message: "option_text must be 200 characters or less"})
		}
	}
	if r.OptionDescription != nil && len(*r.OptionDescription) > MaxOptionDescLength {
		errors = append(errors, FieldError{Field: "option_description", Message: "option_description must be 500 characters or less"})
	}

	return errors
}

// CastBallotRequest represents a request to cast a vote
type CastBallotRequest struct {
	// For FPTP: single option ID
	OptionID *string `json:"option_id,omitempty"`
	// For Ranked Choice: ordered list of option IDs (first = most preferred)
	Rankings []string `json:"rankings,omitempty"`
	// For Approval/Multi-select: list of approved/selected option IDs
	SelectedOptions []string `json:"selected_options,omitempty"`
	// For abstaining
	IsAbstain bool `json:"is_abstain,omitempty"`
}

// Validate checks if the ballot is valid for the given vote type
func (r *CastBallotRequest) ValidateForVoteType(voteType VoteType, maxSelectable *int, allowAbstain bool) []FieldError {
	var errors []FieldError

	if r.IsAbstain {
		if !allowAbstain {
			errors = append(errors, FieldError{Field: "is_abstain", Message: "abstaining is not allowed for this vote"})
		}
		return errors
	}

	switch voteType {
	case VoteTypeFPTP:
		if r.OptionID == nil || *r.OptionID == "" {
			errors = append(errors, FieldError{Field: "option_id", Message: "option_id is required for FPTP voting"})
		}
	case VoteTypeRankedChoice:
		if len(r.Rankings) == 0 {
			errors = append(errors, FieldError{Field: "rankings", Message: "rankings are required for ranked choice voting"})
		}
	case VoteTypeApproval:
		if len(r.SelectedOptions) == 0 {
			errors = append(errors, FieldError{Field: "selected_options", Message: "selected_options are required for approval voting"})
		}
	case VoteTypeMultiSelect:
		if len(r.SelectedOptions) == 0 {
			errors = append(errors, FieldError{Field: "selected_options", Message: "selected_options are required for multi-select voting"})
		}
		if maxSelectable != nil && len(r.SelectedOptions) > *maxSelectable {
			errors = append(errors, FieldError{Field: "selected_options", Message: "too many options selected"})
		}
	}

	return errors
}

// ToBallotData converts the request to ballot data based on vote type
func (r *CastBallotRequest) ToBallotData(voteType VoteType) BallotData {
	if r.IsAbstain {
		return BallotData{}
	}

	switch voteType {
	case VoteTypeFPTP:
		return BallotData{"option_id": *r.OptionID}
	case VoteTypeRankedChoice:
		return BallotData{"rankings": r.Rankings}
	case VoteTypeApproval:
		return BallotData{"approved_options": r.SelectedOptions}
	case VoteTypeMultiSelect:
		return BallotData{"selected_options": r.SelectedOptions}
	default:
		return BallotData{}
	}
}

// VoteSummary provides a lightweight view of a vote
type VoteSummary struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	VoteType    VoteType      `json:"vote_type"`
	ScopeType   VoteScopeType `json:"scope_type"`
	Status      VoteStatus    `json:"status"`
	OpensAt     time.Time     `json:"opens_at"`
	ClosesAt    time.Time     `json:"closes_at"`
	OptionCount int           `json:"option_count"`
	BallotCount int           `json:"ballot_count"`
	HasVoted    bool          `json:"has_voted"`
}
