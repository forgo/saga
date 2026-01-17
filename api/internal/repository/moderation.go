package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// ModerationRepository handles moderation data access
type ModerationRepository struct {
	db database.Database
}

// NewModerationRepository creates a new moderation repository
func NewModerationRepository(db database.Database) *ModerationRepository {
	return &ModerationRepository{db: db}
}

// Report operations

// CreateReport creates a new report
func (r *ModerationRepository) CreateReport(ctx context.Context, report *model.Report) error {
	// Build query dynamically to avoid NULL vs NONE issues for optional fields
	setClause := `reporter_user_id = type::record($reporter_user_id), reported_user_id = type::record($reported_user_id), category = $category, status = $status, created_on = time::now()`
	vars := map[string]interface{}{
		"reporter_user_id": report.ReporterUserID,
		"reported_user_id": report.ReportedUserID,
		"category":         report.Category,
		"status":           report.Status,
	}

	// Only include optional fields if provided
	if report.CircleID != nil && *report.CircleID != "" {
		setClause += ", circle_id = $circle_id"
		vars["circle_id"] = *report.CircleID
	}
	if report.Description != nil && *report.Description != "" {
		setClause += ", description = $description"
		vars["description"] = *report.Description
	}
	if report.ContentType != nil && *report.ContentType != "" {
		setClause += ", content_type = $content_type"
		vars["content_type"] = *report.ContentType
	}
	if report.ContentID != nil && *report.ContentID != "" {
		setClause += ", content_id = $content_id"
		vars["content_id"] = *report.ContentID
	}

	query := "CREATE report SET " + setClause
	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return fmt.Errorf("failed to create report: %w", err)
	}

	created, err := r.extractReportFromResult(result)
	if err != nil {
		return fmt.Errorf("failed to extract report: %w", err)
	}

	report.ID = created.ID
	report.CreatedOn = created.CreatedOn
	return nil
}

// GetReport retrieves a report by ID
func (r *ModerationRepository) GetReport(ctx context.Context, id string) (*model.Report, error) {
	// Direct record access - more efficient than WHERE id =
	query := `SELECT * FROM type::record($id)`
	result, err := r.db.QueryOne(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	m, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}
	return r.parseReportFromMap(m)
}

