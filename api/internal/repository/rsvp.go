package repository

import (
	"context"
	"errors"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// RSVPRepository handles unified RSVP data access
type RSVPRepository struct {
	db database.Database
}

// NewRSVPRepository creates a new RSVP repository
func NewRSVPRepository(db database.Database) *RSVPRepository {
	return &RSVPRepository{db: db}
}

// Create creates a new unified RSVP
func (r *RSVPRepository) Create(ctx context.Context, rsvp *model.UnifiedRSVP) error {
	query := `
		CREATE unified_rsvp CONTENT {
			target_type: $target_type,
			target_id: $target_id,
			user_id: type::record($user_id),
			status: $status,
			role: $role,
			values_aligned: $values_aligned,
			alignment_score: $alignment_score,
			yikes_count: $yikes_count,
			plus_ones: $plus_ones,
			plus_one_names: $plus_one_names,
			note: $note,
			created_on: time::now(),
			updated_on: time::now()
		}
	`

	role := rsvp.Role
	if role == "" {
		role = model.RSVPRoleParticipant
	}

	status := rsvp.Status
	if status == "" {
		status = model.UnifiedRSVPStatusPending
	}

	vars := map[string]interface{}{
		"target_type":     rsvp.TargetType,
		"target_id":       rsvp.TargetID,
		"user_id":         rsvp.UserID,
		"status":          status,
		"role":            role,
		"values_aligned":  rsvp.ValuesAligned,
		"alignment_score": rsvp.AlignmentScore,
		"yikes_count":     rsvp.YikesCount,
		"plus_ones":       rsvp.PlusOnes,
		"plus_one_names":  rsvp.PlusOneNames,
		"note":            rsvp.Note,
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
	rsvp.CreatedOn = created.CreatedOn
	rsvp.UpdatedOn = created.UpdatedOn
	rsvp.Status = status
	rsvp.Role = role
	return nil
}

// Get retrieves an RSVP by ID
func (r *RSVPRepository) Get(ctx context.Context, rsvpID string) (*model.UnifiedRSVP, error) {
	// Direct record access - more efficient than WHERE id =
	query := `SELECT * FROM type::record($rsvp_id)`
	vars := map[string]interface{}{"rsvp_id": rsvpID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.parseRSVPResult(result)
}

// GetByTargetAndUser retrieves an RSVP by target and user
func (r *RSVPRepository) GetByTargetAndUser(ctx context.Context, targetType, targetID, userID string) (*model.UnifiedRSVP, error) {
	query := `
		SELECT * FROM unified_rsvp
		WHERE target_type = $target_type
		AND target_id = $target_id
		AND user_id = type::record($user_id)
		LIMIT 1
	`
	vars := map[string]interface{}{
		"target_type": targetType,
		"target_id":   targetID,
		"user_id":     userID,
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

// Update updates an RSVP
func (r *RSVPRepository) Update(ctx context.Context, rsvpID string, updates map[string]interface{}) (*model.UnifiedRSVP, error) {
	query := `UPDATE unified_rsvp SET updated_on = time::now()`
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

// Delete deletes an RSVP
func (r *RSVPRepository) Delete(ctx context.Context, rsvpID string) error {
	query := `DELETE unified_rsvp WHERE id = type::record($rsvp_id)`
	vars := map[string]interface{}{"rsvp_id": rsvpID}

	return r.db.Execute(ctx, query, vars)
}

// GetByTarget retrieves all RSVPs for a target
func (r *RSVPRepository) GetByTarget(ctx context.Context, targetType, targetID string, filters *model.RSVPFilters) ([]*model.UnifiedRSVP, error) {
	query := `
		SELECT * FROM unified_rsvp
		WHERE target_type = $target_type
		AND target_id = $target_id
	`
	vars := map[string]interface{}{
		"target_type": targetType,
		"target_id":   targetID,
	}

	if filters != nil && filters.Status != nil {
		query += ` AND status = $status`
		vars["status"] = *filters.Status
	}

	if filters != nil && filters.Role != nil {
		query += ` AND role = $role`
		vars["role"] = *filters.Role
	}

	query += ` ORDER BY created_on ASC`

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseRSVPsResult(result)
}

// GetByUser retrieves all RSVPs for a user
func (r *RSVPRepository) GetByUser(ctx context.Context, userID string, filters *model.RSVPFilters) ([]*model.UnifiedRSVP, error) {
	query := `
		SELECT * FROM unified_rsvp
		WHERE user_id = type::record($user_id)
	`
	vars := map[string]interface{}{"user_id": userID}

	if filters != nil && filters.TargetType != nil {
		query += ` AND target_type = $target_type`
		vars["target_type"] = *filters.TargetType
	}

	if filters != nil && filters.Status != nil {
		query += ` AND status = $status`
		vars["status"] = *filters.Status
	}

	query += ` ORDER BY created_on DESC`

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseRSVPsResult(result)
}

// GetPendingByTarget retrieves pending RSVPs for a target
func (r *RSVPRepository) GetPendingByTarget(ctx context.Context, targetType, targetID string) ([]*model.UnifiedRSVP, error) {
	status := model.UnifiedRSVPStatusPending
	return r.GetByTarget(ctx, targetType, targetID, &model.RSVPFilters{Status: &status})
}

// GetApprovedByTarget retrieves approved RSVPs for a target
func (r *RSVPRepository) GetApprovedByTarget(ctx context.Context, targetType, targetID string) ([]*model.UnifiedRSVP, error) {
	status := model.UnifiedRSVPStatusApproved
	return r.GetByTarget(ctx, targetType, targetID, &model.RSVPFilters{Status: &status})
}

// CountByTargetAndStatus counts RSVPs by target and status
func (r *RSVPRepository) CountByTargetAndStatus(ctx context.Context, targetType, targetID, status string) (int, error) {
	query := `
		SELECT count() as count FROM unified_rsvp
		WHERE target_type = $target_type
		AND target_id = $target_id
		AND status = $status
		GROUP ALL
	`
	vars := map[string]interface{}{
		"target_type": targetType,
		"target_id":   targetID,
		"status":      status,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	if data, ok := result.(map[string]interface{}); ok {
		return getInt(data, "count"), nil
	}
	return 0, nil
}

// CountApprovedWithPlusOnes counts approved RSVPs including plus ones
func (r *RSVPRepository) CountApprovedWithPlusOnes(ctx context.Context, targetType, targetID string) (int, error) {
	query := `
		SELECT math::sum(1 + (plus_ones OR 0)) as total FROM unified_rsvp
		WHERE target_type = $target_type
		AND target_id = $target_id
		AND status IN ["approved", "attended"]
		GROUP ALL
	`
	vars := map[string]interface{}{
		"target_type": targetType,
		"target_id":   targetID,
	}

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

// GetStats retrieves aggregate stats for a target
func (r *RSVPRepository) GetStats(ctx context.Context, targetType, targetID string) (*model.RSVPStats, error) {
	query := `
		SELECT
			status,
			count() as count,
			math::sum(plus_ones OR 0) as plus_ones
		FROM unified_rsvp
		WHERE target_type = $target_type
		AND target_id = $target_id
		GROUP BY status
	`
	vars := map[string]interface{}{
		"target_type": targetType,
		"target_id":   targetID,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	stats := &model.RSVPStats{
		TargetType: targetType,
		TargetID:   targetID,
	}

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					if data, ok := item.(map[string]interface{}); ok {
						status := getString(data, "status")
						count := getInt(data, "count")
						plusOnes := getInt(data, "plus_ones")

						switch status {
						case model.UnifiedRSVPStatusApproved:
							stats.ApprovedCount = count
							stats.TotalPlusOnes += plusOnes
						case model.UnifiedRSVPStatusPending:
							stats.PendingCount = count
						case model.UnifiedRSVPStatusWaitlisted:
							stats.WaitlistCount = count
						case model.UnifiedRSVPStatusAttended:
							stats.AttendedCount = count
							stats.TotalPlusOnes += plusOnes
						}
					}
				}
			}
		}
	}

	return stats, nil
}

// ConfirmCompletion marks an RSVP as completion confirmed
func (r *RSVPRepository) ConfirmCompletion(ctx context.Context, rsvpID string, early bool) error {
	updates := map[string]interface{}{
		"completion_confirmed": time.Now(),
		"early_confirmed":      early,
	}

	_, err := r.Update(ctx, rsvpID, updates)
	return err
}

// RecordCheckin records a checkin time
func (r *RSVPRepository) RecordCheckin(ctx context.Context, rsvpID string) error {
	updates := map[string]interface{}{
		"checkin_time": time.Now(),
	}

	_, err := r.Update(ctx, rsvpID, updates)
	return err
}

// RecordFeedback records helpfulness feedback
func (r *RSVPRepository) RecordFeedback(ctx context.Context, rsvpID string, rating string, tags []string) error {
	updates := map[string]interface{}{
		"helpfulness_rating": rating,
		"helpfulness_tags":   tags,
	}

	_, err := r.Update(ctx, rsvpID, updates)
	return err
}

// GetUnconfirmedByTarget retrieves RSVPs that haven't confirmed completion
func (r *RSVPRepository) GetUnconfirmedByTarget(ctx context.Context, targetType, targetID string) ([]*model.UnifiedRSVP, error) {
	query := `
		SELECT * FROM unified_rsvp
		WHERE target_type = $target_type
		AND target_id = $target_id
		AND status IN ["approved", "attended"]
		AND completion_confirmed IS NONE
		ORDER BY created_on ASC
	`
	vars := map[string]interface{}{
		"target_type": targetType,
		"target_id":   targetID,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseRSVPsResult(result)
}

// Helper functions

func (r *RSVPRepository) parseRSVPResult(result interface{}) (*model.UnifiedRSVP, error) {
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

func (r *RSVPRepository) parseRSVPsResult(result []interface{}) ([]*model.UnifiedRSVP, error) {
	rsvps := make([]*model.UnifiedRSVP, 0)

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
