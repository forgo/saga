// Package helpers provides common test utilities for e2e testing.
//
// This package includes HTTP request builders, response validators,
// and assertion helpers for testing API endpoints.
package helpers

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/pkg/jwt"
)

// ============================================================================
// JWT Helpers
// ============================================================================

// JWTHelper provides JWT token generation for tests
type JWTHelper struct {
	privateKey *rsa.PrivateKey
	issuer     string
}

// NewJWTHelper creates a new JWT helper with an in-memory key
func NewJWTHelper(t *testing.T) *JWTHelper {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("helpers: failed to generate RSA key: %v", err)
	}

	return &JWTHelper{
		privateKey: privateKey,
		issuer:     "saga-test",
	}
}

// GenerateToken creates a valid JWT token for testing
func (h *JWTHelper) GenerateToken(user *model.User) string {
	claims := jwt.Claims{
		Subject:   user.ID,
		UserID:    user.ID,
		Email:     user.Email,
		Issuer:    h.issuer,
		IssuedAt:  time.Now().Unix(),
		NotBefore: time.Now().Unix(),
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
	}
	if user.Username != nil {
		claims.Username = *user.Username
	}

	return h.signToken(claims)
}

// GenerateExpiredToken creates an expired JWT token for testing
func (h *JWTHelper) GenerateExpiredToken(user *model.User) string {
	claims := jwt.Claims{
		Subject:   user.ID,
		UserID:    user.ID,
		Email:     user.Email,
		Issuer:    h.issuer,
		IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
		NotBefore: time.Now().Add(-2 * time.Hour).Unix(),
		ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(), // Expired
	}

	return h.signToken(claims)
}

// signToken creates a signed JWT
func (h *JWTHelper) signToken(claims jwt.Claims) string {
	claimsJSON, _ := json.Marshal(claims)
	header := `{"alg":"RS256","typ":"JWT"}`

	headerB64 := base64URLEncode([]byte(header))
	claimsB64 := base64URLEncode(claimsJSON)

	message := headerB64 + "." + claimsB64

	hash := sha256.Sum256([]byte(message))
	signature, _ := rsa.SignPKCS1v15(rand.Reader, h.privateKey, crypto.SHA256, hash[:])

	return message + "." + base64URLEncode(signature)
}

// base64URLEncode encodes data as base64url without padding
func base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

// ============================================================================
// HTTP Request Helpers
// ============================================================================

// RequestBuilder helps construct HTTP requests for testing
type RequestBuilder struct {
	t       *testing.T
	method  string
	path    string
	body    interface{}
	headers map[string]string
	jwt     *JWTHelper
	user    *model.User
}

// NewRequest creates a new request builder
func NewRequest(t *testing.T, method, path string) *RequestBuilder {
	t.Helper()
	return &RequestBuilder{
		t:       t,
		method:  method,
		path:    path,
		headers: make(map[string]string),
	}
}

// WithBody sets the request body (will be JSON encoded)
func (rb *RequestBuilder) WithBody(body interface{}) *RequestBuilder {
	rb.body = body
	return rb
}

// WithHeader adds a header to the request
func (rb *RequestBuilder) WithHeader(key, value string) *RequestBuilder {
	rb.headers[key] = value
	return rb
}

// WithAuth adds authentication for the given user
func (rb *RequestBuilder) WithAuth(jwt *JWTHelper, user *model.User) *RequestBuilder {
	rb.jwt = jwt
	rb.user = user
	return rb
}

