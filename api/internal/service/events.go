package service

import (
	"encoding/json"
	"sync"
	"time"
)

// EventType represents the type of event
type EventType string

const (
	// Guild events
	EventMemberJoined EventType = "guild.member_joined"
	EventMemberLeft   EventType = "guild.member_left"

	// System events
	EventHeartbeat EventType = "heartbeat"

	// Nudge events
	EventNudge EventType = "nudge"
)

// Event represents a server-sent event
type Event struct {
	Type     EventType   `json:"type"`
	Data     interface{} `json:"data"`
	CircleID string      `json:"-"` // Used for routing, not sent to client
}

// Format returns the SSE formatted string
func (e *Event) Format() string {
	data, _ := json.Marshal(e.Data)
	return "event: " + string(e.Type) + "\ndata: " + string(data) + "\n\n"
}

// Subscriber represents a connected SSE client
type Subscriber struct {
	ID       string
	CircleID string
	Events   chan *Event
	Done     chan struct{}
}

// EventHub manages SSE subscriptions and event broadcasting
type EventHub struct {
	mu              sync.RWMutex
	subscribers     map[string]map[string]*Subscriber // circleID -> subscriberID -> subscriber
	userSubscribers map[string]map[string]*Subscriber // userID -> subscriberID -> subscriber (for user-directed events)
	heartbeat       *time.Ticker
	done            chan struct{}
}

// NewEventHub creates a new event hub
func NewEventHub() *EventHub {
	hub := &EventHub{
		subscribers:     make(map[string]map[string]*Subscriber),
		userSubscribers: make(map[string]map[string]*Subscriber),
		done:            make(chan struct{}),
	}
	// Start heartbeat
	hub.heartbeat = time.NewTicker(30 * time.Second)
	go hub.sendHeartbeats()
	return hub
}

// Subscribe adds a new subscriber for a circle
func (h *EventHub) Subscribe(circleID, subscriberID string) *Subscriber {
	h.mu.Lock()
	defer h.mu.Unlock()

	sub := &Subscriber{
		ID:       subscriberID,
		CircleID: circleID,
		Events:   make(chan *Event, 100), // Buffer to prevent blocking
		Done:     make(chan struct{}),
	}

	if h.subscribers[circleID] == nil {
		h.subscribers[circleID] = make(map[string]*Subscriber)
	}
	h.subscribers[circleID][subscriberID] = sub

	return sub
}

// Unsubscribe removes a subscriber
func (h *EventHub) Unsubscribe(circleID, subscriberID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if circleSubs, ok := h.subscribers[circleID]; ok {
		if sub, ok := circleSubs[subscriberID]; ok {
			close(sub.Done)
			close(sub.Events)
			delete(circleSubs, subscriberID)
		}
		if len(circleSubs) == 0 {
			delete(h.subscribers, circleID)
		}
	}
}

// Publish sends an event to all subscribers of a circle
func (h *EventHub) Publish(event *Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	circleSubs, ok := h.subscribers[event.CircleID]
	if !ok {
		return
	}

	for _, sub := range circleSubs {
		select {
		case sub.Events <- event:
			// Event sent successfully
		default:
			// Buffer full, skip this subscriber
		}
	}
}

// SubscribeUser adds a new subscriber for a specific user (for nudges, notifications)
func (h *EventHub) SubscribeUser(userID, subscriberID string) *Subscriber {
	h.mu.Lock()
	defer h.mu.Unlock()

	sub := &Subscriber{
		ID:       subscriberID,
		CircleID: "", // Not circle-bound
		Events:   make(chan *Event, 100),
		Done:     make(chan struct{}),
	}

	if h.userSubscribers[userID] == nil {
		h.userSubscribers[userID] = make(map[string]*Subscriber)
	}
	h.userSubscribers[userID][subscriberID] = sub

	return sub
}

// UnsubscribeUser removes a user subscriber
func (h *EventHub) UnsubscribeUser(userID, subscriberID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if userSubs, ok := h.userSubscribers[userID]; ok {
		if sub, ok := userSubs[subscriberID]; ok {
			close(sub.Done)
			close(sub.Events)
			delete(userSubs, subscriberID)
		}
		if len(userSubs) == 0 {
			delete(h.userSubscribers, userID)
		}
	}
}

// SendToUser sends an event to all subscribers of a specific user
func (h *EventHub) SendToUser(userID string, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userSubs, ok := h.userSubscribers[userID]
	if !ok {
		return
	}

	for _, sub := range userSubs {
		select {
		case sub.Events <- &event:
			// Event sent successfully
		default:
			// Buffer full, skip this subscriber
		}
	}
}

// sendHeartbeats sends periodic heartbeats to all subscribers
func (h *EventHub) sendHeartbeats() {
	for {
		select {
		case <-h.heartbeat.C:
			h.mu.RLock()
			event := &Event{
				Type: EventHeartbeat,
				Data: map[string]string{
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
			}
			for circleID, circleSubs := range h.subscribers {
				event.CircleID = circleID
				for _, sub := range circleSubs {
					select {
					case sub.Events <- event:
					default:
					}
				}
			}
			h.mu.RUnlock()
		case <-h.done:
			return
		}
	}
}

// Close stops the event hub
func (h *EventHub) Close() {
	close(h.done)
	h.heartbeat.Stop()

	h.mu.Lock()
	defer h.mu.Unlock()

	for circleID, circleSubs := range h.subscribers {
		for _, sub := range circleSubs {
			close(sub.Done)
			close(sub.Events)
		}
		delete(h.subscribers, circleID)
	}
}

// SubscriberCount returns the number of subscribers for a circle
func (h *EventHub) SubscriberCount(circleID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if circleSubs, ok := h.subscribers[circleID]; ok {
		return len(circleSubs)
	}
	return 0
}

// Helper functions for creating events

// NewTimerEvent creates a timer event
func NewTimerEvent(eventType EventType, circleID string, data interface{}) *Event {
	return &Event{
		Type:     eventType,
		CircleID: circleID,
		Data:     data,
	}
}

// NewPersonEvent creates a person event
func NewPersonEvent(eventType EventType, circleID string, data interface{}) *Event {
	return &Event{
		Type:     eventType,
		CircleID: circleID,
		Data:     data,
	}
}

// NewActivityEvent creates an activity event
func NewActivityEvent(eventType EventType, circleID string, data interface{}) *Event {
	return &Event{
		Type:     eventType,
		CircleID: circleID,
		Data:     data,
	}
}

// NewMemberEvent creates a member join/leave event
func NewMemberEvent(eventType EventType, circleID string, data interface{}) *Event {
	return &Event{
		Type:     eventType,
		CircleID: circleID,
		Data:     data,
	}
}
