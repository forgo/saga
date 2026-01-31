package service

import (
	"context"
	"testing"

	"github.com/forgo/saga/api/internal/model"
)

// ============================================================================
// Mock Repositories
// ============================================================================

type mockDeviceTokenRepo struct {
	getByUserIDFunc    func(ctx context.Context, userID string) ([]*model.DeviceToken, error)
	getByTokenFunc     func(ctx context.Context, token string) (*model.DeviceToken, error)
	markInactiveFunc   func(ctx context.Context, token string) error
	updateLastUsedFunc func(ctx context.Context, id string) error
}

func (m *mockDeviceTokenRepo) GetByUserID(ctx context.Context, userID string) ([]*model.DeviceToken, error) {
	if m.getByUserIDFunc != nil {
		return m.getByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockDeviceTokenRepo) GetByToken(ctx context.Context, token string) (*model.DeviceToken, error) {
	if m.getByTokenFunc != nil {
		return m.getByTokenFunc(ctx, token)
	}
	return nil, nil
}

func (m *mockDeviceTokenRepo) MarkInactive(ctx context.Context, token string) error {
	if m.markInactiveFunc != nil {
		return m.markInactiveFunc(ctx, token)
	}
	return nil
}

func (m *mockDeviceTokenRepo) UpdateLastUsed(ctx context.Context, id string) error {
	if m.updateLastUsedFunc != nil {
		return m.updateLastUsedFunc(ctx, id)
	}
	return nil
}

// ============================================================================
// Test Helpers
// ============================================================================

func newTestPushService(deviceRepo *mockDeviceTokenRepo, enabled bool) *PushService {
	if deviceRepo == nil {
		deviceRepo = &mockDeviceTokenRepo{}
	}
	svc, _ := NewPushService(PushServiceConfig{
		DeviceRepo: deviceRepo,
		Enabled:    enabled,
	})
	return svc
}

// ============================================================================
// IsEnabled Tests
// ============================================================================

func TestPushService_IsEnabled_True(t *testing.T) {
	t.Parallel()

	svc := newTestPushService(nil, true)
	if !svc.IsEnabled() {
		t.Error("expected IsEnabled to return true")
	}
}

func TestPushService_IsEnabled_False(t *testing.T) {
	t.Parallel()

	svc := newTestPushService(nil, false)
	if svc.IsEnabled() {
		t.Error("expected IsEnabled to return false")
	}
}

// ============================================================================
// SendToUser Tests
// ============================================================================

func TestPushService_SendToUser_Disabled(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := newTestPushService(nil, false)

	_, err := svc.SendToUser(ctx, "user-1", &PushNotification{
		Title: "Test",
		Body:  "Test message",
	})

	if err != ErrPushDisabled {
		t.Errorf("expected ErrPushDisabled, got %v", err)
	}
}

func TestPushService_SendToUser_NoDevices(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	deviceRepo := &mockDeviceTokenRepo{
		getByUserIDFunc: func(ctx context.Context, userID string) ([]*model.DeviceToken, error) {
			return []*model.DeviceToken{}, nil
		},
	}

	svc := newTestPushService(deviceRepo, true)

	_, err := svc.SendToUser(ctx, "user-1", &PushNotification{
		Title: "Test",
		Body:  "Test message",
	})

	if err != ErrNoDeviceTokens {
		t.Errorf("expected ErrNoDeviceTokens, got %v", err)
	}
}

func TestPushService_SendToUser_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	lastUsedUpdated := false
	deviceRepo := &mockDeviceTokenRepo{
		getByUserIDFunc: func(ctx context.Context, userID string) ([]*model.DeviceToken, error) {
			return []*model.DeviceToken{
				{
					ID:       "device-1",
					UserID:   userID,
					Platform: model.PlatformIOS,
					Token:    "apns-token-123",
					Active:   true,
				},
			}, nil
		},
		updateLastUsedFunc: func(ctx context.Context, id string) error {
			lastUsedUpdated = true
			return nil
		},
	}

	svc := newTestPushService(deviceRepo, true)

	results, err := svc.SendToUser(ctx, "user-1", &PushNotification{
		Title: "Test",
		Body:  "Test message",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Success {
		t.Error("expected success=true")
	}
	if !lastUsedUpdated {
		t.Error("expected last_used to be updated")
	}
}

func TestPushService_SendToUser_MultipleDevices(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	deviceRepo := &mockDeviceTokenRepo{
		getByUserIDFunc: func(ctx context.Context, userID string) ([]*model.DeviceToken, error) {
			return []*model.DeviceToken{
				{ID: "device-1", UserID: userID, Platform: model.PlatformIOS, Token: "token-1"},
				{ID: "device-2", UserID: userID, Platform: model.PlatformAndroid, Token: "token-2"},
				{ID: "device-3", UserID: userID, Platform: model.PlatformWeb, Token: "token-3"},
			}, nil
		},
		updateLastUsedFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}

	svc := newTestPushService(deviceRepo, true)

	results, err := svc.SendToUser(ctx, "user-1", &PushNotification{
		Title: "Test",
		Body:  "Test message",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

// ============================================================================
// SendMulticast Tests
// ============================================================================

func TestPushService_SendMulticast_Disabled(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := newTestPushService(nil, false)

	_, err := svc.SendMulticast(ctx, []string{"user-1", "user-2"}, &PushNotification{
		Title: "Test",
		Body:  "Test message",
	})

	if err != ErrPushDisabled {
		t.Errorf("expected ErrPushDisabled, got %v", err)
	}
}

func TestPushService_SendMulticast_MultipleUsers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	deviceRepo := &mockDeviceTokenRepo{
		getByUserIDFunc: func(ctx context.Context, userID string) ([]*model.DeviceToken, error) {
			if userID == "user-no-devices" {
				return []*model.DeviceToken{}, nil
			}
			return []*model.DeviceToken{
				{ID: "device-" + userID, UserID: userID, Token: "token-" + userID},
			}, nil
		},
		updateLastUsedFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}

	svc := newTestPushService(deviceRepo, true)

	results, err := svc.SendMulticast(ctx, []string{"user-1", "user-2", "user-no-devices"}, &PushNotification{
		Title: "Test",
		Body:  "Test message",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have results for 2 users (user-no-devices skipped)
	if len(results) != 2 {
		t.Fatalf("expected results for 2 users, got %d", len(results))
	}
}

// ============================================================================
// maskToken Tests
// ============================================================================

func TestMaskToken_Short(t *testing.T) {
	t.Parallel()

	result := maskToken("short")
	if result != "***" {
		t.Errorf("expected *** for short token, got %s", result)
	}
}

func TestMaskToken_Long(t *testing.T) {
	t.Parallel()

	result := maskToken("abcdefghijklmnop")
	expected := "abcd...mnop"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestMaskToken_ExactlyEight(t *testing.T) {
	t.Parallel()

	result := maskToken("12345678")
	if result != "***" {
		t.Errorf("expected *** for 8-char token, got %s", result)
	}
}

func TestMaskToken_NineChars(t *testing.T) {
	t.Parallel()

	result := maskToken("123456789")
	expected := "1234...6789"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// ============================================================================
// NilRepo Tests
// ============================================================================

func TestPushService_SendToUser_NilRepo(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Create service with nil repo
	svc := &PushService{
		deviceRepo: nil,
		enabled:    true,
	}

	_, err := svc.SendToUser(ctx, "user-1", &PushNotification{
		Title: "Test",
		Body:  "Test message",
	})

	if err == nil {
		t.Error("expected error for nil repo")
	}
}
