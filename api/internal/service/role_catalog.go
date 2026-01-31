package service

import (
	"context"
	"fmt"

	"github.com/forgo/saga/api/internal/model"
)

// RoleCatalogRepository defines the interface for role catalog storage
type RoleCatalogRepository interface {
	Create(ctx context.Context, catalog *model.RoleCatalog) error
	GetByID(ctx context.Context, id string) (*model.RoleCatalog, error)
	GetByScope(ctx context.Context, scopeType model.RoleCatalogScopeType, scopeID string, roleType *model.RoleCatalogRoleType) ([]*model.RoleCatalog, error)
	GetGuildCatalogs(ctx context.Context, guildID string, roleType *model.RoleCatalogRoleType) ([]*model.RoleCatalog, error)
	GetUserCatalogs(ctx context.Context, userID string, roleType *model.RoleCatalogRoleType) ([]*model.RoleCatalog, error)
	Update(ctx context.Context, id string, updates *model.UpdateRoleCatalogRequest) (*model.RoleCatalog, error)
	Delete(ctx context.Context, id string) error
}

// RideshareRoleRepository defines the interface for rideshare role storage
type RideshareRoleRepository interface {
	Create(ctx context.Context, role *model.RideshareRole) error
	GetByID(ctx context.Context, id string) (*model.RideshareRole, error)
	GetByRideshare(ctx context.Context, rideshareID string) ([]*model.RideshareRole, error)
	Update(ctx context.Context, id string, updates *model.UpdateRideshareRoleRequest) (*model.RideshareRole, error)
	Delete(ctx context.Context, id string) error
	CreateAssignment(ctx context.Context, assignment *model.RideshareRoleAssignment) error
	GetAssignmentByID(ctx context.Context, id string) (*model.RideshareRoleAssignment, error)
	GetAssignmentsByRole(ctx context.Context, roleID string) ([]*model.RideshareRoleAssignment, error)
	GetAssignmentsByRideshare(ctx context.Context, rideshareID string) ([]*model.RideshareRoleAssignment, error)
	GetAssignmentsByUser(ctx context.Context, rideshareID, userID string) ([]*model.RideshareRoleAssignment, error)
	DeleteAssignment(ctx context.Context, id string) error
	GetRolesWithAssignments(ctx context.Context, rideshareID string) ([]model.RideshareRoleWithAssignments, error)
}

// RoleCatalogService handles role catalog business logic
type RoleCatalogService struct {
	catalogRepo   RoleCatalogRepository
	rideshareRepo RideshareRoleRepository
	guildRepo     GuildRepository // Uses GuildRepository which has IsMember
}

// RoleCatalogServiceConfig holds configuration for the role catalog service
type RoleCatalogServiceConfig struct {
	CatalogRepo   RoleCatalogRepository
	RideshareRepo RideshareRoleRepository
	MemberRepo    interface{} // Deprecated, kept for backwards compatibility
	GuildRepo     GuildRepository
}

// NewRoleCatalogService creates a new role catalog service
func NewRoleCatalogService(cfg RoleCatalogServiceConfig) *RoleCatalogService {
	return &RoleCatalogService{
		catalogRepo:   cfg.CatalogRepo,
		rideshareRepo: cfg.RideshareRepo,
		guildRepo:     cfg.GuildRepo,
	}
}

// Guild catalog operations

// CreateGuildCatalog creates a role catalog for a guild
func (s *RoleCatalogService) CreateGuildCatalog(ctx context.Context, guildID string, userID string, req *model.CreateRoleCatalogRequest) (*model.RoleCatalog, error) {
	// Validate request
	if errors := req.Validate(); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	// Check if user is guild admin
	if s.guildRepo != nil {
		isAdmin, err := s.guildRepo.IsGuildAdmin(ctx, userID, guildID)
		if err != nil {
			return nil, fmt.Errorf("failed to check admin status: %w", err)
		}
		if !isAdmin {
			return nil, model.NewForbiddenError("must be guild admin to create catalogs")
		}
	}

	catalog := &model.RoleCatalog{
		ScopeType:   model.RoleCatalogScopeGuild,
		ScopeID:     fmt.Sprintf("guild:%s", guildID),
		RoleType:    model.RoleCatalogRoleType(req.RoleType),
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		CreatedBy:   userID,
	}

	if err := s.catalogRepo.Create(ctx, catalog); err != nil {
		return nil, fmt.Errorf("failed to create catalog: %w", err)
	}

	return catalog, nil
}

