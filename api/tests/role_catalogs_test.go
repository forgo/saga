package tests

/*
FEATURE: Role Catalogs
DOMAIN: Role Templates & Assignments

ACCEPTANCE CRITERIA:
===================

AC-ROLE-001: Create Guild Role Catalog
  GIVEN user is guild admin
  WHEN user creates role catalog (type=event)
  THEN catalog created with guild scope

AC-ROLE-002: Create Guild Role - Non-Admin
  GIVEN user is regular member
  WHEN user creates guild role catalog
  THEN fails with 403 Forbidden

AC-ROLE-003: Create User Role Catalog
  GIVEN authenticated user
  WHEN user creates personal role catalog
  THEN catalog created with user scope

AC-ROLE-004: Unique Name Per Scope
  GIVEN role catalog "Driver" in guild
  WHEN creating another "Driver" in same guild
  THEN fails with 409 Conflict

AC-ROLE-005: Create Event Role from Catalog
  GIVEN role catalog exists
  WHEN organizer creates event role from catalog
  THEN role created with catalog reference

AC-ROLE-006: Create Rideshare Role from Catalog
  GIVEN role catalog (type=rideshare)
  WHEN organizer creates rideshare role from catalog
  THEN role created with catalog reference

AC-ROLE-007: Rideshare Role Assignment
  GIVEN rideshare role with max_slots=2
  WHEN user requests role assignment
  THEN assignment created with status=requested

AC-ROLE-008: Role Assignment - Full
  GIVEN role with max_slots=2, filled_slots=2
  WHEN user requests assignment
  THEN fails with 409 Conflict
*/

import (
	"context"
	"fmt"
	"testing"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/repository"
	"github.com/forgo/saga/api/internal/testing/fixtures"
	"github.com/forgo/saga/api/internal/testing/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleCatalog_CreateGuildCatalog(t *testing.T) {
	// AC-ROLE-001: Create Guild Role Catalog
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	catalogRepo := repository.NewRoleCatalogRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	catalog := &model.RoleCatalog{
		ScopeType: model.RoleCatalogScopeGuild,
		ScopeID:   fmt.Sprintf("guild:%s", guild.ID),
		RoleType:  model.RoleCatalogRoleEvent,
		Name:      "DJ",
		CreatedBy: user.ID,
	}

	err := catalogRepo.Create(ctx, catalog)
	require.NoError(t, err)
	assert.NotEmpty(t, catalog.ID)
	assert.True(t, catalog.IsActive)
	assert.NotZero(t, catalog.CreatedOn)

	// Verify catalog can be retrieved
	fetched, err := catalogRepo.GetByID(ctx, catalog.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, "DJ", fetched.Name)
	assert.Equal(t, model.RoleCatalogScopeGuild, fetched.ScopeType)
	assert.Equal(t, model.RoleCatalogRoleEvent, fetched.RoleType)
}

func TestRoleCatalog_CreateUserCatalog(t *testing.T) {
	// AC-ROLE-003: Create User Role Catalog
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	catalogRepo := repository.NewRoleCatalogRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)

	description := "I'm bringing music to the party"
	catalog := &model.RoleCatalog{
		ScopeType:   model.RoleCatalogScopeUser,
		ScopeID:     fmt.Sprintf("user:%s", user.ID),
		RoleType:    model.RoleCatalogRoleEvent,
		Name:        "Music Provider",
		Description: &description,
		CreatedBy:   user.ID,
	}

	err := catalogRepo.Create(ctx, catalog)
	require.NoError(t, err)
	assert.NotEmpty(t, catalog.ID)

	// Verify catalog can be retrieved
	fetched, err := catalogRepo.GetByID(ctx, catalog.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, "Music Provider", fetched.Name)
	assert.Equal(t, model.RoleCatalogScopeUser, fetched.ScopeType)
	require.NotNil(t, fetched.Description)
	assert.Equal(t, description, *fetched.Description)
}

