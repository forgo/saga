package handler

import (
	"net/http"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// AdventureHandler handles adventure HTTP requests
type AdventureHandler struct {
	svc *service.AdventureService
}

// NewAdventureHandler creates a new adventure handler
func NewAdventureHandler(svc *service.AdventureService) *AdventureHandler {
	return &AdventureHandler{svc: svc}
}

// Create handles POST /v1/adventures
func (h *AdventureHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var req model.CreateAdventureRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	adventure, err := h.svc.Create(ctx, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, adventure, nil)
}

// CreateGuildAdventure handles POST /v1/guilds/{guildId}/adventures
func (h *AdventureHandler) CreateGuildAdventure(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	guildID := r.PathValue("guildId")

	var req model.CreateAdventureRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	// Override to guild type
	orgType := string(model.AdventureOrganizerGuild)
	req.OrganizerType = &orgType
	req.GuildID = &guildID

	adventure, err := h.svc.Create(ctx, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, adventure, nil)
}

// CreateUserAdventure handles POST /v1/users/me/adventures
func (h *AdventureHandler) CreateUserAdventure(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var req model.CreateAdventureRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	// Override to user type
	orgType := string(model.AdventureOrganizerUser)
	req.OrganizerType = &orgType

	adventure, err := h.svc.Create(ctx, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, adventure, nil)
}

// GetByID handles GET /v1/adventures/{adventureId}
func (h *AdventureHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	adventureID := r.PathValue("adventureId")

	adventure, err := h.svc.GetByID(ctx, adventureID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, adventure, nil)
}

// Admission Endpoints

// RequestAdmission handles POST /v1/adventures/{adventureId}/admission/request
func (h *AdventureHandler) RequestAdmission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	adventureID := r.PathValue("adventureId")

	var req model.RequestAdmissionRequest
	// Request body is optional for simple admission requests
	_ = DecodeJSON(r, &req)

	admission, err := h.svc.RequestAdmission(ctx, adventureID, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, admission, nil)
}

// GetAdmission handles GET /v1/adventures/{adventureId}/admission
func (h *AdventureHandler) GetAdmission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	adventureID := r.PathValue("adventureId")

	admission, err := h.svc.GetAdmission(ctx, adventureID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, admission, nil)
}

// WithdrawAdmission handles DELETE /v1/adventures/{adventureId}/admission
func (h *AdventureHandler) WithdrawAdmission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	adventureID := r.PathValue("adventureId")

	if err := h.svc.WithdrawAdmission(ctx, adventureID, userID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Admission Management (Organizer)

// GetAdmissions handles GET /v1/adventures/{adventureId}/admissions
func (h *AdventureHandler) GetAdmissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	adventureID := r.PathValue("adventureId")

	limit, offset := getPaginationParams(r)

	var status *model.AdventureAdmissionStatus
	if v := r.URL.Query().Get("status"); v != "" {
		s := model.AdventureAdmissionStatus(v)
		status = &s
	}

	admissions, err := h.svc.GetAdmissions(ctx, adventureID, userID, status, limit, offset)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, admissions, nil, nil)
}

// GetPendingAdmissions handles GET /v1/adventures/{adventureId}/admissions/pending
func (h *AdventureHandler) GetPendingAdmissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	adventureID := r.PathValue("adventureId")

	admissions, err := h.svc.GetPendingAdmissions(ctx, adventureID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, admissions, nil, nil)
}

// RespondToAdmission handles POST /v1/adventures/{adventureId}/admissions/{userId}/respond
func (h *AdventureHandler) RespondToAdmission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	adventureID := r.PathValue("adventureId")
	targetUserID := r.PathValue("userId")

	var req model.RespondToAdmissionRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	admission, err := h.svc.RespondToAdmission(ctx, adventureID, userID, targetUserID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, admission, nil)
}

// InviteToAdventure handles POST /v1/adventures/{adventureId}/admissions/invite
func (h *AdventureHandler) InviteToAdventure(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	adventureID := r.PathValue("adventureId")

	var req model.InviteToAdventureRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	admission, err := h.svc.InviteToAdventure(ctx, adventureID, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, admission, nil)
}

// Organizer Management

// TransferAdventure handles POST /v1/adventures/{adventureId}/transfer
func (h *AdventureHandler) TransferAdventure(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	adventureID := r.PathValue("adventureId")

	var req model.TransferAdventureRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	adventure, err := h.svc.TransferAdventure(ctx, adventureID, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, adventure, nil)
}

// UnfreezeAdventure handles POST /v1/adventures/{adventureId}/unfreeze
func (h *AdventureHandler) UnfreezeAdventure(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	adventureID := r.PathValue("adventureId")

	var req model.UnfreezeAdventureRequest
	// Request body is optional
	_ = DecodeJSON(r, &req)

	adventure, err := h.svc.UnfreezeAdventure(ctx, adventureID, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, adventure, nil)
}

// CheckAdmission handles GET /v1/adventures/{adventureId}/admitted
func (h *AdventureHandler) CheckAdmission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	adventureID := r.PathValue("adventureId")

	isAdmitted, err := h.svc.IsAdmitted(ctx, adventureID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, map[string]bool{"admitted": isAdmitted}, nil)
}

// handleError converts service errors to HTTP responses
func (h *AdventureHandler) handleError(w http.ResponseWriter, err error) {
	if pd, ok := err.(*model.ProblemDetails); ok {
		WriteError(w, pd)
		return
	}
	WriteError(w, model.NewInternalError("internal server error"))
}
