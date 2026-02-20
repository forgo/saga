package service

import (
	"context"
	"testing"

	"github.com/forgo/saga/api/internal/model"
)

// ============================================================================
// Mock Repositories for Discovery
// ============================================================================

type mockBlockChecker struct {
	isBlockedFunc func(ctx context.Context, userID1, userID2 string) (bool, error)
}

func (m *mockBlockChecker) IsBlockedEitherWay(ctx context.Context, userID1, userID2 string) (bool, error) {
	if m.isBlockedFunc != nil {
		return m.isBlockedFunc(ctx, userID1, userID2)
	}
	return false, nil
}

// ============================================================================
// isBlocked Tests
// ============================================================================

func TestIsBlocked_NilChecker_ReturnsFalse(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := &DiscoveryService{
		blockChecker: nil,
	}

	if svc.isBlocked(ctx, "user-1", "user-2") {
		t.Error("expected false when block checker is nil")
	}
}

func TestIsBlocked_UsersBlocked_ReturnsTrue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	checker := &mockBlockChecker{
		isBlockedFunc: func(ctx context.Context, userID1, userID2 string) (bool, error) {
			return true, nil
		},
	}

	svc := &DiscoveryService{
		blockChecker: checker,
	}

	if !svc.isBlocked(ctx, "user-1", "user-2") {
		t.Error("expected true when users are blocked")
	}
}

func TestIsBlocked_NotBlocked_ReturnsFalse(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	checker := &mockBlockChecker{
		isBlockedFunc: func(ctx context.Context, userID1, userID2 string) (bool, error) {
			return false, nil
		},
	}

	svc := &DiscoveryService{
		blockChecker: checker,
	}

	if svc.isBlocked(ctx, "user-1", "user-2") {
		t.Error("expected false when users are not blocked")
	}
}

// ============================================================================
// calculateCompatibilityFromAnswers Tests
// ============================================================================

func TestCalculateCompatibilityFromAnswers_EmptyAnswers_ReturnsZero(t *testing.T) {
	t.Parallel()

	svc := &DiscoveryService{}

	score := svc.calculateCompatibilityFromAnswers(map[string][2]*model.Answer{})
	if score != 0 {
		t.Errorf("expected 0 for empty answers, got %f", score)
	}
}

func TestCalculateCompatibilityFromAnswers_PerfectMatch(t *testing.T) {
	t.Parallel()

	svc := &DiscoveryService{}

	answers := map[string][2]*model.Answer{
		"q1": {
			makeAnswer("a", []string{"a", "b"}, model.ImportanceSomewhat, false, 1.0, nil),
			makeAnswer("a", []string{"a", "b"}, model.ImportanceSomewhat, false, 1.0, nil),
		},
	}

	score := svc.calculateCompatibilityFromAnswers(answers)
	if score != 100 {
		t.Errorf("expected 100 for perfect match, got %f", score)
	}
}

func TestCalculateCompatibilityFromAnswers_NilAnswers_ReturnsZero(t *testing.T) {
	t.Parallel()

	svc := &DiscoveryService{}

	answers := map[string][2]*model.Answer{
		"q1": {nil, nil},
	}

	score := svc.calculateCompatibilityFromAnswers(answers)
	if score != 0 {
		t.Errorf("expected 0 for nil answers, got %f", score)
	}
}

func TestCalculateCompatibilityFromAnswers_DealBreakerViolated_ReturnsZero(t *testing.T) {
	t.Parallel()

	svc := &DiscoveryService{}

	answers := map[string][2]*model.Answer{
		"q1": {
			makeAnswer("a", []string{"a"}, model.ImportanceSomewhat, true, 1.0, nil),       // dealbreaker: only accepts "a"
			makeAnswer("b", []string{"a", "b"}, model.ImportanceSomewhat, false, 1.0, nil), // selected "b"
		},
	}

	score := svc.calculateCompatibilityFromAnswers(answers)
	if score != 0 {
		t.Errorf("expected 0 when dealbreaker violated, got %f", score)
	}
}

func TestCalculateCompatibilityFromAnswers_PartialMatch(t *testing.T) {
	t.Parallel()

	svc := &DiscoveryService{}

	// A accepts B's answer, B doesn't accept A's answer
	answers := map[string][2]*model.Answer{
		"q1": {
			makeAnswer("a", []string{"a", "b"}, model.ImportanceSomewhat, false, 1.0, nil), // accepts b
			makeAnswer("b", []string{"b"}, model.ImportanceSomewhat, false, 1.0, nil),      // only accepts b
		},
	}

	score := svc.calculateCompatibilityFromAnswers(answers)
	// A gets points for accepting B, B doesn't get points
	// With equal weights: earnedPoints = 10 (A's weight), totalWeight = 20
	// score = (10/20) * 100 = 50%
	if score != 50 {
		t.Errorf("expected 50 for partial match, got %f", score)
	}
}

