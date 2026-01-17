// Package tests contains end-to-end acceptance tests for the Saga API.
package tests

import (
	"context"
	"testing"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/repository"
	"github.com/forgo/saga/api/internal/testing/fixtures"
	"github.com/forgo/saga/api/internal/testing/testdb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
FEATURE: Device Tokens
DOMAIN: Notifications

ACCEPTANCE CRITERIA:
===================

AC-DEV-001: Register Device Token
  GIVEN authenticated user
  WHEN user registers device with platform and token
  THEN device token is created and persisted

AC-DEV-002: List User Devices
  GIVEN user with registered devices
  WHEN user lists devices
  THEN all active devices returned

AC-DEV-003: Delete Device
  GIVEN user with registered device
  WHEN user deletes device
  THEN device is removed

AC-DEV-004: Max Devices Limit
  GIVEN user with 10 devices
  WHEN user tries to register another
  THEN should be limited (10 max per user)

AC-DEV-005: Duplicate Token Upsert
  GIVEN device token already registered
  WHEN same token registered again
  THEN existing record is updated (not duplicated)

AC-DEV-006: Mark Token Inactive
  GIVEN active device token
  WHEN token is marked inactive
  THEN token no longer appears in active list
*/

func TestDevice_RegisterToken(t *testing.T) {
	// AC-DEV-001: Register Device Token
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	deviceRepo := repository.NewDeviceTokenRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)

	token := &model.DeviceToken{
		UserID:   user.ID,
		Platform: model.PlatformIOS,
		Token:    "test-apns-token-12345",
		Name:     "iPhone 15",
	}

	err := deviceRepo.Create(ctx, token)
	require.NoError(t, err)
	assert.NotEmpty(t, token.ID)
	assert.True(t, token.Active)

	// Verify token can be retrieved
	retrieved, err := deviceRepo.GetByToken(ctx, "test-apns-token-12345")
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, user.ID, retrieved.UserID)
	assert.Equal(t, model.PlatformIOS, retrieved.Platform)
	assert.Equal(t, "iPhone 15", retrieved.Name)
}

func TestDevice_ListUserDevices(t *testing.T) {
	// AC-DEV-002: List User Devices
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	deviceRepo := repository.NewDeviceTokenRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Register multiple devices
	for i := 0; i < 3; i++ {
		token := &model.DeviceToken{
			UserID:   user.ID,
			Platform: model.PlatformIOS,
			Token:    "token-" + string(rune('a'+i)),
			Name:     "Device " + string(rune('A'+i)),
		}
		err := deviceRepo.Create(ctx, token)
		require.NoError(t, err)
	}

	// List devices
	devices, err := deviceRepo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, devices, 3)
}

func TestDevice_DeleteDevice(t *testing.T) {
	// AC-DEV-003: Delete Device
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	deviceRepo := repository.NewDeviceTokenRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)

	token := &model.DeviceToken{
		UserID:   user.ID,
		Platform: model.PlatformAndroid,
		Token:    "fcm-token-xyz",
	}
	err := deviceRepo.Create(ctx, token)
	require.NoError(t, err)

	// Delete the device
	err = deviceRepo.Delete(ctx, token.ID)
	require.NoError(t, err)

	// Verify it's gone
	retrieved, err := deviceRepo.GetByID(ctx, token.ID)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestDevice_CountDevices(t *testing.T) {
	// AC-DEV-004: Max Devices Limit (test count function)
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	deviceRepo := repository.NewDeviceTokenRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Initially zero
	count, err := deviceRepo.CountByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Add devices
	for i := 0; i < 5; i++ {
		token := &model.DeviceToken{
			UserID:   user.ID,
			Platform: model.PlatformIOS,
			Token:    "token-" + string(rune('0'+i)),
		}
		err := deviceRepo.Create(ctx, token)
		require.NoError(t, err)
	}

	count, err = deviceRepo.CountByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestDevice_DuplicateTokenUpsert(t *testing.T) {
	// AC-DEV-005: Duplicate Token Upsert
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	deviceRepo := repository.NewDeviceTokenRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	tokenValue := "unique-token-abc123"

	// First registration
	token1 := &model.DeviceToken{
		UserID:   user.ID,
		Platform: model.PlatformIOS,
		Token:    tokenValue,
		Name:     "Old Name",
	}
	err := deviceRepo.UpsertByToken(ctx, token1)
	require.NoError(t, err)
	originalID := token1.ID

	// Second registration with same token but different name
	token2 := &model.DeviceToken{
		UserID:   user.ID,
		Platform: model.PlatformIOS,
		Token:    tokenValue,
		Name:     "New Name",
	}
	err = deviceRepo.UpsertByToken(ctx, token2)
	require.NoError(t, err)

	// Should have same ID (updated, not duplicated)
	assert.Equal(t, originalID, token2.ID)

	// Verify only one device exists
	devices, err := deviceRepo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, devices, 1)
}

func TestDevice_MarkTokenInactive(t *testing.T) {
	// AC-DEV-006: Mark Token Inactive
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	deviceRepo := repository.NewDeviceTokenRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)

	token := &model.DeviceToken{
		UserID:   user.ID,
		Platform: model.PlatformIOS,
		Token:    "inactive-token-test",
	}
	err := deviceRepo.Create(ctx, token)
	require.NoError(t, err)

	// Verify initially in active list
	devices, err := deviceRepo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, devices, 1)

	// Mark inactive
	err = deviceRepo.MarkInactive(ctx, "inactive-token-test")
	require.NoError(t, err)

	// Should not appear in active list
	devices, err = deviceRepo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, devices, 0)
}

func TestDevice_UpdateLastUsed(t *testing.T) {
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	deviceRepo := repository.NewDeviceTokenRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)

	token := &model.DeviceToken{
		UserID:   user.ID,
		Platform: model.PlatformIOS,
		Token:    "last-used-test-token",
	}
	err := deviceRepo.Create(ctx, token)
	require.NoError(t, err)

	// Initially no last_used
	retrieved, err := deviceRepo.GetByID(ctx, token.ID)
	require.NoError(t, err)
	assert.Nil(t, retrieved.LastUsed)

	// Update last used
	err = deviceRepo.UpdateLastUsed(ctx, token.ID)
	require.NoError(t, err)

	// Should now have last_used set
	retrieved, err = deviceRepo.GetByID(ctx, token.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved.LastUsed)
}

func TestDevice_DifferentPlatforms(t *testing.T) {
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	deviceRepo := repository.NewDeviceTokenRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Test all platforms
	platforms := []model.DevicePlatform{
		model.PlatformIOS,
		model.PlatformAndroid,
		model.PlatformWeb,
	}

	for i, platform := range platforms {
		token := &model.DeviceToken{
			UserID:   user.ID,
			Platform: platform,
			Token:    "platform-token-" + string(rune('a'+i)),
		}
		err := deviceRepo.Create(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, platform, token.Platform)
	}

	devices, err := deviceRepo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, devices, 3)
}

func TestDevice_DeleteByToken(t *testing.T) {
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	deviceRepo := repository.NewDeviceTokenRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	tokenValue := "delete-by-token-test"

	token := &model.DeviceToken{
		UserID:   user.ID,
		Platform: model.PlatformIOS,
		Token:    tokenValue,
	}
	err := deviceRepo.Create(ctx, token)
	require.NoError(t, err)

	// Delete by token value (not ID)
	err = deviceRepo.DeleteByToken(ctx, tokenValue)
	require.NoError(t, err)

	// Verify it's gone
	retrieved, err := deviceRepo.GetByToken(ctx, tokenValue)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}
