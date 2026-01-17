package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

// PoolRepository handles matching pool data access
type PoolRepository struct {
	db database.Database
}

// NewPoolRepository creates a new pool repository
func NewPoolRepository(db database.Database) *PoolRepository {
	return &PoolRepository{db: db}
}

// CreatePool creates a new matching pool
func (r *PoolRepository) CreatePool(ctx context.Context, pool *model.MatchingPool) error {
	// Build query dynamically to avoid NULL vs NONE issues for optional fields
	setClause := `guild_id = type::record($guild_id), name = $name, frequency = $frequency, match_size = $match_size, next_match_on = $next_match_on, active = true, created_by = type::record($created_by), created_on = time::now(), updated_on = time::now()`
	vars := map[string]interface{}{
		"guild_id":      pool.GuildID,
		"name":          pool.Name,
		"frequency":     pool.Frequency,
		"match_size":    pool.MatchSize,
		"next_match_on": pool.NextMatchOn,
		"created_by":    pool.CreatedBy,
	}

	// Only include optional fields if provided
	if pool.Description != nil && *pool.Description != "" {
		setClause += ", description = $description"
		vars["description"] = *pool.Description
	}
	if pool.ActivitySuggestion != nil && *pool.ActivitySuggestion != "" {
		setClause += ", activity_suggestion = $activity_suggestion"
		vars["activity_suggestion"] = *pool.ActivitySuggestion
	}

	query := "CREATE matching_pool SET " + setClause
	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return fmt.Errorf("failed to create pool: %w", err)
	}

	created, err := extractPoolCreatedRecord(result)
	if err != nil {
		return fmt.Errorf("failed to extract created pool: %w", err)
	}

	pool.ID = created.ID
	pool.CreatedOn = created.CreatedOn
	pool.UpdatedOn = created.UpdatedOn
	pool.Active = true
	return nil
}

// GetPool retrieves a pool by ID
func (r *PoolRepository) GetPool(ctx context.Context, poolID string) (*model.MatchingPool, error) {
	query := `
		SELECT *,
			(SELECT count() FROM pool_member WHERE pool_id = $parent.id AND active = true GROUP ALL)[0].count AS member_count
		FROM matching_pool WHERE id = type::record($id)
	`
	result, err := r.db.QueryOne(ctx, query, map[string]interface{}{"id": poolID})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get pool: %w", err)
	}

	return parsePoolResult(result)
}

// GetPoolsByGuild retrieves all pools for a guild
func (r *PoolRepository) GetPoolsByGuild(ctx context.Context, guildID string) ([]*model.MatchingPool, error) {
	query := `
		SELECT *,
			(SELECT count() FROM pool_member WHERE pool_id = $parent.id AND active = true GROUP ALL)[0].count AS member_count
		FROM matching_pool
		WHERE guild_id = $guild_id
		ORDER BY created_on DESC
	`
	result, err := r.db.Query(ctx, query, map[string]interface{}{"guild_id": guildID})
	if err != nil {
		return nil, fmt.Errorf("failed to get pools: %w", err)
	}

	return parsePoolsResult(result)
}

