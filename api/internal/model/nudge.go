package model

import "time"

// NudgeType represents the kind of nudge being sent
type NudgeType string

const (
	// Hangout-related nudges
	NudgeTypePendingMatch    NudgeType = "pending_match"    // Match created, no action taken
	NudgeTypeStaleHangout    NudgeType = "stale_hangout"    // Hangout scheduled but no follow-up
	NudgeTypeUpcomingHangout NudgeType = "upcoming_hangout" // Reminder for scheduled hangout
	NudgeTypeHangoutFollowUp NudgeType = "hangout_followup" // Post-hangout feedback reminder

	// Availability-related nudges
	NudgeTypePendingRequest     NudgeType = "pending_request"     // Someone wants to hang out
	NudgeTypeUnrespondedRequest NudgeType = "unresponded_request" // Requester hasn't heard back

	// Pool-related nudges
	NudgeTypePoolMatchCreated NudgeType = "pool_match_created" // New pool match available
	NudgeTypePoolMatchStale   NudgeType = "pool_match_stale"   // Pool match not acted on
)

// NudgeChannel represents how the nudge is delivered
type NudgeChannel string

const (
	NudgeChannelSSE  NudgeChannel = "sse"  // Server-sent events (in-app)
	NudgeChannelPush NudgeChannel = "push" // Push notification
)

// Nudge represents a nudge to be sent to a user
type Nudge struct {
	ID        string       `json:"id"`
	UserID    string       `json:"user_id"`
	Type      NudgeType    `json:"type"`
	Channel   NudgeChannel `json:"channel"`
	Title     string       `json:"title"`
	Message   string       `json:"message"`
	Data      NudgeData    `json:"data,omitempty"` // Additional context
	SentAt    time.Time    `json:"sent_at"`
	ReadAt    *time.Time   `json:"read_at,omitempty"`
	ActedAt   *time.Time   `json:"acted_at,omitempty"` // When user took action
	ExpiresAt *time.Time   `json:"expires_at,omitempty"`
}

// NudgeData contains type-specific context
type NudgeData struct {
	// For hangout/match nudges
	HangoutID      *string `json:"hangout_id,omitempty"`
	MatchID        *string `json:"match_id,omitempty"`
	AvailabilityID *string `json:"availability_id,omitempty"`
	PoolID         *string `json:"pool_id,omitempty"`

	// For partner-related nudges
	PartnerUserID  *string  `json:"partner_user_id,omitempty"`
	PartnerUserIDs []string `json:"partner_user_ids,omitempty"`
	PartnerNames   []string `json:"partner_names,omitempty"`

	// For activity nudges
	ActivityDesc  *string    `json:"activity_desc,omitempty"`
	ScheduledTime *time.Time `json:"scheduled_time,omitempty"`

	// Deep link info
	ActionURL *string `json:"action_url,omitempty"` // e.g., "/hangout/123"
}

// NudgeConfig defines when and how nudges are triggered
type NudgeConfig struct {
	Type           NudgeType     `json:"type"`
	Enabled        bool          `json:"enabled"`
	DelayAfter     time.Duration `json:"delay_after"`     // How long to wait before nudging
	RepeatInterval time.Duration `json:"repeat_interval"` // How often to re-nudge
	MaxRepeat      int           `json:"max_repeat"`      // Max times to nudge (0 = unlimited)
	CooldownPeriod time.Duration `json:"cooldown_period"` // Min time between nudges of any type
	Channel        NudgeChannel  `json:"channel"`
}

