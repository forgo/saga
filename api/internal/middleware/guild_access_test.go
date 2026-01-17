package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ============================================================================
// Mock GuildMembershipChecker
// ============================================================================

type mockGuildMembershipChecker struct {
	isMemberFunc func(ctx context.Context, userID, guildID string) (bool, error)
}

func (m *mockGuildMembershipChecker) IsMember(ctx context.Context, userID, guildID string) (bool, error) {
	return m.isMemberFunc(ctx, userID, guildID)
}

// successMembershipChecker always returns true (is a member)
func successMembershipChecker() *mockGuildMembershipChecker {
	return &mockGuildMembershipChecker{
		isMemberFunc: func(ctx context.Context, userID, guildID string) (bool, error) {
			return true, nil
		},
	}
}

// notMemberChecker always returns false (not a member)
func notMemberChecker() *mockGuildMembershipChecker {
	return &mockGuildMembershipChecker{
		isMemberFunc: func(ctx context.Context, userID, guildID string) (bool, error) {
			return false, nil
		},
	}
}

// errorMembershipChecker returns an error
func errorMembershipChecker(err error) *mockGuildMembershipChecker {
	return &mockGuildMembershipChecker{
		isMemberFunc: func(ctx context.Context, userID, guildID string) (bool, error) {
			return false, err
		},
	}
}

// ============================================================================
// GuildAccess Middleware Tests
// ============================================================================

func TestGuildAccess_NoUserID_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()
	checker := successMembershipChecker()
	middleware := GuildAccess(checker)
	handler := &captureHandler{}

	req := httptest.NewRequest(http.MethodGet, "/v1/guilds/guild:123", nil)
	// No user ID in context
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	if handler.called {
		t.Error("handler should not have been called")
	}
}

func TestGuildAccess_InvalidGuildID_ReturnsBadRequest(t *testing.T) {
	t.Parallel()
	checker := successMembershipChecker()
	middleware := GuildAccess(checker)
	handler := &captureHandler{}

	// Path without guild ID
	req := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, "user:123")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if handler.called {
		t.Error("handler should not have been called")
	}
}

func TestGuildAccess_MembershipCheckError_ReturnsNotFound(t *testing.T) {
	t.Parallel()
	checker := errorMembershipChecker(errors.New("database error"))
	middleware := GuildAccess(checker)
	handler := &captureHandler{}

	req := httptest.NewRequest(http.MethodGet, "/v1/guilds/guild:123", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, "user:123")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	// Returns 404 to not leak information about errors
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
	if handler.called {
		t.Error("handler should not have been called")
	}
}

func TestGuildAccess_NotMember_ReturnsNotFound(t *testing.T) {
	t.Parallel()
	checker := notMemberChecker()
	middleware := GuildAccess(checker)
	handler := &captureHandler{}

	req := httptest.NewRequest(http.MethodGet, "/v1/guilds/guild:123", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, "user:123")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	// Returns 404 instead of 403 to not leak guild existence
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
	if handler.called {
		t.Error("handler should not have been called")
	}
}

func TestGuildAccess_IsMember_ProceedsWithGuildID(t *testing.T) {
	t.Parallel()
	checker := successMembershipChecker()
	middleware := GuildAccess(checker)
	handler := &captureHandler{}

	req := httptest.NewRequest(http.MethodGet, "/v1/guilds/guild:123/people", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, "user:456")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !handler.called {
		t.Error("handler should have been called")
	}

	// Check guild ID in context
	guildID := GetGuildID(handler.ctx)
	if guildID != "guild:123" {
		t.Errorf("expected guild ID 'guild:123', got %q", guildID)
	}
}

func TestGuildAccess_PassesCorrectIDsToChecker(t *testing.T) {
	t.Parallel()
	var receivedUserID, receivedGuildID string
	checker := &mockGuildMembershipChecker{
		isMemberFunc: func(ctx context.Context, userID, guildID string) (bool, error) {
			receivedUserID = userID
			receivedGuildID = guildID
			return true, nil
		},
	}
	middleware := GuildAccess(checker)
	handler := &captureHandler{}

	req := httptest.NewRequest(http.MethodGet, "/v1/guilds/guild:abc/people/person:xyz", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, "user:def")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if receivedUserID != "user:def" {
		t.Errorf("expected userID 'user:def', got %q", receivedUserID)
	}
	if receivedGuildID != "guild:abc" {
		t.Errorf("expected guildID 'guild:abc', got %q", receivedGuildID)
	}
}

