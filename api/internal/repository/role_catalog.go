package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// RoleCatalogRepository handles role catalog data access
type RoleCatalogRepository struct {
	db database.Database
}

// NewRoleCatalogRepository creates a new role catalog repository
func NewRoleCatalogRepository(db database.Database) *RoleCatalogRepository {
	return &RoleCatalogRepository{db: db}
}

// Create creates a new role catalog entry
func (r *RoleCatalogRepository) Create(ctx context.Context, catalog *model.RoleCatalog) error {
	// Build query dynamically to avoid NULL vs NONE issues for optional fields
	setClause := `scope_type = $scope_type, scope_id = $scope_id, role_type = $role_type, name = $name, is_active = true, created_by = type::record($created_by), created_on = time::now(), updated_on = time::now()`
	vars := map[string]interface{}{
		"scope_type": catalog.ScopeType,
		"scope_id":   catalog.ScopeID,
		"role_type":  catalog.RoleType,
		"name":       catalog.Name,
		"created_by": catalog.CreatedBy,
	}

	// Add optional fields only when they have values
	if catalog.Description != nil && *catalog.Description != "" {
		setClause += ", description = $description"
		vars["description"] = *catalog.Description
	}
	if catalog.Icon != nil && *catalog.Icon != "" {
		setClause += ", icon = $icon"
		vars["icon"] = *catalog.Icon
	}

	query := "CREATE role_catalog SET " + setClause

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		if isUniqueConstraintError(err) {
			return fmt.Errorf("role catalog with this name already exists")
		}
		return fmt.Errorf("failed to create role catalog: %w", err)
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return fmt.Errorf("failed to extract created catalog: %w", err)
	}

	catalog.ID = created.ID
	catalog.CreatedOn = created.CreatedOn
	catalog.UpdatedOn = created.UpdatedOn
	catalog.IsActive = true
	return nil
}

// GetByID retrieves a role catalog by ID
func (r *RoleCatalogRepository) GetByID(ctx context.Context, id string) (*model.RoleCatalog, error) {
	query := `SELECT * FROM type::record($id)`
	result, err := r.db.QueryOne(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role catalog: %w", err)
	}

	return r.parseRoleCatalog(result)
}

// GetByScope retrieves all role catalogs for a scope
func (r *RoleCatalogRepository) GetByScope(ctx context.Context, scopeType model.RoleCatalogScopeType, scopeID string, roleType *model.RoleCatalogRoleType) ([]*model.RoleCatalog, error) {
	query := `
		SELECT * FROM role_catalog
		WHERE scope_type = $scope_type
		AND scope_id = $scope_id
		AND is_active = true
	`
	vars := map[string]interface{}{
		"scope_type": scopeType,
		"scope_id":   scopeID,
	}

	if roleType != nil {
		query += ` AND role_type = $role_type`
		vars["role_type"] = *roleType
	}

	query += ` ORDER BY name ASC`

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get role catalogs: %w", err)
	}

	return r.parseRoleCatalogs(result)
}

// GetGuildCatalogs retrieves role catalogs for a guild
func (r *RoleCatalogRepository) GetGuildCatalogs(ctx context.Context, guildID string, roleType *model.RoleCatalogRoleType) ([]*model.RoleCatalog, error) {
	scopeID := fmt.Sprintf("guild:%s", guildID)
	return r.GetByScope(ctx, model.RoleCatalogScopeGuild, scopeID, roleType)
}

// GetUserCatalogs retrieves role catalogs for a user
func (r *RoleCatalogRepository) GetUserCatalogs(ctx context.Context, userID string, roleType *model.RoleCatalogRoleType) ([]*model.RoleCatalog, error) {
	scopeID := fmt.Sprintf("user:%s", userID)
	return r.GetByScope(ctx, model.RoleCatalogScopeUser, scopeID, roleType)
}

// Update updates a role catalog
func (r *RoleCatalogRepository) Update(ctx context.Context, id string, updates *model.UpdateRoleCatalogRequest) (*model.RoleCatalog, error) {
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
	if updates.Icon != nil {
		query += `, icon = $icon`
		vars["icon"] = *updates.Icon
	}
	if updates.IsActive != nil {
		query += `, is_active = $is_active`
		vars["is_active"] = *updates.IsActive
	}

	query += ` RETURN AFTER`

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if isUniqueConstraintError(err) {
			return nil, fmt.Errorf("role catalog with this name already exists")
		}
		return nil, fmt.Errorf("failed to update role catalog: %w", err)
	}

	return r.parseRoleCatalog(result)
}

// Delete deletes a role catalog
func (r *RoleCatalogRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE type::record($id)`
	if err := r.db.Execute(ctx, query, map[string]interface{}{"id": id}); err != nil {
		return fmt.Errorf("failed to delete role catalog: %w", err)
	}
	return nil
}

// Parsing helpers

func (r *RoleCatalogRepository) parseRoleCatalog(result interface{}) (*model.RoleCatalog, error) {
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	catalog := &model.RoleCatalog{
		ID:        convertSurrealID(data["id"]),
		ScopeType: model.RoleCatalogScopeType(getString(data, "scope_type")),
		ScopeID:   getString(data, "scope_id"),
		RoleType:  model.RoleCatalogRoleType(getString(data, "role_type")),
		Name:      getString(data, "name"),
		IsActive:  getBool(data, "is_active"),
		CreatedBy: convertSurrealID(data["created_by"]),
	}

	if desc := getString(data, "description"); desc != "" {
		catalog.Description = &desc
	}
	if icon := getString(data, "icon"); icon != "" {
		catalog.Icon = &icon
	}
	if t := getTime(data, "created_on"); t != nil {
		catalog.CreatedOn = *t
	}
	if t := getTime(data, "updated_on"); t != nil {
		catalog.UpdatedOn = *t
	}

	return catalog, nil
}

func (r *RoleCatalogRepository) parseRoleCatalogs(result []interface{}) ([]*model.RoleCatalog, error) {
	catalogs := make([]*model.RoleCatalog, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					catalog, err := r.parseRoleCatalog(item)
					if err != nil {
						continue
					}
					catalogs = append(catalogs, catalog)
				}
			}
		}
	}

	return catalogs, nil
}
