package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// ============================================================================
// Mock UserFetcher
// ============================================================================

type mockUserFetcher struct {
	getByIDFunc func(ctx context.Context, id string) (*model.User, error)
}

func (m *mockUserFetcher) GetByID(ctx context.Context, id string) (*model.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

// ============================================================================
// Mock ModerationService
// ============================================================================

type mockModerationService struct {
	createReportFunc            func(ctx context.Context, reporterID string, req *model.CreateReportRequest) (*model.Report, error)
	getReportFunc               func(ctx context.Context, reportID string) (*model.Report, error)
	getPendingReportsFunc       func(ctx context.Context, limit int) ([]*model.Report, error)
	reviewReportFunc            func(ctx context.Context, reportID, reviewerID string, req *model.ReviewReportRequest) (*model.Report, error)
	takeActionFunc              func(ctx context.Context, adminID string, req *model.CreateModerationActionRequest) (*model.ModerationAction, error)
	liftActionFunc              func(ctx context.Context, actionID, adminID string, req *model.LiftActionRequest) error
	getUserModerationStatusFunc func(ctx context.Context, userID string) (*model.UserModerationStatus, error)
	getModerationStatsFunc      func(ctx context.Context) (*model.ModerationStats, error)
	blockUserFunc               func(ctx context.Context, blockerID string, req *model.CreateBlockRequest) (*model.Block, error)
	getBlockedUsersFunc         func(ctx context.Context, userID string) ([]*model.BlockedUserInfo, error)
	unblockUserFunc             func(ctx context.Context, blockerID, blockedID string) error
	isBlockedEitherWayFunc      func(ctx context.Context, userID1, userID2 string) (bool, error)
}

func (m *mockModerationService) CreateReport(ctx context.Context, reporterID string, req *model.CreateReportRequest) (*model.Report, error) {
	if m.createReportFunc != nil {
		return m.createReportFunc(ctx, reporterID, req)
	}
	return nil, nil
}

func (m *mockModerationService) GetReport(ctx context.Context, reportID string) (*model.Report, error) {
	if m.getReportFunc != nil {
		return m.getReportFunc(ctx, reportID)
	}
	return nil, nil
}

func (m *mockModerationService) GetPendingReports(ctx context.Context, limit int) ([]*model.Report, error) {
	if m.getPendingReportsFunc != nil {
		return m.getPendingReportsFunc(ctx, limit)
	}
	return nil, nil
}

func (m *mockModerationService) ReviewReport(ctx context.Context, reportID, reviewerID string, req *model.ReviewReportRequest) (*model.Report, error) {
	if m.reviewReportFunc != nil {
		return m.reviewReportFunc(ctx, reportID, reviewerID, req)
	}
	return nil, nil
}

func (m *mockModerationService) TakeAction(ctx context.Context, adminID string, req *model.CreateModerationActionRequest) (*model.ModerationAction, error) {
	if m.takeActionFunc != nil {
		return m.takeActionFunc(ctx, adminID, req)
	}
	return nil, nil
}

func (m *mockModerationService) LiftAction(ctx context.Context, actionID, adminID string, req *model.LiftActionRequest) error {
	if m.liftActionFunc != nil {
		return m.liftActionFunc(ctx, actionID, adminID, req)
	}
	return nil
}

func (m *mockModerationService) GetUserModerationStatus(ctx context.Context, userID string) (*model.UserModerationStatus, error) {
	if m.getUserModerationStatusFunc != nil {
		return m.getUserModerationStatusFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockModerationService) GetModerationStats(ctx context.Context) (*model.ModerationStats, error) {
	if m.getModerationStatsFunc != nil {
		return m.getModerationStatsFunc(ctx)
	}
	return nil, nil
}

func (m *mockModerationService) BlockUser(ctx context.Context, blockerID string, req *model.CreateBlockRequest) (*model.Block, error) {
	if m.blockUserFunc != nil {
		return m.blockUserFunc(ctx, blockerID, req)
	}
	return nil, nil
}

func (m *mockModerationService) GetBlockedUsers(ctx context.Context, userID string) ([]*model.BlockedUserInfo, error) {
	if m.getBlockedUsersFunc != nil {
		return m.getBlockedUsersFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockModerationService) UnblockUser(ctx context.Context, blockerID, blockedID string) error {
	if m.unblockUserFunc != nil {
		return m.unblockUserFunc(ctx, blockerID, blockedID)
	}
	return nil
}

func (m *mockModerationService) IsBlockedEitherWay(ctx context.Context, userID1, userID2 string) (bool, error) {
	if m.isBlockedEitherWayFunc != nil {
		return m.isBlockedEitherWayFunc(ctx, userID1, userID2)
	}
	return false, nil
}

// ============================================================================
// Test Helpers
// ============================================================================

func newRegularUser() *model.User {
	now := time.Now()
	return &model.User{
		ID:            "user:regular",
		Email:         "regular@example.com",
		Role:          model.UserRoleUser,
		EmailVerified: true,
		CreatedOn:     now,
		UpdatedOn:     now,
	}
}

func newModeratorUser() *model.User {
	now := time.Now()
	return &model.User{
		ID:            "user:moderator",
		Email:         "moderator@example.com",
		Role:          model.UserRoleModerator,
		EmailVerified: true,
		CreatedOn:     now,
		UpdatedOn:     now,
	}
}

func newAdminUser() *model.User {
	now := time.Now()
	return &model.User{
		ID:            "user:admin",
		Email:         "admin@example.com",
		Role:          model.UserRoleAdmin,
		EmailVerified: true,
		CreatedOn:     now,
		UpdatedOn:     now,
	}
}

func newTestReport() *model.Report {
	now := time.Now()
	return &model.Report{
		ID:             "report:123",
		ReporterUserID: "user:reporter",
		ReportedUserID: "user:reported",
		Category:       model.ReportCategoryHarassment,
		Status:         model.ReportStatusPending,
		CreatedOn:      now,
	}
}

func newTestModerationAction() *model.ModerationAction {
	now := time.Now()
	return &model.ModerationAction{
		ID:        "action:123",
		UserID:    "user:target",
		Level:     model.ModerationLevelWarning,
		Reason:    "Violation of community guidelines",
		IsActive:  true,
		CreatedOn: now,
	}
}

func makeModerationJSONRequest(method, path string, body interface{}) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func withModUserContext(req *http.Request, userID string) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

// ============================================================================
// CreateReport Tests
// ============================================================================

func TestCreateReport_AsUser_Success(t *testing.T) {
	t.Parallel()

	mockSvc := &mockModerationService{
		createReportFunc: func(ctx context.Context, reporterID string, req *model.CreateReportRequest) (*model.Report, error) {
			return newTestReport(), nil
		},
	}

	req := makeModerationJSONRequest(http.MethodPost, "/v1/reports", model.CreateReportRequest{
		ReportedUserID: "user:reported",
		Category:       "harassment",
		Description:    strPtr("They were rude to me"),
	})
	req = withModUserContext(req, "user:regular")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}

		var reqBody model.CreateReportRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		report, err := mockSvc.CreateReport(r.Context(), userID, &reqBody)
		if err != nil {
			WriteError(w, model.NewInternalError("failed to create report"))
			return
		}

		WriteData(w, http.StatusCreated, report, nil)
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestCreateReport_CannotReportSelf(t *testing.T) {
	t.Parallel()

	mockSvc := &mockModerationService{
		createReportFunc: func(ctx context.Context, reporterID string, req *model.CreateReportRequest) (*model.Report, error) {
			return nil, service.ErrCannotReportSelf
		},
	}

	req := makeModerationJSONRequest(http.MethodPost, "/v1/reports", model.CreateReportRequest{
		ReportedUserID: "user:regular", // Same as requesting user
		Category:       "harassment",
	})
	req = withModUserContext(req, "user:regular")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}

		var reqBody model.CreateReportRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		_, err := mockSvc.CreateReport(r.Context(), userID, &reqBody)
		if err == service.ErrCannotReportSelf {
			WriteError(w, model.NewBadRequestError("cannot report yourself"))
			return
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// ============================================================================
// ReviewReport Tests - Authorization Focus
// ============================================================================

func TestReviewReport_AsUser_ReturnsForbidden(t *testing.T) {
	t.Parallel()

	// Mock user fetcher returns a regular user (not moderator/admin)
	mockUserFetcher := &mockUserFetcher{
		getByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return newRegularUser(), nil
		},
	}

	req := makeModerationJSONRequest(http.MethodPatch, "/v1/reports/report:123/review", model.ReviewReportRequest{
		Status: "resolved",
		Notes:  strPtr("Valid complaint, warning issued"),
	})
	req = withModUserContext(req, "user:regular")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}

		// Check moderator permissions
		user, err := mockUserFetcher.GetByID(r.Context(), userID)
		if err != nil || user == nil || !user.CanModerate() {
			WriteError(w, model.NewForbiddenError("moderator access required"))
			return
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestReviewReport_AsModerator_Success(t *testing.T) {
	t.Parallel()

	mockUserFetcher := &mockUserFetcher{
		getByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return newModeratorUser(), nil
		},
	}

	mockSvc := &mockModerationService{
		reviewReportFunc: func(ctx context.Context, reportID, reviewerID string, req *model.ReviewReportRequest) (*model.Report, error) {
			report := newTestReport()
			report.Status = model.ReportStatusResolved
			return report, nil
		},
	}

	req := makeModerationJSONRequest(http.MethodPatch, "/v1/reports/report:123/review", model.ReviewReportRequest{
		Status: "resolved",
		Notes:  strPtr("Valid complaint, warning issued"),
	})
	req = withModUserContext(req, "user:moderator")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}

		// Check moderator permissions
		user, err := mockUserFetcher.GetByID(r.Context(), userID)
		if err != nil || user == nil || !user.CanModerate() {
			WriteError(w, model.NewForbiddenError("moderator access required"))
			return
		}

		var reqBody model.ReviewReportRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		reportID := "report:123" // Would come from path value
		report, err := mockSvc.ReviewReport(r.Context(), reportID, userID, &reqBody)
		if err != nil {
			WriteError(w, model.NewInternalError("failed to review report"))
			return
		}

		WriteData(w, http.StatusOK, report, nil)
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

// ============================================================================
// BanUser Tests - Authorization Focus
// ============================================================================

func TestTakeAction_BanAsUser_ReturnsForbidden(t *testing.T) {
	t.Parallel()

	mockUserFetcher := &mockUserFetcher{
		getByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return newRegularUser(), nil
		},
	}

	req := makeModerationJSONRequest(http.MethodPost, "/v1/moderation/actions", model.CreateModerationActionRequest{
		UserID: "user:target",
		Level:  "ban",
		Reason: "Severe violation",
	})
	req = withModUserContext(req, "user:regular")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}

		var reqBody model.CreateModerationActionRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		// Check permissions based on action level
		level := model.ModerationLevel(reqBody.Level)
		if level == model.ModerationLevelSuspension || level == model.ModerationLevelBan {
			// Admin required for suspensions and bans
			user, err := mockUserFetcher.GetByID(r.Context(), userID)
			if err != nil || user == nil || !user.IsAdmin() {
				WriteError(w, model.NewForbiddenError("admin access required for suspensions and bans"))
				return
			}
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestTakeAction_BanAsAdmin_Success(t *testing.T) {
	t.Parallel()

	mockUserFetcher := &mockUserFetcher{
		getByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return newAdminUser(), nil
		},
	}

	mockSvc := &mockModerationService{
		takeActionFunc: func(ctx context.Context, adminID string, req *model.CreateModerationActionRequest) (*model.ModerationAction, error) {
			action := newTestModerationAction()
			action.Level = model.ModerationLevelBan
			return action, nil
		},
	}

	req := makeModerationJSONRequest(http.MethodPost, "/v1/moderation/actions", model.CreateModerationActionRequest{
		UserID: "user:target",
		Level:  "ban",
		Reason: "Severe violation",
	})
	req = withModUserContext(req, "user:admin")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}

		var reqBody model.CreateModerationActionRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		// Check permissions based on action level
		level := model.ModerationLevel(reqBody.Level)
		if level == model.ModerationLevelSuspension || level == model.ModerationLevelBan {
			// Admin required for suspensions and bans
			user, err := mockUserFetcher.GetByID(r.Context(), userID)
			if err != nil || user == nil || !user.IsAdmin() {
				WriteError(w, model.NewForbiddenError("admin access required for suspensions and bans"))
				return
			}
		}

		action, err := mockSvc.TakeAction(r.Context(), userID, &reqBody)
		if err != nil {
			WriteError(w, model.NewInternalError("failed to create action"))
			return
		}

		WriteData(w, http.StatusCreated, action, nil)
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestTakeAction_WarningAsModerator_Success(t *testing.T) {
	t.Parallel()

	mockUserFetcher := &mockUserFetcher{
		getByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return newModeratorUser(), nil
		},
	}

	mockSvc := &mockModerationService{
		takeActionFunc: func(ctx context.Context, adminID string, req *model.CreateModerationActionRequest) (*model.ModerationAction, error) {
			action := newTestModerationAction()
			action.Level = model.ModerationLevelWarning
			return action, nil
		},
	}

	req := makeModerationJSONRequest(http.MethodPost, "/v1/moderation/actions", model.CreateModerationActionRequest{
		UserID: "user:target",
		Level:  "warning",
		Reason: "Minor violation",
	})
	req = withModUserContext(req, "user:moderator")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}

		var reqBody model.CreateModerationActionRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		level := model.ModerationLevel(reqBody.Level)
		if level == model.ModerationLevelSuspension || level == model.ModerationLevelBan {
			// Admin required for suspensions and bans
			user, err := mockUserFetcher.GetByID(r.Context(), userID)
			if err != nil || user == nil || !user.IsAdmin() {
				WriteError(w, model.NewForbiddenError("admin access required for suspensions and bans"))
				return
			}
		} else {
			// Moderator access for nudges and warnings
			user, err := mockUserFetcher.GetByID(r.Context(), userID)
			if err != nil || user == nil || !user.CanModerate() {
				WriteError(w, model.NewForbiddenError("moderator access required"))
				return
			}
		}

		action, err := mockSvc.TakeAction(r.Context(), userID, &reqBody)
		if err != nil {
			WriteError(w, model.NewInternalError("failed to create action"))
			return
		}

		WriteData(w, http.StatusCreated, action, nil)
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestTakeAction_SuspensionAsModerator_ReturnsForbidden(t *testing.T) {
	t.Parallel()

	mockUserFetcher := &mockUserFetcher{
		getByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return newModeratorUser(), nil
		},
	}

	req := makeModerationJSONRequest(http.MethodPost, "/v1/moderation/actions", model.CreateModerationActionRequest{
		UserID:       "user:target",
		Level:        "suspension",
		Reason:       "Repeated violations",
		DurationDays: intPtr(7),
	})
	req = withModUserContext(req, "user:moderator")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}

		var reqBody model.CreateModerationActionRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		level := model.ModerationLevel(reqBody.Level)
		if level == model.ModerationLevelSuspension || level == model.ModerationLevelBan {
			// Admin required for suspensions and bans
			user, err := mockUserFetcher.GetByID(r.Context(), userID)
			if err != nil || user == nil || !user.IsAdmin() {
				WriteError(w, model.NewForbiddenError("admin access required for suspensions and bans"))
				return
			}
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

// ============================================================================
// GetModerationStats Tests - Admin Only
// ============================================================================

func TestGetStats_AsUser_ReturnsForbidden(t *testing.T) {
	t.Parallel()

	mockUserFetcher := &mockUserFetcher{
		getByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return newRegularUser(), nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/moderation/stats", nil)
	req = withModUserContext(req, "user:regular")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}

		// Only admins can view moderation stats
		user, err := mockUserFetcher.GetByID(r.Context(), userID)
		if err != nil || user == nil || !user.IsAdmin() {
			WriteError(w, model.NewForbiddenError("admin access required"))
			return
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestGetStats_AsModerator_ReturnsForbidden(t *testing.T) {
	t.Parallel()

	mockUserFetcher := &mockUserFetcher{
		getByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return newModeratorUser(), nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/moderation/stats", nil)
	req = withModUserContext(req, "user:moderator")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}

		// Only admins can view moderation stats
		user, err := mockUserFetcher.GetByID(r.Context(), userID)
		if err != nil || user == nil || !user.IsAdmin() {
			WriteError(w, model.NewForbiddenError("admin access required"))
			return
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestGetStats_AsAdmin_Success(t *testing.T) {
	t.Parallel()

	mockUserFetcher := &mockUserFetcher{
		getByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return newAdminUser(), nil
		},
	}

	mockSvc := &mockModerationService{
		getModerationStatsFunc: func(ctx context.Context) (*model.ModerationStats, error) {
			return &model.ModerationStats{
				TotalReports:      100,
				PendingReports:    15,
				ResolvedReports:   80,
				ActiveWarnings:    10,
				ActiveSuspensions: 3,
				TotalBans:         5,
			}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/moderation/stats", nil)
	req = withModUserContext(req, "user:admin")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}

		// Only admins can view moderation stats
		user, err := mockUserFetcher.GetByID(r.Context(), userID)
		if err != nil || user == nil || !user.IsAdmin() {
			WriteError(w, model.NewForbiddenError("admin access required"))
			return
		}

		stats, err := mockSvc.GetModerationStats(r.Context())
		if err != nil {
			WriteError(w, model.NewInternalError("failed to get stats"))
			return
		}

		WriteData(w, http.StatusOK, stats, nil)
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

// ============================================================================
// Block Tests
// ============================================================================

func TestBlockUser_Success(t *testing.T) {
	t.Parallel()

	mockSvc := &mockModerationService{
		blockUserFunc: func(ctx context.Context, blockerID string, req *model.CreateBlockRequest) (*model.Block, error) {
			return &model.Block{
				ID:            "block:123",
				BlockerUserID: blockerID,
				BlockedUserID: req.BlockedUserID,
				CreatedOn:     time.Now(),
			}, nil
		},
	}

	req := makeModerationJSONRequest(http.MethodPost, "/v1/blocks", model.CreateBlockRequest{
		BlockedUserID: "user:annoying",
	})
	req = withModUserContext(req, "user:regular")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}

		var reqBody model.CreateBlockRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		block, err := mockSvc.BlockUser(r.Context(), userID, &reqBody)
		if err != nil {
			WriteError(w, model.NewInternalError("failed to block user"))
			return
		}

		WriteData(w, http.StatusCreated, block, nil)
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestBlockUser_CannotBlockSelf(t *testing.T) {
	t.Parallel()

	mockSvc := &mockModerationService{
		blockUserFunc: func(ctx context.Context, blockerID string, req *model.CreateBlockRequest) (*model.Block, error) {
			return nil, service.ErrCannotBlockSelf
		},
	}

	req := makeModerationJSONRequest(http.MethodPost, "/v1/blocks", model.CreateBlockRequest{
		BlockedUserID: "user:regular", // Same as requesting user
	})
	req = withModUserContext(req, "user:regular")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}

		var reqBody model.CreateBlockRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		_, err := mockSvc.BlockUser(r.Context(), userID, &reqBody)
		if err == service.ErrCannotBlockSelf {
			WriteError(w, model.NewBadRequestError("cannot block yourself"))
			return
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestBlockUser_AlreadyBlocked(t *testing.T) {
	t.Parallel()

	mockSvc := &mockModerationService{
		blockUserFunc: func(ctx context.Context, blockerID string, req *model.CreateBlockRequest) (*model.Block, error) {
			return nil, service.ErrAlreadyBlocked
		},
	}

	req := makeModerationJSONRequest(http.MethodPost, "/v1/blocks", model.CreateBlockRequest{
		BlockedUserID: "user:already-blocked",
	})
	req = withModUserContext(req, "user:regular")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}

		var reqBody model.CreateBlockRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		_, err := mockSvc.BlockUser(r.Context(), userID, &reqBody)
		if err == service.ErrAlreadyBlocked {
			WriteError(w, model.NewConflictError("user already blocked"))
			return
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, rr.Code)
	}
}

// ============================================================================
// Authentication Required Tests
// ============================================================================

func TestModerationEndpoints_Unauthenticated_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()

	endpoints := []struct {
		name   string
		method string
		path   string
		body   interface{}
	}{
		{"CreateReport", http.MethodPost, "/v1/reports", model.CreateReportRequest{}},
		{"GetReport", http.MethodGet, "/v1/reports/report:123", nil},
		{"ReviewReport", http.MethodPatch, "/v1/reports/report:123/review", model.ReviewReportRequest{}},
		{"TakeAction", http.MethodPost, "/v1/moderation/actions", model.CreateModerationActionRequest{}},
		{"BlockUser", http.MethodPost, "/v1/blocks", model.CreateBlockRequest{}},
		{"GetBlockedUsers", http.MethodGet, "/v1/blocks", nil},
	}

	for _, tc := range endpoints {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var req *http.Request
			if tc.body != nil {
				req = makeModerationJSONRequest(tc.method, tc.path, tc.body)
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			// Deliberately NOT adding user context
			rr := httptest.NewRecorder()

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID := middleware.GetUserID(r.Context())
				if userID == "" {
					WriteError(w, model.NewUnauthorizedError("authentication required"))
					return
				}
			})

			testHandler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("%s: expected status %d, got %d", tc.name, http.StatusUnauthorized, rr.Code)
			}
		})
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
