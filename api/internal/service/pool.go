package service

import (
	"context"
	"errors"
	"math"
	"sort"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

var (
	ErrPoolNotFound           = errors.New("pool not found")
	ErrPoolLimitReached       = errors.New("maximum pools per guild reached")
	ErrMemberPoolLimitReached = errors.New("maximum members per pool reached")
	ErrAlreadyPoolMember      = errors.New("already a member of this pool")
	ErrNotPoolMember          = errors.New("not a member of this pool")
	ErrPoolNotInGuild         = errors.New("pool does not belong to this guild")
	ErrInvalidMatchSize       = errors.New("match size must be between 2 and 6")
	ErrInvalidFrequency       = errors.New("invalid frequency")
	ErrMatchNotFound          = errors.New("match not found")
	ErrNotMatchMember         = errors.New("not a member of this match")
	ErrExclusionLimitReached  = errors.New("maximum exclusions reached")
	ErrNotEnoughMembers       = errors.New("not enough active members to create matches")
)

// PoolRepository defines the interface for pool storage
type PoolRepository interface {
	CreatePool(ctx context.Context, pool *model.MatchingPool) error
	GetPool(ctx context.Context, poolID string) (*model.MatchingPool, error)
	GetPoolsByGuild(ctx context.Context, guildID string) ([]*model.MatchingPool, error)
	UpdatePool(ctx context.Context, poolID string, updates map[string]interface{}) (*model.MatchingPool, error)
	DeletePool(ctx context.Context, poolID string) error
	CountPoolsByGuild(ctx context.Context, guildID string) (int, error)

	AddMember(ctx context.Context, member *model.PoolMember) error
	GetMember(ctx context.Context, poolID, memberID string) (*model.PoolMember, error)
	GetMemberByUser(ctx context.Context, poolID, userID string) (*model.PoolMember, error)
	GetPoolMembers(ctx context.Context, poolID string) ([]*model.PoolMember, error)
	UpdateMember(ctx context.Context, membershipID string, updates map[string]interface{}) (*model.PoolMember, error)
	RemoveMember(ctx context.Context, membershipID string) error
	GetUserPoolMemberships(ctx context.Context, userID string) ([]*model.PoolMember, error)

	CreateMatchResult(ctx context.Context, match *model.MatchResult) error
	GetMatchResult(ctx context.Context, matchID string) (*model.MatchResult, error)
	GetMatchesByPool(ctx context.Context, poolID string, limit int) ([]*model.MatchResult, error)
	GetMatchesByRound(ctx context.Context, poolID, round string) ([]*model.MatchResult, error)
	GetUserPendingMatches(ctx context.Context, userID string) ([]*model.MatchResult, error)
	GetRecentMatchesBetween(ctx context.Context, memberIDs []string, days int) ([]*model.MatchResult, error)
	UpdateMatchResult(ctx context.Context, matchID string, updates map[string]interface{}) (*model.MatchResult, error)
	GetPoolsDueForMatching(ctx context.Context) ([]*model.MatchingPool, error)
	GetPoolStats(ctx context.Context, poolID string) (*model.PoolStats, error)
	// Nudge-related
	GetStaleMatches(ctx context.Context, cutoff time.Time, status string) ([]*model.MatchResult, error)
}

// CompatibilityCalculator interface for optional compatibility scoring
type CompatibilityCalculator interface {
	CalculateCompatibility(ctx context.Context, userAID, userBID string) (*model.CompatibilityScore, error)
}

// PoolService handles matching pool business logic
type PoolService struct {
	poolRepo      PoolRepository
	guildRepo     GuildRepository
	memberRepo    MemberRepository
	compatibility CompatibilityCalculator
	config        model.MatchingConfig
}

// PoolServiceConfig holds configuration for the pool service
type PoolServiceConfig struct {
	PoolRepo      PoolRepository
	GuildRepo     GuildRepository
	MemberRepo    MemberRepository
	Compatibility CompatibilityCalculator // Optional
	Config        *model.MatchingConfig   // Optional, uses defaults if nil
}

// NewPoolService creates a new pool service
func NewPoolService(cfg PoolServiceConfig) *PoolService {
	config := model.DefaultMatchingConfig
	if cfg.Config != nil {
		config = *cfg.Config
	}
	return &PoolService{
		poolRepo:      cfg.PoolRepo,
		guildRepo:     cfg.GuildRepo,
		memberRepo:    cfg.MemberRepo,
		compatibility: cfg.Compatibility,
		config:        config,
	}
}

// CreatePool creates a new matching pool in a guild
func (s *PoolService) CreatePool(ctx context.Context, guildID string, req *model.CreatePoolRequest, creatorMemberID string) (*model.MatchingPool, error) {
	// Validate frequency
	if !isValidFrequency(req.Frequency) {
		return nil, ErrInvalidFrequency
	}

	// Validate match size
	matchSize := req.MatchSize
	if matchSize == 0 {
		matchSize = 2 // Default to pairs
	}
	if matchSize < model.MinMatchSize || matchSize > model.MaxMatchSize {
		return nil, ErrInvalidMatchSize
	}

	// Check pool limit for guild
	count, err := s.poolRepo.CountPoolsByGuild(ctx, guildID)
	if err != nil {
		return nil, err
	}
	if count >= model.MaxPoolsPerGuild {
		return nil, ErrPoolLimitReached
	}

	// Validate name length
	if len(req.Name) > model.MaxPoolNameLength {
		req.Name = req.Name[:model.MaxPoolNameLength]
	}
	if req.Description != nil && len(*req.Description) > model.MaxPoolDescLength {
		desc := (*req.Description)[:model.MaxPoolDescLength]
		req.Description = &desc
	}
	if req.ActivitySuggestion != nil && len(*req.ActivitySuggestion) > model.MaxActivitySuggLength {
		sugg := (*req.ActivitySuggestion)[:model.MaxActivitySuggLength]
		req.ActivitySuggestion = &sugg
	}

	// Calculate first match date
	nextMatch := model.GetNextMatchDate(req.Frequency, time.Now())

	pool := &model.MatchingPool{
		GuildID:            guildID,
		Name:               req.Name,
		Description:        req.Description,
		Frequency:          req.Frequency,
		MatchSize:          matchSize,
		ActivitySuggestion: req.ActivitySuggestion,
		NextMatchOn:        nextMatch,
		Active:             true,
		CreatedBy:          creatorMemberID,
	}

	if err := s.poolRepo.CreatePool(ctx, pool); err != nil {
		return nil, err
	}

	return pool, nil
}

// GetPool retrieves a pool by ID
func (s *PoolService) GetPool(ctx context.Context, poolID string) (*model.MatchingPool, error) {
	pool, err := s.poolRepo.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}
	if pool == nil {
		return nil, ErrPoolNotFound
	}
	return pool, nil
}

