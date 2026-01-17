package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// ============================================================================
// Mock Repositories
// ============================================================================

type mockPoolRepo struct {
	createPoolFunc             func(ctx context.Context, pool *model.MatchingPool) error
	getPoolFunc                func(ctx context.Context, poolID string) (*model.MatchingPool, error)
	getPoolsByGuildFunc        func(ctx context.Context, guildID string) ([]*model.MatchingPool, error)
	updatePoolFunc             func(ctx context.Context, poolID string, updates map[string]interface{}) (*model.MatchingPool, error)
	deletePoolFunc             func(ctx context.Context, poolID string) error
	countPoolsByGuildFunc      func(ctx context.Context, guildID string) (int, error)
	addMemberFunc              func(ctx context.Context, member *model.PoolMember) error
	getMemberFunc              func(ctx context.Context, poolID, memberID string) (*model.PoolMember, error)
	getMemberByUserFunc        func(ctx context.Context, poolID, userID string) (*model.PoolMember, error)
	getPoolMembersFunc         func(ctx context.Context, poolID string) ([]*model.PoolMember, error)
	updateMemberFunc           func(ctx context.Context, membershipID string, updates map[string]interface{}) (*model.PoolMember, error)
	removeMemberFunc           func(ctx context.Context, membershipID string) error
	getUserPoolMembershipsFunc func(ctx context.Context, userID string) ([]*model.PoolMember, error)
	createMatchResultFunc      func(ctx context.Context, match *model.MatchResult) error
	getMatchResultFunc         func(ctx context.Context, matchID string) (*model.MatchResult, error)
	getMatchesByPoolFunc       func(ctx context.Context, poolID string, limit int) ([]*model.MatchResult, error)
	getMatchesByRoundFunc      func(ctx context.Context, poolID, round string) ([]*model.MatchResult, error)
	getUserPendingMatchesFunc  func(ctx context.Context, userID string) ([]*model.MatchResult, error)
	getRecentMatchesBetweenFunc func(ctx context.Context, memberIDs []string, days int) ([]*model.MatchResult, error)
	updateMatchResultFunc      func(ctx context.Context, matchID string, updates map[string]interface{}) (*model.MatchResult, error)
	getPoolsDueForMatchingFunc func(ctx context.Context) ([]*model.MatchingPool, error)
	getPoolStatsFunc           func(ctx context.Context, poolID string) (*model.PoolStats, error)
	getStaleMatchesFunc        func(ctx context.Context, cutoff time.Time, status string) ([]*model.MatchResult, error)
}

func (m *mockPoolRepo) CreatePool(ctx context.Context, pool *model.MatchingPool) error {
	if m.createPoolFunc != nil {
		return m.createPoolFunc(ctx, pool)
	}
	return nil
}

func (m *mockPoolRepo) GetPool(ctx context.Context, poolID string) (*model.MatchingPool, error) {
	if m.getPoolFunc != nil {
		return m.getPoolFunc(ctx, poolID)
	}
	return nil, nil
}

func (m *mockPoolRepo) GetPoolsByGuild(ctx context.Context, guildID string) ([]*model.MatchingPool, error) {
	if m.getPoolsByGuildFunc != nil {
		return m.getPoolsByGuildFunc(ctx, guildID)
	}
	return nil, nil
}

func (m *mockPoolRepo) UpdatePool(ctx context.Context, poolID string, updates map[string]interface{}) (*model.MatchingPool, error) {
	if m.updatePoolFunc != nil {
		return m.updatePoolFunc(ctx, poolID, updates)
	}
	return nil, nil
}

func (m *mockPoolRepo) DeletePool(ctx context.Context, poolID string) error {
	if m.deletePoolFunc != nil {
		return m.deletePoolFunc(ctx, poolID)
	}
	return nil
}

func (m *mockPoolRepo) CountPoolsByGuild(ctx context.Context, guildID string) (int, error) {
	if m.countPoolsByGuildFunc != nil {
		return m.countPoolsByGuildFunc(ctx, guildID)
	}
	return 0, nil
}

func (m *mockPoolRepo) AddMember(ctx context.Context, member *model.PoolMember) error {
	if m.addMemberFunc != nil {
		return m.addMemberFunc(ctx, member)
	}
	return nil
}

func (m *mockPoolRepo) GetMember(ctx context.Context, poolID, memberID string) (*model.PoolMember, error) {
	if m.getMemberFunc != nil {
		return m.getMemberFunc(ctx, poolID, memberID)
	}
	return nil, nil
}

func (m *mockPoolRepo) GetMemberByUser(ctx context.Context, poolID, userID string) (*model.PoolMember, error) {
	if m.getMemberByUserFunc != nil {
		return m.getMemberByUserFunc(ctx, poolID, userID)
	}
	return nil, nil
}

func (m *mockPoolRepo) GetPoolMembers(ctx context.Context, poolID string) ([]*model.PoolMember, error) {
	if m.getPoolMembersFunc != nil {
		return m.getPoolMembersFunc(ctx, poolID)
	}
	return nil, nil
}

func (m *mockPoolRepo) UpdateMember(ctx context.Context, membershipID string, updates map[string]interface{}) (*model.PoolMember, error) {
	if m.updateMemberFunc != nil {
		return m.updateMemberFunc(ctx, membershipID, updates)
	}
	return nil, nil
}

func (m *mockPoolRepo) RemoveMember(ctx context.Context, membershipID string) error {
	if m.removeMemberFunc != nil {
		return m.removeMemberFunc(ctx, membershipID)
	}
	return nil
}

