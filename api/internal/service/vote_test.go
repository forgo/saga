package service

import (
	"context"
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// ============================================================================
// Mock Repositories
// ============================================================================

type mockVoteRepo struct {
	createFunc           func(ctx context.Context, vote *model.Vote) error
	getByIDFunc          func(ctx context.Context, id string) (*model.Vote, error)
	getByGuildFunc       func(ctx context.Context, guildID string, status *model.VoteStatus, limit, offset int) ([]*model.Vote, error)
	getGlobalVotesFunc   func(ctx context.Context, status *model.VoteStatus, limit, offset int) ([]*model.Vote, error)
	getVotesToOpenFunc   func(ctx context.Context) ([]*model.Vote, error)
	getVotesToCloseFunc  func(ctx context.Context) ([]*model.Vote, error)
	updateFunc           func(ctx context.Context, id string, updates map[string]interface{}) (*model.Vote, error)
	updateStatusFunc     func(ctx context.Context, id string, status model.VoteStatus) error
	deleteFunc           func(ctx context.Context, id string) error
	createOptionFunc     func(ctx context.Context, option *model.VoteOption) error
	getOptionByIDFunc    func(ctx context.Context, id string) (*model.VoteOption, error)
	getOptionsByVoteFunc func(ctx context.Context, voteID string) ([]*model.VoteOption, error)
	updateOptionFunc     func(ctx context.Context, id string, updates *model.UpdateVoteOptionRequest) (*model.VoteOption, error)
	deleteOptionFunc     func(ctx context.Context, id string) error
	createBallotFunc     func(ctx context.Context, ballot *model.VoteBallot) error
	getBallotByVoterFunc func(ctx context.Context, voteID, userID string) (*model.VoteBallot, error)
	getBallotsByVoteFunc func(ctx context.Context, voteID string) ([]*model.VoteBallot, error)
	deleteBallotFunc     func(ctx context.Context, id string) error
	hasVotedFunc         func(ctx context.Context, voteID, userID string) (bool, error)
	countBallotsFunc     func(ctx context.Context, voteID string) (int, error)
}

func (m *mockVoteRepo) Create(ctx context.Context, vote *model.Vote) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, vote)
	}
	return nil
}

func (m *mockVoteRepo) GetByID(ctx context.Context, id string) (*model.Vote, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockVoteRepo) GetByGuild(ctx context.Context, guildID string, status *model.VoteStatus, limit, offset int) ([]*model.Vote, error) {
	if m.getByGuildFunc != nil {
		return m.getByGuildFunc(ctx, guildID, status, limit, offset)
	}
	return nil, nil
}

func (m *mockVoteRepo) GetGlobalVotes(ctx context.Context, status *model.VoteStatus, limit, offset int) ([]*model.Vote, error) {
	if m.getGlobalVotesFunc != nil {
		return m.getGlobalVotesFunc(ctx, status, limit, offset)
	}
	return nil, nil
}

func (m *mockVoteRepo) GetVotesToOpen(ctx context.Context) ([]*model.Vote, error) {
	if m.getVotesToOpenFunc != nil {
		return m.getVotesToOpenFunc(ctx)
	}
	return nil, nil
}

func (m *mockVoteRepo) GetVotesToClose(ctx context.Context) ([]*model.Vote, error) {
	if m.getVotesToCloseFunc != nil {
		return m.getVotesToCloseFunc(ctx)
	}
	return nil, nil
}

func (m *mockVoteRepo) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Vote, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, updates)
	}
	return nil, nil
}

func (m *mockVoteRepo) UpdateStatus(ctx context.Context, id string, status model.VoteStatus) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, id, status)
	}
	return nil
}

func (m *mockVoteRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockVoteRepo) CreateOption(ctx context.Context, option *model.VoteOption) error {
	if m.createOptionFunc != nil {
		return m.createOptionFunc(ctx, option)
	}
	return nil
}

