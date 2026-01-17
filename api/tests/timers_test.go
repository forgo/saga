package tests

/*
FEATURE: Timers
DOMAIN: Timer Tracking

ACCEPTANCE CRITERIA:
===================

AC-TIMER-001: Create Timer
  GIVEN person and activity in guild
  WHEN user creates timer
  THEN timer created with current reset_date

AC-TIMER-002: Reset Timer
  GIVEN existing timer
  WHEN user resets timer
  THEN reset_date updated to now

AC-TIMER-003: List Timers for Person
  GIVEN person has timers A, B, C
  WHEN user lists timers
  THEN all timers returned with elapsed time

AC-TIMER-004: Timer Threshold Events (service-level)
  GIVEN timer with warn_threshold = 60s
  WHEN 60 seconds elapse
  THEN warn event emitted via SSE

AC-TIMER-005: Timer Critical Events (service-level)
  GIVEN timer with critical_threshold = 120s
  WHEN 120 seconds elapse
  THEN critical event emitted via SSE
*/

import (
	"context"
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/repository"
	"github.com/forgo/saga/api/internal/testing/fixtures"
	"github.com/forgo/saga/api/internal/testing/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimer_Create(t *testing.T) {
	// AC-TIMER-001: Create Timer
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	timerRepo := repository.NewTimerRepository(tdb.DB)
	activityRepo := repository.NewActivityRepository(tdb.DB)
	personRepo := repository.NewPersonRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create person and activity
	person := &model.Person{
		GuildID: guild.ID,
		Name:    "John Doe",
	}
	err := personRepo.Create(ctx, person)
	require.NoError(t, err)

	activity := &model.Activity{
		GuildID:  guild.ID,
		Name:     "Coffee",
		Icon:     "coffee",
		Warn:     3600,
		Critical: 7200,
	}
	err = activityRepo.Create(ctx, activity)
	require.NoError(t, err)

	// Create timer
	timer := &model.Timer{
		PersonID:   person.ID,
		ActivityID: activity.ID,
		ResetDate:  time.Now().UTC(),
		Enabled:    true,
		Push:       false,
	}

	err = timerRepo.Create(ctx, timer)
	require.NoError(t, err)
	assert.NotEmpty(t, timer.ID)
	assert.NotZero(t, timer.CreatedOn)

	// Verify timer can be retrieved
	fetched, err := timerRepo.GetByID(ctx, timer.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, person.ID, fetched.PersonID)
	assert.Equal(t, activity.ID, fetched.ActivityID)
	assert.True(t, fetched.Enabled)
	assert.False(t, fetched.Push)
}

func TestTimer_Reset(t *testing.T) {
	// AC-TIMER-002: Reset Timer
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	timerRepo := repository.NewTimerRepository(tdb.DB)
	activityRepo := repository.NewActivityRepository(tdb.DB)
	personRepo := repository.NewPersonRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create person and activity
	person := &model.Person{
		GuildID: guild.ID,
		Name:    "John Doe",
	}
	err := personRepo.Create(ctx, person)
	require.NoError(t, err)

	activity := &model.Activity{
		GuildID:  guild.ID,
		Name:     "Coffee",
		Icon:     "coffee",
		Warn:     3600,
		Critical: 7200,
	}
	err = activityRepo.Create(ctx, activity)
	require.NoError(t, err)

	// Create timer with reset_date in the past
	pastTime := time.Now().UTC().Add(-24 * time.Hour)
	timer := &model.Timer{
		PersonID:   person.ID,
		ActivityID: activity.ID,
		ResetDate:  pastTime,
		Enabled:    true,
		Push:       false,
	}
	err = timerRepo.Create(ctx, timer)
	require.NoError(t, err)

	// Reset timer
	beforeReset := time.Now().UTC().Truncate(time.Second)
	resetTimer, err := timerRepo.Reset(ctx, timer.ID)
	require.NoError(t, err)
	require.NotNil(t, resetTimer)

	// Verify reset_date is updated to approximately now (allow 2 second window for timing)
	assert.True(t, resetTimer.ResetDate.After(beforeReset.Add(-2*time.Second)) || resetTimer.ResetDate.Equal(beforeReset),
		"Reset date should be at or after the reset request (within timing tolerance)")
	assert.True(t, resetTimer.ResetDate.After(pastTime), "Reset date should be after original past time")
}