// Build creates the HTTP request
func (rb *RequestBuilder) Build() *http.Request {
	rb.t.Helper()

	var bodyReader io.Reader
	if rb.body != nil {
		bodyBytes, err := json.Marshal(rb.body)
		if err != nil {
			rb.t.Fatalf("helpers: failed to marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req := httptest.NewRequest(rb.method, rb.path, bodyReader)

	// Set content type for requests with body
	if rb.body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add custom headers
	for k, v := range rb.headers {
		req.Header.Set(k, v)
	}

	// Add auth header
	if rb.jwt != nil && rb.user != nil {
		token := rb.jwt.GenerateToken(rb.user)
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req
}

// ============================================================================
// Response Assertion Helpers
// ============================================================================

// AssertStatus checks that the response has the expected status code
func AssertStatus(t *testing.T, resp *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if resp.Code != expected {
		t.Errorf("expected status %d, got %d. Body: %s", expected, resp.Code, resp.Body.String())
	}
}

// AssertProblemDetails validates an RFC 9457 Problem Details error response
func AssertProblemDetails(t *testing.T, resp *httptest.ResponseRecorder, expectedStatus int, expectedCode model.ErrorCode) {
	t.Helper()

	AssertStatus(t, resp, expectedStatus)

	var problem model.ProblemDetails
	bodyBytes := resp.Body.Bytes()
	if err := json.Unmarshal(bodyBytes, &problem); err != nil {
		t.Fatalf("failed to decode problem details: %v. Body: %s", err, string(bodyBytes))
	}

	if problem.Status != expectedStatus {
		t.Errorf("expected problem.status %d, got %d", expectedStatus, problem.Status)
	}

	if expectedCode != 0 && problem.Code != expectedCode {
		t.Errorf("expected problem.code %d, got %d", expectedCode, problem.Code)
	}
}

// AssertValidationError checks for a validation error on a specific field
func AssertValidationError(t *testing.T, resp *httptest.ResponseRecorder, field string) {
	t.Helper()

	AssertStatus(t, resp, http.StatusUnprocessableEntity)

	var problem model.ProblemDetails
	bodyBytes := resp.Body.Bytes()
	if err := json.Unmarshal(bodyBytes, &problem); err != nil {
		t.Fatalf("failed to decode problem details: %v", err)
	}

	for _, fe := range problem.Errors {
		if fe.Field == field {
			return // Found the expected field error
		}
	}

	t.Errorf("expected validation error on field %q, but not found. Errors: %+v", field, problem.Errors)
}

// AssertJSONContains checks that the response body contains expected key-value pairs
func AssertJSONContains(t *testing.T, resp *httptest.ResponseRecorder, expected map[string]interface{}) {
	t.Helper()

	var actual map[string]interface{}
	bodyBytes := resp.Body.Bytes()
	if err := json.Unmarshal(bodyBytes, &actual); err != nil {
		t.Fatalf("failed to decode response: %v. Body: %s", err, string(bodyBytes))
	}

	for key, expectedVal := range expected {
		actualVal, ok := actual[key]
		if !ok {
			t.Errorf("expected key %q not found in response", key)
			continue
		}

		if !jsonEqual(expectedVal, actualVal) {
			t.Errorf("for key %q: expected %v, got %v", key, expectedVal, actualVal)
		}
	}
}

// DecodeResponse decodes the response body into the given struct
func DecodeResponse(t *testing.T, resp *httptest.ResponseRecorder, v interface{}) {
	t.Helper()

	bodyBytes := resp.Body.Bytes()
	if err := json.Unmarshal(bodyBytes, v); err != nil {
		t.Fatalf("failed to decode response: %v. Body: %s", err, string(bodyBytes))
	}
}

// GetDataFromResponse extracts the "data" field from a standard response
func GetDataFromResponse(t *testing.T, resp *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()

	var response struct {
		Data map[string]interface{} `json:"data"`
	}
	bodyBytes := resp.Body.Bytes()
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		t.Fatalf("failed to decode response: %v. Body: %s", err, string(bodyBytes))
	}

	return response.Data
}

// ============================================================================
// Database Assertion Helpers
// ============================================================================

// AssertRecordExists checks that a record exists in the database
func AssertRecordExists(t *testing.T, db database.Database, table, id string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Parse the ID to extract just the record part if it's a full thing ID
	recordID := id
	if strings.Contains(id, ":") {
		parts := strings.SplitN(id, ":", 2)
		if len(parts) == 2 {
			recordID = parts[1]
		}
	}

	query := "SELECT * FROM type::record($table, $id)"
	results, err := db.Query(ctx, query, map[string]interface{}{
		"table": table,
		"id":    recordID,
	})
	if err != nil {
		t.Fatalf("failed to query for record: %v", err)
	}

	// Check if we got results
	if !hasResults(results) {
		t.Errorf("expected record %s:%s to exist, but it doesn't", table, recordID)
	}
}

// AssertRecordNotExists checks that a record does not exist
func AssertRecordNotExists(t *testing.T, db database.Database, table, id string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	recordID := id
	if strings.Contains(id, ":") {
		parts := strings.SplitN(id, ":", 2)
		if len(parts) == 2 {
			recordID = parts[1]
		}
	}

	query := "SELECT * FROM type::record($table, $id)"
	results, err := db.Query(ctx, query, map[string]interface{}{
		"table": table,
		"id":    recordID,
	})
	if err != nil {
		// Query error might mean not found, which is what we want
		return
	}

	if hasResults(results) {
		t.Errorf("expected record %s:%s to not exist, but it does", table, recordID)
	}
}

// hasResults checks if SurrealDB query returned any results
func hasResults(results []interface{}) bool {
	if len(results) == 0 {
		return false
	}

	resp, ok := results[0].(map[string]interface{})
	if !ok {
		return false
	}

	result, ok := resp["result"]
	if !ok {
		return false
	}

	switch v := result.(type) {
	case []interface{}:
		return len(v) > 0
	case map[string]interface{}:
		return true
	case nil:
		return false
	default:
		return true
	}
}

// ============================================================================
// Service Factory Helpers
// ============================================================================

// NewTestJWTService creates a JWT service with in-memory keys for testing
func NewTestJWTService(t *testing.T) *jwt.Service {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("helpers: failed to generate RSA key: %v", err)
	}

	return jwt.NewTestService(privateKey, "saga-test", 15*time.Minute)
}

// ============================================================================
// Utility Helpers
// ============================================================================

// jsonEqual compares two JSON values for equality
func jsonEqual(a, b interface{}) bool {
	aBytes, _ := json.Marshal(a)
	bBytes, _ := json.Marshal(b)
	return string(aBytes) == string(bBytes)
}

// StringPtr returns a pointer to the string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the int
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to the bool
func BoolPtr(b bool) *bool {
	return &b
}

// TimePtr returns a pointer to the time
func TimePtr(t time.Time) *time.Time {
	return &t
}

// MustParseTime parses a time string or fails the test
func MustParseTime(t *testing.T, layout, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(layout, value)
	if err != nil {
		t.Fatalf("failed to parse time %q: %v", value, err)
	}
	return parsed
}
