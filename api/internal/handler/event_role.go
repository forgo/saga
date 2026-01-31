package handler

import (
	"errors"
	"net/http"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// EventRoleHandler handles event role endpoints
type EventRoleHandler struct {
	eventRoleService *service.EventRoleService
}

// NewEventRoleHandler creates a new event role handler
func NewEventRoleHandler(eventRoleService *service.EventRoleService) *EventRoleHandler {
	return &EventRoleHandler{
		eventRoleService: eventRoleService,
	}
}

// CreateRole handles POST /v1/events/{eventId}/roles - create a new role (host only)
func (h *EventRoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	eventID := r.PathValue("eventId")
	if eventID == "" {
		WriteError(w, model.NewBadRequestError("event ID required"))
		return
	}

	var req model.CreateEventRoleRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	// Validate
	var fieldErrors []model.FieldError
	if req.Name == "" {
		fieldErrors = append(fieldErrors, model.FieldError{
			Field:   "name",
			Message: "name is required",
		})
	}
	if len(fieldErrors) > 0 {
		WriteError(w, model.NewValidationError(fieldErrors))
		return
	}

	role, err := h.eventRoleService.CreateRole(r.Context(), eventID, userID, &req)
	if err != nil {
		h.handleEventRoleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, role, map[string]string{
		"self": "/v1/events/" + eventID + "/roles/" + role.ID,
	})
}

// GetRoles handles GET /v1/events/{eventId}/roles - list roles for an event
func (h *EventRoleHandler) GetRoles(w http.ResponseWriter, r *http.Request) {
	eventID := r.PathValue("eventId")
	if eventID == "" {
		WriteError(w, model.NewBadRequestError("event ID required"))
		return
	}

	roles, err := h.eventRoleService.GetEventRoles(r.Context(), eventID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get roles"))
		return
	}

	WriteCollection(w, http.StatusOK, roles, nil, map[string]string{
		"self": "/v1/events/" + eventID + "/roles",
	})
}

// GetRolesOverview handles GET /v1/events/{eventId}/roles/overview - get roles with assignments
func (h *EventRoleHandler) GetRolesOverview(w http.ResponseWriter, r *http.Request) {
	eventID := r.PathValue("eventId")
	if eventID == "" {
		WriteError(w, model.NewBadRequestError("event ID required"))
		return
	}

	overview, err := h.eventRoleService.GetEventRolesOverview(r.Context(), eventID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get roles overview"))
		return
	}

	WriteData(w, http.StatusOK, overview, map[string]string{
		"self": "/v1/events/" + eventID + "/roles/overview",
	})
}

// UpdateRole handles PATCH /v1/events/{eventId}/roles/{roleId} - update a role (host only)
func (h *EventRoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	roleID := r.PathValue("roleId")
	if roleID == "" {
		WriteError(w, model.NewBadRequestError("role ID required"))
		return
	}

	var req model.UpdateEventRoleRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	role, err := h.eventRoleService.UpdateRole(r.Context(), roleID, &req)
	if err != nil {
		h.handleEventRoleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, role, nil)
}

// DeleteRole handles DELETE /v1/events/{eventId}/roles/{roleId} - delete a role (host only)
func (h *EventRoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	roleID := r.PathValue("roleId")
	if roleID == "" {
		WriteError(w, model.NewBadRequestError("role ID required"))
		return
	}

	if err := h.eventRoleService.DeleteRole(r.Context(), roleID); err != nil {
		h.handleEventRoleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AssignRole handles POST /v1/events/{eventId}/roles/assign - assign self to a role
func (h *EventRoleHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var req model.AssignRoleRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	if req.RoleID == "" {
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "role_id", Message: "role_id is required"},
		}))
		return
	}

	assignment, err := h.eventRoleService.AssignRole(r.Context(), userID, &req)
	if err != nil {
		h.handleEventRoleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, assignment, nil)
}

// GetMyRoles handles GET /v1/events/{eventId}/roles/mine - get my roles for an event
func (h *EventRoleHandler) GetMyRoles(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	eventID := r.PathValue("eventId")
	if eventID == "" {
		WriteError(w, model.NewBadRequestError("event ID required"))
		return
	}

	roles, err := h.eventRoleService.GetUserRoles(r.Context(), eventID, userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get roles"))
		return
	}

	WriteData(w, http.StatusOK, roles, map[string]string{
		"self": "/v1/events/" + eventID + "/roles/mine",
	})
}

// GetRoleSuggestions handles GET /v1/events/{eventId}/roles/suggestions - get role suggestions based on interests
func (h *EventRoleHandler) GetRoleSuggestions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	eventID := r.PathValue("eventId")
	if eventID == "" {
		WriteError(w, model.NewBadRequestError("event ID required"))
		return
	}

	suggestions, err := h.eventRoleService.GetRoleSuggestions(r.Context(), eventID, userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get suggestions"))
		return
	}

	WriteCollection(w, http.StatusOK, suggestions, nil, map[string]string{
		"self": "/v1/events/" + eventID + "/roles/suggestions",
	})
}

// CancelAssignment handles DELETE /v1/events/{eventId}/roles/assignments/{assignmentId} - cancel own assignment
func (h *EventRoleHandler) CancelAssignment(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	assignmentID := r.PathValue("assignmentId")
	if assignmentID == "" {
		WriteError(w, model.NewBadRequestError("assignment ID required"))
		return
	}

	if err := h.eventRoleService.CancelAssignment(r.Context(), userID, assignmentID); err != nil {
		h.handleEventRoleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *EventRoleHandler) handleEventRoleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrRoleNotFound):
		WriteError(w, model.NewNotFoundError("role"))
	case errors.Is(err, service.ErrAssignmentNotFound):
		WriteError(w, model.NewNotFoundError("assignment"))
	case errors.Is(err, service.ErrRoleFull):
		WriteError(w, model.NewConflictError("role is full"))
	case errors.Is(err, service.ErrAlreadyAssignedToRole):
		WriteError(w, model.NewConflictError("already assigned to this role"))
	case errors.Is(err, service.ErrCannotDeleteDefault):
		WriteError(w, model.NewBadRequestError("cannot delete default role"))
	case errors.Is(err, service.ErrMaxRolesReached):
		WriteError(w, model.NewBadRequestError("maximum roles reached"))
	case errors.Is(err, service.ErrMaxRolesPerUserReached):
		WriteError(w, model.NewBadRequestError("maximum roles per user reached"))
	case errors.Is(err, service.ErrCannotAssignOthers):
		WriteError(w, model.NewForbiddenError("cannot assign roles to others"))
	default:
		WriteError(w, model.NewInternalError("event role operation failed"))
	}
}