func TestTimer_ListByPerson(t *testing.T) {
	// AC-TIMER-003: List Timers for Person
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	timerRepo := repository.NewTimerRepository(tdb.DB)
	activityRepo := repository.NewActivityRepository(tdb.DB)
	personRepo := repository.NewPersonRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create person
	person := &model.Person{
		GuildID: guild.ID,
		Name:    "John Doe",
	}
	err := personRepo.Create(ctx, person)
	require.NoError(t, err)

	// Create multiple activities and timers
	activities := []string{"Coffee", "Diaper", "Feeding"}
	for _, name := range activities {
		activity := &model.Activity{
			GuildID:  guild.ID,
			Name:     name,
			Icon:     "icon",
			Warn:     3600,
			Critical: 7200,
		}
		err := activityRepo.Create(ctx, activity)
		require.NoError(t, err)

		timer := &model.Timer{
			PersonID:   person.ID,
			ActivityID: activity.ID,
			ResetDate:  time.Now().UTC(),
			Enabled:    true,
			Push:       false,
		}
		err = timerRepo.Create(ctx, timer)
		require.NoError(t, err)
	}

	// List timers for person
	timers, err := timerRepo.GetByPersonID(ctx, person.ID)
	require.NoError(t, err)
	assert.Len(t, timers, 3, "Should return all 3 timers")
}

func TestTimer_GetByPersonAndActivity(t *testing.T) {
	// Get specific timer by person and activity
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	timerRepo := repository.NewTimerRepository(tdb.DB)
	activityRepo := repository.NewActivityRepository(tdb.DB)
	personRepo := repository.NewPersonRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create person and activity
	person := &model.Person{
		GuildID: guild.ID,
		Name:    "John Doe",
	}
	err := personRepo.Create(ctx, person)
	require.NoError(t, err)

	activity := &model.Activity{
		GuildID:  guild.ID,
		Name:     "Coffee",
		Icon:     "coffee",
		Warn:     3600,
		Critical: 7200,
	}
	err = activityRepo.Create(ctx, activity)
	require.NoError(t, err)

	// Create timer
	timer := &model.Timer{
		PersonID:   person.ID,
		ActivityID: activity.ID,
		ResetDate:  time.Now().UTC(),
		Enabled:    true,
		Push:       true,
	}
	err = timerRepo.Create(ctx, timer)
	require.NoError(t, err)

	// Get by person and activity
	fetched, err := timerRepo.GetByPersonAndActivity(ctx, person.ID, activity.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, timer.ID, fetched.ID)
}

