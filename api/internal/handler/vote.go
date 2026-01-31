package handler

import (
	"net/http"
	"strconv"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// VoteHandler handles vote HTTP requests
type VoteHandler struct {
	svc *service.VoteService
}

// NewVoteHandler creates a new vote handler
func NewVoteHandler(svc *service.VoteService) *VoteHandler {
	return &VoteHandler{svc: svc}
}

// Vote Management Endpoints

// Create handles POST /v1/votes
func (h *VoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var req model.CreateVoteRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	vote, err := h.svc.Create(ctx, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, vote, nil)
}

// GetByID handles GET /v1/votes/{voteId}
func (h *VoteHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	voteID := r.PathValue("voteId")

	vote, err := h.svc.GetByID(ctx, voteID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, vote, nil)
}

// Update handles PATCH /v1/votes/{voteId}
func (h *VoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	voteID := r.PathValue("voteId")

	var req model.UpdateVoteRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	vote, err := h.svc.Update(ctx, voteID, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, vote, nil)
}

// Open handles POST /v1/votes/{voteId}/open
func (h *VoteHandler) Open(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	voteID := r.PathValue("voteId")

	if err := h.svc.Open(ctx, voteID, userID); err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, map[string]string{"status": "opened"}, nil)
}

// Close handles POST /v1/votes/{voteId}/close
func (h *VoteHandler) Close(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	voteID := r.PathValue("voteId")

	if err := h.svc.Close(ctx, voteID, userID); err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, map[string]string{"status": "closed"}, nil)
}

// Cancel handles POST /v1/votes/{voteId}/cancel
func (h *VoteHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	voteID := r.PathValue("voteId")

	if err := h.svc.Cancel(ctx, voteID, userID); err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, map[string]string{"status": "cancelled"}, nil)
}

// Delete handles DELETE /v1/votes/{voteId}
// Note: Vote deletion is handled via Cancel for now
func (h *VoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	voteID := r.PathValue("voteId")

	// Votes are cancelled rather than deleted to maintain audit trail
	if err := h.svc.Cancel(ctx, voteID, userID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Option Endpoints

// CreateOption handles POST /v1/votes/{voteId}/options
func (h *VoteHandler) CreateOption(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	voteID := r.PathValue("voteId")

	var req model.CreateVoteOptionRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	option, err := h.svc.AddOption(ctx, voteID, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, option, nil)
}

// GetOptions handles GET /v1/votes/{voteId}/options
func (h *VoteHandler) GetOptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	voteID := r.PathValue("voteId")

	vote, err := h.svc.GetByID(ctx, voteID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, vote.Options, nil, nil)
}

// UpdateOption handles PATCH /v1/votes/{voteId}/options/{optionId}
func (h *VoteHandler) UpdateOption(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	optionID := r.PathValue("optionId")

	var req model.UpdateVoteOptionRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	option, err := h.svc.UpdateOption(ctx, optionID, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, option, nil)
}

// DeleteOption handles DELETE /v1/votes/{voteId}/options/{optionId}
func (h *VoteHandler) DeleteOption(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	optionID := r.PathValue("optionId")

	if err := h.svc.DeleteOption(ctx, optionID, userID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Ballot Endpoints

// CastBallot handles POST /v1/votes/{voteId}/ballot
func (h *VoteHandler) CastBallot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	voteID := r.PathValue("voteId")

	var req model.CastBallotRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	ballot, err := h.svc.CastBallot(ctx, voteID, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, ballot, nil)
}

// GetMyBallot handles GET /v1/votes/{voteId}/ballot
func (h *VoteHandler) GetMyBallot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	voteID := r.PathValue("voteId")

	ballot, err := h.svc.GetMyBallot(ctx, voteID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, ballot, nil)
}

// GetBallots handles GET /v1/votes/{voteId}/ballots
func (h *VoteHandler) GetBallots(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	voteID := r.PathValue("voteId")

	ballots, err := h.svc.GetBallots(ctx, voteID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, ballots, nil, nil)
}

// Results Endpoint

// GetResults handles GET /v1/votes/{voteId}/results
func (h *VoteHandler) GetResults(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	voteID := r.PathValue("voteId")

	results, err := h.svc.GetResults(ctx, voteID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, results, nil)
}

// Scoped Query Endpoints

// GetGuildVotes handles GET /v1/guilds/{guildId}/votes
func (h *VoteHandler) GetGuildVotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	guildID := r.PathValue("guildId")

	limit, offset := getVotePaginationParams(r)
	status := getStatusFilterString(r)

	votes, err := h.svc.GetGuildVotes(ctx, guildID, status, limit, offset)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, votes, nil, nil)
}

// GetGlobalVotes handles GET /v1/votes/global
func (h *VoteHandler) GetGlobalVotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit, offset := getVotePaginationParams(r)
	status := getStatusFilterString(r)

	votes, err := h.svc.GetGlobalVotes(ctx, status, limit, offset)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, votes, nil, nil)
}

// GetVoteStats handles GET /v1/votes/{voteId}/stats
func (h *VoteHandler) GetVoteStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	voteID := r.PathValue("voteId")

	vote, err := h.svc.GetByID(ctx, voteID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	stats := map[string]interface{}{
		"vote_id":      vote.Vote.ID,
		"ballot_count": vote.Vote.BallotCount,
		"status":       vote.Vote.Status,
	}

	WriteData(w, http.StatusOK, stats, nil)
}

// BatchCreateOptions handles POST /v1/votes/{voteId}/options/batch
func (h *VoteHandler) BatchCreateOptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	voteID := r.PathValue("voteId")

	var req struct {
		Options []model.CreateVoteOptionRequest `json:"options"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	if len(req.Options) == 0 {
		WriteError(w, model.NewBadRequestError("at least one option required"))
		return
	}

	if len(req.Options) > 50 {
		WriteError(w, model.NewBadRequestError("maximum 50 options per batch"))
		return
	}

	created := make([]*model.VoteOption, 0, len(req.Options))
	for i := range req.Options {
		optReq := &req.Options[i]
		if optReq.SortOrder == nil {
			sortOrder := i + 1
			optReq.SortOrder = &sortOrder
		}
		option, err := h.svc.AddOption(ctx, voteID, userID, optReq)
		if err != nil {
			h.handleError(w, err)
			return
		}
		created = append(created, option)
	}

	WriteData(w, http.StatusCreated, created, nil)
}

// handleError converts service errors to HTTP responses
func (h *VoteHandler) handleError(w http.ResponseWriter, err error) {
	if pd, ok := err.(*model.ProblemDetails); ok {
		WriteError(w, pd)
		return
	}
	WriteError(w, model.NewInternalError("internal server error"))
}

// Helper function for status filter (returns *string for service calls)
func getStatusFilterString(r *http.Request) *string {
	if v := r.URL.Query().Get("status"); v != "" {
		return &v
	}
	return nil
}

// Helper function for pagination (vote-specific to avoid redeclaration)
func getVotePaginationParams(r *http.Request) (limit, offset int) {
	limit = 50
	offset = 0

	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if v := r.URL.Query().Get("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}
