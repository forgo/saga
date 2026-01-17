package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// EventHandler handles event endpoints
type EventHandler struct {
	eventService *service.EventService
}

// NewEventHandler creates a new event handler
func NewEventHandler(eventService *service.EventService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

// CreateEvent handles POST /v1/events - create a new event
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var req model.CreateEventRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	// Validate
	var fieldErrors []model.FieldError
	if req.Title == "" {
		fieldErrors = append(fieldErrors, model.FieldError{
			Field:   "title",
			Message: "title is required",
		})
	}
	if req.StartTime.IsZero() {
		fieldErrors = append(fieldErrors, model.FieldError{
			Field:   "start_time",
			Message: "start_time is required",
		})
	}
	if len(fieldErrors) > 0 {
		WriteError(w, model.NewValidationError(fieldErrors))
		return
	}

	event, err := h.eventService.CreateEvent(r.Context(), userID, &req)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to create event"))
		return
	}

	WriteData(w, http.StatusCreated, event, map[string]string{
		"self": "/v1/events/" + event.ID,
	})
}

// GetEvent handles GET /v1/events/{eventId} - get event details
func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	eventID := r.PathValue("eventId")
	if eventID == "" {
		WriteError(w, model.NewBadRequestError("event ID required"))
		return
	}

	userID := middleware.GetUserID(r.Context())

	eventDetails, err := h.eventService.GetEventWithDetails(r.Context(), eventID, userID)
	if err != nil {
		h.handleEventError(w, err)
		return
	}

	WriteData(w, http.StatusOK, eventDetails, map[string]string{
		"self": "/v1/events/" + eventID,
	})
}

// UpdateEvent handles PATCH /v1/events/{eventId} - update an event
func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
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

	var req model.UpdateEventRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	event, err := h.eventService.UpdateEvent(r.Context(), userID, eventID, &req)
	if err != nil {
		h.handleEventError(w, err)
		return
	}

	WriteData(w, http.StatusOK, event, map[string]string{
		"self": "/v1/events/" + eventID,
	})
}

