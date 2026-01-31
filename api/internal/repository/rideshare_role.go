package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// RideshareRoleRepository handles rideshare role data access
type RideshareRoleRepository struct {
	db database.Database
}

// NewRideshareRoleRepository creates a new rideshare role repository
func NewRideshareRoleRepository(db database.Database) *RideshareRoleRepository {
	return &RideshareRoleRepository{db: db}
}

// Create creates a new rideshare role
func (r *RideshareRoleRepository) Create(ctx context.Context, role *model.RideshareRole) error {
	query := `
		CREATE rideshare_role CONTENT {
			rideshare_id: type::record($rideshare_id),
			catalog_role_id: $catalog_role_id,
			name: $name,
			description: $description,
			max_slots: $max_slots,
			filled_slots: 0,
			sort_order: $sort_order,
			created_by: type::record($created_by),
			created_on: time::now(),
			updated_on: time::now()
		}
	`
	vars := map[string]interface{}{
		"rideshare_id":    role.RideshareID,
		"catalog_role_id": role.CatalogRoleID,
		"name":            role.Name,
		"description":     role.Description,
		"max_slots":       role.MaxSlots,
		"sort_order":      role.SortOrder,
		"created_by":      role.CreatedBy,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return fmt.Errorf("failed to create rideshare role: %w", err)
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return fmt.Errorf("failed to extract created role: %w", err)
	}

	role.ID = created.ID
	role.CreatedOn = created.CreatedOn
	role.UpdatedOn = created.UpdatedOn
	role.FilledSlots = 0
	return nil
}

// GetByID retrieves a rideshare role by ID
func (r *RideshareRoleRepository) GetByID(ctx context.Context, id string) (*model.RideshareRole, error) {
	query := `SELECT * FROM type::record($id)`
	result, err := r.db.QueryOne(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get rideshare role: %w", err)
	}

	return r.parseRideshareRole(result)
}

// GetByRideshare retrieves all roles for a rideshare
func (r *RideshareRoleRepository) GetByRideshare(ctx context.Context, rideshareID string) ([]*model.RideshareRole, error) {
	query := `
		SELECT * FROM rideshare_role
		WHERE rideshare_id = type::record($rideshare_id)
		ORDER BY sort_order ASC, name ASC
	`
	vars := map[string]interface{}{"rideshare_id": rideshareID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get rideshare roles: %w", err)
	}

	return r.parseRideshareRoles(result)
}

// Update updates a rideshare role
func (r *RideshareRoleRepository) Update(ctx context.Context, id string, updates *model.UpdateRideshareRoleRequest) (*model.RideshareRole, error) {
	query := `UPDATE type::record($id) SET updated_on = time::now()`
	vars := map[string]interface{}{"id": id}

	if updates.Name != nil {
		query += `, name = $name`
		vars["name"] = *updates.Name
	}
	if updates.Description != nil {
		query += `, description = $description`
		vars["description"] = *updates.Description
	}
	if updates.MaxSlots != nil {
		query += `, max_slots = $max_slots`
		vars["max_slots"] = *updates.MaxSlots
	}

	query += ` RETURN AFTER`

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to update rideshare role: %w", err)
	}

	return r.parseRideshareRole(result)
}