// ============================================================================
// extractGuildID Tests
// ============================================================================

func TestExtractGuildID_BasicPath_ExtractsID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"guilds with ID", "/v1/guilds/guild:123", "guild:123"},
		{"guilds with people", "/v1/guilds/guild:abc/people", "guild:abc"},
		{"guilds with people and person ID", "/v1/guilds/guild:xyz/people/person:456", "guild:xyz"},
		{"guilds with activities", "/v1/guilds/guild:789/activities", "guild:789"},
		{"guilds with timers", "/v1/guilds/guild:101/timers", "guild:101"},
		{"guilds with members", "/v1/guilds/guild:202/members", "guild:202"},
		{"guilds with events", "/v1/guilds/guild:303/events", "guild:303"},
		{"simple ID", "/guilds/abc123", "abc123"},
		{"no v1 prefix", "/guilds/test-guild-id/people", "test-guild-id"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractGuildID(tt.path)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractGuildID_SkipsSubResourceNames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"guilds followed by people", "/v1/guilds/people", ""},
		{"guilds followed by activities", "/v1/guilds/activities", ""},
		{"guilds followed by timers", "/v1/guilds/timers", ""},
		{"guilds followed by members", "/v1/guilds/members", ""},
		{"guilds followed by events", "/v1/guilds/events", ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractGuildID(tt.path)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractGuildID_InvalidPaths_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
	}{
		{"no guilds segment", "/v1/users/me"},
		{"guilds at end", "/v1/guilds"},
		{"guilds with trailing slash", "/v1/guilds/"},
		{"empty path", ""},
		{"root path", "/"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractGuildID(tt.path)
			if result != "" {
				t.Errorf("expected empty string, got %q", result)
			}
		})
	}
}

// ============================================================================
// extractPersonID Tests
// ============================================================================

func TestExtractPersonID_ValidPaths_ExtractsID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"basic person path", "/v1/guilds/guild:123/people/person:456", "person:456"},
		{"person with timers", "/v1/guilds/guild:123/people/person:abc/timers", "person:abc"},
		{"simple ID", "/people/myPersonId", "myPersonId"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractPersonID(tt.path)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractPersonID_SkipsSubResourceNames(t *testing.T) {
	t.Parallel()
	// "timers" after people should be skipped as sub-resource name
	result := extractPersonID("/v1/guilds/guild:123/people/timers")
	if result != "" {
		t.Errorf("expected empty string for sub-resource name, got %q", result)
	}
}

func TestExtractPersonID_InvalidPaths_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
	}{
		{"no people segment", "/v1/guilds/guild:123"},
		{"people at end", "/v1/guilds/guild:123/people"},
		{"people with trailing slash", "/v1/guilds/guild:123/people/"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractPersonID(tt.path)
			if result != "" {
				t.Errorf("expected empty string, got %q", result)
			}
		})
	}
}

// ============================================================================
// extractActivityID Tests
// ============================================================================

func TestExtractActivityID_ValidPaths_ExtractsID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"basic activity path", "/v1/guilds/guild:123/activities/activity:456", "activity:456"},
		{"simple ID", "/activities/act-xyz", "act-xyz"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractActivityID(tt.path)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractActivityID_InvalidPaths_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
	}{
		{"no activities segment", "/v1/guilds/guild:123"},
		{"activities at end", "/v1/guilds/guild:123/activities"},
		{"activities with trailing slash", "/v1/guilds/guild:123/activities/"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractActivityID(tt.path)
			if result != "" {
				t.Errorf("expected empty string, got %q", result)
			}
		})
	}
}

// ============================================================================
// extractTimerID Tests
// ============================================================================

func TestExtractTimerID_ValidPaths_ExtractsID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"basic timer path", "/v1/guilds/guild:123/people/person:456/timers/timer:789", "timer:789"},
		{"simple ID", "/timers/tmr-abc", "tmr-abc"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractTimerID(tt.path)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractTimerID_SkipsResetAction(t *testing.T) {
	t.Parallel()
	// "reset" after timers should be skipped as action name
	result := extractTimerID("/v1/guilds/guild:123/people/person:456/timers/reset")
	if result != "" {
		t.Errorf("expected empty string for 'reset' action, got %q", result)
	}
}

func TestExtractTimerID_InvalidPaths_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
	}{
		{"no timers segment", "/v1/guilds/guild:123/people/person:456"},
		{"timers at end", "/v1/guilds/guild:123/people/person:456/timers"},
		{"timers with trailing slash", "/v1/guilds/guild:123/people/person:456/timers/"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractTimerID(tt.path)
			if result != "" {
				t.Errorf("expected empty string, got %q", result)
			}
		})
	}
}

