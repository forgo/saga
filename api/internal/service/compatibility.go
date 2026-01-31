package service

import (
	"context"
	"math"

	"github.com/forgo/saga/api/internal/model"
)

// CompatibilityService handles compatibility calculations
type CompatibilityService struct {
	questionnaireRepo QuestionnaireRepository
}

// CompatibilityServiceConfig holds configuration for the compatibility service
type CompatibilityServiceConfig struct {
	QuestionnaireRepo QuestionnaireRepository
}

// NewCompatibilityService creates a new compatibility service
func NewCompatibilityService(cfg CompatibilityServiceConfig) *CompatibilityService {
	return &CompatibilityService{
		questionnaireRepo: cfg.QuestionnaireRepo,
	}
}

// CalculateCompatibility calculates compatibility between two users
// Uses OkCupid-style weighted scoring with alignment weights and yikes detection
func (s *CompatibilityService) CalculateCompatibility(ctx context.Context, userAID, userBID string) (*model.CompatibilityScore, error) {
	// Get shared answers (questions both users have answered)
	sharedAnswers, err := s.questionnaireRepo.GetSharedAnswers(ctx, userAID, userBID)
	if err != nil {
		return nil, err
	}

	if len(sharedAnswers) == 0 {
		return &model.CompatibilityScore{
			UserAID:     userAID,
			UserBID:     userBID,
			Score:       0,
			AToB:        0,
			BToA:        0,
			SharedCount: 0,
			DealBreaker: false,
		}, nil
	}

	// Calculate A→B (how well B matches A's preferences)
	aToBScore, aToBDealBreaker := s.calculateDirectionalScore(sharedAnswers, true)

	// Calculate B→A (how well A matches B's preferences)
	bToAScore, bToADealBreaker := s.calculateDirectionalScore(sharedAnswers, false)

	// Overall score is geometric mean of both directions
	overallScore := math.Sqrt(aToBScore * bToAScore)

	// If either has a dealbreaker, score is 0
	hasDealBreaker := aToBDealBreaker || bToADealBreaker
	if hasDealBreaker {
		overallScore = 0
	}

	return &model.CompatibilityScore{
		UserAID:     userAID,
		UserBID:     userBID,
		Score:       overallScore,
		AToB:        aToBScore,
		BToA:        bToAScore,
		SharedCount: len(sharedAnswers),
		DealBreaker: hasDealBreaker,
	}, nil
}

// CalculateCompatibilityBreakdown provides detailed scoring info
func (s *CompatibilityService) CalculateCompatibilityBreakdown(ctx context.Context, userAID, userBID string) (*model.CompatibilityBreakdown, error) {
	// Get shared answers
	sharedAnswers, err := s.questionnaireRepo.GetSharedAnswers(ctx, userAID, userBID)
	if err != nil {
		return nil, err
	}

	score, err := s.CalculateCompatibility(ctx, userAID, userBID)
	if err != nil {
		return nil, err
	}

	// Get questions for category breakdown
	questions := make(map[string]*model.Question)
	allQuestions, err := s.questionnaireRepo.GetAllQuestions(ctx)
	if err != nil {
		return nil, err
	}
	for _, q := range allQuestions {
		questions[q.ID] = q
	}

	// Calculate per-category scores and find dealbreakers
	categoryScores := make(map[string]float64)
	categoryWeights := make(map[string]float64)
	var dealBreakers []model.DealBreakerViolation

	for questionID, answers := range sharedAnswers {
		answerA := answers[0]
		answerB := answers[1]
		if answerA == nil || answerB == nil {
			continue
		}

		question := questions[questionID]
		if question == nil {
			continue
		}

		// Check dealbreaker violations
		if answerA.IsDealBreaker {
			if !containsString(answerA.AcceptableOptions, answerB.SelectedOption) {
				dealBreakers = append(dealBreakers, model.DealBreakerViolation{
					QuestionID:    questionID,
					QuestionText:  question.Text,
					UserAnswer:    answerA.SelectedOption,
					PartnerAnswer: answerB.SelectedOption,
				})
			}
		}
		if answerB.IsDealBreaker {
			if !containsString(answerB.AcceptableOptions, answerA.SelectedOption) {
				dealBreakers = append(dealBreakers, model.DealBreakerViolation{
					QuestionID:    questionID,
					QuestionText:  question.Text,
					UserAnswer:    answerB.SelectedOption,
					PartnerAnswer: answerA.SelectedOption,
				})
			}
		}

		// Calculate category contribution
		weight := float64(model.ImportanceWeight(answerA.Importance) + model.ImportanceWeight(answerB.Importance))
		earned := 0.0

		if containsString(answerA.AcceptableOptions, answerB.SelectedOption) {
			earned += float64(model.ImportanceWeight(answerA.Importance))
		}
		if containsString(answerB.AcceptableOptions, answerA.SelectedOption) {
			earned += float64(model.ImportanceWeight(answerB.Importance))
		}

		categoryWeights[question.Category] += weight
		categoryScores[question.Category] += earned
	}

	// Normalize category scores
	for cat := range categoryScores {
		if categoryWeights[cat] > 0 {
			categoryScores[cat] = (categoryScores[cat] / categoryWeights[cat]) * 100
		}
	}

	return &model.CompatibilityBreakdown{
		CompatibilityScore: *score,
		CategoryScores:     categoryScores,
		DealBreakers:       dealBreakers,
	}, nil
}

