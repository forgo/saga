package model

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorCode represents API error codes
type ErrorCode int

const (
	// Authentication errors (1xxx)
	ErrCodeUnauthorized ErrorCode = 1001
	ErrCodeTokenExpired ErrorCode = 1002
	ErrCodeTokenInvalid ErrorCode = 1003
	ErrCodeLoginFailed  ErrorCode = 1004

	// Authorization errors (2xxx)
	ErrCodeForbidden ErrorCode = 2001
	ErrCodeNotMember ErrorCode = 2002

	// Resource errors (3xxx)
	ErrCodeNotFound      ErrorCode = 3001
	ErrCodeAlreadyExists ErrorCode = 3002
	ErrCodeConflict      ErrorCode = 3003

	// Validation errors (4xxx)
	ErrCodeValidation    ErrorCode = 4001
	ErrCodeInvalidInput  ErrorCode = 4002
	ErrCodeLimitExceeded ErrorCode = 4003

	// Internal errors (5xxx)
	ErrCodeInternal    ErrorCode = 5001
	ErrCodeDatabase    ErrorCode = 5002
	ErrCodeExternalAPI ErrorCode = 5003
)

// ProblemDetails represents RFC 9457 Problem Details for HTTP APIs
type ProblemDetails struct {
	Type     string       `json:"type"`
	Title    string       `json:"title"`
	Status   int          `json:"status"`
	Detail   string       `json:"detail,omitempty"`
	Instance string       `json:"instance,omitempty"`
	Errors   []FieldError `json:"errors,omitempty"`
	// Extension fields
	Code    ErrorCode `json:"code,omitempty"`
	Limit   *int      `json:"limit,omitempty"`
	Current *int      `json:"current,omitempty"`
}

// FieldError represents a validation error on a specific field
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (p *ProblemDetails) Error() string {
	return fmt.Sprintf("[%d] %s: %s", p.Status, p.Title, p.Detail)
}

// WriteJSON writes the problem details as JSON response
func (p *ProblemDetails) WriteJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(p)
}

// Common error constructors

func NewUnauthorizedError(detail string) *ProblemDetails {
	return &ProblemDetails{
		Type:   "https://saga-api.forgo.software/errors/unauthorized",
		Title:  "Unauthorized",
		Status: http.StatusUnauthorized,
		Detail: detail,
		Code:   ErrCodeUnauthorized,
	}
}

func NewForbiddenError(detail string) *ProblemDetails {
	return &ProblemDetails{
		Type:   "https://saga-api.forgo.software/errors/forbidden",
		Title:  "Forbidden",
		Status: http.StatusForbidden,
		Detail: detail,
		Code:   ErrCodeForbidden,
	}
}

func NewNotFoundError(resource string) *ProblemDetails {
	return &ProblemDetails{
		Type:   "https://saga-api.forgo.software/errors/not-found",
		Title:  "Not Found",
		Status: http.StatusNotFound,
		Detail: fmt.Sprintf("%s not found", resource),
		Code:   ErrCodeNotFound,
	}
}

func NewValidationError(errors []FieldError) *ProblemDetails {
	// Build detailed message from field errors
	detail := "One or more fields failed validation"
	if len(errors) > 0 {
		detail = fmt.Sprintf("%s: %s", errors[0].Field, errors[0].Message)
		if len(errors) > 1 {
			detail = fmt.Sprintf("%s (and %d more errors)", detail, len(errors)-1)
		}
	}
	return &ProblemDetails{
		Type:   "https://saga-api.forgo.software/errors/validation",
		Title:  "Validation Error",
		Status: http.StatusUnprocessableEntity,
		Detail: detail,
		Code:   ErrCodeValidation,
		Errors: errors,
	}
}

func NewLimitExceededError(resource string, limit, current int) *ProblemDetails {
	return &ProblemDetails{
		Type:    "https://saga-api.forgo.software/errors/limit-exceeded",
		Title:   "Limit Exceeded",
		Status:  http.StatusUnprocessableEntity,
		Detail:  fmt.Sprintf("Maximum of %d %s reached", limit, resource),
		Code:    ErrCodeLimitExceeded,
		Limit:   &limit,
		Current: &current,
	}
}

func NewConflictError(detail string) *ProblemDetails {
	return &ProblemDetails{
		Type:   "https://saga-api.forgo.software/errors/conflict",
		Title:  "Conflict",
		Status: http.StatusConflict,
		Detail: detail,
		Code:   ErrCodeConflict,
	}
}

func NewInternalError(detail string) *ProblemDetails {
	if detail == "" {
		detail = "An unexpected error occurred"
	}
	return &ProblemDetails{
		Type:   "https://saga-api.forgo.software/errors/internal",
		Title:  "Internal Server Error",
		Status: http.StatusInternalServerError,
		Detail: detail,
		Code:   ErrCodeInternal,
	}
}

func NewBadRequestError(detail string) *ProblemDetails {
	return &ProblemDetails{
		Type:   "https://saga-api.forgo.software/errors/bad-request",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: detail,
		Code:   ErrCodeInvalidInput,
	}
}

func NewMethodNotAllowedError(allowed string) *ProblemDetails {
	return &ProblemDetails{
		Type:   "https://saga-api.forgo.software/errors/method-not-allowed",
		Title:  "Method Not Allowed",
		Status: http.StatusMethodNotAllowed,
		Detail: fmt.Sprintf("Only %s method is allowed", allowed),
	}
}

func NewRateLimitError(retryAfter int) *ProblemDetails {
	return &ProblemDetails{
		Type:   "https://saga-api.forgo.software/errors/rate-limited",
		Title:  "Too Many Requests",
		Status: http.StatusTooManyRequests,
		Detail: fmt.Sprintf("Rate limit exceeded. Retry after %d seconds", retryAfter),
	}
}
