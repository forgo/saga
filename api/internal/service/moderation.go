package service

import (
	"context"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// ModerationRepository defines the interface for moderation data access
type ModerationRepository interface {
	// Reports
	CreateReport(ctx context.Context, report *model.Report) error
	GetReport(ctx context.Context, id string) (*model.Report, error)
	GetReportsByStatus(ctx context.Context, status model.ReportStatus, limit int) ([]*model.Report, error)
	GetReportsAgainstUser(ctx context.Context, userID string) ([]*model.Report, error)
	GetRecentReportsAgainstUser(ctx context.Context, userID string, days int) ([]*model.Report, error)
	UpdateReport(ctx context.Context, id string, updates map[string]interface{}) (*model.Report, error)

	// Actions
	CreateAction(ctx context.Context, action *model.ModerationAction) error
	GetAction(ctx context.Context, id string) (*model.ModerationAction, error)
	GetActiveActionsForUser(ctx context.Context, userID string) ([]*model.ModerationAction, error)
	GetAllActionsForUser(ctx context.Context, userID string) ([]*model.ModerationAction, error)
	UpdateAction(ctx context.Context, id string, updates map[string]interface{}) error
	ExpireOldActions(ctx context.Context) error

	// Blocks
	CreateBlock(ctx context.Context, block *model.Block) error
	GetBlock(ctx context.Context, blockerID, blockedID string) (*model.Block, error)
	GetBlocksByBlocker(ctx context.Context, blockerID string) ([]*model.Block, error)
	IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error)
	IsBlockedEitherWay(ctx context.Context, userID1, userID2 string) (bool, error)
	DeleteBlock(ctx context.Context, blockerID, blockedID string) error

	// Stats
	GetModerationStats(ctx context.Context) (*model.ModerationStats, error)
}

// Error definitions moved to errors.go

// ModerationService handles moderation operations
type ModerationService struct {
	moderationRepo ModerationRepository
	eventHub       *EventHub
}

// NewModerationService creates a new moderation service
func NewModerationService(moderationRepo ModerationRepository, eventHub *EventHub) *ModerationService {
	return &ModerationService{
		moderationRepo: moderationRepo,
		eventHub:       eventHub,
	}
}

// Report operations

// CreateReport creates a new report
func (s *ModerationService) CreateReport(ctx context.Context, reporterUserID string, req *model.CreateReportRequest) (*model.Report, error) {
	// Validate
	if reporterUserID == req.ReportedUserID {
		return nil, ErrCannotReportSelf
	}
	if !model.IsValidReportCategory(req.Category) {
		return nil, ErrInvalidCategory
	}
	if req.Description != nil && len(*req.Description) > model.MaxReportDescriptionLength {
		return nil, ErrDescriptionTooLong
	}

	report := &model.Report{
		ReporterUserID: reporterUserID,
		ReportedUserID: req.ReportedUserID,
		CircleID:       req.CircleID,
		Category:       model.ReportCategory(req.Category),
		Description:    req.Description,
		ContentType:    req.ContentType,
		ContentID:      req.ContentID,
		Status:         model.ReportStatusPending,
	}

	if err := s.moderationRepo.CreateReport(ctx, report); err != nil {
		return nil, err
	}

	// Emit event for moderation queue
	if s.eventHub != nil {
		s.eventHub.Publish(&Event{
			Type: "moderation.report_created",
			Data: map[string]interface{}{
				"report_id":        report.ID,
				"reported_user_id": report.ReportedUserID,
				"category":         report.Category,
			},
		})
	}

	return report, nil
}

// GetReport retrieves a report by ID
func (s *ModerationService) GetReport(ctx context.Context, id string) (*model.Report, error) {
	report, err := s.moderationRepo.GetReport(ctx, id)
	if err != nil {
		return nil, err
	}
	if report == nil {
		return nil, ErrReportNotFound
	}
	return report, nil
}

// GetPendingReports retrieves pending reports
func (s *ModerationService) GetPendingReports(ctx context.Context, limit int) ([]*model.Report, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.moderationRepo.GetReportsByStatus(ctx, model.ReportStatusPending, limit)
}