// CancelEvent handles DELETE /v1/events/{eventId} - cancel an event
func (h *EventHandler) CancelEvent(w http.ResponseWriter, r *http.Request) {
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

	if err := h.eventService.CancelEvent(r.Context(), userID, eventID); err != nil {
		h.handleEventError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RSVP handles POST /v1/events/{eventId}/rsvp - RSVP to an event
func (h *EventHandler) RSVP(w http.ResponseWriter, r *http.Request) {
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

	var req model.RSVPRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	rsvp, err := h.eventService.RSVP(r.Context(), userID, eventID, &req)
	if err != nil {
		h.handleEventError(w, err)
		return
	}

	if rsvp == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	WriteData(w, http.StatusCreated, rsvp, map[string]string{
		"self": "/v1/events/" + eventID + "/rsvp",
	})
}

// CancelRSVP handles DELETE /v1/events/{eventId}/rsvp - cancel own RSVP
func (h *EventHandler) CancelRSVP(w http.ResponseWriter, r *http.Request) {
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

	if err := h.eventService.CancelRSVP(r.Context(), userID, eventID); err != nil {
		h.handleEventError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetPendingRSVPs handles GET /v1/events/{eventId}/rsvps/pending - get pending RSVPs (host only)
func (h *EventHandler) GetPendingRSVPs(w http.ResponseWriter, r *http.Request) {
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

	rsvps, err := h.eventService.GetPendingRSVPs(r.Context(), userID, eventID)
	if err != nil {
		h.handleEventError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, rsvps, nil, map[string]string{
		"self": "/v1/events/" + eventID + "/rsvps/pending",
	})
}

// RespondToRSVP handles POST /v1/events/{eventId}/rsvps/{userId}/respond - respond to an RSVP (host only)
func (h *EventHandler) RespondToRSVP(w http.ResponseWriter, r *http.Request) {
	hostUserID := middleware.GetUserID(r.Context())
	if hostUserID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	eventID := r.PathValue("eventId")
	rsvpUserID := r.PathValue("userId")
	if eventID == "" || rsvpUserID == "" {
		WriteError(w, model.NewBadRequestError("event ID and user ID required"))
		return
	}

	var req model.RespondToRSVPRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	rsvp, err := h.eventService.RespondToRSVP(r.Context(), hostUserID, eventID, rsvpUserID, &req)
	if err != nil {
		h.handleEventError(w, err)
		return
	}

	WriteData(w, http.StatusOK, rsvp, nil)
}

// AddHost handles POST /v1/events/{eventId}/hosts - add a co-host
func (h *EventHandler) AddHost(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		UserID string `json:"user_id"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	host, err := h.eventService.AddHost(r.Context(), userID, eventID, req.UserID)
	if err != nil {
		h.handleEventError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, host, nil)
}

// ConfirmCompletion handles POST /v1/events/{eventId}/confirm - confirm event attendance
func (h *EventHandler) ConfirmCompletion(w http.ResponseWriter, r *http.Request) {
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

	var req model.ConfirmEventCompletionRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	if err := h.eventService.ConfirmCompletion(r.Context(), userID, eventID, req.Completed); err != nil {
		h.handleEventError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Checkin handles POST /v1/events/{eventId}/checkin - check in to event
func (h *EventHandler) Checkin(w http.ResponseWriter, r *http.Request) {
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

	if err := h.eventService.Checkin(r.Context(), userID, eventID); err != nil {
		h.handleEventError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SubmitFeedback handles POST /v1/events/{eventId}/feedback - submit event feedback
func (h *EventHandler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
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

	var req model.EventFeedbackRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	if err := h.eventService.SubmitFeedback(r.Context(), userID, eventID, &req); err != nil {
		h.handleEventError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetPublicEvents handles GET /v1/discover/events - discover public events
func (h *EventHandler) GetPublicEvents(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if r.URL.Query().Get("limit") != "" {
		if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	var filters model.EventSearchFilters
	if template := r.URL.Query().Get("template"); template != "" {
		filters.Template = &template
	}
	if city := r.URL.Query().Get("city"); city != "" {
		filters.City = &city
	}

	events, err := h.eventService.GetPublicEvents(r.Context(), &filters, limit)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get events"))
		return
	}

	WriteCollection(w, http.StatusOK, events, nil, map[string]string{
		"self": "/v1/discover/events",
	})
}

// GetCircleEvents handles GET /v1/circles/{circleId}/events - get circle events
func (h *EventHandler) GetCircleEvents(w http.ResponseWriter, r *http.Request) {
	circleID := r.PathValue("circleId")
	if circleID == "" {
		WriteError(w, model.NewBadRequestError("circle ID required"))
		return
	}

	events, err := h.eventService.GetCircleEvents(r.Context(), circleID, nil)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get events"))
		return
	}

	WriteCollection(w, http.StatusOK, events, nil, map[string]string{
		"self": "/v1/circles/" + circleID + "/events",
	})
}

func (h *EventHandler) handleEventError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrEventNotFound):
		WriteError(w, model.NewNotFoundError("event"))
	case errors.Is(err, service.ErrRSVPNotFound):
		WriteError(w, model.NewNotFoundError("RSVP"))
	case errors.Is(err, service.ErrNotEventHost):
		WriteError(w, model.NewForbiddenError("not an event host"))
	case errors.Is(err, service.ErrEventFull):
		WriteError(w, model.NewConflictError("event is full"))
	case errors.Is(err, service.ErrAlreadyRSVPd):
		WriteError(w, model.NewConflictError("already RSVP'd"))
	case errors.Is(err, service.ErrValuesCheckRequired):
		WriteError(w, model.NewBadRequestError("values alignment check required"))
	default:
		WriteError(w, model.NewInternalError("event operation failed"))
	}
}
