package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// ActivityRepository handles activity data access
type ActivityRepository struct {
	db database.Database
}

// NewActivityRepository creates a new activity repository
func NewActivityRepository(db database.Database) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// Create creates a new activity
func (r *ActivityRepository) Create(ctx context.Context, activity *model.Activity) error {
	query := `
		CREATE activity SET
			guild_id = type::record($guild_id),
			name = $name,
			icon = $icon,
			warn = $warn,
			critical = $critical,
			created_on = time::now(),
			updated_on = time::now()
	`

	vars := map[string]interface{}{
		"guild_id": activity.GuildID,
		"name":     activity.Name,
		"icon":     activity.Icon,
		"warn":     activity.Warn,
		"critical": activity.Critical,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	activity.ID = created.ID
	activity.CreatedOn = created.CreatedOn
	activity.UpdatedOn = created.UpdatedOn
	return nil
}

// GetByID retrieves an activity by ID
func (r *ActivityRepository) GetByID(ctx context.Context, id string) (*model.Activity, error) {
	query := `SELECT * FROM type::record($id)`
	vars := map[string]interface{}{"id": id}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return parseActivityResult(result)
}

// GetByGuildID retrieves all activities for a guild
func (r *ActivityRepository) GetByGuildID(ctx context.Context, guildID string) ([]*model.Activity, error) {
	query := `SELECT * FROM activity WHERE guild_id = type::record($guild_id) ORDER BY name`
	vars := map[string]interface{}{"guild_id": guildID}

	results, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return parseActivitiesResult(results)
}

// Update updates an activity
func (r *ActivityRepository) Update(ctx context.Context, activity *model.Activity) error {
	query := `
		UPDATE type::record($id) SET
			name = $name,
			icon = $icon,
			warn = $warn,
			critical = $critical,
			updated_on = time::now()
	`
	vars := map[string]interface{}{
		"id":       activity.ID,
		"name":     activity.Name,
		"icon":     activity.Icon,
		"warn":     activity.Warn,
		"critical": activity.Critical,
	}

	return r.db.Execute(ctx, query, vars)
}

// Delete deletes an activity and related timers
func (r *ActivityRepository) Delete(ctx context.Context, id string) error {
	// Delete timers using this activity first
	timerQuery := `DELETE timer WHERE activity_id = type::record($id)`
	if err := r.db.Execute(ctx, timerQuery, map[string]interface{}{"id": id}); err != nil {
		return err
	}

	// Delete activity
	query := `DELETE type::record($id)`
	return r.db.Execute(ctx, query, map[string]interface{}{"id": id})
}

// CountByGuildID counts activities in a guild
func (r *ActivityRepository) CountByGuildID(ctx context.Context, guildID string) (int, error) {
	query := `SELECT count() AS count FROM activity WHERE guild_id = type::record($guild_id) GROUP ALL`
	vars := map[string]interface{}{"guild_id": guildID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return extractCount(result), nil
}

// GetByNameAndGuild finds an activity by name in a guild (for duplicate checking)
func (r *ActivityRepository) GetByNameAndGuild(ctx context.Context, name, guildID string) (*model.Activity, error) {
	query := `SELECT * FROM activity WHERE name = $name AND guild_id = type::record($guild_id) LIMIT 1`
	vars := map[string]interface{}{
		"name":     name,
		"guild_id": guildID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return parseActivityResult(result)
}

// Helper functions

func parseActivityResult(result interface{}) (*model.Activity, error) {
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
	if guildID, ok := data["guild_id"]; ok {
		data["guild_id"] = extractRecordID(guildID)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var activity model.Activity
	if err := json.Unmarshal(jsonBytes, &activity); err != nil {
		return nil, err
	}

	return &activity, nil
}

func parseActivitiesResult(results []interface{}) ([]*model.Activity, error) {
	activities := make([]*model.Activity, 0)

	for _, result := range results {
		if resp, ok := result.(map[string]interface{}); ok {
			if status, ok := resp["status"].(string); ok && status == "OK" {
				if resultData, ok := resp["result"].([]interface{}); ok {
					for _, item := range resultData {
						activity, err := parseActivityResult(item)
						if err == nil && activity != nil {
							activities = append(activities, activity)
						}
					}
				}
			}
		}
	}

	return activities, nil
}
