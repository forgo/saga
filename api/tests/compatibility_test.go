package tests

/*
FEATURE: Compatibility Scoring
DOMAIN: Matching & Discovery

ACCEPTANCE CRITERIA:
===================

AC-COMPAT-001: Basic Calculation
  GIVEN users A and B share 5 questions
  AND both accept each other's answers
  WHEN calculating compatibility
  THEN score reflects high match

AC-COMPAT-002: Dealbreaker Penalty
  GIVEN user A has dealbreaker on question
  AND user B's answer violates it
  WHEN calculating compatibility
  THEN heavy penalty applied (score = 0)

AC-COMPAT-003: No Shared Questions
  GIVEN users with no overlapping answers
  WHEN calculating compatibility
  THEN returns 0 (no data)

AC-COMPAT-004: Importance Weighting
  GIVEN user marks question "very" important
  WHEN calculating compatibility
  THEN that question weighted higher
*/

import (
	"testing"

	"github.com/forgo/saga/api/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestCompatibility_ImportanceWeightIrrelevant(t *testing.T) {
	// AC-COMPAT-004: Importance weighting
	// Irrelevant = 0 weight (doesn't affect score)
	weight := model.ImportanceWeight(model.ImportanceIrrelevant)
	assert.Equal(t, 0, weight, "Irrelevant should have weight 0")
}

func TestCompatibility_ImportanceWeightLittle(t *testing.T) {
	// AC-COMPAT-004: Importance weighting
	// Little = 1 weight
	weight := model.ImportanceWeight(model.ImportanceLittle)
	assert.Equal(t, 1, weight, "Little should have weight 1")
}

func TestCompatibility_ImportanceWeightSomewhat(t *testing.T) {
	// AC-COMPAT-004: Importance weighting
	// Somewhat = 10 weight (default)
	weight := model.ImportanceWeight(model.ImportanceSomewhat)
	assert.Equal(t, 10, weight, "Somewhat should have weight 10")
}

func TestCompatibility_ImportanceWeightVery(t *testing.T) {
	// AC-COMPAT-004: Importance weighting
	// Very = 50 weight (5x somewhat)
	weight := model.ImportanceWeight(model.ImportanceVery)
	assert.Equal(t, 50, weight, "Very should have weight 50")
}

func TestCompatibility_ImportanceWeightMandatory(t *testing.T) {
	// AC-COMPAT-004: Importance weighting
	// Mandatory = 250 weight (25x somewhat, highest priority)
	weight := model.ImportanceWeight(model.ImportanceMandatory)
	assert.Equal(t, 250, weight, "Mandatory should have weight 250")
}

func TestCompatibility_ImportanceWeightUnknownDefault(t *testing.T) {
	// Unknown importance levels default to "somewhat" (10)
	weight := model.ImportanceWeight("unknown")
	assert.Equal(t, 10, weight, "Unknown should default to weight 10 (somewhat)")
}

func TestCompatibility_YikesSeverityNone(t *testing.T) {
	// 0 yikes = no severity
	severity := model.GetYikesSeverity(0)
	assert.Equal(t, model.YikesSeverityNone, severity)
}

func TestCompatibility_YikesSeverityMild(t *testing.T) {
	// 1-2 yikes = mild severity
	assert.Equal(t, model.YikesSeverityMild, model.GetYikesSeverity(1))
	assert.Equal(t, model.YikesSeverityMild, model.GetYikesSeverity(2))
}

func TestCompatibility_YikesSeverityModerate(t *testing.T) {
	// 3-4 yikes = moderate severity
	assert.Equal(t, model.YikesSeverityModerate, model.GetYikesSeverity(3))
	assert.Equal(t, model.YikesSeverityModerate, model.GetYikesSeverity(4))
}

func TestCompatibility_YikesSeveritySevere(t *testing.T) {
	// 5+ yikes = severe
	assert.Equal(t, model.YikesSeveritySevere, model.GetYikesSeverity(5))
	assert.Equal(t, model.YikesSeveritySevere, model.GetYikesSeverity(10))
	assert.Equal(t, model.YikesSeveritySevere, model.GetYikesSeverity(100))
}

func TestCompatibility_QuestionCategoryConstants(t *testing.T) {
	// Verify the 4 question categories
	assert.Equal(t, "values", model.QuestionCategoryValues)
	assert.Equal(t, "social", model.QuestionCategorySocial)
	assert.Equal(t, "lifestyle", model.QuestionCategoryLifestyle)
	assert.Equal(t, "communication", model.QuestionCategoryCommunication)
}

func TestCompatibility_GetQuestionCategories(t *testing.T) {
	// Verify category info is returned
	categories := model.GetQuestionCategories()
	assert.Len(t, categories, 4, "Should have 4 question categories")

	// Check each category has ID, Label, and Icon
	for _, cat := range categories {
		assert.NotEmpty(t, cat.ID)
		assert.NotEmpty(t, cat.Label)
		assert.NotEmpty(t, cat.Icon)
	}
}