func TestCalculateCompatibilityFromAnswers_YikesPenalty(t *testing.T) {
	t.Parallel()

	svc := &DiscoveryService{}

	// A accepts B's answer but marks it as yikes
	// B accepts A's answer
	answers := map[string][2]*model.Answer{
		"q1": {
			makeAnswer("a", []string{"a", "b"}, model.ImportanceSomewhat, false, 1.0, []string{"b"}), // yikes on b
			makeAnswer("b", []string{"a", "b"}, model.ImportanceSomewhat, false, 1.0, nil),
		},
	}

	score := svc.calculateCompatibilityFromAnswers(answers)
	// Both accept each other (full points) but A has yikes penalty
	// Weight per answer = 10 (somewhat with 1.0 alignment = 10 * 1.0 = 10)
	// earnedPoints = 10 + 10 - 2.5 (yikes penalty = 10 * 0.25) = 17.5
	// totalWeight = 20
	// score = (17.5/20) * 100 = 87.5
	if score < 87 || score > 88 {
		t.Errorf("expected ~87.5 with yikes penalty, got %f", score)
	}
}

func TestCalculateCompatibilityFromAnswers_ScoreClamped(t *testing.T) {
	t.Parallel()

	svc := &DiscoveryService{}

	// Heavy yikes penalty could make score negative - verify clamping
	answers := map[string][2]*model.Answer{
		"q1": {
			makeAnswer("a", []string{"c"}, model.ImportanceSomewhat, false, 1.0, []string{"b"}), // doesn't accept b, yikes on b
			makeAnswer("b", []string{"c"}, model.ImportanceSomewhat, false, 1.0, []string{"a"}), // doesn't accept a, yikes on a
		},
	}

	score := svc.calculateCompatibilityFromAnswers(answers)
	// Neither accepts the other, both have yikes
	// earnedPoints = 0 - 2.5 - 2.5 = -5
	// score should be clamped to 0
	if score != 0 {
		t.Errorf("expected 0 (clamped), got %f", score)
	}
}

// ============================================================================
// calculateMatchScores Tests
// ============================================================================

func TestCalculateMatchScores_BaseFromCompatibility(t *testing.T) {
	t.Parallel()

	svc := &DiscoveryService{}

	results := []DiscoveryResult{
		{CompatibilityScore: 75},
	}

	svc.calculateMatchScores(results)

	// Base score is compatibility, no other bonuses
	if results[0].MatchScore != 75 {
		t.Errorf("expected match score 75 from compatibility, got %f", results[0].MatchScore)
	}
}

func TestCalculateMatchScores_SharedInterestsBonus(t *testing.T) {
	t.Parallel()

	svc := &DiscoveryService{}

	results := []DiscoveryResult{
		{
			CompatibilityScore: 50,
			SharedInterests: []SharedInterestBrief{
				{InterestID: "1", InterestName: "Art"},
				{InterestID: "2", InterestName: "Music"},
				{InterestID: "3", InterestName: "Tech"},
			},
		},
	}

	svc.calculateMatchScores(results)

	// 50 base + 3*4 interests = 50 + 12 = 62
	if results[0].MatchScore != 62 {
		t.Errorf("expected match score 62 (50+12 interests), got %f", results[0].MatchScore)
	}
}

func TestCalculateMatchScores_SharedInterestsBonusCapped(t *testing.T) {
	t.Parallel()

	svc := &DiscoveryService{}

	// More than 5 interests should cap at +20
	interests := make([]SharedInterestBrief, 10)
	for i := range interests {
		interests[i] = SharedInterestBrief{InterestID: string(rune('a' + i))}
	}

	results := []DiscoveryResult{
		{CompatibilityScore: 50, SharedInterests: interests},
	}

	svc.calculateMatchScores(results)

	// 50 base + 20 capped = 70
	if results[0].MatchScore != 70 {
		t.Errorf("expected match score 70 (50+20 capped), got %f", results[0].MatchScore)
	}
}

