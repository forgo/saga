package service

import (
	"context"
	"fmt"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// TrustRatingRepository defines the interface for trust rating storage
type TrustRatingRepository interface {
	Create(ctx context.Context, rating *model.TrustRating) error
	GetByID(ctx context.Context, id string) (*model.TrustRating, error)
	GetByRaterRateeAnchor(ctx context.Context, raterID, rateeID, anchorType, anchorID string) (*model.TrustRating, error)
	Update(ctx context.Context, id string, trustLevel model.TrustLevel, trustReview string) (*model.TrustRating, error)
	Delete(ctx context.Context, id string) error
	GetReceivedRatings(ctx context.Context, userID string, limit, offset int) ([]*model.TrustRating, error)
	GetGivenRatings(ctx context.Context, userID string, limit, offset int) ([]*model.TrustRating, error)
	GetAggregate(ctx context.Context, userID string) (*model.TrustAggregate, error)
	GetDailyCount(ctx context.Context, userID string) (int, error)
	CanRate(ctx context.Context, raterID, rateeID, anchorType, anchorID string) (bool, error)
	CreateEndorsement(ctx context.Context, endorsement *model.TrustEndorsement) error
	GetEndorsementsByRating(ctx context.Context, ratingID string) ([]*model.TrustEndorsement, error)
	GetEndorsementCounts(ctx context.Context, ratingID string) (agree, disagree int, err error)
	HasEndorsed(ctx context.Context, endorserID, ratingID string) (bool, error)
	GetDistrustSignals(ctx context.Context, minDistrust int, limit int) ([]*model.DistrustSignal, error)
}

// TrustRatingService handles trust rating business logic
type TrustRatingService struct {
	repo TrustRatingRepository
}

// TrustRatingServiceConfig holds configuration for the trust rating service
type TrustRatingServiceConfig struct {
	Repo TrustRatingRepository
}

// NewTrustRatingService creates a new trust rating service
func NewTrustRatingService(cfg TrustRatingServiceConfig) *TrustRatingService {
	return &TrustRatingService{
		repo: cfg.Repo,
	}
}

// Create creates a new trust rating
func (s *TrustRatingService) Create(ctx context.Context, raterID string, req *model.CreateTrustRatingRequest) (*model.TrustRating, error) {
	// Validate request
	if errors := req.Validate(); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	// Prevent self-rating
	if raterID == req.RateeID {
		return nil, model.NewBadRequestError("cannot rate yourself")
	}

	// Check daily limit
	count, err := s.repo.GetDailyCount(ctx, raterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check daily limit: %w", err)
	}
	if count >= model.MaxTrustRatingsPerDay {
		return nil, model.NewBadRequestError("daily rating limit reached (10 per day)")
	}

	// Check if rating already exists for this anchor
	existing, err := s.repo.GetByRaterRateeAnchor(ctx, raterID, req.RateeID, req.AnchorType, req.AnchorID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing rating: %w", err)
	}
	if existing != nil {
		return nil, model.NewConflictError("rating already exists for this interaction")
	}

	// Validate anchor (both users attended and event is verified)
	canRate, err := s.repo.CanRate(ctx, raterID, req.RateeID, req.AnchorType, req.AnchorID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate anchor: %w", err)
	}
	if !canRate {
		return nil, model.NewForbiddenError("cannot rate - anchor not verified or you didn't both attend")
	}

	rating := &model.TrustRating{
		RaterID:     raterID,
		RateeID:     req.RateeID,
		AnchorType:  model.TrustAnchorType(req.AnchorType),
		AnchorID:    req.AnchorID,
		TrustLevel:  model.TrustLevel(req.TrustLevel),
		TrustReview: req.TrustReview,
	}

	if err := s.repo.Create(ctx, rating); err != nil {
		return nil, fmt.Errorf("failed to create rating: %w", err)
	}

	return rating, nil
}

// GetByID retrieves a trust rating by ID with endorsement counts
func (s *TrustRatingService) GetByID(ctx context.Context, id string, viewerID string) (*model.TrustRating, error) {
	rating, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}
	if rating == nil {
		return nil, model.NewNotFoundError("trust rating not found")
	}

	// Hide distrust reviews from non-admins (except to the rater)
	if rating.ReviewVisibility == model.ReviewVisibilityAdminOnly && rating.RaterID != viewerID {
		rating.TrustReview = "[hidden]"
	}

	// Get endorsement counts
	agree, disagree, err := s.repo.GetEndorsementCounts(ctx, id)
	if err == nil {
		rating.AgreeCount = agree
		rating.DisagreeCount = disagree
		rating.EndorsementCount = agree + disagree
	}

	// Calculate cooldown
	rating.CanEdit = s.canEdit(rating)
	if !rating.CanEdit {
		nextEditable := rating.UpdatedOn.AddDate(0, 0, model.TrustRatingCooldownDays)
		rating.NextEditableAt = &nextEditable
	}

	return rating, nil
}

