package tests

/*
FEATURE: Resonance Scoring
DOMAIN: Gamification & Reputation

ACCEPTANCE CRITERIA:
===================

AC-RES-001: Questing Points - Event Completion
  GIVEN verified event completion
  WHEN resonance calculated
  THEN questing points awarded (10 base)

AC-RES-002: Questing - Early Confirm Bonus
  GIVEN RSVP confirmed 2+ hours before event
  WHEN resonance calculated
  THEN +2 bonus points awarded

AC-RES-003: Questing - On-Time Bonus
  GIVEN check-in within +-10 min of start
  WHEN resonance calculated
  THEN +2 bonus points awarded

AC-RES-004: Wayfinder Points - Host Event
  GIVEN user hosted verified event with 3 attendees
  WHEN resonance calculated
  THEN wayfinder points awarded (8 + attendee bonus)

AC-RES-005: Attunement Points - Answer Question
  GIVEN user answers questionnaire question
  WHEN resonance calculated
  THEN +2 attunement points

AC-RES-006: Anti-Farming - Unique Source
  GIVEN questing points already awarded for event X
  WHEN system attempts to award again for event X
  THEN second award silently skipped (unique constraint)

AC-RES-007: Daily Cap Enforcement
  GIVEN user at daily questing cap
  WHEN completing another event
  THEN points not awarded (silent skip)

AC-RES-008: Ledger Immutability
  GIVEN existing ledger entry
  WHEN attempting to modify
  THEN modification fails

AC-RES-009: Score Aggregation
  GIVEN ledger entries totaling 100 questing, 50 mana
  WHEN user views resonance score
  THEN total = 150, questing = 100, mana = 50
*/

import (
	"context"
	"testing"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/repository"
	"github.com/forgo/saga/api/internal/service"
	"github.com/forgo/saga/api/internal/testing/fixtures"
	"github.com/forgo/saga/api/internal/testing/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createResonanceService(t *testing.T, tdb *testdb.TestDB) (*service.ResonanceService, *repository.ResonanceRepository) {
	resonanceRepo := repository.NewResonanceRepository(tdb.DB)
	resonanceService := service.NewResonanceService(service.ResonanceServiceConfig{
		Repo: resonanceRepo,
	})
	return resonanceService, resonanceRepo
}

func TestResonance_QuestingPointsBase(t *testing.T) {
	// AC-RES-001: Questing Points - Event Completion
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Award questing points for event completion
	err := resonanceService.AwardQuesting(ctx, user.ID, "event123", false, false)
	require.NoError(t, err)

	// Check score
	score, err := resonanceService.GetUserScore(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PointsQuestingBase, score.Questing) // 10 points
	assert.Equal(t, model.PointsQuestingBase, score.Total)
}

func TestResonance_QuestingEarlyConfirmBonus(t *testing.T) {
	// AC-RES-002: Questing - Early Confirm Bonus
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Award with early confirm bonus
	err := resonanceService.AwardQuesting(ctx, user.ID, "event456", true, false)
	require.NoError(t, err)

	score, err := resonanceService.GetUserScore(ctx, user.ID)
	require.NoError(t, err)
	expectedPoints := model.PointsQuestingBase + model.PointsQuestingEarlyConfirm // 10 + 2 = 12
	assert.Equal(t, expectedPoints, score.Questing)
}

func TestResonance_QuestingOnTimeBonus(t *testing.T) {
	// AC-RES-003: Questing - On-Time Bonus
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Award with on-time checkin bonus
	err := resonanceService.AwardQuesting(ctx, user.ID, "event789", false, true)
	require.NoError(t, err)

	score, err := resonanceService.GetUserScore(ctx, user.ID)
	require.NoError(t, err)
	expectedPoints := model.PointsQuestingBase + model.PointsQuestingCheckin // 10 + 2 = 12
	assert.Equal(t, expectedPoints, score.Questing)
}

func TestResonance_QuestingAllBonuses(t *testing.T) {
	// Test all bonuses combined
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Award with both bonuses
	err := resonanceService.AwardQuesting(ctx, user.ID, "eventFull", true, true)
	require.NoError(t, err)

	score, err := resonanceService.GetUserScore(ctx, user.ID)
	require.NoError(t, err)
	// 10 base + 2 early + 2 checkin = 14
	expectedPoints := model.PointsQuestingBase + model.PointsQuestingEarlyConfirm + model.PointsQuestingCheckin
	assert.Equal(t, expectedPoints, score.Questing)
}

