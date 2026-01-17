package tests

import (
	"context"
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/repository"
	"github.com/forgo/saga/api/internal/service"
	"github.com/forgo/saga/api/internal/testing/fixtures"
	"github.com/forgo/saga/api/internal/testing/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
FEATURE: Voting System
DOMAIN: Governance & Decision Making

ACCEPTANCE CRITERIA:
===================

AC-VOTE-001: Create Guild Vote
AC-VOTE-002: Create Global Vote
AC-VOTE-003: Create Global Vote - Non-Admin (skipped - requires sysadmin check)
AC-VOTE-004: Add Vote Options (Draft)
AC-VOTE-005: Cannot Add Options When Open
AC-VOTE-006: Open Vote Manually
AC-VOTE-007: Vote Opens Automatically
AC-VOTE-008: Cast FPTP Ballot
AC-VOTE-009: Cast Ranked Choice Ballot
AC-VOTE-010: Cast Approval Ballot
AC-VOTE-011: Cast Multi-Select Ballot
AC-VOTE-012: Multi-Select Exceeds Max
AC-VOTE-013: Revoting Replaces Previous Ballot
AC-VOTE-014: Ballot Immutability (via database constraints)
AC-VOTE-015: Close Vote Manually
AC-VOTE-016: Cancel Vote
AC-VOTE-017: FPTP Results
AC-VOTE-018: Ranked Choice Results
AC-VOTE-019: Transparent Ballot Ledger
AC-VOTE-020: Guild Vote Membership Required
*/

func createVoteService(t *testing.T, tdb *testdb.TestDB) *service.VoteService {
	voteRepo := repository.NewVoteRepository(tdb.DB)
	userRepo := repository.NewUserRepository(tdb.DB)
	guildRepo := repository.NewGuildRepository(tdb.DB)

	return service.NewVoteService(service.VoteServiceConfig{
		VoteRepo:  voteRepo,
		UserRepo:  userRepo,
		GuildRepo: guildRepo,
	})
}

func createTestVote(t *testing.T, voteService *service.VoteService, userID string, guildID *string, voteType string) *model.Vote {
	ctx := context.Background()
	opensAt := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	closesAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	scopeType := "guild"
	if guildID == nil {
		scopeType = "global"
	}

	vote, err := voteService.Create(ctx, userID, &model.CreateVoteRequest{
		ScopeType: scopeType,
		ScopeID:   guildID,
		Title:     "Test Vote",
		VoteType:  voteType,
		OpensAt:   opensAt,
		ClosesAt:  closesAt,
	})
	require.NoError(t, err)
	return vote
}

func TestVoting_CreateGuildVote(t *testing.T) {
	// AC-VOTE-001: Create Guild Vote
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	opensAt := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	closesAt := time.Now().Add(48 * time.Hour).Format(time.RFC3339)

	vote, err := voteService.Create(ctx, user.ID, &model.CreateVoteRequest{
		ScopeType: "guild",
		ScopeID:   &guild.ID,
		Title:     "Guild Decision",
		VoteType:  "fptp",
		OpensAt:   opensAt,
		ClosesAt:  closesAt,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, vote.ID)
	assert.Equal(t, model.VoteScopeGuild, vote.ScopeType)
	assert.Equal(t, guild.ID, *vote.ScopeID)
	assert.Equal(t, model.VoteStatusDraft, vote.Status)
	assert.Equal(t, user.ID, vote.CreatedBy)
}

func TestVoting_CreateGlobalVote(t *testing.T) {
	// AC-VOTE-002: Create Global Vote
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	opensAt := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	closesAt := time.Now().Add(48 * time.Hour).Format(time.RFC3339)

	vote, err := voteService.Create(ctx, user.ID, &model.CreateVoteRequest{
		ScopeType: "global",
		Title:     "Global Decision",
		VoteType:  "fptp",
		OpensAt:   opensAt,
		ClosesAt:  closesAt,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, vote.ID)
	assert.Equal(t, model.VoteScopeGlobal, vote.ScopeType)
	assert.Nil(t, vote.ScopeID)
	assert.Equal(t, model.VoteStatusDraft, vote.Status)
}

func TestVoting_AddOptionsDraft(t *testing.T) {
	// AC-VOTE-004: Add Vote Options (Draft)
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	vote := createTestVote(t, voteService, user.ID, &guild.ID, "fptp")

	// Add options while in draft status
	opt1, err := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{
		OptionText: "Option A",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, opt1.ID)
	assert.Equal(t, "Option A", opt1.OptionText)

	opt2, err := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{
		OptionText: "Option B",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, opt2.ID)

	// Verify options are retrieved
	details, err := voteService.GetByID(ctx, vote.ID, user.ID)
	require.NoError(t, err)
	assert.Len(t, details.Options, 2)
}

func TestVoting_CannotAddOptionsWhenOpen(t *testing.T) {
	// AC-VOTE-005: Cannot Add Options When Open
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	vote := createTestVote(t, voteService, user.ID, &guild.ID, "fptp")

	// Add initial option
	_, err := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{
		OptionText: "Option A",
	})
	require.NoError(t, err)

	// Open the vote
	err = voteService.Open(ctx, vote.ID, user.ID)
	require.NoError(t, err)

	// Try to add option when open - should fail
	_, err = voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{
		OptionText: "Option B",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "draft")
}

func TestVoting_OpenVoteManually(t *testing.T) {
	// AC-VOTE-006: Open Vote Manually
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	vote := createTestVote(t, voteService, user.ID, &guild.ID, "fptp")

	// Verify starts in draft
	assert.Equal(t, model.VoteStatusDraft, vote.Status)

	// Open manually
	err := voteService.Open(ctx, vote.ID, user.ID)
	require.NoError(t, err)

	// Verify status changed
	details, err := voteService.GetByID(ctx, vote.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, model.VoteStatusOpen, details.Vote.Status)
}

func TestVoting_VoteOpensAutomatically(t *testing.T) {
	// AC-VOTE-007: Vote Opens Automatically
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create vote with opens_at in the past
	opensAt := time.Now().Add(-1 * time.Minute).Format(time.RFC3339)
	closesAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	vote, err := voteService.Create(ctx, user.ID, &model.CreateVoteRequest{
		ScopeType: "guild",
		ScopeID:   &guild.ID,
		Title:     "Auto-Open Vote",
		VoteType:  "fptp",
		OpensAt:   opensAt,
		ClosesAt:  closesAt,
	})
	require.NoError(t, err)
	assert.Equal(t, model.VoteStatusDraft, vote.Status)

	// Run scheduled transitions
	err = voteService.ProcessScheduledTransitions(ctx)
	require.NoError(t, err)

	// Verify vote is now open
	details, err := voteService.GetByID(ctx, vote.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, model.VoteStatusOpen, details.Vote.Status)
}

func TestVoting_CastFPTPBallot(t *testing.T) {
	// AC-VOTE-008: Cast FPTP Ballot
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	voter := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	f.AddMemberToGuild(t, voter, guild)

	vote := createTestVote(t, voteService, user.ID, &guild.ID, "fptp")

	// Add options
	optA, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option A"})
	_, _ = voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option B"})

	// Open vote
	_ = voteService.Open(ctx, vote.ID, user.ID)

	// Cast ballot
	ballot, err := voteService.CastBallot(ctx, vote.ID, voter.ID, &model.CastBallotRequest{
		OptionID: &optA.ID,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, ballot.ID)
	assert.Equal(t, voter.ID, ballot.VoterUserID)
	assert.Equal(t, optA.ID, ballot.BallotData["option_id"])
}

func TestVoting_CastRankedChoiceBallot(t *testing.T) {
	// AC-VOTE-009: Cast Ranked Choice Ballot
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	voter := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	f.AddMemberToGuild(t, voter, guild)

	vote := createTestVote(t, voteService, user.ID, &guild.ID, "ranked_choice")

	// Add options
	optA, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option A"})
	optB, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option B"})
	optC, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option C"})

	// Open vote
	_ = voteService.Open(ctx, vote.ID, user.ID)

	// Cast ranked choice ballot
	ballot, err := voteService.CastBallot(ctx, vote.ID, voter.ID, &model.CastBallotRequest{
		Rankings: []string{optB.ID, optA.ID, optC.ID},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, ballot.ID)
	// Rankings may be []string or []interface{} depending on how it's stored/retrieved
	switch r := ballot.BallotData["rankings"].(type) {
	case []string:
		assert.Len(t, r, 3)
		assert.Equal(t, optB.ID, r[0])
	case []interface{}:
		assert.Len(t, r, 3)
		assert.Equal(t, optB.ID, r[0])
	default:
		t.Fatalf("unexpected rankings type: %T", ballot.BallotData["rankings"])
	}
}

func TestVoting_CastApprovalBallot(t *testing.T) {
	// AC-VOTE-010: Cast Approval Ballot
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	voter := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	f.AddMemberToGuild(t, voter, guild)

	vote := createTestVote(t, voteService, user.ID, &guild.ID, "approval")

	// Add options
	optA, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option A"})
	optB, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option B"})
	_, _ = voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option C"})

	// Open vote
	_ = voteService.Open(ctx, vote.ID, user.ID)

	// Cast approval ballot (approve multiple)
	ballot, err := voteService.CastBallot(ctx, vote.ID, voter.ID, &model.CastBallotRequest{
		SelectedOptions: []string{optA.ID, optB.ID},
	})

	require.NoError(t, err)
	switch a := ballot.BallotData["approved_options"].(type) {
	case []string:
		assert.Len(t, a, 2)
	case []interface{}:
		assert.Len(t, a, 2)
	default:
		t.Fatalf("unexpected approved_options type: %T", ballot.BallotData["approved_options"])
	}
}

func TestVoting_CastMultiSelectBallot(t *testing.T) {
	// AC-VOTE-011: Cast Multi-Select Ballot
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	voter := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	f.AddMemberToGuild(t, voter, guild)

	// Create multi-select vote with max=3
	opensAt := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	closesAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	maxOpts := 3

	vote, err := voteService.Create(ctx, user.ID, &model.CreateVoteRequest{
		ScopeType:            "guild",
		ScopeID:              &guild.ID,
		Title:                "Multi-Select Vote",
		VoteType:             "multi_select",
		OpensAt:              opensAt,
		ClosesAt:             closesAt,
		MaxOptionsSelectable: &maxOpts,
	})
	require.NoError(t, err)

	// Add options
	optA, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option A"})
	optB, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option B"})
	optC, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option C"})

	// Open vote
	_ = voteService.Open(ctx, vote.ID, user.ID)

	// Cast ballot selecting 3 options
	ballot, err := voteService.CastBallot(ctx, vote.ID, voter.ID, &model.CastBallotRequest{
		SelectedOptions: []string{optA.ID, optB.ID, optC.ID},
	})

	require.NoError(t, err)
	switch s := ballot.BallotData["selected_options"].(type) {
	case []string:
		assert.Len(t, s, 3)
	case []interface{}:
		assert.Len(t, s, 3)
	default:
		t.Fatalf("unexpected selected_options type: %T", ballot.BallotData["selected_options"])
	}
}

func TestVoting_MultiSelectExceedsMax(t *testing.T) {
	// AC-VOTE-012: Multi-Select Exceeds Max
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	voter := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	f.AddMemberToGuild(t, voter, guild)

	// Create multi-select vote with max=2
	opensAt := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	closesAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	maxOpts := 2

	vote, err := voteService.Create(ctx, user.ID, &model.CreateVoteRequest{
		ScopeType:            "guild",
		ScopeID:              &guild.ID,
		Title:                "Multi-Select Vote",
		VoteType:             "multi_select",
		OpensAt:              opensAt,
		ClosesAt:             closesAt,
		MaxOptionsSelectable: &maxOpts,
	})
	require.NoError(t, err)

	// Add options
	optA, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option A"})
	optB, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option B"})
	optC, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option C"})

	// Open vote
	_ = voteService.Open(ctx, vote.ID, user.ID)

	// Try to select 3 options when max is 2
	_, err = voteService.CastBallot(ctx, vote.ID, voter.ID, &model.CastBallotRequest{
		SelectedOptions: []string{optA.ID, optB.ID, optC.ID},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many options selected")
}

func TestVoting_RevotingReplacesBallot(t *testing.T) {
	// AC-VOTE-013: Revoting Replaces Previous Ballot
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	voter := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	f.AddMemberToGuild(t, voter, guild)

	vote := createTestVote(t, voteService, user.ID, &guild.ID, "fptp")

	// Add options
	optA, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option A"})
	optB, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option B"})

	// Open vote
	_ = voteService.Open(ctx, vote.ID, user.ID)

	// Cast first ballot
	ballot1, err := voteService.CastBallot(ctx, vote.ID, voter.ID, &model.CastBallotRequest{
		OptionID: &optA.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, optA.ID, ballot1.BallotData["option_id"])

	// Cast second ballot (revote)
	ballot2, err := voteService.CastBallot(ctx, vote.ID, voter.ID, &model.CastBallotRequest{
		OptionID: &optB.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, optB.ID, ballot2.BallotData["option_id"])

	// Verify only one ballot exists
	myBallot, err := voteService.GetMyBallot(ctx, vote.ID, voter.ID)
	require.NoError(t, err)
	assert.Equal(t, optB.ID, myBallot.BallotData["option_id"])
}

func TestVoting_CloseVoteManually(t *testing.T) {
	// AC-VOTE-015: Close Vote Manually
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	vote := createTestVote(t, voteService, user.ID, &guild.ID, "fptp")

	// Open first
	err := voteService.Open(ctx, vote.ID, user.ID)
	require.NoError(t, err)

	// Close manually
	err = voteService.Close(ctx, vote.ID, user.ID)
	require.NoError(t, err)

	// Verify status
	details, err := voteService.GetByID(ctx, vote.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, model.VoteStatusClosed, details.Vote.Status)
}

func TestVoting_CancelVote(t *testing.T) {
	// AC-VOTE-016: Cancel Vote
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	vote := createTestVote(t, voteService, user.ID, &guild.ID, "fptp")

	// Cancel vote (can cancel from draft)
	err := voteService.Cancel(ctx, vote.ID, user.ID)
	require.NoError(t, err)

	// Verify status
	details, err := voteService.GetByID(ctx, vote.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, model.VoteStatusCancelled, details.Vote.Status)
}

func TestVoting_FPTPResults(t *testing.T) {
	// AC-VOTE-017: FPTP Results
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	voter1 := f.CreateUser(t)
	voter2 := f.CreateUser(t)
	voter3 := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	f.AddMemberToGuild(t, voter1, guild)
	f.AddMemberToGuild(t, voter2, guild)
	f.AddMemberToGuild(t, voter3, guild)

	vote := createTestVote(t, voteService, user.ID, &guild.ID, "fptp")

	// Add options
	optA, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option A"})
	optB, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option B"})

	// Open vote
	_ = voteService.Open(ctx, vote.ID, user.ID)

	// Cast ballots: 2 for A, 1 for B
	_, _ = voteService.CastBallot(ctx, vote.ID, voter1.ID, &model.CastBallotRequest{OptionID: &optA.ID})
	_, _ = voteService.CastBallot(ctx, vote.ID, voter2.ID, &model.CastBallotRequest{OptionID: &optA.ID})
	_, _ = voteService.CastBallot(ctx, vote.ID, voter3.ID, &model.CastBallotRequest{OptionID: &optB.ID})

	// Close vote
	_ = voteService.Close(ctx, vote.ID, user.ID)

	// Get results
	results, err := voteService.GetResults(ctx, vote.ID, user.ID)
	require.NoError(t, err)

	assert.Equal(t, 3, results.TotalBallots)
	assert.Equal(t, model.VoteTypeFPTP, results.VoteType)
	assert.NotNil(t, results.Winner)
	assert.Equal(t, optA.ID, *results.Winner)

	// Option A should be winner with 2 votes
	var optAResult *model.OptionResult
	for i := range results.OptionResults {
		if results.OptionResults[i].OptionID == optA.ID {
			optAResult = &results.OptionResults[i]
			break
		}
	}
	require.NotNil(t, optAResult)
	assert.Equal(t, 2, optAResult.VoteCount)
	assert.True(t, optAResult.IsWinner)
	assert.Equal(t, 1, optAResult.Rank)
}

func TestVoting_RankedChoiceResults(t *testing.T) {
	// AC-VOTE-018: Ranked Choice Results
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	voter1 := f.CreateUser(t)
	voter2 := f.CreateUser(t)
	voter3 := f.CreateUser(t)
	voter4 := f.CreateUser(t)
	voter5 := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	f.AddMemberToGuild(t, voter1, guild)
	f.AddMemberToGuild(t, voter2, guild)
	f.AddMemberToGuild(t, voter3, guild)
	f.AddMemberToGuild(t, voter4, guild)
	f.AddMemberToGuild(t, voter5, guild)

	vote := createTestVote(t, voteService, user.ID, &guild.ID, "ranked_choice")

	// Add 3 options
	optA, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Alice"})
	optB, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Bob"})
	optC, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Carol"})

	// Open vote
	_ = voteService.Open(ctx, vote.ID, user.ID)

	// Cast ranked choice ballots
	// Voter 1: A > B > C
	// Voter 2: A > C > B
	// Voter 3: B > C > A
	// Voter 4: C > B > A
	// Voter 5: C > B > A
	// Round 1: A=2, B=1, C=2 -> B eliminated
	// Round 2: A=2, C=3 -> C wins
	_, _ = voteService.CastBallot(ctx, vote.ID, voter1.ID, &model.CastBallotRequest{Rankings: []string{optA.ID, optB.ID, optC.ID}})
	_, _ = voteService.CastBallot(ctx, vote.ID, voter2.ID, &model.CastBallotRequest{Rankings: []string{optA.ID, optC.ID, optB.ID}})
	_, _ = voteService.CastBallot(ctx, vote.ID, voter3.ID, &model.CastBallotRequest{Rankings: []string{optB.ID, optC.ID, optA.ID}})
	_, _ = voteService.CastBallot(ctx, vote.ID, voter4.ID, &model.CastBallotRequest{Rankings: []string{optC.ID, optB.ID, optA.ID}})
	_, _ = voteService.CastBallot(ctx, vote.ID, voter5.ID, &model.CastBallotRequest{Rankings: []string{optC.ID, optB.ID, optA.ID}})

	// Close vote
	voteService.Close(ctx, vote.ID, user.ID)

	// Get results
	results, err := voteService.GetResults(ctx, vote.ID, user.ID)
	require.NoError(t, err)

	assert.Equal(t, 5, results.TotalBallots)
	assert.Equal(t, model.VoteTypeRankedChoice, results.VoteType)
	assert.NotNil(t, results.Winner)
	assert.Equal(t, optC.ID, *results.Winner)
	assert.NotEmpty(t, results.RoundDetails)

	// Verify Carol is marked as winner
	var carolResult *model.OptionResult
	for i := range results.OptionResults {
		if results.OptionResults[i].OptionID == optC.ID {
			carolResult = &results.OptionResults[i]
			break
		}
	}
	require.NotNil(t, carolResult)
	assert.True(t, carolResult.IsWinner)
}

func TestVoting_TransparentBallotLedger(t *testing.T) {
	// AC-VOTE-019: Transparent Ballot Ledger
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	voter1 := f.CreateUser(t)
	voter2 := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	f.AddMemberToGuild(t, voter1, guild)
	f.AddMemberToGuild(t, voter2, guild)

	// Create vote with public results visibility
	opensAt := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	closesAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	visibility := "after_close"

	vote, err := voteService.Create(ctx, user.ID, &model.CreateVoteRequest{
		ScopeType:         "guild",
		ScopeID:           &guild.ID,
		Title:             "Transparent Vote",
		VoteType:          "fptp",
		OpensAt:           opensAt,
		ClosesAt:          closesAt,
		ResultsVisibility: &visibility,
	})
	require.NoError(t, err)

	// Add options
	optA, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option A"})

	// Open vote
	_ = voteService.Open(ctx, vote.ID, user.ID)

	// Cast ballots
	_, _ = voteService.CastBallot(ctx, vote.ID, voter1.ID, &model.CastBallotRequest{OptionID: &optA.ID})
	_, _ = voteService.CastBallot(ctx, vote.ID, voter2.ID, &model.CastBallotRequest{OptionID: &optA.ID})

	// Results not visible while open
	_, err = voteService.GetBallots(ctx, vote.ID, voter1.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not yet visible")

	// Close vote
	voteService.Close(ctx, vote.ID, user.ID)

	// Now ballots are visible
	ballots, err := voteService.GetBallots(ctx, vote.ID, voter1.ID)
	require.NoError(t, err)
	assert.Len(t, ballots, 2)

	// Verify voter info is included
	for _, ballot := range ballots {
		assert.NotEmpty(t, ballot.VoterUserID)
	}
}

func TestVoting_GuildVoteMembershipRequired(t *testing.T) {
	// AC-VOTE-020: Guild Vote Membership Required
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	nonMember := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	vote := createTestVote(t, voteService, user.ID, &guild.ID, "fptp")

	// Add options
	optA, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option A"})

	// Open vote
	_ = voteService.Open(ctx, vote.ID, user.ID)

	// Non-member tries to vote
	_, err := voteService.CastBallot(ctx, vote.ID, nonMember.ID, &model.CastBallotRequest{
		OptionID: &optA.ID,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot vote")
}

func TestVoting_CannotVoteWhenClosed(t *testing.T) {
	// Additional test: Cannot vote when vote is closed
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	voter := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	f.AddMemberToGuild(t, voter, guild)

	vote := createTestVote(t, voteService, user.ID, &guild.ID, "fptp")

	// Add options
	optA, _ := voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option A"})

	// Open and immediately close
	_ = voteService.Open(ctx, vote.ID, user.ID)
	voteService.Close(ctx, vote.ID, user.ID)

	// Try to vote
	_, err := voteService.CastBallot(ctx, vote.ID, voter.ID, &model.CastBallotRequest{
		OptionID: &optA.ID,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not open")
}

func TestVoting_Abstain(t *testing.T) {
	// Additional test: Abstaining from a vote
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	voteService := createVoteService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)
	voter := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	f.AddMemberToGuild(t, voter, guild)

	// Create vote that allows abstaining
	opensAt := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	closesAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	vote, err := voteService.Create(ctx, user.ID, &model.CreateVoteRequest{
		ScopeType:    "guild",
		ScopeID:      &guild.ID,
		Title:        "Vote with Abstain",
		VoteType:     "fptp",
		OpensAt:      opensAt,
		ClosesAt:     closesAt,
		AllowAbstain: true,
	})
	require.NoError(t, err)

	// Add options
	_, _ = voteService.AddOption(ctx, vote.ID, user.ID, &model.CreateVoteOptionRequest{OptionText: "Option A"})

	// Open vote
	_ = voteService.Open(ctx, vote.ID, user.ID)

	// Abstain
	ballot, err := voteService.CastBallot(ctx, vote.ID, voter.ID, &model.CastBallotRequest{
		IsAbstain: true,
	})

	require.NoError(t, err)
	assert.True(t, ballot.IsAbstain)

	// Close and check results
	voteService.Close(ctx, vote.ID, user.ID)
	results, err := voteService.GetResults(ctx, vote.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, results.TotalAbstains)
}
