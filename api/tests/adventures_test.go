package tests

/*
FEATURE: Adventures
DOMAIN: Multi-day, Multi-location Coordination

ACCEPTANCE CRITERIA:
===================

AC-ADV-001: Create Guild Adventure
  GIVEN user is guild member
  WHEN user creates adventure with organizer_type=guild
  THEN adventure created with guild ownership
  AND creator is auto-admitted

AC-ADV-002: Create User Adventure
  GIVEN authenticated user
  WHEN user creates adventure with organizer_type=user
  THEN adventure created with user ownership

AC-ADV-003: Organizer Type Immutable
  GIVEN guild adventure exists
  WHEN attempting to change organizer_type to user
  THEN update only applies to allowed fields

AC-ADV-004: Request Admission
  GIVEN public adventure
  WHEN user requests admission
  THEN admission request created with status=requested

AC-ADV-005: Already Admitted Cannot Request Again
  GIVEN user already admitted to adventure
  WHEN user requests admission again
  THEN fails with 409 Conflict

AC-ADV-006: Approve Admission
  GIVEN pending admission request
  WHEN organizer approves
  THEN status changes to admitted

AC-ADV-007: Reject Admission
  GIVEN pending admission request
  WHEN organizer rejects with reason
  THEN status changes to rejected
  AND reason stored

AC-ADV-008: Invite User
  GIVEN user is adventure organizer
  WHEN organizer invites user
  THEN user immediately admitted

AC-ADV-009: Withdraw Admission Request
  GIVEN pending admission request
  WHEN requester withdraws
  THEN request deleted

AC-ADV-010: Transfer Adventure - Guild
  GIVEN guild adventure
  WHEN organizer transfers to another guild member
  THEN organizer_user_id updated

AC-ADV-011: Transfer Adventure - Non-Member
  GIVEN guild adventure
  WHEN organizer transfers to non-guild-member
  THEN fails with 400 Bad Request

AC-ADV-012: Unfreeze Adventure
  GIVEN frozen adventure
  WHEN guild member unfreezes with new organizer
  THEN status changes from frozen to planning
  AND organizer_user_id updated

AC-ADV-013: Cannot Withdraw After Admitted
  GIVEN user already admitted
  WHEN user tries to withdraw
  THEN fails with 400 Bad Request

AC-ADV-014: Get Adventure
  GIVEN adventure exists
  WHEN user requests adventure by ID
  THEN adventure details returned

AC-ADV-015: Adventure Not Found
  GIVEN adventure does not exist
  WHEN user requests adventure
  THEN fails with 404 Not Found
*/

import (
	"context"
	"testing"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/repository"
	"github.com/forgo/saga/api/internal/service"
	"github.com/forgo/saga/api/internal/testing/fixtures"
	"github.com/forgo/saga/api/internal/testing/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createAdventureService(t *testing.T, tdb *testdb.TestDB) *service.AdventureService {
	adventureRepo := repository.NewAdventureRepository(tdb.DB)
	admissionRepo := repository.NewAdventureAdmissionRepository(tdb.DB)
	guildRepo := repository.NewGuildRepository(tdb.DB)

	return service.NewAdventureService(service.AdventureServiceConfig{
		AdventureRepo: adventureRepo,
		AdmissionRepo: admissionRepo,
		GuildRepo:     guildRepo,
	})
}

func TestAdventure_CreateGuildAdventure(t *testing.T) {
	// AC-ADV-001: Create Guild Adventure
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	// Create user and guild
	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create guild adventure
	orgType := "guild"
	adventure, err := adventureService.Create(ctx, user.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		GuildID:       &guild.ID,
		Title:         "Epic Road Trip",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, adventure.ID)
	assert.Equal(t, model.AdventureOrganizerGuild, adventure.OrganizerType)
	assert.Contains(t, adventure.OrganizerID, "guild:")
	assert.Equal(t, user.ID, adventure.OrganizerUserID)
	assert.Equal(t, "Epic Road Trip", adventure.Title)
	assert.Equal(t, model.AdventureStatusIdea, adventure.Status)

	// Verify creator is auto-admitted
	isAdmitted, err := adventureService.IsAdmitted(ctx, adventure.ID, user.ID)
	require.NoError(t, err)
	assert.True(t, isAdmitted)
}