func TestResonance_WayfinderPoints(t *testing.T) {
	// AC-RES-004: Wayfinder Points - Host Event
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	host := f.CreateUser(t)

	// Award wayfinder points for hosting with 3 verified attendees
	err := resonanceService.AwardWayfinder(ctx, host.ID, "eventHosted1", 3, false)
	require.NoError(t, err)

	score, err := resonanceService.GetUserScore(ctx, host.ID)
	require.NoError(t, err)
	// 8 base + (2 * 3 attendees) = 14
	expectedPoints := model.PointsWayfinderBase + (model.PointsWayfinderPerAttendee * 3)
	assert.Equal(t, expectedPoints, score.Wayfinder)
}

func TestResonance_WayfinderAttendeeCap(t *testing.T) {
	// Attendee bonus is capped at 4 to prevent mega-event farming
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	host := f.CreateUser(t)

	// Award with 10 attendees (should be capped at 4)
	err := resonanceService.AwardWayfinder(ctx, host.ID, "eventMega", 10, false)
	require.NoError(t, err)

	score, err := resonanceService.GetUserScore(ctx, host.ID)
	require.NoError(t, err)
	// 8 base + (2 * 4 capped) = 16
	expectedPoints := model.PointsWayfinderBase + (model.PointsWayfinderPerAttendee * model.PointsWayfinderMaxAttendees)
	assert.Equal(t, expectedPoints, score.Wayfinder)
}

func TestResonance_WayfinderEarlyConfirm(t *testing.T) {
	// Wayfinder early confirm bonus
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	host := f.CreateUser(t)

	// Award with early confirm
	err := resonanceService.AwardWayfinder(ctx, host.ID, "eventEarly", 2, true)
	require.NoError(t, err)

	score, err := resonanceService.GetUserScore(ctx, host.ID)
	require.NoError(t, err)
	// 8 base + (2 * 2 attendees) + 2 early = 14
	expectedPoints := model.PointsWayfinderBase + (model.PointsWayfinderPerAttendee * 2) + model.PointsWayfinderEarlyConfirm
	assert.Equal(t, expectedPoints, score.Wayfinder)
}

func TestResonance_AttunementQuestionAnswer(t *testing.T) {
	// AC-RES-005: Attunement Points - Answer Question
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Award attunement for answering a question
	err := resonanceService.AwardAttunement(ctx, user.ID, "question1")
	require.NoError(t, err)

	score, err := resonanceService.GetUserScore(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PointsAttunementQuestion, score.Attunement) // 2 points
}

func TestResonance_AttunementMultipleQuestions(t *testing.T) {
	// Multiple questions should each award points
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Answer 3 different questions
	require.NoError(t, resonanceService.AwardAttunement(ctx, user.ID, "q1"))
	require.NoError(t, resonanceService.AwardAttunement(ctx, user.ID, "q2"))
	require.NoError(t, resonanceService.AwardAttunement(ctx, user.ID, "q3"))

	// Recalculate to ensure cached score is current
	score, err := resonanceService.RecalculateScore(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PointsAttunementQuestion*3, score.Attunement) // 6 points
}

func TestResonance_UniqueSourceIdempotency(t *testing.T) {
	// AC-RES-006: Anti-Farming - Unique Source
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// First award
	err := resonanceService.AwardQuesting(ctx, user.ID, "eventDupe", false, false)
	require.NoError(t, err)

	// Second award for same event should be silently skipped
	err = resonanceService.AwardQuesting(ctx, user.ID, "eventDupe", false, false)
	require.NoError(t, err) // No error, just skipped

	// Should only have one event's worth of points
	score, err := resonanceService.GetUserScore(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PointsQuestingBase, score.Questing)
}

func TestResonance_AttunementQuestionIdempotency(t *testing.T) {
	// Same question should only award once
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Answer same question twice
	_ = resonanceService.AwardAttunement(ctx, user.ID, "sameQuestion")
	_ = resonanceService.AwardAttunement(ctx, user.ID, "sameQuestion")

	score, err := resonanceService.GetUserScore(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PointsAttunementQuestion, score.Attunement) // Only 2, not 4
}

func TestResonance_DailyCapEnforcement(t *testing.T) {
	// AC-RES-007: Daily Cap Enforcement
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Award enough to hit the daily cap (40 for questing)
	// Each event gives 10 base, so 4 events = 40 points
	for i := 0; i < 4; i++ {
		err := resonanceService.AwardQuesting(ctx, user.ID, "eventCap"+string(rune('a'+i)), false, false)
		require.NoError(t, err)
	}

	// This should be at the cap now
	score, _ := resonanceService.GetUserScore(ctx, user.ID)
	assert.Equal(t, model.DailyCapQuesting, score.Questing) // 40

	// Try to award one more - should be silently capped
	err := resonanceService.AwardQuesting(ctx, user.ID, "eventOver", false, false)
	require.NoError(t, err) // No error, just capped

	// Score should still be at cap
	score, _ = resonanceService.GetUserScore(ctx, user.ID)
	assert.Equal(t, model.DailyCapQuesting, score.Questing)
}