// GetPoolsByGuild retrieves all pools for a guild
func (s *PoolService) GetPoolsByGuild(ctx context.Context, guildID string) ([]*model.MatchingPool, error) {
	return s.poolRepo.GetPoolsByGuild(ctx, guildID)
}

// UpdatePool updates a pool
func (s *PoolService) UpdatePool(ctx context.Context, poolID string, req *model.UpdatePoolRequest) (*model.MatchingPool, error) {
	pool, err := s.poolRepo.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}
	if pool == nil {
		return nil, ErrPoolNotFound
	}

	updates := make(map[string]interface{})

	if req.Name != nil {
		name := *req.Name
		if len(name) > model.MaxPoolNameLength {
			name = name[:model.MaxPoolNameLength]
		}
		updates["name"] = name
	}
	if req.Description != nil {
		desc := *req.Description
		if len(desc) > model.MaxPoolDescLength {
			desc = desc[:model.MaxPoolDescLength]
		}
		updates["description"] = desc
	}
	if req.Frequency != nil {
		if !isValidFrequency(*req.Frequency) {
			return nil, ErrInvalidFrequency
		}
		updates["frequency"] = *req.Frequency
		// Recalculate next match date
		updates["next_match_on"] = model.GetNextMatchDate(*req.Frequency, time.Now())
	}
	if req.MatchSize != nil {
		if *req.MatchSize < model.MinMatchSize || *req.MatchSize > model.MaxMatchSize {
			return nil, ErrInvalidMatchSize
		}
		updates["match_size"] = *req.MatchSize
	}
	if req.ActivitySuggestion != nil {
		sugg := *req.ActivitySuggestion
		if len(sugg) > model.MaxActivitySuggLength {
			sugg = sugg[:model.MaxActivitySuggLength]
		}
		updates["activity_suggestion"] = sugg
	}
	if req.Active != nil {
		updates["active"] = *req.Active
	}

	if len(updates) == 0 {
		return pool, nil
	}

	return s.poolRepo.UpdatePool(ctx, poolID, updates)
}

