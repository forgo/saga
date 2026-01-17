package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

// GuildRepository handles guild data access
type GuildRepository struct {
	db database.Database
}

// NewGuildRepository creates a new guild repository
func NewGuildRepository(db database.Database) *GuildRepository {
	return &GuildRepository{db: db}
}

// Create creates a new guild
func (r *GuildRepository) Create(ctx context.Context, guild *model.Guild) error {
	visibility := guild.Visibility
	if visibility == "" {
		visibility = model.GuildVisibilityPrivate
	}

	query := `
		CREATE guild CONTENT {
			name: $name,
			description: IF $description IS NOT NULL THEN $description ELSE NONE END,
			icon: IF $icon IS NOT NULL THEN $icon ELSE NONE END,
			color: IF $color IS NOT NULL THEN $color ELSE NONE END,
			visibility: $visibility,
			created_on: time::now(),
			updated_on: time::now()
		}
	`

	vars := map[string]interface{}{
		"name":        guild.Name,
		"description": nilIfEmpty(guild.Description),
		"icon":        nilIfEmpty(guild.Icon),
		"color":       nilIfEmpty(guild.Color),
		"visibility":  visibility,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		if isUniqueConstraintError(err) {
			return fmt.Errorf("%w: guild name already exists", database.ErrDuplicate)
		}
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	guild.ID = created.ID
	guild.CreatedOn = created.CreatedOn
	guild.UpdatedOn = created.UpdatedOn
	guild.Visibility = visibility
	return nil
}

// GetByID retrieves a guild by ID
func (r *GuildRepository) GetByID(ctx context.Context, id string) (*model.Guild, error) {
	query := `SELECT * FROM type::record($id)`
	vars := map[string]interface{}{"id": id}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	guild, err := parseGuildResult(result)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return guild, nil
}

// Update updates a guild
func (r *GuildRepository) Update(ctx context.Context, guild *model.Guild) error {
	query := `
		UPDATE type::record($id) SET
			name = $name,
			description = IF $description IS NOT NULL THEN $description ELSE NONE END,
			icon = IF $icon IS NOT NULL THEN $icon ELSE NONE END,
			color = IF $color IS NOT NULL THEN $color ELSE NONE END,
			visibility = $visibility,
			updated_on = time::now()
	`
	vars := map[string]interface{}{
		"id":          guild.ID,
		"name":        guild.Name,
		"description": nilIfEmpty(guild.Description),
		"icon":        nilIfEmpty(guild.Icon),
		"color":       nilIfEmpty(guild.Color),
		"visibility":  guild.Visibility,
	}

	return r.db.Execute(ctx, query, vars)
}

// Delete deletes a guild
func (r *GuildRepository) Delete(ctx context.Context, id string) error {
	// Remove all member relationships first
	relQuery := `DELETE responsible_for WHERE out = type::record($id)`
	if err := r.db.Execute(ctx, relQuery, map[string]interface{}{"id": id}); err != nil {
		return err
	}

	// Delete guild
	query := `DELETE type::record($id)`
	return r.db.Execute(ctx, query, map[string]interface{}{"id": id})
}

// GetGuildsForUser retrieves all guilds a user is a member of
func (r *GuildRepository) GetGuildsForUser(ctx context.Context, userID string) ([]*model.Guild, error) {
	// First get the member for this user
	memberQuery := `SELECT id FROM member WHERE user = type::record($user_id) LIMIT 1`
	memberResult, err := r.db.QueryOne(ctx, memberQuery, map[string]interface{}{"user_id": userID})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return []*model.Guild{}, nil
		}
		return nil, err
	}

	memberID := extractMemberIDFromResult(memberResult)
	if memberID == "" {
		return []*model.Guild{}, nil
	}

	// Get guilds this member is responsible for
	query := `SELECT out.* AS guild FROM responsible_for WHERE in = type::record($member_id)`
	vars := map[string]interface{}{"member_id": memberID}

	results, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return parseGuildsFromRelationResult(results)
}

