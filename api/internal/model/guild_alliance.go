package model

import "time"

// AllianceStatus represents the state of a guild alliance
type AllianceStatus string

const (
	AllianceStatusPending AllianceStatus = "pending" // Awaiting approval from other guild
	AllianceStatusActive  AllianceStatus = "active"  // Both guilds have approved
	AllianceStatusRevoked AllianceStatus = "revoked" // Alliance has been revoked
)

// GuildAlliance represents a bidirectional partnership between two guilds
type GuildAlliance struct {
	ID            string         `json:"id"`
	GuildAID      string         `json:"guild_a_id"`
	GuildBID      string         `json:"guild_b_id"`
	Status        AllianceStatus `json:"status"`
	InitiatedByID string         `json:"initiated_by_id"`          // User ID who initiated
	ApprovedByID  *string        `json:"approved_by_id,omitempty"` // User ID who approved
	CreatedOn     time.Time      `json:"created_on"`
	ApprovedOn    *time.Time     `json:"approved_on,omitempty"`
	RevokedOn     *time.Time     `json:"revoked_on,omitempty"`
}

// GuildAllianceWithGuilds includes full guild information
type GuildAllianceWithGuilds struct {
	Alliance GuildAlliance `json:"alliance"`
	GuildA   GuildSummary  `json:"guild_a"`
	GuildB   GuildSummary  `json:"guild_b"`
}

// GuildSummary provides minimal guild info for display
type GuildSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Color       string `json:"color,omitempty"`
	MemberCount int    `json:"member_count,omitempty"`
}

// CreateAllianceRequest represents a request to create an alliance.
type CreateAllianceRequest struct {
	TargetGuildID string `json:"target_guild_id"`
}

// ApproveAllianceRequest represents approval of a pending alliance
type ApproveAllianceRequest struct {
	Approved bool `json:"approved"`
}

// RevokeAllianceRequest represents revoking an active alliance
type RevokeAllianceRequest struct {
	Reason *string `json:"reason,omitempty"`
}

// AllianceSearchFilters for filtering alliances
type AllianceSearchFilters struct {
	GuildID    *string         `json:"guild_id,omitempty"`
	Status     *AllianceStatus `json:"status,omitempty"`
	ActiveOnly bool            `json:"active_only,omitempty"`
}

// AlliedGuild represents a guild that is allied with another
type AlliedGuild struct {
	Guild      GuildSummary   `json:"guild"`
	AllianceID string         `json:"alliance_id"`
	Status     AllianceStatus `json:"status"`
	SinceDate  time.Time      `json:"since_date"`
}

// Constraints
const (
	MaxAlliancesPerGuild = 50
)
