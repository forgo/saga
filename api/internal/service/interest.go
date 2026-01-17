package service

import (
	"context"

	"github.com/forgo/saga/api/internal/model"
)

// Error definitions moved to errors.go

// InterestRepository defines the interface for interest storage
type InterestRepository interface {
	GetAll(ctx context.Context) ([]*model.Interest, error)
	GetByCategory(ctx context.Context, category string) ([]*model.Interest, error)
	GetByID(ctx context.Context, id string) (*model.Interest, error)
	Create(ctx context.Context, interest *model.Interest) error
	AddUserInterest(ctx context.Context, userID, interestID string, req *model.AddInterestRequest) error
	UpdateUserInterest(ctx context.Context, userID, interestID string, updates map[string]interface{}) error
	RemoveUserInterest(ctx context.Context, userID, interestID string) error
	GetUserInterests(ctx context.Context, userID string) ([]*model.UserInterest, error)
	FindTeachingMatches(ctx context.Context, userID string, limit int) ([]*model.InterestMatch, error)
	FindLearningMatches(ctx context.Context, userID string, limit int) ([]*model.InterestMatch, error)
	FindSharedInterests(ctx context.Context, userID string, limit int) ([]*model.SharedInterestUser, error)
	// Discovery-related methods
	GetUsersWithInterest(ctx context.Context, interestID string) ([]*model.UserInterest, error)
	GetTeachersForInterest(ctx context.Context, interestID string) ([]*model.UserInterest, error)
	GetLearnersForInterest(ctx context.Context, interestID string) ([]*model.UserInterest, error)
}

// InterestService handles interest business logic
type InterestService struct {
	interestRepo InterestRepository
}

// InterestServiceConfig holds configuration for the interest service
type InterestServiceConfig struct {
	InterestRepo InterestRepository
}

// NewInterestService creates a new interest service
func NewInterestService(cfg InterestServiceConfig) *InterestService {
	return &InterestService{
		interestRepo: cfg.InterestRepo,
	}
}

// GetAllInterests retrieves all available interests
func (s *InterestService) GetAllInterests(ctx context.Context) ([]*model.Interest, error) {
	return s.interestRepo.GetAll(ctx)
}

// GetInterestsByCategory retrieves interests by category
func (s *InterestService) GetInterestsByCategory(ctx context.Context, category string) ([]*model.Interest, error) {
	if !isValidInterestCategory(category) {
		return []*model.Interest{}, nil
	}
	return s.interestRepo.GetByCategory(ctx, category)
}

// GetUserInterests retrieves a user's interests
func (s *InterestService) GetUserInterests(ctx context.Context, userID string) ([]*model.UserInterest, error) {
	return s.interestRepo.GetUserInterests(ctx, userID)
}

// AddUserInterest adds an interest to a user's profile
func (s *InterestService) AddUserInterest(ctx context.Context, userID, interestID string, req *model.AddInterestRequest) error {
	// Validate interest exists
	interest, err := s.interestRepo.GetByID(ctx, interestID)
	if err != nil {
		return err
	}
	if interest == nil {
		return ErrInterestNotFound
	}

	// Validate level
	if req.Level != "" && !isValidInterestLevel(req.Level) {
		return ErrInvalidInterestLevel
	}
	if req.Level == "" {
		req.Level = string(model.InterestLevelInterested)
	}

	return s.interestRepo.AddUserInterest(ctx, userID, interestID, req)
}

// UpdateUserInterest updates a user's interest settings
func (s *InterestService) UpdateUserInterest(ctx context.Context, userID, interestID string, req *model.UpdateInterestRequest) error {
	updates := make(map[string]interface{})

	if req.Level != nil {
		if !isValidInterestLevel(*req.Level) {
			return ErrInvalidInterestLevel
		}
		updates["level"] = *req.Level
	}
	if req.WantsToTeach != nil {
		updates["wants_to_teach"] = *req.WantsToTeach
	}
	if req.WantsToLearn != nil {
		updates["wants_to_learn"] = *req.WantsToLearn
	}
	if req.Intent != nil {
		updates["intent"] = *req.Intent
	}

	if len(updates) == 0 {
		return nil
	}

	return s.interestRepo.UpdateUserInterest(ctx, userID, interestID, updates)
}

// RemoveUserInterest removes an interest from a user's profile
func (s *InterestService) RemoveUserInterest(ctx context.Context, userID, interestID string) error {
	return s.interestRepo.RemoveUserInterest(ctx, userID, interestID)
}

// FindTeachingMatches finds users who want to learn what the user can teach
func (s *InterestService) FindTeachingMatches(ctx context.Context, userID string, limit int) ([]*model.InterestMatch, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.interestRepo.FindTeachingMatches(ctx, userID, limit)
}

// FindLearningMatches finds users who can teach what the user wants to learn
func (s *InterestService) FindLearningMatches(ctx context.Context, userID string, limit int) ([]*model.InterestMatch, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.interestRepo.FindLearningMatches(ctx, userID, limit)
}

// FindSharedInterests finds users with the most shared interests
func (s *InterestService) FindSharedInterests(ctx context.Context, userID string, limit int) ([]*model.SharedInterestUser, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.interestRepo.FindSharedInterests(ctx, userID, limit)
}

// GetInterestStats returns statistics about a user's interests
func (s *InterestService) GetInterestStats(ctx context.Context, userID string) (*model.InterestStats, error) {
	interests, err := s.interestRepo.GetUserInterests(ctx, userID)
	if err != nil {
		return nil, err
	}

	stats := &model.InterestStats{
		TotalCount:    len(interests),
		ByCategory:    make(map[string]int),
		TeachingCount: 0,
		LearningCount: 0,
	}

	for _, ui := range interests {
		stats.ByCategory[ui.Category]++
		if ui.WantsToTeach {
			stats.TeachingCount++
		}
		if ui.WantsToLearn {
			stats.LearningCount++
		}
	}

	return stats, nil
}

// Helper functions

func isValidInterestLevel(level string) bool {
	switch model.InterestLevel(level) {
	case model.InterestLevelCurious,
		model.InterestLevelInterested,
		model.InterestLevelExperienced,
		model.InterestLevelExpert:
		return true
	default:
		return false
	}
}

func isValidInterestCategory(category string) bool {
	switch category {
	case model.InterestCategoryHobby,
		model.InterestCategorySport,
		model.InterestCategorySocial,
		model.InterestCategoryLearning,
		model.InterestCategoryOutdoors,
		model.InterestCategorySkill,
		model.InterestCategoryLanguage,
		model.InterestCategoryCuisine,
		model.InterestCategoryMusic,
		model.InterestCategoryArt,
		model.InterestCategoryTech:
		return true
	default:
		return false
	}
}