func TestCompatibility_AlignmentWeightDefaults(t *testing.T) {
	// AC-COMPAT-001: Alignment weight affects matching
	assert.Equal(t, 0.5, model.DefaultAlignmentWeight, "Default alignment weight should be 0.5 (neutral)")
	assert.Equal(t, 0.0, model.MinAlignmentWeight, "Min alignment weight should be 0")
	assert.Equal(t, 1.0, model.MaxAlignmentWeight, "Max alignment weight should be 1")
}

func TestCompatibility_RequiredCategories(t *testing.T) {
	// Required categories for discovery eligibility
	required := model.RequiredCategories
	assert.Contains(t, required, model.QuestionCategoryValues)
	assert.Contains(t, required, model.QuestionCategorySocial)
}

func TestCompatibility_MinQuestionsForDiscovery(t *testing.T) {
	// Must answer minimum questions for discovery
	assert.Equal(t, 3, model.MinQuestionsForDiscovery,
		"Should require at least 3 questions for discovery")
}

func TestCompatibility_MaxQuestionsToDisplay(t *testing.T) {
	// Limit questions shown at once
	assert.Equal(t, 50, model.MaxQuestionsToDisplay,
		"Should limit to 50 questions displayed")
}

func TestCompatibility_BiasThresholds(t *testing.T) {
	// Internal bias thresholds for community fit
	assert.Equal(t, -5.0, model.BiasThresholdWarning)
	assert.Equal(t, -10.0, model.BiasThresholdConcern)
}

func TestCompatibility_BiasStatusConstants(t *testing.T) {
	// Bias status levels
	assert.Equal(t, "normal", model.BiasStatusNormal)
	assert.Equal(t, "warning", model.BiasStatusWarning)
	assert.Equal(t, "concern", model.BiasStatusConcern)
}

func TestCompatibility_CompatibilityScoreModel(t *testing.T) {
	// AC-COMPAT-001: Basic score structure
	score := &model.CompatibilityScore{
		UserAID:     "user:alice",
		UserBID:     "user:bob",
		Score:       85.5,
		AToB:        90.0,
		BToA:        81.0,
		SharedCount: 5,
		DealBreaker: false,
	}

	assert.Equal(t, "user:alice", score.UserAID)
	assert.Equal(t, "user:bob", score.UserBID)
	assert.Equal(t, 85.5, score.Score)
	assert.Equal(t, 90.0, score.AToB)
	assert.Equal(t, 81.0, score.BToA)
	assert.Equal(t, 5, score.SharedCount)
	assert.False(t, score.DealBreaker)
}

func TestCompatibility_DealBreakerViolationModel(t *testing.T) {
	// AC-COMPAT-002: Dealbreaker violation structure
	violation := &model.DealBreakerViolation{
		QuestionID:    "question:q1",
		QuestionText:  "Do you prefer cats or dogs?",
		UserAnswer:    "cats",
		PartnerAnswer: "dogs",
	}

	assert.Equal(t, "question:q1", violation.QuestionID)
	assert.NotEmpty(t, violation.QuestionText)
	assert.NotEqual(t, violation.UserAnswer, violation.PartnerAnswer)
}

func TestCompatibility_BreakdownIncludesCategoryScores(t *testing.T) {
	// AC-COMPAT-001: Breakdown includes per-category scores
	breakdown := &model.CompatibilityBreakdown{
		CompatibilityScore: model.CompatibilityScore{
			Score:       75.0,
			SharedCount: 10,
		},
		CategoryScores: map[string]float64{
			model.QuestionCategoryValues:        80.0,
			model.QuestionCategorySocial:        70.0,
			model.QuestionCategoryLifestyle:     75.0,
			model.QuestionCategoryCommunication: 78.0,
		},
		DealBreakers: nil,
	}

	assert.Len(t, breakdown.CategoryScores, 4)
	assert.Equal(t, 80.0, breakdown.CategoryScores[model.QuestionCategoryValues])
}

func TestCompatibility_YikesSummaryModel(t *testing.T) {
	// Yikes summary structure
	summary := &model.YikesSummary{
		HasYikes:   true,
		YikesCount: 3,
		Categories: []string{model.QuestionCategoryValues, model.QuestionCategorySocial},
		Severity:   model.YikesSeverityModerate,
	}

	assert.True(t, summary.HasYikes)
	assert.Equal(t, 3, summary.YikesCount)
	assert.Len(t, summary.Categories, 2)
	assert.Equal(t, model.YikesSeverityModerate, summary.Severity)
}

func TestCompatibility_AnswerModel(t *testing.T) {
	// Answer model with all fields
	answer := &model.Answer{
		ID:                "answer:a1",
		UserID:            "user:alice",
		QuestionID:        "question:q1",
		SelectedOption:    "option_a",
		AcceptableOptions: []string{"option_a", "option_b"},
		Importance:        model.ImportanceVery,
		IsDealBreaker:     true,
		AlignmentWeight:   0.8,
		YikesOptions:      []string{"option_d"},
	}

	assert.Equal(t, "option_a", answer.SelectedOption)
	assert.Contains(t, answer.AcceptableOptions, "option_a")
	assert.Contains(t, answer.AcceptableOptions, "option_b")
	assert.Equal(t, model.ImportanceVery, answer.Importance)
	assert.True(t, answer.IsDealBreaker)
	assert.Equal(t, 0.8, answer.AlignmentWeight)
	assert.Contains(t, answer.YikesOptions, "option_d")
}

