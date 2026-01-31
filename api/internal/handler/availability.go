package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// AvailabilityHandler handles availability endpoints
type AvailabilityHandler struct {
	availabilityService *service.AvailabilityService
	profileService      *service.ProfileService
}

// NewAvailabilityHandler creates a new availability handler
func NewAvailabilityHandler(
	availabilityService *service.AvailabilityService,
	profileService *service.ProfileService,
) *AvailabilityHandler {
	return &AvailabilityHandler{
		availabilityService: availabilityService,
		profileService:      profileService,
	}
}

// CreateAvailability handles POST /v1/availability - create availability window
func (h *AvailabilityHandler) CreateAvailability(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var req model.CreateAvailabilityRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	// Validate
	var fieldErrors []model.FieldError
	if req.StartTime == "" {
		fieldErrors = append(fieldErrors, model.FieldError{
			Field:   "start_time",
			Message: "start_time is required",
		})
	}
	if req.EndTime == "" {
		fieldErrors = append(fieldErrors, model.FieldError{
			Field:   "end_time",
			Message: "end_time is required",
		})
	}
	if req.HangoutType == "" {
		fieldErrors = append(fieldErrors, model.FieldError{
			Field:   "hangout_type",
			Message: "hangout_type is required",
		})
	}
	if len(fieldErrors) > 0 {
		WriteError(w, model.NewValidationError(fieldErrors))
		return
	}

	availability, err := h.availabilityService.CreateAvailability(r.Context(), userID, &req)
	if err != nil {
		h.handleAvailabilityError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, availability, map[string]string{
		"self": "/v1/availability/" + availability.ID,
	})
}

// GetAvailability handles GET /v1/availability/{availabilityId} - get an availability
func (h *AvailabilityHandler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	availabilityID := r.PathValue("availabilityId")
	if availabilityID == "" {
		WriteError(w, model.NewBadRequestError("availability ID required"))
		return
	}

	availability, err := h.availabilityService.GetAvailability(r.Context(), availabilityID)
	if err != nil {
		h.handleAvailabilityError(w, err)
		return
	}

	WriteData(w, http.StatusOK, availability, map[string]string{
		"self": "/v1/availability/" + availabilityID,
	})
}

// GetMyAvailabilities handles GET /v1/profile/availability - get own availabilities
func (h *AvailabilityHandler) GetMyAvailabilities(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	availabilities, err := h.availabilityService.GetUserAvailabilities(r.Context(), userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get availabilities"))
		return
	}

	WriteCollection(w, http.StatusOK, availabilities, nil, map[string]string{
		"self": "/v1/profile/availability",
	})
}

// UpdateAvailability handles PATCH /v1/availability/{availabilityId} - update availability
func (h *AvailabilityHandler) UpdateAvailability(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	availabilityID := r.PathValue("availabilityId")
	if availabilityID == "" {
		WriteError(w, model.NewBadRequestError("availability ID required"))
		return
	}

	var req model.UpdateAvailabilityRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	availability, err := h.availabilityService.UpdateAvailability(r.Context(), userID, availabilityID, &req)
	if err != nil {
		h.handleAvailabilityError(w, err)
		return
	}

	WriteData(w, http.StatusOK, availability, map[string]string{
		"self": "/v1/availability/" + availabilityID,
	})
}