// CountGuildsForUser counts how many guilds a user is a member of
func (r *GuildRepository) CountGuildsForUser(ctx context.Context, userID string) (int, error) {
	// First get the member for this user
	memberQuery := `SELECT id FROM member WHERE user = type::record($user_id) LIMIT 1`
	memberResult, err := r.db.QueryOne(ctx, memberQuery, map[string]interface{}{"user_id": userID})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	memberID := extractMemberIDFromResult(memberResult)
	if memberID == "" {
		return 0, nil
	}

	// Count guilds this member is responsible for
	query := `SELECT count() AS count FROM responsible_for WHERE in = type::record($member_id) GROUP ALL`
	vars := map[string]interface{}{"member_id": memberID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return extractCount(result), nil
}

// AddMember adds a member to a guild via the responsible_for relation
func (r *GuildRepository) AddMember(ctx context.Context, memberID, guildID string, pendingApproval bool) error {
	query := `RELATE (SELECT * FROM type::record($member_id))->responsible_for->(SELECT * FROM type::record($guild_id)) SET pending_approval = $pending_approval`
	vars := map[string]interface{}{
		"member_id":        memberID,
		"guild_id":         guildID,
		"pending_approval": pendingApproval,
	}

	return r.db.Execute(ctx, query, vars)
}

// RemoveMember removes a member from a guild
func (r *GuildRepository) RemoveMember(ctx context.Context, memberID, guildID string) error {
	query := `DELETE responsible_for WHERE in = type::record($member_id) AND out = type::record($guild_id)`
	vars := map[string]interface{}{
		"member_id": memberID,
		"guild_id":  guildID,
	}

	return r.db.Execute(ctx, query, vars)
}

// IsMember checks if a user is a member of a guild
func (r *GuildRepository) IsMember(ctx context.Context, userID, guildID string) (bool, error) {
	// First get the member for this user
	memberQuery := `SELECT id FROM member WHERE user = type::record($user_id) LIMIT 1`
	memberResult, err := r.db.QueryOne(ctx, memberQuery, map[string]interface{}{"user_id": userID})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	// Extract member ID
	memberID := extractMemberIDFromResult(memberResult)
	if memberID == "" {
		return false, nil
	}

	// Then check if that member is responsible for this guild
	query := `SELECT count() AS count FROM responsible_for WHERE in = type::record($member_id) AND out = type::record($guild_id) GROUP ALL`
	vars := map[string]interface{}{
		"member_id": memberID,
		"guild_id":  guildID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	count := extractCount(result)
	return count > 0, nil
}

// extractMemberIDFromResult extracts member ID from query result
func extractMemberIDFromResult(result interface{}) string {
	if result == nil {
		return ""
	}

	// Handle unwrapped result from QueryOne (direct record)
	if data, ok := result.(map[string]interface{}); ok {
		// Check if this is a wrapped response or direct record
		if _, hasStatus := data["status"]; hasStatus {
			// Wrapped response format
			if status, ok := data["status"].(string); ok && status == "OK" {
				if resultData, ok := data["result"].([]interface{}); ok {
					if len(resultData) == 0 {
						return ""
					}
					if record, ok := resultData[0].(map[string]interface{}); ok {
						if id, ok := record["id"]; ok {
							return convertGuildID(id)
						}
					}
				}
			}
		} else {
			// Direct record format (from QueryOne)
			if id, ok := data["id"]; ok {
				return convertGuildID(id)
			}
		}
	}

	return ""
}

// CountMembers counts members in a guild
func (r *GuildRepository) CountMembers(ctx context.Context, guildID string) (int, error) {
	query := `SELECT count() AS count FROM responsible_for WHERE out = type::record($guild_id) GROUP ALL`
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

// GetMembers retrieves all members of a guild
func (r *GuildRepository) GetMembers(ctx context.Context, guildID string) ([]*model.Member, error) {
	query := `SELECT in.* AS member FROM responsible_for WHERE out = type::record($guild_id)`
	vars := map[string]interface{}{"guild_id": guildID}

	results, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return parseMembersFromRelationResult(results)
}

// GetMemberForUserInGuild gets the member record for a user in a specific guild
func (r *GuildRepository) GetMemberForUserInGuild(ctx context.Context, userID, guildID string) (*model.Member, error) {
	// First get the member for this user
	memberQuery := `SELECT * FROM member WHERE user = type::record($user_id) LIMIT 1`
	memberResult, err := r.db.QueryOne(ctx, memberQuery, map[string]interface{}{"user_id": userID})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	member, err := parseMemberResult(memberResult)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if member == nil {
		return nil, nil
	}

	// Check if this member is in the guild
	checkQuery := `SELECT count() AS count FROM responsible_for WHERE in = type::record($member_id) AND out = type::record($guild_id) GROUP ALL`
	checkResult, err := r.db.QueryOne(ctx, checkQuery, map[string]interface{}{
		"member_id": member.ID,
		"guild_id":  guildID,
	})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	count := extractCount(checkResult)
	if count == 0 {
		return nil, nil
	}

	return member, nil
}

// Helper functions

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func parseGuildResult(result interface{}) (*model.Guild, error) {
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

	// Handle SurrealDB 3 record ID format
	if id, ok := data["id"]; ok {
		data["id"] = convertGuildID(id)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var guild model.Guild
	if err := json.Unmarshal(jsonBytes, &guild); err != nil {
		return nil, err
	}

	return &guild, nil
}

func parseGuildsFromRelationResult(results []interface{}) ([]*model.Guild, error) {
	guilds := make([]*model.Guild, 0)

	for _, result := range results {
		if resp, ok := result.(map[string]interface{}); ok {
			if status, ok := resp["status"].(string); ok && status == "OK" {
				if resultData, ok := resp["result"].([]interface{}); ok {
					for _, item := range resultData {
						if data, ok := item.(map[string]interface{}); ok {
							if guildData, ok := data["guild"].(map[string]interface{}); ok {
								guild, err := parseGuildFromData(guildData)
								if err == nil {
									guilds = append(guilds, guild)
								}
							}
						}
					}
				}
			}
		}
	}

	return guilds, nil
}

func parseGuildFromData(data map[string]interface{}) (*model.Guild, error) {
	if id, ok := data["id"]; ok {
		data["id"] = convertGuildID(id)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var guild model.Guild
	if err := json.Unmarshal(jsonBytes, &guild); err != nil {
		return nil, err
	}

	return &guild, nil
}

func parseMembersFromRelationResult(results []interface{}) ([]*model.Member, error) {
	members := make([]*model.Member, 0)

	for _, result := range results {
		if resp, ok := result.(map[string]interface{}); ok {
			if status, ok := resp["status"].(string); ok && status == "OK" {
				if resultData, ok := resp["result"].([]interface{}); ok {
					for _, item := range resultData {
						if data, ok := item.(map[string]interface{}); ok {
							if memberData, ok := data["member"].(map[string]interface{}); ok {
								member, err := parseMemberFromData(memberData)
								if err == nil {
									members = append(members, member)
								}
							}
						}
					}
				}
			}
		}
	}

	return members, nil
}

func parseMemberFromData(data map[string]interface{}) (*model.Member, error) {
	if id, ok := data["id"]; ok {
		data["id"] = convertGuildID(id)
	}
	if userID, ok := data["user"]; ok {
		data["user_id"] = convertGuildID(userID)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var member model.Member
	if err := json.Unmarshal(jsonBytes, &member); err != nil {
		return nil, err
	}

	return &member, nil
}

// convertGuildID converts a SurrealDB ID to a string
func convertGuildID(id interface{}) string {
	if str, ok := id.(string); ok {
		return str
	}
	if rid, ok := id.(models.RecordID); ok {
		return fmt.Sprintf("%s:%v", rid.Table, rid.ID)
	}
	if rid, ok := id.(*models.RecordID); ok && rid != nil {
		return fmt.Sprintf("%s:%v", rid.Table, rid.ID)
	}
	if m, ok := id.(map[string]interface{}); ok {
		if tb, ok := m["tb"].(string); ok {
			if idVal, ok := m["id"]; ok {
				return fmt.Sprintf("%s:%v", tb, idVal)
			}
		}
		if tb, ok := m["Table"].(string); ok {
			if idVal, ok := m["ID"]; ok {
				return fmt.Sprintf("%s:%v", tb, idVal)
			}
		}
	}
	return fmt.Sprintf("%v", id)
}
