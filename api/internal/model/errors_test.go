package model

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ============================================================================
// Error() Interface Tests
// ============================================================================

func TestProblemDetails_Error_ReturnsFormattedMessage(t *testing.T) {
	t.Parallel()

	pd := &ProblemDetails{
		Status: http.StatusNotFound,
		Title:  "Not Found",
		Detail: "User not found",
	}

	errMsg := pd.Error()

	if !strings.Contains(errMsg, "404") {
		t.Errorf("error message should contain status code, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "Not Found") {
		t.Errorf("error message should contain title, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "User not found") {
		t.Errorf("error message should contain detail, got: %s", errMsg)
	}
}

func TestProblemDetails_Error_EmptyDetail(t *testing.T) {
	t.Parallel()

	pd := &ProblemDetails{
		Status: http.StatusUnauthorized,
		Title:  "Unauthorized",
		Detail: "",
	}

	errMsg := pd.Error()

	// Should still produce valid error string
	if !strings.Contains(errMsg, "401") {
		t.Errorf("error message should contain status code, got: %s", errMsg)
	}
}

// ============================================================================
// WriteJSON Tests
// ============================================================================

func TestProblemDetails_WriteJSON_SetsContentType(t *testing.T) {
	t.Parallel()

	pd := NewNotFoundError("resource")
	rr := httptest.NewRecorder()

	pd.WriteJSON(rr)

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/problem+json" {
		t.Errorf("expected Content-Type 'application/problem+json', got %q", contentType)
	}
}

func TestProblemDetails_WriteJSON_SetsStatusCode(t *testing.T) {
	t.Parallel()

	pd := NewForbiddenError("access denied")
	rr := httptest.NewRecorder()

	pd.WriteJSON(rr)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestProblemDetails_WriteJSON_EncodesBody(t *testing.T) {
	t.Parallel()

	pd := NewBadRequestError("invalid input")
	rr := httptest.NewRecorder()

	pd.WriteJSON(rr)

	var result ProblemDetails
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if result.Title != "Bad Request" {
		t.Errorf("expected title 'Bad Request', got %q", result.Title)
	}
	if result.Detail != "invalid input" {
		t.Errorf("expected detail 'invalid input', got %q", result.Detail)
	}
}

// ============================================================================
// Constructor Tests - NewUnauthorizedError
// ============================================================================

func TestNewUnauthorizedError_ReturnsCorrectValues(t *testing.T) {
	t.Parallel()

	pd := NewUnauthorizedError("token expired")

	if pd.Status != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, pd.Status)
	}
	if pd.Title != "Unauthorized" {
		t.Errorf("expected title 'Unauthorized', got %q", pd.Title)
	}
	if pd.Detail != "token expired" {
		t.Errorf("expected detail 'token expired', got %q", pd.Detail)
	}
	if pd.Code != ErrCodeUnauthorized {
		t.Errorf("expected code %d, got %d", ErrCodeUnauthorized, pd.Code)
	}
	if !strings.Contains(pd.Type, "unauthorized") {
		t.Errorf("expected type to contain 'unauthorized', got %q", pd.Type)
	}
}

// ============================================================================
// Constructor Tests - NewForbiddenError
// ============================================================================

func TestNewForbiddenError_ReturnsCorrectValues(t *testing.T) {
	t.Parallel()

	pd := NewForbiddenError("access denied")

	if pd.Status != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, pd.Status)
	}
	if pd.Title != "Forbidden" {
		t.Errorf("expected title 'Forbidden', got %q", pd.Title)
	}
	if pd.Detail != "access denied" {
		t.Errorf("expected detail 'access denied', got %q", pd.Detail)
	}
	if pd.Code != ErrCodeForbidden {
		t.Errorf("expected code %d, got %d", ErrCodeForbidden, pd.Code)
	}
}

// ============================================================================
// Constructor Tests - NewNotFoundError
// ============================================================================

