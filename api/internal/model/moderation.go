package model

import "time"

// ReportCategory represents the type of report
type ReportCategory string

const (
	ReportCategorySpam           ReportCategory = "spam"
	ReportCategoryHarassment     ReportCategory = "harassment"
	ReportCategoryHateSpeech     ReportCategory = "hate_speech"
	ReportCategoryInappropriate  ReportCategory = "inappropriate_content"
	ReportCategoryUncomfortable  ReportCategory = "made_uncomfortable" // Private, no public action
	ReportCategoryOther          ReportCategory = "other"
)

// ReportStatus represents the state of a report
type ReportStatus string

const (
	ReportStatusPending  ReportStatus = "pending"
	ReportStatusReviewed ReportStatus = "reviewed"
	ReportStatusResolved ReportStatus = "resolved"
	ReportStatusDismissed ReportStatus = "dismissed"
)

// Report represents a user report against another user or content
type Report struct {
	ID              string         `json:"id"`
	ReporterUserID  string         `json:"reporter_user_id"`
	ReportedUserID  string         `json:"reported_user_id"`
	CircleID        *string        `json:"circle_id,omitempty"` // Optional context
	Category        ReportCategory `json:"category"`
	Description     *string        `json:"description,omitempty"`
	ContentType     *string        `json:"content_type,omitempty"` // event, message, profile, etc.
	ContentID       *string        `json:"content_id,omitempty"`   // ID of reported content
	Status          ReportStatus   `json:"status"`
	ReviewedByID    *string        `json:"reviewed_by_id,omitempty"` // Admin who reviewed
	ReviewNotes     *string        `json:"review_notes,omitempty"`
	ActionTaken     *string        `json:"action_taken,omitempty"` // Description of action
	CreatedOn       time.Time      `json:"created_on"`
	ReviewedOn      *time.Time     `json:"reviewed_on,omitempty"`
	ResolvedOn      *time.Time     `json:"resolved_on,omitempty"`
}

// ModerationLevel represents the graduated response level
type ModerationLevel string

const (
	ModerationLevelNudge      ModerationLevel = "nudge"      // Level 0: Private message
	ModerationLevelWarning    ModerationLevel = "warning"    // Level 1: Visible warning
	ModerationLevelSuspension ModerationLevel = "suspension" // Level 2: Temp suspension
	ModerationLevelBan        ModerationLevel = "ban"        // Level 3: Permanent ban
)

// ModerationAction represents an action taken against a user
type ModerationAction struct {
	ID           string          `json:"id"`
	UserID       string          `json:"user_id"`
	Level        ModerationLevel `json:"level"`
	Reason       string          `json:"reason"`
	ReportID     *string         `json:"report_id,omitempty"` // Linked report if any
	AdminUserID  *string         `json:"admin_user_id,omitempty"`
	Duration     *int            `json:"duration_days,omitempty"` // For suspensions
	ExpiresOn    *time.Time      `json:"expires_on,omitempty"`
	IsActive     bool            `json:"is_active"`
	Restrictions []string        `json:"restrictions,omitempty"` // e.g., "create_public_events"
	CreatedOn    time.Time       `json:"created_on"`
	LiftedOn     *time.Time      `json:"lifted_on,omitempty"`
	LiftedByID   *string         `json:"lifted_by_id,omitempty"`
	LiftReason   *string         `json:"lift_reason,omitempty"`
}

// Block represents a user blocking another user
type Block struct {
	ID            string    `json:"id"`
	BlockerUserID string    `json:"blocker_user_id"`
	BlockedUserID string    `json:"blocked_user_id"`
	Reason        *string   `json:"reason,omitempty"` // Private, not shared
	CreatedOn     time.Time `json:"created_on"`
}

// UserModerationStatus represents a user's current moderation standing
type UserModerationStatus struct {
	UserID            string             `json:"user_id"`
	IsBanned          bool               `json:"is_banned"`
	IsSuspended       bool               `json:"is_suspended"`
	SuspensionEndsOn  *time.Time         `json:"suspension_ends_on,omitempty"`
	HasWarning        bool               `json:"has_warning"`
	WarningExpiresOn  *time.Time         `json:"warning_expires_on,omitempty"`
	Restrictions      []string           `json:"restrictions,omitempty"`
	ActiveActions     []ModerationAction `json:"active_actions,omitempty"`
	ReportCount       int                `json:"report_count"`       // Reports against this user
	RecentReportCount int                `json:"recent_report_count"` // Reports in last 30 days
}

