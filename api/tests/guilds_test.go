// Package tests contains end-to-end acceptance tests for the Saga API.
package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/repository"
	"github.com/forgo/saga/api/internal/service"
	"github.com/forgo/saga/api/internal/testing/fixtures"
	"github.com/forgo/saga/api/internal/testing/testdb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
FEATURE: Guilds
DOMAIN: Community

ACCEPTANCE CRITERIA:
===================

AC-GUILD-001: Create Guild
  GIVEN authenticated user
  WHEN user creates guild with name, description, color
  THEN guild is created
  AND creator is added as member

AC-GUILD-002: Create Guild - Name Validation
  GIVEN authenticated user
  WHEN user creates guild with name > 100 chars
  THEN request fails with validation error

AC-GUILD-003: List User's Guilds
  GIVEN user member of guilds A, B, C
  WHEN user requests their guilds
  THEN all three guilds returned

AC-GUILD-004: Get Guild Details
  GIVEN user is member of guild
  WHEN user requests guild details
  THEN full guild info returned with member count

AC-GUILD-005: Get Guild - Non-Member Access
  GIVEN user is NOT member of private guild
  WHEN user requests guild details
  THEN request fails with not member error

AC-GUILD-006: Update Guild
  GIVEN user is guild member
  WHEN user updates guild name/description
  THEN changes are persisted

AC-GUILD-007: Update Guild - Non-Member
  GIVEN user is NOT guild member
  WHEN user attempts to update guild
  THEN request fails with not member error

AC-GUILD-008: Join Guild
  GIVEN public guild
  WHEN user requests to join
  THEN user is added as member
  AND guild member_count incremented

AC-GUILD-009: Leave Guild
  GIVEN user is member (not sole member)
  WHEN user leaves guild
  THEN user removed from members
  AND member_count decremented

AC-GUILD-010: Cannot Leave as Sole Member
  GIVEN user is only member of guild
  WHEN user attempts to leave
  THEN request fails with cannot leave error

AC-GUILD-011: Max Guilds Per User
  GIVEN user is already member of 10 guilds
  WHEN user tries to create another guild
  THEN request fails with max guilds error

AC-GUILD-012: Max Members Per Guild
  GIVEN guild already has 20 members
  WHEN new user tries to join
  THEN request fails with max members error

AC-GUILD-013: Already Guild Member
  GIVEN user is already member of guild
  WHEN user tries to join again
  THEN request fails with already member error
*/

// createGuildService creates a GuildService instance for testing
func createGuildService(t *testing.T, tdb *testdb.TestDB) *service.GuildService {
	t.Helper()

	guildRepo := repository.NewGuildRepository(tdb.DB)
	memberRepo := repository.NewMemberRepository(tdb.DB)
	userRepo := repository.NewUserRepository(tdb.DB)

	return service.NewGuildService(service.GuildServiceConfig{
		GuildRepo:  guildRepo,
		MemberRepo: memberRepo,
		UserRepo:   userRepo,
	})
}

func TestGuild_Create(t *testing.T) {
	// AC-GUILD-001: Create Guild
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	// Create a user
	user := f.CreateUser(t)

	// Create a guild
	guild, err := guildService.CreateGuild(ctx, user.ID, service.CreateGuildRequest{
		Name:        "Test Guild",
		Description: "A test guild for testing",
		Color:       "#FF5733",
	})

	require.NoError(t, err)
	require.NotNil(t, guild)

	// Verify guild was created with correct data
	assert.NotEmpty(t, guild.ID)
	assert.Equal(t, "Test Guild", guild.Name)
	assert.Equal(t, "A test guild for testing", guild.Description)
	assert.Equal(t, "#FF5733", guild.Color)
	assert.Equal(t, model.GuildVisibilityPrivate, guild.Visibility) // Default visibility

	// Verify creator is a member
	isMember, err := guildService.IsMember(ctx, user.ID, guild.ID)
	require.NoError(t, err)
	assert.True(t, isMember, "Creator should be a member of the guild")

	// Verify member count is 1
	memberCount, err := guildService.GetMemberCount(ctx, guild.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, memberCount, "Guild should have 1 member")
}

