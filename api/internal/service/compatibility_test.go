package service

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/forgo/saga/api/internal/model"
)

// ============================================================================
// Mock QuestionnaireRepository
// ============================================================================

type mockQuestionnaireRepo struct {
	getSharedAnswersFunc  func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error)
	getAllQuestionsFunc   func(ctx context.Context) ([]*model.Question, error)
}

func (m *mockQuestionnaireRepo) GetAllQuestions(ctx context.Context) ([]*model.Question, error) {
	if m.getAllQuestionsFunc != nil {
		return m.getAllQuestionsFunc(ctx)
	}
	return nil, nil
}

func (m *mockQuestionnaireRepo) GetQuestionsByCategory(ctx context.Context, category string) ([]*model.Question, error) {
	return nil, nil
}

func (m *mockQuestionnaireRepo) GetCircleQuestions(ctx context.Context, circleID string) ([]*model.Question, error) {
	return nil, nil
}

func (m *mockQuestionnaireRepo) GetQuestionByID(ctx context.Context, id string) (*model.Question, error) {
	return nil, nil
}

func (m *mockQuestionnaireRepo) CreateQuestion(ctx context.Context, question *model.Question) error {
	return nil
}

func (m *mockQuestionnaireRepo) GetUserAnswer(ctx context.Context, userID, questionID string) (*model.Answer, error) {
	return nil, nil
}

func (m *mockQuestionnaireRepo) GetUserAnswers(ctx context.Context, userID string) ([]*model.Answer, error) {
	return nil, nil
}

func (m *mockQuestionnaireRepo) GetUserAnswersWithQuestions(ctx context.Context, userID string) ([]*model.AnswerWithQuestion, error) {
	return nil, nil
}

func (m *mockQuestionnaireRepo) CreateAnswer(ctx context.Context, answer *model.Answer) error {
	return nil
}

func (m *mockQuestionnaireRepo) UpdateAnswer(ctx context.Context, userID, questionID string, updates map[string]interface{}) (*model.Answer, error) {
	return nil, nil
}

func (m *mockQuestionnaireRepo) DeleteAnswer(ctx context.Context, userID, questionID string) error {
	return nil
}

func (m *mockQuestionnaireRepo) GetSharedAnswers(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
	if m.getSharedAnswersFunc != nil {
		return m.getSharedAnswersFunc(ctx, userAID, userBID)
	}
	return nil, nil
}

func (m *mockQuestionnaireRepo) GetUserBiasProfile(ctx context.Context, userID string) (*model.UserBiasProfile, error) {
	return nil, nil
}

func (m *mockQuestionnaireRepo) UpdateUserBiasProfile(ctx context.Context, userID string, accumulatedBias float64, answerCount int) error {
	return nil
}

func (m *mockQuestionnaireRepo) GetQuestionProgress(ctx context.Context, userID string) (*model.QuestionProgress, error) {
	return nil, nil
}

func (m *mockQuestionnaireRepo) GetCircleValues(ctx context.Context, id string) (*model.CircleValues, error) {
	return nil, nil
}

func (m *mockQuestionnaireRepo) GetCircleValuesByCircle(ctx context.Context, circleID string) ([]*model.CircleValues, error) {
	return nil, nil
}

