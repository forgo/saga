package service

import (
	"context"
	"fmt"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// AdventureAdmissionRepository defines the interface for adventure admission storage
type AdventureAdmissionRepository interface {
	Create(ctx context.Context, admission *model.AdventureAdmission) error
	GetByID(ctx context.Context, id string) (*model.AdventureAdmission, error)
	GetByAdventureAndUser(ctx context.Context, adventureID, userID string) (*model.AdventureAdmission, error)
	GetByAdventure(ctx context.Context, adventureID string, status *model.AdventureAdmissionStatus, limit, offset int) ([]*model.AdventureAdmission, error)
	GetByUser(ctx context.Context, userID string, status *model.AdventureAdmissionStatus) ([]*model.AdventureAdmission, error)
	GetAdmittedUsers(ctx context.Context, adventureID string) ([]*model.AdventureAdmission, error)
	GetPendingRequests(ctx context.Context, adventureID string) ([]*model.AdventureAdmission, error)
	Update(ctx context.Context, id string, status model.AdventureAdmissionStatus, rejectionReason *string) (*model.AdventureAdmission, error)
	Admit(ctx context.Context, id string) (*model.AdventureAdmission, error)
	Reject(ctx context.Context, id string, reason string) (*model.AdventureAdmission, error)
	Delete(ctx context.Context, id string) error
	IsAdmitted(ctx context.Context, adventureID, userID string) (bool, error)
	CountAdmitted(ctx context.Context, adventureID string) (int, error)
}

// AdventureRepository defines the interface for adventure storage
type AdventureRepository interface {
	GetByID(ctx context.Context, id string) (*model.Adventure, error)
	Create(ctx context.Context, adventure *model.Adventure) error
	Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Adventure, error)
	UpdateOrganizerUser(ctx context.Context, id string, newOrganizerUserID string) (*model.Adventure, error)
	Freeze(ctx context.Context, id string, reason string) (*model.Adventure, error)
	Unfreeze(ctx context.Context, id string) (*model.Adventure, error)
}

// GuildMembershipChecker defines interface for checking guild membership
// This uses GuildRepository which has the IsMember method
type GuildMembershipChecker interface {
	IsMember(ctx context.Context, userID, guildID string) (bool, error)
	// IsAdmin is implemented via member role check
}

// AdventureService handles adventure business logic
type AdventureService struct {
	adventureRepo AdventureRepository
	admissionRepo AdventureAdmissionRepository
	guildRepo     GuildRepository // Uses GuildRepository which has IsMember
}

// AdventureServiceConfig holds configuration for the adventure service
type AdventureServiceConfig struct {
	AdventureRepo AdventureRepository
	AdmissionRepo AdventureAdmissionRepository
	MemberRepo    interface{} // Not used, kept for backwards compatibility
	GuildRepo     GuildRepository
}

// NewAdventureService creates a new adventure service
func NewAdventureService(cfg AdventureServiceConfig) *AdventureService {
	return &AdventureService{
		adventureRepo: cfg.AdventureRepo,
		admissionRepo: cfg.AdmissionRepo,
		guildRepo:     cfg.GuildRepo,
	}
}