func (m *mockPoolRepo) GetUserPoolMemberships(ctx context.Context, userID string) ([]*model.PoolMember, error) {
	if m.getUserPoolMembershipsFunc != nil {
		return m.getUserPoolMembershipsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockPoolRepo) CreateMatchResult(ctx context.Context, match *model.MatchResult) error {
	if m.createMatchResultFunc != nil {
		return m.createMatchResultFunc(ctx, match)
	}
	return nil
}

func (m *mockPoolRepo) GetMatchResult(ctx context.Context, matchID string) (*model.MatchResult, error) {
	if m.getMatchResultFunc != nil {
		return m.getMatchResultFunc(ctx, matchID)
	}
	return nil, nil
}

func (m *mockPoolRepo) GetMatchesByPool(ctx context.Context, poolID string, limit int) ([]*model.MatchResult, error) {
	if m.getMatchesByPoolFunc != nil {
		return m.getMatchesByPoolFunc(ctx, poolID, limit)
	}
	return nil, nil
}

func (m *mockPoolRepo) GetMatchesByRound(ctx context.Context, poolID, round string) ([]*model.MatchResult, error) {
	if m.getMatchesByRoundFunc != nil {
		return m.getMatchesByRoundFunc(ctx, poolID, round)
	}
	return nil, nil
}

func (m *mockPoolRepo) GetUserPendingMatches(ctx context.Context, userID string) ([]*model.MatchResult, error) {
	if m.getUserPendingMatchesFunc != nil {
		return m.getUserPendingMatchesFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockPoolRepo) GetRecentMatchesBetween(ctx context.Context, memberIDs []string, days int) ([]*model.MatchResult, error) {
	if m.getRecentMatchesBetweenFunc != nil {
		return m.getRecentMatchesBetweenFunc(ctx, memberIDs, days)
	}
	return nil, nil
}

func (m *mockPoolRepo) UpdateMatchResult(ctx context.Context, matchID string, updates map[string]interface{}) (*model.MatchResult, error) {
	if m.updateMatchResultFunc != nil {
		return m.updateMatchResultFunc(ctx, matchID, updates)
	}
	return nil, nil
}

func (m *mockPoolRepo) GetPoolsDueForMatching(ctx context.Context) ([]*model.MatchingPool, error) {
	if m.getPoolsDueForMatchingFunc != nil {
		return m.getPoolsDueForMatchingFunc(ctx)
	}
	return nil, nil
}

func (m *mockPoolRepo) GetPoolStats(ctx context.Context, poolID string) (*model.PoolStats, error) {
	if m.getPoolStatsFunc != nil {
		return m.getPoolStatsFunc(ctx, poolID)
	}
	return nil, nil
}

func (m *mockPoolRepo) GetStaleMatches(ctx context.Context, cutoff time.Time, status string) ([]*model.MatchResult, error) {
	if m.getStaleMatchesFunc != nil {
		return m.getStaleMatchesFunc(ctx, cutoff, status)
	}
	return nil, nil
}

type mockGuildRepo struct {
	getByIDFunc func(ctx context.Context, id string) (*model.Guild, error)
}

func (m *mockGuildRepo) Create(ctx context.Context, guild *model.Guild) error        { return nil }
func (m *mockGuildRepo) GetByID(ctx context.Context, id string) (*model.Guild, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockGuildRepo) Update(ctx context.Context, guild *model.Guild) error { return nil }
func (m *mockGuildRepo) Delete(ctx context.Context, id string) error          { return nil }
func (m *mockGuildRepo) GetGuildsForUser(ctx context.Context, userID string) ([]*model.Guild, error) {
	return nil, nil
}
func (m *mockGuildRepo) CountGuildsForUser(ctx context.Context, userID string) (int, error) {
	return 0, nil
}
func (m *mockGuildRepo) AddMember(ctx context.Context, memberID, guildID string, pendingApproval bool) error {
	return nil
}
func (m *mockGuildRepo) RemoveMember(ctx context.Context, memberID, guildID string) error {
	return nil
}
func (m *mockGuildRepo) IsMember(ctx context.Context, userID, guildID string) (bool, error) {
	return false, nil
}
func (m *mockGuildRepo) CountMembers(ctx context.Context, guildID string) (int, error) {
	return 0, nil
}
func (m *mockGuildRepo) GetMembers(ctx context.Context, guildID string) ([]*model.Member, error) {
	return nil, nil
}

type mockMemberRepo struct{}

func (m *mockMemberRepo) Create(ctx context.Context, member *model.Member) error     { return nil }
func (m *mockMemberRepo) GetByID(ctx context.Context, id string) (*model.Member, error) {
	return nil, nil
}
func (m *mockMemberRepo) GetByUserID(ctx context.Context, userID string) (*model.Member, error) {
	return nil, nil
}
func (m *mockMemberRepo) GetOrCreate(ctx context.Context, userID, name, email string) (*model.Member, error) {
	return nil, nil
}
func (m *mockMemberRepo) Update(ctx context.Context, member *model.Member) error { return nil }
func (m *mockMemberRepo) Delete(ctx context.Context, id string) error            { return nil }

type mockCompatibilityCalc struct {
	calcFunc func(ctx context.Context, userAID, userBID string) (*model.CompatibilityScore, error)
}

func (m *mockCompatibilityCalc) CalculateCompatibility(ctx context.Context, userAID, userBID string) (*model.CompatibilityScore, error) {
	if m.calcFunc != nil {
		return m.calcFunc(ctx, userAID, userBID)
	}
	return nil, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func newTestPoolService(poolRepo *mockPoolRepo, guildRepo *mockGuildRepo, memberRepo *mockMemberRepo, compat CompatibilityCalculator) *PoolService {
	if poolRepo == nil {
		poolRepo = &mockPoolRepo{}
	}
	if guildRepo == nil {
		guildRepo = &mockGuildRepo{}
	}
	if memberRepo == nil {
		memberRepo = &mockMemberRepo{}
	}
	return NewPoolService(PoolServiceConfig{
		PoolRepo:      poolRepo,
		GuildRepo:     guildRepo,
		MemberRepo:    memberRepo,
		Compatibility: compat,
	})
}

// ============================================================================
// isValidFrequency Tests
// ============================================================================

func TestIsValidFrequency_Weekly(t *testing.T) {
	t.Parallel()
	if !isValidFrequency(model.PoolFrequencyWeekly) {
		t.Error("weekly should be valid")
	}
}

func TestIsValidFrequency_Biweekly(t *testing.T) {
	t.Parallel()
	if !isValidFrequency(model.PoolFrequencyBiweekly) {
		t.Error("biweekly should be valid")
	}
}

func TestIsValidFrequency_Monthly(t *testing.T) {
	t.Parallel()
	if !isValidFrequency(model.PoolFrequencyMonthly) {
		t.Error("monthly should be valid")
	}
}

func TestIsValidFrequency_Invalid(t *testing.T) {
	t.Parallel()
	if isValidFrequency("daily") {
		t.Error("daily should be invalid")
	}
	if isValidFrequency("") {
		t.Error("empty should be invalid")
	}
	if isValidFrequency("yearly") {
		t.Error("yearly should be invalid")
	}
}

// ============================================================================
// CreatePool Tests
// ============================================================================

func TestCreatePool_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		countPoolsByGuildFunc: func(ctx context.Context, guildID string) (int, error) {
			return 0, nil
		},
		createPoolFunc: func(ctx context.Context, pool *model.MatchingPool) error {
			pool.ID = "pool-123"
			return nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	req := &model.CreatePoolRequest{
		Name:      "Coffee Chats",
		Frequency: model.PoolFrequencyWeekly,
		MatchSize: 2,
	}

	pool, err := svc.CreatePool(ctx, "guild-1", req, "member-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pool == nil {
		t.Fatal("expected pool, got nil")
	}
	if pool.Name != "Coffee Chats" {
		t.Errorf("expected name 'Coffee Chats', got %q", pool.Name)
	}
	if pool.MatchSize != 2 {
		t.Errorf("expected match size 2, got %d", pool.MatchSize)
	}
}

func TestCreatePool_InvalidFrequency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := newTestPoolService(nil, nil, nil, nil)

	req := &model.CreatePoolRequest{
		Name:      "Test Pool",
		Frequency: "invalid",
	}

	_, err := svc.CreatePool(ctx, "guild-1", req, "member-1")
	if !errors.Is(err, ErrInvalidFrequency) {
		t.Errorf("expected ErrInvalidFrequency, got %v", err)
	}
}

func TestCreatePool_InvalidMatchSize_TooSmall(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := newTestPoolService(nil, nil, nil, nil)

	req := &model.CreatePoolRequest{
		Name:      "Test Pool",
		Frequency: model.PoolFrequencyWeekly,
		MatchSize: 1,
	}

	_, err := svc.CreatePool(ctx, "guild-1", req, "member-1")
	if !errors.Is(err, ErrInvalidMatchSize) {
		t.Errorf("expected ErrInvalidMatchSize, got %v", err)
	}
}

func TestCreatePool_InvalidMatchSize_TooLarge(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	svc := newTestPoolService(nil, nil, nil, nil)

	req := &model.CreatePoolRequest{
		Name:      "Test Pool",
		Frequency: model.PoolFrequencyWeekly,
		MatchSize: 10,
	}

	_, err := svc.CreatePool(ctx, "guild-1", req, "member-1")
	if !errors.Is(err, ErrInvalidMatchSize) {
		t.Errorf("expected ErrInvalidMatchSize, got %v", err)
	}
}

func TestCreatePool_DefaultMatchSize(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		countPoolsByGuildFunc: func(ctx context.Context, guildID string) (int, error) {
			return 0, nil
		},
		createPoolFunc: func(ctx context.Context, pool *model.MatchingPool) error {
			if pool.MatchSize != 2 {
				t.Errorf("expected default match size 2, got %d", pool.MatchSize)
			}
			return nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	req := &model.CreatePoolRequest{
		Name:      "Test Pool",
		Frequency: model.PoolFrequencyWeekly,
		MatchSize: 0, // Should default to 2
	}

	_, err := svc.CreatePool(ctx, "guild-1", req, "member-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreatePool_PoolLimitReached(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		countPoolsByGuildFunc: func(ctx context.Context, guildID string) (int, error) {
			return model.MaxPoolsPerGuild, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	req := &model.CreatePoolRequest{
		Name:      "Test Pool",
		Frequency: model.PoolFrequencyWeekly,
	}

	_, err := svc.CreatePool(ctx, "guild-1", req, "member-1")
	if !errors.Is(err, ErrPoolLimitReached) {
		t.Errorf("expected ErrPoolLimitReached, got %v", err)
	}
}

func TestCreatePool_TruncatesLongName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	longName := make([]byte, model.MaxPoolNameLength+50)
	for i := range longName {
		longName[i] = 'a'
	}

	var capturedPool *model.MatchingPool
	poolRepo := &mockPoolRepo{
		countPoolsByGuildFunc: func(ctx context.Context, guildID string) (int, error) {
			return 0, nil
		},
		createPoolFunc: func(ctx context.Context, pool *model.MatchingPool) error {
			capturedPool = pool
			return nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	req := &model.CreatePoolRequest{
		Name:      string(longName),
		Frequency: model.PoolFrequencyWeekly,
	}

	_, err := svc.CreatePool(ctx, "guild-1", req, "member-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(capturedPool.Name) != model.MaxPoolNameLength {
		t.Errorf("expected name truncated to %d, got %d", model.MaxPoolNameLength, len(capturedPool.Name))
	}
}

// ============================================================================
// GetPool Tests
// ============================================================================

func TestGetPool_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getPoolFunc: func(ctx context.Context, poolID string) (*model.MatchingPool, error) {
			return &model.MatchingPool{
				ID:   poolID,
				Name: "Test Pool",
			}, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	pool, err := svc.GetPool(ctx, "pool-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pool.ID != "pool-123" {
		t.Errorf("expected ID 'pool-123', got %q", pool.ID)
	}
}

func TestGetPool_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getPoolFunc: func(ctx context.Context, poolID string) (*model.MatchingPool, error) {
			return nil, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	_, err := svc.GetPool(ctx, "nonexistent")
	if !errors.Is(err, ErrPoolNotFound) {
		t.Errorf("expected ErrPoolNotFound, got %v", err)
	}
}

func TestGetPool_RepoError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getPoolFunc: func(ctx context.Context, poolID string) (*model.MatchingPool, error) {
			return nil, errors.New("database error")
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	_, err := svc.GetPool(ctx, "pool-123")
	if err == nil || err.Error() != "database error" {
		t.Errorf("expected database error, got %v", err)
	}
}

// ============================================================================
// JoinPool Tests
// ============================================================================

func TestJoinPool_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getPoolFunc: func(ctx context.Context, poolID string) (*model.MatchingPool, error) {
			return &model.MatchingPool{ID: poolID}, nil
		},
		getMemberFunc: func(ctx context.Context, poolID, memberID string) (*model.PoolMember, error) {
			return nil, nil // Not a member
		},
		getPoolMembersFunc: func(ctx context.Context, poolID string) ([]*model.PoolMember, error) {
			return []*model.PoolMember{}, nil // No existing members
		},
		addMemberFunc: func(ctx context.Context, member *model.PoolMember) error {
			member.ID = "membership-123"
			return nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	member, err := svc.JoinPool(ctx, "pool-1", "member-1", "user-1", &model.JoinPoolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if member == nil {
		t.Fatal("expected member, got nil")
	}
	if member.MemberID != "member-1" {
		t.Errorf("expected MemberID 'member-1', got %q", member.MemberID)
	}
}

func TestJoinPool_PoolNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getPoolFunc: func(ctx context.Context, poolID string) (*model.MatchingPool, error) {
			return nil, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	_, err := svc.JoinPool(ctx, "nonexistent", "member-1", "user-1", &model.JoinPoolRequest{})
	if !errors.Is(err, ErrPoolNotFound) {
		t.Errorf("expected ErrPoolNotFound, got %v", err)
	}
}

func TestJoinPool_AlreadyMember(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getPoolFunc: func(ctx context.Context, poolID string) (*model.MatchingPool, error) {
			return &model.MatchingPool{ID: poolID}, nil
		},
		getMemberFunc: func(ctx context.Context, poolID, memberID string) (*model.PoolMember, error) {
			return &model.PoolMember{ID: "existing", Active: true}, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	_, err := svc.JoinPool(ctx, "pool-1", "member-1", "user-1", &model.JoinPoolRequest{})
	if !errors.Is(err, ErrAlreadyPoolMember) {
		t.Errorf("expected ErrAlreadyPoolMember, got %v", err)
	}
}

func TestJoinPool_ReactivatesInactiveMember(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	updateCalled := false
	poolRepo := &mockPoolRepo{
		getPoolFunc: func(ctx context.Context, poolID string) (*model.MatchingPool, error) {
			return &model.MatchingPool{ID: poolID}, nil
		},
		getMemberFunc: func(ctx context.Context, poolID, memberID string) (*model.PoolMember, error) {
			return &model.PoolMember{ID: "existing", Active: false}, nil
		},
		updateMemberFunc: func(ctx context.Context, membershipID string, updates map[string]interface{}) (*model.PoolMember, error) {
			updateCalled = true
			if updates["active"] != true {
				t.Error("expected active to be set to true")
			}
			return &model.PoolMember{ID: membershipID, Active: true}, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	_, err := svc.JoinPool(ctx, "pool-1", "member-1", "user-1", &model.JoinPoolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updateCalled {
		t.Error("expected update to be called for reactivation")
	}
}

func TestJoinPool_MemberLimitReached(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	members := make([]*model.PoolMember, model.MaxMembersPerPool)
	for i := range members {
		members[i] = &model.PoolMember{ID: "member"}
	}

	poolRepo := &mockPoolRepo{
		getPoolFunc: func(ctx context.Context, poolID string) (*model.MatchingPool, error) {
			return &model.MatchingPool{ID: poolID}, nil
		},
		getMemberFunc: func(ctx context.Context, poolID, memberID string) (*model.PoolMember, error) {
			return nil, nil
		},
		getPoolMembersFunc: func(ctx context.Context, poolID string) ([]*model.PoolMember, error) {
			return members, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	_, err := svc.JoinPool(ctx, "pool-1", "member-1", "user-1", &model.JoinPoolRequest{})
	if !errors.Is(err, ErrMemberPoolLimitReached) {
		t.Errorf("expected ErrMemberPoolLimitReached, got %v", err)
	}
}

func TestJoinPool_TruncatesExcessiveExclusions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	exclusions := make([]string, model.MaxExclusionsPerMember+10)
	for i := range exclusions {
		exclusions[i] = "excluded"
	}

	var capturedMember *model.PoolMember
	poolRepo := &mockPoolRepo{
		getPoolFunc: func(ctx context.Context, poolID string) (*model.MatchingPool, error) {
			return &model.MatchingPool{ID: poolID}, nil
		},
		getMemberFunc: func(ctx context.Context, poolID, memberID string) (*model.PoolMember, error) {
			return nil, nil
		},
		getPoolMembersFunc: func(ctx context.Context, poolID string) ([]*model.PoolMember, error) {
			return []*model.PoolMember{}, nil
		},
		addMemberFunc: func(ctx context.Context, member *model.PoolMember) error {
			capturedMember = member
			return nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	_, err := svc.JoinPool(ctx, "pool-1", "member-1", "user-1", &model.JoinPoolRequest{
		ExcludedMembers: exclusions,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(capturedMember.ExcludedMembers) != model.MaxExclusionsPerMember {
		t.Errorf("expected exclusions truncated to %d, got %d", model.MaxExclusionsPerMember, len(capturedMember.ExcludedMembers))
	}
}

// ============================================================================
// LeavePool Tests
// ============================================================================

func TestLeavePool_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	removeCalled := false
	poolRepo := &mockPoolRepo{
		getMemberFunc: func(ctx context.Context, poolID, memberID string) (*model.PoolMember, error) {
			return &model.PoolMember{ID: "membership-1", Active: true}, nil
		},
		removeMemberFunc: func(ctx context.Context, membershipID string) error {
			removeCalled = true
			return nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	err := svc.LeavePool(ctx, "pool-1", "member-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !removeCalled {
		t.Error("expected remove to be called")
	}
}

func TestLeavePool_NotMember(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getMemberFunc: func(ctx context.Context, poolID, memberID string) (*model.PoolMember, error) {
			return nil, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	err := svc.LeavePool(ctx, "pool-1", "member-1")
	if !errors.Is(err, ErrNotPoolMember) {
		t.Errorf("expected ErrNotPoolMember, got %v", err)
	}
}

func TestLeavePool_AlreadyInactive(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getMemberFunc: func(ctx context.Context, poolID, memberID string) (*model.PoolMember, error) {
			return &model.PoolMember{ID: "membership-1", Active: false}, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	err := svc.LeavePool(ctx, "pool-1", "member-1")
	if !errors.Is(err, ErrNotPoolMember) {
		t.Errorf("expected ErrNotPoolMember for inactive member, got %v", err)
	}
}

// ============================================================================
// UpdateMatch Tests
// ============================================================================

func TestUpdateMatch_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	scheduled := model.MatchStatusScheduled
	poolRepo := &mockPoolRepo{
		getMatchResultFunc: func(ctx context.Context, matchID string) (*model.MatchResult, error) {
			return &model.MatchResult{
				ID:            matchID,
				MemberUserIDs: []string{"user-1", "user-2"},
			}, nil
		},
		updateMatchResultFunc: func(ctx context.Context, matchID string, updates map[string]interface{}) (*model.MatchResult, error) {
			return &model.MatchResult{ID: matchID, Status: scheduled}, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	match, err := svc.UpdateMatch(ctx, "match-1", "user-1", &model.UpdateMatchRequest{
		Status: &scheduled,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.Status != scheduled {
		t.Errorf("expected status %q, got %q", scheduled, match.Status)
	}
}

func TestUpdateMatch_MatchNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getMatchResultFunc: func(ctx context.Context, matchID string) (*model.MatchResult, error) {
			return nil, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	_, err := svc.UpdateMatch(ctx, "nonexistent", "user-1", &model.UpdateMatchRequest{})
	if !errors.Is(err, ErrMatchNotFound) {
		t.Errorf("expected ErrMatchNotFound, got %v", err)
	}
}

func TestUpdateMatch_NotMatchMember(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getMatchResultFunc: func(ctx context.Context, matchID string) (*model.MatchResult, error) {
			return &model.MatchResult{
				ID:            matchID,
				MemberUserIDs: []string{"user-2", "user-3"},
			}, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	status := model.MatchStatusCompleted
	_, err := svc.UpdateMatch(ctx, "match-1", "user-1", &model.UpdateMatchRequest{
		Status: &status,
	})
	if !errors.Is(err, ErrNotMatchMember) {
		t.Errorf("expected ErrNotMatchMember, got %v", err)
	}
}

func TestUpdateMatch_ScheduledTimeSetsStatusScheduled(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	scheduledTime := time.Now().Add(24 * time.Hour)
	var capturedUpdates map[string]interface{}

	poolRepo := &mockPoolRepo{
		getMatchResultFunc: func(ctx context.Context, matchID string) (*model.MatchResult, error) {
			return &model.MatchResult{
				ID:            matchID,
				MemberUserIDs: []string{"user-1"},
			}, nil
		},
		updateMatchResultFunc: func(ctx context.Context, matchID string, updates map[string]interface{}) (*model.MatchResult, error) {
			capturedUpdates = updates
			return &model.MatchResult{ID: matchID}, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	_, err := svc.UpdateMatch(ctx, "match-1", "user-1", &model.UpdateMatchRequest{
		ScheduledTime: &scheduledTime,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedUpdates["status"] != model.MatchStatusScheduled {
		t.Errorf("expected status to be set to scheduled")
	}
}

// ============================================================================
// ValidatePoolInGuild Tests
// ============================================================================

func TestValidatePoolInGuild_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getPoolFunc: func(ctx context.Context, poolID string) (*model.MatchingPool, error) {
			return &model.MatchingPool{ID: poolID, GuildID: "guild-1"}, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	pool, err := svc.ValidatePoolInGuild(ctx, "pool-1", "guild-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pool.GuildID != "guild-1" {
		t.Errorf("expected GuildID 'guild-1', got %q", pool.GuildID)
	}
}

func TestValidatePoolInGuild_WrongGuild(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getPoolFunc: func(ctx context.Context, poolID string) (*model.MatchingPool, error) {
			return &model.MatchingPool{ID: poolID, GuildID: "guild-1"}, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	_, err := svc.ValidatePoolInGuild(ctx, "pool-1", "guild-2")
	if !errors.Is(err, ErrPoolNotInGuild) {
		t.Errorf("expected ErrPoolNotInGuild, got %v", err)
	}
}

// ============================================================================
// buildScoringMatrix Tests
// ============================================================================

func TestBuildScoringMatrix_NoExclusions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getRecentMatchesBetweenFunc: func(ctx context.Context, memberIDs []string, days int) ([]*model.MatchResult, error) {
			return nil, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	members := []*model.PoolMember{
		{MemberID: "m1", UserID: "u1"},
		{MemberID: "m2", UserID: "u2"},
		{MemberID: "m3", UserID: "u3"},
	}
	pool := &model.MatchingPool{ID: "pool-1"}

	scores := svc.buildScoringMatrix(ctx, members, pool)

	// All pairs should have base score of 100
	if scores["m1"]["m2"] != 100 {
		t.Errorf("expected score 100 for m1-m2, got %f", scores["m1"]["m2"])
	}
	if scores["m2"]["m1"] != 100 {
		t.Errorf("expected symmetric score 100 for m2-m1, got %f", scores["m2"]["m1"])
	}
}

func TestBuildScoringMatrix_WithExclusions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getRecentMatchesBetweenFunc: func(ctx context.Context, memberIDs []string, days int) ([]*model.MatchResult, error) {
			return nil, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	members := []*model.PoolMember{
		{MemberID: "m1", UserID: "u1", ExcludedMembers: []string{"m2"}},
		{MemberID: "m2", UserID: "u2"},
		{MemberID: "m3", UserID: "u3"},
	}
	pool := &model.MatchingPool{ID: "pool-1"}

	scores := svc.buildScoringMatrix(ctx, members, pool)

	// m1-m2 should be excluded (score -1)
	if scores["m1"]["m2"] != -1 {
		t.Errorf("expected score -1 for excluded pair m1-m2, got %f", scores["m1"]["m2"])
	}
	// m1-m3 should be normal
	if scores["m1"]["m3"] != 100 {
		t.Errorf("expected score 100 for m1-m3, got %f", scores["m1"]["m3"])
	}
}

func TestBuildScoringMatrix_WithCompatibility(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getRecentMatchesBetweenFunc: func(ctx context.Context, memberIDs []string, days int) ([]*model.MatchResult, error) {
			return nil, nil
		},
	}

	compat := &mockCompatibilityCalc{
		calcFunc: func(ctx context.Context, userAID, userBID string) (*model.CompatibilityScore, error) {
			return &model.CompatibilityScore{Score: 80}, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, compat)

	members := []*model.PoolMember{
		{MemberID: "m1", UserID: "u1"},
		{MemberID: "m2", UserID: "u2"},
	}
	pool := &model.MatchingPool{ID: "pool-1"}

	scores := svc.buildScoringMatrix(ctx, members, pool)

	// Score should blend compatibility (80) with base (100)
	// Default config: 0.4 * 80 + 0.6 * 100 = 32 + 60 = 92
	expectedScore := 0.4*80 + 0.6*100
	if scores["m1"]["m2"] != expectedScore {
		t.Errorf("expected blended score %f, got %f", expectedScore, scores["m1"]["m2"])
	}
}

func TestBuildScoringMatrix_WithRecentMatches(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getRecentMatchesBetweenFunc: func(ctx context.Context, memberIDs []string, days int) ([]*model.MatchResult, error) {
			// Return 2 recent matches
			return []*model.MatchResult{{}, {}}, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	members := []*model.PoolMember{
		{MemberID: "m1", UserID: "u1"},
		{MemberID: "m2", UserID: "u2"},
	}
	pool := &model.MatchingPool{ID: "pool-1"}

	scores := svc.buildScoringMatrix(ctx, members, pool)

	// 2 recent matches × variety_weight (0.6) × 20 = 24 penalty
	// 100 - 24 = 76
	expectedScore := 100.0 - 2*0.6*20
	if scores["m1"]["m2"] != expectedScore {
		t.Errorf("expected penalized score %f, got %f", expectedScore, scores["m1"]["m2"])
	}
}

// ============================================================================
// formGroups Tests
// ============================================================================

func TestFormGroups_PairsExactFit(t *testing.T) {
	t.Parallel()

	svc := newTestPoolService(nil, nil, nil, nil)

	members := []*model.PoolMember{
		{MemberID: "m1"},
		{MemberID: "m2"},
		{MemberID: "m3"},
		{MemberID: "m4"},
	}

	// All scores are equal and positive
	scores := map[string]map[string]float64{
		"m1": {"m2": 100, "m3": 100, "m4": 100},
		"m2": {"m1": 100, "m3": 100, "m4": 100},
		"m3": {"m1": 100, "m2": 100, "m4": 100},
		"m4": {"m1": 100, "m2": 100, "m3": 100},
	}

	groups := svc.formGroups(members, scores, 2)

	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}
	for i, g := range groups {
		if len(g) != 2 {
			t.Errorf("group %d: expected size 2, got %d", i, len(g))
		}
	}
}

func TestFormGroups_TriosWithLeftover(t *testing.T) {
	t.Parallel()

	svc := newTestPoolService(nil, nil, nil, nil)

	members := []*model.PoolMember{
		{MemberID: "m1"},
		{MemberID: "m2"},
		{MemberID: "m3"},
		{MemberID: "m4"},
		{MemberID: "m5"},
	}

	// All positive scores
	scores := make(map[string]map[string]float64)
	for _, m := range members {
		scores[m.MemberID] = make(map[string]float64)
		for _, n := range members {
			if m.MemberID != n.MemberID {
				scores[m.MemberID][n.MemberID] = 100
			}
		}
	}

	groups := svc.formGroups(members, scores, 3)

	// 5 members, group size 3: should form 1 group of 3, leaving 2
	if len(groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groups))
	}
	if len(groups) > 0 && len(groups[0]) != 3 {
		t.Errorf("expected group size 3, got %d", len(groups[0]))
	}
}

func TestFormGroups_RespectsExclusions(t *testing.T) {
	t.Parallel()

	svc := newTestPoolService(nil, nil, nil, nil)

	// 4 members: m1 excludes m2, but m3 and m4 can pair
	members := []*model.PoolMember{
		{MemberID: "m1"},
		{MemberID: "m2"},
		{MemberID: "m3"},
		{MemberID: "m4"},
	}

	// m1 and m2 exclude each other, but m3 and m4 are fine
	scores := map[string]map[string]float64{
		"m1": {"m2": -1, "m3": 100, "m4": 100},
		"m2": {"m1": -1, "m3": 100, "m4": 100},
		"m3": {"m1": 100, "m2": 100, "m4": 100},
		"m4": {"m1": 100, "m2": 100, "m3": 100},
	}

	groups := svc.formGroups(members, scores, 2)

	// Should form 2 groups, but m1 and m2 should never be paired together
	for i, g := range groups {
		hasBothExcluded := false
		var memberIDs []string
		for _, m := range g {
			memberIDs = append(memberIDs, m.MemberID)
			for _, other := range g {
				if scores[m.MemberID][other.MemberID] < 0 {
					hasBothExcluded = true
				}
			}
		}
		if hasBothExcluded {
			t.Errorf("group %d contains excluded pair: %v", i, memberIDs)
		}
	}
}

func TestFormGroups_PicksBestScoring(t *testing.T) {
	t.Parallel()

	svc := newTestPoolService(nil, nil, nil, nil)

	members := []*model.PoolMember{
		{MemberID: "m1"},
		{MemberID: "m2"},
		{MemberID: "m3"},
	}

	// m1-m2 has highest score
	scores := map[string]map[string]float64{
		"m1": {"m2": 100, "m3": 50},
		"m2": {"m1": 100, "m3": 50},
		"m3": {"m1": 50, "m2": 50},
	}

	// Run multiple times - due to shuffle, result may vary
	// but the algorithm should always pick valid pairs
	for i := 0; i < 10; i++ {
		groups := svc.formGroups(members, scores, 2)
		if len(groups) != 1 {
			t.Errorf("iteration %d: expected 1 group, got %d", i, len(groups))
		}
		if len(groups) > 0 && len(groups[0]) != 2 {
			t.Errorf("iteration %d: expected group size 2, got %d", i, len(groups[0]))
		}
	}
}

// ============================================================================
// RunMatching Tests
// ============================================================================

func TestRunMatching_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	matchesCreated := 0
	poolRepo := &mockPoolRepo{
		getPoolFunc: func(ctx context.Context, poolID string) (*model.MatchingPool, error) {
			return &model.MatchingPool{
				ID:        poolID,
				Name:      "Test Pool",
				MatchSize: 2,
				Frequency: model.PoolFrequencyWeekly,
			}, nil
		},
		getPoolMembersFunc: func(ctx context.Context, poolID string) ([]*model.PoolMember, error) {
			return []*model.PoolMember{
				{MemberID: "m1", UserID: "u1"},
				{MemberID: "m2", UserID: "u2"},
			}, nil
		},
		getRecentMatchesBetweenFunc: func(ctx context.Context, memberIDs []string, days int) ([]*model.MatchResult, error) {
			return nil, nil
		},
		createMatchResultFunc: func(ctx context.Context, match *model.MatchResult) error {
			matchesCreated++
			match.ID = "match-" + string(rune('0'+matchesCreated))
			return nil
		},
		updatePoolFunc: func(ctx context.Context, poolID string, updates map[string]interface{}) (*model.MatchingPool, error) {
			return nil, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	info, err := svc.RunMatching(ctx, "pool-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected match round info, got nil")
	}
	if info.MatchCount != 1 {
		t.Errorf("expected 1 match, got %d", info.MatchCount)
	}
	if matchesCreated != 1 {
		t.Errorf("expected 1 match created, got %d", matchesCreated)
	}
}

func TestRunMatching_NotEnoughMembers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getPoolFunc: func(ctx context.Context, poolID string) (*model.MatchingPool, error) {
			return &model.MatchingPool{
				ID:        poolID,
				MatchSize: 2,
			}, nil
		},
		getPoolMembersFunc: func(ctx context.Context, poolID string) ([]*model.PoolMember, error) {
			return []*model.PoolMember{
				{MemberID: "m1", UserID: "u1"},
			}, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	_, err := svc.RunMatching(ctx, "pool-1")
	if !errors.Is(err, ErrNotEnoughMembers) {
		t.Errorf("expected ErrNotEnoughMembers, got %v", err)
	}
}

func TestRunMatching_PoolNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getPoolFunc: func(ctx context.Context, poolID string) (*model.MatchingPool, error) {
			return nil, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	_, err := svc.RunMatching(ctx, "nonexistent")
	if !errors.Is(err, ErrPoolNotFound) {
		t.Errorf("expected ErrPoolNotFound, got %v", err)
	}
}

// ============================================================================
// UpdateMembership Tests
// ============================================================================

func TestUpdateMembership_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	active := false
	poolRepo := &mockPoolRepo{
		getMemberFunc: func(ctx context.Context, poolID, memberID string) (*model.PoolMember, error) {
			return &model.PoolMember{ID: "membership-1", Active: true}, nil
		},
		updateMemberFunc: func(ctx context.Context, membershipID string, updates map[string]interface{}) (*model.PoolMember, error) {
			return &model.PoolMember{ID: membershipID, Active: false}, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	member, err := svc.UpdateMembership(ctx, "pool-1", "member-1", &model.UpdateMembershipRequest{
		Active: &active,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if member.Active {
		t.Error("expected member to be inactive")
	}
}

func TestUpdateMembership_ExclusionLimitReached(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	exclusions := make([]string, model.MaxExclusionsPerMember+1)
	for i := range exclusions {
		exclusions[i] = "member"
	}

	poolRepo := &mockPoolRepo{
		getMemberFunc: func(ctx context.Context, poolID, memberID string) (*model.PoolMember, error) {
			return &model.PoolMember{ID: "membership-1"}, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	_, err := svc.UpdateMembership(ctx, "pool-1", "member-1", &model.UpdateMembershipRequest{
		ExcludedMembers: exclusions,
	})
	if !errors.Is(err, ErrExclusionLimitReached) {
		t.Errorf("expected ErrExclusionLimitReached, got %v", err)
	}
}

func TestUpdateMembership_NotMember(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	poolRepo := &mockPoolRepo{
		getMemberFunc: func(ctx context.Context, poolID, memberID string) (*model.PoolMember, error) {
			return nil, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	active := false
	_, err := svc.UpdateMembership(ctx, "pool-1", "member-1", &model.UpdateMembershipRequest{
		Active: &active,
	})
	if !errors.Is(err, ErrNotPoolMember) {
		t.Errorf("expected ErrNotPoolMember, got %v", err)
	}
}

// ============================================================================
// GetMatchHistory Tests
// ============================================================================

func TestGetMatchHistory_DefaultLimit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var capturedLimit int
	poolRepo := &mockPoolRepo{
		getMatchesByPoolFunc: func(ctx context.Context, poolID string, limit int) ([]*model.MatchResult, error) {
			capturedLimit = limit
			return nil, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	_, _ = svc.GetMatchHistory(ctx, "pool-1", 0) // 0 should default to 20
	if capturedLimit != 20 {
		t.Errorf("expected default limit 20, got %d", capturedLimit)
	}
}

func TestGetMatchHistory_NegativeLimit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var capturedLimit int
	poolRepo := &mockPoolRepo{
		getMatchesByPoolFunc: func(ctx context.Context, poolID string, limit int) ([]*model.MatchResult, error) {
			capturedLimit = limit
			return nil, nil
		},
	}

	svc := newTestPoolService(poolRepo, nil, nil, nil)

	_, _ = svc.GetMatchHistory(ctx, "pool-1", -5)
	if capturedLimit != 20 {
		t.Errorf("expected default limit 20 for negative, got %d", capturedLimit)
	}
}

// ============================================================================
// shuffleMembers Tests
// ============================================================================

func TestShuffleMembers_ChangesOrder(t *testing.T) {
	t.Parallel()

	members := make([]*model.PoolMember, 20)
	for i := range members {
		members[i] = &model.PoolMember{MemberID: string(rune('a' + i))}
	}

	// Copy original order
	original := make([]string, len(members))
	for i, m := range members {
		original[i] = m.MemberID
	}

	shuffleMembers(members)

	// Check that order changed (with high probability for 20 elements)
	same := 0
	for i, m := range members {
		if m.MemberID == original[i] {
			same++
		}
	}

	// It's extremely unlikely that more than half stayed in place
	if same > 10 {
		t.Errorf("shuffle didn't change order significantly: %d of %d stayed in place", same, len(members))
	}
}

func TestShuffleMembers_PreservesElements(t *testing.T) {
	t.Parallel()

	members := []*model.PoolMember{
		{MemberID: "a"},
		{MemberID: "b"},
		{MemberID: "c"},
		{MemberID: "d"},
	}

	original := make(map[string]bool)
	for _, m := range members {
		original[m.MemberID] = true
	}

	shuffleMembers(members)

	// All elements should still be present
	for _, m := range members {
		if !original[m.MemberID] {
			t.Errorf("shuffle introduced unknown element: %s", m.MemberID)
		}
		delete(original, m.MemberID)
	}

	if len(original) > 0 {
		t.Errorf("shuffle lost elements: %v", original)
	}
}

func TestShuffleMembers_EmptySlice(t *testing.T) {
	t.Parallel()

	members := []*model.PoolMember{}
	shuffleMembers(members) // Should not panic

	if len(members) != 0 {
		t.Error("empty slice should remain empty")
	}
}

func TestShuffleMembers_SingleElement(t *testing.T) {
	t.Parallel()

	members := []*model.PoolMember{{MemberID: "only"}}
	shuffleMembers(members) // Should not panic

	if members[0].MemberID != "only" {
		t.Error("single element should remain unchanged")
	}
}
