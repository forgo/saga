package handler

import (
	"net/http"
	"strconv"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// AdminUsersHandler handles admin user management endpoints
type AdminUsersHandler struct {
	usersService *service.AdminUsersService
}

// NewAdminUsersHandler creates a new admin users handler
func NewAdminUsersHandler(usersService *service.AdminUsersService) *AdminUsersHandler {
	return &AdminUsersHandler{usersService: usersService}
}

// ListUsers handles GET /v1/admin/users
func (h *AdminUsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))

	req := service.ListUsersRequest{
		Page:     page,
		PageSize: pageSize,
		Search:   q.Get("search"),
		Role:     q.Get("role"),
		SortBy:   q.Get("sort_by"),
		SortDir:  q.Get("sort_dir"),
	}

	result, err := h.usersService.ListUsers(r.Context(), req)
	if err != nil {
		WriteError(w, model.NewInternalError("Failed to list users: "+err.Error()))
		return
	}

	WriteData(w, http.StatusOK, result, nil)
}

// GetUser handles GET /v1/admin/users/{userId}
func (h *AdminUsersHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		WriteError(w, model.NewBadRequestError("userId is required"))
		return
	}

	result, err := h.usersService.GetUserDetail(r.Context(), userID)
	if err != nil {
		if err == service.ErrUserNotFound {
			WriteError(w, model.NewNotFoundError("User"))
			return
		}
		WriteError(w, model.NewInternalError("Failed to get user: "+err.Error()))
		return
	}

	WriteData(w, http.StatusOK, result, nil)
}

// UpdateRole handles PATCH /v1/admin/users/{userId}/role
func (h *AdminUsersHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		WriteError(w, model.NewBadRequestError("userId is required"))
		return
	}

	var req service.UpdateRoleRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("Invalid request body: "+err.Error()))
		return
	}

	// Validate role enum
	role := model.UserRole(req.Role)
	switch role {
	case model.UserRoleUser, model.UserRoleModerator, model.UserRoleAdmin:
		// valid
	default:
		WriteError(w, model.NewBadRequestError("Invalid role. Must be one of: user, moderator, admin"))
		return
	}

	adminUserID := middleware.GetUserID(r.Context())

	if err := h.usersService.UpdateUserRole(r.Context(), adminUserID, userID, role); err != nil {
		if err == service.ErrUserNotFound {
			WriteError(w, model.NewNotFoundError("User"))
			return
		}
		WriteError(w, model.NewBadRequestError(err.Error()))
		return
	}

	WriteNoContent(w)
}

// DeleteUser handles DELETE /v1/admin/users/{userId}
func (h *AdminUsersHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		WriteError(w, model.NewBadRequestError("userId is required"))
		return
	}

	hard := r.URL.Query().Get("hard") == "true"
	adminUserID := middleware.GetUserID(r.Context())

	if err := h.usersService.DeleteUser(r.Context(), adminUserID, userID, hard); err != nil {
		if err == service.ErrUserNotFound {
			WriteError(w, model.NewNotFoundError("User"))
			return
		}
		WriteError(w, model.NewBadRequestError(err.Error()))
		return
	}

	WriteNoContent(w)
}