// Delete deletes a rideshare role
func (r *RideshareRoleRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE type::record($id)`
	if err := r.db.Execute(ctx, query, map[string]interface{}{"id": id}); err != nil {
		return fmt.Errorf("failed to delete rideshare role: %w", err)
	}
	return nil
}

// Assignment operations

// CreateAssignment creates a role assignment
func (r *RideshareRoleRepository) CreateAssignment(ctx context.Context, assignment *model.RideshareRoleAssignment) error {
	query := `
		CREATE rideshare_role_assignment CONTENT {
			rideshare_id: type::record($rideshare_id),
			role_id: type::record($role_id),
			user_id: type::record($user_id),
			note: $note,
			status: $status,
			assigned_on: time::now(),
			updated_on: time::now()
		}
	`
	vars := map[string]interface{}{
		"rideshare_id": assignment.RideshareID,
		"role_id":      assignment.RoleID,
		"user_id":      assignment.UserID,
		"note":         assignment.Note,
		"status":       assignment.Status,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		if isUniqueConstraintError(err) {
			return fmt.Errorf("user already assigned to this role")
		}
		return fmt.Errorf("failed to create assignment: %w", err)
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return fmt.Errorf("failed to extract created assignment: %w", err)
	}

	assignment.ID = created.ID
	assignment.AssignedOn = created.CreatedOn
	assignment.UpdatedOn = created.UpdatedOn
	return nil
}

// GetAssignmentByID retrieves an assignment by ID
func (r *RideshareRoleRepository) GetAssignmentByID(ctx context.Context, id string) (*model.RideshareRoleAssignment, error) {
	query := `SELECT * FROM type::record($id)`
	result, err := r.db.QueryOne(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get assignment: %w", err)
	}

	return r.parseRideshareRoleAssignment(result)
}

// GetAssignmentsByRole retrieves all assignments for a role
func (r *RideshareRoleRepository) GetAssignmentsByRole(ctx context.Context, roleID string) ([]*model.RideshareRoleAssignment, error) {
	query := `
		SELECT * FROM rideshare_role_assignment
		WHERE role_id = type::record($role_id)
		ORDER BY assigned_on ASC
	`
	vars := map[string]interface{}{"role_id": roleID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignments: %w", err)
	}

	return r.parseRideshareRoleAssignments(result)
}

// GetAssignmentsByRideshare retrieves all assignments for a rideshare
func (r *RideshareRoleRepository) GetAssignmentsByRideshare(ctx context.Context, rideshareID string) ([]*model.RideshareRoleAssignment, error) {
	query := `
		SELECT * FROM rideshare_role_assignment
		WHERE rideshare_id = type::record($rideshare_id)
		ORDER BY assigned_on ASC
	`
	vars := map[string]interface{}{"rideshare_id": rideshareID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignments: %w", err)
	}

	return r.parseRideshareRoleAssignments(result)
}

// GetAssignmentsByUser retrieves all assignments for a user in a rideshare
func (r *RideshareRoleRepository) GetAssignmentsByUser(ctx context.Context, rideshareID, userID string) ([]*model.RideshareRoleAssignment, error) {
	query := `
		SELECT * FROM rideshare_role_assignment
		WHERE rideshare_id = type::record($rideshare_id)
		AND user_id = type::record($user_id)
	`
	vars := map[string]interface{}{
		"rideshare_id": rideshareID,
		"user_id":      userID,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get user assignments: %w", err)
	}

	return r.parseRideshareRoleAssignments(result)
}

// DeleteAssignment deletes an assignment
func (r *RideshareRoleRepository) DeleteAssignment(ctx context.Context, id string) error {
	query := `DELETE type::record($id)`
	if err := r.db.Execute(ctx, query, map[string]interface{}{"id": id}); err != nil {
		return fmt.Errorf("failed to delete assignment: %w", err)
	}
	return nil
}

// GetRolesWithAssignments retrieves all roles with their assignments for a rideshare
func (r *RideshareRoleRepository) GetRolesWithAssignments(ctx context.Context, rideshareID string) ([]model.RideshareRoleWithAssignments, error) {
	roles, err := r.GetByRideshare(ctx, rideshareID)
	if err != nil {
		return nil, err
	}

	result := make([]model.RideshareRoleWithAssignments, 0, len(roles))
	for _, role := range roles {
		assignmentPtrs, err := r.GetAssignmentsByRole(ctx, role.ID)
		if err != nil {
			continue
		}

		// Convert []*RideshareRoleAssignment to []RideshareRoleAssignment
		assignments := make([]model.RideshareRoleAssignment, len(assignmentPtrs))
		for i, a := range assignmentPtrs {
			assignments[i] = *a
		}

		spotsLeft := -1 // unlimited
		if role.MaxSlots > 0 {
			spotsLeft = role.MaxSlots - role.FilledSlots
			if spotsLeft < 0 {
				spotsLeft = 0
			}
		}

		result = append(result, model.RideshareRoleWithAssignments{
			Role:        *role,
			Assignments: assignments,
			IsFull:      role.MaxSlots > 0 && role.FilledSlots >= role.MaxSlots,
			SpotsLeft:   spotsLeft,
		})
	}

	return result, nil
}

// Parsing helpers

func (r *RideshareRoleRepository) parseRideshareRole(result interface{}) (*model.RideshareRole, error) {
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	role := &model.RideshareRole{
		ID:          convertSurrealID(data["id"]),
		RideshareID: convertSurrealID(data["rideshare_id"]),
		Name:        getString(data, "name"),
		MaxSlots:    getInt(data, "max_slots"),
		FilledSlots: getInt(data, "filled_slots"),
		SortOrder:   getInt(data, "sort_order"),
		CreatedBy:   convertSurrealID(data["created_by"]),
	}

	if catalogID := convertSurrealID(data["catalog_role_id"]); catalogID != "" {
		role.CatalogRoleID = &catalogID
	}
	if desc := getString(data, "description"); desc != "" {
		role.Description = &desc
	}
	if t := getTime(data, "created_on"); t != nil {
		role.CreatedOn = *t
	}
	if t := getTime(data, "updated_on"); t != nil {
		role.UpdatedOn = *t
	}

	return role, nil
}

func (r *RideshareRoleRepository) parseRideshareRoles(result []interface{}) ([]*model.RideshareRole, error) {
	roles := make([]*model.RideshareRole, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					role, err := r.parseRideshareRole(item)
					if err != nil {
						continue
					}
					roles = append(roles, role)
				}
			}
		}
	}

	return roles, nil
}

func (r *RideshareRoleRepository) parseRideshareRoleAssignment(result interface{}) (*model.RideshareRoleAssignment, error) {
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	assignment := &model.RideshareRoleAssignment{
		ID:          convertSurrealID(data["id"]),
		RideshareID: convertSurrealID(data["rideshare_id"]),
		RoleID:      convertSurrealID(data["role_id"]),
		UserID:      convertSurrealID(data["user_id"]),
		Status:      getString(data, "status"),
	}

	if note := getString(data, "note"); note != "" {
		assignment.Note = &note
	}
	if t := getTime(data, "assigned_on"); t != nil {
		assignment.AssignedOn = *t
	}
	if t := getTime(data, "updated_on"); t != nil {
		assignment.UpdatedOn = *t
	}

	return assignment, nil
}

func (r *RideshareRoleRepository) parseRideshareRoleAssignments(result []interface{}) ([]*model.RideshareRoleAssignment, error) {
	assignments := make([]*model.RideshareRoleAssignment, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					assignment, err := r.parseRideshareRoleAssignment(item)
					if err != nil {
						continue
					}
					assignments = append(assignments, assignment)
				}
			}
		}
	}

	return assignments, nil
}
