package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// EventRoleRepository handles event role data access
type EventRoleRepository struct {
	db database.Database
}

// NewEventRoleRepository creates a new event role repository
func NewEventRoleRepository(db database.Database) *EventRoleRepository {
	return &EventRoleRepository{db: db}
}

// CreateRole creates a new event role
func (r *EventRoleRepository) CreateRole(ctx context.Context, role *model.EventRole) error {
	// Build query dynamically to avoid NULL vs NONE issues for optional fields
	setClause := `event_id = type::record($event_id), name = $name, max_slots = $max_slots, is_default = $is_default, sort_order = $sort_order, created_by = type::record($created_by), created_on = time::now(), updated_on = time::now()`
	vars := map[string]interface{}{
		"event_id":   role.EventID,
		"name":       role.Name,
		"max_slots":  role.MaxSlots,
		"is_default": role.IsDefault,
		"sort_order": role.SortOrder,
		"created_by": role.CreatedBy,
	}

	// Add optional fields only when they have values
	if role.Description != nil && *role.Description != "" {
		setClause += ", description = $description"
		vars["description"] = *role.Description
	}
	if len(role.SuggestedInterests) > 0 {
		setClause += ", suggested_interests = $suggested_interests"
		vars["suggested_interests"] = role.SuggestedInterests
	}

	query := "CREATE event_role SET " + setClause

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	role.ID = created.ID
	role.CreatedOn = created.CreatedOn
	role.UpdatedOn = created.UpdatedOn
	return nil
}

