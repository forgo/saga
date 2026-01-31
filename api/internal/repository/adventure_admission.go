package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// AdventureAdmissionRepository handles adventure admission data access
type AdventureAdmissionRepository struct {
	db database.Database
}

// NewAdventureAdmissionRepository creates a new adventure admission repository
func NewAdventureAdmissionRepository(db database.Database) *AdventureAdmissionRepository {
	return &AdventureAdmissionRepository{db: db}
}

// Create creates a new admission request
func (r *AdventureAdmissionRepository) Create(ctx context.Context, admission *model.AdventureAdmission) error {
	// Build query dynamically to avoid NULL values
	fields := []string{
		"adventure_id: type::record($adventure_id)",
		"user_id: type::record($user_id)",
		"status: $status",
		"requested_by: $requested_by",
		"requested_on: time::now()",
	}
	vars := map[string]interface{}{
		"adventure_id": admission.AdventureID,
		"user_id":      admission.UserID,
		"status":       admission.Status,
		"requested_by": admission.RequestedBy,
	}

	if admission.InvitedByID != nil {
		fields = append(fields, "invited_by_id: type::record($invited_by_id)")
		vars["invited_by_id"] = *admission.InvitedByID
	}

	query := fmt.Sprintf("CREATE adventure_admission CONTENT { %s }", strings.Join(fields, ", "))

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		if isUniqueConstraintError(err) {
			return fmt.Errorf("admission request already exists")
		}
		return fmt.Errorf("failed to create admission: %w", err)
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return fmt.Errorf("failed to extract created admission: %w", err)
	}

	admission.ID = created.ID
	admission.RequestedOn = created.CreatedOn
	return nil
}

// GetByID retrieves an admission by ID
func (r *AdventureAdmissionRepository) GetByID(ctx context.Context, id string) (*model.AdventureAdmission, error) {
	query := `SELECT * FROM type::record($id)`
	result, err := r.db.QueryOne(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get admission: %w", err)
	}

	return r.parseAdmission(result)
}

// GetByAdventureAndUser retrieves an admission for a specific user and adventure
func (r *AdventureAdmissionRepository) GetByAdventureAndUser(ctx context.Context, adventureID, userID string) (*model.AdventureAdmission, error) {
	query := `
		SELECT * FROM adventure_admission
		WHERE adventure_id = type::record($adventure_id)
		AND user_id = type::record($user_id)
	`
	vars := map[string]interface{}{
		"adventure_id": adventureID,
		"user_id":      userID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get admission: %w", err)
	}

	return r.parseAdmission(result)
}

// GetByAdventure retrieves all admissions for an adventure
func (r *AdventureAdmissionRepository) GetByAdventure(ctx context.Context, adventureID string, status *model.AdventureAdmissionStatus, limit, offset int) ([]*model.AdventureAdmission, error) {
	query := `
		SELECT * FROM adventure_admission
		WHERE adventure_id = type::record($adventure_id)
	`
	vars := map[string]interface{}{
		"adventure_id": adventureID,
		"limit":        limit,
		"offset":       offset,
	}

	if status != nil {
		query += ` AND status = $status`
		vars["status"] = *status
	}

	query += ` ORDER BY requested_on DESC LIMIT $limit START $offset`

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get admissions: %w", err)
	}

	return r.parseAdmissions(result)
}

// GetByUser retrieves all admissions for a user
func (r *AdventureAdmissionRepository) GetByUser(ctx context.Context, userID string, status *model.AdventureAdmissionStatus) ([]*model.AdventureAdmission, error) {
	query := `
		SELECT * FROM adventure_admission
		WHERE user_id = type::record($user_id)
	`
	vars := map[string]interface{}{"user_id": userID}

	if status != nil {
		query += ` AND status = $status`
		vars["status"] = *status
	}

	query += ` ORDER BY requested_on DESC`

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get user admissions: %w", err)
	}

	return r.parseAdmissions(result)
}

// GetAdmittedUsers retrieves all admitted users for an adventure
func (r *AdventureAdmissionRepository) GetAdmittedUsers(ctx context.Context, adventureID string) ([]*model.AdventureAdmission, error) {
	status := model.AdmissionStatusAdmitted
	return r.GetByAdventure(ctx, adventureID, &status, 500, 0)
}