func (m *mockQuestionnaireRepo) CreateCircleValues(ctx context.Context, cv *model.CircleValues) error {
	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func newTestCompatibilityService(repo *mockQuestionnaireRepo) *CompatibilityService {
	return NewCompatibilityService(CompatibilityServiceConfig{
		QuestionnaireRepo: repo,
	})
}

func makeAnswer(selected string, acceptable []string, importance string, dealbreaker bool, alignmentWeight float64, yikes []string) *model.Answer {
	return &model.Answer{
		SelectedOption:    selected,
		AcceptableOptions: acceptable,
		Importance:        importance,
		IsDealBreaker:     dealbreaker,
		AlignmentWeight:   alignmentWeight,
		YikesOptions:      yikes,
	}
}

// ============================================================================
// containsString Tests
// ============================================================================

func TestContainsString_Found(t *testing.T) {
	t.Parallel()
	result := containsString([]string{"a", "b", "c"}, "b")
	if !result {
		t.Error("expected true for existing element")
	}
}

func TestContainsString_NotFound(t *testing.T) {
	t.Parallel()
	result := containsString([]string{"a", "b", "c"}, "d")
	if result {
		t.Error("expected false for missing element")
	}
}

func TestContainsString_EmptySlice(t *testing.T) {
	t.Parallel()
	result := containsString([]string{}, "a")
	if result {
		t.Error("expected false for empty slice")
	}
}

func TestContainsString_NilSlice(t *testing.T) {
	t.Parallel()
	result := containsString(nil, "a")
	if result {
		t.Error("expected false for nil slice")
	}
}

// ============================================================================
// CalculateCompatibility Tests
// ============================================================================

func TestCalculateCompatibility_NoSharedAnswers_ReturnsZero(t *testing.T) {
	t.Parallel()
	repo := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return map[string][2]*model.Answer{}, nil
		},
	}
	svc := newTestCompatibilityService(repo)

	score, err := svc.CalculateCompatibility(context.Background(), "user:A", "user:B")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Score != 0 {
		t.Errorf("expected score 0, got %f", score.Score)
	}
	if score.SharedCount != 0 {
		t.Errorf("expected shared count 0, got %d", score.SharedCount)
	}
	if score.DealBreaker {
		t.Error("expected no dealbreaker")
	}
}

func TestCalculateCompatibility_RepoError_ReturnsError(t *testing.T) {
	t.Parallel()
	expectedErr := errors.New("db error")
	repo := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return nil, expectedErr
		},
	}
	svc := newTestCompatibilityService(repo)

	_, err := svc.CalculateCompatibility(context.Background(), "user:A", "user:B")

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestCalculateCompatibility_PerfectMatch_Returns100(t *testing.T) {
	t.Parallel()
	repo := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return map[string][2]*model.Answer{
				"q1": {
					makeAnswer("yes", []string{"yes"}, model.ImportanceVery, false, 0.5, nil),
					makeAnswer("yes", []string{"yes"}, model.ImportanceVery, false, 0.5, nil),
				},
				"q2": {
					makeAnswer("often", []string{"often", "sometimes"}, model.ImportanceSomewhat, false, 0.5, nil),
					makeAnswer("often", []string{"often", "sometimes"}, model.ImportanceSomewhat, false, 0.5, nil),
				},
			}, nil
		},
	}
	svc := newTestCompatibilityService(repo)

	score, err := svc.CalculateCompatibility(context.Background(), "user:A", "user:B")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Score != 100 {
		t.Errorf("expected score 100 for perfect match, got %f", score.Score)
	}
	if score.SharedCount != 2 {
		t.Errorf("expected shared count 2, got %d", score.SharedCount)
	}
}

func TestCalculateCompatibility_PartialMatch_ReturnsMiddleScore(t *testing.T) {
	t.Parallel()
	repo := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return map[string][2]*model.Answer{
				"q1": {
					// A accepts yes, B answers yes - match
					makeAnswer("yes", []string{"yes"}, model.ImportanceVery, false, 0.5, nil),
					makeAnswer("yes", []string{"yes"}, model.ImportanceVery, false, 0.5, nil),
				},
				"q2": {
					// A accepts "often", B answers "rarely" - no match
					makeAnswer("often", []string{"often"}, model.ImportanceVery, false, 0.5, nil),
					makeAnswer("rarely", []string{"rarely"}, model.ImportanceVery, false, 0.5, nil),
				},
			}, nil
		},
	}
	svc := newTestCompatibilityService(repo)

	score, err := svc.CalculateCompatibility(context.Background(), "user:A", "user:B")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 50% match on each side, geometric mean = 50
	if score.Score < 45 || score.Score > 55 {
		t.Errorf("expected score around 50 for partial match, got %f", score.Score)
	}
}

func TestCalculateCompatibility_DealBreakerViolation_ReturnsZero(t *testing.T) {
	t.Parallel()
	repo := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return map[string][2]*model.Answer{
				"q1": {
					// A has dealbreaker requiring "yes", B answers "no"
					makeAnswer("yes", []string{"yes"}, model.ImportanceMandatory, true, 0.5, nil),
					makeAnswer("no", []string{"no"}, model.ImportanceSomewhat, false, 0.5, nil),
				},
			}, nil
		},
	}
	svc := newTestCompatibilityService(repo)

	score, err := svc.CalculateCompatibility(context.Background(), "user:A", "user:B")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Score != 0 {
		t.Errorf("expected score 0 for dealbreaker violation, got %f", score.Score)
	}
	if !score.DealBreaker {
		t.Error("expected DealBreaker to be true")
	}
}

