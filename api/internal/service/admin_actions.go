package service

import (
	"context"
	"fmt"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// AdminActionsService handles admin-triggered actions for testing real-time events
type AdminActionsService struct {
	db       database.Database
	eventHub *EventHub
}

// NewAdminActionsService creates a new admin actions service
func NewAdminActionsService(db database.Database, eventHub *EventHub) *AdminActionsService {
	return &AdminActionsService{
		db:       db,
		eventHub: eventHub,
	}
}

// ActionResult contains the result of an admin action
type ActionResult struct {
	Success   bool        `json:"success"`
	Action    string      `json:"action"`
	ActingAs  string      `json:"acting_as"`
	Target    string      `json:"target,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// UpdateLocationRequest defines the request for updating a user's location
type UpdateLocationRequest struct {
	UserID string  `json:"user_id"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	City   string  `json:"city,omitempty"`
}

// UpdateLocation updates a user's location (as admin)
func (s *AdminActionsService) UpdateLocation(ctx context.Context, req UpdateLocationRequest) (*ActionResult, error) {
	if req.UserID == "" {
		return nil, model.NewBadRequestError("user_id is required")
	}

	city := req.City
	if city == "" {
		city = "Unknown"
	}

	query := `
		UPDATE profile SET
			location = {
				lat: $lat,
				lng: $lng,
				city: $city,
				country: "United States",
				country_code: "US"
			},
			last_active = time::now(),
			updated_on = time::now()
		WHERE user_id = type::record($user_id)
	`

	err := s.db.Execute(ctx, query, map[string]interface{}{
		"user_id": req.UserID,
		"lat":     req.Lat,
		"lng":     req.Lng,
		"city":    city,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update location: %w", err)
	}

	return &ActionResult{
		Success:   true,
		Action:    "location_update",
		ActingAs:  req.UserID,
		Data:      map[string]interface{}{"lat": req.Lat, "lng": req.Lng, "city": city},
		Timestamp: time.Now(),
	}, nil
}

// CreateTrustRatingRequest defines the request for creating a trust rating
type CreateTrustRatingRequest struct {
	RaterID     string `json:"rater_id"`
	RateeID     string `json:"ratee_id"`
	TrustLevel  string `json:"trust_level"` // "high", "medium", "low", "distrust"
	TrustReview string `json:"trust_review,omitempty"`
	AnchorType  string `json:"anchor_type,omitempty"` // "event", "hangout", etc.
	AnchorID    string `json:"anchor_id,omitempty"`
}

// CreateTrustRating creates a trust rating (bypassing validation for testing)
func (s *AdminActionsService) CreateTrustRating(ctx context.Context, req CreateTrustRatingRequest) (*ActionResult, error) {
	if req.RaterID == "" || req.RateeID == "" {
		return nil, model.NewBadRequestError("rater_id and ratee_id are required")
	}

	if req.TrustLevel == "" {
		req.TrustLevel = "medium"
	}

	if req.AnchorType == "" {
		req.AnchorType = "admin_test"
	}
	if req.AnchorID == "" {
		req.AnchorID = fmt.Sprintf("admin_%d", time.Now().Unix())
	}

	query := `
		CREATE trust_rating CONTENT {
			rater_id: type::record($rater_id),
			ratee_id: type::record($ratee_id),
			anchor_type: $anchor_type,
			anchor_id: $anchor_id,
			trust_level: $trust_level,
			trust_review: $trust_review,
			created_on: time::now()
		}
	`

	_, err := s.db.Query(ctx, query, map[string]interface{}{
		"rater_id":     req.RaterID,
		"ratee_id":     req.RateeID,
		"anchor_type":  req.AnchorType,
		"anchor_id":    req.AnchorID,
		"trust_level":  req.TrustLevel,
		"trust_review": req.TrustReview,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create trust rating: %w", err)
	}

	// Send event to the ratee
	if s.eventHub != nil {
		s.eventHub.SendToUser(req.RateeID, Event{
			Type: "trust_rating.received",
			Data: map[string]interface{}{
				"rater_id":    req.RaterID,
				"trust_level": req.TrustLevel,
			},
		})
	}

	return &ActionResult{
		Success:   true,
		Action:    "trust_rating",
		ActingAs:  req.RaterID,
		Target:    req.RateeID,
		Data:      map[string]interface{}{"trust_level": req.TrustLevel},
		Timestamp: time.Now(),
	}, nil
}

// JoinGuildRequest defines the request for joining a guild
type JoinGuildRequest struct {
	UserID  string `json:"user_id"`
	GuildID string `json:"guild_id"`
}

// JoinGuild has a user join a guild (bypassing invite requirements)
func (s *AdminActionsService) JoinGuild(ctx context.Context, req JoinGuildRequest) (*ActionResult, error) {
	if req.UserID == "" || req.GuildID == "" {
		return nil, model.NewBadRequestError("user_id and guild_id are required")
	}

	// Get user info for member record
	userQuery := `SELECT email FROM user WHERE id = type::record($user_id)`
	userResults, err := s.db.Query(ctx, userQuery, map[string]interface{}{
		"user_id": req.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	email := extractFieldString(userResults, "email")
	if email == "" {
		email = fmt.Sprintf("admin_member_%d@test.local", time.Now().Unix())
	}

	// Create member
	memberQuery := `
		CREATE member CONTENT {
			name: $email,
			email: $email,
			user: type::record($user_id),
			created_on: time::now(),
			updated_on: time::now()
		}
	`
	memberResults, err := s.db.Query(ctx, memberQuery, map[string]interface{}{
		"email":   email,
		"user_id": req.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create member: %w", err)
	}

	memberID := extractID(memberResults)
	if memberID == "" {
		return nil, fmt.Errorf("failed to extract member ID")
	}

	// Link member to guild
	relateQuery := `
		LET $m = type::record($member_id);
		LET $g = type::record($guild_id);
		RELATE $m->responsible_for->$g SET role = "member";
	`
	err = s.db.Execute(ctx, relateQuery, map[string]interface{}{
		"member_id": memberID,
		"guild_id":  req.GuildID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to link member to guild: %w", err)
	}

	// Broadcast event
	if s.eventHub != nil {
		s.eventHub.Publish(&Event{
			Type:     EventMemberJoined,
			CircleID: req.GuildID,
			Data: map[string]interface{}{
				"user_id":  req.UserID,
				"guild_id": req.GuildID,
			},
		})
	}

	return &ActionResult{
		Success:   true,
		Action:    "guild_join",
		ActingAs:  req.UserID,
		Target:    req.GuildID,
		Timestamp: time.Now(),
	}, nil
}

// RSVPRequest defines the request for RSVPing to an event
type RSVPRequest struct {
	UserID   string `json:"user_id"`
	EventID  string `json:"event_id"`
	Response string `json:"response"` // "yes", "no", "maybe"
}

// RSVP creates an RSVP for an event
func (s *AdminActionsService) RSVP(ctx context.Context, req RSVPRequest) (*ActionResult, error) {
	if req.UserID == "" || req.EventID == "" {
		return nil, model.NewBadRequestError("user_id and event_id are required")
	}

	if req.Response == "" {
		req.Response = "yes"
	}

	status := model.RSVPStatusPending
	switch req.Response {
	case "yes":
		status = model.RSVPStatusApproved
	case "no":
		status = model.RSVPStatusDeclined
	case "maybe":
		status = model.RSVPStatusPending
	}

	query := `
		CREATE rsvp CONTENT {
			event_id: type::record($event_id),
			user_id: type::record($user_id),
			status: $status,
			created_on: time::now(),
			updated_on: time::now()
		}
	`

	_, err := s.db.Query(ctx, query, map[string]interface{}{
		"event_id": req.EventID,
		"user_id":  req.UserID,
		"status":   status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create RSVP: %w", err)
	}

	// Update event attendee count if confirmed
	if status == model.RSVPStatusApproved {
		updateQuery := `UPDATE event SET attendee_count += 1 WHERE id = type::record($event_id)`
		_ = s.db.Execute(ctx, updateQuery, map[string]interface{}{
			"event_id": req.EventID,
		})
	}

	// Send event to event creator
	eventQuery := `SELECT created_by FROM event WHERE id = type::record($event_id)`
	eventResults, _ := s.db.Query(ctx, eventQuery, map[string]interface{}{
		"event_id": req.EventID,
	})
	creatorID := extractFieldString(eventResults, "created_by")

	if s.eventHub != nil && creatorID != "" {
		s.eventHub.SendToUser(creatorID, Event{
			Type: "event.rsvp",
			Data: map[string]interface{}{
				"event_id": req.EventID,
				"user_id":  req.UserID,
				"response": req.Response,
			},
		})
	}

	return &ActionResult{
		Success:   true,
		Action:    "rsvp",
		ActingAs:  req.UserID,
		Target:    req.EventID,
		Data:      map[string]interface{}{"response": req.Response},
		Timestamp: time.Now(),
	}, nil
}

// CreateEventRequest defines the request for creating an event
type CreateEventRequest struct {
	UserID      string    `json:"user_id"`
	GuildID     string    `json:"guild_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	StartsAt    time.Time `json:"starts_at,omitempty"`
}

// CreateEvent creates an event as a user
func (s *AdminActionsService) CreateEvent(ctx context.Context, req CreateEventRequest) (*ActionResult, error) {
	if req.UserID == "" || req.GuildID == "" || req.Title == "" {
		return nil, model.NewBadRequestError("user_id, guild_id, and title are required")
	}

	if req.StartsAt.IsZero() {
		req.StartsAt = time.Now().Add(24 * time.Hour) // Default to tomorrow
	}

	endsAt := req.StartsAt.Add(2 * time.Hour)
	confirmationDeadline := endsAt.Add(48 * time.Hour)

	// Get member ID for the user in this guild
	memberQuery := `
		SELECT in AS member_id FROM responsible_for
		WHERE out = type::record($guild_id)
		AND in.user = type::record($user_id)
		LIMIT 1
	`
	memberResults, err := s.db.Query(ctx, memberQuery, map[string]interface{}{
		"guild_id": req.GuildID,
		"user_id":  req.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find member: %w", err)
	}

	memberID := extractFieldString(memberResults, "member_id")
	if memberID == "" {
		// User isn't a member, join them first
		joinResult, err := s.JoinGuild(ctx, JoinGuildRequest{
			UserID:  req.UserID,
			GuildID: req.GuildID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to join guild: %w", err)
		}
		_ = joinResult
		// Re-query for member ID
		memberResults, _ = s.db.Query(ctx, memberQuery, map[string]interface{}{
			"guild_id": req.GuildID,
			"user_id":  req.UserID,
		})
		memberID = extractFieldString(memberResults, "member_id")
	}

	query := `
		CREATE event SET
			guild_id = type::record($guild_id),
			title = $title,
			description = $description,
			template = "casual",
			visibility = "guilds",
			starts_at = $starts_at,
			ends_at = $ends_at,
			is_support_event = false,
			status = "published",
			created_by = type::record($created_by),
			confirmation_deadline = $confirmation_deadline,
			confirmed_count = 0,
			requires_confirmation = true,
			completion_verified = false,
			attendee_count = 0,
			created_on = time::now(),
			updated_on = time::now()
	`

	results, err := s.db.Query(ctx, query, map[string]interface{}{
		"guild_id":              req.GuildID,
		"title":                 req.Title,
		"description":           req.Description,
		"starts_at":             req.StartsAt,
		"ends_at":               endsAt,
		"created_by":            memberID,
		"confirmation_deadline": confirmationDeadline,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	eventID := extractID(results)

	// Broadcast event to guild
	if s.eventHub != nil {
		s.eventHub.Publish(&Event{
			Type:     "event.created",
			CircleID: req.GuildID,
			Data: map[string]interface{}{
				"event_id":   eventID,
				"title":      req.Title,
				"created_by": req.UserID,
			},
		})
	}

	return &ActionResult{
		Success:   true,
		Action:    "event_create",
		ActingAs:  req.UserID,
		Target:    req.GuildID,
		Data:      map[string]interface{}{"event_id": eventID, "title": req.Title},
		Timestamp: time.Now(),
	}, nil
}

// GetUsersRequest defines the request for listing users
type GetUsersRequest struct {
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
	Prefix string `json:"prefix,omitempty"` // Filter by email prefix
}

// GetUsers lists users (for admin action panel user selection)
func (s *AdminActionsService) GetUsers(ctx context.Context, req GetUsersRequest) ([]map[string]interface{}, error) {
	if req.Limit <= 0 {
		req.Limit = 50
	}

	query := `
		SELECT
			id,
			email,
			username,
			firstname,
			lastname,
			role,
			created_on
		FROM user
		ORDER BY created_on DESC
		LIMIT $limit
		START $offset
	`

	if req.Prefix != "" {
		query = fmt.Sprintf(`
			SELECT
				id,
				email,
				username,
				firstname,
				lastname,
				role,
				created_on
			FROM user
			WHERE email CONTAINS '%s'
			ORDER BY created_on DESC
			LIMIT $limit
			START $offset
		`, req.Prefix)
	}

	results, err := s.db.Query(ctx, query, map[string]interface{}{
		"limit":  req.Limit,
		"offset": req.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	return extractResultArray(results), nil
}

// GetGuilds lists guilds (for admin action panel)
func (s *AdminActionsService) GetGuilds(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT
			id,
			name,
			description,
			visibility,
			created_on
		FROM guild
		ORDER BY created_on DESC
		LIMIT $limit
	`

	results, err := s.db.Query(ctx, query, map[string]interface{}{
		"limit": limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get guilds: %w", err)
	}

	return extractResultArray(results), nil
}

// GetEvents lists events (for admin action panel)
func (s *AdminActionsService) GetEvents(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT
			id,
			title,
			guild_id,
			status,
			starts_at,
			attendee_count,
			created_on
		FROM event
		WHERE status = "published"
		ORDER BY starts_at ASC
		LIMIT $limit
	`

	results, err := s.db.Query(ctx, query, map[string]interface{}{
		"limit": limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	return extractResultArray(results), nil
}

// Helper to extract a field from query results
func extractFieldString(results []interface{}, field string) string {
	if len(results) == 0 {
		return ""
	}

	resp, ok := results[0].(map[string]interface{})
	if !ok {
		return ""
	}

	result, ok := resp["result"]
	if !ok {
		return ""
	}

	arr, ok := result.([]interface{})
	if !ok || len(arr) == 0 {
		return ""
	}

	data, ok := arr[0].(map[string]interface{})
	if !ok {
		return ""
	}

	if v, ok := data[field]; ok {
		return formatID(v)
	}
	return ""
}

// Helper to extract result array
func extractResultArray(results []interface{}) []map[string]interface{} {
	var items []map[string]interface{}
	if len(results) == 0 {
		return items
	}

	resp, ok := results[0].(map[string]interface{})
	if !ok {
		return items
	}

	result, ok := resp["result"]
	if !ok {
		return items
	}

	arr, ok := result.([]interface{})
	if !ok {
		return items
	}

	for _, item := range arr {
		if m, ok := item.(map[string]interface{}); ok {
			// Format IDs
			if id, ok := m["id"]; ok {
				m["id"] = formatID(id)
			}
			if guildID, ok := m["guild_id"]; ok {
				m["guild_id"] = formatID(guildID)
			}
			items = append(items, m)
		}
	}

	return items
}
