package service

import "errors"

// Centralized service layer errors.
// All errors returned by service methods are defined here for consistency
// and to make error handling in handlers predictable.

// ===== Authentication Errors =====
var (
	ErrInvalidCredentials     = errors.New("invalid email or password")
	ErrEmailAlreadyExists     = errors.New("email already registered")
	ErrUserNotFound           = errors.New("user not found")
	ErrPasswordRequired       = errors.New("password is required")
	ErrPasswordTooShort       = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong        = errors.New("password must be at most 128 characters")
	ErrInvalidEmail           = errors.New("invalid email format")
	ErrAccountLinkingRequired = errors.New("account linking required")
)

// ===== Token Errors =====
var (
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrRefreshTokenRevoked = errors.New("refresh token revoked")
)

// ===== OAuth Errors =====
var (
	ErrInvalidAuthCode    = errors.New("invalid authorization code")
	ErrPKCEVerifyFailed   = errors.New("PKCE verification failed")
	ErrProviderError      = errors.New("OAuth provider error")
	ErrInvalidIDToken     = errors.New("invalid ID token")
	ErrEmailNotVerified   = errors.New("email not verified by provider")
	ErrAccountLinkPending = errors.New("account linking required")
)

// ===== Passkey Errors =====
var (
	ErrPasskeyNotFound      = errors.New("passkey not found")
	ErrInvalidChallenge     = errors.New("invalid or expired challenge")
	ErrInvalidCredential    = errors.New("invalid credential")
	ErrCredentialNotAllowed = errors.New("credential not allowed for this user")
	ErrSignCountMismatch    = errors.New("sign count mismatch - potential cloned authenticator")
	ErrPasskeyLimitReached  = errors.New("maximum number of passkeys reached")
)

// ===== Guild Errors =====
var (
	ErrGuildNotFound               = errors.New("guild not found")
	ErrGuildNameRequired           = errors.New("guild name is required")
	ErrGuildNameTooLong            = errors.New("guild name exceeds maximum length")
	ErrGuildDescTooLong            = errors.New("guild description exceeds maximum length")
	ErrNotGuildMember              = errors.New("not a member of this guild")
	ErrNotGuildAdmin               = errors.New("not authorized to perform this action")
	ErrCannotLeaveSoleMember       = errors.New("cannot leave guild as the only member")
	ErrAlreadyGuildMember          = errors.New("already a member of this guild")
	ErrMaxGuildsReached            = errors.New("maximum number of guilds reached")
	ErrMaxMembersReached           = errors.New("guild has reached maximum member limit")
	ErrGuildNameExists             = errors.New("a guild with this name already exists")
	ErrMergeRequiresDualMembership = errors.New("must be a member of both guilds to merge")
)

// ===== Event Errors =====
var (
	ErrEventNotFound       = errors.New("event not found")
	ErrRSVPNotFound        = errors.New("RSVP not found")
	ErrNotEventHost        = errors.New("not an event host")
	ErrEventFull           = errors.New("event is full")
	ErrAlreadyRSVPd        = errors.New("already RSVP'd")
	ErrRSVPNotAllowed      = errors.New("RSVP not allowed for this event")
	ErrValuesCheckRequired = errors.New("values alignment check required")
	ErrMaxHostsReached     = errors.New("maximum hosts reached")
	ErrAlreadyHost         = errors.New("already a host")
)

// ===== Event Role Errors =====
var (
	ErrRoleNotFound           = errors.New("role not found")
	ErrAssignmentNotFound     = errors.New("assignment not found")
	ErrRoleFull               = errors.New("role is full")
	ErrAlreadyAssignedToRole  = errors.New("already assigned to this role")
	ErrCannotDeleteDefault    = errors.New("cannot delete default role")
	ErrMaxRolesReached        = errors.New("maximum roles reached")
	ErrCannotAssignOthers     = errors.New("cannot assign roles to others")
	ErrMaxRolesPerUserReached = errors.New("maximum roles per user reached")
)

// ===== Profile Errors =====
var (
	ErrProfileNotFound   = errors.New("profile not found")
	ErrProfileExists     = errors.New("profile already exists")
	ErrInvalidVisibility = errors.New("invalid visibility setting")
	ErrBioTooLong        = errors.New("bio exceeds maximum length")
	ErrTaglineTooLong    = errors.New("tagline exceeds maximum length")
	ErrTooManyLanguages  = errors.New("too many languages")
)