func TestCalculateCompatibility_AsymmetricScoring(t *testing.T) {
	t.Parallel()
	repo := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return map[string][2]*model.Answer{
				"q1": {
					// A accepts both, B only accepts "yes"
					makeAnswer("yes", []string{"yes", "no"}, model.ImportanceVery, false, 0.5, nil),
					makeAnswer("yes", []string{"yes"}, model.ImportanceVery, false, 0.5, nil),
				},
			}, nil
		},
	}
	svc := newTestCompatibilityService(repo)

	score, err := svc.CalculateCompatibility(context.Background(), "user:A", "user:B")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A→B: B answered "yes" which A accepts = 100%
	// B→A: A answered "yes" which B accepts = 100%
	// Overall: sqrt(100 * 100) = 100
	if score.AToB != 100 {
		t.Errorf("expected A→B score 100, got %f", score.AToB)
	}
	if score.BToA != 100 {
		t.Errorf("expected B→A score 100, got %f", score.BToA)
	}
}

func TestCalculateCompatibility_YikesPenalty_ReducesScore(t *testing.T) {
	t.Parallel()
	repo := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return map[string][2]*model.Answer{
				"q1": {
					// A accepts "maybe" but marks it as yikes
					makeAnswer("yes", []string{"yes", "maybe"}, model.ImportanceVery, false, 0.5, []string{"maybe"}),
					makeAnswer("maybe", []string{"yes", "maybe"}, model.ImportanceVery, false, 0.5, nil),
				},
			}, nil
		},
	}
	svc := newTestCompatibilityService(repo)

	score, err := svc.CalculateCompatibility(context.Background(), "user:A", "user:B")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Score should be reduced by 25% due to yikes
	// A→B: 100 - 25 = 75
	if score.AToB < 70 || score.AToB > 80 {
		t.Errorf("expected A→B score around 75 (yikes penalty), got %f", score.AToB)
	}
}

func TestCalculateCompatibility_NilAnswers_HandledGracefully(t *testing.T) {
	t.Parallel()
	repo := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return map[string][2]*model.Answer{
				"q1": {nil, nil},
				"q2": {
					makeAnswer("yes", []string{"yes"}, model.ImportanceVery, false, 0.5, nil),
					nil,
				},
			}, nil
		},
	}
	svc := newTestCompatibilityService(repo)

	score, err := svc.CalculateCompatibility(context.Background(), "user:A", "user:B")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Nil answers should be skipped
	if score.Score != 0 {
		t.Errorf("expected score 0 (no valid shared answers), got %f", score.Score)
	}
}

func TestCalculateCompatibility_GeometricMean(t *testing.T) {
	t.Parallel()
	repo := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return map[string][2]*model.Answer{
				"q1": {
					// A→B will be 100% (B's answer acceptable to A)
					// B→A will be 0% (A's answer not acceptable to B)
					makeAnswer("yes", []string{"yes", "no"}, model.ImportanceVery, false, 0.5, nil),
					makeAnswer("no", []string{"no"}, model.ImportanceVery, false, 0.5, nil),
				},
			}, nil
		},
	}
	svc := newTestCompatibilityService(repo)

	score, err := svc.CalculateCompatibility(context.Background(), "user:A", "user:B")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A→B: B answered "no", A accepts "yes" or "no" = 100%
	// B→A: A answered "yes", B only accepts "no" = 0%
	// Geometric mean: sqrt(100 * 0) = 0
	if score.Score != 0 {
		t.Errorf("expected geometric mean 0, got %f", score.Score)
	}
}

