// Package tests contains end-to-end acceptance tests for the Saga API.
package tests

import (
	"context"
	"testing"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/repository"
	"github.com/forgo/saga/api/internal/testing/fixtures"
	"github.com/forgo/saga/api/internal/testing/testdb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
FEATURE: Guild Roles
DOMAIN: Community

ACCEPTANCE CRITERIA:
===================

AC-ROLE-001: Creator Is Admin
  GIVEN user creates a guild
  THEN creator is assigned "admin" role

AC-ROLE-002: Default Member Role
  GIVEN existing guild
  WHEN new user joins
  THEN user is assigned "member" role

AC-ROLE-003: Get Member Role
  GIVEN user is member of guild with specific role
  WHEN role is queried
  THEN correct role is returned

AC-ROLE-004: Is Guild Admin
  GIVEN user with admin role
  WHEN IsGuildAdmin is called
  THEN returns true

AC-ROLE-005: Is Guild Moderator
  GIVEN user with moderator role
  WHEN IsGuildModerator is called
  THEN returns true

AC-ROLE-006: Admin Includes Moderator
  GIVEN user with admin role
  WHEN IsGuildModerator is called
  THEN returns true (admin implies moderator)

AC-ROLE-007: Update Member Role
  GIVEN admin user
  WHEN admin promotes member to moderator
  THEN member role is updated

AC-ROLE-008: Member Not Admin
  GIVEN user with member role
  WHEN IsGuildAdmin is called
  THEN returns false
*/

func TestGuildRoles_CreatorIsAdmin(t *testing.T) {
	// AC-ROLE-001: Creator Is Admin
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildRepo := repository.NewGuildRepository(tdb.DB)
	ctx := context.Background()

	// Create user and guild
	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Verify creator has admin role
	role, err := guildRepo.GetMemberRole(ctx, user.ID, guild.ID)
	require.NoError(t, err)
	assert.Equal(t, model.GuildRoleAdmin, role)

	// Verify IsGuildAdmin returns true
	isAdmin, err := guildRepo.IsGuildAdmin(ctx, user.ID, guild.ID)
	require.NoError(t, err)
	assert.True(t, isAdmin, "Creator should be admin")
}

func TestGuildRoles_DefaultMemberRole(t *testing.T) {
	// AC-ROLE-002: Default Member Role
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildRepo := repository.NewGuildRepository(tdb.DB)
	ctx := context.Background()

	// Create admin and member
	admin := f.CreateUser(t)
	guild := f.CreateGuild(t, admin)

	// Add a new member (not admin)
	member := f.CreateUser(t)
	f.AddMemberToGuild(t, member, guild)

	// Verify member has default member role
	role, err := guildRepo.GetMemberRole(ctx, member.ID, guild.ID)
	require.NoError(t, err)
	assert.Equal(t, model.GuildRoleMember, role)

	// Verify member is not admin
	isAdmin, err := guildRepo.IsGuildAdmin(ctx, member.ID, guild.ID)
	require.NoError(t, err)
	assert.False(t, isAdmin, "Regular member should not be admin")
}

func TestGuildRoles_IsGuildAdmin(t *testing.T) {
	// AC-ROLE-004: Is Guild Admin
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildRepo := repository.NewGuildRepository(tdb.DB)
	ctx := context.Background()

	admin := f.CreateUser(t)
	guild := f.CreateGuild(t, admin)

	isAdmin, err := guildRepo.IsGuildAdmin(ctx, admin.ID, guild.ID)
	require.NoError(t, err)
	assert.True(t, isAdmin)
}

func TestGuildRoles_IsGuildModerator(t *testing.T) {
	// AC-ROLE-005: Is Guild Moderator
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildRepo := repository.NewGuildRepository(tdb.DB)
	ctx := context.Background()

	admin := f.CreateUser(t)
	guild := f.CreateGuild(t, admin)
	moderator := f.CreateUser(t)
	f.AddMemberToGuild(t, moderator, guild)

	// Update moderator's role
	err := guildRepo.UpdateMemberRole(ctx, moderator.ID, guild.ID, model.GuildRoleModerator)
	require.NoError(t, err)

	isMod, err := guildRepo.IsGuildModerator(ctx, moderator.ID, guild.ID)
	require.NoError(t, err)
	assert.True(t, isMod)
}

func TestGuildRoles_AdminIncludesModerator(t *testing.T) {
	// AC-ROLE-006: Admin Includes Moderator
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildRepo := repository.NewGuildRepository(tdb.DB)
	ctx := context.Background()

	admin := f.CreateUser(t)
	guild := f.CreateGuild(t, admin)

	// Admin should also be considered a moderator
	isMod, err := guildRepo.IsGuildModerator(ctx, admin.ID, guild.ID)
	require.NoError(t, err)
	assert.True(t, isMod, "Admin should also be moderator")
}

func TestGuildRoles_UpdateMemberRole(t *testing.T) {
	// AC-ROLE-007: Update Member Role
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildRepo := repository.NewGuildRepository(tdb.DB)
	ctx := context.Background()

	admin := f.CreateUser(t)
	guild := f.CreateGuild(t, admin)
	member := f.CreateUser(t)
	f.AddMemberToGuild(t, member, guild)

	// Verify initial role is member
	role, err := guildRepo.GetMemberRole(ctx, member.ID, guild.ID)
	require.NoError(t, err)
	assert.Equal(t, model.GuildRoleMember, role)

	// Promote to moderator
	err = guildRepo.UpdateMemberRole(ctx, member.ID, guild.ID, model.GuildRoleModerator)
	require.NoError(t, err)

	// Verify role changed
	role, err = guildRepo.GetMemberRole(ctx, member.ID, guild.ID)
	require.NoError(t, err)
	assert.Equal(t, model.GuildRoleModerator, role)

	// Promote to admin
	err = guildRepo.UpdateMemberRole(ctx, member.ID, guild.ID, model.GuildRoleAdmin)
	require.NoError(t, err)

	role, err = guildRepo.GetMemberRole(ctx, member.ID, guild.ID)
	require.NoError(t, err)
	assert.Equal(t, model.GuildRoleAdmin, role)
}

func TestGuildRoles_MemberNotAdmin(t *testing.T) {
	// AC-ROLE-008: Member Not Admin
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildRepo := repository.NewGuildRepository(tdb.DB)
	ctx := context.Background()

	admin := f.CreateUser(t)
	guild := f.CreateGuild(t, admin)
	member := f.CreateUser(t)
	f.AddMemberToGuild(t, member, guild)

	isAdmin, err := guildRepo.IsGuildAdmin(ctx, member.ID, guild.ID)
	require.NoError(t, err)
	assert.False(t, isAdmin, "Regular member should not be admin")

	// Member should also not be moderator (unless explicitly set)
	isMod, err := guildRepo.IsGuildModerator(ctx, member.ID, guild.ID)
	require.NoError(t, err)
	assert.False(t, isMod, "Regular member should not be moderator")
}

func TestGuildRoles_NonMemberReturnsEmptyRole(t *testing.T) {
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildRepo := repository.NewGuildRepository(tdb.DB)
	ctx := context.Background()

	admin := f.CreateUser(t)
	guild := f.CreateGuild(t, admin)
	nonMember := f.CreateUser(t)

	// Non-member should return empty role or error
	role, err := guildRepo.GetMemberRole(ctx, nonMember.ID, guild.ID)
	// Depending on implementation, this may return empty string or error
	// Accept either behavior
	if err == nil {
		assert.Equal(t, model.GuildRole(""), role, "Non-member should have empty role")
	}

	// IsGuildAdmin should return false for non-member
	isAdmin, err := guildRepo.IsGuildAdmin(ctx, nonMember.ID, guild.ID)
	require.NoError(t, err)
	assert.False(t, isAdmin)
}
