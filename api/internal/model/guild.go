package model

import "time"

// Member represents a member linked to a user
type Member struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	UserID    string    `json:"user_id"`
	CreatedOn time.Time `json:"created_on"`
	UpdatedOn time.Time `json:"updated_on"`
}

// Guild represents a community with shared purpose (formerly Circle)
type Guild struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Icon        string    `json:"icon,omitempty"`
	Color       string    `json:"color,omitempty"`
	Visibility  string    `json:"visibility"` // private, public
	CreatedOn   time.Time `json:"created_on"`
	UpdatedOn   time.Time `json:"updated_on"`
}

// GuildVisibility constants
const (
	GuildVisibilityPrivate = "private"
	GuildVisibilityPublic  = "public"
)

// GuildRole represents a member's role within a guild
type GuildRole string

const (
	GuildRoleMember    GuildRole = "member"    // Default - can participate
	GuildRoleModerator GuildRole = "moderator" // Can manage content
	GuildRoleAdmin     GuildRole = "admin"     // Full guild management
)

// IsAdmin returns true if the role has admin privileges
func (r GuildRole) IsAdmin() bool {
	return r == GuildRoleAdmin
}

// IsModerator returns true if the role has moderator privileges (includes admin)
func (r GuildRole) IsModerator() bool {
	return r == GuildRoleModerator || r == GuildRoleAdmin
}

// IsValid returns true if the role is a valid guild role
func (r GuildRole) IsValid() bool {
	switch r {
	case GuildRoleMember, GuildRoleModerator, GuildRoleAdmin:
		return true
	default:
		return false
	}
}

// GuildMembership represents a member's relationship to a guild
type GuildMembership struct {
	MemberID        string    `json:"member_id"`
	GuildID         string    `json:"guild_id"`
	Role            GuildRole `json:"role"`
	PendingApproval bool      `json:"pending_approval"`
}

// GuildData is a complete guild with all related data
type GuildData struct {
	Guild   Guild    `json:"guild"`
	Members []Member `json:"members"`
}

// Business constraints
const (
	MaxMembersPerGuild = 20
	MaxGuildsPerUser   = 10

	MaxGuildNameLength = 100
	MaxGuildDescLength = 500
)

// CreateGuildRequest represents a request to create a guild
type CreateGuildRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Color       string `json:"color,omitempty"`
	Visibility  string `json:"visibility,omitempty"` // defaults to "private"
}

// UpdateGuildRequest represents a request to update a guild
type UpdateGuildRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	Color       *string `json:"color,omitempty"`
	Visibility  *string `json:"visibility,omitempty"`
}

// CreatePersonRequest represents a request to create a person
type CreatePersonRequest struct {
	Name     string     `json:"name"`
	Nickname string     `json:"nickname,omitempty"`
	Birthday *time.Time `json:"birthday,omitempty"`
	Notes    string     `json:"notes,omitempty"`
}

// UpdatePersonRequest represents a request to update a person
type UpdatePersonRequest struct {
	Name     *string    `json:"name,omitempty"`
	Nickname *string    `json:"nickname,omitempty"`
	Birthday *time.Time `json:"birthday,omitempty"`
	Notes    *string    `json:"notes,omitempty"`
}

// CreateActivityRequest represents a request to create an activity
type CreateActivityRequest struct {
	Name     string  `json:"name"`
	Icon     string  `json:"icon"`
	Warn     float64 `json:"warn"`
	Critical float64 `json:"critical"`
}

// UpdateActivityRequest represents a request to update an activity
type UpdateActivityRequest struct {
	Name     *string  `json:"name,omitempty"`
	Icon     *string  `json:"icon,omitempty"`
	Warn     *float64 `json:"warn,omitempty"`
	Critical *float64 `json:"critical,omitempty"`
}

// CreateTimerRequest represents a request to create a timer
type CreateTimerRequest struct {
	ActivityID string `json:"activity_id"`
	Enabled    *bool  `json:"enabled,omitempty"`
	Push       *bool  `json:"push,omitempty"`
}

// UpdateTimerRequest represents a request to update a timer
type UpdateTimerRequest struct {
	Enabled *bool `json:"enabled,omitempty"`
	Push    *bool `json:"push,omitempty"`
}

// UpdateMemberRoleRequest represents a request to change a member's role
type UpdateMemberRoleRequest struct {
	Role GuildRole `json:"role"`
}

// Backward compatibility type aliases (deprecated, will be removed)
type Circle = Guild
type CircleMembership = GuildMembership
type CircleData = GuildData
