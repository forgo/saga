package model

import "time"

// ResonanceStat represents the different scoring categories
type ResonanceStat string

const (
	ResonanceStatQuesting   ResonanceStat = "questing"   // Verified follow-through
	ResonanceStatMana       ResonanceStat = "mana"       // Support that landed
	ResonanceStatWayfinder  ResonanceStat = "wayfinder"  // Hosting that happened
	ResonanceStatAttunement ResonanceStat = "attunement" // Profile clarity
	ResonanceStatNexus      ResonanceStat = "nexus"      // Active circles + bridging
)

// ResonanceLedger represents an immutable ledger entry for point awards
type ResonanceLedger struct {
	ID             string        `json:"id"`
	UserID         string        `json:"user_id"`
	Stat           ResonanceStat `json:"stat"`
	Points         int           `json:"points"`
	SourceObjectID string        `json:"source_object_id"` // event:xyz, question:abc, month:2026-01
	ReasonCode     string        `json:"reason_code"`
	CreatedOn      time.Time     `json:"created_on"`
}

// ReasonCode constants for auditing
const (
	// Questing reasons
	ReasonQuestingCompletion   = "COMPLETION"
	ReasonQuestingEarlyConfirm = "EARLY_CONFIRM"
	ReasonQuestingCheckinBonus = "CHECKIN_BONUS"

	// Mana reasons
	ReasonManaSupport       = "SUPPORT_HELPFUL"
	ReasonManaEarlyConfirm  = "MANA_EARLY_CONFIRM"
	ReasonManaFeedbackTag   = "MANA_FEEDBACK_TAG"

	// Wayfinder reasons
	ReasonWayfinderHosting   = "HOSTING_VERIFIED"
	ReasonWayfinderAttendees = "ATTENDEES"
	ReasonWayfinderEarly     = "WAYFINDER_EARLY"

	// Attunement reasons
	ReasonAttunementQuestion      = "QUESTION_ANSWERED"
	ReasonAttunementProfileRefresh = "PROFILE_REFRESH"

	// Nexus reasons
	ReasonNexusMonthly = "NEXUS_MONTHLY"
)

// ResonanceScore represents cached totals for a user
type ResonanceScore struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Total          int       `json:"total"`
	Questing       int       `json:"questing"`
	Mana           int       `json:"mana"`
	Wayfinder      int       `json:"wayfinder"`
	Attunement     int       `json:"attunement"`
	Nexus          int       `json:"nexus"`
	LastCalculated time.Time `json:"last_calculated"`
}

// ResonanceDailyCap tracks daily earning limits
type ResonanceDailyCap struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	Date             string `json:"date"` // "2026-01-06"
	QuestingEarned   int    `json:"questing_earned"`
	ManaEarned       int    `json:"mana_earned"`
	WayfinderEarned  int    `json:"wayfinder_earned"`
	AttunementEarned int    `json:"attunement_earned"`
}

// Daily caps (configurable)
const (
	DailyCapQuesting   = 40
	DailyCapMana       = 32
	DailyCapWayfinder  = 30
	DailyCapAttunement = 20
	MonthlyCapNexus    = 200
)

// SupportPairCount tracks pairwise support sessions for diminishing returns
type SupportPairCount struct {
	ID          string    `json:"id"`
	HelperID    string    `json:"helper_id"`
	ReceiverID  string    `json:"receiver_id"`
	Count       int       `json:"count"`       // Sessions in last 30 days
	LastSession time.Time `json:"last_session"`
}

// GetManaMultiplier returns the multiplier based on pairwise session count
func GetManaMultiplier(count int) float64 {
	switch {
	case count <= 2:
		return 1.0 // Sessions 1-3
	case count <= 5:
		return 0.5 // Sessions 4-6
	default:
		return 0.25 // Sessions 7+
	}
}