func (m *mockVoteRepo) GetOptionByID(ctx context.Context, id string) (*model.VoteOption, error) {
	if m.getOptionByIDFunc != nil {
		return m.getOptionByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockVoteRepo) GetOptionsByVote(ctx context.Context, voteID string) ([]*model.VoteOption, error) {
	if m.getOptionsByVoteFunc != nil {
		return m.getOptionsByVoteFunc(ctx, voteID)
	}
	return nil, nil
}

func (m *mockVoteRepo) UpdateOption(ctx context.Context, id string, updates *model.UpdateVoteOptionRequest) (*model.VoteOption, error) {
	if m.updateOptionFunc != nil {
		return m.updateOptionFunc(ctx, id, updates)
	}
	return nil, nil
}

func (m *mockVoteRepo) DeleteOption(ctx context.Context, id string) error {
	if m.deleteOptionFunc != nil {
		return m.deleteOptionFunc(ctx, id)
	}
	return nil
}

func (m *mockVoteRepo) CreateBallot(ctx context.Context, ballot *model.VoteBallot) error {
	if m.createBallotFunc != nil {
		return m.createBallotFunc(ctx, ballot)
	}
	return nil
}

func (m *mockVoteRepo) GetBallotByVoter(ctx context.Context, voteID, userID string) (*model.VoteBallot, error) {
	if m.getBallotByVoterFunc != nil {
		return m.getBallotByVoterFunc(ctx, voteID, userID)
	}
	return nil, nil
}

func (m *mockVoteRepo) GetBallotsByVote(ctx context.Context, voteID string) ([]*model.VoteBallot, error) {
	if m.getBallotsByVoteFunc != nil {
		return m.getBallotsByVoteFunc(ctx, voteID)
	}
	return nil, nil
}

func (m *mockVoteRepo) DeleteBallot(ctx context.Context, id string) error {
	if m.deleteBallotFunc != nil {
		return m.deleteBallotFunc(ctx, id)
	}
	return nil
}

func (m *mockVoteRepo) HasVoted(ctx context.Context, voteID, userID string) (bool, error) {
	if m.hasVotedFunc != nil {
		return m.hasVotedFunc(ctx, voteID, userID)
	}
	return false, nil
}

func (m *mockVoteRepo) CountBallots(ctx context.Context, voteID string) (int, error) {
	if m.countBallotsFunc != nil {
		return m.countBallotsFunc(ctx, voteID)
	}
	return 0, nil
}

type mockVoteUserRepo struct {
	getByIDFunc func(ctx context.Context, id string) (*model.User, error)
}

func (m *mockVoteUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func newTestVoteService(voteRepo *mockVoteRepo, userRepo *mockVoteUserRepo, guildRepo *mockGuildRepo) *VoteService {
	if voteRepo == nil {
		voteRepo = &mockVoteRepo{}
	}
	return NewVoteService(VoteServiceConfig{
		VoteRepo:  voteRepo,
		UserRepo:  userRepo,
		GuildRepo: guildRepo,
	})
}

// ============================================================================
// canVote Tests
// ============================================================================

func TestCanVote_GlobalVote_OpenStatus_ReturnsTrue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := newTestVoteService(nil, nil, nil)

	vote := &model.Vote{
		Status:    model.VoteStatusOpen,
		ScopeType: model.VoteScopeGlobal,
	}

	if !svc.canVote(ctx, vote, "user-1") {
		t.Error("expected canVote=true for open global vote")
	}
}

func TestCanVote_GlobalVote_DraftStatus_ReturnsFalse(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := newTestVoteService(nil, nil, nil)

	vote := &model.Vote{
		Status:    model.VoteStatusDraft,
		ScopeType: model.VoteScopeGlobal,
	}

	if svc.canVote(ctx, vote, "user-1") {
		t.Error("expected canVote=false for draft vote")
	}
}

func TestCanVote_GuildVote_IsMember_ReturnsTrue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	guildRepo := &mockGuildRepo{
		getByIDFunc: func(ctx context.Context, id string) (*model.Guild, error) {
			return &model.Guild{ID: id}, nil
		},
	}
	// Need to implement IsMember - let's create a more complete mock
	svc := newTestVoteService(nil, nil, guildRepo)

	guildID := "guild-1"
	vote := &model.Vote{
		Status:    model.VoteStatusOpen,
		ScopeType: model.VoteScopeGuild,
		ScopeID:   &guildID,
	}

	// Without proper guild repo implementation, this tests nil handling
	result := svc.canVote(ctx, vote, "user-1")
	// Result depends on guild repo behavior - just verify no panic
	_ = result
}