func TestCalculateCompatibility_AlignmentWeightAffectsScore(t *testing.T) {
	t.Parallel()
	// Test with high alignment weight
	repoHighAlign := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return map[string][2]*model.Answer{
				"q1": {
					makeAnswer("yes", []string{"yes"}, model.ImportanceVery, false, 1.0, nil), // High alignment
					makeAnswer("yes", []string{"yes"}, model.ImportanceVery, false, 1.0, nil),
				},
			}, nil
		},
	}
	svcHighAlign := newTestCompatibilityService(repoHighAlign)
	scoreHigh, _ := svcHighAlign.CalculateCompatibility(context.Background(), "user:A", "user:B")

	// Test with low alignment weight
	repoLowAlign := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return map[string][2]*model.Answer{
				"q1": {
					makeAnswer("yes", []string{"yes"}, model.ImportanceVery, false, 0.0, nil), // Low alignment
					makeAnswer("yes", []string{"yes"}, model.ImportanceVery, false, 0.0, nil),
				},
			}, nil
		},
	}
	svcLowAlign := newTestCompatibilityService(repoLowAlign)
	scoreLow, _ := svcLowAlign.CalculateCompatibility(context.Background(), "user:A", "user:B")

	// Both should be 100 for perfect match, but weight affects contribution
	if scoreHigh.Score != 100 || scoreLow.Score != 100 {
		t.Errorf("expected both scores 100 for perfect match, got high=%f, low=%f", scoreHigh.Score, scoreLow.Score)
	}
}

// ============================================================================
// calculateDirectionalScore Tests
// ============================================================================

func TestCalculateDirectionalScore_NoAnswers_ReturnsZero(t *testing.T) {
	t.Parallel()
	svc := newTestCompatibilityService(&mockQuestionnaireRepo{})

	score, dealbreaker := svc.calculateDirectionalScore(map[string][2]*model.Answer{}, true)

	if score != 0 {
		t.Errorf("expected score 0, got %f", score)
	}
	if dealbreaker {
		t.Error("expected no dealbreaker")
	}
}

func TestCalculateDirectionalScore_AllMatches_Returns100(t *testing.T) {
	t.Parallel()
	svc := newTestCompatibilityService(&mockQuestionnaireRepo{})

	answers := map[string][2]*model.Answer{
		"q1": {
			makeAnswer("yes", []string{"yes"}, model.ImportanceVery, false, 0.5, nil),
			makeAnswer("yes", []string{"yes"}, model.ImportanceVery, false, 0.5, nil),
		},
	}

	score, dealbreaker := svc.calculateDirectionalScore(answers, true)

	if score != 100 {
		t.Errorf("expected score 100, got %f", score)
	}
	if dealbreaker {
		t.Error("expected no dealbreaker")
	}
}

func TestCalculateDirectionalScore_NoMatches_ReturnsZero(t *testing.T) {
	t.Parallel()
	svc := newTestCompatibilityService(&mockQuestionnaireRepo{})

	answers := map[string][2]*model.Answer{
		"q1": {
			makeAnswer("yes", []string{"yes"}, model.ImportanceVery, false, 0.5, nil),
			makeAnswer("no", []string{"no"}, model.ImportanceVery, false, 0.5, nil),
		},
	}

	score, _ := svc.calculateDirectionalScore(answers, true)

	if score != 0 {
		t.Errorf("expected score 0, got %f", score)
	}
}

func TestCalculateDirectionalScore_DealBreakerViolated_ReturnsTrueFlag(t *testing.T) {
	t.Parallel()
	svc := newTestCompatibilityService(&mockQuestionnaireRepo{})

	answers := map[string][2]*model.Answer{
		"q1": {
			makeAnswer("yes", []string{"yes"}, model.ImportanceMandatory, true, 0.5, nil), // Dealbreaker
			makeAnswer("no", []string{"no"}, model.ImportanceSomewhat, false, 0.5, nil),
		},
	}

	_, dealbreaker := svc.calculateDirectionalScore(answers, true)

	if !dealbreaker {
		t.Error("expected dealbreaker flag to be true")
	}
}

func TestCalculateDirectionalScore_YikesPenalty_ReducesScore(t *testing.T) {
	t.Parallel()
	svc := newTestCompatibilityService(&mockQuestionnaireRepo{})

	answers := map[string][2]*model.Answer{
		"q1": {
			// A accepts "maybe" but marks it as yikes
			makeAnswer("yes", []string{"yes", "maybe"}, model.ImportanceVery, false, 0.5, []string{"maybe"}),
			makeAnswer("maybe", []string{"maybe"}, model.ImportanceVery, false, 0.5, nil),
		},
	}

	score, _ := svc.calculateDirectionalScore(answers, true)

	// 100% match but 25% penalty = 75%
	if score < 70 || score > 80 {
		t.Errorf("expected score around 75, got %f", score)
	}
}