// ReviewReport reviews a report
func (s *ModerationService) ReviewReport(ctx context.Context, reportID, reviewerID string, req *model.ReviewReportRequest) (*model.Report, error) {
	if !model.IsValidReportStatus(req.Status) {
		return nil, ErrInvalidStatus
	}

	report, err := s.moderationRepo.GetReport(ctx, reportID)
	if err != nil {
		return nil, err
	}
	if report == nil {
		return nil, ErrReportNotFound
	}

	updates := map[string]interface{}{
		"status":         req.Status,
		"reviewed_by_id": reviewerID,
		"reviewed_on":    time.Now(),
	}
	if req.Notes != nil {
		updates["review_notes"] = *req.Notes
	}
	if req.ActionTaken != nil {
		updates["action_taken"] = *req.ActionTaken
	}
	if req.Status == string(model.ReportStatusResolved) {
		updates["resolved_on"] = time.Now()
	}

	return s.moderationRepo.UpdateReport(ctx, reportID, updates)
}

// Moderation Action operations

// TakeAction creates a moderation action against a user
func (s *ModerationService) TakeAction(ctx context.Context, adminUserID string, req *model.CreateModerationActionRequest) (*model.ModerationAction, error) {
	// Validate
	if !model.IsValidModerationLevel(req.Level) {
		return nil, ErrInvalidLevel
	}
	if req.Reason == "" {
		return nil, ErrReasonRequired
	}
	if len(req.Reason) > model.MaxActionReasonLength {
		return nil, ErrDescriptionTooLong
	}

	action := &model.ModerationAction{
		UserID:       req.UserID,
		Level:        model.ModerationLevel(req.Level),
		Reason:       req.Reason,
		ReportID:     req.ReportID,
		AdminUserID:  &adminUserID,
		IsActive:     true,
		Restrictions: req.Restrictions,
	}

	// Set expiration based on level
	switch action.Level {
	case model.ModerationLevelWarning:
		expires := time.Now().AddDate(0, 0, model.WarningDurationDays)
		action.ExpiresOn = &expires
		dur := model.WarningDurationDays
		action.Duration = &dur
	case model.ModerationLevelSuspension:
		days := model.DefaultSuspensionDays
		if req.DurationDays != nil && *req.DurationDays > 0 {
			days = *req.DurationDays
		}
		expires := time.Now().AddDate(0, 0, days)
		action.ExpiresOn = &expires
		action.Duration = &days
	case model.ModerationLevelBan:
		// Bans don't expire
		action.ExpiresOn = nil
		action.Duration = nil
	}

	if err := s.moderationRepo.CreateAction(ctx, action); err != nil {
		return nil, err
	}

	// Emit event
	if s.eventHub != nil {
		s.eventHub.Publish(&Event{
			Type: "moderation.action_taken",
			Data: map[string]interface{}{
				"action_id": action.ID,
				"user_id":   action.UserID,
				"level":     action.Level,
			},
		})
	}

	return action, nil
}

// LiftAction lifts an active moderation action
func (s *ModerationService) LiftAction(ctx context.Context, actionID, adminUserID string, req *model.LiftActionRequest) error {
	action, err := s.moderationRepo.GetAction(ctx, actionID)
	if err != nil {
		return err
	}
	if action == nil {
		return ErrActionNotFound
	}

	if req.Reason == "" {
		return ErrReasonRequired
	}

	updates := map[string]interface{}{
		"is_active":    false,
		"lifted_on":    time.Now(),
		"lifted_by_id": adminUserID,
		"lift_reason":  req.Reason,
	}

	return s.moderationRepo.UpdateAction(ctx, actionID, updates)
}

