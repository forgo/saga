package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// TimerRepository handles timer data access
type TimerRepository struct {
	db database.Database
}

// NewTimerRepository creates a new timer repository
func NewTimerRepository(db database.Database) *TimerRepository {
	return &TimerRepository{db: db}
}

// Create creates a new timer
func (r *TimerRepository) Create(ctx context.Context, timer *model.Timer) error {
	query := `
		CREATE timer SET
			person_id = type::record($person_id),
			activity_id = type::record($activity_id),
			reset_date = <datetime>$reset_date,
			enabled = $enabled,
			push = $push,
			created_on = time::now(),
			updated_on = time::now()
	`

	vars := map[string]interface{}{
		"person_id":   timer.PersonID,
		"activity_id": timer.ActivityID,
		"reset_date":  timer.ResetDate.Format(time.RFC3339),
		"enabled":     timer.Enabled,
		"push":        timer.Push,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	timer.ID = created.ID
	timer.CreatedOn = created.CreatedOn
	timer.UpdatedOn = created.UpdatedOn
	return nil
}

// GetByID retrieves a timer by ID
func (r *TimerRepository) GetByID(ctx context.Context, id string) (*model.Timer, error) {
	query := `SELECT * FROM type::record($id)`
	vars := map[string]interface{}{"id": id}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return parseTimerResult(result)
}

// GetByPersonID retrieves all timers for a person
func (r *TimerRepository) GetByPersonID(ctx context.Context, personID string) ([]*model.Timer, error) {
	query := `SELECT * FROM timer WHERE person_id = type::record($person_id)`
	vars := map[string]interface{}{"person_id": personID}

	results, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return parseTimersResult(results)
}

// GetByPersonAndActivity retrieves a timer for a specific person and activity
func (r *TimerRepository) GetByPersonAndActivity(ctx context.Context, personID, activityID string) (*model.Timer, error) {
	query := `SELECT * FROM timer WHERE person_id = type::record($person_id) AND activity_id = type::record($activity_id) LIMIT 1`
	vars := map[string]interface{}{
		"person_id":   personID,
		"activity_id": activityID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return parseTimerResult(result)
}

// Update updates a timer
func (r *TimerRepository) Update(ctx context.Context, timer *model.Timer) error {
	query := `
		UPDATE type::record($id) SET
			enabled = $enabled,
			push = $push,
			updated_on = time::now()
	`
	vars := map[string]interface{}{
		"id":      timer.ID,
		"enabled": timer.Enabled,
		"push":    timer.Push,
	}

	return r.db.Execute(ctx, query, vars)
}

// Reset resets a timer's reset_date to now
func (r *TimerRepository) Reset(ctx context.Context, id string) (*model.Timer, error) {
	now := time.Now().UTC()
	query := `UPDATE type::record($id) SET reset_date = <datetime>$reset_date, updated_on = time::now() RETURN AFTER`
	vars := map[string]interface{}{
		"id":         id,
		"reset_date": now.Format(time.RFC3339),
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return parseTimerResult(result)
}

// Delete deletes a timer
func (r *TimerRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE type::record($id)`
	return r.db.Execute(ctx, query, map[string]interface{}{"id": id})
}

// CountByPersonID counts timers for a person
func (r *TimerRepository) CountByPersonID(ctx context.Context, personID string) (int, error) {
	query := `SELECT count() AS count FROM timer WHERE person_id = type::record($person_id) GROUP ALL`
	vars := map[string]interface{}{"person_id": personID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return extractCount(result), nil
}

// GetTimerWithActivity retrieves a timer with its activity
func (r *TimerRepository) GetTimerWithActivity(ctx context.Context, timerID string) (*model.TimerWithActivity, error) {
	timer, err := r.GetByID(ctx, timerID)
	if err != nil {
		return nil, err
	}
	if timer == nil {
		return nil, nil
	}

	// Get activity
	activityQuery := `SELECT * FROM type::record($id)`
	actResult, err := r.db.QueryOne(ctx, activityQuery, map[string]interface{}{"id": timer.ActivityID})
	if err != nil {
		return nil, err
	}

	activity, err := parseActivityResult(actResult)
	if err != nil {
		return nil, err
	}

	return &model.TimerWithActivity{
		Timer:    *timer,
		Activity: *activity,
	}, nil
}

// GetTimersWithActivities retrieves all timers for a person with their activities
func (r *TimerRepository) GetTimersWithActivities(ctx context.Context, personID string) ([]model.TimerWithActivity, error) {
	query := `
		SELECT
			*,
			(SELECT * FROM activity WHERE id = timer.activity_id)[0] AS activity
		FROM timer
		WHERE person_id = type::record($person_id)
		ORDER BY activity.name
	`
	vars := map[string]interface{}{"person_id": personID}

	results, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return parseTimersWithActivities(results), nil
}

// Helper functions

func parseTimerResult(result interface{}) (*model.Timer, error) {
	if result == nil {
		return nil, database.ErrNotFound
	}

	// Navigate through SurrealDB response structure
	if resp, ok := result.(map[string]interface{}); ok {
		if status, ok := resp["status"].(string); ok && status == "OK" {
			if resultData, ok := resp["result"].([]interface{}); ok {
				if len(resultData) == 0 {
					return nil, database.ErrNotFound
				}
				result = resultData[0]
			}
		}
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

	// Convert record IDs to strings before JSON marshaling
	if id, ok := data["id"]; ok {
		data["id"] = extractRecordID(id)
	}
	if personID, ok := data["person_id"]; ok {
		data["person_id"] = extractRecordID(personID)
	}
	if activityID, ok := data["activity_id"]; ok {
		data["activity_id"] = extractRecordID(activityID)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var timer model.Timer
	if err := json.Unmarshal(jsonBytes, &timer); err != nil {
		return nil, err
	}

	return &timer, nil
}

func parseTimersResult(results []interface{}) ([]*model.Timer, error) {
	timers := make([]*model.Timer, 0)

	for _, result := range results {
		if resp, ok := result.(map[string]interface{}); ok {
			if status, ok := resp["status"].(string); ok && status == "OK" {
				if resultData, ok := resp["result"].([]interface{}); ok {
					for _, item := range resultData {
						timer, err := parseTimerResult(item)
						if err == nil && timer != nil {
							timers = append(timers, timer)
						}
					}
				}
			}
		}
	}

	return timers, nil
}

// GetAllActiveTimersWithActivities retrieves all enabled timers with their activities
func (r *TimerRepository) GetAllActiveTimersWithActivities(ctx context.Context) ([]model.TimerWithActivity, error) {
	query := `
		SELECT
			*,
			(SELECT * FROM activity WHERE id = timer.activity_id)[0] AS activity
		FROM timer
		WHERE enabled = true
	`

	results, err := r.db.Query(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	return parseTimersWithActivities(results), nil
}

// GetCircleIDForPerson retrieves the circle ID for a person
func (r *TimerRepository) GetCircleIDForPerson(ctx context.Context, personID string) (string, error) {
	query := `SELECT circle_id FROM $id`
	vars := map[string]interface{}{"id": personID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return "", err
	}

	if resp, ok := result.(map[string]interface{}); ok {
		if status, ok := resp["status"].(string); ok && status == "OK" {
			if resultData, ok := resp["result"].([]interface{}); ok && len(resultData) > 0 {
				if data, ok := resultData[0].(map[string]interface{}); ok {
					if circleID, ok := data["circle_id"].(string); ok {
						return circleID, nil
					}
				}
			}
		}
	}

	if arr, ok := result.([]interface{}); ok && len(arr) > 0 {
		if data, ok := arr[0].(map[string]interface{}); ok {
			if circleID, ok := data["circle_id"].(string); ok {
				return circleID, nil
			}
		}
	}

	return "", errors.New("circle_id not found")
}
