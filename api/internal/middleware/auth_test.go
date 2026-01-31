package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgo/saga/api/pkg/jwt"
)

// ============================================================================
// Mock AuthService
// ============================================================================

type mockAuthService struct {
	validateFunc func(token string) (*jwt.Claims, error)
}

func (m *mockAuthService) ValidateAccessToken(token string) (*jwt.Claims, error) {
	return m.validateFunc(token)
}

// successAuthService returns valid claims for any token
func successAuthService(userID, email string) *mockAuthService {
	return &mockAuthService{
		validateFunc: func(token string) (*jwt.Claims, error) {
			return &jwt.Claims{
				UserID: userID,
				Email:  email,
			}, nil
		},
	}
}

// errorAuthService returns the specified error
func errorAuthService(err error) *mockAuthService {
	return &mockAuthService{
		validateFunc: func(token string) (*jwt.Claims, error) {
			return nil, err
		},
	}
}

// ============================================================================
// Test Helpers
// ============================================================================

func newTestRequest(authHeader string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	return req
}

// captureHandler captures the request context for inspection
type captureHandler struct {
	called bool
	ctx    context.Context
}

func (h *captureHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.called = true
	h.ctx = r.Context()
	w.WriteHeader(http.StatusOK)
}

// ============================================================================
// Auth() Middleware Tests
// ============================================================================

func TestAuth_MissingAuthorizationHeader_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	authSvc := successAuthService("user:123", "test@example.com")
	middleware := Auth(authSvc)
	handler := &captureHandler{}

	req := newTestRequest("") // No auth header
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if handler.called {
		t.Error("handler should not have been called")
	}
}

func TestAuth_InvalidHeaderFormat_NoBearerPrefix_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	authSvc := successAuthService("user:123", "test@example.com")
	middleware := Auth(authSvc)
	handler := &captureHandler{}

	req := newTestRequest("Basic sometoken") // Wrong scheme
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if handler.called {
		t.Error("handler should not have been called")
	}
}

func TestAuth_InvalidHeaderFormat_OnlyBearer_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	authSvc := successAuthService("user:123", "test@example.com")
	middleware := Auth(authSvc)
	handler := &captureHandler{}

	req := newTestRequest("Bearer") // No token
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if handler.called {
		t.Error("handler should not have been called")
	}
}

func TestAuth_InvalidHeaderFormat_BearerNoSpace_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	authSvc := successAuthService("user:123", "test@example.com")
	middleware := Auth(authSvc)
	handler := &captureHandler{}

	req := newTestRequest("Bearertoken") // No space
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if handler.called {
		t.Error("handler should not have been called")
	}
}

func TestAuth_ValidToken_SetsContext_CallsNext(t *testing.T) {
	t.Parallel()
	authSvc := successAuthService("user:123", "test@example.com")
	middleware := Auth(authSvc)
	handler := &captureHandler{}

	req := newTestRequest("Bearer valid-token")
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !handler.called {
		t.Error("handler should have been called")
	}

	// Check context values
	if GetUserID(handler.ctx) != "user:123" {
		t.Errorf("expected UserID 'user:123', got %q", GetUserID(handler.ctx))
	}
	if GetUserEmail(handler.ctx) != "test@example.com" {
		t.Errorf("expected Email 'test@example.com', got %q", GetUserEmail(handler.ctx))
	}
}

func TestAuth_ValidToken_CaseInsensitiveBearer(t *testing.T) {
	t.Parallel()
	authSvc := successAuthService("user:123", "test@example.com")
	middleware := Auth(authSvc)
	handler := &captureHandler{}

	// Test lowercase "bearer"
	req := newTestRequest("bearer valid-token")
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !handler.called {
		t.Error("handler should have been called")
	}
}

func TestAuth_ExpiredToken_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	authSvc := errorAuthService(jwt.ErrTokenExpired)
	middleware := Auth(authSvc)
	handler := &captureHandler{}

	req := newTestRequest("Bearer expired-token")
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if handler.called {
		t.Error("handler should not have been called")
	}
}

func TestAuth_InvalidSignature_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	authSvc := errorAuthService(jwt.ErrInvalidSignature)
	middleware := Auth(authSvc)
	handler := &captureHandler{}

	req := newTestRequest("Bearer invalid-signature")
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if handler.called {
		t.Error("handler should not have been called")
	}
}

func TestAuth_GenericError_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	authSvc := errorAuthService(jwt.ErrInvalidToken)
	middleware := Auth(authSvc)
	handler := &captureHandler{}

	req := newTestRequest("Bearer bad-token")
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if handler.called {
		t.Error("handler should not have been called")
	}
}

func TestAuth_SetsClaims_InContext(t *testing.T) {
	t.Parallel()
	expectedClaims := &jwt.Claims{
		UserID:   "user:456",
		Email:    "user@test.com",
		Username: "testuser",
		Subject:  "sub:456",
	}
	authSvc := &mockAuthService{
		validateFunc: func(token string) (*jwt.Claims, error) {
			return expectedClaims, nil
		},
	}
	middleware := Auth(authSvc)
	handler := &captureHandler{}

	req := newTestRequest("Bearer valid-token")
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	claims := GetClaims(handler.ctx)
	if claims == nil {
		t.Fatal("expected claims in context")
	}
	if claims.UserID != expectedClaims.UserID {
		t.Errorf("UserID: expected %q, got %q", expectedClaims.UserID, claims.UserID)
	}
	if claims.Email != expectedClaims.Email {
		t.Errorf("Email: expected %q, got %q", expectedClaims.Email, claims.Email)
	}
	if claims.Username != expectedClaims.Username {
		t.Errorf("Username: expected %q, got %q", expectedClaims.Username, claims.Username)
	}
}