func TestRoleCatalog_UniqueNamePerScope(t *testing.T) {
	// AC-ROLE-004: Unique Name Per Scope
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	catalogRepo := repository.NewRoleCatalogRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	scopeID := fmt.Sprintf("guild:%s", guild.ID)

	// Create first catalog
	catalog1 := &model.RoleCatalog{
		ScopeType: model.RoleCatalogScopeGuild,
		ScopeID:   scopeID,
		RoleType:  model.RoleCatalogRoleEvent,
		Name:      "Driver",
		CreatedBy: user.ID,
	}
	err := catalogRepo.Create(ctx, catalog1)
	require.NoError(t, err)

	// Try to create duplicate with same name in same scope and type
	catalog2 := &model.RoleCatalog{
		ScopeType: model.RoleCatalogScopeGuild,
		ScopeID:   scopeID,
		RoleType:  model.RoleCatalogRoleEvent,
		Name:      "Driver",
		CreatedBy: user.ID,
	}
	err = catalogRepo.Create(ctx, catalog2)
	assert.Error(t, err, "Should fail to create duplicate role catalog name in same scope")
	assert.Contains(t, err.Error(), "already exists")
}

func TestRoleCatalog_DifferentTypesSameName(t *testing.T) {
	// Same name allowed for different role types (event vs rideshare)
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	catalogRepo := repository.NewRoleCatalogRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	scopeID := fmt.Sprintf("guild:%s", guild.ID)

	// Create event role catalog
	eventCatalog := &model.RoleCatalog{
		ScopeType: model.RoleCatalogScopeGuild,
		ScopeID:   scopeID,
		RoleType:  model.RoleCatalogRoleEvent,
		Name:      "Driver",
		CreatedBy: user.ID,
	}
	err := catalogRepo.Create(ctx, eventCatalog)
	require.NoError(t, err)

	// Create rideshare role catalog with same name - should succeed
	rideshareCatalog := &model.RoleCatalog{
		ScopeType: model.RoleCatalogScopeGuild,
		ScopeID:   scopeID,
		RoleType:  model.RoleCatalogRoleRideshare,
		Name:      "Driver",
		CreatedBy: user.ID,
	}
	err = catalogRepo.Create(ctx, rideshareCatalog)
	require.NoError(t, err, "Same name should be allowed for different role types")
}

func TestRoleCatalog_ListByScope(t *testing.T) {
	// List role catalogs for a specific scope
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	catalogRepo := repository.NewRoleCatalogRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create multiple catalogs
	roles := []string{"DJ", "Dessert Bringer", "Photographer"}
	for _, name := range roles {
		catalog := &model.RoleCatalog{
			ScopeType: model.RoleCatalogScopeGuild,
			ScopeID:   fmt.Sprintf("guild:%s", guild.ID),
			RoleType:  model.RoleCatalogRoleEvent,
			Name:      name,
			CreatedBy: user.ID,
		}
		err := catalogRepo.Create(ctx, catalog)
		require.NoError(t, err)
	}

	// List catalogs for guild
	catalogs, err := catalogRepo.GetGuildCatalogs(ctx, guild.ID, nil)
	require.NoError(t, err)
	assert.Len(t, catalogs, 3, "Should return all 3 guild catalogs")

	// List with role type filter
	eventType := model.RoleCatalogRoleEvent
	eventCatalogs, err := catalogRepo.GetGuildCatalogs(ctx, guild.ID, &eventType)
	require.NoError(t, err)
	assert.Len(t, eventCatalogs, 3, "All catalogs are event type")
}

