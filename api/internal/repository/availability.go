package repository

import (
	"context"
	"errors"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// AvailabilityRepository handles availability data access
type AvailabilityRepository struct {
	db database.Database
}

// NewAvailabilityRepository creates a new availability repository
func NewAvailabilityRepository(db database.Database) *AvailabilityRepository {
	return &AvailabilityRepository{db: db}
}

// Create creates a new availability window
func (r *AvailabilityRepository) Create(ctx context.Context, av *model.Availability) error {
	query := `
		CREATE availability CONTENT {
			user: type::record($user_id),
			status: $status,
			start_time: $start_time,
			end_time: $end_time,
			location: $location,
			hangout_type: $hangout_type,
			activity_description: $activity_description,
			activity_venue: $activity_venue,
			interest_id: $interest_id,
			max_people: $max_people,
			note: $note,
			visibility: $visibility,
			expires_at: $expires_at,
			created_on: time::now(),
			updated_on: time::now()
		}
	`

	var locationData interface{}
	if av.Location != nil {
		locationData = map[string]interface{}{
			"lat":    av.Location.Lat,
			"lng":    av.Location.Lng,
			"radius": av.Location.Radius,
		}
	}

	vars := map[string]interface{}{
		"user_id":              av.UserID,
		"status":               av.Status,
		"start_time":           av.StartTime,
		"end_time":             av.EndTime,
		"location":             locationData,
		"hangout_type":         av.HangoutType,
		"activity_description": av.ActivityDescription,
		"activity_venue":       av.ActivityVenue,
		"interest_id":          av.InterestID,
		"max_people":           av.MaxPeople,
		"note":                 av.Note,
		"visibility":           av.Visibility,
		"expires_at":           av.ExpiresAt,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	av.ID = created.ID
	av.CreatedOn = created.CreatedOn
	av.UpdatedOn = created.UpdatedOn
	return nil
}

// GetByID retrieves an availability by ID
func (r *AvailabilityRepository) GetByID(ctx context.Context, id string) (*model.Availability, error) {
	// Direct record access - more efficient than WHERE id =
	query := `SELECT * FROM type::record($id)`
	vars := map[string]interface{}{"id": id}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.parseAvailabilityResult(result)
}

// GetByUser retrieves all active availability windows for a user
func (r *AvailabilityRepository) GetByUser(ctx context.Context, userID string) ([]*model.Availability, error) {
	query := `
		SELECT * FROM availability
		WHERE user = type::record($user_id)
		AND expires_at > time::now()
		ORDER BY start_time
	`
	vars := map[string]interface{}{"user_id": userID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseAvailabilitiesResult(result)
}

// GetNearby finds availabilities within a bounding box and time range
func (r *AvailabilityRepository) GetNearby(ctx context.Context, minLat, maxLat, minLng, maxLng float64, startTime, endTime time.Time, excludeUserID string, limit int) ([]*model.Availability, error) {
	query := `
		SELECT * FROM availability
		WHERE location != NONE
			AND location.lat >= $min_lat
			AND location.lat <= $max_lat
			AND location.lng >= $min_lng
			AND location.lng <= $max_lng
			AND start_time <= $end_time
			AND end_time >= $start_time
			AND expires_at > time::now()
			AND user != type::record($exclude_user)
			AND visibility != "private"
		ORDER BY start_time
		LIMIT $limit
	`

	vars := map[string]interface{}{
		"min_lat":      minLat,
		"max_lat":      maxLat,
		"min_lng":      minLng,
		"max_lng":      maxLng,
		"start_time":   startTime,
		"end_time":     endTime,
		"exclude_user": excludeUserID,
		"limit":        limit,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseAvailabilitiesResult(result)
}

// GetByHangoutType finds availabilities by type
func (r *AvailabilityRepository) GetByHangoutType(ctx context.Context, hangoutType string, excludeUserID string, limit int) ([]*model.Availability, error) {
	query := `
		SELECT * FROM availability
		WHERE hangout_type = $hangout_type
			AND expires_at > time::now()
			AND user != type::record($exclude_user)
			AND visibility != "private"
		ORDER BY start_time
		LIMIT $limit
	`

	vars := map[string]interface{}{
		"hangout_type": hangoutType,
		"exclude_user": excludeUserID,
		"limit":        limit,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseAvailabilitiesResult(result)
}

// Update updates an availability
func (r *AvailabilityRepository) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Availability, error) {
	query := `UPDATE availability SET updated_on = time::now()`
	vars := map[string]interface{}{"id": id}

	if status, ok := updates["status"]; ok {
		query += ", status = $status"
		vars["status"] = status
	}
	if startTime, ok := updates["start_time"]; ok {
		query += ", start_time = $start_time"
		vars["start_time"] = startTime
	}
	if endTime, ok := updates["end_time"]; ok {
		query += ", end_time = $end_time"
		vars["end_time"] = endTime
	}
	if location, ok := updates["location"]; ok {
		query += ", location = $location"
		vars["location"] = location
	}
	if hangoutType, ok := updates["hangout_type"]; ok {
		query += ", hangout_type = $hangout_type"
		vars["hangout_type"] = hangoutType
	}
	if activityDescription, ok := updates["activity_description"]; ok {
		query += ", activity_description = $activity_description"
		vars["activity_description"] = activityDescription
	}
	if activityVenue, ok := updates["activity_venue"]; ok {
		query += ", activity_venue = $activity_venue"
		vars["activity_venue"] = activityVenue
	}
	if maxPeople, ok := updates["max_people"]; ok {
		query += ", max_people = $max_people"
		vars["max_people"] = maxPeople
	}
	if note, ok := updates["note"]; ok {
		query += ", note = $note"
		vars["note"] = note
	}
	if visibility, ok := updates["visibility"]; ok {
		query += ", visibility = $visibility"
		vars["visibility"] = visibility
	}

	query += ` WHERE id = type::record($id) RETURN AFTER`

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseAvailabilityResult(result)
}

// Delete deletes an availability
func (r *AvailabilityRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE availability WHERE id = type::record($id)`
	vars := map[string]interface{}{"id": id}

	return r.db.Execute(ctx, query, vars)
}

// CreateHangoutRequest creates a hangout request
func (r *AvailabilityRepository) CreateHangoutRequest(ctx context.Context, req *model.HangoutRequest) error {
	query := `
		CREATE hangout_request CONTENT {
			availability: type::record($availability_id),
			requester: type::record($requester_id),
			note: $note,
			status: "pending",
			created_on: time::now()
		}
	`

	vars := map[string]interface{}{
		"availability_id": req.AvailabilityID,
		"requester_id":    req.RequesterID,
		"note":            req.Note,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	req.ID = created.ID
	req.CreatedOn = created.CreatedOn
	return nil
}

// GetHangoutRequest retrieves a hangout request by ID
func (r *AvailabilityRepository) GetHangoutRequest(ctx context.Context, id string) (*model.HangoutRequest, error) {
	// Direct record access - more efficient than WHERE id =
	query := `SELECT * FROM type::record($id)`
	vars := map[string]interface{}{"id": id}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.parseHangoutRequestResult(result)
}

// GetPendingRequests retrieves pending requests for an availability
func (r *AvailabilityRepository) GetPendingRequests(ctx context.Context, availabilityID string) ([]*model.HangoutRequest, error) {
	query := `
		SELECT * FROM hangout_request
		WHERE availability = type::record($availability_id)
		AND status = "pending"
		ORDER BY created_on
	`
	vars := map[string]interface{}{"availability_id": availabilityID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseHangoutRequestsResult(result)
}

// UpdateHangoutRequestStatus updates a request status
func (r *AvailabilityRepository) UpdateHangoutRequestStatus(ctx context.Context, id, status string) error {
	query := `UPDATE hangout_request SET status = $status, responded_on = time::now() WHERE id = type::record($id)`
	vars := map[string]interface{}{
		"id":     id,
		"status": status,
	}

	return r.db.Execute(ctx, query, vars)
}

// CreateHangout creates a hangout record
func (r *AvailabilityRepository) CreateHangout(ctx context.Context, hangout *model.Hangout) error {
	query := `
		CREATE hangout CONTENT {
			participants: $participants,
			availability: $availability_id,
			hangout_type: $hangout_type,
			activity_description: $activity_description,
			scheduled_time: $scheduled_time,
			location: $location,
			is_support_session: $is_support_session,
			status: "scheduled",
			created_on: time::now(),
			updated_on: time::now()
		}
	`

	participantRefs := make([]string, len(hangout.Participants))
	copy(participantRefs, hangout.Participants)

	var locationData interface{}
	if hangout.Location != nil {
		locationData = map[string]interface{}{
			"lat":   hangout.Location.Lat,
			"lng":   hangout.Location.Lng,
			"venue": hangout.Location.Venue,
		}
	}

	vars := map[string]interface{}{
		"participants":         participantRefs,
		"availability_id":      hangout.AvailabilityID,
		"hangout_type":         hangout.HangoutType,
		"activity_description": hangout.ActivityDescription,
		"scheduled_time":       hangout.ScheduledTime,
		"location":             locationData,
		"is_support_session":   hangout.IsSupportSession,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	hangout.ID = created.ID
	hangout.CreatedOn = created.CreatedOn
	hangout.UpdatedOn = created.UpdatedOn
	return nil
}

// GetHangout retrieves a hangout by ID
func (r *AvailabilityRepository) GetHangout(ctx context.Context, id string) (*model.Hangout, error) {
	// Direct record access - more efficient than WHERE id =
	query := `SELECT * FROM type::record($id)`
	vars := map[string]interface{}{"id": id}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.parseHangoutResult(result)
}

// GetUserHangouts retrieves hangouts for a user
func (r *AvailabilityRepository) GetUserHangouts(ctx context.Context, userID string, limit int) ([]*model.Hangout, error) {
	query := `
		SELECT * FROM hangout
		WHERE type::record($user_id) IN participants
		ORDER BY scheduled_time DESC
		LIMIT $limit
	`
	vars := map[string]interface{}{
		"user_id": userID,
		"limit":   limit,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseHangoutsResult(result)
}

// UpdateHangoutStatus updates a hangout status
func (r *AvailabilityRepository) UpdateHangoutStatus(ctx context.Context, id, status string) error {
	query := `UPDATE hangout SET status = $status, updated_on = time::now() WHERE id = type::record($id)`
	vars := map[string]interface{}{
		"id":     id,
		"status": status,
	}

	return r.db.Execute(ctx, query, vars)
}

// Helper functions

func (r *AvailabilityRepository) parseAvailabilityResult(result interface{}) (*model.Availability, error) {
	if result == nil {
		return nil, database.ErrNotFound
	}

	if arr, ok := result.([]interface{}); ok {
		if len(arr) == 0 {
			return nil, database.ErrNotFound
		}
		result = arr[0]
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	av := &model.Availability{
		ID:          convertSurrealID(data["id"]),
		UserID:      convertSurrealID(data["user"]),
		Status:      model.AvailabilityStatus(getString(data, "status")),
		HangoutType: model.HangoutType(getString(data, "hangout_type")),
		MaxPeople:   getInt(data, "max_people"),
		Visibility:  getString(data, "visibility"),
	}

	if startTime := getTime(data, "start_time"); startTime != nil {
		av.StartTime = *startTime
	}
	if endTime := getTime(data, "end_time"); endTime != nil {
		av.EndTime = *endTime
	}
	if expiresAt := getTime(data, "expires_at"); expiresAt != nil {
		av.ExpiresAt = *expiresAt
	}
	if createdOn := getTime(data, "created_on"); createdOn != nil {
		av.CreatedOn = *createdOn
	}
	if updatedOn := getTime(data, "updated_on"); updatedOn != nil {
		av.UpdatedOn = *updatedOn
	}

	if locData, ok := data["location"].(map[string]interface{}); ok {
		av.Location = &model.AvailabilityLocation{
			Lat:    getFloat(locData, "lat"),
			Lng:    getFloat(locData, "lng"),
			Radius: getFloat(locData, "radius"),
		}
	}

	if desc, ok := data["activity_description"].(string); ok {
		av.ActivityDescription = &desc
	}
	if venue, ok := data["activity_venue"].(string); ok {
		av.ActivityVenue = &venue
	}
	if interestID, ok := data["interest_id"].(string); ok {
		av.InterestID = &interestID
	}
	if note, ok := data["note"].(string); ok {
		av.Note = &note
	}

	return av, nil
}

func (r *AvailabilityRepository) parseAvailabilitiesResult(result []interface{}) ([]*model.Availability, error) {
	availabilities := make([]*model.Availability, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					av, err := r.parseAvailabilityResult(item)
					if err != nil {
						continue
					}
					availabilities = append(availabilities, av)
				}
				continue
			}
		}

		av, err := r.parseAvailabilityResult(res)
		if err != nil {
			continue
		}
		availabilities = append(availabilities, av)
	}

	return availabilities, nil
}

func (r *AvailabilityRepository) parseHangoutRequestResult(result interface{}) (*model.HangoutRequest, error) {
	if result == nil {
		return nil, database.ErrNotFound
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	req := &model.HangoutRequest{
		ID:             convertSurrealID(data["id"]),
		AvailabilityID: convertSurrealID(data["availability"]),
		RequesterID:    convertSurrealID(data["requester"]),
		Note:           getString(data, "note"),
		Status:         getString(data, "status"),
	}

	if createdOn := getTime(data, "created_on"); createdOn != nil {
		req.CreatedOn = *createdOn
	}
	if respondedOn := getTime(data, "responded_on"); respondedOn != nil {
		req.RespondedOn = respondedOn
	}

	return req, nil
}

func (r *AvailabilityRepository) parseHangoutRequestsResult(result []interface{}) ([]*model.HangoutRequest, error) {
	requests := make([]*model.HangoutRequest, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					req, err := r.parseHangoutRequestResult(item)
					if err != nil {
						continue
					}
					requests = append(requests, req)
				}
			}
		}
	}

	return requests, nil
}

func (r *AvailabilityRepository) parseHangoutResult(result interface{}) (*model.Hangout, error) {
	if result == nil {
		return nil, database.ErrNotFound
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	hangout := &model.Hangout{
		ID:               convertSurrealID(data["id"]),
		HangoutType:      model.HangoutType(getString(data, "hangout_type")),
		IsSupportSession: getBool(data, "is_support_session"),
		Status:           getString(data, "status"),
	}

	if participants, ok := data["participants"].([]interface{}); ok {
		for _, p := range participants {
			hangout.Participants = append(hangout.Participants, convertSurrealID(p))
		}
	}

	if avID, ok := data["availability"].(string); ok {
		hangout.AvailabilityID = &avID
	}
	if desc, ok := data["activity_description"].(string); ok {
		hangout.ActivityDescription = &desc
	}

	if scheduledTime := getTime(data, "scheduled_time"); scheduledTime != nil {
		hangout.ScheduledTime = *scheduledTime
	}
	if createdOn := getTime(data, "created_on"); createdOn != nil {
		hangout.CreatedOn = *createdOn
	}
	if updatedOn := getTime(data, "updated_on"); updatedOn != nil {
		hangout.UpdatedOn = *updatedOn
	}

	if locData, ok := data["location"].(map[string]interface{}); ok {
		hangout.Location = &model.HangoutLocation{
			Lat: getFloat(locData, "lat"),
			Lng: getFloat(locData, "lng"),
		}
		if venue, ok := locData["venue"].(string); ok {
			hangout.Location.Venue = &venue
		}
	}

	return hangout, nil
}

func (r *AvailabilityRepository) parseHangoutsResult(result []interface{}) ([]*model.Hangout, error) {
	hangouts := make([]*model.Hangout, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					h, err := r.parseHangoutResult(item)
					if err != nil {
						continue
					}
					hangouts = append(hangouts, h)
				}
			}
		}
	}

	return hangouts, nil
}

// Nudge-related methods

// GetStaleHangouts retrieves hangouts that are past their scheduled time but still marked as the given status
func (r *AvailabilityRepository) GetStaleHangouts(ctx context.Context, cutoff time.Time, status string) ([]*model.Hangout, error) {
	query := `
		SELECT * FROM hangout
		WHERE status = $status
		AND scheduled_time < $cutoff
		ORDER BY scheduled_time ASC
		LIMIT 100
	`
	vars := map[string]interface{}{
		"status": status,
		"cutoff": cutoff,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseHangoutsResult(result)
}

// GetUpcomingHangouts retrieves hangouts scheduled within the given time window
func (r *AvailabilityRepository) GetUpcomingHangouts(ctx context.Context, windowStart, windowEnd time.Time) ([]*model.Hangout, error) {
	query := `
		SELECT * FROM hangout
		WHERE status = "scheduled"
		AND scheduled_time >= $start
		AND scheduled_time <= $end
		ORDER BY scheduled_time ASC
		LIMIT 100
	`
	vars := map[string]interface{}{
		"start": windowStart,
		"end":   windowEnd,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseHangoutsResult(result)
}

// GetAllPendingRequests retrieves all pending hangout requests (for nudge processing)
func (r *AvailabilityRepository) GetAllPendingRequests(ctx context.Context) ([]*model.HangoutRequest, error) {
	query := `
		SELECT * FROM hangout_request
		WHERE status = "pending"
		ORDER BY created_on ASC
		LIMIT 100
	`

	result, err := r.db.Query(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	return r.parseHangoutRequestsResult(result)
}

// GetPendingRequestsForUser retrieves pending hangout requests where user owns the availability
func (r *AvailabilityRepository) GetPendingRequestsForUser(ctx context.Context, userID string) ([]*model.HangoutRequest, error) {
	query := `
		SELECT * FROM hangout_request
		WHERE status = "pending"
		AND availability_id IN (
			SELECT id FROM availability WHERE user = type::record($user_id)
		)
		ORDER BY created_on ASC
	`
	vars := map[string]interface{}{
		"user_id": userID,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseHangoutRequestsResult(result)
}

// GetUserUpcomingHangouts retrieves upcoming hangouts for a specific user
func (r *AvailabilityRepository) GetUserUpcomingHangouts(ctx context.Context, userID string, windowStart, windowEnd time.Time) ([]*model.Hangout, error) {
	query := `
		SELECT * FROM hangout
		WHERE type::record($user_id) IN participants
		AND status = "scheduled"
		AND scheduled_time >= $start
		AND scheduled_time <= $end
		ORDER BY scheduled_time ASC
	`
	vars := map[string]interface{}{
		"user_id": userID,
		"start":   windowStart,
		"end":     windowEnd,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseHangoutsResult(result)
}