// Point values (configurable)
const (
	// Questing
	PointsQuestingBase        = 10
	PointsQuestingEarlyConfirm = 2
	PointsQuestingCheckin     = 2

	// Mana
	PointsManaBase         = 12
	PointsManaEarlyConfirm = 2
	PointsManaFeedbackTag  = 2

	// Wayfinder
	PointsWayfinderBase        = 8
	PointsWayfinderPerAttendee = 2
	PointsWayfinderMaxAttendees = 4
	PointsWayfinderEarlyConfirm = 2

	// Attunement
	PointsAttunementQuestion       = 2
	PointsAttunementProfileRefresh = 10
)

// Time windows (configurable)
const (
	ConfirmWindowHours   = 48
	CheckinWindowMinutes = 10
	EarlyConfirmHours    = 2
	NexusWindowDays      = 30
)

// HelpfulnessRating for support sessions
type HelpfulnessRating string

const (
	HelpfulnessYes       HelpfulnessRating = "YES"
	HelpfulnessSomewhat  HelpfulnessRating = "SOMEWHAT"
	HelpfulnessNotReally HelpfulnessRating = "NOT_REALLY"
	HelpfulnessSkip      HelpfulnessRating = "SKIP"
)

// IsHelpful returns true if the rating qualifies for Mana
func (r HelpfulnessRating) IsHelpful() bool {
	return r == HelpfulnessYes || r == HelpfulnessSomewhat
}

// Note: EventParticipant fields are now part of EventRSVP in event.go
// (CompletionConfirmed, CheckinTime, HelpfulnessRating, HelpfulnessTags)

// ResonanceDisplay is what users see on profiles
type ResonanceDisplay struct {
	Total      int    `json:"total"`
	Questing   int    `json:"questing"`
	Mana       int    `json:"mana"`
	Wayfinder  int    `json:"wayfinder"`
	Attunement int    `json:"attunement"`
	Nexus      int    `json:"nexus"`
}

// ResonanceLedgerEntry is used for creating and reading ledger entries
type ResonanceLedgerEntry struct {
	ID             string        `json:"id,omitempty"`
	UserID         string        `json:"user_id,omitempty"`
	Stat           ResonanceStat `json:"stat"`
	Points         int           `json:"points"`
	SourceObjectID string        `json:"source_object_id,omitempty"` // event:xyz, question:abc, month:2026-01
	ReasonCode     string        `json:"reason_code,omitempty"`
	Description    string        `json:"description,omitempty"` // Human-readable
	CreatedOn      time.Time     `json:"created_on"`
}

// CircleNexusContribution represents a circle's contribution to Nexus score
type CircleNexusContribution struct {
	CircleID       string  `json:"circle_id"`
	CircleName     string  `json:"circle_name"`
	Points         int     `json:"points"`
	ActivityFactor float64 `json:"activity_factor"` // 0-1 based on participation
	ActiveMembers  int     `json:"active_members"`
}

// NexusCircleData represents circle activity data for Nexus calculation
type NexusCircleData struct {
	CircleID        string
	CircleName      string
	ActiveMembers   int     // Members who participated in events in last 30 days
	TotalEvents     int     // Verified events in last 30 days
	UserCompletions int     // User's verified event completions in this circle
	ActivityFactor  float64 // min(1, completions / 3)
	IsActive        bool    // Has ≥2 events AND ≥3 active members
}

// ConfirmEventRequest is used to mark an event as complete
type ConfirmEventRequest struct {
	EventID string `json:"event_id"`
}

// CheckinRequest is used for on-time checkin
type CheckinRequest struct {
	EventID string `json:"event_id"`
}

// HelpfulnessFeedbackRequest is used for support session feedback
type HelpfulnessFeedbackRequest struct {
	EventID string   `json:"event_id"`
	Rating  string   `json:"rating"` // YES, SOMEWHAT, NOT_REALLY, SKIP
	Tags    []string `json:"tags,omitempty"`
}
