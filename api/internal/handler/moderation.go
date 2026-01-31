package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// UserFetcher defines the interface for fetching users
type UserFetcher interface {
	GetByID(ctx context.Context, id string) (*model.User, error)
}

// ModerationHandler handles moderation HTTP requests
type ModerationHandler struct {
	moderationService *service.ModerationService
	userFetcher       UserFetcher
}

// NewModerationHandler creates a new moderation handler
func NewModerationHandler(moderationService *service.ModerationService, userFetcher UserFetcher) *ModerationHandler {
	return &ModerationHandler{
		moderationService: moderationService,
		userFetcher:       userFetcher,
	}
}

// requireModerator checks if the current user has moderator or admin role
func (h *ModerationHandler) requireModerator(ctx context.Context, userID string) (*model.User, error) {
	user, err := h.userFetcher.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if !user.CanModerate() {
		return nil, errors.New("insufficient permissions")
	}
	return user, nil
}

// requireAdmin checks if the current user has admin role
func (h *ModerationHandler) requireAdmin(ctx context.Context, userID string) (*model.User, error) {
	user, err := h.userFetcher.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if !user.IsAdmin() {
		return nil, errors.New("admin access required")
	}
	return user, nil
}

// RegisterRoutes registers moderation routes
func (h *ModerationHandler) RegisterRoutes(mux *http.ServeMux) {
	// Reports
	mux.HandleFunc("POST /v1/reports", h.CreateReport)
	mux.HandleFunc("GET /v1/reports/{reportId}", h.GetReport)
	mux.HandleFunc("GET /v1/reports/pending", h.GetPendingReports)
	mux.HandleFunc("PATCH /v1/reports/{reportId}/review", h.ReviewReport)

	// Moderation actions (admin)
	mux.HandleFunc("POST /v1/moderation/actions", h.TakeAction)
	mux.HandleFunc("GET /v1/moderation/actions/{actionId}", h.GetAction)
	mux.HandleFunc("POST /v1/moderation/actions/{actionId}/lift", h.LiftAction)
	mux.HandleFunc("GET /v1/moderation/users/{userId}/status", h.GetUserStatus)
	mux.HandleFunc("GET /v1/moderation/users/{userId}/actions", h.GetUserActions)
	mux.HandleFunc("GET /v1/moderation/stats", h.GetStats)

	// Blocks (user-facing)
	mux.HandleFunc("POST /v1/blocks", h.BlockUser)
	mux.HandleFunc("GET /v1/blocks", h.GetBlockedUsers)
	mux.HandleFunc("DELETE /v1/blocks/{blockedUserId}", h.UnblockUser)
	mux.HandleFunc("GET /v1/blocks/{userId}/check", h.CheckBlock)
}

// Report handlers

// CreateReport creates a new report
func (h *ModerationHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var req model.CreateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	report, err := h.moderationService.CreateReport(ctx, userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCannotReportSelf):
			WriteError(w, model.NewBadRequestError("cannot report yourself"))
		case errors.Is(err, service.ErrInvalidCategory):
			WriteError(w, model.NewBadRequestError("invalid report category"))
		case errors.Is(err, service.ErrDescriptionTooLong):
			WriteError(w, model.NewBadRequestError("description too long"))
		default:
			WriteError(w, model.NewInternalError("failed to create report"))
		}
		return
	}

	WriteData(w, http.StatusCreated, report, nil)
}

// GetReport retrieves a report by ID
func (h *ModerationHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	reportID := r.PathValue("reportId")
	if reportID == "" {
		WriteError(w, model.NewBadRequestError("report ID required"))
		return
	}

	report, err := h.moderationService.GetReport(ctx, reportID)
	if err != nil {
		if errors.Is(err, service.ErrReportNotFound) {
			WriteError(w, model.NewNotFoundError("report not found"))
			return
		}
		WriteError(w, model.NewInternalError("failed to get report"))
		return
	}

	WriteData(w, http.StatusOK, report, nil)
}

