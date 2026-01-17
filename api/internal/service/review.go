package service

import (
	"context"

	"github.com/forgo/saga/api/internal/model"
)

// Error definitions moved to errors.go

// ReviewRepository defines the interface for review storage
type ReviewRepository interface {
	Create(ctx context.Context, review *model.Review) error
	GetByID(ctx context.Context, id string) (*model.Review, error)
	GetReviewsGiven(ctx context.Context, userID string, limit, offset int) ([]*model.Review, error)
	GetReviewsReceived(ctx context.Context, userID string, limit, offset int) ([]*model.Review, error)
	HasReviewed(ctx context.Context, reviewerID, revieweeID, referenceID string) (bool, error)
	GetReputation(ctx context.Context, userID string) (*model.Reputation, error)
	GetReputationDisplay(ctx context.Context, userID string) (*model.ReputationDisplay, error)
}

// ReviewService handles review business logic
type ReviewService struct {
	repo ReviewRepository
}

// ReviewServiceConfig holds configuration for the review service
type ReviewServiceConfig struct {
	Repo ReviewRepository
}

// NewReviewService creates a new review service
func NewReviewService(cfg ReviewServiceConfig) *ReviewService {
	return &ReviewService{
		repo: cfg.Repo,
	}
}

// CreateReview creates a new review
func (s *ReviewService) CreateReview(ctx context.Context, reviewerID string, req *model.CreateReviewRequest) (*model.Review, error) {
	// Cannot review yourself
	if reviewerID == req.RevieweeID {
		return nil, ErrCannotReviewSelf
	}

	// Validate context
	if !isValidReviewContext(req.Context) {
		return nil, ErrInvalidReviewContext
	}

	// Validate tag counts
	if len(req.PositiveTags) > model.MaxTagsPerReview {
		return nil, ErrTooManyTags
	}
	if len(req.ImprovementTags) > model.MaxTagsPerReview {
		return nil, ErrTooManyTags
	}

	// Validate private note length
	if req.PrivateNote != nil && len(*req.PrivateNote) > model.MaxPrivateNoteLength {
		return nil, ErrPrivateNoteTooLong
	}

	// Check for duplicate review if reference ID provided
	if req.ReferenceID != nil {
		alreadyReviewed, err := s.repo.HasReviewed(ctx, reviewerID, req.RevieweeID, *req.ReferenceID)
		if err != nil {
			return nil, err
		}
		if alreadyReviewed {
			return nil, ErrAlreadyReviewed
		}
	}

	// Validate tags are from known sets
	validPositiveTags := filterValidTags(req.PositiveTags, getPositiveTagSet())
	validImprovementTags := filterValidTags(req.ImprovementTags, getImprovementTagSet())

	review := &model.Review{
		ReviewerID:      reviewerID,
		RevieweeID:      req.RevieweeID,
		Context:         req.Context,
		ReferenceID:     req.ReferenceID,
		WouldMeetAgain:  req.WouldMeetAgain,
		PositiveTags:    validPositiveTags,
		ImprovementTags: validImprovementTags,
		PrivateNote:     req.PrivateNote,
	}

	if err := s.repo.Create(ctx, review); err != nil {
		return nil, err
	}

	return review, nil
}

// GetReview retrieves a review by ID
func (s *ReviewService) GetReview(ctx context.Context, id string) (*model.Review, error) {
	review, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if review == nil {
		return nil, ErrReviewNotFound
	}
	return review, nil
}

// GetReviewsGiven retrieves reviews given by a user
func (s *ReviewService) GetReviewsGiven(ctx context.Context, userID string, limit, offset int) ([]*model.Review, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.repo.GetReviewsGiven(ctx, userID, limit, offset)
}

// GetReviewsReceived retrieves reviews received by a user
func (s *ReviewService) GetReviewsReceived(ctx context.Context, userID string, limit, offset int) ([]*model.Review, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.repo.GetReviewsReceived(ctx, userID, limit, offset)
}

// GetReputation retrieves full reputation data for a user
func (s *ReviewService) GetReputation(ctx context.Context, userID string) (*model.Reputation, error) {
	return s.repo.GetReputation(ctx, userID)
}

// GetReputationDisplay retrieves profile-ready reputation display
func (s *ReviewService) GetReputationDisplay(ctx context.Context, userID string) (*model.ReputationDisplay, error) {
	return s.repo.GetReputationDisplay(ctx, userID)
}

// SubmitEventFeedback processes post-event feedback
func (s *ReviewService) SubmitEventFeedback(ctx context.Context, userID string, req *model.PostEventFeedbackRequest) error {
	// If user didn't attend, no feedback to process
	if req.Attended == "no" {
		return nil
	}

	// Convert to reviews based on context
	// The caller should pass in the host and other attendees separately
	// This is a simplified version - real implementation would look up event details

	return nil
}

// Helper functions

func isValidReviewContext(ctx string) bool {
	switch ctx {
	case model.ReviewContextHosted,
		model.ReviewContextWasGuest,
		model.ReviewContextEvent,
		model.ReviewContextMatched,
		model.ReviewContextHangout:
		return true
	default:
		return false
	}
}

func getPositiveTagSet() map[string]bool {
	tags := model.GetPositiveTags()
	set := make(map[string]bool, len(tags))
	for _, t := range tags {
		set[t.Tag] = true
	}
	return set
}

func getImprovementTagSet() map[string]bool {
	tags := model.GetImprovementTags()
	set := make(map[string]bool, len(tags))
	for _, t := range tags {
		set[t.Tag] = true
	}
	return set
}

func filterValidTags(input []string, validSet map[string]bool) []string {
	result := make([]string, 0, len(input))
	for _, tag := range input {
		if validSet[tag] {
			result = append(result, tag)
		}
	}
	return result
}