// GetPendingRequests retrieves all pending admission requests for an adventure
func (r *AdventureAdmissionRepository) GetPendingRequests(ctx context.Context, adventureID string) ([]*model.AdventureAdmission, error) {
	status := model.AdmissionStatusRequested
	return r.GetByAdventure(ctx, adventureID, &status, 100, 0)
}

// Update updates an admission status
func (r *AdventureAdmissionRepository) Update(ctx context.Context, id string, status model.AdventureAdmissionStatus, rejectionReason *string) (*model.AdventureAdmission, error) {
	// Build query dynamically to avoid NULL
	setClause := "status = $status, decided_on = time::now()"
	vars := map[string]interface{}{
		"id":     id,
		"status": status,
	}

	if rejectionReason != nil {
		setClause += ", rejection_reason = $rejection_reason"
		vars["rejection_reason"] = *rejectionReason
	}

	query := fmt.Sprintf("UPDATE type::record($id) SET %s RETURN AFTER", setClause)

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to update admission: %w", err)
	}

	return r.parseAdmission(result)
}

// Admit admits a user to an adventure
func (r *AdventureAdmissionRepository) Admit(ctx context.Context, id string) (*model.AdventureAdmission, error) {
	return r.Update(ctx, id, model.AdmissionStatusAdmitted, nil)
}

// Reject rejects an admission request
func (r *AdventureAdmissionRepository) Reject(ctx context.Context, id string, reason string) (*model.AdventureAdmission, error) {
	return r.Update(ctx, id, model.AdmissionStatusRejected, &reason)
}

// Delete deletes an admission (withdraws request)
func (r *AdventureAdmissionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE type::record($id)`
	if err := r.db.Execute(ctx, query, map[string]interface{}{"id": id}); err != nil {
		return fmt.Errorf("failed to delete admission: %w", err)
	}
	return nil
}

// IsAdmitted checks if a user is admitted to an adventure
func (r *AdventureAdmissionRepository) IsAdmitted(ctx context.Context, adventureID, userID string) (bool, error) {
	// Use string::concat to compare record IDs as strings - SurrealDB 3 beta compatible
	query := `
		SELECT count() as count FROM adventure_admission
		WHERE string::concat("", adventure_id) = $adventure_id
		AND string::concat("", user_id) = $user_id
		AND status = "admitted"
		GROUP ALL
	`
	vars := map[string]interface{}{
		"adventure_id": adventureID,
		"user_id":      userID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	if data, ok := result.(map[string]interface{}); ok {
		return getInt(data, "count") > 0, nil
	}
	return false, nil
}

// CountAdmitted counts admitted users for an adventure
func (r *AdventureAdmissionRepository) CountAdmitted(ctx context.Context, adventureID string) (int, error) {
	query := `
		SELECT count() as count FROM adventure_admission
		WHERE adventure_id = type::record($adventure_id)
		AND status = "admitted"
		GROUP ALL
	`
	vars := map[string]interface{}{"adventure_id": adventureID}

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

// Parsing helpers

func (r *AdventureAdmissionRepository) parseAdmission(result interface{}) (*model.AdventureAdmission, error) {
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	admission := &model.AdventureAdmission{
		ID:          convertSurrealID(data["id"]),
		AdventureID: convertSurrealID(data["adventure_id"]),
		UserID:      convertSurrealID(data["user_id"]),
		Status:      model.AdventureAdmissionStatus(getString(data, "status")),
		RequestedBy: model.AdventureAdmissionRequestedBy(getString(data, "requested_by")),
	}

	if invitedBy := convertSurrealID(data["invited_by_id"]); invitedBy != "" {
		admission.InvitedByID = &invitedBy
	}
	if reason := getString(data, "rejection_reason"); reason != "" {
		admission.RejectionReason = &reason
	}
	if t := getTime(data, "requested_on"); t != nil {
		admission.RequestedOn = *t
	}
	if t := getTime(data, "decided_on"); t != nil {
		admission.DecidedOn = t
	}

	return admission, nil
}

func (r *AdventureAdmissionRepository) parseAdmissions(result []interface{}) ([]*model.AdventureAdmission, error) {
	admissions := make([]*model.AdventureAdmission, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					admission, err := r.parseAdmission(item)
					if err != nil {
						continue
					}
					admissions = append(admissions, admission)
				}
			}
		}
	}

	return admissions, nil
}