// UpdatePool updates a pool
func (r *PoolRepository) UpdatePool(ctx context.Context, poolID string, updates map[string]interface{}) (*model.MatchingPool, error) {
	updates["updated_on"] = time.Now()

	// Build dynamic SET clause
	setClause := ""
	for key := range updates {
		if setClause != "" {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%s", key, key)
	}

	query := fmt.Sprintf("UPDATE type::record($id) SET %s", setClause)
	updates["id"] = poolID

	if err := r.db.Execute(ctx, query, updates); err != nil {
		return nil, fmt.Errorf("failed to update pool: %w", err)
	}

	return r.GetPool(ctx, poolID)
}

// DeletePool deletes a pool and its memberships
func (r *PoolRepository) DeletePool(ctx context.Context, poolID string) error {
	queries := []string{
		`DELETE pool_member WHERE pool_id = type::record($pool_id)`,
		`DELETE match_result WHERE pool_id = type::record($pool_id)`,
		`DELETE matching_pool WHERE id = type::record($pool_id)`,
	}

	for _, q := range queries {
		if err := r.db.Execute(ctx, q, map[string]interface{}{"pool_id": poolID}); err != nil {
			return fmt.Errorf("failed to delete pool: %w", err)
		}
	}
	return nil
}

// AddMember adds a member to a pool
func (r *PoolRepository) AddMember(ctx context.Context, member *model.PoolMember) error {
	query := `
		CREATE pool_member CONTENT {
			pool_id: $pool_id,
			member_id: $member_id,
			user_id: $user_id,
			active: true,
			excluded_members: $excluded_members,
			joined_on: time::now()
		}
	`
	result, err := r.db.Query(ctx, query, map[string]interface{}{
		"pool_id":          member.PoolID,
		"member_id":        member.MemberID,
		"user_id":          member.UserID,
		"excluded_members": member.ExcludedMembers,
	})
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	created, err := extractPoolCreatedRecord(result)
	if err != nil {
		return fmt.Errorf("failed to extract created member: %w", err)
	}

	member.ID = created.ID
	member.JoinedOn = created.CreatedOn
	member.Active = true
	return nil
}

// GetMember retrieves a pool member
func (r *PoolRepository) GetMember(ctx context.Context, poolID, memberID string) (*model.PoolMember, error) {
	query := `
		SELECT * FROM pool_member
		WHERE pool_id = $pool_id AND member_id = $member_id
	`
	result, err := r.db.QueryOne(ctx, query, map[string]interface{}{
		"pool_id":   poolID,
		"member_id": memberID,
	})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	return parsePoolMemberResult(result)
}

// GetMemberByUser retrieves a pool member by user ID
func (r *PoolRepository) GetMemberByUser(ctx context.Context, poolID, userID string) (*model.PoolMember, error) {
	query := `
		SELECT * FROM pool_member
		WHERE pool_id = $pool_id AND user_id = $user_id
	`
	result, err := r.db.QueryOne(ctx, query, map[string]interface{}{
		"pool_id": poolID,
		"user_id": userID,
	})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	return parsePoolMemberResult(result)
}

// GetPoolMembers retrieves all active members of a pool
func (r *PoolRepository) GetPoolMembers(ctx context.Context, poolID string) ([]*model.PoolMember, error) {
	query := `
		SELECT * FROM pool_member
		WHERE pool_id = $pool_id AND active = true
		ORDER BY joined_on ASC
	`
	result, err := r.db.Query(ctx, query, map[string]interface{}{"pool_id": poolID})
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}

	return parsePoolMembersResult(result)
}