// Create creates a new adventure
func (s *AdventureService) Create(ctx context.Context, userID string, req *model.CreateAdventureRequest) (*model.Adventure, error) {
	if errors := req.Validate(); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	// Parse dates
	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return nil, model.NewValidationError([]model.FieldError{{Field: "start_date", Message: "invalid date format, use RFC3339"}})
	}
	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return nil, model.NewValidationError([]model.FieldError{{Field: "end_date", Message: "invalid date format, use RFC3339"}})
	}

	orgType := req.GetOrganizerType()
	adventure := &model.Adventure{
		Title:           req.Title,
		Description:     req.Description,
		StartDate:       startDate,
		EndDate:         endDate,
		OrganizerType:   orgType,
		OrganizerUserID: userID,
		Status:          model.AdventureStatusIdea,
		CreatedByID:     userID,
	}

	// Set organizer ID based on type
	if orgType == model.AdventureOrganizerGuild {
		if req.GuildID == nil {
			return nil, model.NewBadRequestError("guild_id required for guild-organized adventures")
		}
		adventure.OrganizerID = fmt.Sprintf("guild:%s", *req.GuildID)
		adventure.GuildID = req.GuildID

		// Verify user is member of guild
		if s.guildRepo != nil {
			isMember, err := s.guildRepo.IsMember(ctx, userID, *req.GuildID)
			if err != nil {
				return nil, fmt.Errorf("failed to check guild membership: %w", err)
			}
			if !isMember {
				return nil, model.NewForbiddenError("must be guild member to create guild adventure")
			}
		}
	} else {
		adventure.OrganizerID = fmt.Sprintf("user:%s", userID)
	}

	if err := s.adventureRepo.Create(ctx, adventure); err != nil {
		return nil, fmt.Errorf("failed to create adventure: %w", err)
	}

	// Auto-admit the creator
	admission := &model.AdventureAdmission{
		AdventureID: adventure.ID,
		UserID:      userID,
		Status:      model.AdmissionStatusAdmitted,
		RequestedBy: model.AdmissionRequestedBySelf,
	}
	// Create admission (non-fatal error)
	_ = s.admissionRepo.Create(ctx, admission)

	return adventure, nil
}

// GetByID retrieves an adventure by ID
func (s *AdventureService) GetByID(ctx context.Context, id string) (*model.Adventure, error) {
	adventure, err := s.adventureRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get adventure: %w", err)
	}
	if adventure == nil {
		return nil, model.NewNotFoundError("adventure not found")
	}
	return adventure, nil
}

// Admission Operations

// RequestAdmission requests admission to an adventure
func (s *AdventureService) RequestAdmission(ctx context.Context, adventureID string, userID string, req *model.RequestAdmissionRequest) (*model.AdventureAdmission, error) {
	// Check adventure exists
	adventure, err := s.adventureRepo.GetByID(ctx, adventureID)
	if err != nil {
		return nil, fmt.Errorf("failed to get adventure: %w", err)
	}
	if adventure == nil {
		return nil, model.NewNotFoundError("adventure not found")
	}

	// Check if already admitted or pending
	existing, err := s.admissionRepo.GetByAdventureAndUser(ctx, adventureID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing admission: %w", err)
	}
	if existing != nil {
		if existing.Status == model.AdmissionStatusAdmitted {
			return nil, model.NewConflictError("already admitted")
		}
		if existing.Status == model.AdmissionStatusRequested {
			return nil, model.NewConflictError("request already pending")
		}
	}

	admission := &model.AdventureAdmission{
		AdventureID: adventureID,
		UserID:      userID,
		Status:      model.AdmissionStatusRequested,
		RequestedBy: model.AdmissionRequestedBySelf,
	}

	if err := s.admissionRepo.Create(ctx, admission); err != nil {
		return nil, fmt.Errorf("failed to create admission request: %w", err)
	}

	return admission, nil
}

// GetAdmission gets admission status for a user
func (s *AdventureService) GetAdmission(ctx context.Context, adventureID string, userID string) (*model.AdventureAdmission, error) {
	admission, err := s.admissionRepo.GetByAdventureAndUser(ctx, adventureID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get admission: %w", err)
	}
	if admission == nil {
		return nil, model.NewNotFoundError("no admission found")
	}
	return admission, nil
}

// WithdrawAdmission withdraws an admission request
func (s *AdventureService) WithdrawAdmission(ctx context.Context, adventureID string, userID string) error {
	admission, err := s.admissionRepo.GetByAdventureAndUser(ctx, adventureID, userID)
	if err != nil {
		return fmt.Errorf("failed to get admission: %w", err)
	}
	if admission == nil {
		return model.NewNotFoundError("no admission found")
	}

	if admission.Status == model.AdmissionStatusAdmitted {
		return model.NewBadRequestError("cannot withdraw after being admitted")
	}

	return s.admissionRepo.Delete(ctx, admission.ID)
}