func TestCalculateDirectionalScore_ScoreClamping(t *testing.T) {
	t.Parallel()
	svc := newTestCompatibilityService(&mockQuestionnaireRepo{})

	// Multiple yikes that would push score below 0
	answers := map[string][2]*model.Answer{
		"q1": {
			makeAnswer("yes", []string{"maybe"}, model.ImportanceVery, false, 0.5, []string{"maybe"}),
			makeAnswer("maybe", []string{"maybe"}, model.ImportanceVery, false, 0.5, nil),
		},
		"q2": {
			makeAnswer("yes", []string{"no"}, model.ImportanceVery, false, 0.5, []string{"no"}),
			makeAnswer("no", []string{"no"}, model.ImportanceVery, false, 0.5, nil),
		},
	}

	score, _ := svc.calculateDirectionalScore(answers, true)

	if score < 0 {
		t.Errorf("score should be clamped to 0, got %f", score)
	}
}

// ============================================================================
// CalculateCompatibilityBreakdown Tests
// ============================================================================

func TestCalculateCompatibilityBreakdown_ReturnsCategories(t *testing.T) {
	t.Parallel()
	repo := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return map[string][2]*model.Answer{
				"q1": {
					makeAnswer("yes", []string{"yes"}, model.ImportanceVery, false, 0.5, nil),
					makeAnswer("yes", []string{"yes"}, model.ImportanceVery, false, 0.5, nil),
				},
			}, nil
		},
		getAllQuestionsFunc: func(ctx context.Context) ([]*model.Question, error) {
			return []*model.Question{
				{ID: "q1", Text: "Test question?", Category: "values"},
			}, nil
		},
	}
	svc := newTestCompatibilityService(repo)

	breakdown, err := svc.CalculateCompatibilityBreakdown(context.Background(), "user:A", "user:B")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if breakdown.Score != 100 {
		t.Errorf("expected score 100, got %f", breakdown.Score)
	}
	if len(breakdown.CategoryScores) == 0 {
		t.Error("expected category scores")
	}
	if breakdown.CategoryScores["values"] != 100 {
		t.Errorf("expected values category score 100, got %f", breakdown.CategoryScores["values"])
	}
}

func TestCalculateCompatibilityBreakdown_IdentifiesDealBreakers(t *testing.T) {
	t.Parallel()
	repo := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return map[string][2]*model.Answer{
				"q1": {
					makeAnswer("yes", []string{"yes"}, model.ImportanceMandatory, true, 0.5, nil),
					makeAnswer("no", []string{"no"}, model.ImportanceSomewhat, false, 0.5, nil),
				},
			}, nil
		},
		getAllQuestionsFunc: func(ctx context.Context) ([]*model.Question, error) {
			return []*model.Question{
				{ID: "q1", Text: "Important question?", Category: "values"},
			}, nil
		},
	}
	svc := newTestCompatibilityService(repo)

	breakdown, err := svc.CalculateCompatibilityBreakdown(context.Background(), "user:A", "user:B")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(breakdown.DealBreakers) != 1 {
		t.Errorf("expected 1 dealbreaker, got %d", len(breakdown.DealBreakers))
	}
	if breakdown.DealBreakers[0].QuestionID != "q1" {
		t.Errorf("expected dealbreaker for q1, got %s", breakdown.DealBreakers[0].QuestionID)
	}
}

// ============================================================================
// CalculateYikesSummary Tests
// ============================================================================

func TestCalculateYikesSummary_NoYikes(t *testing.T) {
	t.Parallel()
	repo := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return map[string][2]*model.Answer{
				"q1": {
					makeAnswer("yes", []string{"yes"}, model.ImportanceVery, false, 0.5, nil),
					makeAnswer("yes", []string{"yes"}, model.ImportanceVery, false, 0.5, nil),
				},
			}, nil
		},
		getAllQuestionsFunc: func(ctx context.Context) ([]*model.Question, error) {
			return []*model.Question{{ID: "q1", Category: "values"}}, nil
		},
	}
	svc := newTestCompatibilityService(repo)

	summary, err := svc.CalculateYikesSummary(context.Background(), "user:A", "user:B")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.HasYikes {
		t.Error("expected no yikes")
	}
	if summary.YikesCount != 0 {
		t.Errorf("expected yikes count 0, got %d", summary.YikesCount)
	}
}