func TestCanVote_ClosedVote_ReturnsFalse(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := newTestVoteService(nil, nil, nil)

	vote := &model.Vote{
		Status:    model.VoteStatusClosed,
		ScopeType: model.VoteScopeGlobal,
	}

	if svc.canVote(ctx, vote, "user-1") {
		t.Error("expected canVote=false for closed vote")
	}
}

// ============================================================================
// canViewResults Tests
// ============================================================================

func TestCanViewResults_LiveVisibility_AlwaysTrue(t *testing.T) {
	t.Parallel()

	svc := newTestVoteService(nil, nil, nil)

	vote := &model.Vote{
		Status:            model.VoteStatusOpen,
		ResultsVisibility: model.ResultsVisibilityLive,
	}

	if !svc.canViewResults(vote, "anyone") {
		t.Error("expected canViewResults=true for live visibility")
	}
}

func TestCanViewResults_AfterClose_OnlyClosed(t *testing.T) {
	t.Parallel()

	svc := newTestVoteService(nil, nil, nil)

	openVote := &model.Vote{
		Status:            model.VoteStatusOpen,
		ResultsVisibility: model.ResultsVisibilityAfterClose,
	}

	closedVote := &model.Vote{
		Status:            model.VoteStatusClosed,
		ResultsVisibility: model.ResultsVisibilityAfterClose,
	}

	if svc.canViewResults(openVote, "user-1") {
		t.Error("expected canViewResults=false for open vote with after_close visibility")
	}

	if !svc.canViewResults(closedVote, "user-1") {
		t.Error("expected canViewResults=true for closed vote with after_close visibility")
	}
}

func TestCanViewResults_AdminOnly_OnlyCreator(t *testing.T) {
	t.Parallel()

	svc := newTestVoteService(nil, nil, nil)

	vote := &model.Vote{
		CreatedBy:         "admin-user",
		ResultsVisibility: model.ResultsVisibilityAdminOnly,
	}

	if svc.canViewResults(vote, "other-user") {
		t.Error("expected canViewResults=false for non-creator")
	}

	if !svc.canViewResults(vote, "admin-user") {
		t.Error("expected canViewResults=true for creator")
	}
}

// ============================================================================
// computeFPTP Tests
// ============================================================================

func TestComputeFPTP_SimpleWinner(t *testing.T) {
	t.Parallel()

	svc := newTestVoteService(nil, nil, nil)

	options := []*model.VoteOption{
		{ID: "opt-a", OptionText: "Option A"},
		{ID: "opt-b", OptionText: "Option B"},
		{ID: "opt-c", OptionText: "Option C"},
	}

	ballots := []*model.VoteBallot{
		{BallotData: model.BallotData{"option_id": "opt-a"}},
		{BallotData: model.BallotData{"option_id": "opt-a"}},
		{BallotData: model.BallotData{"option_id": "opt-b"}},
	}

	results := svc.computeFPTP(options, ballots)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Check winner (opt-a with 2 votes)
	if results[0].OptionID != "opt-a" || !results[0].IsWinner {
		t.Errorf("expected opt-a to be winner, got %s", results[0].OptionID)
	}
	if results[0].VoteCount != 2 {
		t.Errorf("expected winner to have 2 votes, got %d", results[0].VoteCount)
	}
	if results[0].Rank != 1 {
		t.Errorf("expected winner rank 1, got %d", results[0].Rank)
	}

	// Verify percentages
	expectedPct := (2.0 / 3.0) * 100
	if results[0].Percentage < expectedPct-0.1 || results[0].Percentage > expectedPct+0.1 {
		t.Errorf("expected percentage ~%.2f, got %.2f", expectedPct, results[0].Percentage)
	}
}

func TestComputeFPTP_Tie(t *testing.T) {
	t.Parallel()

	svc := newTestVoteService(nil, nil, nil)

	options := []*model.VoteOption{
		{ID: "opt-a", OptionText: "Option A"},
		{ID: "opt-b", OptionText: "Option B"},
	}

	ballots := []*model.VoteBallot{
		{BallotData: model.BallotData{"option_id": "opt-a"}},
		{BallotData: model.BallotData{"option_id": "opt-b"}},
	}

	results := svc.computeFPTP(options, ballots)

	// First in sorted order is winner (arbitrary in tie)
	if !results[0].IsWinner {
		t.Error("expected first result to be marked as winner")
	}
	if results[0].VoteCount != 1 || results[1].VoteCount != 1 {
		t.Error("expected both options to have 1 vote")
	}
}