// DeletePool deletes a pool and all associated data
func (s *PoolService) DeletePool(ctx context.Context, poolID string) error {
	pool, err := s.poolRepo.GetPool(ctx, poolID)
	if err != nil {
		return err
	}
	if pool == nil {
		return ErrPoolNotFound
	}
	return s.poolRepo.DeletePool(ctx, poolID)
}

// JoinPool adds a user to a pool
func (s *PoolService) JoinPool(ctx context.Context, poolID, memberID, userID string, req *model.JoinPoolRequest) (*model.PoolMember, error) {
	pool, err := s.poolRepo.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}
	if pool == nil {
		return nil, ErrPoolNotFound
	}

	// Check if already a member
	existing, err := s.poolRepo.GetMember(ctx, poolID, memberID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if existing.Active {
			return nil, ErrAlreadyPoolMember
		}
		// Reactivate membership
		updates := map[string]interface{}{
			"active":           true,
			"excluded_members": req.ExcludedMembers,
		}
		return s.poolRepo.UpdateMember(ctx, existing.ID, updates)
	}

	// Check member limit
	members, err := s.poolRepo.GetPoolMembers(ctx, poolID)
	if err != nil {
		return nil, err
	}
	if len(members) >= model.MaxMembersPerPool {
		return nil, ErrMemberPoolLimitReached
	}

	// Validate exclusions
	if len(req.ExcludedMembers) > model.MaxExclusionsPerMember {
		req.ExcludedMembers = req.ExcludedMembers[:model.MaxExclusionsPerMember]
	}

	member := &model.PoolMember{
		PoolID:          poolID,
		MemberID:        memberID,
		UserID:          userID,
		Active:          true,
		ExcludedMembers: req.ExcludedMembers,
	}

	if err := s.poolRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// LeavePool removes a user from a pool
func (s *PoolService) LeavePool(ctx context.Context, poolID, memberID string) error {
	member, err := s.poolRepo.GetMember(ctx, poolID, memberID)
	if err != nil {
		return err
	}
	if member == nil || !member.Active {
		return ErrNotPoolMember
	}
	return s.poolRepo.RemoveMember(ctx, member.ID)
}

// UpdateMembership updates a user's pool membership settings
func (s *PoolService) UpdateMembership(ctx context.Context, poolID, memberID string, req *model.UpdateMembershipRequest) (*model.PoolMember, error) {
	member, err := s.poolRepo.GetMember(ctx, poolID, memberID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotPoolMember
	}

	updates := make(map[string]interface{})
	if req.Active != nil {
		updates["active"] = *req.Active
	}
	if req.ExcludedMembers != nil {
		if len(req.ExcludedMembers) > model.MaxExclusionsPerMember {
			return nil, ErrExclusionLimitReached
		}
		updates["excluded_members"] = req.ExcludedMembers
	}

	if len(updates) == 0 {
		return member, nil
	}

	return s.poolRepo.UpdateMember(ctx, member.ID, updates)
}

// GetPoolMembers retrieves all members of a pool
func (s *PoolService) GetPoolMembers(ctx context.Context, poolID string) ([]*model.PoolMember, error) {
	return s.poolRepo.GetPoolMembers(ctx, poolID)
}

// GetUserMemberships retrieves all pools a user belongs to
func (s *PoolService) GetUserMemberships(ctx context.Context, userID string) ([]*model.PoolMember, error) {
	return s.poolRepo.GetUserPoolMemberships(ctx, userID)
}

// GetPoolWithMembers retrieves a pool with its member list
func (s *PoolService) GetPoolWithMembers(ctx context.Context, poolID string) (*model.PoolWithMembers, error) {
	pool, err := s.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}

	members, err := s.poolRepo.GetPoolMembers(ctx, poolID)
	if err != nil {
		return nil, err
	}

	memberList := make([]model.PoolMember, len(members))
	for i, m := range members {
		memberList[i] = *m
	}

	return &model.PoolWithMembers{
		Pool:    *pool,
		Members: memberList,
	}, nil
}