// GetUserModerationStatus retrieves a user's current moderation standing
func (s *ModerationService) GetUserModerationStatus(ctx context.Context, userID string) (*model.UserModerationStatus, error) {
	// Get active actions
	actionPtrs, err := s.moderationRepo.GetActiveActionsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get report counts
	allReports, err := s.moderationRepo.GetReportsAgainstUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	recentReports, err := s.moderationRepo.GetRecentReportsAgainstUser(ctx, userID, model.RecentReportWindowDays)
	if err != nil {
		return nil, err
	}

	// Convert pointer slice to value slice
	actions := make([]model.ModerationAction, len(actionPtrs))
	for i, a := range actionPtrs {
		actions[i] = *a
	}

	status := &model.UserModerationStatus{
		UserID:            userID,
		ActiveActions:     actions,
		ReportCount:       len(allReports),
		RecentReportCount: len(recentReports),
	}

	// Check for active restrictions
	var restrictions []string
	for _, action := range actions {
		switch action.Level {
		case model.ModerationLevelBan:
			status.IsBanned = true
		case model.ModerationLevelSuspension:
			status.IsSuspended = true
			status.SuspensionEndsOn = action.ExpiresOn
		case model.ModerationLevelWarning:
			status.HasWarning = true
			status.WarningExpiresOn = action.ExpiresOn
		}
		restrictions = append(restrictions, action.Restrictions...)
	}
	status.Restrictions = restrictions

	return status, nil
}

// IsUserRestricted checks if a user can perform an action
func (s *ModerationService) IsUserRestricted(ctx context.Context, userID, restriction string) (bool, error) {
	status, err := s.GetUserModerationStatus(ctx, userID)
	if err != nil {
		return false, err
	}

	if status.IsBanned {
		return true, nil
	}
	if status.IsSuspended {
		return true, nil
	}

	for _, r := range status.Restrictions {
		if r == restriction {
			return true, nil
		}
	}

	return false, nil
}

// CanUserAccess checks if a user can access the platform
func (s *ModerationService) CanUserAccess(ctx context.Context, userID string) error {
	status, err := s.GetUserModerationStatus(ctx, userID)
	if err != nil {
		return err
	}

	if status.IsBanned {
		return ErrUserBanned
	}
	if status.IsSuspended {
		return ErrUserSuspended
	}

	return nil
}

// ExpireOldActions expires old moderation actions (should be run periodically)
func (s *ModerationService) ExpireOldActions(ctx context.Context) error {
	return s.moderationRepo.ExpireOldActions(ctx)
}

// Block operations

// BlockUser blocks another user
func (s *ModerationService) BlockUser(ctx context.Context, blockerUserID string, req *model.CreateBlockRequest) (*model.Block, error) {
	if blockerUserID == req.BlockedUserID {
		return nil, ErrCannotBlockSelf
	}

	// Check if already blocked
	existing, err := s.moderationRepo.GetBlock(ctx, blockerUserID, req.BlockedUserID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyBlocked
	}

	block := &model.Block{
		BlockerUserID: blockerUserID,
		BlockedUserID: req.BlockedUserID,
		Reason:        req.Reason,
	}

	if err := s.moderationRepo.CreateBlock(ctx, block); err != nil {
		return nil, err
	}

	return block, nil
}

// UnblockUser removes a block
func (s *ModerationService) UnblockUser(ctx context.Context, blockerUserID, blockedUserID string) error {
	// Check if blocked
	existing, err := s.moderationRepo.GetBlock(ctx, blockerUserID, blockedUserID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNotBlocked
	}

	return s.moderationRepo.DeleteBlock(ctx, blockerUserID, blockedUserID)
}

// GetBlockedUsers retrieves users blocked by the given user
func (s *ModerationService) GetBlockedUsers(ctx context.Context, userID string) ([]*model.Block, error) {
	return s.moderationRepo.GetBlocksByBlocker(ctx, userID)
}

// IsBlocked checks if one user has blocked another
func (s *ModerationService) IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error) {
	return s.moderationRepo.IsBlocked(ctx, blockerID, blockedID)
}

// IsBlockedEitherWay checks if either user has blocked the other
func (s *ModerationService) IsBlockedEitherWay(ctx context.Context, userID1, userID2 string) (bool, error) {
	return s.moderationRepo.IsBlockedEitherWay(ctx, userID1, userID2)
}

// Stats operations

// GetModerationStats retrieves moderation statistics (admin only)
func (s *ModerationService) GetModerationStats(ctx context.Context) (*model.ModerationStats, error) {
	return s.moderationRepo.GetModerationStats(ctx)
}
