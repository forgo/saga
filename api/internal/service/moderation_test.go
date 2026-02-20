package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// ============================================================================
// Mock Repository
// ============================================================================

type mockModerationRepo struct {
	createReportFunc          func(ctx context.Context, report *model.Report) error
	getReportFunc             func(ctx context.Context, id string) (*model.Report, error)
	getReportsByStatusFunc    func(ctx context.Context, status model.ReportStatus, limit int) ([]*model.Report, error)
	getReportsAgainstUserFunc func(ctx context.Context, userID string) ([]*model.Report, error)
	getRecentReportsFunc      func(ctx context.Context, userID string, days int) ([]*model.Report, error)
	updateReportFunc          func(ctx context.Context, id string, updates map[string]interface{}) (*model.Report, error)
	createActionFunc          func(ctx context.Context, action *model.ModerationAction) error
	getActionFunc             func(ctx context.Context, id string) (*model.ModerationAction, error)
	getActiveActionsFunc      func(ctx context.Context, userID string) ([]*model.ModerationAction, error)
	getAllActionsFunc         func(ctx context.Context, userID string) ([]*model.ModerationAction, error)
	updateActionFunc          func(ctx context.Context, id string, updates map[string]interface{}) error
	expireOldActionsFunc      func(ctx context.Context) error
	createBlockFunc           func(ctx context.Context, block *model.Block) error
	getBlockFunc              func(ctx context.Context, blockerID, blockedID string) (*model.Block, error)
	getBlocksByBlockerFunc    func(ctx context.Context, blockerID string) ([]*model.Block, error)
	isBlockedFunc             func(ctx context.Context, blockerID, blockedID string) (bool, error)
	isBlockedEitherWayFunc    func(ctx context.Context, userID1, userID2 string) (bool, error)
	deleteBlockFunc           func(ctx context.Context, blockerID, blockedID string) error
	getModerationStatsFunc    func(ctx context.Context) (*model.ModerationStats, error)
}

func (m *mockModerationRepo) CreateReport(ctx context.Context, report *model.Report) error {
	if m.createReportFunc != nil {
		return m.createReportFunc(ctx, report)
	}
	report.ID = "report:1"
	return nil
}