func TestRoleCatalog_Update(t *testing.T) {
	// Update role catalog
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	catalogRepo := repository.NewRoleCatalogRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	catalog := &model.RoleCatalog{
		ScopeType: model.RoleCatalogScopeGuild,
		ScopeID:   fmt.Sprintf("guild:%s", guild.ID),
		RoleType:  model.RoleCatalogRoleEvent,
		Name:      "DJ",
		CreatedBy: user.ID,
	}
	err := catalogRepo.Create(ctx, catalog)
	require.NoError(t, err)

	// Update catalog
	newName := "Music DJ"
	newDesc := "Spin the tunes!"
	updated, err := catalogRepo.Update(ctx, catalog.ID, &model.UpdateRoleCatalogRequest{
		Name:        &newName,
		Description: &newDesc,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Music DJ", updated.Name)
	require.NotNil(t, updated.Description)
	assert.Equal(t, "Spin the tunes!", *updated.Description)
}

func TestRoleCatalog_Deactivate(t *testing.T) {
	// Deactivate role catalog
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	catalogRepo := repository.NewRoleCatalogRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	catalog := &model.RoleCatalog{
		ScopeType: model.RoleCatalogScopeGuild,
		ScopeID:   fmt.Sprintf("guild:%s", guild.ID),
		RoleType:  model.RoleCatalogRoleEvent,
		Name:      "DJ",
		CreatedBy: user.ID,
	}
	err := catalogRepo.Create(ctx, catalog)
	require.NoError(t, err)
	assert.True(t, catalog.IsActive)

	// Deactivate
	isActive := false
	updated, err := catalogRepo.Update(ctx, catalog.ID, &model.UpdateRoleCatalogRequest{
		IsActive: &isActive,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.False(t, updated.IsActive)

	// Deactivated catalog should not appear in listings
	catalogs, err := catalogRepo.GetGuildCatalogs(ctx, guild.ID, nil)
	require.NoError(t, err)
	assert.Len(t, catalogs, 0, "Deactivated catalog should not appear in list")
}

func TestRoleCatalog_Delete(t *testing.T) {
	// Delete role catalog
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	catalogRepo := repository.NewRoleCatalogRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	catalog := &model.RoleCatalog{
		ScopeType: model.RoleCatalogScopeGuild,
		ScopeID:   fmt.Sprintf("guild:%s", guild.ID),
		RoleType:  model.RoleCatalogRoleEvent,
		Name:      "DJ",
		CreatedBy: user.ID,
	}
	err := catalogRepo.Create(ctx, catalog)
	require.NoError(t, err)

	// Delete
	err = catalogRepo.Delete(ctx, catalog.ID)
	require.NoError(t, err)

	// Verify deleted
	fetched, err := catalogRepo.GetByID(ctx, catalog.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched, "Catalog should be deleted")
}

func TestEventRole_Create(t *testing.T) {
	// AC-ROLE-005: Create Event Role
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	roleRepo := repository.NewEventRoleRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	event := f.CreateEvent(t, guild, user)

	role := &model.EventRole{
		EventID:   event.ID,
		Name:      "DJ",
		MaxSlots:  1,
		IsDefault: false,
		SortOrder: 0,
		CreatedBy: user.ID,
	}

	err := roleRepo.CreateRole(ctx, role)
	require.NoError(t, err)
	assert.NotEmpty(t, role.ID)
	assert.NotZero(t, role.CreatedOn)

	// Verify role can be retrieved
	fetched, err := roleRepo.GetRole(ctx, role.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, "DJ", fetched.Name)
	assert.Equal(t, 1, fetched.MaxSlots)
}

func TestEventRole_ListByEvent(t *testing.T) {
	// List roles for an event
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	roleRepo := repository.NewEventRoleRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	event := f.CreateEvent(t, guild, user)

	// Create multiple roles
	roleNames := []string{"DJ", "Dessert Bringer", "Photographer"}
	for i, name := range roleNames {
		role := &model.EventRole{
			EventID:   event.ID,
			Name:      name,
			MaxSlots:  2,
			IsDefault: false,
			SortOrder: i,
			CreatedBy: user.ID,
		}
		err := roleRepo.CreateRole(ctx, role)
		require.NoError(t, err)
	}

	// List roles
	roles, err := roleRepo.GetRolesByEvent(ctx, event.ID)
	require.NoError(t, err)
	assert.Len(t, roles, 3, "Should return all 3 roles")

	// Verify sorted by sort_order
	assert.Equal(t, "DJ", roles[0].Name)
	assert.Equal(t, "Dessert Bringer", roles[1].Name)
	assert.Equal(t, "Photographer", roles[2].Name)
}

func TestEventRoleAssignment_Create(t *testing.T) {
	// AC-ROLE-007: Role Assignment
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	roleRepo := repository.NewEventRoleRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	event := f.CreateEvent(t, guild, user)

	// Create role with slots
	role := &model.EventRole{
		EventID:   event.ID,
		Name:      "DJ",
		MaxSlots:  2,
		IsDefault: false,
		SortOrder: 0,
		CreatedBy: user.ID,
	}
	err := roleRepo.CreateRole(ctx, role)
	require.NoError(t, err)

	// Create assignment
	assignment := &model.EventRoleAssignment{
		EventID: event.ID,
		RoleID:  role.ID,
		UserID:  user.ID,
		Status:  model.RoleAssignmentStatusConfirmed,
	}
	err = roleRepo.CreateAssignment(ctx, assignment)
	require.NoError(t, err)
	assert.NotEmpty(t, assignment.ID)
	assert.NotZero(t, assignment.AssignedOn)

	// Verify assignment can be retrieved
	fetched, err := roleRepo.GetAssignment(ctx, assignment.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, user.ID, fetched.UserID)
	assert.Equal(t, model.RoleAssignmentStatusConfirmed, fetched.Status)
}

func TestEventRoleAssignment_WithNote(t *testing.T) {
	// Assignment with note
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	roleRepo := repository.NewEventRoleRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	event := f.CreateEvent(t, guild, user)

	role := &model.EventRole{
		EventID:   event.ID,
		Name:      "Dessert Bringer",
		MaxSlots:  5,
		IsDefault: false,
		SortOrder: 0,
		CreatedBy: user.ID,
	}
	err := roleRepo.CreateRole(ctx, role)
	require.NoError(t, err)

	note := "I'll bring vegan brownies"
	assignment := &model.EventRoleAssignment{
		EventID: event.ID,
		RoleID:  role.ID,
		UserID:  user.ID,
		Note:    &note,
		Status:  model.RoleAssignmentStatusConfirmed,
	}
	err = roleRepo.CreateAssignment(ctx, assignment)
	require.NoError(t, err)

	fetched, err := roleRepo.GetAssignment(ctx, assignment.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.NotNil(t, fetched.Note)
	assert.Equal(t, "I'll bring vegan brownies", *fetched.Note)
}

func TestEventRoleAssignment_UniquePerUserRole(t *testing.T) {
	// User can only have one assignment per role
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	roleRepo := repository.NewEventRoleRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	event := f.CreateEvent(t, guild, user)

	role := &model.EventRole{
		EventID:   event.ID,
		Name:      "DJ",
		MaxSlots:  2,
		IsDefault: false,
		SortOrder: 0,
		CreatedBy: user.ID,
	}
	err := roleRepo.CreateRole(ctx, role)
	require.NoError(t, err)

	// First assignment
	assignment1 := &model.EventRoleAssignment{
		EventID: event.ID,
		RoleID:  role.ID,
		UserID:  user.ID,
		Status:  model.RoleAssignmentStatusConfirmed,
	}
	err = roleRepo.CreateAssignment(ctx, assignment1)
	require.NoError(t, err)

	// Second assignment for same user+role should fail
	assignment2 := &model.EventRoleAssignment{
		EventID: event.ID,
		RoleID:  role.ID,
		UserID:  user.ID,
		Status:  model.RoleAssignmentStatusConfirmed,
	}
	err = roleRepo.CreateAssignment(ctx, assignment2)
	assert.Error(t, err, "Should fail to create duplicate assignment")
}

func TestEventRoleAssignment_MultipleRolesPerUser(t *testing.T) {
	// User can have multiple role assignments per event (different roles)
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	roleRepo := repository.NewEventRoleRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	event := f.CreateEvent(t, guild, user)

	// Create two different roles
	djRole := &model.EventRole{
		EventID:   event.ID,
		Name:      "DJ",
		MaxSlots:  1,
		IsDefault: false,
		SortOrder: 0,
		CreatedBy: user.ID,
	}
	err := roleRepo.CreateRole(ctx, djRole)
	require.NoError(t, err)

	dessertRole := &model.EventRole{
		EventID:   event.ID,
		Name:      "Dessert Bringer",
		MaxSlots:  5,
		IsDefault: false,
		SortOrder: 1,
		CreatedBy: user.ID,
	}
	err = roleRepo.CreateRole(ctx, dessertRole)
	require.NoError(t, err)

	// Assign user to both roles
	djAssignment := &model.EventRoleAssignment{
		EventID: event.ID,
		RoleID:  djRole.ID,
		UserID:  user.ID,
		Status:  model.RoleAssignmentStatusConfirmed,
	}
	err = roleRepo.CreateAssignment(ctx, djAssignment)
	require.NoError(t, err)

	dessertAssignment := &model.EventRoleAssignment{
		EventID: event.ID,
		RoleID:  dessertRole.ID,
		UserID:  user.ID,
		Status:  model.RoleAssignmentStatusConfirmed,
	}
	err = roleRepo.CreateAssignment(ctx, dessertAssignment)
	require.NoError(t, err)

	// Get all user assignments for event
	assignments, err := roleRepo.GetUserAssignmentsForEvent(ctx, event.ID, user.ID)
	require.NoError(t, err)
	assert.Len(t, assignments, 2, "User should have 2 role assignments")
}

func TestEventRole_RolesWithAssignments(t *testing.T) {
	// Get roles with their assignments
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	roleRepo := repository.NewEventRoleRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	user2 := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	event := f.CreateEvent(t, guild, user)

	// Create role with max_slots=2
	role := &model.EventRole{
		EventID:   event.ID,
		Name:      "DJ",
		MaxSlots:  2,
		IsDefault: false,
		SortOrder: 0,
		CreatedBy: user.ID,
	}
	err := roleRepo.CreateRole(ctx, role)
	require.NoError(t, err)

	// Add one assignment
	assignment := &model.EventRoleAssignment{
		EventID: event.ID,
		RoleID:  role.ID,
		UserID:  user.ID,
		Status:  model.RoleAssignmentStatusConfirmed,
	}
	err = roleRepo.CreateAssignment(ctx, assignment)
	require.NoError(t, err)

	// Get roles with assignments
	rolesWithAssignments, err := roleRepo.GetRolesWithAssignments(ctx, event.ID)
	require.NoError(t, err)
	require.Len(t, rolesWithAssignments, 1)

	roleWithAssignment := rolesWithAssignments[0]
	assert.Equal(t, "DJ", roleWithAssignment.Role.Name)
	assert.Len(t, roleWithAssignment.Assignments, 1)
	assert.False(t, roleWithAssignment.IsFull, "Role with 1/2 slots should not be full")
	assert.Equal(t, 1, roleWithAssignment.SpotsLeft)

	// Add second assignment to fill the role
	assignment2 := &model.EventRoleAssignment{
		EventID: event.ID,
		RoleID:  role.ID,
		UserID:  user2.ID,
		Status:  model.RoleAssignmentStatusConfirmed,
	}
	err = roleRepo.CreateAssignment(ctx, assignment2)
	require.NoError(t, err)

	// Check again
	rolesWithAssignments, err = roleRepo.GetRolesWithAssignments(ctx, event.ID)
	require.NoError(t, err)
	require.Len(t, rolesWithAssignments, 1)

	roleWithAssignment = rolesWithAssignments[0]
	assert.Len(t, roleWithAssignment.Assignments, 2)
	assert.True(t, roleWithAssignment.IsFull, "Role with 2/2 slots should be full")
	assert.Equal(t, 0, roleWithAssignment.SpotsLeft)
}

func TestEventRole_UnlimitedSlots(t *testing.T) {
	// Role with unlimited slots (max_slots=0)
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	roleRepo := repository.NewEventRoleRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	event := f.CreateEvent(t, guild, user)

	// Create role with unlimited slots
	role := &model.EventRole{
		EventID:   event.ID,
		Name:      "Guest",
		MaxSlots:  0, // 0 = unlimited
		IsDefault: true,
		SortOrder: 0,
		CreatedBy: user.ID,
	}
	err := roleRepo.CreateRole(ctx, role)
	require.NoError(t, err)

	// Add multiple assignments
	for i := 0; i < 10; i++ {
		newUser := f.CreateUser(t)
		assignment := &model.EventRoleAssignment{
			EventID: event.ID,
			RoleID:  role.ID,
			UserID:  newUser.ID,
			Status:  model.RoleAssignmentStatusConfirmed,
		}
		err = roleRepo.CreateAssignment(ctx, assignment)
		require.NoError(t, err)
	}

	// Get roles with assignments
	rolesWithAssignments, err := roleRepo.GetRolesWithAssignments(ctx, event.ID)
	require.NoError(t, err)
	require.Len(t, rolesWithAssignments, 1)

	roleWithAssignment := rolesWithAssignments[0]
	assert.Len(t, roleWithAssignment.Assignments, 10)
	assert.False(t, roleWithAssignment.IsFull, "Unlimited role should never be full")
	assert.Equal(t, -1, roleWithAssignment.SpotsLeft, "Unlimited role should have -1 spots left")
}