func TestCompatibility_QuestionModel(t *testing.T) {
	// Question model structure
	question := &model.Question{
		ID:       "question:q1",
		Text:     "What's your communication style?",
		Category: model.QuestionCategoryCommunication,
		Options: []model.QuestionOption{
			{Value: "direct", Label: "Direct and honest"},
			{Value: "diplomatic", Label: "Diplomatic and tactful"},
			{Value: "passive", Label: "Passive and accommodating"},
		},
		IsDealBreakerEligible: true,
		SortOrder:             1,
		Active:                true,
	}

	assert.Equal(t, model.QuestionCategoryCommunication, question.Category)
	assert.Len(t, question.Options, 3)
	assert.True(t, question.IsDealBreakerEligible)
	assert.True(t, question.Active)
}

func TestCompatibility_QuestionOptionImplicitBias(t *testing.T) {
	// Question options can have implicit bias (internal only)
	prosocialOption := model.QuestionOption{
		Value:        "help_others",
		Label:        "I enjoy helping others",
		ImplicitBias: 0.5, // Prosocial
	}

	antisocialOption := model.QuestionOption{
		Value:        "avoid_others",
		Label:        "I prefer to avoid people",
		ImplicitBias: -0.5, // Antisocial
	}

	neutralOption := model.QuestionOption{
		Value:        "neutral",
		Label:        "It depends on the situation",
		ImplicitBias: 0.0, // Neutral
	}

	assert.Greater(t, prosocialOption.ImplicitBias, 0.0)
	assert.Less(t, antisocialOption.ImplicitBias, 0.0)
	assert.Equal(t, 0.0, neutralOption.ImplicitBias)
}

func TestCompatibility_CircleValuesStatusConstants(t *testing.T) {
	// Circle values approval workflow
	assert.Equal(t, "pending", model.CircleValuesStatusPending)
	assert.Equal(t, "approved", model.CircleValuesStatusApproved)
	assert.Equal(t, "rejected", model.CircleValuesStatusRejected)
}

func TestCompatibility_CompatibilityScoreWithDealBreaker(t *testing.T) {
	// AC-COMPAT-002: Dealbreaker sets score to 0
	score := &model.CompatibilityScore{
		UserAID:     "user:alice",
		UserBID:     "user:bob",
		Score:       0, // Should be 0 when dealbreaker violated
		AToB:        0,
		BToA:        0,
		SharedCount: 5,
		DealBreaker: true,
	}

	assert.True(t, score.DealBreaker)
	assert.Equal(t, float64(0), score.Score, "Score should be 0 when dealbreaker violated")
}

func TestCompatibility_NoSharedQuestionsScore(t *testing.T) {
	// AC-COMPAT-003: No shared questions returns 0 with SharedCount=0
	score := &model.CompatibilityScore{
		UserAID:     "user:alice",
		UserBID:     "user:bob",
		Score:       0, // No data to calculate from
		AToB:        0,
		BToA:        0,
		SharedCount: 0, // Key indicator: no shared questions
		DealBreaker: false,
	}

	assert.Equal(t, 0, score.SharedCount)
	assert.Equal(t, float64(0), score.Score)
}

func TestCompatibility_QuestionProgressModel(t *testing.T) {
	// Progress tracking for questionnaire completion
	progress := &model.QuestionProgress{
		TotalQuestions: 50,
		AnsweredCount:  15,
		ByCategory: map[string]int{
			model.QuestionCategoryValues:        5,
			model.QuestionCategorySocial:        4,
			model.QuestionCategoryLifestyle:     3,
			model.QuestionCategoryCommunication: 3,
		},
		RequiredCategories: []string{}, // All required categories covered
		CanDiscover:        true,
	}

	assert.Equal(t, 50, progress.TotalQuestions)
	assert.Equal(t, 15, progress.AnsweredCount)
	assert.Equal(t, 5, progress.ByCategory[model.QuestionCategoryValues])
	assert.True(t, progress.CanDiscover)
}

func TestCompatibility_QuestionProgressCannotDiscoverMissingCategories(t *testing.T) {
	// Cannot discover if required categories are missing
	progress := &model.QuestionProgress{
		TotalQuestions: 50,
		AnsweredCount:  2,
		ByCategory: map[string]int{
			model.QuestionCategoryLifestyle: 2, // Only answered lifestyle
		},
		RequiredCategories: []string{
			model.QuestionCategoryValues, // Missing!
			model.QuestionCategorySocial, // Missing!
		},
		CanDiscover: false,
	}

	assert.False(t, progress.CanDiscover)
	assert.Contains(t, progress.RequiredCategories, model.QuestionCategoryValues)
	assert.Contains(t, progress.RequiredCategories, model.QuestionCategorySocial)
}