func TestResonance_DailyCapPartialAward(t *testing.T) {
	// When close to cap, partial points should be awarded
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Award 3 events = 30 points, leaving 10 remaining
	for i := 0; i < 3; i++ {
		_ = resonanceService.AwardQuesting(ctx, user.ID, "eventPartial"+string(rune('a'+i)), false, false)
	}

	// Award with bonuses (would be 14, but only 10 remaining)
	_ = resonanceService.AwardQuesting(ctx, user.ID, "eventPartialFinal", true, true)

	score, _ := resonanceService.GetUserScore(ctx, user.ID)
	assert.Equal(t, model.DailyCapQuesting, score.Questing) // Should cap at 40
}

func TestResonance_ScoreAggregation(t *testing.T) {
	// AC-RES-009: Score Aggregation
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Award various stats
	_ = resonanceService.AwardQuesting(ctx, user.ID, "ev1", false, false) // 10 questing
	_ = resonanceService.AwardQuesting(ctx, user.ID, "ev2", false, false) // 10 questing
	_ = resonanceService.AwardWayfinder(ctx, user.ID, "host1", 2, false)  // 8 + 4 = 12 wayfinder
	_ = resonanceService.AwardAttunement(ctx, user.ID, "q1")              // 2 attunement
	_ = resonanceService.AwardAttunement(ctx, user.ID, "q2")              // 2 attunement

	score, err := resonanceService.GetUserScore(ctx, user.ID)
	require.NoError(t, err)

	assert.Equal(t, 20, score.Questing)
	assert.Equal(t, 12, score.Wayfinder)
	assert.Equal(t, 4, score.Attunement)
	assert.Equal(t, 0, score.Mana)
	assert.Equal(t, 0, score.Nexus)
	assert.Equal(t, 36, score.Total) // 20 + 12 + 4
}

func TestResonance_GetLedger(t *testing.T) {
	// Test that ledger entries are retrievable
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Create some entries
	_ = resonanceService.AwardQuesting(ctx, user.ID, "ledgerEv1", false, false)
	_ = resonanceService.AwardQuesting(ctx, user.ID, "ledgerEv2", true, false)
	_ = resonanceService.AwardAttunement(ctx, user.ID, "ledgerQ1")

	// Get ledger
	ledger, err := resonanceService.GetUserLedger(ctx, user.ID, 50, 0)
	require.NoError(t, err)
	assert.Len(t, ledger, 3)

	// Verify entries have proper data
	for _, entry := range ledger {
		assert.NotEmpty(t, entry.ID)
		assert.Equal(t, user.ID, entry.UserID)
		assert.NotEmpty(t, entry.Stat)
		assert.Greater(t, entry.Points, 0)
		assert.NotEmpty(t, entry.SourceObjectID)
		assert.NotEmpty(t, entry.ReasonCode)
	}
}

func TestResonance_ManaPoints(t *testing.T) {
	// Test Mana points for helpful support sessions
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	helper := f.CreateUser(t)
	receiver := f.CreateUser(t)

	// Award mana for helpful session
	err := resonanceService.AwardMana(ctx, helper.ID, receiver.ID, "hangout1", "YES", false, false)
	require.NoError(t, err)

	score, err := resonanceService.GetUserScore(ctx, helper.ID)
	require.NoError(t, err)
	assert.Equal(t, model.PointsManaBase, score.Mana) // 12 points
}

func TestResonance_ManaNotHelpfulNoPoints(t *testing.T) {
	// Not helpful rating should not award points
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	helper := f.CreateUser(t)
	receiver := f.CreateUser(t)

	// Award mana with NOT_REALLY rating - should not award
	err := resonanceService.AwardMana(ctx, helper.ID, receiver.ID, "hangout2", "NOT_REALLY", false, false)
	require.NoError(t, err)

	score, err := resonanceService.GetUserScore(ctx, helper.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, score.Mana)
}