// ============================================================================
// OptionalAuth() Middleware Tests
// ============================================================================

func TestOptionalAuth_NoHeader_Proceeds(t *testing.T) {
	t.Parallel()
	authSvc := successAuthService("user:123", "test@example.com")
	middleware := OptionalAuth(authSvc)
	handler := &captureHandler{}

	req := newTestRequest("") // No auth header
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !handler.called {
		t.Error("handler should have been called")
	}

	// Context should NOT have user info
	if GetUserID(handler.ctx) != "" {
		t.Errorf("expected empty UserID, got %q", GetUserID(handler.ctx))
	}
}

func TestOptionalAuth_InvalidHeaderFormat_Proceeds(t *testing.T) {
	t.Parallel()
	authSvc := successAuthService("user:123", "test@example.com")
	middleware := OptionalAuth(authSvc)
	handler := &captureHandler{}

	req := newTestRequest("Basic sometoken") // Wrong scheme
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !handler.called {
		t.Error("handler should have been called")
	}

	// Context should NOT have user info
	if GetUserID(handler.ctx) != "" {
		t.Errorf("expected empty UserID, got %q", GetUserID(handler.ctx))
	}
}

func TestOptionalAuth_ValidToken_SetsContext(t *testing.T) {
	t.Parallel()
	authSvc := successAuthService("user:123", "test@example.com")
	middleware := OptionalAuth(authSvc)
	handler := &captureHandler{}

	req := newTestRequest("Bearer valid-token")
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !handler.called {
		t.Error("handler should have been called")
	}

	// Context SHOULD have user info
	if GetUserID(handler.ctx) != "user:123" {
		t.Errorf("expected UserID 'user:123', got %q", GetUserID(handler.ctx))
	}
	if GetUserEmail(handler.ctx) != "test@example.com" {
		t.Errorf("expected Email 'test@example.com', got %q", GetUserEmail(handler.ctx))
	}
}

func TestOptionalAuth_InvalidToken_ProceedsWithoutAuth(t *testing.T) {
	t.Parallel()
	authSvc := errorAuthService(jwt.ErrInvalidToken)
	middleware := OptionalAuth(authSvc)
	handler := &captureHandler{}

	req := newTestRequest("Bearer invalid-token")
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	// Should still proceed
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !handler.called {
		t.Error("handler should have been called")
	}

	// Context should NOT have user info
	if GetUserID(handler.ctx) != "" {
		t.Errorf("expected empty UserID, got %q", GetUserID(handler.ctx))
	}
}

func TestOptionalAuth_ExpiredToken_ProceedsWithoutAuth(t *testing.T) {
	t.Parallel()
	authSvc := errorAuthService(jwt.ErrTokenExpired)
	middleware := OptionalAuth(authSvc)
	handler := &captureHandler{}

	req := newTestRequest("Bearer expired-token")
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	// Should still proceed
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !handler.called {
		t.Error("handler should have been called")
	}

	// Context should NOT have user info
	if GetUserID(handler.ctx) != "" {
		t.Errorf("expected empty UserID, got %q", GetUserID(handler.ctx))
	}
}

// ============================================================================
// Context Helper Tests
// ============================================================================

func TestGetUserID_Present_ReturnsValue(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), UserIDKey, "user:999")

	result := GetUserID(ctx)

	if result != "user:999" {
		t.Errorf("expected 'user:999', got %q", result)
	}
}

func TestGetUserID_Missing_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	result := GetUserID(ctx)

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestGetUserID_WrongType_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), UserIDKey, 12345) // Wrong type

	result := GetUserID(ctx)

	if result != "" {
		t.Errorf("expected empty string for wrong type, got %q", result)
	}
}

func TestGetUserEmail_Present_ReturnsValue(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), UserEmailKey, "user@test.com")

	result := GetUserEmail(ctx)

	if result != "user@test.com" {
		t.Errorf("expected 'user@test.com', got %q", result)
	}
}

func TestGetUserEmail_Missing_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	result := GetUserEmail(ctx)

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestGetClaims_Present_ReturnsClaims(t *testing.T) {
	t.Parallel()
	expectedClaims := &jwt.Claims{
		UserID: "user:123",
		Email:  "test@example.com",
	}
	ctx := context.WithValue(context.Background(), ClaimsKey, expectedClaims)

	result := GetClaims(ctx)

	if result == nil {
		t.Fatal("expected claims, got nil")
	}
	if result.UserID != expectedClaims.UserID {
		t.Errorf("expected UserID %q, got %q", expectedClaims.UserID, result.UserID)
	}
}

func TestGetClaims_Missing_ReturnsNil(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	result := GetClaims(ctx)

	if result != nil {
		t.Errorf("expected nil, got %+v", result)
	}
}

func TestGetClaims_WrongType_ReturnsNil(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), ClaimsKey, "not claims") // Wrong type

	result := GetClaims(ctx)

	if result != nil {
		t.Errorf("expected nil for wrong type, got %+v", result)
	}
}