// GetPendingReports retrieves pending reports (moderator/admin only)
func (h *ModerationHandler) GetPendingReports(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	// Check moderator permissions
	if _, err := h.requireModerator(ctx, userID); err != nil {
		WriteError(w, model.NewForbiddenError("moderator access required"))
		return
	}

	reports, err := h.moderationService.GetPendingReports(ctx, 50)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get reports"))
		return
	}

	WriteData(w, http.StatusOK, map[string]interface{}{
		"reports": reports,
	}, nil)
}

// ReviewReport reviews a report (moderator/admin only)
func (h *ModerationHandler) ReviewReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	// Check moderator permissions
	if _, err := h.requireModerator(ctx, userID); err != nil {
		WriteError(w, model.NewForbiddenError("moderator access required"))
		return
	}

	reportID := r.PathValue("reportId")
	if reportID == "" {
		WriteError(w, model.NewBadRequestError("report ID required"))
		return
	}

	var req model.ReviewReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	report, err := h.moderationService.ReviewReport(ctx, reportID, userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrReportNotFound):
			WriteError(w, model.NewNotFoundError("report not found"))
		case errors.Is(err, service.ErrInvalidStatus):
			WriteError(w, model.NewBadRequestError("invalid status"))
		default:
			WriteError(w, model.NewInternalError("failed to review report"))
		}
		return
	}

	WriteData(w, http.StatusOK, report, nil)
}

// Action handlers

// TakeAction takes a moderation action
// - Moderators can issue nudges and warnings
// - Admins can issue suspensions and bans
func (h *ModerationHandler) TakeAction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var req model.CreateModerationActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	// Check permissions based on action level
	level := model.ModerationLevel(req.Level)
	if level == model.ModerationLevelSuspension || level == model.ModerationLevelBan {
		// Admin required for suspensions and bans
		if _, err := h.requireAdmin(ctx, userID); err != nil {
			WriteError(w, model.NewForbiddenError("admin access required for suspensions and bans"))
			return
		}
	} else {
		// Moderator access for nudges and warnings
		if _, err := h.requireModerator(ctx, userID); err != nil {
			WriteError(w, model.NewForbiddenError("moderator access required"))
			return
		}
	}

	action, err := h.moderationService.TakeAction(ctx, userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidLevel):
			WriteError(w, model.NewBadRequestError("invalid moderation level"))
		case errors.Is(err, service.ErrReasonRequired):
			WriteError(w, model.NewBadRequestError("reason is required"))
		case errors.Is(err, service.ErrDescriptionTooLong):
			WriteError(w, model.NewBadRequestError("reason too long"))
		default:
			WriteError(w, model.NewInternalError("failed to create action"))
		}
		return
	}

	WriteData(w, http.StatusCreated, action, nil)
}

// GetAction retrieves a moderation action (moderator/admin only)
func (h *ModerationHandler) GetAction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	// Check moderator permissions
	if _, err := h.requireModerator(ctx, userID); err != nil {
		WriteError(w, model.NewForbiddenError("moderator access required"))
		return
	}

	actionID := r.PathValue("actionId")
	if actionID == "" {
		WriteError(w, model.NewBadRequestError("action ID required"))
		return
	}

	// Get action from repository through service
	status, err := h.moderationService.GetUserModerationStatus(ctx, userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get action"))
		return
	}

	// Find the specific action
	for _, a := range status.ActiveActions {
		if a.ID == actionID {
			WriteData(w, http.StatusOK, a, nil)
			return
		}
	}

	WriteError(w, model.NewNotFoundError("action not found"))
}

// LiftAction lifts a moderation action (admin only)
func (h *ModerationHandler) LiftAction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	// Only admins can lift moderation actions
	if _, err := h.requireAdmin(ctx, userID); err != nil {
		WriteError(w, model.NewForbiddenError("admin access required"))
		return
	}

	actionID := r.PathValue("actionId")
	if actionID == "" {
		WriteError(w, model.NewBadRequestError("action ID required"))
		return
	}

	var req model.LiftActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	if err := h.moderationService.LiftAction(ctx, actionID, userID, &req); err != nil {
		switch {
		case errors.Is(err, service.ErrActionNotFound):
			WriteError(w, model.NewNotFoundError("action not found"))
		case errors.Is(err, service.ErrReasonRequired):
			WriteError(w, model.NewBadRequestError("reason is required"))
		default:
			WriteError(w, model.NewInternalError("failed to lift action"))
		}
		return
	}

	WriteData(w, http.StatusOK, map[string]interface{}{
		"message": "action lifted successfully",
	}, nil)
}