// GetRole retrieves a role by ID
func (r *EventRoleRepository) GetRole(ctx context.Context, roleID string) (*model.EventRole, error) {
	// Direct record access - more efficient than WHERE id =
	query := `SELECT * FROM type::record($role_id)`
	vars := map[string]interface{}{"role_id": roleID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.parseRoleResult(result)
}

// GetRolesByEvent retrieves all roles for an event
func (r *EventRoleRepository) GetRolesByEvent(ctx context.Context, eventID string) ([]*model.EventRole, error) {
	query := `
		SELECT * FROM event_role
		WHERE event_id = type::record($event_id)
		ORDER BY sort_order ASC
	`
	vars := map[string]interface{}{"event_id": eventID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseRolesResult(result)
}

// UpdateRole updates a role
func (r *EventRoleRepository) UpdateRole(ctx context.Context, roleID string, updates map[string]interface{}) (*model.EventRole, error) {
	query := `UPDATE event_role SET updated_on = time::now()`

	vars := map[string]interface{}{
		"role_id": roleID,
	}

	if name, ok := updates["name"]; ok {
		query += ", name = $name"
		vars["name"] = name
	}
	if description, ok := updates["description"]; ok {
		query += ", description = $description"
		vars["description"] = description
	}
	if maxSlots, ok := updates["max_slots"]; ok {
		query += ", max_slots = $max_slots"
		vars["max_slots"] = maxSlots
	}
	if suggestedInterests, ok := updates["suggested_interests"]; ok {
		query += ", suggested_interests = $suggested_interests"
		vars["suggested_interests"] = suggestedInterests
	}

	query += ` WHERE id = type::record($role_id) RETURN AFTER`

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseRoleResult(result)
}

// DeleteRole deletes a role
func (r *EventRoleRepository) DeleteRole(ctx context.Context, roleID string) error {
	query := `DELETE event_role WHERE id = type::record($role_id)`
	vars := map[string]interface{}{"role_id": roleID}

	return r.db.Execute(ctx, query, vars)
}

// CreateAssignment creates a role assignment
func (r *EventRoleRepository) CreateAssignment(ctx context.Context, assignment *model.EventRoleAssignment) error {
	// Build query dynamically to avoid NULL vs NONE issues for optional fields
	setClause := `event_id = type::record($event_id), role_id = type::record($role_id), user_id = type::record($user_id), status = $status, assigned_on = time::now(), updated_on = time::now()`
	vars := map[string]interface{}{
		"event_id": assignment.EventID,
		"role_id":  assignment.RoleID,
		"user_id":  assignment.UserID,
		"status":   assignment.Status,
	}

	// Add optional fields only when they have values
	if assignment.Note != nil && *assignment.Note != "" {
		setClause += ", note = $note"
		vars["note"] = *assignment.Note
	}

	query := "CREATE event_role_assignment SET " + setClause

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	// Extract created record - assignments use assigned_on instead of created_on
	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	assignment.ID = created.ID
	// For assignments, assigned_on is set to time::now() in the query, same as updated_on
	assignment.AssignedOn = created.UpdatedOn // Updated_on is set at creation time
	assignment.UpdatedOn = created.UpdatedOn
	return nil
}

// GetAssignment retrieves an assignment by ID
func (r *EventRoleRepository) GetAssignment(ctx context.Context, assignmentID string) (*model.EventRoleAssignment, error) {
	// Direct record access - more efficient than WHERE id =
	query := `SELECT * FROM type::record($assignment_id)`
	vars := map[string]interface{}{"assignment_id": assignmentID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.parseAssignmentResult(result)
}

// GetUserAssignmentForRole retrieves a user's assignment for a specific role
func (r *EventRoleRepository) GetUserAssignmentForRole(ctx context.Context, roleID, userID string) (*model.EventRoleAssignment, error) {
	query := `
		SELECT * FROM event_role_assignment
		WHERE role_id = type::record($role_id) AND user_id = type::record($user_id)
		LIMIT 1
	`
	vars := map[string]interface{}{
		"role_id": roleID,
		"user_id": userID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.parseAssignmentResult(result)
}

// GetUserAssignmentsForEvent retrieves all of a user's assignments for an event
// A user can have multiple role assignments per event
func (r *EventRoleRepository) GetUserAssignmentsForEvent(ctx context.Context, eventID, userID string) ([]*model.EventRoleAssignment, error) {
	query := `
		SELECT * FROM event_role_assignment
		WHERE event_id = type::record($event_id) AND user_id = type::record($user_id)
		ORDER BY assigned_on ASC
	`
	vars := map[string]interface{}{
		"event_id": eventID,
		"user_id":  userID,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseAssignmentsResult(result)
}

// GetAssignmentsByRole retrieves all assignments for a role
func (r *EventRoleRepository) GetAssignmentsByRole(ctx context.Context, roleID string) ([]*model.EventRoleAssignment, error) {
	query := `
		SELECT * FROM event_role_assignment
		WHERE role_id = type::record($role_id) AND status != "cancelled"
		ORDER BY assigned_on ASC
	`
	vars := map[string]interface{}{"role_id": roleID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseAssignmentsResult(result)
}

// GetAssignmentsByEvent retrieves all assignments for an event
func (r *EventRoleRepository) GetAssignmentsByEvent(ctx context.Context, eventID string) ([]*model.EventRoleAssignment, error) {
	query := `
		SELECT * FROM event_role_assignment
		WHERE event_id = type::record($event_id) AND status != "cancelled"
		ORDER BY assigned_on ASC
	`
	vars := map[string]interface{}{"event_id": eventID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseAssignmentsResult(result)
}

// UpdateAssignment updates an assignment
func (r *EventRoleRepository) UpdateAssignment(ctx context.Context, assignmentID string, updates map[string]interface{}) (*model.EventRoleAssignment, error) {
	query := `UPDATE event_role_assignment SET updated_on = time::now()`

	vars := map[string]interface{}{
		"assignment_id": assignmentID,
	}

	if note, ok := updates["note"]; ok {
		query += ", note = $note"
		vars["note"] = note
	}
	if status, ok := updates["status"]; ok {
		query += ", status = $status"
		vars["status"] = status
	}

	query += ` WHERE id = type::record($assignment_id) RETURN AFTER`

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseAssignmentResult(result)
}

// DeleteAssignment deletes an assignment
func (r *EventRoleRepository) DeleteAssignment(ctx context.Context, assignmentID string) error {
	query := `DELETE event_role_assignment WHERE id = type::record($assignment_id)`
	vars := map[string]interface{}{"assignment_id": assignmentID}

	return r.db.Execute(ctx, query, vars)
}

// CountAssignmentsByRole counts confirmed assignments for a role
func (r *EventRoleRepository) CountAssignmentsByRole(ctx context.Context, roleID string) (int, error) {
	query := `
		SELECT count() as count FROM event_role_assignment
		WHERE role_id = $role_id AND status = "confirmed"
		GROUP ALL
	`
	vars := map[string]interface{}{"role_id": roleID}

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

// GetRolesWithAssignments retrieves roles with their assignments
func (r *EventRoleRepository) GetRolesWithAssignments(ctx context.Context, eventID string) ([]model.EventRoleWithAssignments, error) {
	roles, err := r.GetRolesByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	result := make([]model.EventRoleWithAssignments, 0, len(roles))
	for _, role := range roles {
		assignments, err := r.GetAssignmentsByRole(ctx, role.ID)
		if err != nil {
			return nil, err
		}

		// Convert to non-pointer slice
		assignmentsList := make([]model.EventRoleAssignment, 0, len(assignments))
		for _, a := range assignments {
			assignmentsList = append(assignmentsList, *a)
		}

		spotsLeft := -1 // Unlimited
		if role.MaxSlots > 0 {
			spotsLeft = role.MaxSlots - len(assignmentsList)
			if spotsLeft < 0 {
				spotsLeft = 0
			}
		}

		result = append(result, model.EventRoleWithAssignments{
			Role:        *role,
			Assignments: assignmentsList,
			IsFull:      role.MaxSlots > 0 && len(assignmentsList) >= role.MaxSlots,
			SpotsLeft:   spotsLeft,
		})
	}

	return result, nil
}

// Helper functions

func (r *EventRoleRepository) parseRoleResult(result interface{}) (*model.EventRole, error) {
	if result == nil {
		return nil, database.ErrNotFound
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	// Convert record IDs to strings before JSON marshaling
	if id, ok := data["id"]; ok {
		data["id"] = convertSurrealID(id)
	}
	if eventID, ok := data["event_id"]; ok {
		data["event_id"] = convertSurrealID(eventID)
	}
	if createdBy, ok := data["created_by"]; ok {
		data["created_by"] = convertSurrealID(createdBy)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var role model.EventRole
	if err := json.Unmarshal(jsonBytes, &role); err != nil {
		return nil, err
	}

	role.SuggestedInterests = getStringSlice(data, "suggested_interests")

	if t := getTime(data, "created_on"); t != nil {
		role.CreatedOn = *t
	}
	if t := getTime(data, "updated_on"); t != nil {
		role.UpdatedOn = *t
	}

	return &role, nil
}

func (r *EventRoleRepository) parseRolesResult(result []interface{}) ([]*model.EventRole, error) {
	roles := make([]*model.EventRole, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					role, err := r.parseRoleResult(item)
					if err != nil {
						continue
					}
					roles = append(roles, role)
				}
				continue
			}
		}

		role, err := r.parseRoleResult(res)
		if err != nil {
			continue
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *EventRoleRepository) parseAssignmentResult(result interface{}) (*model.EventRoleAssignment, error) {
	if result == nil {
		return nil, database.ErrNotFound
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	// Convert record IDs to strings before JSON marshaling
	if id, ok := data["id"]; ok {
		data["id"] = convertSurrealID(id)
	}
	if eventID, ok := data["event_id"]; ok {
		data["event_id"] = convertSurrealID(eventID)
	}
	if roleID, ok := data["role_id"]; ok {
		data["role_id"] = convertSurrealID(roleID)
	}
	if userID, ok := data["user_id"]; ok {
		data["user_id"] = convertSurrealID(userID)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var assignment model.EventRoleAssignment
	if err := json.Unmarshal(jsonBytes, &assignment); err != nil {
		return nil, err
	}

	if t := getTime(data, "assigned_on"); t != nil {
		assignment.AssignedOn = *t
	}
	if t := getTime(data, "updated_on"); t != nil {
		assignment.UpdatedOn = *t
	}

	return &assignment, nil
}

func (r *EventRoleRepository) parseAssignmentsResult(result []interface{}) ([]*model.EventRoleAssignment, error) {
	assignments := make([]*model.EventRoleAssignment, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					assignment, err := r.parseAssignmentResult(item)
					if err != nil {
						continue
					}
					assignments = append(assignments, assignment)
				}
				continue
			}
		}

		assignment, err := r.parseAssignmentResult(res)
		if err != nil {
			continue
		}
		assignments = append(assignments, assignment)
	}

	return assignments, nil
}

// Unused - silence linter
var _ = time.Now