// GetAdmissions gets all admissions for an adventure (organizer only)
func (s *AdventureService) GetAdmissions(ctx context.Context, adventureID string, userID string, status *model.AdventureAdmissionStatus, limit, offset int) ([]*model.AdventureAdmission, error) {
	// Check permission
	adventure, err := s.adventureRepo.GetByID(ctx, adventureID)
	if err != nil {
		return nil, fmt.Errorf("failed to get adventure: %w", err)
	}
	if adventure == nil {
		return nil, model.NewNotFoundError("adventure not found")
	}

	if err := s.checkOrganizerPermission(ctx, adventure, userID); err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	return s.admissionRepo.GetByAdventure(ctx, adventureID, status, limit, offset)
}

// GetPendingAdmissions gets pending admission requests
func (s *AdventureService) GetPendingAdmissions(ctx context.Context, adventureID string, userID string) ([]*model.AdventureAdmission, error) {
	adventure, err := s.adventureRepo.GetByID(ctx, adventureID)
	if err != nil {
		return nil, fmt.Errorf("failed to get adventure: %w", err)
	}
	if adventure == nil {
		return nil, model.NewNotFoundError("adventure not found")
	}

	if err := s.checkOrganizerPermission(ctx, adventure, userID); err != nil {
		return nil, err
	}

	return s.admissionRepo.GetPendingRequests(ctx, adventureID)
}

// RespondToAdmission responds to an admission request (organizer only)
func (s *AdventureService) RespondToAdmission(ctx context.Context, adventureID string, userID string, targetUserID string, req *model.RespondToAdmissionRequest) (*model.AdventureAdmission, error) {
	if errors := req.Validate(); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	adventure, err := s.adventureRepo.GetByID(ctx, adventureID)
	if err != nil {
		return nil, fmt.Errorf("failed to get adventure: %w", err)
	}
	if adventure == nil {
		return nil, model.NewNotFoundError("adventure not found")
	}

	if err := s.checkOrganizerPermission(ctx, adventure, userID); err != nil {
		return nil, err
	}

	admission, err := s.admissionRepo.GetByAdventureAndUser(ctx, adventureID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get admission: %w", err)
	}
	if admission == nil {
		return nil, model.NewNotFoundError("admission request not found")
	}

	if admission.Status != model.AdmissionStatusRequested {
		return nil, model.NewBadRequestError("admission already processed")
	}

	if req.Admit {
		return s.admissionRepo.Admit(ctx, admission.ID)
	}

	reason := ""
	if req.RejectionReason != nil {
		reason = *req.RejectionReason
	}
	return s.admissionRepo.Reject(ctx, admission.ID, reason)
}

// InviteToAdventure invites a user to an adventure
func (s *AdventureService) InviteToAdventure(ctx context.Context, adventureID string, userID string, req *model.InviteToAdventureRequest) (*model.AdventureAdmission, error) {
	if errors := req.Validate(); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	adventure, err := s.adventureRepo.GetByID(ctx, adventureID)
	if err != nil {
		return nil, fmt.Errorf("failed to get adventure: %w", err)
	}
	if adventure == nil {
		return nil, model.NewNotFoundError("adventure not found")
	}

	if err := s.checkOrganizerPermission(ctx, adventure, userID); err != nil {
		return nil, err
	}

	// Check if already exists
	existing, err := s.admissionRepo.GetByAdventureAndUser(ctx, adventureID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing admission: %w", err)
	}
	if existing != nil {
		return nil, model.NewConflictError("user already has admission record")
	}

	admission := &model.AdventureAdmission{
		AdventureID: adventureID,
		UserID:      req.UserID,
		Status:      model.AdmissionStatusAdmitted, // Auto-admit invites
		RequestedBy: model.AdmissionRequestedByInvited,
		InvitedByID: &userID,
	}

	if err := s.admissionRepo.Create(ctx, admission); err != nil {
		return nil, fmt.Errorf("failed to create admission: %w", err)
	}

	return admission, nil
}