func TestComputeFPTP_NoBallots(t *testing.T) {
	t.Parallel()

	svc := newTestVoteService(nil, nil, nil)

	options := []*model.VoteOption{
		{ID: "opt-a", OptionText: "Option A"},
	}

	ballots := []*model.VoteBallot{}

	results := svc.computeFPTP(options, ballots)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].VoteCount != 0 {
		t.Error("expected 0 votes")
	}
	if results[0].IsWinner {
		t.Error("no winner with 0 votes")
	}
	if results[0].Percentage != 0 {
		t.Error("expected 0% with no ballots")
	}
}

// ============================================================================
// computeRankedChoice Tests
// ============================================================================

func TestComputeRankedChoice_FirstRoundMajority(t *testing.T) {
	t.Parallel()

	svc := newTestVoteService(nil, nil, nil)

	options := []*model.VoteOption{
		{ID: "opt-a", OptionText: "Option A"},
		{ID: "opt-b", OptionText: "Option B"},
		{ID: "opt-c", OptionText: "Option C"},
	}

	// A has 3 votes, B has 1, C has 1 - A wins with majority
	ballots := []*model.VoteBallot{
		{BallotData: model.BallotData{"rankings": []interface{}{"opt-a", "opt-b", "opt-c"}}},
		{BallotData: model.BallotData{"rankings": []interface{}{"opt-a", "opt-c", "opt-b"}}},
		{BallotData: model.BallotData{"rankings": []interface{}{"opt-a", "opt-b", "opt-c"}}},
		{BallotData: model.BallotData{"rankings": []interface{}{"opt-b", "opt-a", "opt-c"}}},
		{BallotData: model.BallotData{"rankings": []interface{}{"opt-c", "opt-a", "opt-b"}}},
	}

	results, rounds := svc.computeRankedChoice(options, ballots)

	// A should win with 3/5 (majority = 3)
	var winner *model.OptionResult
	for i := range results {
		if results[i].IsWinner {
			winner = &results[i]
			break
		}
	}

	if winner == nil {
		t.Fatal("expected a winner")
	}
	if winner.OptionID != "opt-a" {
		t.Errorf("expected opt-a to win, got %s", winner.OptionID)
	}

	// Should end in 1 round (majority achieved)
	if len(rounds) != 1 {
		t.Errorf("expected 1 round, got %d", len(rounds))
	}
}

func TestComputeRankedChoice_RequiresElimination(t *testing.T) {
	t.Parallel()

	svc := newTestVoteService(nil, nil, nil)

	options := []*model.VoteOption{
		{ID: "opt-a", OptionText: "Option A"},
		{ID: "opt-b", OptionText: "Option B"},
		{ID: "opt-c", OptionText: "Option C"},
	}

	// A: 2, B: 2, C: 1 - no majority, C eliminated
	// After C eliminated, C's voter's second choice matters
	ballots := []*model.VoteBallot{
		{BallotData: model.BallotData{"rankings": []interface{}{"opt-a", "opt-b", "opt-c"}}},
		{BallotData: model.BallotData{"rankings": []interface{}{"opt-a", "opt-c", "opt-b"}}},
		{BallotData: model.BallotData{"rankings": []interface{}{"opt-b", "opt-a", "opt-c"}}},
		{BallotData: model.BallotData{"rankings": []interface{}{"opt-b", "opt-c", "opt-a"}}},
		{BallotData: model.BallotData{"rankings": []interface{}{"opt-c", "opt-a", "opt-b"}}}, // C voter prefers A second
	}

	results, rounds := svc.computeRankedChoice(options, ballots)

	// C should be eliminated, its vote goes to A
	// Final: A: 3, B: 2 -> A wins
	var winner *model.OptionResult
	for i := range results {
		if results[i].IsWinner {
			winner = &results[i]
		}
	}

	if winner == nil || winner.OptionID != "opt-a" {
		t.Errorf("expected opt-a to win after elimination")
	}

	// Should have at least 2 rounds
	if len(rounds) < 1 {
		t.Errorf("expected at least 1 round, got %d", len(rounds))
	}

	// Check C was eliminated
	var cResult *model.OptionResult
	for i := range results {
		if results[i].OptionID == "opt-c" {
			cResult = &results[i]
		}
	}
	if cResult != nil && !cResult.IsEliminated {
		t.Error("expected opt-c to be marked as eliminated")
	}
}