func (m *mockModerationRepo) GetReport(ctx context.Context, id string) (*model.Report, error) {
	if m.getReportFunc != nil {
		return m.getReportFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockModerationRepo) GetReportsByStatus(ctx context.Context, status model.ReportStatus, limit int) ([]*model.Report, error) {
	if m.getReportsByStatusFunc != nil {
		return m.getReportsByStatusFunc(ctx, status, limit)
	}
	return nil, nil
}

func (m *mockModerationRepo) GetReportsAgainstUser(ctx context.Context, userID string) ([]*model.Report, error) {
	if m.getReportsAgainstUserFunc != nil {
		return m.getReportsAgainstUserFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockModerationRepo) GetRecentReportsAgainstUser(ctx context.Context, userID string, days int) ([]*model.Report, error) {
	if m.getRecentReportsFunc != nil {
		return m.getRecentReportsFunc(ctx, userID, days)
	}
	return nil, nil
}

func (m *mockModerationRepo) UpdateReport(ctx context.Context, id string, updates map[string]interface{}) (*model.Report, error) {
	if m.updateReportFunc != nil {
		return m.updateReportFunc(ctx, id, updates)
	}
	return &model.Report{ID: id}, nil
}

func (m *mockModerationRepo) CreateAction(ctx context.Context, action *model.ModerationAction) error {
	if m.createActionFunc != nil {
		return m.createActionFunc(ctx, action)
	}
	action.ID = "action:1"
	return nil
}

func (m *mockModerationRepo) GetAction(ctx context.Context, id string) (*model.ModerationAction, error) {
	if m.getActionFunc != nil {
		return m.getActionFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockModerationRepo) GetActiveActionsForUser(ctx context.Context, userID string) ([]*model.ModerationAction, error) {
	if m.getActiveActionsFunc != nil {
		return m.getActiveActionsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockModerationRepo) GetAllActionsForUser(ctx context.Context, userID string) ([]*model.ModerationAction, error) {
	if m.getAllActionsFunc != nil {
		return m.getAllActionsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockModerationRepo) UpdateAction(ctx context.Context, id string, updates map[string]interface{}) error {
	if m.updateActionFunc != nil {
		return m.updateActionFunc(ctx, id, updates)
	}
	return nil
}

func (m *mockModerationRepo) ExpireOldActions(ctx context.Context) error {
	if m.expireOldActionsFunc != nil {
		return m.expireOldActionsFunc(ctx)
	}
	return nil
}

func (m *mockModerationRepo) CreateBlock(ctx context.Context, block *model.Block) error {
	if m.createBlockFunc != nil {
		return m.createBlockFunc(ctx, block)
	}
	block.ID = "block:1"
	return nil
}

func (m *mockModerationRepo) GetBlock(ctx context.Context, blockerID, blockedID string) (*model.Block, error) {
	if m.getBlockFunc != nil {
		return m.getBlockFunc(ctx, blockerID, blockedID)
	}
	return nil, nil
}

func (m *mockModerationRepo) GetBlocksByBlocker(ctx context.Context, blockerID string) ([]*model.Block, error) {
	if m.getBlocksByBlockerFunc != nil {
		return m.getBlocksByBlockerFunc(ctx, blockerID)
	}
	return nil, nil
}

func (m *mockModerationRepo) IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error) {
	if m.isBlockedFunc != nil {
		return m.isBlockedFunc(ctx, blockerID, blockedID)
	}
	return false, nil
}

func (m *mockModerationRepo) IsBlockedEitherWay(ctx context.Context, userID1, userID2 string) (bool, error) {
	if m.isBlockedEitherWayFunc != nil {
		return m.isBlockedEitherWayFunc(ctx, userID1, userID2)
	}
	return false, nil
}

func (m *mockModerationRepo) DeleteBlock(ctx context.Context, blockerID, blockedID string) error {
	if m.deleteBlockFunc != nil {
		return m.deleteBlockFunc(ctx, blockerID, blockedID)
	}
	return nil
}

func (m *mockModerationRepo) GetModerationStats(ctx context.Context) (*model.ModerationStats, error) {
	if m.getModerationStatsFunc != nil {
		return m.getModerationStatsFunc(ctx)
	}
	return &model.ModerationStats{}, nil
}

// ============================================================================
// CreateReport Tests
// ============================================================================

func TestCreateReport_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{}
	svc := NewModerationService(repo, nil)

	req := &model.CreateReportRequest{
		ReportedUserID: "user:other",
		Category:       "spam",
	}

	report, err := svc.CreateReport(ctx, "user:me", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.ID != "report:1" {
		t.Errorf("expected report ID, got %s", report.ID)
	}
	if report.Status != model.ReportStatusPending {
		t.Errorf("expected pending status, got %s", report.Status)
	}
}

func TestCreateReport_CannotReportSelf(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{}
	svc := NewModerationService(repo, nil)

	req := &model.CreateReportRequest{
		ReportedUserID: "user:me",
		Category:       "spam",
	}

	_, err := svc.CreateReport(ctx, "user:me", req)
	if !errors.Is(err, ErrCannotReportSelf) {
		t.Errorf("expected ErrCannotReportSelf, got %v", err)
	}
}

func TestCreateReport_InvalidCategory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{}
	svc := NewModerationService(repo, nil)

	req := &model.CreateReportRequest{
		ReportedUserID: "user:other",
		Category:       "invalid_category",
	}

	_, err := svc.CreateReport(ctx, "user:me", req)
	if !errors.Is(err, ErrInvalidCategory) {
		t.Errorf("expected ErrInvalidCategory, got %v", err)
	}
}

func TestCreateReport_DescriptionTooLong(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{}
	svc := NewModerationService(repo, nil)

	longDesc := make([]byte, model.MaxReportDescriptionLength+1)
	for i := range longDesc {
		longDesc[i] = 'a'
	}
	longDescStr := string(longDesc)

	req := &model.CreateReportRequest{
		ReportedUserID: "user:other",
		Category:       "spam",
		Description:    &longDescStr,
	}

	_, err := svc.CreateReport(ctx, "user:me", req)
	if !errors.Is(err, ErrDescriptionTooLong) {
		t.Errorf("expected ErrDescriptionTooLong, got %v", err)
	}
}

func TestCreateReport_AllValidCategories(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	categories := []string{"spam", "harassment", "hate_speech", "inappropriate_content", "made_uncomfortable", "other"}

	for _, cat := range categories {
		repo := &mockModerationRepo{}
		svc := NewModerationService(repo, nil)

		req := &model.CreateReportRequest{
			ReportedUserID: "user:other",
			Category:       cat,
		}

		report, err := svc.CreateReport(ctx, "user:me", req)
		if err != nil {
			t.Errorf("category %s: unexpected error: %v", cat, err)
		}
		if report == nil {
			t.Errorf("category %s: expected report", cat)
		}
	}
}

// ============================================================================
// GetReport Tests
// ============================================================================

func TestGetReport_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getReportFunc: func(ctx context.Context, id string) (*model.Report, error) {
			return &model.Report{ID: id, Status: model.ReportStatusPending}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	report, err := svc.GetReport(ctx, "report:123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.ID != "report:123" {
		t.Errorf("expected report ID, got %s", report.ID)
	}
}

func TestGetReport_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getReportFunc: func(ctx context.Context, id string) (*model.Report, error) {
			return nil, nil
		},
	}
	svc := NewModerationService(repo, nil)

	_, err := svc.GetReport(ctx, "report:nonexistent")
	if !errors.Is(err, ErrReportNotFound) {
		t.Errorf("expected ErrReportNotFound, got %v", err)
	}
}

// ============================================================================
// GetPendingReports Tests
// ============================================================================

func TestGetPendingReports_DefaultLimit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var capturedLimit int
	repo := &mockModerationRepo{
		getReportsByStatusFunc: func(ctx context.Context, status model.ReportStatus, limit int) ([]*model.Report, error) {
			capturedLimit = limit
			return []*model.Report{}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	_, err := svc.GetPendingReports(ctx, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedLimit != 50 {
		t.Errorf("expected default limit 50, got %d", capturedLimit)
	}
}

func TestGetPendingReports_CustomLimit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var capturedLimit int
	repo := &mockModerationRepo{
		getReportsByStatusFunc: func(ctx context.Context, status model.ReportStatus, limit int) ([]*model.Report, error) {
			capturedLimit = limit
			return []*model.Report{}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	_, err := svc.GetPendingReports(ctx, 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedLimit != 25 {
		t.Errorf("expected limit 25, got %d", capturedLimit)
	}
}

// ============================================================================
// ReviewReport Tests
// ============================================================================

func TestReviewReport_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getReportFunc: func(ctx context.Context, id string) (*model.Report, error) {
			return &model.Report{ID: id, Status: model.ReportStatusPending}, nil
		},
		updateReportFunc: func(ctx context.Context, id string, updates map[string]interface{}) (*model.Report, error) {
			return &model.Report{ID: id, Status: model.ReportStatus(updates["status"].(string))}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	req := &model.ReviewReportRequest{Status: "resolved"}
	report, err := svc.ReviewReport(ctx, "report:1", "admin:1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Status != model.ReportStatusResolved {
		t.Errorf("expected resolved status, got %s", report.Status)
	}
}

func TestReviewReport_InvalidStatus(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{}
	svc := NewModerationService(repo, nil)

	req := &model.ReviewReportRequest{Status: "invalid"}
	_, err := svc.ReviewReport(ctx, "report:1", "admin:1", req)
	if !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestReviewReport_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getReportFunc: func(ctx context.Context, id string) (*model.Report, error) {
			return nil, nil
		},
	}
	svc := NewModerationService(repo, nil)

	req := &model.ReviewReportRequest{Status: "resolved"}
	_, err := svc.ReviewReport(ctx, "report:1", "admin:1", req)
	if !errors.Is(err, ErrReportNotFound) {
		t.Errorf("expected ErrReportNotFound, got %v", err)
	}
}

// ============================================================================
// TakeAction Tests
// ============================================================================

func TestTakeAction_Warning(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{}
	svc := NewModerationService(repo, nil)

	req := &model.CreateModerationActionRequest{
		UserID: "user:1",
		Level:  "warning",
		Reason: "First offense",
	}

	action, err := svc.TakeAction(ctx, "admin:1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action.Level != model.ModerationLevelWarning {
		t.Errorf("expected warning level, got %s", action.Level)
	}
	if action.ExpiresOn == nil {
		t.Error("expected expiration for warning")
	}
	if action.Duration == nil || *action.Duration != model.WarningDurationDays {
		t.Errorf("expected duration %d, got %v", model.WarningDurationDays, action.Duration)
	}
}

func TestTakeAction_Suspension_DefaultDuration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{}
	svc := NewModerationService(repo, nil)

	req := &model.CreateModerationActionRequest{
		UserID: "user:1",
		Level:  "suspension",
		Reason: "Repeated violations",
	}

	action, err := svc.TakeAction(ctx, "admin:1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action.Level != model.ModerationLevelSuspension {
		t.Errorf("expected suspension level, got %s", action.Level)
	}
	if action.Duration == nil || *action.Duration != model.DefaultSuspensionDays {
		t.Errorf("expected default suspension duration %d, got %v", model.DefaultSuspensionDays, action.Duration)
	}
}

func TestTakeAction_Suspension_CustomDuration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{}
	svc := NewModerationService(repo, nil)

	customDays := 14
	req := &model.CreateModerationActionRequest{
		UserID:       "user:1",
		Level:        "suspension",
		Reason:       "Repeated violations",
		DurationDays: &customDays,
	}

	action, err := svc.TakeAction(ctx, "admin:1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action.Duration == nil || *action.Duration != customDays {
		t.Errorf("expected custom duration %d, got %v", customDays, action.Duration)
	}
}

func TestTakeAction_Ban_NoExpiration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{}
	svc := NewModerationService(repo, nil)

	req := &model.CreateModerationActionRequest{
		UserID: "user:1",
		Level:  "ban",
		Reason: "Serious violation",
	}

	action, err := svc.TakeAction(ctx, "admin:1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action.Level != model.ModerationLevelBan {
		t.Errorf("expected ban level, got %s", action.Level)
	}
	if action.ExpiresOn != nil {
		t.Error("expected no expiration for ban")
	}
	if action.Duration != nil {
		t.Error("expected no duration for ban")
	}
}

func TestTakeAction_InvalidLevel(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{}
	svc := NewModerationService(repo, nil)

	req := &model.CreateModerationActionRequest{
		UserID: "user:1",
		Level:  "invalid",
		Reason: "Test",
	}

	_, err := svc.TakeAction(ctx, "admin:1", req)
	if !errors.Is(err, ErrInvalidLevel) {
		t.Errorf("expected ErrInvalidLevel, got %v", err)
	}
}

func TestTakeAction_MissingReason(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{}
	svc := NewModerationService(repo, nil)

	req := &model.CreateModerationActionRequest{
		UserID: "user:1",
		Level:  "warning",
		Reason: "",
	}

	_, err := svc.TakeAction(ctx, "admin:1", req)
	if !errors.Is(err, ErrReasonRequired) {
		t.Errorf("expected ErrReasonRequired, got %v", err)
	}
}

func TestTakeAction_ReasonTooLong(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{}
	svc := NewModerationService(repo, nil)

	longReason := make([]byte, model.MaxActionReasonLength+1)
	for i := range longReason {
		longReason[i] = 'a'
	}

	req := &model.CreateModerationActionRequest{
		UserID: "user:1",
		Level:  "warning",
		Reason: string(longReason),
	}

	_, err := svc.TakeAction(ctx, "admin:1", req)
	if !errors.Is(err, ErrDescriptionTooLong) {
		t.Errorf("expected ErrDescriptionTooLong, got %v", err)
	}
}

// ============================================================================
// LiftAction Tests
// ============================================================================

func TestLiftAction_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getActionFunc: func(ctx context.Context, id string) (*model.ModerationAction, error) {
			return &model.ModerationAction{ID: id, IsActive: true}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	req := &model.LiftActionRequest{Reason: "Appeal granted"}
	err := svc.LiftAction(ctx, "action:1", "admin:1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLiftAction_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getActionFunc: func(ctx context.Context, id string) (*model.ModerationAction, error) {
			return nil, nil
		},
	}
	svc := NewModerationService(repo, nil)

	req := &model.LiftActionRequest{Reason: "Appeal granted"}
	err := svc.LiftAction(ctx, "action:1", "admin:1", req)
	if !errors.Is(err, ErrActionNotFound) {
		t.Errorf("expected ErrActionNotFound, got %v", err)
	}
}

