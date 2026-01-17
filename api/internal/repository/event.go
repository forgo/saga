package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// EventRepository handles event data access
type EventRepository struct {
	db database.Database
}

// NewEventRepository creates a new event repository
func NewEventRepository(db database.Database) *EventRepository {
	return &EventRepository{db: db}
}

// Create creates a new event
func (r *EventRepository) Create(ctx context.Context, event *model.Event) error {
	// Build query dynamically to handle optional fields (SurrealDB option<T> requires NONE, not NULL)
	vars := map[string]interface{}{
		"title":      event.Title,
		"start_time": event.StartTime,
		"template":   event.Template,
		"visibility": event.Visibility,
		"status":     event.Status,
		"created_by": event.CreatedBy,
	}

	// Build the SET clause with required fields
	setClause := `
		title = $title,
		start_time = $start_time,
		template = $template,
		visibility = $visibility,
		status = $status,
		created_by = $created_by,
		created_on = time::now(),
		updated_on = time::now()`

	// Add optional fields only when they have values
	if event.GuildID != nil {
		setClause += ", guild_id = $guild_id"
		vars["guild_id"] = event.GuildID
	}
	if event.Description != nil {
		setClause += ", description = $description"
		vars["description"] = event.Description
	}
	if event.Location != nil {
		setClause += ", location = $location"
		vars["location"] = map[string]interface{}{
			"name":         event.Location.Name,
			"address":      event.Location.Address,
			"neighborhood": event.Location.Neighborhood,
			"city":         event.Location.City,
			"lat":          event.Location.Lat,
			"lng":          event.Location.Lng,
			"is_virtual":   event.Location.IsVirtual,
			"meet_link":    event.Location.MeetLink,
		}
	}
	if event.EndTime != nil {
		setClause += ", end_time = $end_time"
		vars["end_time"] = event.EndTime
	}
	if event.MaxAttendees != nil {
		setClause += ", max_attendees = $max_attendees"
		vars["max_attendees"] = event.MaxAttendees
	}
	if event.WaitlistEnabled {
		setClause += ", waitlist_enabled = $waitlist_enabled"
		vars["waitlist_enabled"] = event.WaitlistEnabled
	}
	if event.CoverImage != nil {
		setClause += ", cover_image = $cover_image"
		vars["cover_image"] = event.CoverImage
	}
	if event.ThemeColor != nil {
		setClause += ", theme_color = $theme_color"
		vars["theme_color"] = event.ThemeColor
	}
	if event.ValuesRequired {
		setClause += ", values_required = $values_required"
		vars["values_required"] = event.ValuesRequired
	}
	if len(event.ValuesQuestions) > 0 {
		setClause += ", values_questions = $values_questions"
		vars["values_questions"] = event.ValuesQuestions
	}
	if event.AutoApproveAligned {
		setClause += ", auto_approve_aligned = $auto_approve_aligned"
		vars["auto_approve_aligned"] = event.AutoApproveAligned
	}
	if event.YikesThreshold != 0 {
		setClause += ", yikes_threshold = $yikes_threshold"
		vars["yikes_threshold"] = event.YikesThreshold
	}
	if event.IsSupportEvent {
		setClause += ", is_support_event = $is_support_event"
		vars["is_support_event"] = event.IsSupportEvent
	}
	if event.RequiresConfirmation {
		setClause += ", requires_confirmation = $requires_confirmation"
		vars["requires_confirmation"] = event.RequiresConfirmation
	}
	if event.ConfirmationDeadline != nil {
		setClause += ", confirmation_deadline = $confirmation_deadline"
		vars["confirmation_deadline"] = event.ConfirmationDeadline
	}

	query := "CREATE event SET " + setClause

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	event.ID = created.ID
	event.CreatedOn = created.CreatedOn
	event.UpdatedOn = created.UpdatedOn
	return nil
}

