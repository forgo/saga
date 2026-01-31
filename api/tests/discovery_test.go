package tests

/*
FEATURE: Discovery
DOMAIN: People Matching & Privacy

ACCEPTANCE CRITERIA:
===================

AC-DISC-001: Discovery Eligibility
  GIVEN user answered 3+ questions from all 4 required categories
  THEN user.discovery_eligible = true

AC-DISC-002: Discovery Ineligible
  GIVEN user answered only 2 categories
  THEN user.discovery_eligible = false
  AND user cannot appear in discovery

AC-DISC-003: Discover People
  GIVEN eligible user
  WHEN user discovers people
  THEN matching profiles returned
  AND blocked users excluded

AC-DISC-004: Daily Discovery Limit
  GIVEN user has viewed 10 people today
  WHEN user requests more
  THEN fails with 429 Too Many Requests

AC-DISC-005: Distance Bucket Privacy
  GIVEN two users 3km apart
  WHEN one discovers the other
  THEN distance shown as "~5km" (bucket)
  AND exact coordinates never exposed

AC-DISC-006: Blocked Users Hidden
  GIVEN user A blocked user B
  WHEN user A discovers people
  THEN user B not returned
  WHEN user B discovers people
  THEN user A not returned
*/

import (
	"context"
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/repository"
	"github.com/forgo/saga/api/internal/service"
	"github.com/forgo/saga/api/internal/testing/fixtures"
	"github.com/forgo/saga/api/internal/testing/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const DailyDiscoveryLimit = 10 // Default daily limit for testing

func TestDiscovery_DailyLimit_Tracking(t *testing.T) {
	// AC-DISC-004: Daily Discovery Limit tracking
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	discoveryRepo := repository.NewDiscoveryRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	today := time.Now().Format("2006-01-02")

	// Initially no count
	count, err := discoveryRepo.GetDailyCount(ctx, user.ID, today)
	require.NoError(t, err)
	assert.Equal(t, 0, count.PeopleShown)

	// Increment count multiple times
	for i := 0; i < 5; i++ {
		err := discoveryRepo.IncrementCount(ctx, user.ID, today, "people")
		require.NoError(t, err)
	}

	// Verify count increased
	count, err = discoveryRepo.GetDailyCount(ctx, user.ID, today)
	require.NoError(t, err)
	assert.Equal(t, 5, count.PeopleShown)
}

func TestDiscovery_DailyLimit_AtLimit(t *testing.T) {
	// AC-DISC-004: Daily Discovery Limit - at limit
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	discoveryRepo := repository.NewDiscoveryRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	today := time.Now().Format("2006-01-02")

	// Simulate reaching the daily limit
	for i := 0; i < DailyDiscoveryLimit; i++ {
		err := discoveryRepo.IncrementCount(ctx, user.ID, today, "people")
		require.NoError(t, err)
	}

	// Verify at limit
	count, err := discoveryRepo.GetDailyCount(ctx, user.ID, today)
	require.NoError(t, err)
	assert.Equal(t, DailyDiscoveryLimit, count.PeopleShown)

	// Check if limit exceeded
	assert.True(t, count.PeopleShown >= DailyDiscoveryLimit, "Should be at daily limit")
}

func TestDiscovery_DailyLimit_DifferentDays(t *testing.T) {
	// Daily limits should reset each day
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	discoveryRepo := repository.NewDiscoveryRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Increment yesterday
	for i := 0; i < 5; i++ {
		err := discoveryRepo.IncrementCount(ctx, user.ID, yesterday, "people")
		require.NoError(t, err)
	}

	// Increment today
	for i := 0; i < 3; i++ {
		err := discoveryRepo.IncrementCount(ctx, user.ID, today, "people")
		require.NoError(t, err)
	}

	// Verify counts are separate
	yesterdayCount, err := discoveryRepo.GetDailyCount(ctx, user.ID, yesterday)
	require.NoError(t, err)
	assert.Equal(t, 5, yesterdayCount.PeopleShown)

	todayCount, err := discoveryRepo.GetDailyCount(ctx, user.ID, today)
	require.NoError(t, err)
	assert.Equal(t, 3, todayCount.PeopleShown)
}

func TestDiscovery_DailyLimit_DifferentCountTypes(t *testing.T) {
	// Track people, events, and guilds separately
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	discoveryRepo := repository.NewDiscoveryRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	today := time.Now().Format("2006-01-02")

	// Increment different types
	for i := 0; i < 5; i++ {
		err := discoveryRepo.IncrementCount(ctx, user.ID, today, "people")
		require.NoError(t, err)
	}
	for i := 0; i < 3; i++ {
		err := discoveryRepo.IncrementCount(ctx, user.ID, today, "events")
		require.NoError(t, err)
	}
	for i := 0; i < 2; i++ {
		err := discoveryRepo.IncrementCount(ctx, user.ID, today, "guilds")
		require.NoError(t, err)
	}

	// Verify counts are separate
	count, err := discoveryRepo.GetDailyCount(ctx, user.ID, today)
	require.NoError(t, err)
	assert.Equal(t, 5, count.PeopleShown)
	assert.Equal(t, 3, count.EventsShown)
	assert.Equal(t, 2, count.GuildsShown)
}

func TestDiscovery_DistanceBucket_Nearby(t *testing.T) {
	// AC-DISC-005: Distance Bucket Privacy - nearby
	bucket := model.GetDistanceBucket(0.5)
	assert.Equal(t, model.DistanceNearby, bucket, "0.5km should be 'nearby'")
}

func TestDiscovery_DistanceBucket_2km(t *testing.T) {
	// AC-DISC-005: Distance Bucket Privacy - ~2km
	bucket := model.GetDistanceBucket(1.5)
	assert.Equal(t, model.Distance2km, bucket, "1.5km should be '~2km'")
}

func TestDiscovery_DistanceBucket_5km(t *testing.T) {
	// AC-DISC-005: Distance Bucket Privacy - ~5km
	bucket := model.GetDistanceBucket(3.0)
	assert.Equal(t, model.Distance5km, bucket, "3.0km should be '~5km'")
}

func TestDiscovery_DistanceBucket_10km(t *testing.T) {
	// AC-DISC-005: Distance Bucket Privacy - ~10km
	bucket := model.GetDistanceBucket(7.0)
	assert.Equal(t, model.Distance10km, bucket, "7.0km should be '~10km'")
}

func TestDiscovery_DistanceBucket_20kmPlus(t *testing.T) {
	// AC-DISC-005: Distance Bucket Privacy - >20km
	bucket := model.GetDistanceBucket(30.0)
	assert.Equal(t, model.Distance20kmPlus, bucket, "30.0km should be '>20km'")
}

func TestDiscovery_DistanceBucket_GeoService(t *testing.T) {
	// AC-DISC-005: GeoService integration
	geoService := service.NewGeoService()

	// San Francisco coordinates
	lat1, lng1 := 37.7749, -122.4194
	// Oakland coordinates (~13km away)
	lat2, lng2 := 37.8044, -122.2712

	// Calculate distance
	distance := geoService.HaversineDistance(lat1, lng1, lat2, lng2)
	assert.InDelta(t, 13.0, distance, 2.0, "SF to Oakland should be approximately 13km")

	// Get bucket
	bucket := geoService.GetDistanceBucket(distance)
	assert.Equal(t, model.Distance20kmPlus, bucket, "13km should be '>20km' bucket (since >10km)")
}

func TestDiscovery_BlockedUsersHidden(t *testing.T) {
	// AC-DISC-006: Blocked Users Hidden
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	moderationRepo := repository.NewModerationRepository(tdb.DB)
	ctx := context.Background()

	userA := f.CreateUser(t)
	userB := f.CreateUser(t)

	// Initially not blocked
	blocked, err := moderationRepo.IsBlockedEitherWay(ctx, userA.ID, userB.ID)
	require.NoError(t, err)
	assert.False(t, blocked, "Users should not be blocked initially")

	// A blocks B
	err = moderationRepo.CreateBlock(ctx, &model.Block{
		BlockerUserID: userA.ID,
		BlockedUserID: userB.ID,
	})
	require.NoError(t, err)

	// Check block in both directions
	blockedAtoB, err := moderationRepo.IsBlockedEitherWay(ctx, userA.ID, userB.ID)
	require.NoError(t, err)
	assert.True(t, blockedAtoB, "A->B should show as blocked either way")

	blockedBtoA, err := moderationRepo.IsBlockedEitherWay(ctx, userB.ID, userA.ID)
	require.NoError(t, err)
	assert.True(t, blockedBtoA, "B->A should also show as blocked either way")
}

func TestDiscovery_BlockBidirectional(t *testing.T) {
	// AC-DISC-006: Block is bidirectional in effect
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	moderationRepo := repository.NewModerationRepository(tdb.DB)
	ctx := context.Background()

	userA := f.CreateUser(t)
	userB := f.CreateUser(t)
	userC := f.CreateUser(t)

	// A blocks B
	err := moderationRepo.CreateBlock(ctx, &model.Block{
		BlockerUserID: userA.ID,
		BlockedUserID: userB.ID,
	})
	require.NoError(t, err)

	// C is not involved
	blockedAC, err := moderationRepo.IsBlockedEitherWay(ctx, userA.ID, userC.ID)
	require.NoError(t, err)
	assert.False(t, blockedAC, "A and C should not be blocked")

	blockedBC, err := moderationRepo.IsBlockedEitherWay(ctx, userB.ID, userC.ID)
	require.NoError(t, err)
	assert.False(t, blockedBC, "B and C should not be blocked")

	// Verify block only affects A-B relationship
	blockedAB, err := moderationRepo.IsBlockedEitherWay(ctx, userA.ID, userB.ID)
	require.NoError(t, err)
	assert.True(t, blockedAB, "A and B should be blocked")
}

func TestDiscovery_CannotSelfBlock(t *testing.T) {
	// AC-DISC-006: Cannot self-block
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	moderationRepo := repository.NewModerationRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Try to block self - should fail (database trigger enforces this)
	err := moderationRepo.CreateBlock(ctx, &model.Block{
		BlockerUserID: user.ID,
		BlockedUserID: user.ID,
	})
	assert.Error(t, err, "Should not be able to block self")
}

func TestDiscovery_IsBlocked_OneWay(t *testing.T) {
	// Test IsBlocked (one-directional check)
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	moderationRepo := repository.NewModerationRepository(tdb.DB)
	ctx := context.Background()

	userA := f.CreateUser(t)
	userB := f.CreateUser(t)

	// A blocks B
	err := moderationRepo.CreateBlock(ctx, &model.Block{
		BlockerUserID: userA.ID,
		BlockedUserID: userB.ID,
	})
	require.NoError(t, err)

	// Check one-way block
	blockedAB, err := moderationRepo.IsBlocked(ctx, userA.ID, userB.ID)
	require.NoError(t, err)
	assert.True(t, blockedAB, "A->B should show as blocked")

	// B did not block A (one-way check)
	blockedBA, err := moderationRepo.IsBlocked(ctx, userB.ID, userA.ID)
	require.NoError(t, err)
	assert.False(t, blockedBA, "B->A should NOT show as blocked (one-way)")
}

func TestDiscovery_DeleteBlock(t *testing.T) {
	// Unblock functionality
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	moderationRepo := repository.NewModerationRepository(tdb.DB)
	ctx := context.Background()

	userA := f.CreateUser(t)
	userB := f.CreateUser(t)

	// Create block
	err := moderationRepo.CreateBlock(ctx, &model.Block{
		BlockerUserID: userA.ID,
		BlockedUserID: userB.ID,
	})
	require.NoError(t, err)

	// Verify blocked
	blocked, err := moderationRepo.IsBlockedEitherWay(ctx, userA.ID, userB.ID)
	require.NoError(t, err)
	assert.True(t, blocked)

	// Delete block
	err = moderationRepo.DeleteBlock(ctx, userA.ID, userB.ID)
	require.NoError(t, err)

	// Verify unblocked
	blocked, err = moderationRepo.IsBlockedEitherWay(ctx, userA.ID, userB.ID)
	require.NoError(t, err)
	assert.False(t, blocked, "Block should be removed")
}

func TestDiscovery_ProfileEligibility_Update(t *testing.T) {
	// AC-DISC-001/002: Profile eligibility can be updated via Update method
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	profileRepo := repository.NewProfileRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Create profile
	profile := &model.UserProfile{
		UserID:            user.ID,
		Visibility:        model.VisibilityPublic,
		DiscoveryEligible: false,
	}
	err := profileRepo.Create(ctx, profile)
	require.NoError(t, err)

	// Initially not eligible
	fetched, err := profileRepo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.False(t, fetched.DiscoveryEligible)

	// Update to eligible
	updated, err := profileRepo.Update(ctx, user.ID, map[string]interface{}{
		"discovery_eligible": true,
	})
	require.NoError(t, err)
	assert.True(t, updated.DiscoveryEligible, "Should be discovery eligible after update")
}
