package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// AdventureRepository handles adventure data access
type AdventureRepository struct {
	db database.Database
}

// NewAdventureRepository creates a new adventure repository
func NewAdventureRepository(db database.Database) *AdventureRepository {
	return &AdventureRepository{db: db}
}

// Create creates a new adventure
func (r *AdventureRepository) Create(ctx context.Context, adventure *model.Adventure) error {
	// Build query dynamically to avoid NULL values
	fields := []string{
		"title: $title",
		"start_date: <datetime> $start_date",
		"end_date: <datetime> $end_date",
		"organizer_type: $organizer_type",
		"organizer_id: $organizer_id",
		"organizer_user_id: type::record($organizer_user_id)",
		"status: $status",
		"created_by_id: type::record($created_by_id)",
		"created_on: time::now()",
	}
	vars := map[string]interface{}{
		"title":             adventure.Title,
		"start_date":        adventure.StartDate.Format(time.RFC3339),
		"end_date":          adventure.EndDate.Format(time.RFC3339),
		"organizer_type":    adventure.OrganizerType,
		"organizer_id":      adventure.OrganizerID,
		"organizer_user_id": adventure.OrganizerUserID,
		"status":            adventure.Status,
		"created_by_id":     adventure.CreatedByID,
	}

	if adventure.Description != nil {
		fields = append(fields, "description: $description")
		vars["description"] = *adventure.Description
	}
	if adventure.GuildID != nil {
		fields = append(fields, "guild_id: type::record($guild_id)")
		vars["guild_id"] = *adventure.GuildID
	}
	if adventure.Visibility != "" {
		fields = append(fields, "visibility: $visibility")
		vars["visibility"] = adventure.Visibility
	}

	query := fmt.Sprintf("CREATE adventure CONTENT { %s }", strings.Join(fields, ", "))

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return fmt.Errorf("failed to create adventure: %w", err)
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return fmt.Errorf("failed to extract created adventure: %w", err)
	}

	adventure.ID = created.ID
	adventure.CreatedOn = created.CreatedOn
	return nil
}

// GetByID retrieves an adventure by ID
func (r *AdventureRepository) GetByID(ctx context.Context, id string) (*model.Adventure, error) {
	query := `SELECT * FROM type::record($id)`
	result, err := r.db.QueryOne(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get adventure: %w", err)
	}

	return r.parseAdventure(result)
}

// Update updates an adventure
func (r *AdventureRepository) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Adventure, error) {
	// Build dynamic update
	setClause := ""
	vars := map[string]interface{}{"id": id}
	i := 0
	for key, value := range updates {
		if i > 0 {
			setClause += ", "
		}
		varName := fmt.Sprintf("v%d", i)
		setClause += fmt.Sprintf("%s = $%s", key, varName)
		vars[varName] = value
		i++
	}

	if setClause == "" {
		return r.GetByID(ctx, id)
	}

	query := fmt.Sprintf(`
		UPDATE type::record($id) SET %s, updated_on = time::now()
		RETURN AFTER
	`, setClause)

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to update adventure: %w", err)
	}

	return r.parseAdventure(result)
}

// UpdateOrganizerUser updates the organizer user for an adventure
func (r *AdventureRepository) UpdateOrganizerUser(ctx context.Context, id string, newOrganizerUserID string) (*model.Adventure, error) {
	query := `
		UPDATE type::record($id) SET
			organizer_user_id = type::record($new_organizer_user_id),
			updated_on = time::now()
		RETURN AFTER
	`
	vars := map[string]interface{}{
		"id":                    id,
		"new_organizer_user_id": newOrganizerUserID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to update organizer: %w", err)
	}

	return r.parseAdventure(result)
}

// Freeze freezes an adventure
func (r *AdventureRepository) Freeze(ctx context.Context, id string, reason string) (*model.Adventure, error) {
	query := `
		UPDATE type::record($id) SET
			status = "frozen",
			freeze_reason = $reason,
			frozen_on = time::now(),
			updated_on = time::now()
		RETURN AFTER
	`
	vars := map[string]interface{}{
		"id":     id,
		"reason": reason,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to freeze adventure: %w", err)
	}

	return r.parseAdventure(result)
}

// Unfreeze unfreezes an adventure
func (r *AdventureRepository) Unfreeze(ctx context.Context, id string) (*model.Adventure, error) {
	query := `
		UPDATE type::record($id) SET
			status = "planning",
			freeze_reason = NONE,
			frozen_on = NONE,
			updated_on = time::now()
		RETURN AFTER
	`
	vars := map[string]interface{}{"id": id}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to unfreeze adventure: %w", err)
	}

	return r.parseAdventure(result)
}

// GetByGuild retrieves adventures for a guild
func (r *AdventureRepository) GetByGuild(ctx context.Context, guildID string, limit, offset int) ([]*model.Adventure, error) {
	query := `
		SELECT * FROM adventure
		WHERE organizer_type = "guild"
		AND organizer_id = $organizer_id
		ORDER BY created_on DESC
		LIMIT $limit START $offset
	`
	vars := map[string]interface{}{
		"organizer_id": fmt.Sprintf("guild:%s", guildID),
		"limit":        limit,
		"offset":       offset,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild adventures: %w", err)
	}

	return r.parseAdventures(result)
}

// GetByUser retrieves adventures for a user (as organizer)
func (r *AdventureRepository) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*model.Adventure, error) {
	query := `
		SELECT * FROM adventure
		WHERE organizer_type = "user"
		AND organizer_user_id = type::record($user_id)
		ORDER BY created_on DESC
		LIMIT $limit START $offset
	`
	vars := map[string]interface{}{
		"user_id": userID,
		"limit":   limit,
		"offset":  offset,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get user adventures: %w", err)
	}

	return r.parseAdventures(result)
}

// Delete deletes an adventure
func (r *AdventureRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE type::record($id)`
	if err := r.db.Execute(ctx, query, map[string]interface{}{"id": id}); err != nil {
		return fmt.Errorf("failed to delete adventure: %w", err)
	}
	return nil
}

// Parsing helpers

func (r *AdventureRepository) parseAdventure(result interface{}) (*model.Adventure, error) {
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	adventure := &model.Adventure{
		ID:              convertSurrealID(data["id"]),
		Title:           getString(data, "title"),
		Description:     getStringPtr(data, "description"),
		OrganizerType:   model.AdventureOrganizerType(getString(data, "organizer_type")),
		OrganizerID:     getString(data, "organizer_id"),
		OrganizerUserID: convertSurrealID(data["organizer_user_id"]),
		Status:          model.AdventureStatus(getString(data, "status")),
		Visibility:      model.AdventureVisibility(getString(data, "visibility")),
		CreatedByID:     convertSurrealID(data["created_by_id"]),
	}

	if guildID := convertSurrealID(data["guild_id"]); guildID != "" {
		adventure.GuildID = &guildID
	}
	if reason := getString(data, "freeze_reason"); reason != "" {
		adventure.FreezeReason = &reason
	}
	if t := getTime(data, "created_on"); t != nil {
		adventure.CreatedOn = *t
	}
	if t := getTime(data, "updated_on"); t != nil {
		adventure.UpdatedOn = *t
	}
	adventure.FrozenOn = getTime(data, "frozen_on")

	return adventure, nil
}

func (r *AdventureRepository) parseAdventures(result []interface{}) ([]*model.Adventure, error) {
	adventures := make([]*model.Adventure, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					adventure, err := r.parseAdventure(item)
					if err != nil {
						continue
					}
					adventures = append(adventures, adventure)
				}
			}
		}
	}

	return adventures, nil
}
