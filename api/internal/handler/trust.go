package handler

import (
	"errors"
	"net/http"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// TrustHandler handles trust and IRL verification endpoints
type TrustHandler struct {
	trustService *service.TrustService
}

// NewTrustHandler creates a new trust handler
func NewTrustHandler(trustService *service.TrustService) *TrustHandler {
	return &TrustHandler{
		trustService: trustService,
	}
}

// GrantTrust handles POST /v1/trust/{userId} - grant trust to another user
func (h *TrustHandler) GrantTrust(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	targetUserID := r.PathValue("userId")
	if targetUserID == "" {
		WriteError(w, model.NewBadRequestError("user ID required"))
		return
	}

	trust, err := h.trustService.GrantTrust(r.Context(), userID, targetUserID)
	if err != nil {
		h.handleTrustError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, trust, map[string]string{
		"self": "/v1/trust/" + targetUserID,
	})
}

// RevokeTrust handles DELETE /v1/trust/{userId} - revoke trust from another user
func (h *TrustHandler) RevokeTrust(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	targetUserID := r.PathValue("userId")
	if targetUserID == "" {
		WriteError(w, model.NewBadRequestError("user ID required"))
		return
	}

	if err := h.trustService.RevokeTrust(r.Context(), userID, targetUserID); err != nil {
		h.handleTrustError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetTrustSummary handles GET /v1/trust/{userId} - get trust summary with another user
func (h *TrustHandler) GetTrustSummary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	targetUserID := r.PathValue("userId")
	if targetUserID == "" {
		WriteError(w, model.NewBadRequestError("user ID required"))
		return
	}

	summary, err := h.trustService.GetTrustSummary(r.Context(), userID, targetUserID)
	if err != nil {
		h.handleTrustError(w, err)
		return
	}

	WriteData(w, http.StatusOK, summary, map[string]string{
		"self": "/v1/trust/" + targetUserID,
	})
}

// GetTrustedUsers handles GET /v1/trust - get list of trusted users
func (h *TrustHandler) GetTrustedUsers(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	users, err := h.trustService.GetTrustedUsers(r.Context(), userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get trusted users"))
		return
	}

	WriteCollection(w, http.StatusOK, users, nil, map[string]string{
		"self": "/v1/trust",
	})
}

// GetTrustProfile handles GET /v1/profile/trust - get own trust profile
func (h *TrustHandler) GetTrustProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	profile, err := h.trustService.GetTrustProfile(r.Context(), userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get trust profile"))
		return
	}

	WriteData(w, http.StatusOK, profile, map[string]string{
		"self": "/v1/profile/trust",
	})
}

// ConfirmIRL handles POST /v1/irl/{userId} - confirm IRL meeting with another user
func (h *TrustHandler) ConfirmIRL(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	targetUserID := r.PathValue("userId")
	if targetUserID == "" {
		WriteError(w, model.NewBadRequestError("user ID required"))
		return
	}

	var req model.ConfirmIRLRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	// Set the target user ID from path
	req.UserID = targetUserID

	// Default context if not provided
	if req.Context == "" {
		req.Context = model.IRLContextOther
	}

	irl, err := h.trustService.ConfirmIRL(r.Context(), userID, &req)
	if err != nil {
		h.handleTrustError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, irl, map[string]string{
		"self": "/v1/irl/" + targetUserID,
	})
}

// GetIRLConnections handles GET /v1/irl - get list of IRL connections
func (h *TrustHandler) GetIRLConnections(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	connections, err := h.trustService.GetIRLConnections(r.Context(), userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get IRL connections"))
		return
	}

	WriteCollection(w, http.StatusOK, connections, nil, map[string]string{
		"self": "/v1/irl",
	})
}

func (h *TrustHandler) handleTrustError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrTrustNotFound):
		WriteError(w, model.NewNotFoundError("trust relation"))
	case errors.Is(err, service.ErrIRLNotFound):
		WriteError(w, model.NewNotFoundError("IRL verification"))
	case errors.Is(err, service.ErrCannotTrustSelf):
		WriteError(w, model.NewBadRequestError("cannot trust yourself"))
	case errors.Is(err, service.ErrAlreadyTrusted):
		WriteError(w, model.NewConflictError("already trusted"))
	case errors.Is(err, service.ErrInvalidContext):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "context", Message: "invalid IRL context"},
		}))
	default:
		WriteError(w, model.NewInternalError("trust operation failed"))
	}
}