// GetPendingMatches retrieves pending matches for a user
func (s *PoolService) GetPendingMatches(ctx context.Context, userID string) ([]*model.PendingMatch, error) {
	matches, err := s.poolRepo.GetUserPendingMatches(ctx, userID)
	if err != nil {
		return nil, err
	}

	var pending []*model.PendingMatch
	for _, match := range matches {
		pool, err := s.poolRepo.GetPool(ctx, match.PoolID)
		if err != nil || pool == nil {
			continue
		}

		// Get guild info (would need guild repo)
		guild, err := s.guildRepo.GetByID(ctx, pool.GuildID)
		if err != nil || guild == nil {
			continue
		}

		// Get partner info
		partnerIDs := make([]string, 0)
		partnerNames := make([]string, 0)
		for _, mid := range match.Members {
			member, err := s.poolRepo.GetMember(ctx, pool.ID, mid)
			if err == nil && member != nil && member.UserID != userID {
				partnerIDs = append(partnerIDs, mid)
				if member.MemberName != nil {
					partnerNames = append(partnerNames, *member.MemberName)
				}
			}
		}

		pm := &model.PendingMatch{
			Match:        *match,
			PoolName:     pool.Name,
			GuildID:      pool.GuildID,
			GuildName:    guild.Name,
			PartnerIDs:   partnerIDs,
			PartnerNames: partnerNames,
			Suggestion:   pool.ActivitySuggestion,
			DueBy:        &pool.NextMatchOn,
		}
		pending = append(pending, pm)
	}

	return pending, nil
}

// UpdateMatch updates a match result (status, scheduled time, etc.)
func (s *PoolService) UpdateMatch(ctx context.Context, matchID, userID string, req *model.UpdateMatchRequest) (*model.MatchResult, error) {
	match, err := s.poolRepo.GetMatchResult(ctx, matchID)
	if err != nil {
		return nil, err
	}
	if match == nil {
		return nil, ErrMatchNotFound
	}

	// Verify user is part of this match
	isMember := false
	for _, uid := range match.MemberUserIDs {
		if uid == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, ErrNotMatchMember
	}

	updates := make(map[string]interface{})
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.ScheduledTime != nil {
		updates["scheduled_time"] = *req.ScheduledTime
		updates["status"] = model.MatchStatusScheduled
	}

	if len(updates) == 0 {
		return match, nil
	}

	return s.poolRepo.UpdateMatchResult(ctx, matchID, updates)
}

// GetPoolStats retrieves statistics for a pool
func (s *PoolService) GetPoolStats(ctx context.Context, poolID string) (*model.PoolStats, error) {
	return s.poolRepo.GetPoolStats(ctx, poolID)
}

// GetMatchHistory retrieves match history for a pool
func (s *PoolService) GetMatchHistory(ctx context.Context, poolID string, limit int) ([]*model.MatchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.poolRepo.GetMatchesByPool(ctx, poolID, limit)
}

// ValidatePoolInGuild checks if a pool belongs to a guild
func (s *PoolService) ValidatePoolInGuild(ctx context.Context, poolID, guildID string) (*model.MatchingPool, error) {
	pool, err := s.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}
	if pool.GuildID != guildID {
		return nil, ErrPoolNotInGuild
	}
	return pool, nil
}

