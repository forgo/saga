package handler

import (
	"net/http"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// AdminSeederHandler handles admin seeding endpoints
type AdminSeederHandler struct {
	seederService *service.SeederService
}

// NewAdminSeederHandler creates a new admin seeder handler
func NewAdminSeederHandler(seederService *service.SeederService) *AdminSeederHandler {
	return &AdminSeederHandler{seederService: seederService}
}

// SeedUsers handles POST /v1/admin/seed/users
func (h *AdminSeederHandler) SeedUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	var req service.SeedUsersRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("Invalid request body: "+err.Error()))
		return
	}

	if req.Count <= 0 {
		WriteError(w, model.NewBadRequestError("count must be greater than 0"))
		return
	}

	result, err := h.seederService.SeedUsers(r.Context(), req)
	if err != nil {
		WriteError(w, model.NewInternalError("Failed to seed users: "+err.Error()))
		return
	}

	WriteData(w, http.StatusCreated, result, map[string]string{
		"self":    "/v1/admin/seed/users",
		"cleanup": "/v1/admin/seed/cleanup",
	})
}

// SeedGuilds handles POST /v1/admin/seed/guilds
func (h *AdminSeederHandler) SeedGuilds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	var req service.SeedGuildsRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("Invalid request body: "+err.Error()))
		return
	}

	if req.Count <= 0 {
		WriteError(w, model.NewBadRequestError("count must be greater than 0"))
		return
	}

	result, err := h.seederService.SeedGuilds(r.Context(), req)
	if err != nil {
		WriteError(w, model.NewInternalError("Failed to seed guilds: "+err.Error()))
		return
	}

	WriteData(w, http.StatusCreated, result, map[string]string{
		"self":    "/v1/admin/seed/guilds",
		"cleanup": "/v1/admin/seed/cleanup",
	})
}

// SeedEvents handles POST /v1/admin/seed/events
func (h *AdminSeederHandler) SeedEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	var req service.SeedEventsRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("Invalid request body: "+err.Error()))
		return
	}

	if req.Count <= 0 {
		WriteError(w, model.NewBadRequestError("count must be greater than 0"))
		return
	}

	result, err := h.seederService.SeedEvents(r.Context(), req)
	if err != nil {
		WriteError(w, model.NewInternalError("Failed to seed events: "+err.Error()))
		return
	}

	WriteData(w, http.StatusCreated, result, map[string]string{
		"self":    "/v1/admin/seed/events",
		"cleanup": "/v1/admin/seed/cleanup",
	})
}

// SeedScenario handles POST /v1/admin/seed/scenario
func (h *AdminSeederHandler) SeedScenario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	var req service.SeedScenarioRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("Invalid request body: "+err.Error()))
		return
	}

	if req.Scenario == "" {
		WriteError(w, model.NewBadRequestError("scenario is required"))
		return
	}

	result, err := h.seederService.SeedScenario(r.Context(), req)
	if err != nil {
		WriteError(w, model.NewInternalError("Failed to run scenario: "+err.Error()))
		return
	}

	WriteData(w, http.StatusCreated, result, map[string]string{
		"self":    "/v1/admin/seed/scenario",
		"cleanup": "/v1/admin/seed/cleanup",
	})
}

// CleanupRequest defines the cleanup request body
type CleanupRequest struct {
	Prefix string `json:"prefix,omitempty"`
}

// Cleanup handles DELETE /v1/admin/seed/cleanup
func (h *AdminSeederHandler) Cleanup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		WriteError(w, model.NewMethodNotAllowedError("DELETE"))
		return
	}

	// Prefix can come from query param or request body
	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		// Try to read from body (optional)
		var req CleanupRequest
		_ = DecodeJSON(r, &req) // Ignore error, body is optional
		prefix = req.Prefix
	}

	if prefix == "" {
		prefix = "seed_"
	}

	result, err := h.seederService.Cleanup(r.Context(), prefix)
	if err != nil {
		WriteError(w, model.NewInternalError("Failed to cleanup: "+err.Error()))
		return
	}

	WriteData(w, http.StatusOK, result, map[string]string{
		"self": "/v1/admin/seed/cleanup",
	})
}

// ListScenarios handles GET /v1/admin/seed/scenarios
func (h *AdminSeederHandler) ListScenarios(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, model.NewMethodNotAllowedError("GET"))
		return
	}

	scenarios := []map[string]string{
		{
			"id":          "sf_discovery_pool",
			"name":        "SF Discovery Pool",
			"description": "20 users in San Francisco for discovery testing",
		},
		{
			"id":          "active_guild",
			"name":        "Active Guild",
			"description": "A guild with 10 members and 5 upcoming events",
		},
		{
			"id":          "event_with_attendees",
			"name":        "Event with Attendees",
			"description": "An event with 20 attendees for RSVP testing",
		},
	}

	WriteData(w, http.StatusOK, scenarios, map[string]string{
		"self": "/v1/admin/seed/scenarios",
	})
}