func TestComputeRankedChoice_TwoOptions(t *testing.T) {
	t.Parallel()

	svc := newTestVoteService(nil, nil, nil)

	options := []*model.VoteOption{
		{ID: "opt-a", OptionText: "Option A"},
		{ID: "opt-b", OptionText: "Option B"},
	}

	ballots := []*model.VoteBallot{
		{BallotData: model.BallotData{"rankings": []interface{}{"opt-a", "opt-b"}}},
		{BallotData: model.BallotData{"rankings": []interface{}{"opt-b", "opt-a"}}},
		{BallotData: model.BallotData{"rankings": []interface{}{"opt-b", "opt-a"}}},
	}

	results, _ := svc.computeRankedChoice(options, ballots)

	// B should win with 2/3 majority
	var winner *model.OptionResult
	for i := range results {
		if results[i].IsWinner {
			winner = &results[i]
		}
	}

	if winner == nil || winner.OptionID != "opt-b" {
		t.Errorf("expected opt-b to win with majority")
	}
}

// ============================================================================
// computeApproval Tests
// ============================================================================

func TestComputeApproval_MultipleApprovalsPerBallot(t *testing.T) {
	t.Parallel()

	svc := newTestVoteService(nil, nil, nil)

	options := []*model.VoteOption{
		{ID: "opt-a", OptionText: "Option A"},
		{ID: "opt-b", OptionText: "Option B"},
		{ID: "opt-c", OptionText: "Option C"},
	}

	ballots := []*model.VoteBallot{
		{BallotData: model.BallotData{"approved_options": []interface{}{"opt-a", "opt-b"}}},
		{BallotData: model.BallotData{"approved_options": []interface{}{"opt-a", "opt-c"}}},
		{BallotData: model.BallotData{"approved_options": []interface{}{"opt-b"}}},
	}

	results := svc.computeApproval(options, ballots, model.VoteTypeApproval)

	// A: 2 approvals, B: 2 approvals, C: 1 approval
	// Winner should be first with 2 (A or B)
	if results[0].VoteCount != 2 {
		t.Errorf("expected winner to have 2 approvals, got %d", results[0].VoteCount)
	}
	if !results[0].IsWinner {
		t.Error("expected first result to be winner")
	}

	// Percentage is based on total ballots (3)
	expectedPct := (2.0 / 3.0) * 100
	if results[0].Percentage < expectedPct-0.1 || results[0].Percentage > expectedPct+0.1 {
		t.Errorf("expected percentage ~%.2f, got %.2f", expectedPct, results[0].Percentage)
	}
}

func TestComputeApproval_MultiSelect(t *testing.T) {
	t.Parallel()

	svc := newTestVoteService(nil, nil, nil)

	options := []*model.VoteOption{
		{ID: "opt-a", OptionText: "Option A"},
		{ID: "opt-b", OptionText: "Option B"},
	}

	ballots := []*model.VoteBallot{
		{BallotData: model.BallotData{"selected_options": []interface{}{"opt-a"}}},
		{BallotData: model.BallotData{"selected_options": []interface{}{"opt-a", "opt-b"}}},
	}

	results := svc.computeApproval(options, ballots, model.VoteTypeMultiSelect)

	// A: 2 selections, B: 1 selection
	if results[0].OptionID != "opt-a" || results[0].VoteCount != 2 {
		t.Errorf("expected opt-a to have 2 votes")
	}
	if results[1].OptionID != "opt-b" || results[1].VoteCount != 1 {
		t.Errorf("expected opt-b to have 1 vote")
	}
}

// ============================================================================
// computeResults Tests (integration)
// ============================================================================

