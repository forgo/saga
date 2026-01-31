package handler

import (
	"net/http"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// RoleCatalogHandler handles role catalog HTTP requests
type RoleCatalogHandler struct {
	svc *service.RoleCatalogService
}

// NewRoleCatalogHandler creates a new role catalog handler
func NewRoleCatalogHandler(svc *service.RoleCatalogService) *RoleCatalogHandler {
	return &RoleCatalogHandler{svc: svc}
}

// Guild Catalog Endpoints

// CreateGuildCatalog handles POST /v1/guilds/{guildId}/role-catalogs
func (h *RoleCatalogHandler) CreateGuildCatalog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	guildID := r.PathValue("guildId")

	var req model.CreateRoleCatalogRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	catalog, err := h.svc.CreateGuildCatalog(ctx, guildID, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, catalog, nil)
}

// GetGuildCatalogs handles GET /v1/guilds/{guildId}/role-catalogs
func (h *RoleCatalogHandler) GetGuildCatalogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	guildID := r.PathValue("guildId")

	var roleType *string
	if v := r.URL.Query().Get("role_type"); v != "" {
		roleType = &v
	}

	catalogs, err := h.svc.GetGuildCatalogs(ctx, guildID, roleType)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, catalogs, nil, nil)
}

// User Catalog Endpoints

// CreateUserCatalog handles POST /v1/users/me/role-catalogs
func (h *RoleCatalogHandler) CreateUserCatalog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var req model.CreateRoleCatalogRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	catalog, err := h.svc.CreateUserCatalog(ctx, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, catalog, nil)
}

// GetUserCatalogs handles GET /v1/users/me/role-catalogs
func (h *RoleCatalogHandler) GetUserCatalogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var roleType *string
	if v := r.URL.Query().Get("role_type"); v != "" {
		roleType = &v
	}

	catalogs, err := h.svc.GetUserCatalogs(ctx, userID, roleType)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, catalogs, nil, nil)
}

// Common Catalog Endpoints

// GetCatalogByID handles GET /v1/role-catalogs/{catalogId}
func (h *RoleCatalogHandler) GetCatalogByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	catalogID := r.PathValue("catalogId")

	catalog, err := h.svc.GetCatalogByID(ctx, catalogID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, catalog, nil)
}

// UpdateCatalog handles PATCH /v1/role-catalogs/{catalogId}
func (h *RoleCatalogHandler) UpdateCatalog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	catalogID := r.PathValue("catalogId")

	var req model.UpdateRoleCatalogRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	catalog, err := h.svc.UpdateCatalog(ctx, catalogID, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, catalog, nil)
}

// DeleteCatalog handles DELETE /v1/role-catalogs/{catalogId}
func (h *RoleCatalogHandler) DeleteCatalog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	catalogID := r.PathValue("catalogId")

	if err := h.svc.DeleteCatalog(ctx, catalogID, userID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Rideshare Role Endpoints

// CreateRideshareRole handles POST /v1/rideshares/{rideshareId}/roles
func (h *RoleCatalogHandler) CreateRideshareRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	rideshareID := r.PathValue("rideshareId")

	var req model.CreateRideshareRoleRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	role, err := h.svc.CreateRideshareRole(ctx, rideshareID, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, role, nil)
}

// GetRideshareRoles handles GET /v1/rideshares/{rideshareId}/roles
func (h *RoleCatalogHandler) GetRideshareRoles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rideshareID := r.PathValue("rideshareId")

	roles, err := h.svc.GetRideshareRoles(ctx, rideshareID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, roles, nil, nil)
}

// GetRideshareRolesWithAssignments handles GET /v1/rideshares/{rideshareId}/roles/detailed
func (h *RoleCatalogHandler) GetRideshareRolesWithAssignments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rideshareID := r.PathValue("rideshareId")

	roles, err := h.svc.GetRideshareRolesWithAssignments(ctx, rideshareID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, roles, nil, nil)
}

// UpdateRideshareRole handles PATCH /v1/rideshares/{rideshareId}/roles/{roleId}
func (h *RoleCatalogHandler) UpdateRideshareRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	roleID := r.PathValue("roleId")

	var req model.UpdateRideshareRoleRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	role, err := h.svc.UpdateRideshareRole(ctx, roleID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, role, nil)
}

// DeleteRideshareRole handles DELETE /v1/rideshares/{rideshareId}/roles/{roleId}
func (h *RoleCatalogHandler) DeleteRideshareRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	roleID := r.PathValue("roleId")

	if err := h.svc.DeleteRideshareRole(ctx, roleID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AssignRideshareRole handles POST /v1/rideshares/{rideshareId}/roles/assign
func (h *RoleCatalogHandler) AssignRideshareRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	rideshareID := r.PathValue("rideshareId")

	var req model.AssignRideshareRoleRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	assignment, err := h.svc.AssignRideshareRole(ctx, rideshareID, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, assignment, nil)
}

// UnassignRideshareRole handles DELETE /v1/rideshares/{rideshareId}/roles/assignments/{assignmentId}
func (h *RoleCatalogHandler) UnassignRideshareRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	assignmentID := r.PathValue("assignmentId")

	if err := h.svc.UnassignRideshareRole(ctx, assignmentID, userID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetUserRideshareRoles handles GET /v1/rideshares/{rideshareId}/my-roles
func (h *RoleCatalogHandler) GetUserRideshareRoles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	rideshareID := r.PathValue("rideshareId")

	assignments, err := h.svc.GetUserRideshareRoles(ctx, rideshareID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, assignments, nil, nil)
}

// handleError converts service errors to HTTP responses
func (h *RoleCatalogHandler) handleError(w http.ResponseWriter, err error) {
	if pd, ok := err.(*model.ProblemDetails); ok {
		WriteError(w, pd)
		return
	}
	WriteError(w, model.NewInternalError("internal server error"))
}