// UpdateMember updates a pool member
func (r *PoolRepository) UpdateMember(ctx context.Context, membershipID string, updates map[string]interface{}) (*model.PoolMember, error) {
	// Build dynamic SET clause
	setClause := ""
	for key := range updates {
		if setClause != "" {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%s", key, key)
	}

	query := fmt.Sprintf("UPDATE type::record($id) SET %s", setClause)
	updates["id"] = membershipID

	if err := r.db.Execute(ctx, query, updates); err != nil {
		return nil, fmt.Errorf("failed to update member: %w", err)
	}

	// Retrieve updated record (direct record access)
	getQuery := `SELECT * FROM type::record($id)`
	result, err := r.db.QueryOne(ctx, getQuery, map[string]interface{}{"id": membershipID})
	if err != nil {
		return nil, fmt.Errorf("failed to get updated member: %w", err)
	}

	return parsePoolMemberResult(result)
}

// RemoveMember removes a member from a pool (soft delete)
func (r *PoolRepository) RemoveMember(ctx context.Context, membershipID string) error {
	query := `UPDATE type::record($id) SET active = false`
	return r.db.Execute(ctx, query, map[string]interface{}{"id": membershipID})
}

// CreateMatchResult creates a new match result
func (r *PoolRepository) CreateMatchResult(ctx context.Context, match *model.MatchResult) error {
	query := `
		CREATE match_result CONTENT {
			pool_id: $pool_id,
			members: $members,
			member_user_ids: $member_user_ids,
			status: $status,
			match_round: $match_round,
			created_on: time::now(),
			updated_on: time::now()
		}
	`
	result, err := r.db.Query(ctx, query, map[string]interface{}{
		"pool_id":         match.PoolID,
		"members":         match.Members,
		"member_user_ids": match.MemberUserIDs,
		"status":          match.Status,
		"match_round":     match.MatchRound,
	})
	if err != nil {
		return fmt.Errorf("failed to create match: %w", err)
	}

	created, err := extractPoolCreatedRecord(result)
	if err != nil {
		return fmt.Errorf("failed to extract created match: %w", err)
	}

	match.ID = created.ID
	match.CreatedOn = created.CreatedOn
	match.UpdatedOn = created.UpdatedOn
	return nil
}

// GetMatchResult retrieves a match result by ID
func (r *PoolRepository) GetMatchResult(ctx context.Context, matchID string) (*model.MatchResult, error) {
	// Direct record access - more efficient than WHERE id =
	query := `SELECT * FROM type::record($id)`
	result, err := r.db.QueryOne(ctx, query, map[string]interface{}{"id": matchID})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	return parseMatchResult(result)
}

// GetMatchesByPool retrieves matches for a pool
func (r *PoolRepository) GetMatchesByPool(ctx context.Context, poolID string, limit int) ([]*model.MatchResult, error) {
	query := `
		SELECT * FROM match_result
		WHERE pool_id = $pool_id
		ORDER BY created_on DESC
		LIMIT $limit
	`
	result, err := r.db.Query(ctx, query, map[string]interface{}{
		"pool_id": poolID,
		"limit":   limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get matches: %w", err)
	}

	return parseMatchResultsFromQuery(result)
}

// GetMatchesByRound retrieves matches for a specific round
func (r *PoolRepository) GetMatchesByRound(ctx context.Context, poolID, round string) ([]*model.MatchResult, error) {
	query := `
		SELECT * FROM match_result
		WHERE pool_id = $pool_id AND match_round = $round
		ORDER BY created_on ASC
	`
	result, err := r.db.Query(ctx, query, map[string]interface{}{
		"pool_id": poolID,
		"round":   round,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get matches: %w", err)
	}

	return parseMatchResultsFromQuery(result)
}

// GetUserPendingMatches retrieves pending matches for a user
func (r *PoolRepository) GetUserPendingMatches(ctx context.Context, userID string) ([]*model.MatchResult, error) {
	query := `
		SELECT * FROM match_result
		WHERE $user_id IN member_user_ids AND status = 'pending'
		ORDER BY created_on DESC
	`
	result, err := r.db.Query(ctx, query, map[string]interface{}{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to get pending matches: %w", err)
	}

	return parseMatchResultsFromQuery(result)
}

// GetRecentMatchesBetween gets recent matches between specific members
func (r *PoolRepository) GetRecentMatchesBetween(ctx context.Context, memberIDs []string, days int) ([]*model.MatchResult, error) {
	// Build query to find matches that contain ALL specified member IDs
	query := `
		SELECT * FROM match_result
		WHERE created_on > time::now() - $days_duration
		ORDER BY created_on DESC
	`
	result, err := r.db.Query(ctx, query, map[string]interface{}{
		"days_duration": fmt.Sprintf("%dd", days),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get recent matches: %w", err)
	}

	allMatches, err := parseMatchResultsFromQuery(result)
	if err != nil {
		return nil, err
	}

	// Filter to matches containing all specified members
	var filteredMatches []*model.MatchResult
	for _, match := range allMatches {
		if containsAllMembers(match.Members, memberIDs) {
			filteredMatches = append(filteredMatches, match)
		}
	}

	return filteredMatches, nil
}

// UpdateMatchResult updates a match result
func (r *PoolRepository) UpdateMatchResult(ctx context.Context, matchID string, updates map[string]interface{}) (*model.MatchResult, error) {
	updates["updated_on"] = time.Now()

	// Build dynamic SET clause
	setClause := ""
	for key := range updates {
		if setClause != "" {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%s", key, key)
	}

	query := fmt.Sprintf("UPDATE type::record($id) SET %s", setClause)
	updates["id"] = matchID

	if err := r.db.Execute(ctx, query, updates); err != nil {
		return nil, fmt.Errorf("failed to update match: %w", err)
	}

	return r.GetMatchResult(ctx, matchID)
}

// GetPoolsDueForMatching retrieves pools that need matching run
func (r *PoolRepository) GetPoolsDueForMatching(ctx context.Context) ([]*model.MatchingPool, error) {
	query := `
		SELECT * FROM matching_pool
		WHERE active = true AND next_match_on <= time::now()
		ORDER BY next_match_on ASC
	`
	result, err := r.db.Query(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get pools due: %w", err)
	}

	return parsePoolsResult(result)
}

// GetUserPoolMemberships retrieves all pools a user is a member of
func (r *PoolRepository) GetUserPoolMemberships(ctx context.Context, userID string) ([]*model.PoolMember, error) {
	query := `
		SELECT * FROM pool_member
		WHERE user_id = $user_id AND active = true
		ORDER BY joined_on DESC
	`
	result, err := r.db.Query(ctx, query, map[string]interface{}{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to get memberships: %w", err)
	}

	return parsePoolMembersResult(result)
}

// CountPoolsByGuild returns the number of pools in a guild
func (r *PoolRepository) CountPoolsByGuild(ctx context.Context, guildID string) (int, error) {
	query := `SELECT count() AS count FROM matching_pool WHERE guild_id = $guild_id GROUP ALL`
	result, err := r.db.QueryOne(ctx, query, map[string]interface{}{"guild_id": guildID})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to count pools: %w", err)
	}

	return extractPoolCount(result), nil
}

// GetPoolStats retrieves statistics for a pool
func (r *PoolRepository) GetPoolStats(ctx context.Context, poolID string) (*model.PoolStats, error) {
	// Get total members
	totalQuery := `SELECT count() AS count FROM pool_member WHERE pool_id = $pool_id GROUP ALL`
	totalResult, _ := r.db.QueryOne(ctx, totalQuery, map[string]interface{}{"pool_id": poolID})
	total := extractPoolCount(totalResult)

	// Get active members
	activeQuery := `SELECT count() AS count FROM pool_member WHERE pool_id = $pool_id AND active = true GROUP ALL`
	activeResult, _ := r.db.QueryOne(ctx, activeQuery, map[string]interface{}{"pool_id": poolID})
	active := extractPoolCount(activeResult)

	// Get completed matches
	completedQuery := `SELECT count() AS count FROM match_result WHERE pool_id = $pool_id AND status = 'completed' GROUP ALL`
	completedResult, _ := r.db.QueryOne(ctx, completedQuery, map[string]interface{}{"pool_id": poolID})
	completed := extractPoolCount(completedResult)

	// Get skipped matches
	skippedQuery := `SELECT count() AS count FROM match_result WHERE pool_id = $pool_id AND status = 'skipped' GROUP ALL`
	skippedResult, _ := r.db.QueryOne(ctx, skippedQuery, map[string]interface{}{"pool_id": poolID})
	skipped := extractPoolCount(skippedResult)

	// Get total matches
	totalMatchesQuery := `SELECT count() AS count FROM match_result WHERE pool_id = $pool_id GROUP ALL`
	totalMatchesResult, _ := r.db.QueryOne(ctx, totalMatchesQuery, map[string]interface{}{"pool_id": poolID})
	totalMatches := extractPoolCount(totalMatchesResult)

	// Get distinct rounds
	roundsQuery := `SELECT count() AS count FROM (SELECT DISTINCT match_round FROM match_result WHERE pool_id = $pool_id) GROUP ALL`
	roundsResult, _ := r.db.QueryOne(ctx, roundsQuery, map[string]interface{}{"pool_id": poolID})
	rounds := extractPoolCount(roundsResult)

	completionRate := 0.0
	if totalMatches > 0 {
		completionRate = float64(completed) / float64(totalMatches) * 100
	}

	return &model.PoolStats{
		PoolID:           poolID,
		TotalMembers:     total,
		ActiveMembers:    active,
		TotalRounds:      rounds,
		CompletedMatches: completed,
		SkippedMatches:   skipped,
		CompletionRate:   completionRate,
	}, nil
}

// Helper functions

type poolCreatedRecord struct {
	ID        string
	CreatedOn time.Time
	UpdatedOn time.Time
}

func extractPoolCreatedRecord(result []interface{}) (*poolCreatedRecord, error) {
	if len(result) == 0 {
		return nil, errors.New("no result returned")
	}

	first := result[0]
	if resp, ok := first.(map[string]interface{}); ok {
		if resultData, ok := resp["result"].([]interface{}); ok && len(resultData) > 0 {
			first = resultData[0]
		}
	}

	data, ok := first.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	record := &poolCreatedRecord{}

	if id, ok := data["id"]; ok {
		record.ID = convertPoolID(id)
	}
	if createdOn, ok := data["created_on"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdOn); err == nil {
			record.CreatedOn = t
		}
	} else {
		record.CreatedOn = time.Now()
	}
	if updatedOn, ok := data["updated_on"].(string); ok {
		if t, err := time.Parse(time.RFC3339, updatedOn); err == nil {
			record.UpdatedOn = t
		}
	} else {
		record.UpdatedOn = time.Now()
	}
	if joinedOn, ok := data["joined_on"].(string); ok {
		if t, err := time.Parse(time.RFC3339, joinedOn); err == nil {
			record.CreatedOn = t
		}
	}

	return record, nil
}

func convertPoolID(id interface{}) string {
	if str, ok := id.(string); ok {
		return str
	}
	if rid, ok := id.(models.RecordID); ok {
		return fmt.Sprintf("%s:%v", rid.Table, rid.ID)
	}
	if rid, ok := id.(*models.RecordID); ok && rid != nil {
		return fmt.Sprintf("%s:%v", rid.Table, rid.ID)
	}
	return fmt.Sprintf("%v", id)
}

func parsePoolResult(result interface{}) (*model.MatchingPool, error) {
	if result == nil {
		return nil, nil
	}

	// Navigate through SurrealDB response structure
	if resp, ok := result.(map[string]interface{}); ok {
		if status, ok := resp["status"].(string); ok && status == "OK" {
			if resultData, ok := resp["result"].([]interface{}); ok {
				if len(resultData) == 0 {
					return nil, nil
				}
				result = resultData[0]
			}
		}
	}

	if arr, ok := result.([]interface{}); ok {
		if len(arr) == 0 {
			return nil, nil
		}
		result = arr[0]
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	// Handle SurrealDB's complex ID format
	if id, ok := data["id"]; ok {
		data["id"] = convertPoolID(id)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var pool model.MatchingPool
	if err := json.Unmarshal(jsonBytes, &pool); err != nil {
		return nil, err
	}

	return &pool, nil
}

func parsePoolsResult(results []interface{}) ([]*model.MatchingPool, error) {
	pools := make([]*model.MatchingPool, 0)

	for _, result := range results {
		if resp, ok := result.(map[string]interface{}); ok {
			if status, ok := resp["status"].(string); ok && status == "OK" {
				if resultData, ok := resp["result"].([]interface{}); ok {
					for _, item := range resultData {
						pool, err := parsePoolResult(item)
						if err == nil && pool != nil {
							pools = append(pools, pool)
						}
					}
				}
			}
		}
	}

	return pools, nil
}

func parsePoolMemberResult(result interface{}) (*model.PoolMember, error) {
	if result == nil {
		return nil, nil
	}

	// Navigate through SurrealDB response structure
	if resp, ok := result.(map[string]interface{}); ok {
		if status, ok := resp["status"].(string); ok && status == "OK" {
			if resultData, ok := resp["result"].([]interface{}); ok {
				if len(resultData) == 0 {
					return nil, nil
				}
				result = resultData[0]
			}
		}
	}

	if arr, ok := result.([]interface{}); ok {
		if len(arr) == 0 {
			return nil, nil
		}
		result = arr[0]
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	if id, ok := data["id"]; ok {
		data["id"] = convertPoolID(id)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var member model.PoolMember
	if err := json.Unmarshal(jsonBytes, &member); err != nil {
		return nil, err
	}

	return &member, nil
}

func parsePoolMembersResult(results []interface{}) ([]*model.PoolMember, error) {
	members := make([]*model.PoolMember, 0)

	for _, result := range results {
		if resp, ok := result.(map[string]interface{}); ok {
			if status, ok := resp["status"].(string); ok && status == "OK" {
				if resultData, ok := resp["result"].([]interface{}); ok {
					for _, item := range resultData {
						member, err := parsePoolMemberResult(item)
						if err == nil && member != nil {
							members = append(members, member)
						}
					}
				}
			}
		}
	}

	return members, nil
}

func parseMatchResult(result interface{}) (*model.MatchResult, error) {
	if result == nil {
		return nil, nil
	}

	// Navigate through SurrealDB response structure
	if resp, ok := result.(map[string]interface{}); ok {
		if status, ok := resp["status"].(string); ok && status == "OK" {
			if resultData, ok := resp["result"].([]interface{}); ok {
				if len(resultData) == 0 {
					return nil, nil
				}
				result = resultData[0]
			}
		}
	}

	if arr, ok := result.([]interface{}); ok {
		if len(arr) == 0 {
			return nil, nil
		}
		result = arr[0]
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	if id, ok := data["id"]; ok {
		data["id"] = convertPoolID(id)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var match model.MatchResult
	if err := json.Unmarshal(jsonBytes, &match); err != nil {
		return nil, err
	}

	return &match, nil
}

func parseMatchResultsFromQuery(results []interface{}) ([]*model.MatchResult, error) {
	matches := make([]*model.MatchResult, 0)

	for _, result := range results {
		if resp, ok := result.(map[string]interface{}); ok {
			if status, ok := resp["status"].(string); ok && status == "OK" {
				if resultData, ok := resp["result"].([]interface{}); ok {
					for _, item := range resultData {
						match, err := parseMatchResult(item)
						if err == nil && match != nil {
							matches = append(matches, match)
						}
					}
				}
			}
		}
	}

	return matches, nil
}

func extractPoolCount(result interface{}) int {
	if result == nil {
		return 0
	}

	if resp, ok := result.(map[string]interface{}); ok {
		if status, ok := resp["status"].(string); ok && status == "OK" {
			if resultData, ok := resp["result"].([]interface{}); ok && len(resultData) > 0 {
				if data, ok := resultData[0].(map[string]interface{}); ok {
					if count, ok := data["count"].(float64); ok {
						return int(count)
					}
				}
			}
		}
		// Direct access
		if count, ok := resp["count"].(float64); ok {
			return int(count)
		}
	}
	return 0
}

func containsAllMembers(members []string, targets []string) bool {
	memberSet := make(map[string]bool)
	for _, m := range members {
		memberSet[m] = true
	}
	for _, t := range targets {
		if !memberSet[t] {
			return false
		}
	}
	return true
}

// GetStaleMatches retrieves matches that are older than cutoff and have the specified status
func (r *PoolRepository) GetStaleMatches(ctx context.Context, cutoff time.Time, status string) ([]*model.MatchResult, error) {
	query := `
		SELECT * FROM match_result
		WHERE status = $status
		AND created_on < $cutoff
		ORDER BY created_on ASC
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

	return parseMatchResultsFromQuery(result)
}