// CalculateYikesSummary calculates red flag summary between two users
func (s *CompatibilityService) CalculateYikesSummary(ctx context.Context, userAID, userBID string) (*model.YikesSummary, error) {
	sharedAnswers, err := s.questionnaireRepo.GetSharedAnswers(ctx, userAID, userBID)
	if err != nil {
		return nil, err
	}

	// Get questions for categories
	questions := make(map[string]*model.Question)
	allQuestions, err := s.questionnaireRepo.GetAllQuestions(ctx)
	if err != nil {
		return nil, err
	}
	for _, q := range allQuestions {
		questions[q.ID] = q
	}

	yikesCount := 0
	categoriesWithYikes := make(map[string]bool)

	for questionID, answers := range sharedAnswers {
		answerA := answers[0]
		answerB := answers[1]
		if answerA == nil || answerB == nil {
			continue
		}

		question := questions[questionID]

		// Check if B's answer is in A's yikes list
		if containsString(answerA.YikesOptions, answerB.SelectedOption) {
			yikesCount++
			if question != nil {
				categoriesWithYikes[question.Category] = true
			}
		}

		// Check if A's answer is in B's yikes list
		if containsString(answerB.YikesOptions, answerA.SelectedOption) {
			yikesCount++
			if question != nil {
				categoriesWithYikes[question.Category] = true
			}
		}
	}

	categories := make([]string, 0, len(categoriesWithYikes))
	for cat := range categoriesWithYikes {
		categories = append(categories, cat)
	}

	return &model.YikesSummary{
		HasYikes:   yikesCount > 0,
		YikesCount: yikesCount,
		Categories: categories,
		Severity:   model.GetYikesSeverity(yikesCount),
	}, nil
}

// calculateDirectionalScore calculates how well one user matches the other's preferences
// If aToB is true, calculates how well B matches A's preferences
// If aToB is false, calculates how well A matches B's preferences
func (s *CompatibilityService) calculateDirectionalScore(sharedAnswers map[string][2]*model.Answer, aToB bool) (float64, bool) {
	var totalWeight float64
	var earnedPoints float64
	dealBreakerViolated := false

	for _, answers := range sharedAnswers {
		var evaluator, evaluated *model.Answer
		if aToB {
			evaluator = answers[0] // A evaluates B
			evaluated = answers[1] // B's answer
		} else {
			evaluator = answers[1] // B evaluates A
			evaluated = answers[0] // A's answer
		}

		if evaluator == nil || evaluated == nil {
			continue
		}

		// Get base importance weight
		baseWeight := float64(model.ImportanceWeight(evaluator.Importance))

		// Apply alignment weight (0-1 scale affects contribution)
		// Higher alignment weight = this question matters more for matching
		effectiveWeight := baseWeight * (0.5 + evaluator.AlignmentWeight*0.5)

		totalWeight += effectiveWeight

		// Check if evaluated's answer is acceptable to evaluator
		if containsString(evaluator.AcceptableOptions, evaluated.SelectedOption) {
			earnedPoints += effectiveWeight
		} else {
			// Check if dealbreaker
			if evaluator.IsDealBreaker {
				dealBreakerViolated = true
			}
		}

		// Apply yikes penalty (reduce score even if "acceptable")
		if containsString(evaluator.YikesOptions, evaluated.SelectedOption) {
			// Reduce earned points by 25% for each yikes
			earnedPoints -= effectiveWeight * 0.25
		}
	}

	if totalWeight == 0 {
		return 0, dealBreakerViolated
	}

	// Calculate percentage (0-100)
	score := (earnedPoints / totalWeight) * 100

	// Clamp to 0-100 range (yikes can reduce below 0)
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score, dealBreakerViolated
}

// Helper function
func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