func TestNewNotFoundError_ReturnsCorrectValues(t *testing.T) {
	t.Parallel()

	pd := NewNotFoundError("user")

	if pd.Status != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, pd.Status)
	}
	if pd.Title != "Not Found" {
		t.Errorf("expected title 'Not Found', got %q", pd.Title)
	}
	if pd.Detail != "user not found" {
		t.Errorf("expected detail 'user not found', got %q", pd.Detail)
	}
	if pd.Code != ErrCodeNotFound {
		t.Errorf("expected code %d, got %d", ErrCodeNotFound, pd.Code)
	}
}

func TestNewNotFoundError_FormatsResourceName(t *testing.T) {
	t.Parallel()

	pd := NewNotFoundError("guild")

	if !strings.Contains(pd.Detail, "guild") {
		t.Errorf("detail should contain resource name, got %q", pd.Detail)
	}
	if !strings.Contains(pd.Detail, "not found") {
		t.Errorf("detail should contain 'not found', got %q", pd.Detail)
	}
}

// ============================================================================
// Constructor Tests - NewValidationError
// ============================================================================

func TestNewValidationError_SingleField_ReturnsCorrectValues(t *testing.T) {
	t.Parallel()

	errors := []FieldError{
		{Field: "email", Message: "invalid format"},
	}
	pd := NewValidationError(errors)

	if pd.Status != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, pd.Status)
	}
	if pd.Title != "Validation Error" {
		t.Errorf("expected title 'Validation Error', got %q", pd.Title)
	}
	if pd.Code != ErrCodeValidation {
		t.Errorf("expected code %d, got %d", ErrCodeValidation, pd.Code)
	}
	if len(pd.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(pd.Errors))
	}
	if !strings.Contains(pd.Detail, "email") {
		t.Errorf("detail should contain field name, got %q", pd.Detail)
	}
	if !strings.Contains(pd.Detail, "invalid format") {
		t.Errorf("detail should contain error message, got %q", pd.Detail)
	}
}

func TestNewValidationError_MultipleFields_SummarizesCount(t *testing.T) {
	t.Parallel()

	errors := []FieldError{
		{Field: "email", Message: "required"},
		{Field: "name", Message: "too short"},
		{Field: "age", Message: "must be positive"},
	}
	pd := NewValidationError(errors)

	if len(pd.Errors) != 3 {
		t.Errorf("expected 3 errors, got %d", len(pd.Errors))
	}
	if !strings.Contains(pd.Detail, "2 more errors") {
		t.Errorf("detail should mention count of additional errors, got %q", pd.Detail)
	}
}

func TestNewValidationError_EmptyErrors_ReturnsDefaultMessage(t *testing.T) {
	t.Parallel()

	pd := NewValidationError([]FieldError{})

	if pd.Detail != "One or more fields failed validation" {
		t.Errorf("expected default detail message, got %q", pd.Detail)
	}
	if len(pd.Errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(pd.Errors))
	}
}

// ============================================================================
// Constructor Tests - NewLimitExceededError
// ============================================================================

func TestNewLimitExceededError_ReturnsCorrectValues(t *testing.T) {
	t.Parallel()

	pd := NewLimitExceededError("guilds", 5, 5)

	if pd.Status != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, pd.Status)
	}
	if pd.Title != "Limit Exceeded" {
		t.Errorf("expected title 'Limit Exceeded', got %q", pd.Title)
	}
	if pd.Code != ErrCodeLimitExceeded {
		t.Errorf("expected code %d, got %d", ErrCodeLimitExceeded, pd.Code)
	}
	if pd.Limit == nil || *pd.Limit != 5 {
		t.Errorf("expected limit 5, got %v", pd.Limit)
	}
	if pd.Current == nil || *pd.Current != 5 {
		t.Errorf("expected current 5, got %v", pd.Current)
	}
	if !strings.Contains(pd.Detail, "5") {
		t.Errorf("detail should contain limit number, got %q", pd.Detail)
	}
	if !strings.Contains(pd.Detail, "guilds") {
		t.Errorf("detail should contain resource name, got %q", pd.Detail)
	}
}

// ============================================================================
// Constructor Tests - NewConflictError
// ============================================================================