// GetGuildCatalogs retrieves role catalogs for a guild
func (s *RoleCatalogService) GetGuildCatalogs(ctx context.Context, guildID string, roleType *string) ([]*model.RoleCatalog, error) {
	var rt *model.RoleCatalogRoleType
	if roleType != nil {
		t := model.RoleCatalogRoleType(*roleType)
		rt = &t
	}
	return s.catalogRepo.GetGuildCatalogs(ctx, guildID, rt)
}

// User catalog operations

// CreateUserCatalog creates a role catalog for a user
func (s *RoleCatalogService) CreateUserCatalog(ctx context.Context, userID string, req *model.CreateRoleCatalogRequest) (*model.RoleCatalog, error) {
	// Validate request
	if errors := req.Validate(); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	catalog := &model.RoleCatalog{
		ScopeType:   model.RoleCatalogScopeUser,
		ScopeID:     fmt.Sprintf("user:%s", userID),
		RoleType:    model.RoleCatalogRoleType(req.RoleType),
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		CreatedBy:   userID,
	}

	if err := s.catalogRepo.Create(ctx, catalog); err != nil {
		return nil, fmt.Errorf("failed to create catalog: %w", err)
	}

	return catalog, nil
}

// GetUserCatalogs retrieves role catalogs for a user
func (s *RoleCatalogService) GetUserCatalogs(ctx context.Context, userID string, roleType *string) ([]*model.RoleCatalog, error) {
	var rt *model.RoleCatalogRoleType
	if roleType != nil {
		t := model.RoleCatalogRoleType(*roleType)
		rt = &t
	}
	return s.catalogRepo.GetUserCatalogs(ctx, userID, rt)
}

// Common catalog operations

// GetCatalogByID retrieves a catalog by ID
func (s *RoleCatalogService) GetCatalogByID(ctx context.Context, id string) (*model.RoleCatalog, error) {
	catalog, err := s.catalogRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog: %w", err)
	}
	if catalog == nil {
		return nil, model.NewNotFoundError("role catalog not found")
	}
	return catalog, nil
}

// UpdateCatalog updates a role catalog
func (s *RoleCatalogService) UpdateCatalog(ctx context.Context, id string, userID string, req *model.UpdateRoleCatalogRequest) (*model.RoleCatalog, error) {
	// Validate request
	if errors := req.Validate(); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	// Get existing catalog
	catalog, err := s.catalogRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog: %w", err)
	}
	if catalog == nil {
		return nil, model.NewNotFoundError("role catalog not found")
	}

	// Check permission
	if err := s.checkCatalogPermission(ctx, catalog, userID); err != nil {
		return nil, err
	}

	updated, err := s.catalogRepo.Update(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update catalog: %w", err)
	}

	return updated, nil
}

// DeleteCatalog deletes a role catalog
func (s *RoleCatalogService) DeleteCatalog(ctx context.Context, id string, userID string) error {
	catalog, err := s.catalogRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get catalog: %w", err)
	}
	if catalog == nil {
		return model.NewNotFoundError("role catalog not found")
	}

	if err := s.checkCatalogPermission(ctx, catalog, userID); err != nil {
		return err
	}

	if err := s.catalogRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete catalog: %w", err)
	}

	return nil
}

// Rideshare role operations

