package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// PoolHandler handles pool HTTP requests
type PoolHandler struct {
	poolService  *service.PoolService
	guildService *service.GuildService
}

// NewPoolHandler creates a new pool handler
func NewPoolHandler(poolService *service.PoolService, guildService *service.GuildService) *PoolHandler {
	return &PoolHandler{
		poolService:  poolService,
		guildService: guildService,
	}
}

// ListPools handles GET /v1/guilds/{guildId}/pools - list pools in a guild
func (h *PoolHandler) ListPools(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	guildID := r.PathValue("guildId")
	if guildID == "" {
		WriteError(w, model.NewBadRequestError("guild ID required"))
		return
	}

	pools, err := h.poolService.GetPoolsByGuild(ctx, guildID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, pools, nil)
}

// CreatePool handles POST /v1/guilds/{guildId}/pools - create a new pool
func (h *PoolHandler) CreatePool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	guildID := r.PathValue("guildId")
	if guildID == "" {
		WriteError(w, model.NewBadRequestError("guild ID required"))
		return
	}

	var req model.CreatePoolRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	// Validate required fields
	var fieldErrors []model.FieldError
	if req.Name == "" {
		fieldErrors = append(fieldErrors, model.FieldError{
			Field:   "name",
			Message: "name is required",
		})
	}
	if req.Frequency == "" {
		fieldErrors = append(fieldErrors, model.FieldError{
			Field:   "frequency",
			Message: "frequency is required",
		})
	}
	if len(fieldErrors) > 0 {
		WriteError(w, model.NewValidationError(fieldErrors))
		return
	}

	pool, err := h.poolService.CreatePool(ctx, guildID, &req, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, pool, nil)
}

// GetPool handles GET /v1/guilds/{guildId}/pools/{poolId} - get pool details
func (h *PoolHandler) GetPool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	guildID := r.PathValue("guildId")
	poolID := r.PathValue("poolId")
	if guildID == "" || poolID == "" {
		WriteError(w, model.NewBadRequestError("guild ID and pool ID required"))
		return
	}

	// Validate pool belongs to guild
	pool, err := h.poolService.ValidatePoolInGuild(ctx, poolID, guildID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Get pool with members
	poolWithMembers, err := h.poolService.GetPoolWithMembers(ctx, poolID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	_ = pool // Used for validation
	WriteData(w, http.StatusOK, poolWithMembers, nil)
}

// UpdatePool handles PATCH /v1/guilds/{guildId}/pools/{poolId} - update a pool
func (h *PoolHandler) UpdatePool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	guildID := r.PathValue("guildId")
	poolID := r.PathValue("poolId")
	if guildID == "" || poolID == "" {
		WriteError(w, model.NewBadRequestError("guild ID and pool ID required"))
		return
	}

	// Validate pool belongs to guild
	if _, err := h.poolService.ValidatePoolInGuild(ctx, poolID, guildID); err != nil {
		h.handleError(w, err)
		return
	}

	var req model.UpdatePoolRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	pool, err := h.poolService.UpdatePool(ctx, poolID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, pool, nil)
}

// DeletePool handles DELETE /v1/guilds/{guildId}/pools/{poolId} - delete a pool
func (h *PoolHandler) DeletePool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	guildID := r.PathValue("guildId")
	poolID := r.PathValue("poolId")
	if guildID == "" || poolID == "" {
		WriteError(w, model.NewBadRequestError("guild ID and pool ID required"))
		return
	}

	// Validate pool belongs to guild
	if _, err := h.poolService.ValidatePoolInGuild(ctx, poolID, guildID); err != nil {
		h.handleError(w, err)
		return
	}

	if err := h.poolService.DeletePool(ctx, poolID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// JoinPool handles POST /v1/guilds/{guildId}/pools/{poolId}/join - join a pool
func (h *PoolHandler) JoinPool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	guildID := r.PathValue("guildId")
	poolID := r.PathValue("poolId")
	if guildID == "" || poolID == "" {
		WriteError(w, model.NewBadRequestError("guild ID and pool ID required"))
		return
	}

	// Validate pool belongs to guild
	if _, err := h.poolService.ValidatePoolInGuild(ctx, poolID, guildID); err != nil {
		h.handleError(w, err)
		return
	}

	var req model.JoinPoolRequest
	if err := DecodeJSON(r, &req); err != nil {
		// JoinPoolRequest may be empty, that's fine
		req = model.JoinPoolRequest{}
	}

	member, err := h.poolService.JoinPool(ctx, poolID, userID, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, member, nil)
}

// LeavePool handles POST /v1/guilds/{guildId}/pools/{poolId}/leave - leave a pool
func (h *PoolHandler) LeavePool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	guildID := r.PathValue("guildId")
	poolID := r.PathValue("poolId")
	if guildID == "" || poolID == "" {
		WriteError(w, model.NewBadRequestError("guild ID and pool ID required"))
		return
	}

	// Validate pool belongs to guild
	if _, err := h.poolService.ValidatePoolInGuild(ctx, poolID, guildID); err != nil {
		h.handleError(w, err)
		return
	}

	if err := h.poolService.LeavePool(ctx, poolID, userID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetPoolMembers handles GET /v1/guilds/{guildId}/pools/{poolId}/members - list pool members
func (h *PoolHandler) GetPoolMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	guildID := r.PathValue("guildId")
	poolID := r.PathValue("poolId")
	if guildID == "" || poolID == "" {
		WriteError(w, model.NewBadRequestError("guild ID and pool ID required"))
		return
	}

	// Validate pool belongs to guild
	if _, err := h.poolService.ValidatePoolInGuild(ctx, poolID, guildID); err != nil {
		h.handleError(w, err)
		return
	}

	members, err := h.poolService.GetPoolMembers(ctx, poolID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, members, nil)
}