func TestNewConflictError_ReturnsCorrectValues(t *testing.T) {
	t.Parallel()

	pd := NewConflictError("email already in use")

	if pd.Status != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, pd.Status)
	}
	if pd.Title != "Conflict" {
		t.Errorf("expected title 'Conflict', got %q", pd.Title)
	}
	if pd.Detail != "email already in use" {
		t.Errorf("expected detail 'email already in use', got %q", pd.Detail)
	}
	if pd.Code != ErrCodeConflict {
		t.Errorf("expected code %d, got %d", ErrCodeConflict, pd.Code)
	}
}

// ============================================================================
// Constructor Tests - NewInternalError
// ============================================================================

func TestNewInternalError_ReturnsCorrectValues(t *testing.T) {
	t.Parallel()

	pd := NewInternalError("database connection failed")

	if pd.Status != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, pd.Status)
	}
	if pd.Title != "Internal Server Error" {
		t.Errorf("expected title 'Internal Server Error', got %q", pd.Title)
	}
	if pd.Detail != "database connection failed" {
		t.Errorf("expected detail 'database connection failed', got %q", pd.Detail)
	}
	if pd.Code != ErrCodeInternal {
		t.Errorf("expected code %d, got %d", ErrCodeInternal, pd.Code)
	}
}

func TestNewInternalError_EmptyDetail_UsesDefault(t *testing.T) {
	t.Parallel()

	pd := NewInternalError("")

	if pd.Detail != "An unexpected error occurred" {
		t.Errorf("expected default detail message, got %q", pd.Detail)
	}
}

// ============================================================================
// Constructor Tests - NewBadRequestError
// ============================================================================

func TestNewBadRequestError_ReturnsCorrectValues(t *testing.T) {
	t.Parallel()

	pd := NewBadRequestError("missing required field")

	if pd.Status != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, pd.Status)
	}
	if pd.Title != "Bad Request" {
		t.Errorf("expected title 'Bad Request', got %q", pd.Title)
	}
	if pd.Detail != "missing required field" {
		t.Errorf("expected detail 'missing required field', got %q", pd.Detail)
	}
	if pd.Code != ErrCodeInvalidInput {
		t.Errorf("expected code %d, got %d", ErrCodeInvalidInput, pd.Code)
	}
}

// ============================================================================
// Constructor Tests - NewMethodNotAllowedError
// ============================================================================

func TestNewMethodNotAllowedError_ReturnsCorrectValues(t *testing.T) {
	t.Parallel()

	pd := NewMethodNotAllowedError("POST")

	if pd.Status != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, pd.Status)
	}
	if pd.Title != "Method Not Allowed" {
		t.Errorf("expected title 'Method Not Allowed', got %q", pd.Title)
	}
	if !strings.Contains(pd.Detail, "POST") {
		t.Errorf("detail should contain allowed method, got %q", pd.Detail)
	}
}

// ============================================================================
// Constructor Tests - NewRateLimitError
// ============================================================================

func TestNewRateLimitError_ReturnsCorrectValues(t *testing.T) {
	t.Parallel()

	pd := NewRateLimitError(60)

	if pd.Status != http.StatusTooManyRequests {
		t.Errorf("expected status %d, got %d", http.StatusTooManyRequests, pd.Status)
	}
	if pd.Title != "Too Many Requests" {
		t.Errorf("expected title 'Too Many Requests', got %q", pd.Title)
	}
	if !strings.Contains(pd.Detail, "60") {
		t.Errorf("detail should contain retry seconds, got %q", pd.Detail)
	}
}

// ============================================================================
// Error Code Constants Tests
// ============================================================================