// DeleteAvailability handles DELETE /v1/availability/{availabilityId}
func (h *AvailabilityHandler) DeleteAvailability(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	availabilityID := r.PathValue("availabilityId")
	if availabilityID == "" {
		WriteError(w, model.NewBadRequestError("availability ID required"))
		return
	}

	if err := h.availabilityService.DeleteAvailability(r.Context(), userID, availabilityID); err != nil {
		h.handleAvailabilityError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// FindNearby handles GET /v1/discover/availability - find nearby available people
func (h *AvailabilityHandler) FindNearby(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	// Get user's location
	location, err := h.profileService.GetLocationInternal(r.Context(), userID)
	if err != nil || location == nil {
		WriteError(w, model.NewBadRequestError("location required to discover nearby availability"))
		return
	}

	// Parse query params
	radiusKm := 25.0
	if r.URL.Query().Get("radius_km") != "" {
		if radius, err := strconv.ParseFloat(r.URL.Query().Get("radius_km"), 64); err == nil {
			radiusKm = radius
		}
	}

	// Time window defaults to now + 24 hours
	startTime := time.Now()
	endTime := time.Now().Add(24 * time.Hour)

	if r.URL.Query().Get("start_time") != "" {
		if t, err := time.Parse(time.RFC3339, r.URL.Query().Get("start_time")); err == nil {
			startTime = t
		}
	}
	if r.URL.Query().Get("end_time") != "" {
		if t, err := time.Parse(time.RFC3339, r.URL.Query().Get("end_time")); err == nil {
			endTime = t
		}
	}

	limit := 20
	if r.URL.Query().Get("limit") != "" {
		if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	availabilities, err := h.availabilityService.FindNearbyAvailabilities(
		r.Context(), userID, location.Lat, location.Lng, radiusKm, startTime, endTime, limit,
	)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to find nearby availability"))
		return
	}

	WriteCollection(w, http.StatusOK, availabilities, nil, map[string]string{
		"self": "/v1/discover/availability",
	})
}

// FindByType handles GET /v1/discover/availability/type/{type} - find by hangout type
func (h *AvailabilityHandler) FindByType(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	hangoutType := r.PathValue("type")
	if hangoutType == "" {
		WriteError(w, model.NewBadRequestError("hangout type required"))
		return
	}

	limit := 20
	if r.URL.Query().Get("limit") != "" {
		if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	availabilities, err := h.availabilityService.FindByHangoutType(r.Context(), userID, hangoutType, limit)
	if err != nil {
		h.handleAvailabilityError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, availabilities, nil, map[string]string{
		"self": "/v1/discover/availability/type/" + hangoutType,
	})
}

// RequestHangout handles POST /v1/availability/{availabilityId}/request - request to join
func (h *AvailabilityHandler) RequestHangout(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	availabilityID := r.PathValue("availabilityId")
	if availabilityID == "" {
		WriteError(w, model.NewBadRequestError("availability ID required"))
		return
	}

	var req struct {
		Note string `json:"note"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	// Validate note length
	if len(req.Note) < model.MinHangoutNoteLength {
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "note", Message: "note must be at least 20 characters"},
		}))
		return
	}

	hangoutRequest, err := h.availabilityService.RequestHangout(r.Context(), userID, availabilityID, req.Note)
	if err != nil {
		h.handleAvailabilityError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, hangoutRequest, map[string]string{
		"self": "/v1/availability/" + availabilityID + "/request",
	})
}

// GetPendingRequests handles GET /v1/availability/{availabilityId}/requests - get pending requests
func (h *AvailabilityHandler) GetPendingRequests(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	availabilityID := r.PathValue("availabilityId")
	if availabilityID == "" {
		WriteError(w, model.NewBadRequestError("availability ID required"))
		return
	}

	requests, err := h.availabilityService.GetPendingRequests(r.Context(), userID, availabilityID)
	if err != nil {
		h.handleAvailabilityError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, requests, nil, map[string]string{
		"self": "/v1/availability/" + availabilityID + "/requests",
	})
}

// RespondToRequest handles POST /v1/requests/{requestId}/respond - accept or decline
func (h *AvailabilityHandler) RespondToRequest(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	requestID := r.PathValue("requestId")
	if requestID == "" {
		WriteError(w, model.NewBadRequestError("request ID required"))
		return
	}

	var req struct {
		Accept bool `json:"accept"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	hangout, err := h.availabilityService.RespondToRequest(r.Context(), userID, requestID, req.Accept)
	if err != nil {
		h.handleAvailabilityError(w, err)
		return
	}

	if hangout != nil {
		WriteData(w, http.StatusOK, hangout, map[string]string{
			"self": "/v1/hangouts/" + hangout.ID,
		})
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

// GetHangoutTypes handles GET /v1/hangout-types - list hangout types
func (h *AvailabilityHandler) GetHangoutTypes(w http.ResponseWriter, r *http.Request) {
	types := model.GetHangoutTypeInfo()
	WriteCollection(w, http.StatusOK, types, nil, map[string]string{
		"self": "/v1/hangout-types",
	})
}

// GetUserHangouts handles GET /v1/profile/hangouts - get user's hangouts
func (h *AvailabilityHandler) GetUserHangouts(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	limit := 20
	if r.URL.Query().Get("limit") != "" {
		if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	hangouts, err := h.availabilityService.GetUserHangouts(r.Context(), userID, limit)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get hangouts"))
		return
	}

	WriteCollection(w, http.StatusOK, hangouts, nil, map[string]string{
		"self": "/v1/profile/hangouts",
	})
}

// UpdateHangoutStatus handles PATCH /v1/hangouts/{hangoutId}/status
func (h *AvailabilityHandler) UpdateHangoutStatus(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	hangoutID := r.PathValue("hangoutId")
	if hangoutID == "" {
		WriteError(w, model.NewBadRequestError("hangout ID required"))
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	if err := h.availabilityService.UpdateHangoutStatus(r.Context(), userID, hangoutID, req.Status); err != nil {
		h.handleAvailabilityError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AvailabilityHandler) handleAvailabilityError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrAvailabilityNotFound):
		WriteError(w, model.NewNotFoundError("availability"))
	case errors.Is(err, service.ErrHangoutRequestNotFound):
		WriteError(w, model.NewNotFoundError("hangout request"))
	case errors.Is(err, service.ErrHangoutNotFound):
		WriteError(w, model.NewNotFoundError("hangout"))
	case errors.Is(err, service.ErrInvalidHangoutType):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "hangout_type", Message: "invalid hangout type"},
		}))
	case errors.Is(err, service.ErrInvalidTimeRange):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "end_time", Message: "end time must be after start time"},
		}))
	case errors.Is(err, service.ErrNoteTooShort):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "note", Message: "note must be at least 20 characters"},
		}))
	case errors.Is(err, service.ErrAlreadyRequested):
		WriteError(w, model.NewConflictError("already requested this hangout"))
	case errors.Is(err, service.ErrCannotRequestOwn):
		WriteError(w, model.NewBadRequestError("cannot request your own availability"))
	default:
		WriteError(w, model.NewInternalError("availability operation failed"))
	}
}