// CircleModerationSettings represents moderation settings for a circle
type CircleModerationSettings struct {
	CircleID              string `json:"circle_id"`
	RequireApprovalToJoin bool   `json:"require_approval_to_join"`
	MinReputationToJoin   *int   `json:"min_reputation_to_join,omitempty"`
	AllowPublicEvents     bool   `json:"allow_public_events"`
	PublicEventMinRep     *int   `json:"public_event_min_reputation,omitempty"`
}

// Constraints
const (
	MaxReportDescriptionLength = 1000
	MaxBlockReasonLength       = 500
	MaxActionReasonLength      = 1000
	WarningDurationDays        = 7
	DefaultSuspensionDays      = 30
	RecentReportWindowDays     = 30
)

// CreateReportRequest represents a request to create a report
type CreateReportRequest struct {
	ReportedUserID string  `json:"reported_user_id"`
	CircleID       *string `json:"circle_id,omitempty"`
	Category       string  `json:"category"`
	Description    *string `json:"description,omitempty"`
	ContentType    *string `json:"content_type,omitempty"`
	ContentID      *string `json:"content_id,omitempty"`
}

// ReviewReportRequest represents a request to review a report
type ReviewReportRequest struct {
	Status      string  `json:"status"` // reviewed, resolved, dismissed
	Notes       *string `json:"notes,omitempty"`
	ActionTaken *string `json:"action_taken,omitempty"`
}

// CreateModerationActionRequest represents a request to take moderation action
type CreateModerationActionRequest struct {
	UserID       string   `json:"user_id"`
	Level        string   `json:"level"` // nudge, warning, suspension, ban
	Reason       string   `json:"reason"`
	ReportID     *string  `json:"report_id,omitempty"`
	DurationDays *int     `json:"duration_days,omitempty"` // For suspensions
	Restrictions []string `json:"restrictions,omitempty"`
}

// LiftActionRequest represents a request to lift a moderation action
type LiftActionRequest struct {
	Reason string `json:"reason"`
}

// CreateBlockRequest represents a request to block a user
type CreateBlockRequest struct {
	BlockedUserID string  `json:"blocked_user_id"`
	Reason        *string `json:"reason,omitempty"`
}

// ReportSummary provides a lightweight view of a report
type ReportSummary struct {
	ID             string         `json:"id"`
	ReportedUserID string         `json:"reported_user_id"`
	Category       ReportCategory `json:"category"`
	Status         ReportStatus   `json:"status"`
	CreatedOn      time.Time      `json:"created_on"`
}

// BlockedUserInfo provides info about a blocked user
type BlockedUserInfo struct {
	UserID    string    `json:"user_id"`
	Name      string    `json:"name,omitempty"`
	BlockedOn time.Time `json:"blocked_on"`
}

// ModerationStats provides moderation statistics
type ModerationStats struct {
	TotalReports      int `json:"total_reports"`
	PendingReports    int `json:"pending_reports"`
	ResolvedReports   int `json:"resolved_reports"`
	ActiveWarnings    int `json:"active_warnings"`
	ActiveSuspensions int `json:"active_suspensions"`
	TotalBans         int `json:"total_bans"`
}

// Valid report categories
func IsValidReportCategory(cat string) bool {
	switch ReportCategory(cat) {
	case ReportCategorySpam,
		ReportCategoryHarassment,
		ReportCategoryHateSpeech,
		ReportCategoryInappropriate,
		ReportCategoryUncomfortable,
		ReportCategoryOther:
		return true
	}
	return false
}

// Valid moderation levels
func IsValidModerationLevel(level string) bool {
	switch ModerationLevel(level) {
	case ModerationLevelNudge,
		ModerationLevelWarning,
		ModerationLevelSuspension,
		ModerationLevelBan:
		return true
	}
	return false
}

// Valid report statuses
func IsValidReportStatus(status string) bool {
	switch ReportStatus(status) {
	case ReportStatusPending,
		ReportStatusReviewed,
		ReportStatusResolved,
		ReportStatusDismissed:
		return true
	}
	return false
}