// RunMatching executes the matching algorithm for a pool
func (s *PoolService) RunMatching(ctx context.Context, poolID string) (*model.MatchRoundInfo, error) {
	pool, err := s.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}

	members, err := s.poolRepo.GetPoolMembers(ctx, poolID)
	if err != nil {
		return nil, err
	}

	// Need at least match_size members
	if len(members) < pool.MatchSize {
		return nil, ErrNotEnoughMembers
	}

	// Build scoring matrix
	scores := s.buildScoringMatrix(ctx, members, pool)

	// Run matching algorithm
	groups := s.formGroups(members, scores, pool.MatchSize)

	// Create match results
	round := model.GetMatchRound(time.Now())
	var matches []model.MatchResult

	for _, group := range groups {
		memberIDs := make([]string, len(group))
		userIDs := make([]string, len(group))
		for i, m := range group {
			memberIDs[i] = m.MemberID
			userIDs[i] = m.UserID
		}

		match := &model.MatchResult{
			PoolID:        poolID,
			Members:       memberIDs,
			MemberUserIDs: userIDs,
			Status:        model.MatchStatusPending,
			MatchRound:    round,
		}

		if err := s.poolRepo.CreateMatchResult(ctx, match); err != nil {
			return nil, err
		}
		matches = append(matches, *match)
	}

	// Update pool's next match date and last match date
	now := time.Now()
	nextMatch := model.GetNextMatchDate(pool.Frequency, now)
	_, err = s.poolRepo.UpdatePool(ctx, poolID, map[string]interface{}{
		"next_match_on": nextMatch,
		"last_match_on": now,
	})
	if err != nil {
		return nil, err
	}

	return &model.MatchRoundInfo{
		PoolID:     poolID,
		PoolName:   pool.Name,
		Round:      round,
		RanOn:      now,
		MatchCount: len(matches),
		Matches:    matches,
	}, nil
}

// GetPoolsDueForMatching retrieves pools that need matching run
func (s *PoolService) GetPoolsDueForMatching(ctx context.Context) ([]*model.MatchingPool, error) {
	return s.poolRepo.GetPoolsDueForMatching(ctx)
}

// buildScoringMatrix creates a scoring matrix between all members
// Higher scores = better matches
func (s *PoolService) buildScoringMatrix(ctx context.Context, members []*model.PoolMember, pool *model.MatchingPool) map[string]map[string]float64 {
	scores := make(map[string]map[string]float64)

	for _, m := range members {
		scores[m.MemberID] = make(map[string]float64)
	}

	// Build exclusion sets for quick lookup
	exclusions := make(map[string]map[string]bool)
	for _, m := range members {
		exclusions[m.MemberID] = make(map[string]bool)
		for _, ex := range m.ExcludedMembers {
			exclusions[m.MemberID][ex] = true
		}
	}

	// Calculate scores for each pair
	for i, a := range members {
		for j, b := range members {
			if i >= j {
				continue // Only calculate upper triangle
			}

			// Start with base score
			score := 100.0

			// Check exclusions (either direction)
			if exclusions[a.MemberID][b.MemberID] || exclusions[b.MemberID][a.MemberID] {
				score = -1 // Cannot be matched
				scores[a.MemberID][b.MemberID] = score
				scores[b.MemberID][a.MemberID] = score
				continue
			}

			// Apply compatibility score if available
			if s.compatibility != nil {
				compat, err := s.compatibility.CalculateCompatibility(ctx, a.UserID, b.UserID)
				if err == nil && compat != nil {
					// Blend compatibility: weight * compat + (1-weight) * base
					score = s.config.CompatibilityWeight*compat.Score +
						(1-s.config.CompatibilityWeight)*score
				}
			}

			// Apply variety penalty for recent matches
			recentMatches, err := s.poolRepo.GetRecentMatchesBetween(ctx, []string{a.MemberID, b.MemberID}, s.config.RecencyDays)
			if err == nil && len(recentMatches) > 0 {
				// Penalize based on number of recent matches
				// Each recent match reduces score by variety_weight * 20
				penalty := float64(len(recentMatches)) * s.config.VarietyWeight * 20
				score = math.Max(0, score-penalty)
			}

			scores[a.MemberID][b.MemberID] = score
			scores[b.MemberID][a.MemberID] = score
		}
	}

	return scores
}