func TestErrorCodes_UniqueValues(t *testing.T) {
	t.Parallel()

	codes := map[ErrorCode]string{
		ErrCodeUnauthorized:  "ErrCodeUnauthorized",
		ErrCodeTokenExpired:  "ErrCodeTokenExpired",
		ErrCodeTokenInvalid:  "ErrCodeTokenInvalid",
		ErrCodeLoginFailed:   "ErrCodeLoginFailed",
		ErrCodeForbidden:     "ErrCodeForbidden",
		ErrCodeNotMember:     "ErrCodeNotMember",
		ErrCodeNotFound:      "ErrCodeNotFound",
		ErrCodeAlreadyExists: "ErrCodeAlreadyExists",
		ErrCodeConflict:      "ErrCodeConflict",
		ErrCodeValidation:    "ErrCodeValidation",
		ErrCodeInvalidInput:  "ErrCodeInvalidInput",
		ErrCodeLimitExceeded: "ErrCodeLimitExceeded",
		ErrCodeInternal:      "ErrCodeInternal",
		ErrCodeDatabase:      "ErrCodeDatabase",
		ErrCodeExternalAPI:   "ErrCodeExternalAPI",
	}

	seen := make(map[ErrorCode]string)
	for code, name := range codes {
		if existing, exists := seen[code]; exists {
			t.Errorf("duplicate error code: %s and %s both have value %d", existing, name, code)
		}
		seen[code] = name
	}
}

func TestErrorCodes_CorrectRanges(t *testing.T) {
	t.Parallel()

	// Authentication errors should be 1xxx
	authCodes := []ErrorCode{ErrCodeUnauthorized, ErrCodeTokenExpired, ErrCodeTokenInvalid, ErrCodeLoginFailed}
	for _, code := range authCodes {
		if code < 1000 || code >= 2000 {
			t.Errorf("auth error code %d should be in 1xxx range", code)
		}
	}

	// Authorization errors should be 2xxx
	authzCodes := []ErrorCode{ErrCodeForbidden, ErrCodeNotMember}
	for _, code := range authzCodes {
		if code < 2000 || code >= 3000 {
			t.Errorf("authz error code %d should be in 2xxx range", code)
		}
	}

	// Resource errors should be 3xxx
	resourceCodes := []ErrorCode{ErrCodeNotFound, ErrCodeAlreadyExists, ErrCodeConflict}
	for _, code := range resourceCodes {
		if code < 3000 || code >= 4000 {
			t.Errorf("resource error code %d should be in 3xxx range", code)
		}
	}

	// Validation errors should be 4xxx
	validationCodes := []ErrorCode{ErrCodeValidation, ErrCodeInvalidInput, ErrCodeLimitExceeded}
	for _, code := range validationCodes {
		if code < 4000 || code >= 5000 {
			t.Errorf("validation error code %d should be in 4xxx range", code)
		}
	}

	// Internal errors should be 5xxx
	internalCodes := []ErrorCode{ErrCodeInternal, ErrCodeDatabase, ErrCodeExternalAPI}
	for _, code := range internalCodes {
		if code < 5000 || code >= 6000 {
			t.Errorf("internal error code %d should be in 5xxx range", code)
		}
	}
}

// ============================================================================
// JSON Serialization Tests
// ============================================================================

func TestProblemDetails_JSON_OmitsEmptyFields(t *testing.T) {
	t.Parallel()

	pd := &ProblemDetails{
		Type:   "test",
		Title:  "Test",
		Status: 400,
		// Detail, Instance, Errors, etc. are empty
	}

	data, err := json.Marshal(pd)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	jsonStr := string(data)
	if strings.Contains(jsonStr, "detail") {
		t.Error("empty detail should be omitted from JSON")
	}
	if strings.Contains(jsonStr, "instance") {
		t.Error("empty instance should be omitted from JSON")
	}
	if strings.Contains(jsonStr, "errors") {
		t.Error("empty errors should be omitted from JSON")
	}
}

func TestProblemDetails_JSON_IncludesAllFields(t *testing.T) {
	t.Parallel()

	limit := 10
	current := 5
	pd := &ProblemDetails{
		Type:     "test-type",
		Title:    "Test Title",
		Status:   422,
		Detail:   "Test detail",
		Instance: "/api/test",
		Errors:   []FieldError{{Field: "name", Message: "required"}},
		Code:     ErrCodeValidation,
		Limit:    &limit,
		Current:  &current,
	}

	data, err := json.Marshal(pd)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedFields := []string{"type", "title", "status", "detail", "instance", "errors", "code", "limit", "current"}
	for _, field := range expectedFields {
		if _, ok := result[field]; !ok {
			t.Errorf("expected field %q in JSON output", field)
		}
	}
}