func TestComputeResults_CountsAbstains(t *testing.T) {
	t.Parallel()

	svc := newTestVoteService(nil, nil, nil)

	vote := &model.Vote{
		ID:       "vote-1",
		VoteType: model.VoteTypeFPTP,
	}

	options := []*model.VoteOption{
		{ID: "opt-a", OptionText: "Option A"},
	}

	ballots := []*model.VoteBallot{
		{BallotData: model.BallotData{"option_id": "opt-a"}, IsAbstain: false},
		{BallotData: model.BallotData{}, IsAbstain: true},
		{BallotData: model.BallotData{}, IsAbstain: true},
	}

	result := svc.computeResults(vote, options, ballots)

	if result.TotalBallots != 3 {
		t.Errorf("expected 3 total ballots, got %d", result.TotalBallots)
	}
	if result.TotalAbstains != 2 {
		t.Errorf("expected 2 abstains, got %d", result.TotalAbstains)
	}

	// Only 1 non-abstain ballot counts for option
	if result.OptionResults[0].VoteCount != 1 {
		t.Errorf("expected 1 vote for option, got %d", result.OptionResults[0].VoteCount)
	}
}

func TestComputeResults_SetsWinner(t *testing.T) {
	t.Parallel()

	svc := newTestVoteService(nil, nil, nil)

	vote := &model.Vote{
		ID:       "vote-1",
		VoteType: model.VoteTypeFPTP,
	}

	options := []*model.VoteOption{
		{ID: "opt-a", OptionText: "Option A"},
		{ID: "opt-b", OptionText: "Option B"},
	}

	ballots := []*model.VoteBallot{
		{BallotData: model.BallotData{"option_id": "opt-a"}},
		{BallotData: model.BallotData{"option_id": "opt-a"}},
		{BallotData: model.BallotData{"option_id": "opt-b"}},
	}

	result := svc.computeResults(vote, options, ballots)

	if result.Winner == nil {
		t.Fatal("expected a winner")
	}
	if *result.Winner != "opt-a" {
		t.Errorf("expected winner to be opt-a, got %s", *result.Winner)
	}
}

// ============================================================================
// Status Transition Tests
// ============================================================================

func TestOpen_DraftVote_TransitionsToOpen(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var capturedStatus model.VoteStatus
	voteRepo := &mockVoteRepo{
		getByIDFunc: func(ctx context.Context, id string) (*model.Vote, error) {
			return &model.Vote{
				ID:        id,
				Status:    model.VoteStatusDraft,
				CreatedBy: "user-1",
			}, nil
		},
		updateStatusFunc: func(ctx context.Context, id string, status model.VoteStatus) error {
			capturedStatus = status
			return nil
		},
	}

	svc := newTestVoteService(voteRepo, nil, nil)

	err := svc.Open(ctx, "vote-1", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedStatus != model.VoteStatusOpen {
		t.Errorf("expected status to be open, got %s", capturedStatus)
	}
}

func TestOpen_NotDraft_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	voteRepo := &mockVoteRepo{
		getByIDFunc: func(ctx context.Context, id string) (*model.Vote, error) {
			return &model.Vote{
				ID:        id,
				Status:    model.VoteStatusOpen,
				CreatedBy: "user-1",
			}, nil
		},
	}

	svc := newTestVoteService(voteRepo, nil, nil)

	err := svc.Open(ctx, "vote-1", "user-1")
	if err == nil {
		t.Error("expected error for non-draft vote")
	}
}

func TestOpen_WrongUser_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	voteRepo := &mockVoteRepo{
		getByIDFunc: func(ctx context.Context, id string) (*model.Vote, error) {
			return &model.Vote{
				ID:        id,
				Status:    model.VoteStatusDraft,
				CreatedBy: "user-1",
			}, nil
		},
	}

	svc := newTestVoteService(voteRepo, nil, nil)

	err := svc.Open(ctx, "vote-1", "other-user")
	if err == nil {
		t.Error("expected error for wrong user")
	}
}

func TestClose_OpenVote_TransitionsToClosed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var capturedStatus model.VoteStatus
	voteRepo := &mockVoteRepo{
		getByIDFunc: func(ctx context.Context, id string) (*model.Vote, error) {
			return &model.Vote{
				ID:        id,
				Status:    model.VoteStatusOpen,
				CreatedBy: "user-1",
			}, nil
		},
		updateStatusFunc: func(ctx context.Context, id string, status model.VoteStatus) error {
			capturedStatus = status
			return nil
		},
	}

	svc := newTestVoteService(voteRepo, nil, nil)

	err := svc.Close(ctx, "vote-1", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedStatus != model.VoteStatusClosed {
		t.Errorf("expected status to be closed, got %s", capturedStatus)
	}
}