// Update updates a trust rating (subject to 30-day cooldown)
func (s *TrustRatingService) Update(ctx context.Context, id string, userID string, req *model.UpdateTrustRatingRequest) (*model.TrustRating, error) {
	// Validate request
	if errors := req.Validate(); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	// Get existing rating
	rating, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}
	if rating == nil {
		return nil, model.NewNotFoundError("trust rating not found")
	}

	// Check ownership
	if rating.RaterID != userID {
		return nil, model.NewForbiddenError("not your rating")
	}

	// Check cooldown
	if !s.canEdit(rating) {
		nextEditable := rating.UpdatedOn.AddDate(0, 0, model.TrustRatingCooldownDays)
		return nil, model.NewBadRequestError(fmt.Sprintf("rating cannot be changed until %s", nextEditable.Format("2006-01-02")))
	}

	// Apply updates
	trustLevel := rating.TrustLevel
	trustReview := rating.TrustReview
	if req.TrustLevel != nil {
		trustLevel = model.TrustLevel(*req.TrustLevel)
	}
	if req.TrustReview != nil {
		trustReview = *req.TrustReview
	}

	updated, err := s.repo.Update(ctx, id, trustLevel, trustReview)
	if err != nil {
		return nil, fmt.Errorf("failed to update rating: %w", err)
	}

	return updated, nil
}

// Delete deletes a trust rating (returns to neutral state)
func (s *TrustRatingService) Delete(ctx context.Context, id string, userID string) error {
	rating, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get rating: %w", err)
	}
	if rating == nil {
		return model.NewNotFoundError("trust rating not found")
	}

	if rating.RaterID != userID {
		return model.NewForbiddenError("not your rating")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete rating: %w", err)
	}

	return nil
}

// GetReceivedRatings retrieves public ratings received by a user
func (s *TrustRatingService) GetReceivedRatings(ctx context.Context, userID string, limit, offset int) ([]*model.TrustRating, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.GetReceivedRatings(ctx, userID, limit, offset)
}

// GetGivenRatings retrieves ratings given by a user
func (s *TrustRatingService) GetGivenRatings(ctx context.Context, userID string, limit, offset int) ([]*model.TrustRating, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.GetGivenRatings(ctx, userID, limit, offset)
}

// GetAggregate retrieves aggregated trust stats for a user
func (s *TrustRatingService) GetAggregate(ctx context.Context, userID string) (*model.TrustAggregate, error) {
	return s.repo.GetAggregate(ctx, userID)
}

// CreateEndorsement creates an endorsement on a trust rating
func (s *TrustRatingService) CreateEndorsement(ctx context.Context, ratingID string, endorserID string, req *model.CreateEndorsementRequest) (*model.TrustEndorsement, error) {
	// Validate request
	if errors := req.Validate(); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	// Get the rating
	rating, err := s.repo.GetByID(ctx, ratingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}
	if rating == nil {
		return nil, model.NewNotFoundError("trust rating not found")
	}

	// Prevent self-endorsement
	if rating.RaterID == endorserID {
		return nil, model.NewBadRequestError("cannot endorse your own rating")
	}

	// Check if already endorsed
	hasEndorsed, err := s.repo.HasEndorsed(ctx, endorserID, ratingID)
	if err != nil {
		return nil, fmt.Errorf("failed to check endorsement: %w", err)
	}
	if hasEndorsed {
		return nil, model.NewConflictError("already endorsed this rating")
	}

	// Verify endorser was also present at the anchor
	canRate, err := s.repo.CanRate(ctx, endorserID, rating.RateeID, string(rating.AnchorType), rating.AnchorID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate endorser eligibility: %w", err)
	}
	if !canRate {
		return nil, model.NewForbiddenError("you must have attended the same event to endorse")
	}

	endorsement := &model.TrustEndorsement{
		TrustRatingID:   ratingID,
		EndorserID:      endorserID,
		EndorsementType: model.EndorsementType(req.EndorsementType),
		Note:            req.Note,
	}

	if err := s.repo.CreateEndorsement(ctx, endorsement); err != nil {
		return nil, fmt.Errorf("failed to create endorsement: %w", err)
	}

	return endorsement, nil
}

// GetEndorsements retrieves endorsements for a rating
func (s *TrustRatingService) GetEndorsements(ctx context.Context, ratingID string) ([]*model.TrustEndorsement, error) {
	return s.repo.GetEndorsementsByRating(ctx, ratingID)
}

// GetDistrustSignals retrieves users with high distrust for admin review
func (s *TrustRatingService) GetDistrustSignals(ctx context.Context, minDistrust int, limit int) ([]*model.DistrustSignal, error) {
	if minDistrust <= 0 {
		minDistrust = 3
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.GetDistrustSignals(ctx, minDistrust, limit)
}

// Helper methods

func (s *TrustRatingService) canEdit(rating *model.TrustRating) bool {
	cooldownEnd := rating.UpdatedOn.AddDate(0, 0, model.TrustRatingCooldownDays)
	return time.Now().After(cooldownEnd)
}
