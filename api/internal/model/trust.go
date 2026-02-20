package model

import "time"

// TrustRelation represents trust status between two users
type TrustRelation struct {
	ID        string    `json:"id"`
	UserAID   string    `json:"user_a_id"` // The user who granted trust
	UserBID   string    `json:"user_b_id"` // The user being trusted
	Status    string    `json:"status"`    // pending, active, revoked
	CreatedOn time.Time `json:"created_on"`
	UpdatedOn time.Time `json:"updated_on"`
}

// TrustStatus constants
const (
	TrustStatusPending = "pending" // Trust requested but not confirmed
	TrustStatusActive  = "active"  // Mutual trust established
	TrustStatusRevoked = "revoked" // Trust was revoked by one party
)

// IRLVerification represents confirmed in-person interaction
type IRLVerification struct {
	ID          string    `json:"id"`
	UserAID     string    `json:"user_a_id"`
	UserBID     string    `json:"user_b_id"`
	VerifiedOn  time.Time `json:"verified_on"`
	Context     string    `json:"context"`                // event, hangout, introduced
	ReferenceID *string   `json:"reference_id,omitempty"` // event_id if from event
	// Both users must confirm
	UserAConfirmed   bool       `json:"user_a_confirmed"`
	UserBConfirmed   bool       `json:"user_b_confirmed"`
	UserAConfirmedOn *time.Time `json:"user_a_confirmed_on,omitempty"`
	UserBConfirmedOn *time.Time `json:"user_b_confirmed_on,omitempty"`
}

// IRLContext constants
const (
	IRLContextEvent      = "event"      // Met at an event
	IRLContextHangout    = "hangout"    // Met through availability hangout
	IRLContextIntroduced = "introduced" // Introduced by mutual friend
	IRLContextOther      = "other"      // Met outside the app
)

// TrustSummary provides a summary of trust status between two users
type TrustSummary struct {
	UserAID      string `json:"user_a_id"`
	UserBID      string `json:"user_b_id"`
	IRLConfirmed bool   `json:"irl_confirmed"` // Both confirmed IRL meeting
	MutualTrust  bool   `json:"mutual_trust"`  // Both trust each other
	CanCommute   bool   `json:"can_commute"`   // Unlocked commute features
	TrustLevel   string `json:"trust_level"`   // none, irl_only, trusted
}

// TrustLevel constants
const (
	TrustLevelNone    = "none"     // No verified interaction
	TrustLevelIRL     = "irl_only" // IRL confirmed but not trusted
	TrustLevelTrusted = "trusted"  // IRL confirmed + mutual trust
)

// UserTrustProfile provides trust info for a user
type UserTrustProfile struct {
	UserID             string   `json:"user_id"`
	IRLConnectionCount int      `json:"irl_connection_count"`
	TrustedByCount     int      `json:"trusted_by_count"`
	TrustsCount        int      `json:"trusts_count"`
	MutualTrustCount   int      `json:"mutual_trust_count"`
	CanOfferCommute    bool     `json:"can_offer_commute"` // Has enough trust to offer rides
	TrustedUserIDs     []string `json:"-"`                 // Internal use only
}

// Trust requirements for features
const (
	// MinTrustForCommute: minimum mutual trust connections to offer rides
	MinTrustForCommute = 1
	// MinIRLForCommute: minimum IRL verifications to participate in commute
	MinIRLForCommute = 1
)

// RequestTrustRequest represents a request to establish trust
type RequestTrustRequest struct {
	UserID string `json:"user_id"` // User to request trust from
}

// ConfirmIRLRequest represents a request to confirm IRL meeting
type ConfirmIRLRequest struct {
	UserID      string  `json:"user_id"`                // User you met
	Context     string  `json:"context"`                // event, hangout, introduced, other
	ReferenceID *string `json:"reference_id,omitempty"` // Event ID if applicable
}

// TrustAction represents actions on trust relations
type TrustAction string

const (
	TrustActionGrant  TrustAction = "grant"
	TrustActionRevoke TrustAction = "revoke"
)

// UpdateTrustRequest represents a request to update trust status
type UpdateTrustRequest struct {
	Action TrustAction `json:"action"` // grant or revoke
}

// TrustedUser represents a user in the trust network
type TrustedUser struct {
	UserID       string    `json:"user_id"`
	DisplayName  string    `json:"display_name,omitempty"`
	IRLConfirmed bool      `json:"irl_confirmed"`
	TrustStatus  string    `json:"trust_status"` // i_trust_them, they_trust_me, mutual
	Since        time.Time `json:"since"`
}

// TrustDirection constants for display
const (
	TrustDirectionITrustThem  = "i_trust_them"
	TrustDirectionTheyTrustMe = "they_trust_me"
	TrustDirectionMutual      = "mutual"
)

// PendingTrustRequest represents a pending trust request
type PendingTrustRequest struct {
	ID          string    `json:"id"`
	FromUserID  string    `json:"from_user_id"`
	ToUserID    string    `json:"to_user_id"`
	RequestedOn time.Time `json:"requested_on"`
	// Populated
	FromUserName *string `json:"from_user_name,omitempty"`
}
