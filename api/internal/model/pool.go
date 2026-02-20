package model

import "time"

// MatchingPool represents a Donut-style matching pool within a guild
type MatchingPool struct {
	ID                 string     `json:"id"`
	GuildID            string     `json:"guild_id"`
	Name               string     `json:"name"`
	Description        *string    `json:"description,omitempty"`
	Frequency          string     `json:"frequency"`  // weekly, biweekly, monthly
	MatchSize          int        `json:"match_size"` // 2 for pairs, 3+ for groups
	ActivitySuggestion *string    `json:"activity_suggestion,omitempty"`
	NextMatchOn        time.Time  `json:"next_match_on"`
	LastMatchOn        *time.Time `json:"last_match_on,omitempty"`
	Active             bool       `json:"active"`
	CreatedBy          string     `json:"created_by"` // Member ID
	CreatedOn          time.Time  `json:"created_on"`
	UpdatedOn          time.Time  `json:"updated_on"`
	// Computed fields
	MemberCount int `json:"member_count,omitempty"`
}

// PoolFrequency constants
const (
	PoolFrequencyWeekly   = "weekly"
	PoolFrequencyBiweekly = "biweekly"
	PoolFrequencyMonthly  = "monthly"
)

// GetNextMatchDate calculates the next match date based on frequency
func GetNextMatchDate(frequency string, from time.Time) time.Time {
	switch frequency {
	case PoolFrequencyWeekly:
		return from.AddDate(0, 0, 7)
	case PoolFrequencyBiweekly:
		return from.AddDate(0, 0, 14)
	case PoolFrequencyMonthly:
		return from.AddDate(0, 1, 0)
	default:
		return from.AddDate(0, 0, 7) // Default to weekly
	}
}

// PoolMember represents a member's participation in a matching pool
type PoolMember struct {
	ID              string    `json:"id"`
	PoolID          string    `json:"pool_id"`
	MemberID        string    `json:"member_id"`
	UserID          string    `json:"user_id"` // For easier querying
	Active          bool      `json:"active"`
	ExcludedMembers []string  `json:"excluded_members,omitempty"` // Member IDs to never match with
	JoinedOn        time.Time `json:"joined_on"`
	// Populated fields
	MemberName *string `json:"member_name,omitempty"`
}

// MatchResult represents a generated match from the pool
type MatchResult struct {
	ID             string     `json:"id"`
	PoolID         string     `json:"pool_id"`
	Members        []string   `json:"members"`                   // Member IDs
	MemberUserIDs  []string   `json:"member_user_ids"`           // User IDs for notifications
	Status         string     `json:"status"`                    // pending, scheduled, completed, skipped
	MatchRound     string     `json:"match_round"`               // e.g., "2026-W02"
	ScheduledEvent *string    `json:"scheduled_event,omitempty"` // Event ID if created
	ScheduledTime  *time.Time `json:"scheduled_time,omitempty"`
	CreatedOn      time.Time  `json:"created_on"`
	UpdatedOn      time.Time  `json:"updated_on"`
	// Populated fields
	MemberNames []string `json:"member_names,omitempty"`
	PoolName    *string  `json:"pool_name,omitempty"`
}

// MatchStatus constants
const (
	MatchStatusPending   = "pending"   // Awaiting members to schedule
	MatchStatusScheduled = "scheduled" // Meeting scheduled
	MatchStatusCompleted = "completed" // Meeting happened
	MatchStatusSkipped   = "skipped"   // Members opted out
)

// PoolWithMembers includes pool details with member list
type PoolWithMembers struct {
	Pool    MatchingPool `json:"pool"`
	Members []PoolMember `json:"members"`
}

// PoolMatchHistory shows recent matches for a user
type PoolMatchHistory struct {
	UserID        string        `json:"user_id"`
	RecentMatches []MatchResult `json:"recent_matches"`
	// Maps member_id -> count of matches in last 30 days (for variety scoring)
	MatchCounts map[string]int `json:"match_counts,omitempty"`
}