// GetUserStatus retrieves a user's moderation status
// Users can view their own status, moderators/admins can view anyone's
func (h *ModerationHandler) GetUserStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestingUserID := middleware.GetUserID(ctx)
	if requestingUserID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	targetUserID := r.PathValue("userId")
	if targetUserID == "" {
		WriteError(w, model.NewBadRequestError("user ID required"))
		return
	}

	// Users can view their own status, moderators/admins can view anyone's
	if targetUserID != requestingUserID {
		if _, err := h.requireModerator(ctx, requestingUserID); err != nil {
			WriteError(w, model.NewForbiddenError("can only view your own moderation status"))
			return
		}
	}

	status, err := h.moderationService.GetUserModerationStatus(ctx, targetUserID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get user status"))
		return
	}

	WriteData(w, http.StatusOK, status, nil)
}

// GetUserActions retrieves all moderation actions for a user (moderator/admin only)
func (h *ModerationHandler) GetUserActions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	// Check moderator permissions
	if _, err := h.requireModerator(ctx, userID); err != nil {
		WriteError(w, model.NewForbiddenError("moderator access required"))
		return
	}

	targetUserID := r.PathValue("userId")
	if targetUserID == "" {
		WriteError(w, model.NewBadRequestError("user ID required"))
		return
	}

	status, err := h.moderationService.GetUserModerationStatus(ctx, targetUserID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get user actions"))
		return
	}

	WriteData(w, http.StatusOK, map[string]interface{}{
		"actions": status.ActiveActions,
	}, nil)
}

// GetStats retrieves moderation statistics (admin only)
func (h *ModerationHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	// Only admins can view moderation stats
	if _, err := h.requireAdmin(ctx, userID); err != nil {
		WriteError(w, model.NewForbiddenError("admin access required"))
		return
	}

	stats, err := h.moderationService.GetModerationStats(ctx)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get stats"))
		return
	}

	WriteData(w, http.StatusOK, stats, nil)
}

// Block handlers

// BlockUser blocks another user
func (h *ModerationHandler) BlockUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var req model.CreateBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	block, err := h.moderationService.BlockUser(ctx, userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCannotBlockSelf):
			WriteError(w, model.NewBadRequestError("cannot block yourself"))
		case errors.Is(err, service.ErrAlreadyBlocked):
			WriteError(w, model.NewConflictError("user already blocked"))
		default:
			WriteError(w, model.NewInternalError("failed to block user"))
		}
		return
	}

	WriteData(w, http.StatusCreated, block, nil)
}

// GetBlockedUsers retrieves users blocked by the current user
func (h *ModerationHandler) GetBlockedUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	blocks, err := h.moderationService.GetBlockedUsers(ctx, userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get blocked users"))
		return
	}

	WriteData(w, http.StatusOK, map[string]interface{}{
		"blocks": blocks,
	}, nil)
}

// UnblockUser unblocks a user
func (h *ModerationHandler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	blockedUserID := r.PathValue("blockedUserId")
	if blockedUserID == "" {
		WriteError(w, model.NewBadRequestError("blocked user ID required"))
		return
	}

	if err := h.moderationService.UnblockUser(ctx, userID, blockedUserID); err != nil {
		switch {
		case errors.Is(err, service.ErrNotBlocked):
			WriteError(w, model.NewNotFoundError("user not blocked"))
		default:
			WriteError(w, model.NewInternalError("failed to unblock user"))
		}
		return
	}

	WriteData(w, http.StatusOK, map[string]interface{}{
		"message": "user unblocked successfully",
	}, nil)
}

// CheckBlock checks if a block exists between users
func (h *ModerationHandler) CheckBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	otherUserID := r.PathValue("userId")
	if otherUserID == "" {
		WriteError(w, model.NewBadRequestError("user ID required"))
		return
	}

	blocked, err := h.moderationService.IsBlockedEitherWay(ctx, userID, otherUserID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to check block status"))
		return
	}

	WriteData(w, http.StatusOK, map[string]interface{}{
		"blocked": blocked,
	}, nil)
}