// IsAdmitted checks if a user is admitted to an adventure
func (s *AdventureService) IsAdmitted(ctx context.Context, adventureID, userID string) (bool, error) {
	return s.admissionRepo.IsAdmitted(ctx, adventureID, userID)
}

// Organizer Operations

// TransferAdventure transfers organizer role to another user
func (s *AdventureService) TransferAdventure(ctx context.Context, adventureID string, userID string, req *model.TransferAdventureRequest) (*model.Adventure, error) {
	if errors := req.Validate(); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	adventure, err := s.adventureRepo.GetByID(ctx, adventureID)
	if err != nil {
		return nil, fmt.Errorf("failed to get adventure: %w", err)
	}
	if adventure == nil {
		return nil, model.NewNotFoundError("adventure not found")
	}

	if err := s.checkOrganizerPermission(ctx, adventure, userID); err != nil {
		return nil, err
	}

	// For guild adventures, new organizer must be guild member
	if adventure.IsGuildOrganized() && s.guildRepo != nil {
		guildID := adventure.OrganizerID[6:] // Remove "guild:" prefix
		isMember, err := s.guildRepo.IsMember(ctx, req.NewOrganizerUserID, guildID)
		if err != nil {
			return nil, fmt.Errorf("failed to check membership: %w", err)
		}
		if !isMember {
			return nil, model.NewBadRequestError("new organizer must be guild member")
		}
	}

	return s.adventureRepo.UpdateOrganizerUser(ctx, adventureID, req.NewOrganizerUserID)
}

// UnfreezeAdventure unfreezes a frozen adventure
func (s *AdventureService) UnfreezeAdventure(ctx context.Context, adventureID string, userID string, req *model.UnfreezeAdventureRequest) (*model.Adventure, error) {
	adventure, err := s.adventureRepo.GetByID(ctx, adventureID)
	if err != nil {
		return nil, fmt.Errorf("failed to get adventure: %w", err)
	}
	if adventure == nil {
		return nil, model.NewNotFoundError("adventure not found")
	}

	if !adventure.IsFrozen() {
		return nil, model.NewBadRequestError("adventure is not frozen")
	}

	// Must be guild admin for guild adventures
	if adventure.IsGuildOrganized() && s.guildRepo != nil {
		guildID := adventure.OrganizerID[6:]
		isAdmin, err := s.guildRepo.IsGuildAdmin(ctx, userID, guildID)
		if err != nil {
			return nil, fmt.Errorf("failed to check admin status: %w", err)
		}
		if !isAdmin {
			return nil, model.NewForbiddenError("must be guild admin to unfreeze")
		}
	}

	// If new organizer specified, update it
	if req.NewOrganizerUserID != "" {
		if adventure.IsGuildOrganized() && s.guildRepo != nil {
			guildID := adventure.OrganizerID[6:]
			isMember, err := s.guildRepo.IsMember(ctx, req.NewOrganizerUserID, guildID)
			if err != nil {
				return nil, fmt.Errorf("failed to check membership: %w", err)
			}
			if !isMember {
				return nil, model.NewBadRequestError("new organizer must be guild member")
			}
		}
		if _, err := s.adventureRepo.UpdateOrganizerUser(ctx, adventureID, req.NewOrganizerUserID); err != nil {
			return nil, fmt.Errorf("failed to update organizer: %w", err)
		}
	}

	return s.adventureRepo.Unfreeze(ctx, adventureID)
}

// Helper methods

func (s *AdventureService) checkOrganizerPermission(ctx context.Context, adventure *model.Adventure, userID string) error {
	// Current organizer user always has permission
	if adventure.OrganizerUserID == userID {
		return nil
	}

	// For guild adventures, guild admins have permission
	if adventure.IsGuildOrganized() && s.guildRepo != nil {
		guildID := adventure.OrganizerID[6:] // Remove "guild:" prefix
		isAdmin, err := s.guildRepo.IsGuildAdmin(ctx, userID, guildID)
		if err != nil {
			return fmt.Errorf("failed to check admin status: %w", err)
		}
		if isAdmin {
			return nil
		}
	}

	return model.NewForbiddenError("not authorized to manage this adventure")
}
