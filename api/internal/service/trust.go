package service

import (
	"context"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// Error definitions moved to errors.go

// TrustRepositoryInterface defines the repository interface
type TrustRepositoryInterface interface {
	CreateTrustRelation(ctx context.Context, trust *model.TrustRelation) error
	GetTrustRelation(ctx context.Context, userAID, userBID string) (*model.TrustRelation, error)
	UpdateTrustRelation(ctx context.Context, id, status string) error
	DeleteTrustRelation(ctx context.Context, id string) error
	GetUserTrustRelations(ctx context.Context, userID string) ([]*model.TrustRelation, error)
	CheckMutualTrust(ctx context.Context, userAID, userBID string) (bool, error)
	CreateIRLVerification(ctx context.Context, irl *model.IRLVerification) error
	GetIRLVerification(ctx context.Context, userAID, userBID string) (*model.IRLVerification, error)
	UpdateIRLVerification(ctx context.Context, id string, updates map[string]interface{}) error
	CheckIRLConfirmed(ctx context.Context, userAID, userBID string) (bool, error)
	GetUserIRLConnections(ctx context.Context, userID string) ([]*model.IRLVerification, error)
	GetTrustProfile(ctx context.Context, userID string) (*model.UserTrustProfile, error)
}

// TrustService handles trust and IRL verification business logic
type TrustService struct {
	repo TrustRepositoryInterface
}

// NewTrustService creates a new trust service
func NewTrustService(repo TrustRepositoryInterface) *TrustService {
	return &TrustService{repo: repo}
}

// GrantTrust grants trust from user A to user B
func (s *TrustService) GrantTrust(ctx context.Context, fromUserID, toUserID string) (*model.TrustRelation, error) {
	if fromUserID == toUserID {
		return nil, ErrCannotTrustSelf
	}

	// Check if already trusted
	existing, err := s.repo.GetTrustRelation(ctx, fromUserID, toUserID)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.Status == model.TrustStatusActive {
		return nil, ErrAlreadyTrusted
	}

	// If there's a revoked trust, update it
	if existing != nil {
		if err := s.repo.UpdateTrustRelation(ctx, existing.ID, model.TrustStatusActive); err != nil {
			return nil, err
		}
		existing.Status = model.TrustStatusActive
		existing.UpdatedOn = time.Now()
		return existing, nil
	}

	// Create new trust relation
	trust := &model.TrustRelation{
		UserAID: fromUserID,
		UserBID: toUserID,
		Status:  model.TrustStatusActive,
	}

	if err := s.repo.CreateTrustRelation(ctx, trust); err != nil {
		return nil, err
	}

	return trust, nil
}

// RevokeTrust revokes trust from user A to user B
func (s *TrustService) RevokeTrust(ctx context.Context, fromUserID, toUserID string) error {
	if fromUserID == toUserID {
		return ErrCannotTrustSelf
	}

	existing, err := s.repo.GetTrustRelation(ctx, fromUserID, toUserID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrTrustNotFound
	}

	return s.repo.UpdateTrustRelation(ctx, existing.ID, model.TrustStatusRevoked)
}

// CheckTrust checks if user A trusts user B
func (s *TrustService) CheckTrust(ctx context.Context, fromUserID, toUserID string) (bool, error) {
	trust, err := s.repo.GetTrustRelation(ctx, fromUserID, toUserID)
	if err != nil {
		return false, err
	}
	return trust != nil && trust.Status == model.TrustStatusActive, nil
}

// CheckMutualTrust checks if two users mutually trust each other
func (s *TrustService) CheckMutualTrust(ctx context.Context, userAID, userBID string) (bool, error) {
	return s.repo.CheckMutualTrust(ctx, userAID, userBID)
}

// ConfirmIRL confirms an IRL meeting between two users
func (s *TrustService) ConfirmIRL(ctx context.Context, userID string, req *model.ConfirmIRLRequest) (*model.IRLVerification, error) {
	if userID == req.UserID {
		return nil, ErrCannotTrustSelf
	}

	// Validate context
	validContexts := map[string]bool{
		model.IRLContextEvent:      true,
		model.IRLContextHangout:    true,
		model.IRLContextIntroduced: true,
		model.IRLContextOther:      true,
	}
	if !validContexts[req.Context] {
		return nil, ErrInvalidContext
	}

	// Check for existing IRL verification
	existing, err := s.repo.GetIRLVerification(ctx, userID, req.UserID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	if existing != nil {
		// Update existing verification
		updates := make(map[string]interface{})

		// Determine which user is confirming
		if existing.UserAID == userID {
			if !existing.UserAConfirmed {
				updates["user_a_confirmed"] = true
				updates["user_a_confirmed_on"] = now
			}
		} else {
			if !existing.UserBConfirmed {
				updates["user_b_confirmed"] = true
				updates["user_b_confirmed_on"] = now
			}
		}

		if len(updates) > 0 {
			if err := s.repo.UpdateIRLVerification(ctx, existing.ID, updates); err != nil {
				return nil, err
			}
		}

		// Refresh and return
		return s.repo.GetIRLVerification(ctx, userID, req.UserID)
	}

	// Create new IRL verification
	irl := &model.IRLVerification{
		UserAID:          userID,
		UserBID:          req.UserID,
		Context:          req.Context,
		ReferenceID:      req.ReferenceID,
		UserAConfirmed:   true,
		UserBConfirmed:   false,
		UserAConfirmedOn: &now,
	}

	if err := s.repo.CreateIRLVerification(ctx, irl); err != nil {
		return nil, err
	}

	return irl, nil
}

// CheckIRLConfirmed checks if IRL is confirmed by both users
func (s *TrustService) CheckIRLConfirmed(ctx context.Context, userAID, userBID string) (bool, error) {
	return s.repo.CheckIRLConfirmed(ctx, userAID, userBID)
}

// GetTrustSummary returns a summary of trust status between two users
func (s *TrustService) GetTrustSummary(ctx context.Context, userAID, userBID string) (*model.TrustSummary, error) {
	irlConfirmed, err := s.CheckIRLConfirmed(ctx, userAID, userBID)
	if err != nil {
		return nil, err
	}

	mutualTrust, err := s.CheckMutualTrust(ctx, userAID, userBID)
	if err != nil {
		return nil, err
	}

	var trustLevel string
	switch {
	case mutualTrust && irlConfirmed:
		trustLevel = model.TrustLevelTrusted
	case irlConfirmed:
		trustLevel = model.TrustLevelIRL
	default:
		trustLevel = model.TrustLevelNone
	}

	return &model.TrustSummary{
		UserAID:      userAID,
		UserBID:      userBID,
		IRLConfirmed: irlConfirmed,
		MutualTrust:  mutualTrust,
		CanCommute:   irlConfirmed && mutualTrust,
		TrustLevel:   trustLevel,
	}, nil
}

// GetTrustProfile gets trust statistics for a user
func (s *TrustService) GetTrustProfile(ctx context.Context, userID string) (*model.UserTrustProfile, error) {
	return s.repo.GetTrustProfile(ctx, userID)
}

// GetTrustedUsers returns list of users that the given user trusts
func (s *TrustService) GetTrustedUsers(ctx context.Context, userID string) ([]model.TrustedUser, error) {
	relations, err := s.repo.GetUserTrustRelations(ctx, userID)
	if err != nil {
		return nil, err
	}

	users := make([]model.TrustedUser, 0)
	for _, rel := range relations {
		// Determine the other user and direction
		var otherID string
		var direction string

		if rel.UserAID == userID {
			otherID = rel.UserBID
			direction = model.TrustDirectionITrustThem
		} else {
			otherID = rel.UserAID
			direction = model.TrustDirectionTheyTrustMe
		}

		// Check if mutual
		if rel.UserAID == userID {
			// Check if they also trust us
			reverse, _ := s.repo.GetTrustRelation(ctx, rel.UserBID, userID)
			if reverse != nil && reverse.Status == model.TrustStatusActive {
				direction = model.TrustDirectionMutual
			}
		}

		// Check IRL status
		irlConfirmed, _ := s.CheckIRLConfirmed(ctx, userID, otherID)

		users = append(users, model.TrustedUser{
			UserID:       otherID,
			IRLConfirmed: irlConfirmed,
			TrustStatus:  direction,
			Since:        rel.CreatedOn,
		})
	}

	return users, nil
}

// GetIRLConnections returns all confirmed IRL connections
func (s *TrustService) GetIRLConnections(ctx context.Context, userID string) ([]*model.IRLVerification, error) {
	return s.repo.GetUserIRLConnections(ctx, userID)
}

// CanAccessCommuteFeature checks if a user can use commute features
func (s *TrustService) CanAccessCommuteFeature(ctx context.Context, userID string) (bool, error) {
	profile, err := s.repo.GetTrustProfile(ctx, userID)
	if err != nil {
		return false, err
	}
	return profile.CanOfferCommute, nil
}

// ValidateCommuteParticipation checks if two users can share a commute
func (s *TrustService) ValidateCommuteParticipation(ctx context.Context, driverID, passengerID string) error {
	// Check IRL requirement
	irlConfirmed, err := s.CheckIRLConfirmed(ctx, driverID, passengerID)
	if err != nil {
		return err
	}
	if !irlConfirmed {
		return ErrIRLRequired
	}

	// Check mutual trust
	mutualTrust, err := s.CheckMutualTrust(ctx, driverID, passengerID)
	if err != nil {
		return err
	}
	if !mutualTrust {
		return ErrTrustNotEstablished
	}

	return nil
}