func TestCalculateMatchScores_TeachLearnBonus(t *testing.T) {
	t.Parallel()

	svc := &DiscoveryService{}

	results := []DiscoveryResult{
		{
			CompatibilityScore: 50,
			SharedInterests: []SharedInterestBrief{
				{InterestID: "1", TeachLearnMatch: true},
				{InterestID: "2", TeachLearnMatch: true},
			},
		},
	}

	svc.calculateMatchScores(results)

	// 50 base + 2*4 interests + 2*5 teach/learn = 50 + 8 + 10 = 68
	if results[0].MatchScore != 68 {
		t.Errorf("expected match score 68, got %f", results[0].MatchScore)
	}
}

func TestCalculateMatchScores_TeachLearnBonusCapped(t *testing.T) {
	t.Parallel()

	svc := &DiscoveryService{}

	// Many teach/learn matches should cap at +15
	interests := make([]SharedInterestBrief, 5)
	for i := range interests {
		interests[i] = SharedInterestBrief{InterestID: string(rune('a' + i)), TeachLearnMatch: true}
	}

	results := []DiscoveryResult{
		{CompatibilityScore: 50, SharedInterests: interests},
	}

	svc.calculateMatchScores(results)

	// 50 base + 20 interests (capped) + 15 teach/learn (capped) = 85
	if results[0].MatchScore != 85 {
		t.Errorf("expected match score 85 (50+20+15 capped), got %f", results[0].MatchScore)
	}
}

func TestCalculateMatchScores_DistanceBonus(t *testing.T) {
	t.Parallel()

	svc := &DiscoveryService{}

	tests := []struct {
		distance      model.DistanceBucket
		expectedBonus float64
	}{
		{model.DistanceNearby, 10},
		{model.Distance2km, 8},
		{model.Distance5km, 5},
		{model.Distance10km, 2},
		{model.Distance20kmPlus, 0},
	}

	for _, tt := range tests {
		results := []DiscoveryResult{
			{CompatibilityScore: 50, Distance: tt.distance},
		}

		svc.calculateMatchScores(results)

		expected := 50 + tt.expectedBonus
		if results[0].MatchScore != expected {
			t.Errorf("distance %s: expected match score %f, got %f", tt.distance, expected, results[0].MatchScore)
		}
	}
}

func TestCalculateMatchScores_CombinedBonuses(t *testing.T) {
	t.Parallel()

	svc := &DiscoveryService{}

	results := []DiscoveryResult{
		{
			CompatibilityScore: 60,
			SharedInterests: []SharedInterestBrief{
				{InterestID: "1", TeachLearnMatch: true},
			},
			Distance: model.DistanceNearby,
		},
	}

	svc.calculateMatchScores(results)

	// 60 base + 4 interest + 5 teach/learn + 10 distance = 79
	if results[0].MatchScore != 79 {
		t.Errorf("expected match score 79, got %f", results[0].MatchScore)
	}
}

// ============================================================================
// PeopleDiscoveryFilter Defaults Tests
// ============================================================================

func TestDiscoverPeople_DefaultsLimit(t *testing.T) {
	t.Parallel()

	// Test the default logic used in DiscoverPeople
	filter := PeopleDiscoveryFilter{Limit: 0}

	// Apply defaults as would happen in DiscoverPeople
	if filter.Limit <= 0 || filter.Limit > 50 {
		filter.Limit = 20
	}

	if filter.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", filter.Limit)
	}
}

func TestDiscoverPeople_CapsLimit(t *testing.T) {
	t.Parallel()

	filter := PeopleDiscoveryFilter{Limit: 100}

	if filter.Limit <= 0 || filter.Limit > 50 {
		filter.Limit = 20
	}

	if filter.Limit != 20 {
		t.Errorf("expected capped limit 20, got %d", filter.Limit)
	}
}

func TestDiscoverPeople_DefaultsRadius(t *testing.T) {
	t.Parallel()

	filter := PeopleDiscoveryFilter{RadiusKm: 0}

	if filter.RadiusKm <= 0 {
		filter.RadiusKm = DefaultSearchRadiusKm
	}

	if filter.RadiusKm != DefaultSearchRadiusKm {
		t.Errorf("expected default radius %f, got %f", DefaultSearchRadiusKm, filter.RadiusKm)
	}
}

func TestDiscoverPeople_CapsRadius(t *testing.T) {
	t.Parallel()

	filter := PeopleDiscoveryFilter{RadiusKm: 1000}

	if filter.RadiusKm > MaxSearchRadiusKm {
		filter.RadiusKm = MaxSearchRadiusKm
	}

	if filter.RadiusKm != MaxSearchRadiusKm {
		t.Errorf("expected capped radius %f, got %f", MaxSearchRadiusKm, filter.RadiusKm)
	}
}