// UpdateMembership handles PATCH /v1/guilds/{guildId}/pools/{poolId}/membership - update membership settings
func (h *PoolHandler) UpdateMembership(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	guildID := r.PathValue("guildId")
	poolID := r.PathValue("poolId")
	if guildID == "" || poolID == "" {
		WriteError(w, model.NewBadRequestError("guild ID and pool ID required"))
		return
	}

	// Validate pool belongs to guild
	if _, err := h.poolService.ValidatePoolInGuild(ctx, poolID, guildID); err != nil {
		h.handleError(w, err)
		return
	}

	var req model.UpdateMembershipRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	member, err := h.poolService.UpdateMembership(ctx, poolID, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, member, nil)
}

// GetPoolStats handles GET /v1/guilds/{guildId}/pools/{poolId}/stats - get pool statistics
func (h *PoolHandler) GetPoolStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	guildID := r.PathValue("guildId")
	poolID := r.PathValue("poolId")
	if guildID == "" || poolID == "" {
		WriteError(w, model.NewBadRequestError("guild ID and pool ID required"))
		return
	}

	// Validate pool belongs to guild
	if _, err := h.poolService.ValidatePoolInGuild(ctx, poolID, guildID); err != nil {
		h.handleError(w, err)
		return
	}

	stats, err := h.poolService.GetPoolStats(ctx, poolID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, stats, nil)
}

// GetMatchHistory handles GET /v1/guilds/{guildId}/pools/{poolId}/matches - get match history
func (h *PoolHandler) GetMatchHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	guildID := r.PathValue("guildId")
	poolID := r.PathValue("poolId")
	if guildID == "" || poolID == "" {
		WriteError(w, model.NewBadRequestError("guild ID and pool ID required"))
		return
	}

	// Validate pool belongs to guild
	if _, err := h.poolService.ValidatePoolInGuild(ctx, poolID, guildID); err != nil {
		h.handleError(w, err)
		return
	}

	// Parse optional limit parameter
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	matches, err := h.poolService.GetMatchHistory(ctx, poolID, limit)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, matches, nil)
}

// GetPendingMatches handles GET /v1/profile/matches/pending - get user's pending matches
func (h *PoolHandler) GetPendingMatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	matches, err := h.poolService.GetPendingMatches(ctx, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, matches, nil)
}

// UpdateMatch handles PATCH /v1/matches/{matchId} - update a match (schedule, status)
func (h *PoolHandler) UpdateMatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	matchID := r.PathValue("matchId")
	if matchID == "" {
		WriteError(w, model.NewBadRequestError("match ID required"))
		return
	}

	var req model.UpdateMatchRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	match, err := h.poolService.UpdateMatch(ctx, matchID, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, match, nil)
}

// handleError converts service errors to HTTP responses
func (h *PoolHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrPoolNotFound):
		WriteError(w, model.NewNotFoundError("pool not found"))
	case errors.Is(err, service.ErrPoolNotInGuild):
		WriteError(w, model.NewNotFoundError("pool not found"))
	case errors.Is(err, service.ErrMatchNotFound):
		WriteError(w, model.NewNotFoundError("match not found"))
	case errors.Is(err, service.ErrNotPoolMember):
		WriteError(w, model.NewNotFoundError("not a pool member"))
	case errors.Is(err, service.ErrNotMatchMember):
		WriteError(w, model.NewForbiddenError("not a member of this match"))
	case errors.Is(err, service.ErrAlreadyPoolMember):
		WriteError(w, model.NewConflictError("already a member of this pool"))
	case errors.Is(err, service.ErrPoolLimitReached):
		WriteError(w, model.NewLimitExceededError("maximum pools per guild reached", model.MaxPoolsPerGuild, model.MaxPoolsPerGuild))
	case errors.Is(err, service.ErrMemberPoolLimitReached):
		WriteError(w, model.NewLimitExceededError("maximum members per pool reached", model.MaxMembersPerPool, model.MaxMembersPerPool))
	case errors.Is(err, service.ErrExclusionLimitReached):
		WriteError(w, model.NewLimitExceededError("maximum exclusions reached", model.MaxExclusionsPerMember, model.MaxExclusionsPerMember))
	case errors.Is(err, service.ErrInvalidMatchSize):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "match_size", Message: "match size must be between 2 and 6"},
		}))
	case errors.Is(err, service.ErrInvalidFrequency):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "frequency", Message: "invalid frequency (use weekly, biweekly, or monthly)"},
		}))
	default:
		WriteError(w, model.NewInternalError("an unexpected error occurred"))
	}
}