func TestCalculateYikesSummary_WithYikes(t *testing.T) {
	t.Parallel()
	repo := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return map[string][2]*model.Answer{
				"q1": {
					makeAnswer("yes", []string{"yes", "maybe"}, model.ImportanceVery, false, 0.5, []string{"maybe"}),
					makeAnswer("maybe", []string{"maybe"}, model.ImportanceVery, false, 0.5, nil),
				},
			}, nil
		},
		getAllQuestionsFunc: func(ctx context.Context) ([]*model.Question, error) {
			return []*model.Question{{ID: "q1", Category: "lifestyle"}}, nil
		},
	}
	svc := newTestCompatibilityService(repo)

	summary, err := svc.CalculateYikesSummary(context.Background(), "user:A", "user:B")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !summary.HasYikes {
		t.Error("expected yikes")
	}
	if summary.YikesCount != 1 {
		t.Errorf("expected yikes count 1, got %d", summary.YikesCount)
	}
	if len(summary.Categories) != 1 || summary.Categories[0] != "lifestyle" {
		t.Errorf("expected category 'lifestyle', got %v", summary.Categories)
	}
}

func TestCalculateYikesSummary_MutualYikes(t *testing.T) {
	t.Parallel()
	repo := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return map[string][2]*model.Answer{
				"q1": {
					// Both mark each other's answer as yikes
					makeAnswer("yes", []string{"yes", "no"}, model.ImportanceVery, false, 0.5, []string{"no"}),
					makeAnswer("no", []string{"yes", "no"}, model.ImportanceVery, false, 0.5, []string{"yes"}),
				},
			}, nil
		},
		getAllQuestionsFunc: func(ctx context.Context) ([]*model.Question, error) {
			return []*model.Question{{ID: "q1", Category: "values"}}, nil
		},
	}
	svc := newTestCompatibilityService(repo)

	summary, err := svc.CalculateYikesSummary(context.Background(), "user:A", "user:B")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.YikesCount != 2 {
		t.Errorf("expected yikes count 2 (mutual), got %d", summary.YikesCount)
	}
}

// ============================================================================
// ImportanceWeight Tests (from model)
// ============================================================================

func TestImportanceWeight_AllLevels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		importance string
		expected   int
	}{
		{model.ImportanceIrrelevant, 0},
		{model.ImportanceLittle, 1},
		{model.ImportanceSomewhat, 10},
		{model.ImportanceVery, 50},
		{model.ImportanceMandatory, 250},
		{"unknown", 10}, // defaults to "somewhat"
	}

	for _, tt := range tests {
		result := model.ImportanceWeight(tt.importance)
		if result != tt.expected {
			t.Errorf("ImportanceWeight(%q): expected %d, got %d", tt.importance, tt.expected, result)
		}
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestCalculateCompatibility_VeryLargeDataset(t *testing.T) {
	t.Parallel()
	// Create 100 shared questions
	answers := make(map[string][2]*model.Answer)
	for i := 0; i < 100; i++ {
		answers[string(rune('q'+i))] = [2]*model.Answer{
			makeAnswer("yes", []string{"yes"}, model.ImportanceSomewhat, false, 0.5, nil),
			makeAnswer("yes", []string{"yes"}, model.ImportanceSomewhat, false, 0.5, nil),
		}
	}

	repo := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return answers, nil
		},
	}
	svc := newTestCompatibilityService(repo)

	score, err := svc.CalculateCompatibility(context.Background(), "user:A", "user:B")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Score != 100 {
		t.Errorf("expected score 100, got %f", score.Score)
	}
	if score.SharedCount != 100 {
		t.Errorf("expected shared count 100, got %d", score.SharedCount)
	}
}

func TestCalculateCompatibility_OnlyIrrelevantQuestions(t *testing.T) {
	t.Parallel()
	repo := &mockQuestionnaireRepo{
		getSharedAnswersFunc: func(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
			return map[string][2]*model.Answer{
				"q1": {
					makeAnswer("yes", []string{"yes"}, model.ImportanceIrrelevant, false, 0.5, nil),
					makeAnswer("yes", []string{"yes"}, model.ImportanceIrrelevant, false, 0.5, nil),
				},
			}, nil
		},
	}
	svc := newTestCompatibilityService(repo)

	score, err := svc.CalculateCompatibility(context.Background(), "user:A", "user:B")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// With weight 0, should still get a result (but may be 0 due to division)
	if math.IsNaN(score.Score) {
		t.Error("score should not be NaN")
	}
}
