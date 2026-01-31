package handler

import (
	"fmt"
	"net/http"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
	"github.com/google/uuid"
)

// EventsHandler handles SSE event streaming
type EventsHandler struct {
	eventHub *service.EventHub
}

// NewEventsHandler creates a new events handler
func NewEventsHandler(eventHub *service.EventHub) *EventsHandler {
	return &EventsHandler{
		eventHub: eventHub,
	}
}

// Stream handles GET /v1/guilds/{guildId}/events
// This endpoint streams SSE events for the guild
func (h *EventsHandler) Stream(w http.ResponseWriter, r *http.Request) {
	guildID := middleware.GetGuildID(r.Context())
	if guildID == "" {
		WriteError(w, model.NewBadRequestError("guild ID required"))
		return
	}

	// Check if the client supports SSE
	flusher, ok := w.(http.Flusher)
	if !ok {
		WriteError(w, model.NewInternalError("streaming not supported"))
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Generate subscriber ID
	subscriberID := uuid.New().String()

	// Subscribe to events
	sub := h.eventHub.Subscribe(guildID, subscriberID)
	defer h.eventHub.Unsubscribe(guildID, subscriberID)

	// Send initial connection event
	fmt.Fprintf(w, "event: connected\ndata: {\"subscriber_id\":\"%s\"}\n\n", subscriberID)
	flusher.Flush()

	// Stream events
	for {
		select {
		case event, ok := <-sub.Events:
			if !ok {
				return
			}
			fmt.Fprint(w, event.Format())
			flusher.Flush()

		case <-sub.Done:
			return

		case <-r.Context().Done():
			// Client disconnected
			return
		}
	}
}
