package tests

/*
FEATURE: Matching Pools (Donut-style)
DOMAIN: Community Building & Connection

ACCEPTANCE CRITERIA:
===================

AC-POOL-001: Create Pool
  GIVEN guild admin
  WHEN creating pool with frequency=weekly
  THEN pool created

AC-POOL-002: Join Pool
  GIVEN guild member
  WHEN joining pool
  THEN pool_member created

AC-POOL-003: Exclusion List
  GIVEN pool member
  WHEN adding exclusions
  THEN up to 20 allowed

AC-POOL-004: Exclusion Limit
  GIVEN pool member with 20 exclusions
  WHEN adding 21st exclusion
  THEN fails (limit enforced)

AC-POOL-005: Match Generation
  GIVEN pool with 10 members
  WHEN match job runs
  THEN 5 pairs created respecting exclusions
*/

import (
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestMatchingPool_ExclusionLimit(t *testing.T) {
	// AC-POOL-004: Exclusion Limit (max 20)
	// This test verifies the model constant is set correctly
	assert.Equal(t, 20, model.MaxExclusionsPerMember,
		"Max exclusions should be 20 per member")
}

func TestMatchingPool_FrequencyOptions(t *testing.T) {
	// Verify frequency options
	assert.Equal(t, "weekly", model.PoolFrequencyWeekly)
	assert.Equal(t, "biweekly", model.PoolFrequencyBiweekly)
	assert.Equal(t, "monthly", model.PoolFrequencyMonthly)
}

func TestMatchingPool_NextMatchDateWeekly(t *testing.T) {
	// Weekly pool should schedule next match 7 days out
	now := time.Now()
	next := model.GetNextMatchDate(model.PoolFrequencyWeekly, now)
	assert.Equal(t, 7, int(next.Sub(now).Hours()/24))
}

func TestMatchingPool_NextMatchDateBiweekly(t *testing.T) {
	// Biweekly pool should schedule next match 14 days out
	now := time.Now()
	next := model.GetNextMatchDate(model.PoolFrequencyBiweekly, now)
	assert.Equal(t, 14, int(next.Sub(now).Hours()/24))
}

func TestMatchingPool_NextMatchDateMonthly(t *testing.T) {
	// Monthly pool should schedule next match 1 month out
	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	next := model.GetNextMatchDate(model.PoolFrequencyMonthly, now)
	expected := time.Date(2026, 2, 15, 12, 0, 0, 0, time.UTC)
	assert.Equal(t, expected.Format(time.RFC3339), next.Format(time.RFC3339))
}

func TestMatchingPool_GetMatchRound(t *testing.T) {
	// Match round format: YYYY-Www
	t1 := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC) // Week 3 of 2026
	round := model.GetMatchRound(t1)
	assert.Equal(t, "2026-W03", round)
}

func TestMatchingPool_MatchStatuses(t *testing.T) {
	// Verify match status constants
	assert.Equal(t, "pending", model.MatchStatusPending)
	assert.Equal(t, "scheduled", model.MatchStatusScheduled)
	assert.Equal(t, "completed", model.MatchStatusCompleted)
	assert.Equal(t, "skipped", model.MatchStatusSkipped)
}

func TestMatchingPool_Constraints(t *testing.T) {
	// Verify pool constraints
	assert.Equal(t, 10, model.MaxPoolsPerGuild)
	assert.Equal(t, 100, model.MaxMembersPerPool)
	assert.Equal(t, 2, model.MinMatchSize)
	assert.Equal(t, 6, model.MaxMatchSize)
	assert.Equal(t, 100, model.MaxPoolNameLength)
	assert.Equal(t, 500, model.MaxPoolDescLength)
}

func TestMatchingPool_MatchConfig(t *testing.T) {
	// Verify default matching config
	config := model.DefaultMatchingConfig
	assert.Equal(t, 0.6, config.VarietyWeight)
	assert.Equal(t, 0.4, config.CompatibilityWeight)
	assert.Equal(t, 30, config.RecencyDays)
}

func TestMatchingPool_PoolModelDefaults(t *testing.T) {
	// Verify pool model has expected default behavior
	pool := &model.MatchingPool{
		GuildID:   "guild:test",
		Name:      "Test Pool",
		Frequency: model.PoolFrequencyWeekly,
		MatchSize: 2,
	}

	assert.Equal(t, "Test Pool", pool.Name)
	assert.Equal(t, model.PoolFrequencyWeekly, pool.Frequency)
	assert.Equal(t, 2, pool.MatchSize)
	assert.Nil(t, pool.Description)
	assert.Nil(t, pool.ActivitySuggestion)
}