func TestLiftAction_MissingReason(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getActionFunc: func(ctx context.Context, id string) (*model.ModerationAction, error) {
			return &model.ModerationAction{ID: id, IsActive: true}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	req := &model.LiftActionRequest{Reason: ""}
	err := svc.LiftAction(ctx, "action:1", "admin:1", req)
	if !errors.Is(err, ErrReasonRequired) {
		t.Errorf("expected ErrReasonRequired, got %v", err)
	}
}

// ============================================================================
// GetUserModerationStatus Tests
// ============================================================================

func TestGetUserModerationStatus_Clean(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getActiveActionsFunc: func(ctx context.Context, userID string) ([]*model.ModerationAction, error) {
			return []*model.ModerationAction{}, nil
		},
		getReportsAgainstUserFunc: func(ctx context.Context, userID string) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
		getRecentReportsFunc: func(ctx context.Context, userID string, days int) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	status, err := svc.GetUserModerationStatus(ctx, "user:1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.IsBanned {
		t.Error("expected not banned")
	}
	if status.IsSuspended {
		t.Error("expected not suspended")
	}
	if status.HasWarning {
		t.Error("expected no warning")
	}
}

func TestGetUserModerationStatus_Banned(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getActiveActionsFunc: func(ctx context.Context, userID string) ([]*model.ModerationAction, error) {
			return []*model.ModerationAction{
				{ID: "action:1", Level: model.ModerationLevelBan, IsActive: true},
			}, nil
		},
		getReportsAgainstUserFunc: func(ctx context.Context, userID string) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
		getRecentReportsFunc: func(ctx context.Context, userID string, days int) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	status, err := svc.GetUserModerationStatus(ctx, "user:1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.IsBanned {
		t.Error("expected banned")
	}
}

func TestGetUserModerationStatus_Suspended(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	expires := time.Now().Add(24 * time.Hour)
	repo := &mockModerationRepo{
		getActiveActionsFunc: func(ctx context.Context, userID string) ([]*model.ModerationAction, error) {
			return []*model.ModerationAction{
				{ID: "action:1", Level: model.ModerationLevelSuspension, IsActive: true, ExpiresOn: &expires},
			}, nil
		},
		getReportsAgainstUserFunc: func(ctx context.Context, userID string) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
		getRecentReportsFunc: func(ctx context.Context, userID string, days int) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	status, err := svc.GetUserModerationStatus(ctx, "user:1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.IsSuspended {
		t.Error("expected suspended")
	}
	if status.SuspensionEndsOn == nil {
		t.Error("expected suspension end time")
	}
}

func TestGetUserModerationStatus_WithWarning(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	expires := time.Now().Add(7 * 24 * time.Hour)
	repo := &mockModerationRepo{
		getActiveActionsFunc: func(ctx context.Context, userID string) ([]*model.ModerationAction, error) {
			return []*model.ModerationAction{
				{ID: "action:1", Level: model.ModerationLevelWarning, IsActive: true, ExpiresOn: &expires},
			}, nil
		},
		getReportsAgainstUserFunc: func(ctx context.Context, userID string) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
		getRecentReportsFunc: func(ctx context.Context, userID string, days int) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	status, err := svc.GetUserModerationStatus(ctx, "user:1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.HasWarning {
		t.Error("expected has warning")
	}
	if status.WarningExpiresOn == nil {
		t.Error("expected warning expiration")
	}
}

func TestGetUserModerationStatus_WithRestrictions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getActiveActionsFunc: func(ctx context.Context, userID string) ([]*model.ModerationAction, error) {
			return []*model.ModerationAction{
				{
					ID:           "action:1",
					Level:        model.ModerationLevelWarning,
					IsActive:     true,
					Restrictions: []string{"create_public_events", "send_messages"},
				},
			}, nil
		},
		getReportsAgainstUserFunc: func(ctx context.Context, userID string) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
		getRecentReportsFunc: func(ctx context.Context, userID string, days int) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	status, err := svc.GetUserModerationStatus(ctx, "user:1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(status.Restrictions) != 2 {
		t.Errorf("expected 2 restrictions, got %d", len(status.Restrictions))
	}
}

// ============================================================================
// IsUserRestricted Tests
// ============================================================================

func TestIsUserRestricted_NotRestricted(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getActiveActionsFunc: func(ctx context.Context, userID string) ([]*model.ModerationAction, error) {
			return []*model.ModerationAction{}, nil
		},
		getReportsAgainstUserFunc: func(ctx context.Context, userID string) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
		getRecentReportsFunc: func(ctx context.Context, userID string, days int) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	restricted, err := svc.IsUserRestricted(ctx, "user:1", "create_public_events")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if restricted {
		t.Error("expected not restricted")
	}
}

func TestIsUserRestricted_Banned(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getActiveActionsFunc: func(ctx context.Context, userID string) ([]*model.ModerationAction, error) {
			return []*model.ModerationAction{
				{ID: "action:1", Level: model.ModerationLevelBan, IsActive: true},
			}, nil
		},
		getReportsAgainstUserFunc: func(ctx context.Context, userID string) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
		getRecentReportsFunc: func(ctx context.Context, userID string, days int) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	restricted, err := svc.IsUserRestricted(ctx, "user:1", "anything")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !restricted {
		t.Error("expected restricted (banned)")
	}
}

func TestIsUserRestricted_SpecificRestriction(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getActiveActionsFunc: func(ctx context.Context, userID string) ([]*model.ModerationAction, error) {
			return []*model.ModerationAction{
				{
					ID:           "action:1",
					Level:        model.ModerationLevelWarning,
					IsActive:     true,
					Restrictions: []string{"create_public_events"},
				},
			}, nil
		},
		getReportsAgainstUserFunc: func(ctx context.Context, userID string) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
		getRecentReportsFunc: func(ctx context.Context, userID string, days int) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	// Should be restricted for create_public_events
	restricted, err := svc.IsUserRestricted(ctx, "user:1", "create_public_events")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !restricted {
		t.Error("expected restricted for create_public_events")
	}

	// Should not be restricted for other actions
	restricted, err = svc.IsUserRestricted(ctx, "user:1", "send_messages")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if restricted {
		t.Error("expected not restricted for send_messages")
	}
}

// ============================================================================
// CanUserAccess Tests
// ============================================================================

func TestCanUserAccess_Allowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getActiveActionsFunc: func(ctx context.Context, userID string) ([]*model.ModerationAction, error) {
			return []*model.ModerationAction{}, nil
		},
		getReportsAgainstUserFunc: func(ctx context.Context, userID string) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
		getRecentReportsFunc: func(ctx context.Context, userID string, days int) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	err := svc.CanUserAccess(ctx, "user:1")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCanUserAccess_Banned(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getActiveActionsFunc: func(ctx context.Context, userID string) ([]*model.ModerationAction, error) {
			return []*model.ModerationAction{
				{ID: "action:1", Level: model.ModerationLevelBan, IsActive: true},
			}, nil
		},
		getReportsAgainstUserFunc: func(ctx context.Context, userID string) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
		getRecentReportsFunc: func(ctx context.Context, userID string, days int) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	err := svc.CanUserAccess(ctx, "user:1")
	if !errors.Is(err, ErrUserBanned) {
		t.Errorf("expected ErrUserBanned, got %v", err)
	}
}

func TestCanUserAccess_Suspended(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	expires := time.Now().Add(24 * time.Hour)
	repo := &mockModerationRepo{
		getActiveActionsFunc: func(ctx context.Context, userID string) ([]*model.ModerationAction, error) {
			return []*model.ModerationAction{
				{ID: "action:1", Level: model.ModerationLevelSuspension, IsActive: true, ExpiresOn: &expires},
			}, nil
		},
		getReportsAgainstUserFunc: func(ctx context.Context, userID string) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
		getRecentReportsFunc: func(ctx context.Context, userID string, days int) ([]*model.Report, error) {
			return []*model.Report{}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	err := svc.CanUserAccess(ctx, "user:1")
	if !errors.Is(err, ErrUserSuspended) {
		t.Errorf("expected ErrUserSuspended, got %v", err)
	}
}

// ============================================================================
// Block Tests
// ============================================================================

func TestBlockUser_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{}
	svc := NewModerationService(repo, nil)

	req := &model.CreateBlockRequest{BlockedUserID: "user:other"}
	block, err := svc.BlockUser(ctx, "user:me", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if block.ID != "block:1" {
		t.Errorf("expected block ID, got %s", block.ID)
	}
}

func TestBlockUser_CannotBlockSelf(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{}
	svc := NewModerationService(repo, nil)

	req := &model.CreateBlockRequest{BlockedUserID: "user:me"}
	_, err := svc.BlockUser(ctx, "user:me", req)
	if !errors.Is(err, ErrCannotBlockSelf) {
		t.Errorf("expected ErrCannotBlockSelf, got %v", err)
	}
}

func TestBlockUser_AlreadyBlocked(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getBlockFunc: func(ctx context.Context, blockerID, blockedID string) (*model.Block, error) {
			return &model.Block{ID: "block:existing"}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	req := &model.CreateBlockRequest{BlockedUserID: "user:other"}
	_, err := svc.BlockUser(ctx, "user:me", req)
	if !errors.Is(err, ErrAlreadyBlocked) {
		t.Errorf("expected ErrAlreadyBlocked, got %v", err)
	}
}

func TestUnblockUser_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getBlockFunc: func(ctx context.Context, blockerID, blockedID string) (*model.Block, error) {
			return &model.Block{ID: "block:1"}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	err := svc.UnblockUser(ctx, "user:me", "user:other")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnblockUser_NotBlocked(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getBlockFunc: func(ctx context.Context, blockerID, blockedID string) (*model.Block, error) {
			return nil, nil
		},
	}
	svc := NewModerationService(repo, nil)

	err := svc.UnblockUser(ctx, "user:me", "user:other")
	if !errors.Is(err, ErrNotBlocked) {
		t.Errorf("expected ErrNotBlocked, got %v", err)
	}
}

func TestGetBlockedUsers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		getBlocksByBlockerFunc: func(ctx context.Context, blockerID string) ([]*model.Block, error) {
			return []*model.Block{
				{ID: "block:1", BlockedUserID: "user:a"},
				{ID: "block:2", BlockedUserID: "user:b"},
			}, nil
		},
	}
	svc := NewModerationService(repo, nil)

	blocks, err := svc.GetBlockedUsers(ctx, "user:me")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(blocks) != 2 {
		t.Errorf("expected 2 blocks, got %d", len(blocks))
	}
}

func TestIsBlocked(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		isBlockedFunc: func(ctx context.Context, blockerID, blockedID string) (bool, error) {
			return blockerID == "user:a" && blockedID == "user:b", nil
		},
	}
	svc := NewModerationService(repo, nil)

	blocked, err := svc.IsBlocked(ctx, "user:a", "user:b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !blocked {
		t.Error("expected blocked")
	}

	blocked, err = svc.IsBlocked(ctx, "user:b", "user:a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if blocked {
		t.Error("expected not blocked (opposite direction)")
	}
}

func TestIsBlockedEitherWay(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &mockModerationRepo{
		isBlockedEitherWayFunc: func(ctx context.Context, userID1, userID2 string) (bool, error) {
			// Return true if either has blocked the other
			return true, nil
		},
	}
	svc := NewModerationService(repo, nil)

	blocked, err := svc.IsBlockedEitherWay(ctx, "user:a", "user:b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !blocked {
		t.Error("expected blocked either way")
	}
}

// ============================================================================
// Model Validation Helper Tests
// ============================================================================

func TestIsValidReportCategory_Valid(t *testing.T) {
	t.Parallel()

	validCategories := []string{"spam", "harassment", "hate_speech", "inappropriate_content", "made_uncomfortable", "other"}
	for _, cat := range validCategories {
		if !model.IsValidReportCategory(cat) {
			t.Errorf("expected %s to be valid", cat)
		}
	}
}

func TestIsValidReportCategory_Invalid(t *testing.T) {
	t.Parallel()

	if model.IsValidReportCategory("invalid") {
		t.Error("expected invalid category to be invalid")
	}
}

func TestIsValidModerationLevel_Valid(t *testing.T) {
	t.Parallel()

	validLevels := []string{"nudge", "warning", "suspension", "ban"}
	for _, level := range validLevels {
		if !model.IsValidModerationLevel(level) {
			t.Errorf("expected %s to be valid", level)
		}
	}
}

func TestIsValidModerationLevel_Invalid(t *testing.T) {
	t.Parallel()

	if model.IsValidModerationLevel("invalid") {
		t.Error("expected invalid level to be invalid")
	}
}

func TestIsValidReportStatus_Valid(t *testing.T) {
	t.Parallel()

	validStatuses := []string{"pending", "reviewed", "resolved", "dismissed"}
	for _, status := range validStatuses {
		if !model.IsValidReportStatus(status) {
			t.Errorf("expected %s to be valid", status)
		}
	}
}

func TestIsValidReportStatus_Invalid(t *testing.T) {
	t.Parallel()

	if model.IsValidReportStatus("invalid") {
		t.Error("expected invalid status to be invalid")
	}
}