func TestResonance_ManaDiminishingReturns(t *testing.T) {
	// Repeated sessions between same pair should have diminishing returns
	// Note: Daily cap of 32 limits total mana per day
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, resonanceRepo := createResonanceService(t, tdb)
	ctx := context.Background()

	helper := f.CreateUser(t)
	receiver := f.CreateUser(t)

	// Session 1 (count=0, full value)
	_ = resonanceService.AwardMana(ctx, helper.ID, receiver.ID, "session1", "YES", false, false)
	count1, _ := resonanceRepo.GetSupportPairCount(ctx, helper.ID, receiver.ID)
	assert.Equal(t, 1, count1)

	score1, _ := resonanceService.RecalculateScore(ctx, helper.ID)
	assert.Equal(t, model.PointsManaBase, score1.Mana) // 12 points

	// Session 2 (count=1, full value)
	_ = resonanceService.AwardMana(ctx, helper.ID, receiver.ID, "session2", "YES", false, false)
	count2, _ := resonanceRepo.GetSupportPairCount(ctx, helper.ID, receiver.ID)
	assert.Equal(t, 2, count2)

	score2, _ := resonanceService.RecalculateScore(ctx, helper.ID)
	assert.Equal(t, model.PointsManaBase*2, score2.Mana) // 24 points

	// Session 3 (count=2, full value but capped)
	// Would award 12, but only 8 remaining to daily cap of 32
	_ = resonanceService.AwardMana(ctx, helper.ID, receiver.ID, "session3", "YES", false, false)
	count3, _ := resonanceRepo.GetSupportPairCount(ctx, helper.ID, receiver.ID)
	assert.Equal(t, 3, count3)

	score3, _ := resonanceService.RecalculateScore(ctx, helper.ID)
	assert.Equal(t, model.DailyCapMana, score3.Mana) // 32 points (capped)

	// Session 4 (count=3, would be 0.5x but already at cap)
	_ = resonanceService.AwardMana(ctx, helper.ID, receiver.ID, "session4", "YES", false, false)

	// Pair count should NOT increment when no points awarded
	count4, _ := resonanceRepo.GetSupportPairCount(ctx, helper.ID, receiver.ID)
	assert.Equal(t, 3, count4) // Still 3, not 4

	score4, _ := resonanceService.RecalculateScore(ctx, helper.ID)
	assert.Equal(t, model.DailyCapMana, score4.Mana) // Still 32 (at cap)
}

func TestResonance_ProfileRefreshMonthly(t *testing.T) {
	// Profile refresh is once per month
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// First refresh
	err := resonanceService.AwardMonthlyProfileRefresh(ctx, user.ID)
	require.NoError(t, err)

	score, _ := resonanceService.GetUserScore(ctx, user.ID)
	assert.Equal(t, model.PointsAttunementProfileRefresh, score.Attunement) // 10 points

	// Second refresh same month - should be idempotent
	err = resonanceService.AwardMonthlyProfileRefresh(ctx, user.ID)
	require.NoError(t, err)

	score, _ = resonanceService.GetUserScore(ctx, user.ID)
	assert.Equal(t, model.PointsAttunementProfileRefresh, score.Attunement) // Still 10
}

func TestResonance_NexusPoints(t *testing.T) {
	// Test Nexus point awarding
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Award nexus points with circle contributions
	contributions := []model.CircleNexusContribution{
		{CircleID: "circle1", CircleName: "Circle A", Points: 20, ActivityFactor: 1.0, ActiveMembers: 5},
		{CircleID: "circle2", CircleName: "Circle B", Points: 15, ActivityFactor: 0.5, ActiveMembers: 3},
	}

	err := resonanceService.AwardNexus(ctx, user.ID, contributions)
	require.NoError(t, err)

	score, _ := resonanceService.GetUserScore(ctx, user.ID)
	assert.Equal(t, 35, score.Nexus) // 20 + 15
}

func TestResonance_NexusMonthlyCap(t *testing.T) {
	// Nexus is capped at 200 per month
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Award more than cap
	contributions := []model.CircleNexusContribution{
		{CircleID: "bigCircle", CircleName: "Big Circle", Points: 250, ActivityFactor: 1.0, ActiveMembers: 10},
	}

	err := resonanceService.AwardNexus(ctx, user.ID, contributions)
	require.NoError(t, err)

	score, _ := resonanceService.GetUserScore(ctx, user.ID)
	assert.Equal(t, model.MonthlyCapNexus, score.Nexus) // Capped at 200
}

func TestResonance_RecalculateScore(t *testing.T) {
	// Test score recalculation
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	resonanceService, _ := createResonanceService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Award some points
	_ = resonanceService.AwardQuesting(ctx, user.ID, "recalcEv", false, false)
	_ = resonanceService.AwardAttunement(ctx, user.ID, "recalcQ")

	// Recalculate
	newScore, err := resonanceService.RecalculateScore(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, 10, newScore.Questing)
	assert.Equal(t, 2, newScore.Attunement)
	assert.Equal(t, 12, newScore.Total)
}
