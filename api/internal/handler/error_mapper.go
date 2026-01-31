package handler

import (
	"errors"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// MapServiceError converts a service error to a ProblemDetails response.
// This centralizes error handling logic for all handlers, ensuring consistent
// HTTP status codes and error messages across the API.
func MapServiceError(err error) *model.ProblemDetails {
	if err == nil {
		return nil
	}

	// ===== Authentication Errors → 401 =====
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		return model.NewUnauthorizedError(err.Error())
	case errors.Is(err, service.ErrInvalidRefreshToken),
		errors.Is(err, service.ErrRefreshTokenExpired),
		errors.Is(err, service.ErrRefreshTokenRevoked):
		return model.NewUnauthorizedError(err.Error())
	case errors.Is(err, service.ErrInvalidChallenge),
		errors.Is(err, service.ErrInvalidCredential),
		errors.Is(err, service.ErrCredentialNotAllowed):
		return model.NewUnauthorizedError(err.Error())

	// ===== Authorization Errors → 403 =====
	case errors.Is(err, service.ErrNotGuildAdmin),
		errors.Is(err, service.ErrNotGuildMember),
		errors.Is(err, service.ErrNotEventHost),
		errors.Is(err, service.ErrNotPoolMember),
		errors.Is(err, service.ErrNotMatchMember),
		errors.Is(err, service.ErrCannotAssignOthers):
		return model.NewForbiddenError(err.Error())

	// ===== Not Found Errors → 404 =====
	case errors.Is(err, service.ErrUserNotFound):
		return model.NewNotFoundError("user")
	case errors.Is(err, service.ErrGuildNotFound):
		return model.NewNotFoundError("guild")
	case errors.Is(err, service.ErrEventNotFound):
		return model.NewNotFoundError("event")
	case errors.Is(err, service.ErrRSVPNotFound):
		return model.NewNotFoundError("RSVP")
	case errors.Is(err, service.ErrRoleNotFound):
		return model.NewNotFoundError("role")
	case errors.Is(err, service.ErrAssignmentNotFound):
		return model.NewNotFoundError("assignment")
	case errors.Is(err, service.ErrProfileNotFound):
		return model.NewNotFoundError("profile")
	case errors.Is(err, service.ErrAvailabilityNotFound):
		return model.NewNotFoundError("availability")
	case errors.Is(err, service.ErrHangoutNotFound):
		return model.NewNotFoundError("hangout")
	case errors.Is(err, service.ErrHangoutRequestNotFound):
		return model.NewNotFoundError("hangout request")
	case errors.Is(err, service.ErrTrustNotFound):
		return model.NewNotFoundError("trust relation")
	case errors.Is(err, service.ErrIRLNotFound):
		return model.NewNotFoundError("IRL verification")
	case errors.Is(err, service.ErrReviewNotFound):
		return model.NewNotFoundError("review")
	case errors.Is(err, service.ErrQuestionNotFound):
		return model.NewNotFoundError("question")
	case errors.Is(err, service.ErrAnswerNotFound):
		return model.NewNotFoundError("answer")
	case errors.Is(err, service.ErrInterestNotFound):
		return model.NewNotFoundError("interest")
	case errors.Is(err, service.ErrPoolNotFound):
		return model.NewNotFoundError("pool")
	case errors.Is(err, service.ErrMatchNotFound):
		return model.NewNotFoundError("match")
	case errors.Is(err, service.ErrReportNotFound):
		return model.NewNotFoundError("report")
	case errors.Is(err, service.ErrActionNotFound):
		return model.NewNotFoundError("moderation action")
	case errors.Is(err, service.ErrPasskeyNotFound):
		return model.NewNotFoundError("passkey")

	// ===== Conflict Errors → 409 =====
	case errors.Is(err, service.ErrEmailAlreadyExists),
		errors.Is(err, service.ErrGuildNameExists):
		return model.NewConflictError(err.Error())
	case errors.Is(err, service.ErrAlreadyGuildMember),
		errors.Is(err, service.ErrAlreadyRSVPd),
		errors.Is(err, service.ErrAlreadyAssignedToRole),
		errors.Is(err, service.ErrAlreadyPoolMember),
		errors.Is(err, service.ErrAlreadyTrusted),
		errors.Is(err, service.ErrAlreadyReviewed),
		errors.Is(err, service.ErrAlreadyBlocked),
		errors.Is(err, service.ErrAlreadyHost),
		errors.Is(err, service.ErrAlreadyRequested),
		errors.Is(err, service.ErrInterestAlreadyExists),
		errors.Is(err, service.ErrProfileExists):
		return model.NewConflictError(err.Error())

	// ===== Validation Errors → 422 =====
	// Self-action prevention
	case errors.Is(err, service.ErrCannotTrustSelf),
		errors.Is(err, service.ErrCannotReviewSelf),
		errors.Is(err, service.ErrCannotReportSelf),
		errors.Is(err, service.ErrCannotBlockSelf),
		errors.Is(err, service.ErrCannotRequestOwn):
		return model.NewValidationError([]model.FieldError{{Field: "target", Message: err.Error()}})

	// Format/input validation
	case errors.Is(err, service.ErrInvalidEmail),
		errors.Is(err, service.ErrPasswordRequired),
		errors.Is(err, service.ErrPasswordTooShort),
		errors.Is(err, service.ErrPasswordTooLong):
		return model.NewValidationError([]model.FieldError{{Field: "credentials", Message: err.Error()}})

	case errors.Is(err, service.ErrGuildNameRequired),
		errors.Is(err, service.ErrGuildNameTooLong),
		errors.Is(err, service.ErrGuildDescTooLong):
		return model.NewValidationError([]model.FieldError{{Field: "guild", Message: err.Error()}})

	case errors.Is(err, service.ErrInvalidHangoutType),
		errors.Is(err, service.ErrInvalidTimeRange),
		errors.Is(err, service.ErrInvalidStartTimeFormat),
		errors.Is(err, service.ErrInvalidEndTimeFormat),
		errors.Is(err, service.ErrNoteTooShort):
		return model.NewValidationError([]model.FieldError{{Field: "availability", Message: err.Error()}})

	case errors.Is(err, service.ErrInvalidOption),
		errors.Is(err, service.ErrInvalidImportance),
		errors.Is(err, service.ErrDealBreakerNotAllowed),
		errors.Is(err, service.ErrInvalidAlignmentWeight):
		return model.NewValidationError([]model.FieldError{{Field: "questionnaire", Message: err.Error()}})

	case errors.Is(err, service.ErrInvalidInterestLevel):
		return model.NewValidationError([]model.FieldError{{Field: "interest", Message: err.Error()}})

	case errors.Is(err, service.ErrInvalidVisibility),
		errors.Is(err, service.ErrBioTooLong),
		errors.Is(err, service.ErrTaglineTooLong),
		errors.Is(err, service.ErrTooManyLanguages):
		return model.NewValidationError([]model.FieldError{{Field: "profile", Message: err.Error()}})

	case errors.Is(err, service.ErrInvalidReviewContext),
		errors.Is(err, service.ErrTooManyTags),
		errors.Is(err, service.ErrPrivateNoteTooLong):
		return model.NewValidationError([]model.FieldError{{Field: "review", Message: err.Error()}})

	case errors.Is(err, service.ErrInvalidMatchSize),
		errors.Is(err, service.ErrInvalidFrequency),
		errors.Is(err, service.ErrPoolNotInGuild):
		return model.NewValidationError([]model.FieldError{{Field: "pool", Message: err.Error()}})

	case errors.Is(err, service.ErrInvalidCategory),
		errors.Is(err, service.ErrInvalidLevel),
		errors.Is(err, service.ErrInvalidStatus),
		errors.Is(err, service.ErrReasonRequired),
		errors.Is(err, service.ErrDescriptionTooLong):
		return model.NewValidationError([]model.FieldError{{Field: "moderation", Message: err.Error()}})

	case errors.Is(err, service.ErrInvalidContext):
		return model.NewValidationError([]model.FieldError{{Field: "context", Message: err.Error()}})

	// Limit/capacity errors → 422
	case errors.Is(err, service.ErrMaxGuildsReached),
		errors.Is(err, service.ErrMaxMembersReached),
		errors.Is(err, service.ErrMaxHostsReached),
		errors.Is(err, service.ErrMaxRolesReached),
		errors.Is(err, service.ErrMaxRolesPerUserReached),
		errors.Is(err, service.ErrPoolLimitReached),
		errors.Is(err, service.ErrMemberPoolLimitReached),
		errors.Is(err, service.ErrExclusionLimitReached),
		errors.Is(err, service.ErrPasskeyLimitReached),
		errors.Is(err, service.ErrEventFull),
		errors.Is(err, service.ErrRoleFull),
		errors.Is(err, service.ErrNotEnoughMembers):
		return model.NewValidationError([]model.FieldError{{Field: "limit", Message: err.Error()}})

	// State errors → 422
	case errors.Is(err, service.ErrCannotLeaveSoleMember),
		errors.Is(err, service.ErrCannotDeleteDefault),
		errors.Is(err, service.ErrNotBlocked),
		errors.Is(err, service.ErrTrustNotEstablished),
		errors.Is(err, service.ErrIRLRequired),
		errors.Is(err, service.ErrValuesCheckRequired),
		errors.Is(err, service.ErrRSVPNotAllowed):
		return model.NewValidationError([]model.FieldError{{Field: "state", Message: err.Error()}})

	// ===== Security Errors → 400 =====
	case errors.Is(err, service.ErrSignCountMismatch):
		return model.NewBadRequestError("Security verification failed: " + err.Error())
	case errors.Is(err, service.ErrInvalidAuthCode),
		errors.Is(err, service.ErrPKCEVerifyFailed),
		errors.Is(err, service.ErrInvalidIDToken),
		errors.Is(err, service.ErrEmailNotVerified):
		return model.NewBadRequestError(err.Error())
	case errors.Is(err, service.ErrAccountLinkingRequired),
		errors.Is(err, service.ErrAccountLinkPending):
		return model.NewBadRequestError(err.Error())

	// ===== Banned/Suspended Users → 403 =====
	case errors.Is(err, service.ErrUserBanned),
		errors.Is(err, service.ErrUserSuspended):
		return model.NewForbiddenError(err.Error())

	// ===== Push Notification Errors → 400 =====
	case errors.Is(err, service.ErrPushDisabled),
		errors.Is(err, service.ErrNoDeviceTokens),
		errors.Is(err, service.ErrInvalidDeviceToken):
		return model.NewBadRequestError(err.Error())

	// ===== Provider/External Errors → 502 =====
	case errors.Is(err, service.ErrProviderError):
		return &model.ProblemDetails{
			Type:   "https://saga-api.forgo.software/errors/external-service",
			Title:  "External Service Error",
			Status: 502,
			Detail: err.Error(),
		}

	// ===== Default → 500 =====
	default:
		return model.NewInternalError("")
	}
}

// MapServiceErrorWithContext converts a service error to a ProblemDetails response
// with additional context about the operation that failed.
func MapServiceErrorWithContext(err error, operation string) *model.ProblemDetails {
	pd := MapServiceError(err)
	if pd != nil && pd.Status == 500 {
		pd.Detail = operation + ": an unexpected error occurred"
	}
	return pd
}