// ============================================================================
// Context Helper Tests
// ============================================================================

func TestGetGuildID_Present_ReturnsValue(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), GuildIDKey, "guild:999")

	result := GetGuildID(ctx)

	if result != "guild:999" {
		t.Errorf("expected 'guild:999', got %q", result)
	}
}

func TestGetGuildID_Missing_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	result := GetGuildID(ctx)

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestGetGuildID_WrongType_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), GuildIDKey, 12345) // Wrong type

	result := GetGuildID(ctx)

	if result != "" {
		t.Errorf("expected empty string for wrong type, got %q", result)
	}
}

func TestGetPersonID_Present_ReturnsValue(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), PersonIDKey, "person:888")

	result := GetPersonID(ctx)

	if result != "person:888" {
		t.Errorf("expected 'person:888', got %q", result)
	}
}

func TestGetPersonID_Missing_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	result := GetPersonID(ctx)

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestGetActivityID_Present_ReturnsValue(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), ActivityIDKey, "activity:777")

	result := GetActivityID(ctx)

	if result != "activity:777" {
		t.Errorf("expected 'activity:777', got %q", result)
	}
}

func TestGetActivityID_Missing_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	result := GetActivityID(ctx)

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestGetTimerID_Present_ReturnsValue(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), TimerIDKey, "timer:666")

	result := GetTimerID(ctx)

	if result != "timer:666" {
		t.Errorf("expected 'timer:666', got %q", result)
	}
}

func TestGetTimerID_Missing_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	result := GetTimerID(ctx)

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

// ============================================================================
// ExtractPathParams Tests
// ============================================================================

func TestExtractPathParams_ExtractsAllIDs(t *testing.T) {
	t.Parallel()
	handler := &captureHandler{}
	middleware := ExtractPathParams

	req := httptest.NewRequest(http.MethodGet, "/v1/guilds/guild:123/people/person:456/activities/activity:789/timers/timer:012", nil)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if !handler.called {
		t.Error("handler should have been called")
	}

	// Note: activities are not at this path level typically, but test extraction logic
	if GetGuildID(handler.ctx) != "guild:123" {
		t.Errorf("expected guild ID 'guild:123', got %q", GetGuildID(handler.ctx))
	}
	if GetPersonID(handler.ctx) != "person:456" {
		t.Errorf("expected person ID 'person:456', got %q", GetPersonID(handler.ctx))
	}
	if GetActivityID(handler.ctx) != "activity:789" {
		t.Errorf("expected activity ID 'activity:789', got %q", GetActivityID(handler.ctx))
	}
	if GetTimerID(handler.ctx) != "timer:012" {
		t.Errorf("expected timer ID 'timer:012', got %q", GetTimerID(handler.ctx))
	}
}

func TestExtractPathParams_PartialPath_ExtractsAvailable(t *testing.T) {
	t.Parallel()
	handler := &captureHandler{}
	middleware := ExtractPathParams

	req := httptest.NewRequest(http.MethodGet, "/v1/guilds/guild:abc/people/person:def", nil)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if GetGuildID(handler.ctx) != "guild:abc" {
		t.Errorf("expected guild ID 'guild:abc', got %q", GetGuildID(handler.ctx))
	}
	if GetPersonID(handler.ctx) != "person:def" {
		t.Errorf("expected person ID 'person:def', got %q", GetPersonID(handler.ctx))
	}
	// Activity and Timer should be empty
	if GetActivityID(handler.ctx) != "" {
		t.Errorf("expected empty activity ID, got %q", GetActivityID(handler.ctx))
	}
	if GetTimerID(handler.ctx) != "" {
		t.Errorf("expected empty timer ID, got %q", GetTimerID(handler.ctx))
	}
}

func TestExtractPathParams_NoMatchingSegments_ProceesWithEmptyContext(t *testing.T) {
	t.Parallel()
	handler := &captureHandler{}
	middleware := ExtractPathParams

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if !handler.called {
		t.Error("handler should have been called")
	}
	if GetGuildID(handler.ctx) != "" {
		t.Errorf("expected empty guild ID, got %q", GetGuildID(handler.ctx))
	}
}