// CreateRideshareRole creates a role for a rideshare
func (s *RoleCatalogService) CreateRideshareRole(ctx context.Context, rideshareID string, userID string, req *model.CreateRideshareRoleRequest) (*model.RideshareRole, error) {
	// Validate request
	if errors := req.Validate(); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	role := &model.RideshareRole{
		RideshareID:   rideshareID,
		CatalogRoleID: req.CatalogRoleID,
		Name:          req.Name,
		Description:   req.Description,
		MaxSlots:      req.MaxSlots,
		CreatedBy:     userID,
	}

	// If creating from catalog, copy values
	if req.CatalogRoleID != nil {
		catalog, err := s.catalogRepo.GetByID(ctx, *req.CatalogRoleID)
		if err != nil {
			return nil, fmt.Errorf("failed to get catalog: %w", err)
		}
		if catalog != nil {
			if role.Name == "" {
				role.Name = catalog.Name
			}
			if role.Description == nil {
				role.Description = catalog.Description
			}
		}
	}

	if err := s.rideshareRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

// GetRideshareRoles retrieves all roles for a rideshare
func (s *RoleCatalogService) GetRideshareRoles(ctx context.Context, rideshareID string) ([]*model.RideshareRole, error) {
	return s.rideshareRepo.GetByRideshare(ctx, rideshareID)
}

// GetRideshareRolesWithAssignments retrieves roles with their assignments
func (s *RoleCatalogService) GetRideshareRolesWithAssignments(ctx context.Context, rideshareID string) ([]model.RideshareRoleWithAssignments, error) {
	return s.rideshareRepo.GetRolesWithAssignments(ctx, rideshareID)
}

// UpdateRideshareRole updates a rideshare role
func (s *RoleCatalogService) UpdateRideshareRole(ctx context.Context, roleID string, req *model.UpdateRideshareRoleRequest) (*model.RideshareRole, error) {
	if errors := req.Validate(); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	return s.rideshareRepo.Update(ctx, roleID, req)
}

// DeleteRideshareRole deletes a rideshare role
func (s *RoleCatalogService) DeleteRideshareRole(ctx context.Context, roleID string) error {
	return s.rideshareRepo.Delete(ctx, roleID)
}

// AssignRideshareRole assigns a user to a rideshare role
func (s *RoleCatalogService) AssignRideshareRole(ctx context.Context, rideshareID string, userID string, req *model.AssignRideshareRoleRequest) (*model.RideshareRoleAssignment, error) {
	if errors := req.Validate(); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	// Get the role to check capacity
	role, err := s.rideshareRepo.GetByID(ctx, req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return nil, model.NewNotFoundError("role not found")
	}

	// Check if role is full
	if role.MaxSlots > 0 && role.FilledSlots >= role.MaxSlots {
		return nil, model.NewConflictError("role is full")
	}

	assignment := &model.RideshareRoleAssignment{
		RideshareID: rideshareID,
		RoleID:      req.RoleID,
		UserID:      userID,
		Note:        req.Note,
		Status:      "confirmed",
	}

	if err := s.rideshareRepo.CreateAssignment(ctx, assignment); err != nil {
		return nil, fmt.Errorf("failed to create assignment: %w", err)
	}

	return assignment, nil
}

// UnassignRideshareRole removes a user's role assignment
func (s *RoleCatalogService) UnassignRideshareRole(ctx context.Context, assignmentID string, userID string) error {
	assignment, err := s.rideshareRepo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return fmt.Errorf("failed to get assignment: %w", err)
	}
	if assignment == nil {
		return model.NewNotFoundError("assignment not found")
	}

	// Only the assigned user can unassign themselves (or rideshare owner)
	if assignment.UserID != userID {
		return model.NewForbiddenError("not your assignment")
	}

	return s.rideshareRepo.DeleteAssignment(ctx, assignmentID)
}

// GetUserRideshareRoles gets all roles a user has in a rideshare
func (s *RoleCatalogService) GetUserRideshareRoles(ctx context.Context, rideshareID, userID string) ([]*model.RideshareRoleAssignment, error) {
	return s.rideshareRepo.GetAssignmentsByUser(ctx, rideshareID, userID)
}

// Helper methods

func (s *RoleCatalogService) checkCatalogPermission(ctx context.Context, catalog *model.RoleCatalog, userID string) error {
	if catalog.ScopeType == model.RoleCatalogScopeUser {
		// User catalogs - only owner can modify
		expectedScopeID := fmt.Sprintf("user:%s", userID)
		if catalog.ScopeID != expectedScopeID {
			return model.NewForbiddenError("not your catalog")
		}
		return nil
	}

	// Guild catalogs - must be admin
	if s.guildRepo != nil {
		// Extract guild ID from scope_id (format: "guild:<id>")
		guildID := catalog.ScopeID[6:] // Remove "guild:" prefix
		isAdmin, err := s.guildRepo.IsGuildAdmin(ctx, userID, guildID)
		if err != nil {
			return fmt.Errorf("failed to check admin status: %w", err)
		}
		if !isAdmin {
			return model.NewForbiddenError("must be guild admin")
		}
	}

	return nil
}