// DefaultNudgeConfigs provides sensible defaults for each nudge type
var DefaultNudgeConfigs = map[NudgeType]NudgeConfig{
	NudgeTypePendingMatch: {
		Type:           NudgeTypePendingMatch,
		Enabled:        true,
		DelayAfter:     24 * time.Hour, // Wait 24h after match created
		RepeatInterval: 48 * time.Hour, // Remind every 48h
		MaxRepeat:      2,              // Max 2 reminders
		CooldownPeriod: 6 * time.Hour,
		Channel:        NudgeChannelSSE,
	},
	NudgeTypeStaleHangout: {
		Type:           NudgeTypeStaleHangout,
		Enabled:        true,
		DelayAfter:     72 * time.Hour,  // Wait 3 days
		RepeatInterval: 168 * time.Hour, // Weekly reminder
		MaxRepeat:      1,
		CooldownPeriod: 24 * time.Hour,
		Channel:        NudgeChannelSSE,
	},
	NudgeTypeUpcomingHangout: {
		Type:           NudgeTypeUpcomingHangout,
		Enabled:        true,
		DelayAfter:     0, // Triggered before event
		RepeatInterval: 0, // No repeat
		MaxRepeat:      0,
		CooldownPeriod: 0,
		Channel:        NudgeChannelPush,
	},
	NudgeTypeHangoutFollowUp: {
		Type:           NudgeTypeHangoutFollowUp,
		Enabled:        true,
		DelayAfter:     2 * time.Hour, // 2h after scheduled time
		RepeatInterval: 24 * time.Hour,
		MaxRepeat:      1,
		CooldownPeriod: 6 * time.Hour,
		Channel:        NudgeChannelSSE,
	},
	NudgeTypePendingRequest: {
		Type:           NudgeTypePendingRequest,
		Enabled:        true,
		DelayAfter:     0, // Immediate
		RepeatInterval: 4 * time.Hour,
		MaxRepeat:      2,
		CooldownPeriod: 2 * time.Hour,
		Channel:        NudgeChannelSSE,
	},
	NudgeTypeUnrespondedRequest: {
		Type:           NudgeTypeUnrespondedRequest,
		Enabled:        true,
		DelayAfter:     12 * time.Hour, // Wait 12h for response
		RepeatInterval: 24 * time.Hour,
		MaxRepeat:      1,
		CooldownPeriod: 12 * time.Hour,
		Channel:        NudgeChannelSSE,
	},
	NudgeTypePoolMatchCreated: {
		Type:           NudgeTypePoolMatchCreated,
		Enabled:        true,
		DelayAfter:     0, // Immediate
		RepeatInterval: 0,
		MaxRepeat:      0,
		CooldownPeriod: 0,
		Channel:        NudgeChannelPush,
	},
	NudgeTypePoolMatchStale: {
		Type:           NudgeTypePoolMatchStale,
		Enabled:        true,
		DelayAfter:     48 * time.Hour, // 2 days after match
		RepeatInterval: 72 * time.Hour, // Every 3 days
		MaxRepeat:      2,
		CooldownPeriod: 24 * time.Hour,
		Channel:        NudgeChannelSSE,
	},
}

// NudgeHistory tracks sent nudges to prevent over-nudging
type NudgeHistory struct {
	UserID    string    `json:"user_id"`
	Type      NudgeType `json:"type"`
	TargetID  string    `json:"target_id"` // Hangout/Match/Pool ID
	SentCount int       `json:"sent_count"`
	LastSent  time.Time `json:"last_sent"`
}

// NudgePreference allows users to control nudge settings
type NudgePreference struct {
	UserID  string        `json:"user_id"`
	Type    NudgeType     `json:"type"`
	Enabled bool          `json:"enabled"`
	Channel *NudgeChannel `json:"channel,omitempty"` // Override default channel
}

// NudgeSummary provides a summary of pending nudges for a user
type NudgeSummary struct {
	UserID           string `json:"user_id"`
	PendingMatches   int    `json:"pending_matches"`
	StaleHangouts    int    `json:"stale_hangouts"`
	UpcomingHangouts int    `json:"upcoming_hangouts"`
	PendingRequests  int    `json:"pending_requests"`
	TotalActionable  int    `json:"total_actionable"`
}

// NudgeTemplates for generating nudge messages
var NudgeTemplates = map[NudgeType]struct {
	Title   string
	Message string
}{
	NudgeTypePendingMatch: {
		Title:   "You have a pending match!",
		Message: "You were matched with %s. Ready to connect?",
	},
	NudgeTypeStaleHangout: {
		Title:   "How did it go?",
		Message: "Your hangout with %s was a while ago. Mark it complete or let us know if you need to reschedule.",
	},
	NudgeTypeUpcomingHangout: {
		Title:   "Hangout coming up!",
		Message: "Your hangout with %s is in %s. Don't forget!",
	},
	NudgeTypeHangoutFollowUp: {
		Title:   "How was your hangout?",
		Message: "We'd love to know how your time with %s went.",
	},
	NudgeTypePendingRequest: {
		Title:   "Someone wants to hang out!",
		Message: "%s is interested in joining your hangout.",
	},
	NudgeTypeUnrespondedRequest: {
		Title:   "Still waiting...",
		Message: "You haven't heard back about your hangout request. Maybe reach out directly?",
	},
	NudgeTypePoolMatchCreated: {
		Title:   "New match from %s!",
		Message: "You've been matched with %s. Time to connect!",
	},
	NudgeTypePoolMatchStale: {
		Title:   "Don't forget your match!",
		Message: "You were matched with %s but haven't connected yet. The next round is coming up!",
	},
}

// GetNudgeMessage generates a nudge message from template
func GetNudgeMessage(nudgeType NudgeType, args ...interface{}) (title, message string) {
	template, ok := NudgeTemplates[nudgeType]
	if !ok {
		return "Reminder", "You have a pending action"
	}

	title = template.Title
	message = template.Message

	// Simple substitution - in production, use proper formatting
	// This is just a placeholder implementation
	return title, message
}
