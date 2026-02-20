package handler

import (
	"net/http"
	"strconv"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// AdminDiscoveryHandler handles admin discovery lab endpoints
type AdminDiscoveryHandler struct {
	discoveryService *service.AdminDiscoveryService
}

// NewAdminDiscoveryHandler creates a new admin discovery handler
func NewAdminDiscoveryHandler(discoveryService *service.AdminDiscoveryService) *AdminDiscoveryHandler {
	return &AdminDiscoveryHandler{discoveryService: discoveryService}
}

// GetUsersWithLocations handles GET /v1/admin/discovery/users
func (h *AdminDiscoveryHandler) GetUsersWithLocations(w http.ResponseWriter, r *http.Request) {
	limit := 200
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	users, err := h.discoveryService.GetUsersWithLocations(r.Context(), limit)
	if err != nil {
		WriteError(w, model.NewInternalError("Failed to get users with locations: "+err.Error()))
		return
	}

	WriteData(w, http.StatusOK, users, nil)
}

// SimulateDiscovery handles POST /v1/admin/discovery/simulate
func (h *AdminDiscoveryHandler) SimulateDiscovery(w http.ResponseWriter, r *http.Request) {
	var req service.AdminDiscoveryRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("Invalid request body: "+err.Error()))
		return
	}

	if req.ViewerID == "" {
		WriteError(w, model.NewBadRequestError("viewer_id is required"))
		return
	}

	result, err := h.discoveryService.SimulateDiscovery(r.Context(), req)
	if err != nil {
		WriteError(w, model.NewInternalError("Discovery simulation failed: "+err.Error()))
		return
	}

	WriteData(w, http.StatusOK, result, nil)
}

// GetCompatibility handles GET /v1/admin/discovery/compatibility/{userAId}/{userBId}
func (h *AdminDiscoveryHandler) GetCompatibility(w http.ResponseWriter, r *http.Request) {
	userAID := r.PathValue("userAId")
	userBID := r.PathValue("userBId")

	if userAID == "" || userBID == "" {
		WriteError(w, model.NewBadRequestError("Both userAId and userBId are required"))
		return
	}

	result, err := h.discoveryService.GetCompatibility(r.Context(), userAID, userBID)
	if err != nil {
		WriteError(w, model.NewInternalError("Failed to get compatibility: "+err.Error()))
		return
	}

	WriteData(w, http.StatusOK, result, nil)
}
