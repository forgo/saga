package handler

import (
	"net/http"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// AdminActionsHandler handles admin action endpoints for testing real-time events
type AdminActionsHandler struct {
	actionsService *service.AdminActionsService
}

// NewAdminActionsHandler creates a new admin actions handler
func NewAdminActionsHandler(actionsService *service.AdminActionsService) *AdminActionsHandler {
	return &AdminActionsHandler{actionsService: actionsService}
}

// UpdateLocation handles POST /v1/admin/actions/location
func (h *AdminActionsHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	var req service.UpdateLocationRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("Invalid request body: "+err.Error()))
		return
	}

	result, err := h.actionsService.UpdateLocation(r.Context(), req)
	if err != nil {
		WriteError(w, model.NewInternalError("Failed to update location: "+err.Error()))
		return
	}

	WriteData(w, http.StatusOK, result, nil)
}

// CreateTrustRating handles POST /v1/admin/actions/trust-rating
func (h *AdminActionsHandler) CreateTrustRating(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	var req service.CreateTrustRatingRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("Invalid request body: "+err.Error()))
		return
	}

	result, err := h.actionsService.CreateTrustRating(r.Context(), req)
	if err != nil {
		WriteError(w, model.NewInternalError("Failed to create trust rating: "+err.Error()))
		return
	}

	WriteData(w, http.StatusCreated, result, nil)
}

// JoinGuild handles POST /v1/admin/actions/guild-join
func (h *AdminActionsHandler) JoinGuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	var req service.JoinGuildRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("Invalid request body: "+err.Error()))
		return
	}

	result, err := h.actionsService.JoinGuild(r.Context(), req)
	if err != nil {
		WriteError(w, model.NewInternalError("Failed to join guild: "+err.Error()))
		return
	}

	WriteData(w, http.StatusCreated, result, nil)
}

// RSVP handles POST /v1/admin/actions/rsvp
func (h *AdminActionsHandler) RSVP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	var req service.RSVPRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("Invalid request body: "+err.Error()))
		return
	}

	result, err := h.actionsService.RSVP(r.Context(), req)
	if err != nil {
		WriteError(w, model.NewInternalError("Failed to RSVP: "+err.Error()))
		return
	}

	WriteData(w, http.StatusCreated, result, nil)
}

// CreateEvent handles POST /v1/admin/actions/event-create
func (h *AdminActionsHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	var req service.CreateEventRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("Invalid request body: "+err.Error()))
		return
	}

	result, err := h.actionsService.CreateEvent(r.Context(), req)
	if err != nil {
		WriteError(w, model.NewInternalError("Failed to create event: "+err.Error()))
		return
	}

	WriteData(w, http.StatusCreated, result, nil)
}

// GetUsers handles GET /v1/admin/actions/users
func (h *AdminActionsHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, model.NewMethodNotAllowedError("GET"))
		return
	}

	prefix := r.URL.Query().Get("prefix")

	users, err := h.actionsService.GetUsers(r.Context(), service.GetUsersRequest{
		Limit:  50,
		Prefix: prefix,
	})
	if err != nil {
		WriteError(w, model.NewInternalError("Failed to get users: "+err.Error()))
		return
	}

	WriteData(w, http.StatusOK, users, nil)
}

// GetGuilds handles GET /v1/admin/actions/guilds
func (h *AdminActionsHandler) GetGuilds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, model.NewMethodNotAllowedError("GET"))
		return
	}

	guilds, err := h.actionsService.GetGuilds(r.Context(), 50)
	if err != nil {
		WriteError(w, model.NewInternalError("Failed to get guilds: "+err.Error()))
		return
	}

	WriteData(w, http.StatusOK, guilds, nil)
}

// GetEvents handles GET /v1/admin/actions/events
func (h *AdminActionsHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, model.NewMethodNotAllowedError("GET"))
		return
	}

	events, err := h.actionsService.GetEvents(r.Context(), 50)
	if err != nil {
		WriteError(w, model.NewInternalError("Failed to get events: "+err.Error()))
		return
	}

	WriteData(w, http.StatusOK, events, nil)
}