func TestAdventure_CreateUserAdventure(t *testing.T) {
	// AC-ADV-002: Create User Adventure
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Create user-organized adventure (no guild)
	orgType := "user"
	adventure, err := adventureService.Create(ctx, user.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		Title:         "Solo Backpacking Adventure",
		StartDate:     "2025-07-01T00:00:00Z",
		EndDate:       "2025-07-10T00:00:00Z",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, adventure.ID)
	assert.Equal(t, model.AdventureOrganizerUser, adventure.OrganizerType)
	assert.Contains(t, adventure.OrganizerID, "user:")
	assert.Equal(t, user.ID, adventure.OrganizerUserID)
}

func TestAdventure_CreateGuildAdventureNonMember(t *testing.T) {
	// GIVEN user is NOT a guild member
	// WHEN user tries to create guild adventure
	// THEN fails with 403 Forbidden
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	guildOwner := f.CreateUser(t)
	guild := f.CreateGuild(t, guildOwner)
	nonMember := f.CreateUser(t)

	orgType := "guild"
	_, err := adventureService.Create(ctx, nonMember.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		GuildID:       &guild.ID,
		Title:         "Should Fail",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestAdventure_RequestAdmission(t *testing.T) {
	// AC-ADV-004: Request Admission
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	organizer := f.CreateUser(t)
	requester := f.CreateUser(t)
	guild := f.CreateGuild(t, organizer)

	// Create adventure
	orgType := "guild"
	adventure, _ := adventureService.Create(ctx, organizer.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		GuildID:       &guild.ID,
		Title:         "Open Adventure",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	// Request admission
	admission, err := adventureService.RequestAdmission(ctx, adventure.ID, requester.ID, &model.RequestAdmissionRequest{})

	require.NoError(t, err)
	assert.NotEmpty(t, admission.ID)
	assert.Equal(t, model.AdmissionStatusRequested, admission.Status)
	assert.Equal(t, model.AdmissionRequestedBySelf, admission.RequestedBy)
	assert.Equal(t, requester.ID, admission.UserID)
}

func TestAdventure_AlreadyAdmittedCannotRequestAgain(t *testing.T) {
	// AC-ADV-005: Already Admitted Cannot Request Again
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	organizer := f.CreateUser(t)
	guild := f.CreateGuild(t, organizer)

	orgType := "guild"
	adventure, _ := adventureService.Create(ctx, organizer.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		GuildID:       &guild.ID,
		Title:         "Test Adventure",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	// Organizer is auto-admitted, try to request again
	_, err := adventureService.RequestAdmission(ctx, adventure.ID, organizer.ID, &model.RequestAdmissionRequest{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already admitted")
}

func TestAdventure_PendingCannotRequestAgain(t *testing.T) {
	// Duplicate request should fail
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	organizer := f.CreateUser(t)
	requester := f.CreateUser(t)
	guild := f.CreateGuild(t, organizer)

	orgType := "guild"
	adventure, _ := adventureService.Create(ctx, organizer.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		GuildID:       &guild.ID,
		Title:         "Test Adventure",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	// First request
	_, _ = adventureService.RequestAdmission(ctx, adventure.ID, requester.ID, &model.RequestAdmissionRequest{})

	// Second request should fail
	_, err := adventureService.RequestAdmission(ctx, adventure.ID, requester.ID, &model.RequestAdmissionRequest{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "pending")
}

func TestAdventure_ApproveAdmission(t *testing.T) {
	// AC-ADV-006: Approve Admission
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	organizer := f.CreateUser(t)
	requester := f.CreateUser(t)
	guild := f.CreateGuild(t, organizer)

	orgType := "guild"
	adventure, _ := adventureService.Create(ctx, organizer.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		GuildID:       &guild.ID,
		Title:         "Test Adventure",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	// Request admission
	_, _ = adventureService.RequestAdmission(ctx, adventure.ID, requester.ID, &model.RequestAdmissionRequest{})

	// Approve
	admitted, err := adventureService.RespondToAdmission(ctx, adventure.ID, organizer.ID, requester.ID, &model.RespondToAdmissionRequest{
		Admit: true,
	})

	require.NoError(t, err)
	assert.Equal(t, model.AdmissionStatusAdmitted, admitted.Status)

	// Verify they are now admitted
	isAdmitted, _ := adventureService.IsAdmitted(ctx, adventure.ID, requester.ID)
	assert.True(t, isAdmitted)
}

func TestAdventure_RejectAdmission(t *testing.T) {
	// AC-ADV-007: Reject Admission
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	organizer := f.CreateUser(t)
	requester := f.CreateUser(t)
	guild := f.CreateGuild(t, organizer)

	orgType := "guild"
	adventure, _ := adventureService.Create(ctx, organizer.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		GuildID:       &guild.ID,
		Title:         "Exclusive Adventure",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	// Request admission
	_, _ = adventureService.RequestAdmission(ctx, adventure.ID, requester.ID, &model.RequestAdmissionRequest{})

	// Reject with reason
	reason := "Application incomplete"
	rejected, err := adventureService.RespondToAdmission(ctx, adventure.ID, organizer.ID, requester.ID, &model.RespondToAdmissionRequest{
		Admit:           false,
		RejectionReason: &reason,
	})

	require.NoError(t, err)
	assert.Equal(t, model.AdmissionStatusRejected, rejected.Status)
	assert.NotNil(t, rejected.RejectionReason)
	assert.Equal(t, "Application incomplete", *rejected.RejectionReason)

	// Verify they are not admitted
	isAdmitted, _ := adventureService.IsAdmitted(ctx, adventure.ID, requester.ID)
	assert.False(t, isAdmitted)
}

func TestAdventure_RejectRequiresReason(t *testing.T) {
	// Rejection without reason should fail validation
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	organizer := f.CreateUser(t)
	requester := f.CreateUser(t)
	guild := f.CreateGuild(t, organizer)

	orgType := "guild"
	adventure, _ := adventureService.Create(ctx, organizer.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		GuildID:       &guild.ID,
		Title:         "Test Adventure",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	_, _ = adventureService.RequestAdmission(ctx, adventure.ID, requester.ID, &model.RequestAdmissionRequest{})

	// Reject without reason
	_, err := adventureService.RespondToAdmission(ctx, adventure.ID, organizer.ID, requester.ID, &model.RespondToAdmissionRequest{
		Admit: false,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rejection_reason")
}

func TestAdventure_InviteUser(t *testing.T) {
	// AC-ADV-008: Invite User
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	organizer := f.CreateUser(t)
	invitee := f.CreateUser(t)
	guild := f.CreateGuild(t, organizer)

	orgType := "guild"
	adventure, _ := adventureService.Create(ctx, organizer.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		GuildID:       &guild.ID,
		Title:         "Private Adventure",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	// Invite user
	admission, err := adventureService.InviteToAdventure(ctx, adventure.ID, organizer.ID, &model.InviteToAdventureRequest{
		UserID: invitee.ID,
	})

	require.NoError(t, err)
	assert.Equal(t, model.AdmissionStatusAdmitted, admission.Status) // Auto-admitted
	assert.Equal(t, model.AdmissionRequestedByInvited, admission.RequestedBy)
	assert.NotNil(t, admission.InvitedByID)
	assert.Equal(t, organizer.ID, *admission.InvitedByID)

	// Verify they are admitted
	isAdmitted, _ := adventureService.IsAdmitted(ctx, adventure.ID, invitee.ID)
	assert.True(t, isAdmitted)
}

func TestAdventure_WithdrawAdmissionRequest(t *testing.T) {
	// AC-ADV-009: Withdraw Admission Request
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	organizer := f.CreateUser(t)
	requester := f.CreateUser(t)
	guild := f.CreateGuild(t, organizer)

	orgType := "guild"
	adventure, _ := adventureService.Create(ctx, organizer.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		GuildID:       &guild.ID,
		Title:         "Test Adventure",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	// Request admission
	_, _ = adventureService.RequestAdmission(ctx, adventure.ID, requester.ID, &model.RequestAdmissionRequest{})

	// Withdraw
	err := adventureService.WithdrawAdmission(ctx, adventure.ID, requester.ID)
	require.NoError(t, err)

	// Verify no admission exists
	_, err = adventureService.GetAdmission(ctx, adventure.ID, requester.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAdventure_CannotWithdrawAfterAdmitted(t *testing.T) {
	// AC-ADV-013: Cannot Withdraw After Admitted
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	organizer := f.CreateUser(t)
	guild := f.CreateGuild(t, organizer)

	orgType := "guild"
	adventure, _ := adventureService.Create(ctx, organizer.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		GuildID:       &guild.ID,
		Title:         "Test Adventure",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	// Organizer is auto-admitted, try to withdraw
	err := adventureService.WithdrawAdmission(ctx, adventure.ID, organizer.ID)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot withdraw after being admitted")
}

func TestAdventure_TransferGuild(t *testing.T) {
	// AC-ADV-010: Transfer Adventure - Guild
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	organizer := f.CreateUser(t)
	newOrganizer := f.CreateUser(t)
	guild := f.CreateGuild(t, organizer)
	f.AddMemberToGuild(t, newOrganizer, guild)

	orgType := "guild"
	adventure, _ := adventureService.Create(ctx, organizer.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		GuildID:       &guild.ID,
		Title:         "Transferable Adventure",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	// Transfer to new organizer
	updated, err := adventureService.TransferAdventure(ctx, adventure.ID, organizer.ID, &model.TransferAdventureRequest{
		NewOrganizerUserID: newOrganizer.ID,
	})

	require.NoError(t, err)
	assert.Equal(t, newOrganizer.ID, updated.OrganizerUserID)
}

func TestAdventure_TransferNonMember(t *testing.T) {
	// AC-ADV-011: Transfer Adventure - Non-Member
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	organizer := f.CreateUser(t)
	nonMember := f.CreateUser(t)
	guild := f.CreateGuild(t, organizer)

	orgType := "guild"
	adventure, _ := adventureService.Create(ctx, organizer.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		GuildID:       &guild.ID,
		Title:         "Cannot Transfer",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	// Try to transfer to non-guild-member
	_, err := adventureService.TransferAdventure(ctx, adventure.ID, organizer.ID, &model.TransferAdventureRequest{
		NewOrganizerUserID: nonMember.ID,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "guild member")
}

func TestAdventure_UnfreezeAdventure(t *testing.T) {
	// AC-ADV-012: Unfreeze Adventure
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	adventureRepo := repository.NewAdventureRepository(tdb.DB)
	ctx := context.Background()

	organizer := f.CreateUser(t)
	newOrganizer := f.CreateUser(t)
	guild := f.CreateGuild(t, organizer)
	f.AddMemberToGuild(t, newOrganizer, guild)

	orgType := "guild"
	adventure, _ := adventureService.Create(ctx, organizer.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		GuildID:       &guild.ID,
		Title:         "Freezable Adventure",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	// Freeze the adventure directly via repository (simulating organizer leaving guild)
	_, err := adventureRepo.Freeze(ctx, adventure.ID, "Organizer left guild")
	require.NoError(t, err)

	// Verify it's frozen
	frozen, _ := adventureService.GetByID(ctx, adventure.ID)
	assert.Equal(t, model.AdventureStatusFrozen, frozen.Status)
	assert.NotNil(t, frozen.FreezeReason)

	// Unfreeze with new organizer
	unfrozen, err := adventureService.UnfreezeAdventure(ctx, adventure.ID, newOrganizer.ID, &model.UnfreezeAdventureRequest{
		NewOrganizerUserID: newOrganizer.ID,
	})

	require.NoError(t, err)
	assert.Equal(t, model.AdventureStatusPlanning, unfrozen.Status)
	assert.Equal(t, newOrganizer.ID, unfrozen.OrganizerUserID)
}

func TestAdventure_CannotUnfreezeNonFrozen(t *testing.T) {
	// Unfreezing a non-frozen adventure should fail
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	organizer := f.CreateUser(t)
	guild := f.CreateGuild(t, organizer)

	orgType := "guild"
	adventure, _ := adventureService.Create(ctx, organizer.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		GuildID:       &guild.ID,
		Title:         "Normal Adventure",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	// Try to unfreeze non-frozen adventure
	_, err := adventureService.UnfreezeAdventure(ctx, adventure.ID, organizer.ID, &model.UnfreezeAdventureRequest{
		NewOrganizerUserID: organizer.ID,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not frozen")
}

func TestAdventure_GetByID(t *testing.T) {
	// AC-ADV-014: Get Adventure
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	orgType := "user"
	adventure, _ := adventureService.Create(ctx, user.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		Title:         "My Adventure",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	// Get by ID
	found, err := adventureService.GetByID(ctx, adventure.ID)

	require.NoError(t, err)
	assert.Equal(t, adventure.ID, found.ID)
	assert.Equal(t, "My Adventure", found.Title)
}

func TestAdventure_NotFound(t *testing.T) {
	// AC-ADV-015: Adventure Not Found
	tdb := testdb.New(t)
	defer tdb.Close()

	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	_, err := adventureService.GetByID(ctx, "adventure:nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAdventure_InviteAlreadyAdmitted(t *testing.T) {
	// Cannot invite someone who already has an admission record
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	organizer := f.CreateUser(t)
	invitee := f.CreateUser(t)
	guild := f.CreateGuild(t, organizer)

	orgType := "guild"
	adventure, _ := adventureService.Create(ctx, organizer.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		GuildID:       &guild.ID,
		Title:         "Test Adventure",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	// First invite
	_, _ = adventureService.InviteToAdventure(ctx, adventure.ID, organizer.ID, &model.InviteToAdventureRequest{
		UserID: invitee.ID,
	})

	// Second invite should fail
	_, err := adventureService.InviteToAdventure(ctx, adventure.ID, organizer.ID, &model.InviteToAdventureRequest{
		UserID: invitee.ID,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already has admission")
}

func TestAdventure_NonOrganizerCannotInvite(t *testing.T) {
	// Non-organizers cannot invite
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	organizer := f.CreateUser(t)
	randomUser := f.CreateUser(t)
	invitee := f.CreateUser(t)

	orgType := "user"
	adventure, _ := adventureService.Create(ctx, organizer.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		Title:         "Private Adventure",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	// Random user tries to invite
	_, err := adventureService.InviteToAdventure(ctx, adventure.ID, randomUser.ID, &model.InviteToAdventureRequest{
		UserID: invitee.ID,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestAdventure_NonOrganizerCannotApprove(t *testing.T) {
	// Non-organizers cannot approve admission
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	adventureService := createAdventureService(t, tdb)
	ctx := context.Background()

	organizer := f.CreateUser(t)
	randomUser := f.CreateUser(t)
	requester := f.CreateUser(t)

	orgType := "user"
	adventure, _ := adventureService.Create(ctx, organizer.ID, &model.CreateAdventureRequest{
		OrganizerType: &orgType,
		Title:         "Test Adventure",
		StartDate:     "2025-06-01T00:00:00Z",
		EndDate:       "2025-06-15T00:00:00Z",
	})

	// Request admission
	_, _ = adventureService.RequestAdmission(ctx, adventure.ID, requester.ID, &model.RequestAdmissionRequest{})

	// Random user tries to approve
	_, err := adventureService.RespondToAdmission(ctx, adventure.ID, randomUser.ID, requester.ID, &model.RespondToAdmissionRequest{
		Admit: true,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}
