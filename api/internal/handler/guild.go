package handler

import (
	"errors"
	"net/http"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// GuildHandler handles guild HTTP requests
type GuildHandler struct {
	svc *service.GuildService
}

// NewGuildHandler creates a new guild handler
func NewGuildHandler(svc *service.GuildService) *GuildHandler {
	return &GuildHandler{svc: svc}
}

// List handles GET /v1/guilds - list user's guilds
func (h *GuildHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	guilds, err := h.svc.ListUserGuilds(ctx, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, guilds, nil)
}

// Create handles POST /v1/guilds - create a new guild
func (h *GuildHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var req service.CreateGuildRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	guild, err := h.svc.CreateGuild(ctx, userID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, guild, nil)
}

// Get handles GET /v1/guilds/{guildId} - get guild details
func (h *GuildHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	guildData, err := h.svc.GetGuildWithMembers(ctx, userID, guildID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, guildData, nil)
}

// Update handles PATCH /v1/guilds/{guildId} - update a guild
func (h *GuildHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var req service.UpdateGuildRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	guild, err := h.svc.UpdateGuild(ctx, userID, guildID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, guild, nil)
}

// Delete handles DELETE /v1/guilds/{guildId} - delete a guild
func (h *GuildHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.svc.DeleteGuild(ctx, userID, guildID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Join handles POST /v1/guilds/{guildId}/join - join a guild
func (h *GuildHandler) Join(w http.ResponseWriter, r *http.Request) {
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

	if err := h.svc.JoinGuild(ctx, userID, guildID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Leave handles POST /v1/guilds/{guildId}/leave - leave a guild
func (h *GuildHandler) Leave(w http.ResponseWriter, r *http.Request) {
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

	if err := h.svc.LeaveGuild(ctx, userID, guildID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetMembers handles GET /v1/guilds/{guildId}/members - list guild members
func (h *GuildHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
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

	// GetGuildWithMembers already checks membership for private guilds
	guildData, err := h.svc.GetGuildWithMembers(ctx, userID, guildID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, guildData.Members, nil)
}

// GetMemberRole handles GET /v1/guilds/{guildId}/members/{userId}/role - get member's role
func (h *GuildHandler) GetMemberRole(w http.ResponseWriter, r *http.Request) {
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

	targetUserID := r.PathValue("userId")
	if targetUserID == "" {
		WriteError(w, model.NewBadRequestError("user ID required"))
		return
	}

	// Check that requester is a member
	isMember, err := h.svc.IsMember(ctx, userID, guildID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	if !isMember {
		WriteError(w, model.NewNotFoundError("guild not found"))
		return
	}

	role, err := h.svc.GetMemberRole(ctx, targetUserID, guildID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, map[string]string{"role": string(role)}, nil)
}

// UpdateMemberRole handles PATCH /v1/guilds/{guildId}/members/{userId}/role - update member's role
func (h *GuildHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
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

	targetUserID := r.PathValue("userId")
	if targetUserID == "" {
		WriteError(w, model.NewBadRequestError("user ID required"))
		return
	}

	var req model.UpdateMemberRoleRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	if !req.Role.IsValid() {
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "role", Message: "invalid role (must be member, moderator, or admin)"},
		}))
		return
	}

	if err := h.svc.UpdateMemberRole(ctx, userID, targetUserID, guildID, req.Role); err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, map[string]string{"role": string(req.Role)}, nil)
}

// handleError converts service errors to HTTP responses
func (h *GuildHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrGuildNotFound):
		WriteError(w, model.NewNotFoundError("guild not found"))
	case errors.Is(err, service.ErrGuildNameRequired):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "name", Message: "guild name is required"},
		}))
	case errors.Is(err, service.ErrGuildNameTooLong):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "name", Message: "guild name exceeds maximum length"},
		}))
	case errors.Is(err, service.ErrGuildDescTooLong):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "description", Message: "guild description exceeds maximum length"},
		}))
	case errors.Is(err, service.ErrNotGuildMember):
		WriteError(w, model.NewNotFoundError("guild not found")) // Don't reveal existence
	case errors.Is(err, service.ErrNotGuildAdmin):
		WriteError(w, model.NewForbiddenError("not authorized to perform this action"))
	case errors.Is(err, service.ErrCannotLeaveSoleMember):
		WriteError(w, model.NewConflictError("cannot leave guild as the only member"))
	case errors.Is(err, service.ErrAlreadyGuildMember):
		WriteError(w, model.NewConflictError("already a member of this guild"))
	case errors.Is(err, service.ErrMaxGuildsReached):
		WriteError(w, model.NewLimitExceededError("maximum number of guilds reached", model.MaxGuildsPerUser, model.MaxGuildsPerUser))
	case errors.Is(err, service.ErrMaxMembersReached):
		WriteError(w, model.NewLimitExceededError("guild has reached maximum member limit", model.MaxMembersPerGuild, model.MaxMembersPerGuild))
	case errors.Is(err, service.ErrGuildNameExists):
		WriteError(w, model.NewConflictError("a guild with this name already exists"))
	case errors.Is(err, service.ErrUserNotFound):
		WriteError(w, model.NewNotFoundError("user not found"))
	default:
		WriteError(w, model.NewInternalError("an unexpected error occurred"))
	}
}