// GetReportsByStatus retrieves reports by status
func (r *ModerationRepository) GetReportsByStatus(ctx context.Context, status model.ReportStatus, limit int) ([]*model.Report, error) {
	query := `
		SELECT * FROM report
		WHERE status = $status
		ORDER BY created_on DESC
		LIMIT $limit
	`
	result, err := r.db.Query(ctx, query, map[string]interface{}{
		"status": status,
		"limit":  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get reports: %w", err)
	}

	return r.parseReportsFromQuery(result)
}

// GetReportsAgainstUser retrieves reports against a specific user
func (r *ModerationRepository) GetReportsAgainstUser(ctx context.Context, userID string) ([]*model.Report, error) {
	query := `
		SELECT * FROM report
		WHERE reported_user_id = type::record($user_id)
		ORDER BY created_on DESC
	`
	result, err := r.db.Query(ctx, query, map[string]interface{}{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to get reports: %w", err)
	}

	return r.parseReportsFromQuery(result)
}

// GetRecentReportsAgainstUser retrieves recent reports against a user
func (r *ModerationRepository) GetRecentReportsAgainstUser(ctx context.Context, userID string, days int) ([]*model.Report, error) {
	query := `
		SELECT * FROM report
		WHERE reported_user_id = type::record($user_id)
		AND created_on > time::now() - type::duration($days + "d")
		ORDER BY created_on DESC
	`
	result, err := r.db.Query(ctx, query, map[string]interface{}{
		"user_id": userID,
		"days":    fmt.Sprintf("%d", days),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get recent reports: %w", err)
	}

	return r.parseReportsFromQuery(result)
}

// UpdateReport updates a report
func (r *ModerationRepository) UpdateReport(ctx context.Context, id string, updates map[string]interface{}) (*model.Report, error) {
	query := "UPDATE report SET "
	params := map[string]interface{}{"id": id}

	// Record fields that need special casting
	recordFields := map[string]bool{
		"reviewed_by_id": true,
	}

	first := true
	for key, value := range updates {
		if !first {
			query += ", "
		}
		if recordFields[key] {
			query += fmt.Sprintf("%s = type::record($%s)", key, key)
		} else {
			query += fmt.Sprintf("%s = $%s", key, key)
		}
		params[key] = value
		first = false
	}
	query += " WHERE id = type::record($id) RETURN AFTER"

	result, err := r.db.Query(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update report: %w", err)
	}

	return r.extractReportFromResult(result)
}

// Moderation Action operations

// CreateAction creates a moderation action
func (r *ModerationRepository) CreateAction(ctx context.Context, action *model.ModerationAction) error {
	query := `
		CREATE moderation_action CONTENT {
			user_id: $user_id,
			level: $level,
			reason: $reason,
			report_id: $report_id,
			admin_user_id: $admin_user_id,
			duration_days: $duration_days,
			expires_on: $expires_on,
			is_active: $is_active,
			restrictions: $restrictions,
			created_on: time::now()
		}
	`
	result, err := r.db.Query(ctx, query, map[string]interface{}{
		"user_id":       action.UserID,
		"level":         action.Level,
		"reason":        action.Reason,
		"report_id":     action.ReportID,
		"admin_user_id": action.AdminUserID,
		"duration_days": action.Duration,
		"expires_on":    action.ExpiresOn,
		"is_active":     action.IsActive,
		"restrictions":  action.Restrictions,
	})
	if err != nil {
		return fmt.Errorf("failed to create action: %w", err)
	}

	created, err := r.extractActionFromResult(result)
	if err != nil {
		return fmt.Errorf("failed to extract action: %w", err)
	}

	action.ID = created.ID
	action.CreatedOn = created.CreatedOn
	return nil
}

// GetAction retrieves an action by ID
func (r *ModerationRepository) GetAction(ctx context.Context, id string) (*model.ModerationAction, error) {
	// Direct record access - more efficient than WHERE id =
	query := `SELECT * FROM type::record($id)`
	result, err := r.db.QueryOne(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get action: %w", err)
	}

	m, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}
	return r.parseActionFromMap(m)
}

// GetActiveActionsForUser retrieves active moderation actions for a user
func (r *ModerationRepository) GetActiveActionsForUser(ctx context.Context, userID string) ([]*model.ModerationAction, error) {
	query := `
		SELECT * FROM moderation_action
		WHERE user_id = $user_id
		AND is_active = true
		AND (expires_on IS NULL OR expires_on > time::now())
		ORDER BY level DESC, created_on DESC
	`
	result, err := r.db.Query(ctx, query, map[string]interface{}{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to get active actions: %w", err)
	}

	return r.parseActionsFromQuery(result)
}

// GetAllActionsForUser retrieves all moderation actions for a user
func (r *ModerationRepository) GetAllActionsForUser(ctx context.Context, userID string) ([]*model.ModerationAction, error) {
	query := `
		SELECT * FROM moderation_action
		WHERE user_id = $user_id
		ORDER BY created_on DESC
	`
	result, err := r.db.Query(ctx, query, map[string]interface{}{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to get actions: %w", err)
	}

	return r.parseActionsFromQuery(result)
}

// UpdateAction updates a moderation action
func (r *ModerationRepository) UpdateAction(ctx context.Context, id string, updates map[string]interface{}) error {
	query := "UPDATE moderation_action SET "
	params := map[string]interface{}{"id": id}

	first := true
	for key, value := range updates {
		if !first {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%s", key, key)
		params[key] = value
		first = false
	}
	query += " WHERE id = type::record($id)"

	_, err := r.db.Query(ctx, query, params)
	return err
}

// ExpireOldActions marks expired actions as inactive
func (r *ModerationRepository) ExpireOldActions(ctx context.Context) error {
	query := `
		UPDATE moderation_action
		SET is_active = false
		WHERE is_active = true
		AND expires_on IS NOT NULL
		AND expires_on < time::now()
	`
	_, err := r.db.Query(ctx, query, nil)
	return err
}

// Block operations

// CreateBlock creates a block
func (r *ModerationRepository) CreateBlock(ctx context.Context, block *model.Block) error {
	// Build query dynamically to avoid NULL vs NONE issues for optional fields
	setClause := `blocker_user_id = type::record($blocker_user_id), blocked_user_id = type::record($blocked_user_id), created_on = time::now()`
	vars := map[string]interface{}{
		"blocker_user_id": block.BlockerUserID,
		"blocked_user_id": block.BlockedUserID,
	}

	// Only include reason if provided
	if block.Reason != nil {
		setClause += ", reason = $reason"
		vars["reason"] = *block.Reason
	}

	query := "CREATE block SET " + setClause
	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return fmt.Errorf("failed to create block: %w", err)
	}

	created, err := r.extractBlockFromResult(result)
	if err != nil {
		return fmt.Errorf("failed to extract block: %w", err)
	}

	block.ID = created.ID
	block.CreatedOn = created.CreatedOn
	return nil
}

// GetBlock retrieves a specific block
func (r *ModerationRepository) GetBlock(ctx context.Context, blockerID, blockedID string) (*model.Block, error) {
	query := `
		SELECT * FROM block
		WHERE blocker_user_id = type::record($blocker_id) AND blocked_user_id = type::record($blocked_id)
	`
	result, err := r.db.QueryOne(ctx, query, map[string]interface{}{
		"blocker_id": blockerID,
		"blocked_id": blockedID,
	})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	m, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}
	return r.parseBlockFromMap(m)
}

// GetBlocksByBlocker retrieves all blocks by a user
func (r *ModerationRepository) GetBlocksByBlocker(ctx context.Context, blockerID string) ([]*model.Block, error) {
	query := `
		SELECT * FROM block
		WHERE blocker_user_id = type::record($blocker_id)
		ORDER BY created_on DESC
	`
	result, err := r.db.Query(ctx, query, map[string]interface{}{"blocker_id": blockerID})
	if err != nil {
		return nil, fmt.Errorf("failed to get blocks: %w", err)
	}

	return r.parseBlocksFromQuery(result)
}

// IsBlocked checks if one user has blocked another
func (r *ModerationRepository) IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error) {
	block, err := r.GetBlock(ctx, blockerID, blockedID)
	if err != nil {
		return false, err
	}
	return block != nil, nil
}

// IsBlockedEitherWay checks if either user has blocked the other
func (r *ModerationRepository) IsBlockedEitherWay(ctx context.Context, userID1, userID2 string) (bool, error) {
	// Check if user1 blocked user2
	blocked1to2, err := r.GetBlock(ctx, userID1, userID2)
	if err != nil {
		return false, err
	}
	if blocked1to2 != nil {
		return true, nil
	}

	// Check if user2 blocked user1
	blocked2to1, err := r.GetBlock(ctx, userID2, userID1)
	if err != nil {
		return false, err
	}
	return blocked2to1 != nil, nil
}

// DeleteBlock removes a block
func (r *ModerationRepository) DeleteBlock(ctx context.Context, blockerID, blockedID string) error {
	query := `DELETE block WHERE blocker_user_id = type::record($blocker_id) AND blocked_user_id = type::record($blocked_id)`
	_, err := r.db.Query(ctx, query, map[string]interface{}{
		"blocker_id": blockerID,
		"blocked_id": blockedID,
	})
	return err
}

// Stats operations

// GetModerationStats retrieves moderation statistics
func (r *ModerationRepository) GetModerationStats(ctx context.Context) (*model.ModerationStats, error) {
	stats := &model.ModerationStats{}

	// Total reports
	totalQuery := `SELECT count() FROM report GROUP ALL`
	if result, err := r.db.QueryOne(ctx, totalQuery, nil); err == nil {
		if m, ok := result.(map[string]interface{}); ok {
			stats.TotalReports = getInt(m, "count")
		}
	}

	// Pending reports
	pendingQuery := `SELECT count() FROM report WHERE status = "pending" GROUP ALL`
	if result, err := r.db.QueryOne(ctx, pendingQuery, nil); err == nil {
		if m, ok := result.(map[string]interface{}); ok {
			stats.PendingReports = getInt(m, "count")
		}
	}

	// Resolved reports
	resolvedQuery := `SELECT count() FROM report WHERE status = "resolved" GROUP ALL`
	if result, err := r.db.QueryOne(ctx, resolvedQuery, nil); err == nil {
		if m, ok := result.(map[string]interface{}); ok {
			stats.ResolvedReports = getInt(m, "count")
		}
	}

	// Active warnings
	warningsQuery := `SELECT count() FROM moderation_action WHERE level = "warning" AND is_active = true GROUP ALL`
	if result, err := r.db.QueryOne(ctx, warningsQuery, nil); err == nil {
		if m, ok := result.(map[string]interface{}); ok {
			stats.ActiveWarnings = getInt(m, "count")
		}
	}

	// Active suspensions
	suspensionsQuery := `SELECT count() FROM moderation_action WHERE level = "suspension" AND is_active = true GROUP ALL`
	if result, err := r.db.QueryOne(ctx, suspensionsQuery, nil); err == nil {
		if m, ok := result.(map[string]interface{}); ok {
			stats.ActiveSuspensions = getInt(m, "count")
		}
	}

	// Total bans
	bansQuery := `SELECT count() FROM moderation_action WHERE level = "ban" AND is_active = true GROUP ALL`
	if result, err := r.db.QueryOne(ctx, bansQuery, nil); err == nil {
		if m, ok := result.(map[string]interface{}); ok {
			stats.TotalBans = getInt(m, "count")
		}
	}

	return stats, nil
}

// Parsing helpers

func (r *ModerationRepository) extractReportFromResult(result interface{}) (*model.Report, error) {
	rows, ok := extractQueryResults(result)
	if !ok || len(rows) == 0 {
		return nil, errors.New("no report returned")
	}
	m, ok := rows[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}
	return r.parseReportFromMap(m)
}

func (r *ModerationRepository) parseReportFromMap(m map[string]interface{}) (*model.Report, error) {
	rep := &model.Report{}

	if id, ok := m["id"]; ok {
		rep.ID = extractRecordID(id)
	}
	// reporter_user_id and reported_user_id come back as record objects
	if v, ok := m["reporter_user_id"]; ok {
		rep.ReporterUserID = convertSurrealID(v)
	}
	if v, ok := m["reported_user_id"]; ok {
		rep.ReportedUserID = convertSurrealID(v)
	}
	if v, ok := m["circle_id"].(string); ok {
		rep.CircleID = &v
	}
	if v, ok := m["category"].(string); ok {
		rep.Category = model.ReportCategory(v)
	}
	if v, ok := m["description"].(string); ok {
		rep.Description = &v
	}
	if v, ok := m["content_type"].(string); ok {
		rep.ContentType = &v
	}
	if v, ok := m["content_id"].(string); ok {
		rep.ContentID = &v
	}
	if v, ok := m["status"].(string); ok {
		rep.Status = model.ReportStatus(v)
	}
	if v, ok := m["reviewed_by_id"].(string); ok {
		rep.ReviewedByID = &v
	}
	if v, ok := m["review_notes"].(string); ok {
		rep.ReviewNotes = &v
	}
	if v, ok := m["action_taken"].(string); ok {
		rep.ActionTaken = &v
	}
	if v, ok := m["created_on"]; ok {
		rep.CreatedOn = parseTime(v)
	}
	if v, ok := m["reviewed_on"]; ok && v != nil {
		t := parseTime(v)
		if !t.IsZero() {
			rep.ReviewedOn = &t
		}
	}
	if v, ok := m["resolved_on"]; ok && v != nil {
		t := parseTime(v)
		if !t.IsZero() {
			rep.ResolvedOn = &t
		}
	}

	return rep, nil
}

func (r *ModerationRepository) parseReportsFromQuery(result interface{}) ([]*model.Report, error) {
	rows, ok := extractQueryResults(result)
	if !ok {
		return []*model.Report{}, nil
	}

	reports := make([]*model.Report, 0, len(rows))
	for _, row := range rows {
		if m, ok := row.(map[string]interface{}); ok {
			rep, err := r.parseReportFromMap(m)
			if err == nil {
				reports = append(reports, rep)
			}
		}
	}
	return reports, nil
}

func (r *ModerationRepository) extractActionFromResult(result interface{}) (*model.ModerationAction, error) {
	rows, ok := extractQueryResults(result)
	if !ok || len(rows) == 0 {
		return nil, errors.New("no action returned")
	}
	m, ok := rows[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}
	return r.parseActionFromMap(m)
}

func (r *ModerationRepository) parseActionFromMap(m map[string]interface{}) (*model.ModerationAction, error) {
	action := &model.ModerationAction{}

	if id, ok := m["id"]; ok {
		action.ID = extractRecordID(id)
	}
	if v, ok := m["user_id"].(string); ok {
		action.UserID = v
	}
	if v, ok := m["level"].(string); ok {
		action.Level = model.ModerationLevel(v)
	}
	if v, ok := m["reason"].(string); ok {
		action.Reason = v
	}
	if v, ok := m["report_id"].(string); ok {
		action.ReportID = &v
	}
	if v, ok := m["admin_user_id"].(string); ok {
		action.AdminUserID = &v
	}
	if v, ok := m["duration_days"].(float64); ok {
		dur := int(v)
		action.Duration = &dur
	}
	if v, ok := m["expires_on"]; ok && v != nil {
		t := parseTime(v)
		if !t.IsZero() {
			action.ExpiresOn = &t
		}
	}
	if v, ok := m["is_active"].(bool); ok {
		action.IsActive = v
	}
	if v, ok := m["restrictions"].([]interface{}); ok {
		action.Restrictions = make([]string, len(v))
		for i, r := range v {
			if s, ok := r.(string); ok {
				action.Restrictions[i] = s
			}
		}
	}
	if v, ok := m["created_on"]; ok {
		action.CreatedOn = parseTime(v)
	}
	if v, ok := m["lifted_on"]; ok && v != nil {
		t := parseTime(v)
		if !t.IsZero() {
			action.LiftedOn = &t
		}
	}
	if v, ok := m["lifted_by_id"].(string); ok {
		action.LiftedByID = &v
	}
	if v, ok := m["lift_reason"].(string); ok {
		action.LiftReason = &v
	}

	return action, nil
}

func (r *ModerationRepository) parseActionsFromQuery(result interface{}) ([]*model.ModerationAction, error) {
	rows, ok := extractQueryResults(result)
	if !ok {
		return []*model.ModerationAction{}, nil
	}

	actions := make([]*model.ModerationAction, 0, len(rows))
	for _, row := range rows {
		if m, ok := row.(map[string]interface{}); ok {
			action, err := r.parseActionFromMap(m)
			if err == nil {
				actions = append(actions, action)
			}
		}
	}
	return actions, nil
}

func (r *ModerationRepository) extractBlockFromResult(result interface{}) (*model.Block, error) {
	rows, ok := extractQueryResults(result)
	if !ok || len(rows) == 0 {
		return nil, errors.New("no block returned")
	}
	m, ok := rows[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}
	return r.parseBlockFromMap(m)
}

func (r *ModerationRepository) parseBlockFromMap(m map[string]interface{}) (*model.Block, error) {
	block := &model.Block{}

	if id, ok := m["id"]; ok {
		block.ID = extractRecordID(id)
	}
	// blocker_user_id and blocked_user_id come back as record objects
	if v, ok := m["blocker_user_id"]; ok {
		block.BlockerUserID = convertSurrealID(v)
	}
	if v, ok := m["blocked_user_id"]; ok {
		block.BlockedUserID = convertSurrealID(v)
	}
	if v, ok := m["reason"].(string); ok {
		block.Reason = &v
	}
	if v, ok := m["created_on"]; ok {
		block.CreatedOn = parseTime(v)
	}

	return block, nil
}

func (r *ModerationRepository) parseBlocksFromQuery(result interface{}) ([]*model.Block, error) {
	rows, ok := extractQueryResults(result)
	if !ok {
		return []*model.Block{}, nil
	}

	blocks := make([]*model.Block, 0, len(rows))
	for _, row := range rows {
		if m, ok := row.(map[string]interface{}); ok {
			block, err := r.parseBlockFromMap(m)
			if err == nil {
				blocks = append(blocks, block)
			}
		}
	}
	return blocks, nil
}
