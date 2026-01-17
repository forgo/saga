package repository

import (
	"context"
	"errors"
	"time"

	"github.com/forgo/saga/api/internal/database"
)

// DiscoveryRepository handles discovery-related data access
type DiscoveryRepository struct {
	db database.Database
}

// DiscoveryDailyCount tracks a user's daily discovery usage
type DiscoveryDailyCount struct {
	UserID      string    `json:"user_id"`
	Date        string    `json:"date"` // YYYY-MM-DD
	PeopleShown int       `json:"people_shown"`
	EventsShown int       `json:"events_shown"`
	GuildsShown int       `json:"guilds_shown"`
	CreatedOn   time.Time `json:"created_on"`
	UpdatedOn   time.Time `json:"updated_on"`
}

// NewDiscoveryRepository creates a new discovery repository
func NewDiscoveryRepository(db database.Database) *DiscoveryRepository {
	return &DiscoveryRepository{db: db}
}

// GetDailyCount retrieves a user's discovery count for a date
func (r *DiscoveryRepository) GetDailyCount(ctx context.Context, userID, date string) (*DiscoveryDailyCount, error) {
	query := `
		SELECT * FROM discovery_daily_count
		WHERE user_id = type::record($user_id) AND date = $date
		LIMIT 1
	`
	vars := map[string]interface{}{
		"user_id": userID,
		"date":    date,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// Return empty count for today
			return &DiscoveryDailyCount{
				UserID: userID,
				Date:   date,
			}, nil
		}
		return nil, err
	}

	return r.parseDailyCountResult(result)
}

// IncrementCount increments a discovery count for a user
func (r *DiscoveryRepository) IncrementCount(ctx context.Context, userID, date, countType string) error {
	field := ""
	switch countType {
	case "people":
		field = "people_shown"
	case "events":
		field = "events_shown"
	case "guilds":
		field = "guilds_shown"
	default:
		return errors.New("invalid count type")
	}

	// Try to update existing record first
	updateQuery := `
		UPDATE discovery_daily_count
		SET ` + field + ` = ` + field + ` + 1, updated_on = time::now()
		WHERE user_id = type::record($user_id) AND date = $date
		RETURN AFTER
	`
	vars := map[string]interface{}{
		"user_id": userID,
		"date":    date,
	}

	result, err := r.db.QueryOne(ctx, updateQuery, vars)
	if err == nil && result != nil {
		return nil
	}

	// Record doesn't exist, create it
	createQuery := `
		CREATE discovery_daily_count CONTENT {
			user_id: type::record($user_id),
			date: $date,
			people_shown: $people,
			events_shown: $events,
			guilds_shown: $guilds,
			created_on: time::now(),
			updated_on: time::now()
		}
	`
	vars["people"] = 0
	vars["events"] = 0
	vars["guilds"] = 0

	switch countType {
	case "people":
		vars["people"] = 1
	case "events":
		vars["events"] = 1
	case "guilds":
		vars["guilds"] = 1
	}

	_, err = r.db.Query(ctx, createQuery, vars)
	return err
}

// HasQuota checks if user has remaining discovery quota for a type
func (r *DiscoveryRepository) HasQuota(ctx context.Context, userID, date, countType string, limit int) (bool, error) {
	count, err := r.GetDailyCount(ctx, userID, date)
	if err != nil {
		return false, err
	}

	switch countType {
	case "people":
		return count.PeopleShown < limit, nil
	case "events":
		return count.EventsShown < limit, nil
	case "guilds":
		return count.GuildsShown < limit, nil
	default:
		return true, nil
	}
}

// GetRemainingQuota returns remaining quota for each type
func (r *DiscoveryRepository) GetRemainingQuota(ctx context.Context, userID, date string, limits map[string]int) (map[string]int, error) {
	count, err := r.GetDailyCount(ctx, userID, date)
	if err != nil {
		return nil, err
	}

	remaining := make(map[string]int)

	if limit, ok := limits["people"]; ok {
		remaining["people"] = max(0, limit-count.PeopleShown)
	}
	if limit, ok := limits["events"]; ok {
		remaining["events"] = max(0, limit-count.EventsShown)
	}
	if limit, ok := limits["guilds"]; ok {
		remaining["guilds"] = max(0, limit-count.GuildsShown)
	}

	return remaining, nil
}

// CleanupOldCounts removes discovery counts older than retention days
func (r *DiscoveryRepository) CleanupOldCounts(ctx context.Context, retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays).Format("2006-01-02")

	query := `DELETE discovery_daily_count WHERE date < $cutoff_date`
	vars := map[string]interface{}{"cutoff_date": cutoffDate}

	return r.db.Execute(ctx, query, vars)
}

// Helper functions

func (r *DiscoveryRepository) parseDailyCountResult(result interface{}) (*DiscoveryDailyCount, error) {
	if result == nil {
		return nil, database.ErrNotFound
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	count := &DiscoveryDailyCount{
		UserID:      convertSurrealID(data["user_id"]),
		Date:        getString(data, "date"),
		PeopleShown: getInt(data, "people_shown"),
		EventsShown: getInt(data, "events_shown"),
		GuildsShown: getInt(data, "guilds_shown"),
	}

	if t := getTime(data, "created_on"); t != nil {
		count.CreatedOn = *t
	}
	if t := getTime(data, "updated_on"); t != nil {
		count.UpdatedOn = *t
	}

	return count, nil
}

// Discovery limit constants
const (
	DefaultDailyPeopleLimit = 10
	DefaultDailyEventsLimit = 20
	DefaultDailyGuildsLimit = 10
)

// GetDefaultLimits returns the default discovery limits
func GetDefaultLimits() map[string]int {
	return map[string]int{
		"people": DefaultDailyPeopleLimit,
		"events": DefaultDailyEventsLimit,
		"guilds": DefaultDailyGuildsLimit,
	}
}