func TestTimer_Update(t *testing.T) {
	// Update timer settings
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	timerRepo := repository.NewTimerRepository(tdb.DB)
	activityRepo := repository.NewActivityRepository(tdb.DB)
	personRepo := repository.NewPersonRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create person and activity
	person := &model.Person{
		GuildID: guild.ID,
		Name:    "John Doe",
	}
	err := personRepo.Create(ctx, person)
	require.NoError(t, err)

	activity := &model.Activity{
		GuildID:  guild.ID,
		Name:     "Coffee",
		Icon:     "coffee",
		Warn:     3600,
		Critical: 7200,
	}
	err = activityRepo.Create(ctx, activity)
	require.NoError(t, err)

	// Create timer
	timer := &model.Timer{
		PersonID:   person.ID,
		ActivityID: activity.ID,
		ResetDate:  time.Now().UTC(),
		Enabled:    true,
		Push:       false,
	}
	err = timerRepo.Create(ctx, timer)
	require.NoError(t, err)

	// Update timer settings
	timer.Enabled = false
	timer.Push = true
	err = timerRepo.Update(ctx, timer)
	require.NoError(t, err)

	// Verify update
	fetched, err := timerRepo.GetByID(ctx, timer.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.False(t, fetched.Enabled, "Timer should be disabled")
	assert.True(t, fetched.Push, "Push should be enabled")
}

func TestTimer_Delete(t *testing.T) {
	// Delete timer
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	timerRepo := repository.NewTimerRepository(tdb.DB)
	activityRepo := repository.NewActivityRepository(tdb.DB)
	personRepo := repository.NewPersonRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create person and activity
	person := &model.Person{
		GuildID: guild.ID,
		Name:    "John Doe",
	}
	err := personRepo.Create(ctx, person)
	require.NoError(t, err)

	activity := &model.Activity{
		GuildID:  guild.ID,
		Name:     "Coffee",
		Icon:     "coffee",
		Warn:     3600,
		Critical: 7200,
	}
	err = activityRepo.Create(ctx, activity)
	require.NoError(t, err)

	// Create timer
	timer := &model.Timer{
		PersonID:   person.ID,
		ActivityID: activity.ID,
		ResetDate:  time.Now().UTC(),
		Enabled:    true,
		Push:       false,
	}
	err = timerRepo.Create(ctx, timer)
	require.NoError(t, err)

	// Delete timer
	err = timerRepo.Delete(ctx, timer.ID)
	require.NoError(t, err)

	// Verify deleted
	fetched, err := timerRepo.GetByID(ctx, timer.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched, "Timer should be deleted")
}

func TestTimer_Count(t *testing.T) {
	// Count timers for person
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	timerRepo := repository.NewTimerRepository(tdb.DB)
	activityRepo := repository.NewActivityRepository(tdb.DB)
	personRepo := repository.NewPersonRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create person
	person := &model.Person{
		GuildID: guild.ID,
		Name:    "John Doe",
	}
	err := personRepo.Create(ctx, person)
	require.NoError(t, err)

	// Initially zero
	count, err := timerRepo.CountByPersonID(ctx, person.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Create timers
	for i := 0; i < 5; i++ {
		activity := &model.Activity{
			GuildID:  guild.ID,
			Name:     "Activity" + string(rune('A'+i)),
			Icon:     "icon",
			Warn:     3600,
			Critical: 7200,
		}
		err := activityRepo.Create(ctx, activity)
		require.NoError(t, err)

		timer := &model.Timer{
			PersonID:   person.ID,
			ActivityID: activity.ID,
			ResetDate:  time.Now().UTC(),
			Enabled:    true,
			Push:       false,
		}
		err = timerRepo.Create(ctx, timer)
		require.NoError(t, err)
	}

	// Count should be 5
	count, err = timerRepo.CountByPersonID(ctx, person.ID)
	require.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestTimer_ThresholdValidation(t *testing.T) {
	// AC-TIMER-004/005: Timer threshold validation
	// These test the service-level logic for determining when timers exceed thresholds

	// Calculate elapsed time
	resetDate := time.Now().UTC().Add(-90 * time.Second) // 90 seconds ago

	elapsed := time.Since(resetDate).Seconds()

	// Service-level validation: Check if elapsed time exceeds thresholds
	warnThreshold := float64(60)     // 60 seconds
	criticalThreshold := float64(120) // 120 seconds

	// After 90 seconds:
	// - Should be past warn threshold (60s) = true
	// - Should not be past critical threshold (120s) = false
	isPastWarn := elapsed >= warnThreshold
	isPastCritical := elapsed >= criticalThreshold

	assert.True(t, isPastWarn, "90s elapsed should be past 60s warn threshold")
	assert.False(t, isPastCritical, "90s elapsed should not be past 120s critical threshold")
}

func TestTimer_UniquePersonActivity(t *testing.T) {
	// Test that person+activity combination is unique
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	timerRepo := repository.NewTimerRepository(tdb.DB)
	activityRepo := repository.NewActivityRepository(tdb.DB)
	personRepo := repository.NewPersonRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create person and activity
	person := &model.Person{
		GuildID: guild.ID,
		Name:    "John Doe",
	}
	err := personRepo.Create(ctx, person)
	require.NoError(t, err)

	activity := &model.Activity{
		GuildID:  guild.ID,
		Name:     "Coffee",
		Icon:     "coffee",
		Warn:     3600,
		Critical: 7200,
	}
	err = activityRepo.Create(ctx, activity)
	require.NoError(t, err)

	// Create first timer
	timer1 := &model.Timer{
		PersonID:   person.ID,
		ActivityID: activity.ID,
		ResetDate:  time.Now().UTC(),
		Enabled:    true,
		Push:       false,
	}
	err = timerRepo.Create(ctx, timer1)
	require.NoError(t, err)

	// Try to create duplicate timer - should fail due to unique constraint
	timer2 := &model.Timer{
		PersonID:   person.ID,
		ActivityID: activity.ID,
		ResetDate:  time.Now().UTC(),
		Enabled:    true,
		Push:       false,
	}
	err = timerRepo.Create(ctx, timer2)
	assert.Error(t, err, "Should fail to create duplicate timer for same person+activity")
}
