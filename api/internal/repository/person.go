package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// PersonRepository handles person data access
type PersonRepository struct {
	db database.Database
}

// NewPersonRepository creates a new person repository
func NewPersonRepository(db database.Database) *PersonRepository {
	return &PersonRepository{db: db}
}

// Create creates a new person
func (r *PersonRepository) Create(ctx context.Context, person *model.Person) error {
	// Build query dynamically to avoid NULL vs NONE issues for optional fields
	setClause := `guild_id = type::record($guild_id), name = $name, created_on = time::now(), updated_on = time::now()`
	vars := map[string]interface{}{
		"guild_id": person.GuildID,
		"name":     person.Name,
	}

	// Add optional fields only when they have values
	if person.Nickname != "" {
		setClause += ", nickname = $nickname"
		vars["nickname"] = person.Nickname
	}
	if person.Birthday != nil {
		setClause += ", birthday = $birthday"
		vars["birthday"] = person.Birthday
	}
	if person.Notes != "" {
		setClause += ", notes = $notes"
		vars["notes"] = person.Notes
	}
	if person.Avatar != "" {
		setClause += ", avatar = $avatar"
		vars["avatar"] = person.Avatar
	}

	query := "CREATE person SET " + setClause
	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	person.ID = created.ID
	person.CreatedOn = created.CreatedOn
	person.UpdatedOn = created.UpdatedOn
	return nil
}

// GetByID retrieves a person by ID
func (r *PersonRepository) GetByID(ctx context.Context, id string) (*model.Person, error) {
	query := `SELECT * FROM type::record($id)`
	vars := map[string]interface{}{"id": id}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return parsePersonResult(result)
}

// GetByGuildID retrieves all people for a guild
func (r *PersonRepository) GetByGuildID(ctx context.Context, guildID string) ([]*model.Person, error) {
	query := `SELECT * FROM person WHERE guild_id = type::record($guild_id) ORDER BY name`
	vars := map[string]interface{}{"guild_id": guildID}

	results, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return parsePeopleResult(results)
}

// Update updates a person
func (r *PersonRepository) Update(ctx context.Context, person *model.Person) error {
	// Build query dynamically to avoid NULL vs NONE issues for optional fields
	setClause := `name = $name, updated_on = time::now()`
	vars := map[string]interface{}{
		"id":   person.ID,
		"name": person.Name,
	}

	// Add optional fields only when they have values
	if person.Nickname != "" {
		setClause += ", nickname = $nickname"
		vars["nickname"] = person.Nickname
	}
	if person.Birthday != nil {
		setClause += ", birthday = $birthday"
		vars["birthday"] = person.Birthday
	}
	if person.Notes != "" {
		setClause += ", notes = $notes"
		vars["notes"] = person.Notes
	}
	if person.Avatar != "" {
		setClause += ", avatar = $avatar"
		vars["avatar"] = person.Avatar
	}

	query := "UPDATE type::record($id) SET " + setClause
	return r.db.Execute(ctx, query, vars)
}

// Delete deletes a person and their timers
func (r *PersonRepository) Delete(ctx context.Context, id string) error {
	// Delete timers first
	timerQuery := `DELETE timer WHERE person_id = type::record($id)`
	if err := r.db.Execute(ctx, timerQuery, map[string]interface{}{"id": id}); err != nil {
		return err
	}

	// Delete person
	query := `DELETE type::record($id)`
	return r.db.Execute(ctx, query, map[string]interface{}{"id": id})
}

// CountByGuildID counts people in a guild
func (r *PersonRepository) CountByGuildID(ctx context.Context, guildID string) (int, error) {
	query := `SELECT count() AS count FROM person WHERE guild_id = type::record($guild_id) GROUP ALL`
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

// GetPersonWithTimers retrieves a person with all their timers and activities
func (r *PersonRepository) GetPersonWithTimers(ctx context.Context, personID string) (*model.PersonWithTimers, error) {
	person, err := r.GetByID(ctx, personID)
	if err != nil {
		return nil, err
	}
	if person == nil {
		return nil, nil
	}

	// Get timers with activities
	query := `
		SELECT
			timer.*,
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

	timers := parseTimersWithActivities(results)

	return &model.PersonWithTimers{
		Person: *person,
		Timers: timers,
	}, nil
}

// Helper functions

func parsePersonResult(result interface{}) (*model.Person, error) {
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

	var person model.Person
	if err := json.Unmarshal(jsonBytes, &person); err != nil {
		return nil, err
	}

	return &person, nil
}

func parsePeopleResult(results []interface{}) ([]*model.Person, error) {
	people := make([]*model.Person, 0)

	for _, result := range results {
		if resp, ok := result.(map[string]interface{}); ok {
			if status, ok := resp["status"].(string); ok && status == "OK" {
				if resultData, ok := resp["result"].([]interface{}); ok {
					for _, item := range resultData {
						person, err := parsePersonResult(item)
						if err == nil && person != nil {
							people = append(people, person)
						}
					}
				}
			}
		}
	}

	return people, nil
}

func parseTimersWithActivities(results []interface{}) []model.TimerWithActivity {
	timers := make([]model.TimerWithActivity, 0)

	for _, result := range results {
		if resp, ok := result.(map[string]interface{}); ok {
			if status, ok := resp["status"].(string); ok && status == "OK" {
				if resultData, ok := resp["result"].([]interface{}); ok {
					for _, item := range resultData {
						if data, ok := item.(map[string]interface{}); ok {
							var timer model.Timer
							var activity model.Activity

							// Parse timer fields
							timerBytes, _ := json.Marshal(data)
							_ = json.Unmarshal(timerBytes, &timer)

							// Parse activity
							if actData, ok := data["activity"].(map[string]interface{}); ok {
								actBytes, _ := json.Marshal(actData)
								_ = json.Unmarshal(actBytes, &activity)
							}

							timers = append(timers, model.TimerWithActivity{
								Timer:    timer,
								Activity: activity,
							})
						}
					}
				}
			}
		}
	}

	return timers
}