// ===== Availability Errors =====
var (
	ErrAvailabilityNotFound   = errors.New("availability not found")
	ErrHangoutRequestNotFound = errors.New("hangout request not found")
	ErrHangoutNotFound        = errors.New("hangout not found")
	ErrInvalidHangoutType     = errors.New("invalid hangout type")
	ErrInvalidTimeRange       = errors.New("end time must be after start time")
	ErrNoteTooShort           = errors.New("note must be at least 20 characters")
	ErrAlreadyRequested       = errors.New("already requested this hangout")
	ErrCannotRequestOwn       = errors.New("cannot request your own availability")
	ErrInvalidStartTimeFormat = errors.New("invalid start_time format")
	ErrInvalidEndTimeFormat   = errors.New("invalid end_time format")
)

// ===== Trust Errors =====
var (
	ErrTrustNotFound       = errors.New("trust relation not found")
	ErrIRLNotFound         = errors.New("IRL verification not found")
	ErrCannotTrustSelf     = errors.New("cannot trust yourself")
	ErrAlreadyTrusted      = errors.New("already trusted")
	ErrTrustNotEstablished = errors.New("trust not established")
	ErrIRLRequired         = errors.New("IRL verification required")
	ErrInvalidContext      = errors.New("invalid IRL context")
)

// ===== Review Errors =====
var (
	ErrReviewNotFound       = errors.New("review not found")
	ErrCannotReviewSelf     = errors.New("cannot review yourself")
	ErrAlreadyReviewed      = errors.New("already reviewed for this reference")
	ErrInvalidReviewContext = errors.New("invalid review context")
	ErrTooManyTags          = errors.New("too many tags")
	ErrPrivateNoteTooLong   = errors.New("private note too long")
)

// ===== Questionnaire Errors =====
var (
	ErrQuestionNotFound       = errors.New("question not found")
	ErrAnswerNotFound         = errors.New("answer not found")
	ErrInvalidOption          = errors.New("invalid option selected")
	ErrInvalidImportance      = errors.New("invalid importance level")
	ErrDealBreakerNotAllowed  = errors.New("this question cannot be a dealbreaker")
	ErrInvalidAlignmentWeight = errors.New("alignment weight must be between 0 and 1")
)

// ===== Interest Errors =====
var (
	ErrInterestNotFound      = errors.New("interest not found")
	ErrInterestAlreadyExists = errors.New("user already has this interest")
	ErrInvalidInterestLevel  = errors.New("invalid interest level")
)

// ===== Pool Errors =====
var (
	ErrPoolNotFound           = errors.New("pool not found")
	ErrPoolLimitReached       = errors.New("maximum pools per guild reached")
	ErrMemberPoolLimitReached = errors.New("maximum members per pool reached")
	ErrAlreadyPoolMember      = errors.New("already a member of this pool")
	ErrNotPoolMember          = errors.New("not a member of this pool")
	ErrPoolNotInGuild         = errors.New("pool does not belong to this guild")
	ErrInvalidMatchSize       = errors.New("match size must be between 2 and 6")
	ErrInvalidFrequency       = errors.New("invalid frequency")
	ErrMatchNotFound          = errors.New("match not found")
	ErrNotMatchMember         = errors.New("not a member of this match")
	ErrExclusionLimitReached  = errors.New("maximum exclusions reached")
	ErrNotEnoughMembers       = errors.New("not enough active members to create matches")
)

// ===== Moderation Errors =====
var (
	ErrReportNotFound     = errors.New("report not found")
	ErrActionNotFound     = errors.New("moderation action not found")
	ErrCannotReportSelf   = errors.New("cannot report yourself")
	ErrCannotBlockSelf    = errors.New("cannot block yourself")
	ErrAlreadyBlocked     = errors.New("user already blocked")
	ErrNotBlocked         = errors.New("user not blocked")
	ErrUserBanned         = errors.New("user is banned")
	ErrUserSuspended      = errors.New("user is suspended")
	ErrInvalidCategory    = errors.New("invalid report category")
	ErrInvalidLevel       = errors.New("invalid moderation level")
	ErrInvalidStatus      = errors.New("invalid report status")
	ErrReasonRequired     = errors.New("reason is required")
	ErrDescriptionTooLong = errors.New("description too long")
)

// ===== Push Notification Errors =====
var (
	ErrPushDisabled       = errors.New("push notifications are disabled")
	ErrNoDeviceTokens     = errors.New("no device tokens found for user")
	ErrInvalidDeviceToken = errors.New("invalid device token")
)