// formGroups uses a greedy algorithm to form groups
func (s *PoolService) formGroups(members []*model.PoolMember, scores map[string]map[string]float64, groupSize int) [][]*model.PoolMember {
	var groups [][]*model.PoolMember
	remaining := make([]*model.PoolMember, len(members))
	copy(remaining, members)

	// Shuffle to avoid bias
	shuffleMembers(remaining)

	for len(remaining) >= groupSize {
		// Pick first remaining member
		group := []*model.PoolMember{remaining[0]}
		remaining = remaining[1:]

		// Greedily add best-scoring members
		for len(group) < groupSize && len(remaining) > 0 {
			bestIdx := -1
			bestScore := -2.0

			for i, candidate := range remaining {
				// Calculate average score with current group
				avgScore := 0.0
				valid := true
				for _, member := range group {
					s := scores[member.MemberID][candidate.MemberID]
					if s < 0 {
						valid = false
						break
					}
					avgScore += s
				}

				if !valid {
					continue
				}
				avgScore /= float64(len(group))

				if avgScore > bestScore {
					bestScore = avgScore
					bestIdx = i
				}
			}

			if bestIdx < 0 {
				// No valid candidate found, skip this group
				break
			}

			group = append(group, remaining[bestIdx])
			remaining = append(remaining[:bestIdx], remaining[bestIdx+1:]...)
		}

		if len(group) == groupSize {
			groups = append(groups, group)
		} else {
			// Put incomplete group members back
			remaining = append(remaining, group...)
		}
	}

	// Handle remaining members (unmatched this round)
	// They'll have better chances next round
	return groups
}

// shuffleMembers randomly shuffles the members slice
func shuffleMembers(members []*model.PoolMember) {
	// Use a simple Fisher-Yates shuffle with timestamp seed
	n := len(members)
	seed := time.Now().UnixNano()
	for i := n - 1; i > 0; i-- {
		seed = (seed*1103515245 + 12345) % (1 << 31)
		j := int(seed) % (i + 1)
		if j < 0 {
			j = -j
		}
		members[i], members[j] = members[j], members[i]
	}
}

// isValidFrequency checks if a frequency string is valid
func isValidFrequency(freq string) bool {
	switch freq {
	case model.PoolFrequencyWeekly, model.PoolFrequencyBiweekly, model.PoolFrequencyMonthly:
		return true
	default:
		return false
	}
}

// GetRoundMatches retrieves matches for a specific round
func (s *PoolService) GetRoundMatches(ctx context.Context, poolID, round string) ([]*model.MatchResult, error) {
	return s.poolRepo.GetMatchesByRound(ctx, poolID, round)
}

// GetMatchWithDetails retrieves a match with member names populated
func (s *PoolService) GetMatchWithDetails(ctx context.Context, matchID string) (*model.MatchResult, error) {
	match, err := s.poolRepo.GetMatchResult(ctx, matchID)
	if err != nil {
		return nil, err
	}
	if match == nil {
		return nil, ErrMatchNotFound
	}

	// Populate member names
	names := make([]string, 0, len(match.Members))
	for _, memberID := range match.Members {
		// Try to get member from any pool they're in
		memberships, err := s.poolRepo.GetUserPoolMemberships(ctx, "")
		if err == nil {
			for _, m := range memberships {
				if m.MemberID == memberID && m.MemberName != nil {
					names = append(names, *m.MemberName)
					break
				}
			}
		}
	}
	match.MemberNames = names

	// Get pool name
	pool, err := s.poolRepo.GetPool(ctx, match.PoolID)
	if err == nil && pool != nil {
		match.PoolName = &pool.Name
	}

	return match, nil
}

// GetUserMatchHistory retrieves match history for a user in a pool
func (s *PoolService) GetUserMatchHistory(ctx context.Context, poolID, userID string, days int) (*model.PoolMatchHistory, error) {
	// Get member
	member, err := s.poolRepo.GetMemberByUser(ctx, poolID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotPoolMember
	}

	// Get recent matches
	matches, err := s.poolRepo.GetMatchesByPool(ctx, poolID, 50)
	if err != nil {
		return nil, err
	}

	// Filter to matches including this user
	cutoff := time.Now().AddDate(0, 0, -days)
	var userMatches []model.MatchResult
	matchCounts := make(map[string]int)

	for _, match := range matches {
		if match.CreatedOn.Before(cutoff) {
			continue
		}

		// Check if user is in this match
		isInMatch := false
		for _, mid := range match.Members {
			if mid == member.MemberID {
				isInMatch = true
			} else {
				// Count matches with other members
				matchCounts[mid]++
			}
		}

		if isInMatch {
			userMatches = append(userMatches, *match)
		}
	}

	// Sort by recency
	sort.Slice(userMatches, func(i, j int) bool {
		return userMatches[i].CreatedOn.After(userMatches[j].CreatedOn)
	})

	return &model.PoolMatchHistory{
		UserID:        userID,
		RecentMatches: userMatches,
		MatchCounts:   matchCounts,
	}, nil
}