// PoolStats provides statistics for a pool
type PoolStats struct {
	PoolID           string  `json:"pool_id"`
	TotalMembers     int     `json:"total_members"`
	ActiveMembers    int     `json:"active_members"`
	TotalRounds      int     `json:"total_rounds"`
	CompletedMatches int     `json:"completed_matches"`
	SkippedMatches   int     `json:"skipped_matches"`
	CompletionRate   float64 `json:"completion_rate"` // Percentage
}

// Constraints
const (
	MaxPoolsPerGuild       = 10
	MaxMembersPerPool      = 100
	MaxExclusionsPerMember = 20
	MinMatchSize           = 2
	MaxMatchSize           = 6
	MaxPoolNameLength      = 100
	MaxPoolDescLength      = 500
	MaxActivitySuggLength  = 200
)

// CreatePoolRequest represents a request to create a matching pool
type CreatePoolRequest struct {
	Name               string  `json:"name"`
	Description        *string `json:"description,omitempty"`
	Frequency          string  `json:"frequency"`
	MatchSize          int     `json:"match_size,omitempty"` // Default: 2
	ActivitySuggestion *string `json:"activity_suggestion,omitempty"`
}

// UpdatePoolRequest represents a request to update a pool
type UpdatePoolRequest struct {
	Name               *string `json:"name,omitempty"`
	Description        *string `json:"description,omitempty"`
	Frequency          *string `json:"frequency,omitempty"`
	MatchSize          *int    `json:"match_size,omitempty"`
	ActivitySuggestion *string `json:"activity_suggestion,omitempty"`
	Active             *bool   `json:"active,omitempty"`
}

// JoinPoolRequest represents a request to join a pool
type JoinPoolRequest struct {
	ExcludedMembers []string `json:"excluded_members,omitempty"`
}

// UpdateMembershipRequest represents updating pool membership settings
type UpdateMembershipRequest struct {
	Active          *bool    `json:"active,omitempty"`
	ExcludedMembers []string `json:"excluded_members,omitempty"`
}

// UpdateMatchRequest represents updating a match result
type UpdateMatchRequest struct {
	Status        *string    `json:"status,omitempty"`
	ScheduledTime *time.Time `json:"scheduled_time,omitempty"`
}

// MatchRoundInfo provides info about a matching round
type MatchRoundInfo struct {
	PoolID     string        `json:"pool_id"`
	PoolName   string        `json:"pool_name"`
	Round      string        `json:"round"` // e.g., "2026-W02"
	RanOn      time.Time     `json:"ran_on"`
	MatchCount int           `json:"match_count"`
	Matches    []MatchResult `json:"matches"`
}

// PendingMatch is a match waiting for the user to act on
type PendingMatch struct {
	Match        MatchResult `json:"match"`
	PoolName     string      `json:"pool_name"`
	GuildID      string      `json:"guild_id"`
	GuildName    string      `json:"guild_name"`
	PartnerIDs   []string    `json:"partner_ids"`   // Other member IDs in the match
	PartnerNames []string    `json:"partner_names"` // Other member names
	Suggestion   *string     `json:"suggestion,omitempty"`
	DueBy        *time.Time  `json:"due_by,omitempty"` // When next round happens
}

// MatchingConfig holds configuration for the matching algorithm
type MatchingConfig struct {
	// VarietyWeight: how much to penalize recent matches (0-1)
	// Higher = more variety, may result in suboptimal compatibility matches
	VarietyWeight float64 `json:"variety_weight"`
	// CompatibilityWeight: how much to weight compatibility scores (0-1)
	CompatibilityWeight float64 `json:"compatibility_weight"`
	// RecencyDays: how many days to consider for "recent" matches
	RecencyDays int `json:"recency_days"`
}

// DefaultMatchingConfig provides sensible defaults
var DefaultMatchingConfig = MatchingConfig{
	VarietyWeight:       0.6, // Prioritize variety over compatibility
	CompatibilityWeight: 0.4,
	RecencyDays:         30,
}

// GetMatchRound returns the match round string for a given time
// Format: "YYYY-Www" where ww is the ISO week number
func GetMatchRound(t time.Time) string {
	_, week := t.ISOWeek()
	return t.Format("2006") + "-W" + padWeek(week)
}

func padWeek(week int) string {
	if week < 10 {
		return "0" + string(rune('0'+week))
	}
	return string(rune('0'+week/10)) + string(rune('0'+week%10))
}