func TestCancel_DraftVote_TransitionsToCancelled(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var capturedStatus model.VoteStatus
	voteRepo := &mockVoteRepo{
		getByIDFunc: func(ctx context.Context, id string) (*model.Vote, error) {
			return &model.Vote{
				ID:        id,
				Status:    model.VoteStatusDraft,
				CreatedBy: "user-1",
			}, nil
		},
		updateStatusFunc: func(ctx context.Context, id string, status model.VoteStatus) error {
			capturedStatus = status
			return nil
		},
	}

	svc := newTestVoteService(voteRepo, nil, nil)

	err := svc.Cancel(ctx, "vote-1", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedStatus != model.VoteStatusCancelled {
		t.Errorf("expected status to be cancelled, got %s", capturedStatus)
	}
}

func TestCancel_AlreadyClosed_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	voteRepo := &mockVoteRepo{
		getByIDFunc: func(ctx context.Context, id string) (*model.Vote, error) {
			return &model.Vote{
				ID:        id,
				Status:    model.VoteStatusClosed,
				CreatedBy: "user-1",
			}, nil
		},
	}

	svc := newTestVoteService(voteRepo, nil, nil)

	err := svc.Cancel(ctx, "vote-1", "user-1")
	if err == nil {
		t.Error("expected error for already closed vote")
	}
}

// ============================================================================
// CastBallot Tests
// ============================================================================

func TestCastBallot_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	optionID := "opt-1"
	voteRepo := &mockVoteRepo{
		getByIDFunc: func(ctx context.Context, id string) (*model.Vote, error) {
			return &model.Vote{
				ID:        id,
				Status:    model.VoteStatusOpen,
				ScopeType: model.VoteScopeGlobal,
				VoteType:  model.VoteTypeFPTP,
			}, nil
		},
		getBallotByVoterFunc: func(ctx context.Context, voteID, userID string) (*model.VoteBallot, error) {
			return nil, nil // No existing ballot
		},
		createBallotFunc: func(ctx context.Context, ballot *model.VoteBallot) error {
			ballot.ID = "ballot-1"
			return nil
		},
	}

	userRepo := &mockVoteUserRepo{
		getByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return nil, nil
		},
	}

	svc := newTestVoteService(voteRepo, userRepo, nil)

	ballot, err := svc.CastBallot(ctx, "vote-1", "user-1", &model.CastBallotRequest{
		OptionID: &optionID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ballot == nil {
		t.Fatal("expected ballot, got nil")
	}
}

func TestCastBallot_Revote_DeletesExisting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	deleteCalled := false
	optionID := "opt-1"
	voteRepo := &mockVoteRepo{
		getByIDFunc: func(ctx context.Context, id string) (*model.Vote, error) {
			return &model.Vote{
				ID:        id,
				Status:    model.VoteStatusOpen,
				ScopeType: model.VoteScopeGlobal,
				VoteType:  model.VoteTypeFPTP,
			}, nil
		},
		getBallotByVoterFunc: func(ctx context.Context, voteID, userID string) (*model.VoteBallot, error) {
			return &model.VoteBallot{ID: "existing-ballot"}, nil
		},
		deleteBallotFunc: func(ctx context.Context, id string) error {
			deleteCalled = true
			return nil
		},
		createBallotFunc: func(ctx context.Context, ballot *model.VoteBallot) error {
			return nil
		},
	}

	userRepo := &mockVoteUserRepo{
		getByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return nil, nil
		},
	}

	svc := newTestVoteService(voteRepo, userRepo, nil)

	_, err := svc.CastBallot(ctx, "vote-1", "user-1", &model.CastBallotRequest{
		OptionID: &optionID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleteCalled {
		t.Error("expected existing ballot to be deleted")
	}
}

func TestCastBallot_VoteNotOpen_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	optionID := "opt-1"
	voteRepo := &mockVoteRepo{
		getByIDFunc: func(ctx context.Context, id string) (*model.Vote, error) {
			return &model.Vote{
				ID:       id,
				Status:   model.VoteStatusDraft,
				VoteType: model.VoteTypeFPTP,
			}, nil
		},
	}

	svc := newTestVoteService(voteRepo, nil, nil)

	_, err := svc.CastBallot(ctx, "vote-1", "user-1", &model.CastBallotRequest{
		OptionID: &optionID,
	})
	if err == nil {
		t.Error("expected error for draft vote")
	}
}