// Get retrieves an event by ID
func (r *EventRepository) Get(ctx context.Context, eventID string) (*model.Event, error) {
	// Direct record access - more efficient than WHERE id =
	query := `SELECT * FROM type::record($event_id)`
	vars := map[string]interface{}{"event_id": eventID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.parseEventResult(result)
}

// Update updates an event
func (r *EventRepository) Update(ctx context.Context, eventID string, updates map[string]interface{}) (*model.Event, error) {
	query := `UPDATE event SET updated_on = time::now()`
	vars := map[string]interface{}{"event_id": eventID}

	for key, value := range updates {
		query += ", " + key + " = $" + key
		vars[key] = value
	}

	query += ` WHERE id = type::record($event_id) RETURN AFTER`

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseEventResult(result)
}

// Delete deletes an event
func (r *EventRepository) Delete(ctx context.Context, eventID string) error {
	query := `DELETE event WHERE id = type::record($event_id)`
	vars := map[string]interface{}{"event_id": eventID}

	return r.db.Execute(ctx, query, vars)
}

// GetByGuild retrieves events for a guild
func (r *EventRepository) GetByGuild(ctx context.Context, guildID string, filters *model.EventSearchFilters) ([]*model.Event, error) {
	query := `
		SELECT * FROM event
		WHERE guild_id = $guild_id AND status IN ["published", "completed"]
	`
	vars := map[string]interface{}{"guild_id": guildID}

	if filters != nil && filters.StartAfter != nil {
		query += ` AND start_time >= $start_after`
		vars["start_after"] = *filters.StartAfter
	}

	query += ` ORDER BY start_time ASC`

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseEventsResult(result)
}

// GetPublicEvents retrieves public events
func (r *EventRepository) GetPublicEvents(ctx context.Context, filters *model.EventSearchFilters, limit int) ([]*model.Event, error) {
	query := `
		SELECT * FROM event
		WHERE visibility = "public" AND status = "published"
	`
	vars := map[string]interface{}{"limit": limit}

	if filters != nil {
		if filters.StartAfter != nil {
			query += ` AND start_time >= $start_after`
			vars["start_after"] = *filters.StartAfter
		}
		if filters.StartBefore != nil {
			query += ` AND start_time <= $start_before`
			vars["start_before"] = *filters.StartBefore
		}
		if filters.Template != nil {
			query += ` AND template = $template`
			vars["template"] = *filters.Template
		}
		if filters.City != nil {
			query += ` AND location.city = $city`
			vars["city"] = *filters.City
		}
	}

	query += ` ORDER BY start_time ASC LIMIT $limit`

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseEventsResult(result)
}

// CreateHost adds a host to an event
func (r *EventRepository) CreateHost(ctx context.Context, host *model.EventHost) error {
	query := `
		CREATE event_host CONTENT {
			event_id: $event_id,
			user_id: $user_id,
			role: $role,
			added_on: time::now(),
			added_by: $added_by
		}
	`

	vars := map[string]interface{}{
		"event_id": host.EventID,
		"user_id":  host.UserID,
		"role":     host.Role,
		"added_by": host.AddedBy,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	host.ID = created.ID
	host.AddedOn = created.CreatedOn
	return nil
}

// GetHosts retrieves hosts for an event
func (r *EventRepository) GetHosts(ctx context.Context, eventID string) ([]*model.EventHost, error) {
	query := `
		SELECT * FROM event_host
		WHERE event_id = $event_id
		ORDER BY added_on ASC
	`
	vars := map[string]interface{}{"event_id": eventID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseHostsResult(result)
}

// IsHost checks if user is a host of the event
func (r *EventRepository) IsHost(ctx context.Context, eventID, userID string) (bool, error) {
	query := `
		SELECT count() as cnt FROM event_host
		WHERE event_id = $event_id AND user_id = $user_id
		GROUP ALL
	`
	vars := map[string]interface{}{
		"event_id": eventID,
		"user_id":  userID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	if data, ok := result.(map[string]interface{}); ok {
		return getInt(data, "cnt") > 0, nil
	}
	return false, nil
}

// CreateRSVP creates an RSVP
func (r *EventRepository) CreateRSVP(ctx context.Context, rsvp *model.EventRSVP) error {
	query := `
		CREATE event_rsvp CONTENT {
			event_id: $event_id,
			user_id: $user_id,
			status: $status,
			rsvp_type: $rsvp_type,
			values_aligned: $values_aligned,
			alignment_score: $alignment_score,
			yikes_count: $yikes_count,
			waiting_reason: $waiting_reason,
			plus_ones: $plus_ones,
			plus_one_names: $plus_one_names,
			requested_on: time::now(),
			updated_on: time::now()
		}
	`

	vars := map[string]interface{}{
		"event_id":        rsvp.EventID,
		"user_id":         rsvp.UserID,
		"status":          rsvp.Status,
		"rsvp_type":       rsvp.RSVPType,
		"values_aligned":  rsvp.ValuesAligned,
		"alignment_score": rsvp.AlignmentScore,
		"yikes_count":     rsvp.YikesCount,
		"waiting_reason":  rsvp.WaitingReason,
		"plus_ones":       rsvp.PlusOnes,
		"plus_one_names":  rsvp.PlusOneNames,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	rsvp.ID = created.ID
	rsvp.RequestedOn = created.CreatedOn
	rsvp.UpdatedOn = created.UpdatedOn
	return nil
}

// GetRSVP retrieves an RSVP
func (r *EventRepository) GetRSVP(ctx context.Context, eventID, userID string) (*model.EventRSVP, error) {
	query := `
		SELECT * FROM event_rsvp
		WHERE event_id = $event_id AND user_id = $user_id
		LIMIT 1
	`
	vars := map[string]interface{}{
		"event_id": eventID,
		"user_id":  userID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.parseRSVPResult(result)
}

// UpdateRSVP updates an RSVP
func (r *EventRepository) UpdateRSVP(ctx context.Context, rsvpID string, updates map[string]interface{}) (*model.EventRSVP, error) {
	query := `UPDATE event_rsvp SET updated_on = time::now()`
	vars := map[string]interface{}{"rsvp_id": rsvpID}

	for key, value := range updates {
		query += ", " + key + " = $" + key
		vars[key] = value
	}

	query += ` WHERE id = type::record($rsvp_id) RETURN AFTER`

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseRSVPResult(result)
}

// GetRSVPsByEvent retrieves all RSVPs for an event
func (r *EventRepository) GetRSVPsByEvent(ctx context.Context, eventID string) ([]*model.EventRSVP, error) {
	query := `
		SELECT * FROM event_rsvp
		WHERE event_id = $event_id
		ORDER BY requested_on ASC
	`
	vars := map[string]interface{}{"event_id": eventID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseRSVPsResult(result)
}

// GetPendingRSVPs retrieves pending RSVPs for an event
func (r *EventRepository) GetPendingRSVPs(ctx context.Context, eventID string) ([]*model.EventRSVP, error) {
	query := `
		SELECT * FROM event_rsvp
		WHERE event_id = $event_id AND status = "pending"
		ORDER BY requested_on ASC
	`
	vars := map[string]interface{}{"event_id": eventID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseRSVPsResult(result)
}

// CountApprovedRSVPs counts approved RSVPs including plus ones
func (r *EventRepository) CountApprovedRSVPs(ctx context.Context, eventID string) (int, error) {
	query := `
		SELECT math::sum(1 + plus_ones) as total FROM event_rsvp
		WHERE event_id = $event_id AND status = "approved"
		GROUP ALL
	`
	vars := map[string]interface{}{"event_id": eventID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	if data, ok := result.(map[string]interface{}); ok {
		return getInt(data, "total"), nil
	}
	return 0, nil
}

// Helper functions

func (r *EventRepository) parseEventResult(result interface{}) (*model.Event, error) {
	if result == nil {
		return nil, database.ErrNotFound
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	if id, ok := data["id"]; ok {
		data["id"] = convertSurrealID(id)
	}

	// Handle created_by which may be returned as record object or string
	if cb, ok := data["created_by"]; ok {
		data["created_by"] = convertSurrealID(cb)
	}

	// Handle guild_id which may be returned as record object
	if gid, ok := data["guild_id"]; ok {
		if gidStr := convertSurrealID(gid); gidStr != "" {
			data["guild_id"] = gidStr
		}
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var event model.Event
	if err := json.Unmarshal(jsonBytes, &event); err != nil {
		return nil, err
	}

	// Parse location
	if locData, ok := data["location"].(map[string]interface{}); ok {
		event.Location = &model.EventLocation{
			Name: getString(locData, "name"),
			City: getString(locData, "city"),
			Lat:  getFloat(locData, "lat"),
			Lng:  getFloat(locData, "lng"),
		}
		if addr, ok := locData["address"].(string); ok {
			event.Location.Address = &addr
		}
		if neighborhood, ok := locData["neighborhood"].(string); ok {
			event.Location.Neighborhood = &neighborhood
		}
		if meetLink, ok := locData["meet_link"].(string); ok {
			event.Location.MeetLink = &meetLink
		}
		event.Location.IsVirtual = getBool(locData, "is_virtual")
	}

	event.ValuesQuestions = getStringSlice(data, "values_questions")

	if t := getTime(data, "start_time"); t != nil {
		event.StartTime = *t
	}
	event.EndTime = getTime(data, "end_time")
	if t := getTime(data, "created_on"); t != nil {
		event.CreatedOn = *t
	}
	if t := getTime(data, "updated_on"); t != nil {
		event.UpdatedOn = *t
	}

	return &event, nil
}

func (r *EventRepository) parseEventsResult(result []interface{}) ([]*model.Event, error) {
	events := make([]*model.Event, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					event, err := r.parseEventResult(item)
					if err != nil {
						continue
					}
					events = append(events, event)
				}
				continue
			}
		}

		event, err := r.parseEventResult(res)
		if err != nil {
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *EventRepository) parseHostResult(result interface{}) (*model.EventHost, error) {
	if result == nil {
		return nil, database.ErrNotFound
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	if id, ok := data["id"]; ok {
		data["id"] = convertSurrealID(id)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var host model.EventHost
	if err := json.Unmarshal(jsonBytes, &host); err != nil {
		return nil, err
	}

	if t := getTime(data, "added_on"); t != nil {
		host.AddedOn = *t
	}

	return &host, nil
}

func (r *EventRepository) parseHostsResult(result []interface{}) ([]*model.EventHost, error) {
	hosts := make([]*model.EventHost, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					host, err := r.parseHostResult(item)
					if err != nil {
						continue
					}
					hosts = append(hosts, host)
				}
				continue
			}
		}

		host, err := r.parseHostResult(res)
		if err != nil {
			continue
		}
		hosts = append(hosts, host)
	}

	return hosts, nil
}

func (r *EventRepository) parseRSVPResult(result interface{}) (*model.EventRSVP, error) {
	if result == nil {
		return nil, database.ErrNotFound
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	if id, ok := data["id"]; ok {
		data["id"] = convertSurrealID(id)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var rsvp model.EventRSVP
	if err := json.Unmarshal(jsonBytes, &rsvp); err != nil {
		return nil, err
	}

	rsvp.PlusOneNames = getStringSlice(data, "plus_one_names")
	rsvp.ValuesAligned = getBool(data, "values_aligned")
	rsvp.AlignmentScore = getFloat(data, "alignment_score")
	rsvp.YikesCount = getInt(data, "yikes_count")

	if t := getTime(data, "requested_on"); t != nil {
		rsvp.RequestedOn = *t
	}
	if t := getTime(data, "updated_on"); t != nil {
		rsvp.UpdatedOn = *t
	}
	rsvp.RespondedOn = getTime(data, "responded_on")

	return &rsvp, nil
}

func (r *EventRepository) parseRSVPsResult(result []interface{}) ([]*model.EventRSVP, error) {
	rsvps := make([]*model.EventRSVP, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					rsvp, err := r.parseRSVPResult(item)
					if err != nil {
						continue
					}
					rsvps = append(rsvps, rsvp)
				}
				continue
			}
		}

		rsvp, err := r.parseRSVPResult(res)
		if err != nil {
			continue
		}
		rsvps = append(rsvps, rsvp)
	}

	return rsvps, nil
}

// === Bridge methods for unified RSVP migration ===

// CreateUnifiedRSVP creates an RSVP using the unified system
// This is the preferred method for new code
func (r *EventRepository) CreateUnifiedRSVP(ctx context.Context, rsvp *model.UnifiedRSVP) error {
	rsvp.TargetType = model.RSVPTargetEvent

	role := rsvp.Role
	if role == "" {
		role = model.RSVPRoleParticipant
	}

	status := rsvp.Status
	if status == "" {
		status = model.UnifiedRSVPStatusPending
	}

	// Build query dynamically to handle optional fields (SurrealDB option<T> requires NONE, not NULL)
	vars := map[string]interface{}{
		"target_type": rsvp.TargetType,
		"target_id":   rsvp.TargetID,
		"user_id":     rsvp.UserID,
		"status":      status,
		"role":        role,
	}

	setClause := `
		target_type = $target_type,
		target_id = $target_id,
		user_id = type::record($user_id),
		status = $status,
		role = $role,
		created_on = time::now(),
		updated_on = time::now()`

	// Add optional fields only when they have values
	if rsvp.ValuesAligned != nil {
		setClause += ", values_aligned = $values_aligned"
		vars["values_aligned"] = *rsvp.ValuesAligned
	}
	if rsvp.AlignmentScore != nil {
		setClause += ", alignment_score = $alignment_score"
		vars["alignment_score"] = *rsvp.AlignmentScore
	}
	if rsvp.YikesCount != nil && *rsvp.YikesCount > 0 {
		setClause += ", yikes_count = $yikes_count"
		vars["yikes_count"] = *rsvp.YikesCount
	}
	if rsvp.PlusOnes != nil && *rsvp.PlusOnes > 0 {
		setClause += ", plus_ones = $plus_ones"
		vars["plus_ones"] = *rsvp.PlusOnes
	}
	if len(rsvp.PlusOneNames) > 0 {
		setClause += ", plus_one_names = $plus_one_names"
		vars["plus_one_names"] = rsvp.PlusOneNames
	}
	if rsvp.Note != nil {
		setClause += ", note = $note"
		vars["note"] = *rsvp.Note
	}

	query := "CREATE unified_rsvp SET " + setClause

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	rsvp.ID = created.ID
	rsvp.CreatedOn = created.CreatedOn
	rsvp.UpdatedOn = created.UpdatedOn
	rsvp.Status = status
	rsvp.Role = role
	return nil
}

// GetUnifiedRSVP retrieves a unified RSVP for an event
func (r *EventRepository) GetUnifiedRSVP(ctx context.Context, eventID, userID string) (*model.UnifiedRSVP, error) {
	query := `
		SELECT * FROM unified_rsvp
		WHERE target_type = "event"
		AND target_id = $event_id
		AND user_id = type::record($user_id)
		LIMIT 1
	`
	vars := map[string]interface{}{
		"event_id": eventID,
		"user_id":  userID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.parseUnifiedRSVPResult(result)
}

// ConfirmEventCompletion confirms event completion for resonance scoring
func (r *EventRepository) ConfirmEventCompletion(ctx context.Context, eventID, userID string, early bool) error {
	// Update unified_rsvp
	query := `
		UPDATE unified_rsvp
		SET completion_confirmed = time::now(),
			early_confirmed = $early,
			updated_on = time::now()
		WHERE target_type = "event"
		AND target_id = $event_id
		AND user_id = type::record($user_id)
	`
	vars := map[string]interface{}{
		"event_id": eventID,
		"user_id":  userID,
		"early":    early,
	}

	if err := r.db.Execute(ctx, query, vars); err != nil {
		return err
	}

	// Also update legacy event_rsvp for backwards compatibility
	legacyQuery := `
		UPDATE event_rsvp
		SET completion_confirmed = time::now(),
			updated_on = time::now()
		WHERE event_id = $event_id
		AND user_id = $user_id
	`

	return r.db.Execute(ctx, legacyQuery, vars)
}

// RecordCheckin records a checkin for resonance on-time bonus
func (r *EventRepository) RecordCheckin(ctx context.Context, eventID, userID string) error {
	now := time.Now()

	// Update unified_rsvp
	query := `
		UPDATE unified_rsvp
		SET checkin_time = $checkin_time,
			updated_on = time::now()
		WHERE target_type = "event"
		AND target_id = $event_id
		AND user_id = type::record($user_id)
	`
	vars := map[string]interface{}{
		"event_id":     eventID,
		"user_id":      userID,
		"checkin_time": now,
	}

	if err := r.db.Execute(ctx, query, vars); err != nil {
		return err
	}

	// Also update legacy event_rsvp for backwards compatibility
	legacyQuery := `
		UPDATE event_rsvp
		SET checkin_time = $checkin_time,
			updated_on = time::now()
		WHERE event_id = $event_id
		AND user_id = $user_id
	`

	return r.db.Execute(ctx, legacyQuery, vars)
}

// IncrementConfirmedCount increments the event's confirmed count
func (r *EventRepository) IncrementConfirmedCount(ctx context.Context, eventID string) error {
	query := `
		UPDATE event
		SET confirmed_count = confirmed_count + 1,
			updated_on = time::now()
		WHERE id = type::record($event_id)
	`
	vars := map[string]interface{}{"event_id": eventID}

	return r.db.Execute(ctx, query, vars)
}

// MarkEventVerified marks an event as completion verified
func (r *EventRepository) MarkEventVerified(ctx context.Context, eventID string) error {
	query := `
		UPDATE event
		SET completion_verified = true,
			completion_verified_on = time::now(),
			updated_on = time::now()
		WHERE id = type::record($event_id)
	`
	vars := map[string]interface{}{"event_id": eventID}

	return r.db.Execute(ctx, query, vars)
}

// SetConfirmationDeadline sets the confirmation deadline for an event
func (r *EventRepository) SetConfirmationDeadline(ctx context.Context, eventID string, deadline time.Time) error {
	query := `
		UPDATE event
		SET confirmation_deadline = $deadline,
			updated_on = time::now()
		WHERE id = type::record($event_id)
	`
	vars := map[string]interface{}{
		"event_id": eventID,
		"deadline": deadline,
	}

	return r.db.Execute(ctx, query, vars)
}

// GetEventsNeedingVerification retrieves events that might need verification
func (r *EventRepository) GetEventsNeedingVerification(ctx context.Context) ([]*model.Event, error) {
	query := `
		SELECT * FROM event
		WHERE requires_confirmation = true
		AND completion_verified = false
		AND confirmation_deadline IS NOT NONE
		AND confirmation_deadline < time::now()
		AND status = "completed"
		ORDER BY confirmation_deadline ASC
		LIMIT 100
	`

	result, err := r.db.Query(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	return r.parseEventsResult(result)
}

func (r *EventRepository) parseUnifiedRSVPResult(result interface{}) (*model.UnifiedRSVP, error) {
	if result == nil {
		return nil, database.ErrNotFound
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	rsvp := &model.UnifiedRSVP{
		ID:         convertSurrealID(data["id"]),
		TargetType: getString(data, "target_type"),
		TargetID:   getString(data, "target_id"),
		UserID:     convertSurrealID(data["user_id"]),
		Status:     getString(data, "status"),
		Role:       getString(data, "role"),
	}

	// Optional fields
	if val, ok := data["values_aligned"].(bool); ok {
		rsvp.ValuesAligned = &val
	}
	if val, ok := data["alignment_score"].(float64); ok {
		rsvp.AlignmentScore = &val
	}
	if val, ok := data["yikes_count"].(float64); ok {
		count := int(val)
		rsvp.YikesCount = &count
	}
	if val, ok := data["plus_ones"].(float64); ok {
		count := int(val)
		rsvp.PlusOnes = &count
	}
	rsvp.PlusOneNames = getStringSlice(data, "plus_one_names")

	if note, ok := data["note"].(string); ok {
		rsvp.Note = &note
	}
	if note, ok := data["host_note"].(string); ok {
		rsvp.HostNote = &note
	}

	if rating, ok := data["helpfulness_rating"].(string); ok {
		rsvp.HelpfulnessRating = &rating
	}
	rsvp.HelpfulnessTags = getStringSlice(data, "helpfulness_tags")

	if val, ok := data["early_confirmed"].(bool); ok {
		rsvp.EarlyConfirmed = &val
	}

	// Timestamps
	if t := getTime(data, "created_on"); t != nil {
		rsvp.CreatedOn = *t
	}
	if t := getTime(data, "updated_on"); t != nil {
		rsvp.UpdatedOn = *t
	}
	rsvp.CheckinTime = getTime(data, "checkin_time")
	rsvp.CompletionConfirmed = getTime(data, "completion_confirmed")

	return rsvp, nil
}

// Unused - silence linter
var _ = time.Now