func TestGuild_CreatePublic(t *testing.T) {
	// AC-GUILD-001 (variation): Create public guild
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	guild, err := guildService.CreateGuild(ctx, user.ID, service.CreateGuildRequest{
		Name:       "Public Guild",
		Visibility: model.GuildVisibilityPublic,
	})

	require.NoError(t, err)
	assert.Equal(t, model.GuildVisibilityPublic, guild.Visibility)
}

func TestGuild_CreateNameValidation(t *testing.T) {
	// AC-GUILD-002: Create Guild - Name Validation
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	tests := []struct {
		name      string
		guildName string
		wantErr   error
	}{
		{
			name:      "empty name",
			guildName: "",
			wantErr:   service.ErrGuildNameRequired,
		},
		{
			name:      "whitespace only name",
			guildName: "   ",
			wantErr:   service.ErrGuildNameRequired,
		},
		{
			name:      "name too long",
			guildName: strings.Repeat("a", model.MaxGuildNameLength+1),
			wantErr:   service.ErrGuildNameTooLong,
		},
		{
			name:      "exactly max length is valid",
			guildName: strings.Repeat("a", model.MaxGuildNameLength),
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := guildService.CreateGuild(ctx, user.ID, service.CreateGuildRequest{
				Name: tt.guildName,
			})

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGuild_CreateDescriptionValidation(t *testing.T) {
	// AC-GUILD-002 (variation): Description validation
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Description too long should fail
	_, err := guildService.CreateGuild(ctx, user.ID, service.CreateGuildRequest{
		Name:        "Valid Name",
		Description: strings.Repeat("a", model.MaxGuildDescLength+1),
	})

	require.ErrorIs(t, err, service.ErrGuildDescTooLong)
}

func TestGuild_ListUserGuilds(t *testing.T) {
	// AC-GUILD-003: List User's Guilds
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Create multiple guilds
	guild1, err := guildService.CreateGuild(ctx, user.ID, service.CreateGuildRequest{Name: "Guild One"})
	require.NoError(t, err)

	guild2, err := guildService.CreateGuild(ctx, user.ID, service.CreateGuildRequest{Name: "Guild Two"})
	require.NoError(t, err)

	guild3, err := guildService.CreateGuild(ctx, user.ID, service.CreateGuildRequest{Name: "Guild Three"})
	require.NoError(t, err)

	// List user's guilds
	guilds, err := guildService.ListUserGuilds(ctx, user.ID)

	require.NoError(t, err)
	require.Len(t, guilds, 3)

	// Verify all guilds are present
	guildIDs := make(map[string]bool)
	for _, g := range guilds {
		guildIDs[g.ID] = true
	}

	assert.True(t, guildIDs[guild1.ID])
	assert.True(t, guildIDs[guild2.ID])
	assert.True(t, guildIDs[guild3.ID])
}

func TestGuild_ListUserGuildsEmpty(t *testing.T) {
	// AC-GUILD-003 (variation): User with no guilds
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// List guilds for user with no guilds
	guilds, err := guildService.ListUserGuilds(ctx, user.ID)

	require.NoError(t, err)
	assert.Empty(t, guilds)
}

func TestGuild_GetGuildDetails(t *testing.T) {
	// AC-GUILD-004: Get Guild Details
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Create a guild
	created, err := guildService.CreateGuild(ctx, user.ID, service.CreateGuildRequest{
		Name:        "My Guild",
		Description: "Guild description",
		Color:       "#123456",
	})
	require.NoError(t, err)

	// Get guild details
	guild, err := guildService.GetGuild(ctx, user.ID, created.ID)

	require.NoError(t, err)
	require.NotNil(t, guild)

	assert.Equal(t, created.ID, guild.ID)
	assert.Equal(t, "My Guild", guild.Name)
	assert.Equal(t, "Guild description", guild.Description)
	assert.Equal(t, "#123456", guild.Color)
}

func TestGuild_GetGuildWithMembers(t *testing.T) {
	// AC-GUILD-004 (variation): Get guild with member list
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Create a guild
	created, err := guildService.CreateGuild(ctx, user.ID, service.CreateGuildRequest{
		Name: "Guild With Members",
	})
	require.NoError(t, err)

	// Get guild with members
	guildData, err := guildService.GetGuildWithMembers(ctx, user.ID, created.ID)

	require.NoError(t, err)
	require.NotNil(t, guildData)

	assert.Equal(t, created.ID, guildData.Guild.ID)
	assert.Len(t, guildData.Members, 1) // Creator should be the only member
}

func TestGuild_GetGuildNotFound(t *testing.T) {
	// AC-GUILD-004 (error): Guild not found
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Try to get non-existent guild
	_, err := guildService.GetGuild(ctx, user.ID, "guild:nonexistent")

	require.ErrorIs(t, err, service.ErrGuildNotFound)
}

func TestGuild_GetGuildNonMemberAccess(t *testing.T) {
	// AC-GUILD-005: Get Guild - Non-Member Access
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	owner := f.CreateUser(t)
	nonMember := f.CreateUser(t)

	// Create a private guild
	guild, err := guildService.CreateGuild(ctx, owner.ID, service.CreateGuildRequest{
		Name:       "Private Guild",
		Visibility: model.GuildVisibilityPrivate,
	})
	require.NoError(t, err)

	// Non-member tries to access
	_, err = guildService.GetGuild(ctx, nonMember.ID, guild.ID)

	require.ErrorIs(t, err, service.ErrNotGuildMember)
}

func TestGuild_GetPublicGuildNonMemberAccess(t *testing.T) {
	// AC-GUILD-005 (variation): Public guilds are accessible to non-members
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	owner := f.CreateUser(t)
	nonMember := f.CreateUser(t)

	// Create a public guild
	guild, err := guildService.CreateGuild(ctx, owner.ID, service.CreateGuildRequest{
		Name:       "Public Guild",
		Visibility: model.GuildVisibilityPublic,
	})
	require.NoError(t, err)

	// Non-member can access public guild
	retrieved, err := guildService.GetGuild(ctx, nonMember.ID, guild.ID)

	require.NoError(t, err)
	assert.Equal(t, guild.ID, retrieved.ID)
}

func TestGuild_Update(t *testing.T) {
	// AC-GUILD-006: Update Guild
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Create a guild
	guild, err := guildService.CreateGuild(ctx, user.ID, service.CreateGuildRequest{
		Name:        "Original Name",
		Description: "Original description",
	})
	require.NoError(t, err)

	// Update the guild
	newName := "Updated Name"
	newDesc := "Updated description"
	updated, err := guildService.UpdateGuild(ctx, user.ID, guild.ID, service.UpdateGuildRequest{
		Name:        &newName,
		Description: &newDesc,
	})

	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated description", updated.Description)

	// Verify changes persisted
	retrieved, err := guildService.GetGuild(ctx, user.ID, guild.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", retrieved.Name)
}

func TestGuild_UpdatePartial(t *testing.T) {
	// AC-GUILD-006 (variation): Partial update only changes specified fields
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Create a guild
	guild, err := guildService.CreateGuild(ctx, user.ID, service.CreateGuildRequest{
		Name:        "Original Name",
		Description: "Original description",
		Color:       "#FF0000",
	})
	require.NoError(t, err)

	// Update only name
	newName := "New Name Only"
	updated, err := guildService.UpdateGuild(ctx, user.ID, guild.ID, service.UpdateGuildRequest{
		Name: &newName,
	})

	require.NoError(t, err)

	assert.Equal(t, "New Name Only", updated.Name)
	assert.Equal(t, "Original description", updated.Description) // Unchanged
	assert.Equal(t, "#FF0000", updated.Color)                    // Unchanged
}

func TestGuild_UpdateNonMember(t *testing.T) {
	// AC-GUILD-007: Update Guild - Non-Member
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	owner := f.CreateUser(t)
	nonMember := f.CreateUser(t)

	// Create a guild
	guild, err := guildService.CreateGuild(ctx, owner.ID, service.CreateGuildRequest{
		Name: "Owner's Guild",
	})
	require.NoError(t, err)

	// Non-member tries to update
	newName := "Hacked Name"
	_, err = guildService.UpdateGuild(ctx, nonMember.ID, guild.ID, service.UpdateGuildRequest{
		Name: &newName,
	})

	require.ErrorIs(t, err, service.ErrNotGuildMember)

	// Verify name wasn't changed
	retrieved, err := guildService.GetGuild(ctx, owner.ID, guild.ID)
	require.NoError(t, err)
	assert.Equal(t, "Owner's Guild", retrieved.Name)
}

func TestGuild_JoinPublicGuild(t *testing.T) {
	// AC-GUILD-008: Join Guild
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	owner := f.CreateUser(t)
	joiner := f.CreateUser(t)

	// Create a public guild
	guild, err := guildService.CreateGuild(ctx, owner.ID, service.CreateGuildRequest{
		Name:       "Public Guild",
		Visibility: model.GuildVisibilityPublic,
	})
	require.NoError(t, err)

	// Initial member count
	initialCount, err := guildService.GetMemberCount(ctx, guild.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, initialCount)

	// User joins guild
	err = guildService.JoinGuild(ctx, joiner.ID, guild.ID)

	require.NoError(t, err)

	// Verify membership
	isMember, err := guildService.IsMember(ctx, joiner.ID, guild.ID)
	require.NoError(t, err)
	assert.True(t, isMember)

	// Verify member count increased
	newCount, err := guildService.GetMemberCount(ctx, guild.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, newCount)
}

func TestGuild_LeaveGuild(t *testing.T) {
	// AC-GUILD-009: Leave Guild
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	owner := f.CreateUser(t)
	member := f.CreateUser(t)

	// Create a public guild
	guild, err := guildService.CreateGuild(ctx, owner.ID, service.CreateGuildRequest{
		Name:       "Guild To Leave",
		Visibility: model.GuildVisibilityPublic,
	})
	require.NoError(t, err)

	// Second user joins
	err = guildService.JoinGuild(ctx, member.ID, guild.ID)
	require.NoError(t, err)

	// Verify 2 members
	count, err := guildService.GetMemberCount(ctx, guild.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// Second user leaves
	err = guildService.LeaveGuild(ctx, member.ID, guild.ID)

	require.NoError(t, err)

	// Verify no longer a member
	isMember, err := guildService.IsMember(ctx, member.ID, guild.ID)
	require.NoError(t, err)
	assert.False(t, isMember)

	// Verify member count decreased
	count, err = guildService.GetMemberCount(ctx, guild.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestGuild_CannotLeaveSoleMember(t *testing.T) {
	// AC-GUILD-010: Cannot Leave as Sole Member
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Create a guild (user is sole member)
	guild, err := guildService.CreateGuild(ctx, user.ID, service.CreateGuildRequest{
		Name: "Solo Guild",
	})
	require.NoError(t, err)

	// Verify only 1 member
	count, err := guildService.GetMemberCount(ctx, guild.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Try to leave
	err = guildService.LeaveGuild(ctx, user.ID, guild.ID)

	require.ErrorIs(t, err, service.ErrCannotLeaveSoleMember)

	// Verify still a member
	isMember, err := guildService.IsMember(ctx, user.ID, guild.ID)
	require.NoError(t, err)
	assert.True(t, isMember)
}

func TestGuild_LeaveNonMember(t *testing.T) {
	// AC-GUILD-009 (error): Cannot leave guild you're not a member of
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	owner := f.CreateUser(t)
	nonMember := f.CreateUser(t)

	// Create a guild
	guild, err := guildService.CreateGuild(ctx, owner.ID, service.CreateGuildRequest{
		Name: "Not Their Guild",
	})
	require.NoError(t, err)

	// Non-member tries to leave
	err = guildService.LeaveGuild(ctx, nonMember.ID, guild.ID)

	require.ErrorIs(t, err, service.ErrNotGuildMember)
}

func TestGuild_MaxGuildsPerUser(t *testing.T) {
	// AC-GUILD-011: Max Guilds Per User
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Create maximum number of guilds
	for i := 0; i < model.MaxGuildsPerUser; i++ {
		_, err := guildService.CreateGuild(ctx, user.ID, service.CreateGuildRequest{
			Name: strings.Repeat("Guild ", 1) + string(rune('A'+i)),
		})
		require.NoError(t, err)
	}

	// Try to create one more
	_, err := guildService.CreateGuild(ctx, user.ID, service.CreateGuildRequest{
		Name: "One Too Many",
	})

	require.ErrorIs(t, err, service.ErrMaxGuildsReached)
}

func TestGuild_MaxMembersPerGuild(t *testing.T) {
	// AC-GUILD-012: Max Members Per Guild
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	// Create a public guild
	owner := f.CreateUser(t)
	guild, err := guildService.CreateGuild(ctx, owner.ID, service.CreateGuildRequest{
		Name:       "Full Guild",
		Visibility: model.GuildVisibilityPublic,
	})
	require.NoError(t, err)

	// Add members up to limit (owner is already 1)
	for i := 1; i < model.MaxMembersPerGuild; i++ {
		member := f.CreateUser(t)
		err := guildService.JoinGuild(ctx, member.ID, guild.ID)
		require.NoError(t, err)
	}

	// Verify at max
	count, err := guildService.GetMemberCount(ctx, guild.ID)
	require.NoError(t, err)
	assert.Equal(t, model.MaxMembersPerGuild, count)

	// Try to add one more
	extraUser := f.CreateUser(t)
	err = guildService.JoinGuild(ctx, extraUser.ID, guild.ID)

	require.ErrorIs(t, err, service.ErrMaxMembersReached)
}

func TestGuild_AlreadyMember(t *testing.T) {
	// AC-GUILD-013: Already Guild Member
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	owner := f.CreateUser(t)
	member := f.CreateUser(t)

	// Create a public guild
	guild, err := guildService.CreateGuild(ctx, owner.ID, service.CreateGuildRequest{
		Name:       "Join Once Guild",
		Visibility: model.GuildVisibilityPublic,
	})
	require.NoError(t, err)

	// Member joins
	err = guildService.JoinGuild(ctx, member.ID, guild.ID)
	require.NoError(t, err)

	// Member tries to join again
	err = guildService.JoinGuild(ctx, member.ID, guild.ID)

	require.ErrorIs(t, err, service.ErrAlreadyGuildMember)
}

func TestGuild_DeleteGuild(t *testing.T) {
	// Additional test: Delete guild
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Create a guild
	guild, err := guildService.CreateGuild(ctx, user.ID, service.CreateGuildRequest{
		Name: "Guild To Delete",
	})
	require.NoError(t, err)

	// Delete the guild
	err = guildService.DeleteGuild(ctx, user.ID, guild.ID)

	require.NoError(t, err)

	// Verify guild no longer exists
	_, err = guildService.GetGuild(ctx, user.ID, guild.ID)
	require.ErrorIs(t, err, service.ErrGuildNotFound)

	// Verify guild not in user's list
	guilds, err := guildService.ListUserGuilds(ctx, user.ID)
	require.NoError(t, err)
	assert.Empty(t, guilds)
}

func TestGuild_DeleteGuildNonMember(t *testing.T) {
	// Additional test: Non-member cannot delete guild
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	guildService := createGuildService(t, tdb)
	ctx := context.Background()

	owner := f.CreateUser(t)
	nonMember := f.CreateUser(t)

	// Create a guild
	guild, err := guildService.CreateGuild(ctx, owner.ID, service.CreateGuildRequest{
		Name: "Protected Guild",
	})
	require.NoError(t, err)

	// Non-member tries to delete
	err = guildService.DeleteGuild(ctx, nonMember.ID, guild.ID)

	require.ErrorIs(t, err, service.ErrNotGuildMember)

	// Verify guild still exists
	retrieved, err := guildService.GetGuild(ctx, owner.ID, guild.ID)
	require.NoError(t, err)
	assert.Equal(t, guild.ID, retrieved.ID)
}