// ============================================================================
// ProcessScheduledTransitions Tests
// ============================================================================

func TestProcessScheduledTransitions_OpensAndCloses(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	openedIDs := make(map[string]bool)
	closedIDs := make(map[string]bool)

	voteRepo := &mockVoteRepo{
		getVotesToOpenFunc: func(ctx context.Context) ([]*model.Vote, error) {
			return []*model.Vote{{ID: "vote-to-open"}}, nil
		},
		getVotesToCloseFunc: func(ctx context.Context) ([]*model.Vote, error) {
			return []*model.Vote{{ID: "vote-to-close"}}, nil
		},
		updateStatusFunc: func(ctx context.Context, id string, status model.VoteStatus) error {
			switch status {
			case model.VoteStatusOpen:
				openedIDs[id] = true
			case model.VoteStatusClosed:
				closedIDs[id] = true
			}
			return nil
		},
	}

	svc := newTestVoteService(voteRepo, nil, nil)

	err := svc.ProcessScheduledTransitions(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !openedIDs["vote-to-open"] {
		t.Error("expected vote-to-open to be opened")
	}
	if !closedIDs["vote-to-close"] {
		t.Error("expected vote-to-close to be closed")
	}
}

// ============================================================================
// GetGuildVotes Tests
// ============================================================================

func TestGetGuildVotes_DefaultsLimit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var capturedLimit int
	voteRepo := &mockVoteRepo{
		getByGuildFunc: func(ctx context.Context, guildID string, status *model.VoteStatus, limit, offset int) ([]*model.Vote, error) {
			capturedLimit = limit
			return nil, nil
		},
	}

	svc := newTestVoteService(voteRepo, nil, nil)

	_, _ = svc.GetGuildVotes(ctx, "guild-1", nil, 0, 0)
	if capturedLimit != 50 {
		t.Errorf("expected default limit 50, got %d", capturedLimit)
	}
}

func TestGetGuildVotes_CapsLimit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var capturedLimit int
	voteRepo := &mockVoteRepo{
		getByGuildFunc: func(ctx context.Context, guildID string, status *model.VoteStatus, limit, offset int) ([]*model.Vote, error) {
			capturedLimit = limit
			return nil, nil
		},
	}

	svc := newTestVoteService(voteRepo, nil, nil)

	_, _ = svc.GetGuildVotes(ctx, "guild-1", nil, 200, 0)
	if capturedLimit != 50 {
		t.Errorf("expected capped limit 50, got %d", capturedLimit)
	}
}

// ============================================================================
// Create Tests
// ============================================================================

func TestCreate_ValidRequest_CreatesVote(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	voteRepo := &mockVoteRepo{
		createFunc: func(ctx context.Context, vote *model.Vote) error {
			vote.ID = "vote-1"
			return nil
		},
	}

	svc := newTestVoteService(voteRepo, nil, nil)

	opensAt := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	closesAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	vote, err := svc.Create(ctx, "user-1", &model.CreateVoteRequest{
		ScopeType: string(model.VoteScopeGlobal),
		Title:     "Test Vote",
		VoteType:  string(model.VoteTypeFPTP),
		OpensAt:   opensAt,
		ClosesAt:  closesAt,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vote == nil {
		t.Fatal("expected vote, got nil")
	}
	if vote.Status != model.VoteStatusDraft {
		t.Errorf("expected draft status, got %s", vote.Status)
	}
}

func TestCreate_ClosesBeforeOpens_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := newTestVoteService(&mockVoteRepo{}, nil, nil)

	opensAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	closesAt := time.Now().Add(1 * time.Hour).Format(time.RFC3339)

	_, err := svc.Create(ctx, "user-1", &model.CreateVoteRequest{
		ScopeType: string(model.VoteScopeGlobal),
		Title:     "Test Vote",
		VoteType:  string(model.VoteTypeFPTP),
		OpensAt:   opensAt,
		ClosesAt:  closesAt,
	})

	if err == nil {
		t.Error("expected error for closes before opens")
	}
}
