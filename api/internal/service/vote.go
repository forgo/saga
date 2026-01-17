package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// VoteRepository defines the interface for vote storage
type VoteRepository interface {
	Create(ctx context.Context, vote *model.Vote) error
	GetByID(ctx context.Context, id string) (*model.Vote, error)
	GetByGuild(ctx context.Context, guildID string, status *model.VoteStatus, limit, offset int) ([]*model.Vote, error)
	GetGlobalVotes(ctx context.Context, status *model.VoteStatus, limit, offset int) ([]*model.Vote, error)
	GetVotesToOpen(ctx context.Context) ([]*model.Vote, error)
	GetVotesToClose(ctx context.Context) ([]*model.Vote, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Vote, error)
	UpdateStatus(ctx context.Context, id string, status model.VoteStatus) error
	Delete(ctx context.Context, id string) error
	// Options
	CreateOption(ctx context.Context, option *model.VoteOption) error
	GetOptionByID(ctx context.Context, id string) (*model.VoteOption, error)
	GetOptionsByVote(ctx context.Context, voteID string) ([]*model.VoteOption, error)
	UpdateOption(ctx context.Context, id string, updates *model.UpdateVoteOptionRequest) (*model.VoteOption, error)
	DeleteOption(ctx context.Context, id string) error
	// Ballots
	CreateBallot(ctx context.Context, ballot *model.VoteBallot) error
	GetBallotByVoter(ctx context.Context, voteID, userID string) (*model.VoteBallot, error)
	GetBallotsByVote(ctx context.Context, voteID string) ([]*model.VoteBallot, error)
	DeleteBallot(ctx context.Context, id string) error
	HasVoted(ctx context.Context, voteID, userID string) (bool, error)
	CountBallots(ctx context.Context, voteID string) (int, error)
}

// VoteUserRepository defines interface for getting user info for snapshots
type VoteUserRepository interface {
	GetByID(ctx context.Context, id string) (*model.User, error)
}

// VoteService handles vote business logic
type VoteService struct {
	repo      VoteRepository
	userRepo  VoteUserRepository
	guildRepo GuildRepository // Uses GuildRepository which has IsMember
}

// VoteServiceConfig holds configuration for the vote service
type VoteServiceConfig struct {
	VoteRepo   VoteRepository
	UserRepo   VoteUserRepository
	GuildRepo  GuildRepository
	MemberRepo interface{} // Deprecated, kept for backwards compatibility
}

// NewVoteService creates a new vote service
func NewVoteService(cfg VoteServiceConfig) *VoteService {
	return &VoteService{
		repo:      cfg.VoteRepo,
		userRepo:  cfg.UserRepo,
		guildRepo: cfg.GuildRepo,
	}
}

// Create creates a new vote
func (s *VoteService) Create(ctx context.Context, userID string, req *model.CreateVoteRequest) (*model.Vote, error) {
	// Validate request
	if errors := req.Validate(); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	// Parse times
	opensAt, err := time.Parse(time.RFC3339, req.OpensAt)
	if err != nil {
		return nil, model.NewBadRequestError("invalid opens_at format")
	}
	closesAt, err := time.Parse(time.RFC3339, req.ClosesAt)
	if err != nil {
		return nil, model.NewBadRequestError("invalid closes_at format")
	}

	// Validate dates
	if closesAt.Before(opensAt) {
		return nil, model.NewBadRequestError("closes_at must be after opens_at")
	}

	// Check permissions for guild votes (TODO: restrict to admin role when available)
	if req.ScopeType == string(model.VoteScopeGuild) && req.ScopeID != nil {
		if s.guildRepo != nil {
			isMember, err := s.guildRepo.IsMember(ctx, userID, *req.ScopeID)
			if err != nil {
				return nil, fmt.Errorf("failed to check membership: %w", err)
			}
			if !isMember {
				return nil, model.NewForbiddenError("must be guild member to create votes")
			}
		}
	}

	// For global votes, would need sysadmin check (not implemented here)

	resultsVisibility := model.ResultsVisibilityAfterClose
	if req.ResultsVisibility != nil {
		resultsVisibility = model.ResultsVisibility(*req.ResultsVisibility)
	}

	vote := &model.Vote{
		ScopeType:            model.VoteScopeType(req.ScopeType),
		ScopeID:              req.ScopeID,
		CreatedBy:            userID,
		Title:                req.Title,
		Description:          req.Description,
		VoteType:             model.VoteType(req.VoteType),
		OpensAt:              opensAt,
		ClosesAt:             closesAt,
		Status:               model.VoteStatusDraft,
		ResultsVisibility:    resultsVisibility,
		MaxOptionsSelectable: req.MaxOptionsSelectable,
		AllowAbstain:         req.AllowAbstain,
	}

	if err := s.repo.Create(ctx, vote); err != nil {
		return nil, fmt.Errorf("failed to create vote: %w", err)
	}

	return vote, nil
}

// GetByID retrieves a vote by ID with options
func (s *VoteService) GetByID(ctx context.Context, id string, userID string) (*model.VoteWithDetails, error) {
	vote, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get vote: %w", err)
	}
	if vote == nil {
		return nil, model.NewNotFoundError("vote not found")
	}

	optionPtrs, err := s.repo.GetOptionsByVote(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}

	// Convert []*VoteOption to []VoteOption
	options := make([]model.VoteOption, len(optionPtrs))
	for i, opt := range optionPtrs {
		options[i] = *opt
	}

	hasVoted, _ := s.repo.HasVoted(ctx, id, userID)
	canVote := s.canVote(ctx, vote, userID)

	details := &model.VoteWithDetails{
		Vote:     *vote,
		Options:  options,
		HasVoted: hasVoted,
		CanVote:  canVote,
	}

	// Get user's ballot if they voted
	if hasVoted {
		ballot, _ := s.repo.GetBallotByVoter(ctx, id, userID)
		details.MyBallot = ballot
	}

	return details, nil
}

// GetGuildVotes retrieves votes for a guild
func (s *VoteService) GetGuildVotes(ctx context.Context, guildID string, status *string, limit, offset int) ([]*model.Vote, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var statusFilter *model.VoteStatus
	if status != nil {
		st := model.VoteStatus(*status)
		statusFilter = &st
	}

	return s.repo.GetByGuild(ctx, guildID, statusFilter, limit, offset)
}

// GetGlobalVotes retrieves global votes
func (s *VoteService) GetGlobalVotes(ctx context.Context, status *string, limit, offset int) ([]*model.Vote, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var statusFilter *model.VoteStatus
	if status != nil {
		st := model.VoteStatus(*status)
		statusFilter = &st
	}

	return s.repo.GetGlobalVotes(ctx, statusFilter, limit, offset)
}

// Update updates a vote (only when draft)
func (s *VoteService) Update(ctx context.Context, id string, userID string, req *model.UpdateVoteRequest) (*model.Vote, error) {
	if errors := req.Validate(); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	vote, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get vote: %w", err)
	}
	if vote == nil {
		return nil, model.NewNotFoundError("vote not found")
	}

	// Only draft votes can be updated
	if vote.Status != model.VoteStatusDraft {
		return nil, model.NewBadRequestError("can only update draft votes")
	}

	// Check permission
	if vote.CreatedBy != userID {
		return nil, model.NewForbiddenError("not your vote")
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.OpensAt != nil {
		t, err := time.Parse(time.RFC3339, *req.OpensAt)
		if err != nil {
			return nil, model.NewBadRequestError("invalid opens_at format")
		}
		updates["opens_at"] = t
	}
	if req.ClosesAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ClosesAt)
		if err != nil {
			return nil, model.NewBadRequestError("invalid closes_at format")
		}
		updates["closes_at"] = t
	}
	if req.ResultsVisibility != nil {
		updates["results_visibility"] = *req.ResultsVisibility
	}
	if req.MaxOptionsSelectable != nil {
		updates["max_options_selectable"] = *req.MaxOptionsSelectable
	}
	if req.AllowAbstain != nil {
		updates["allow_abstain"] = *req.AllowAbstain
	}

	return s.repo.Update(ctx, id, updates)
}

// Open manually opens a vote
func (s *VoteService) Open(ctx context.Context, id string, userID string) error {
	vote, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get vote: %w", err)
	}
	if vote == nil {
		return model.NewNotFoundError("vote not found")
	}

	if vote.Status != model.VoteStatusDraft {
		return model.NewBadRequestError("can only open draft votes")
	}

	if vote.CreatedBy != userID {
		return model.NewForbiddenError("not your vote")
	}

	return s.repo.UpdateStatus(ctx, id, model.VoteStatusOpen)
}

// Close manually closes a vote
func (s *VoteService) Close(ctx context.Context, id string, userID string) error {
	vote, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get vote: %w", err)
	}
	if vote == nil {
		return model.NewNotFoundError("vote not found")
	}

	if vote.Status != model.VoteStatusOpen {
		return model.NewBadRequestError("can only close open votes")
	}

	if vote.CreatedBy != userID {
		return model.NewForbiddenError("not your vote")
	}

	return s.repo.UpdateStatus(ctx, id, model.VoteStatusClosed)
}

// Cancel cancels a vote
func (s *VoteService) Cancel(ctx context.Context, id string, userID string) error {
	vote, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get vote: %w", err)
	}
	if vote == nil {
		return model.NewNotFoundError("vote not found")
	}

	if vote.Status == model.VoteStatusClosed || vote.Status == model.VoteStatusCancelled {
		return model.NewBadRequestError("vote already ended")
	}

	if vote.CreatedBy != userID {
		return model.NewForbiddenError("not your vote")
	}

	return s.repo.UpdateStatus(ctx, id, model.VoteStatusCancelled)
}

// Option operations

// AddOption adds an option to a vote
func (s *VoteService) AddOption(ctx context.Context, voteID string, userID string, req *model.CreateVoteOptionRequest) (*model.VoteOption, error) {
	if errors := req.Validate(); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	vote, err := s.repo.GetByID(ctx, voteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vote: %w", err)
	}
	if vote == nil {
		return nil, model.NewNotFoundError("vote not found")
	}

	if vote.Status != model.VoteStatusDraft {
		return nil, model.NewBadRequestError("can only add options to draft votes")
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	option := &model.VoteOption{
		VoteID:            voteID,
		OptionText:        req.OptionText,
		OptionDescription: req.OptionDescription,
		SortOrder:         sortOrder,
		CreatedBy:         userID,
	}

	if err := s.repo.CreateOption(ctx, option); err != nil {
		return nil, fmt.Errorf("failed to create option: %w", err)
	}

	return option, nil
}

// UpdateOption updates a vote option
func (s *VoteService) UpdateOption(ctx context.Context, optionID string, userID string, req *model.UpdateVoteOptionRequest) (*model.VoteOption, error) {
	if errors := req.Validate(); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	option, err := s.repo.GetOptionByID(ctx, optionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get option: %w", err)
	}
	if option == nil {
		return nil, model.NewNotFoundError("option not found")
	}

	// Check vote status
	vote, _ := s.repo.GetByID(ctx, option.VoteID)
	if vote != nil && vote.Status != model.VoteStatusDraft {
		return nil, model.NewBadRequestError("can only update options for draft votes")
	}

	return s.repo.UpdateOption(ctx, optionID, req)
}

// DeleteOption deletes a vote option
func (s *VoteService) DeleteOption(ctx context.Context, optionID string, userID string) error {
	option, err := s.repo.GetOptionByID(ctx, optionID)
	if err != nil {
		return fmt.Errorf("failed to get option: %w", err)
	}
	if option == nil {
		return model.NewNotFoundError("option not found")
	}

	// Check vote status
	vote, _ := s.repo.GetByID(ctx, option.VoteID)
	if vote != nil && vote.Status != model.VoteStatusDraft {
		return model.NewBadRequestError("can only delete options for draft votes")
	}

	return s.repo.DeleteOption(ctx, optionID)
}

// Ballot operations

// CastBallot casts a vote
func (s *VoteService) CastBallot(ctx context.Context, voteID string, userID string, req *model.CastBallotRequest) (*model.VoteBallot, error) {
	vote, err := s.repo.GetByID(ctx, voteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vote: %w", err)
	}
	if vote == nil {
		return nil, model.NewNotFoundError("vote not found")
	}

	// Validate ballot for vote type
	if errors := req.ValidateForVoteType(vote.VoteType, vote.MaxOptionsSelectable, vote.AllowAbstain); len(errors) > 0 {
		return nil, model.NewValidationError(errors)
	}

	// Check if voting is open
	if vote.Status != model.VoteStatusOpen {
		return nil, model.NewBadRequestError("voting is not open")
	}

	// Check if user can vote
	if !s.canVote(ctx, vote, userID) {
		return nil, model.NewForbiddenError("you cannot vote in this poll")
	}

	// Check if already voted - delete existing ballot to allow revoting
	existingBallot, _ := s.repo.GetBallotByVoter(ctx, voteID, userID)
	if existingBallot != nil {
		if err := s.repo.DeleteBallot(ctx, existingBallot.ID); err != nil {
			return nil, fmt.Errorf("failed to remove existing ballot: %w", err)
		}
	}

	// Get user info for snapshot
	var snapshot model.VoterSnapshot
	if s.userRepo != nil {
		user, _ := s.userRepo.GetByID(ctx, userID)
		if user != nil {
			if user.Username != nil {
				snapshot.Username = *user.Username
			}
			// Use email as display name fallback
			snapshot.DisplayName = user.Email
		}
	}

	ballot := &model.VoteBallot{
		VoteID:        voteID,
		VoterUserID:   userID,
		VoterSnapshot: snapshot,
		BallotData:    req.ToBallotData(vote.VoteType),
		IsAbstain:     req.IsAbstain,
	}

	if err := s.repo.CreateBallot(ctx, ballot); err != nil {
		return nil, fmt.Errorf("failed to cast ballot: %w", err)
	}

	return ballot, nil
}

// GetMyBallot retrieves the user's ballot for a vote
func (s *VoteService) GetMyBallot(ctx context.Context, voteID string, userID string) (*model.VoteBallot, error) {
	return s.repo.GetBallotByVoter(ctx, voteID, userID)
}

// GetBallots retrieves all ballots for a vote (transparent ledger)
func (s *VoteService) GetBallots(ctx context.Context, voteID string, userID string) ([]*model.VoteBallot, error) {
	vote, err := s.repo.GetByID(ctx, voteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vote: %w", err)
	}
	if vote == nil {
		return nil, model.NewNotFoundError("vote not found")
	}

	// Check visibility
	if !s.canViewResults(vote, userID) {
		return nil, model.NewForbiddenError("results not yet visible")
	}

	return s.repo.GetBallotsByVote(ctx, voteID)
}

// GetResults computes and returns vote results
func (s *VoteService) GetResults(ctx context.Context, voteID string, userID string) (*model.VoteResult, error) {
	vote, err := s.repo.GetByID(ctx, voteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vote: %w", err)
	}
	if vote == nil {
		return nil, model.NewNotFoundError("vote not found")
	}

	// Check visibility
	if !s.canViewResults(vote, userID) {
		return nil, model.NewForbiddenError("results not yet visible")
	}

	options, err := s.repo.GetOptionsByVote(ctx, voteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}

	ballots, err := s.repo.GetBallotsByVote(ctx, voteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ballots: %w", err)
	}

	return s.computeResults(vote, options, ballots), nil
}

// ProcessScheduledTransitions processes votes that should open/close based on time
func (s *VoteService) ProcessScheduledTransitions(ctx context.Context) error {
	// Open votes that should be open
	toOpen, err := s.repo.GetVotesToOpen(ctx)
	if err != nil {
		return fmt.Errorf("failed to get votes to open: %w", err)
	}
	for _, vote := range toOpen {
		_ = s.repo.UpdateStatus(ctx, vote.ID, model.VoteStatusOpen)
	}

	// Close votes that should be closed
	toClose, err := s.repo.GetVotesToClose(ctx)
	if err != nil {
		return fmt.Errorf("failed to get votes to close: %w", err)
	}
	for _, vote := range toClose {
		_ = s.repo.UpdateStatus(ctx, vote.ID, model.VoteStatusClosed)
	}

	return nil
}

// Helper methods

func (s *VoteService) canVote(ctx context.Context, vote *model.Vote, userID string) bool {
	if vote.Status != model.VoteStatusOpen {
		return false
	}

	// Global votes - any authenticated user can vote
	if vote.ScopeType == model.VoteScopeGlobal {
		return true
	}

	// Guild votes - must be guild member
	if vote.ScopeType == model.VoteScopeGuild && vote.ScopeID != nil && s.guildRepo != nil {
		isMember, _ := s.guildRepo.IsMember(ctx, userID, *vote.ScopeID)
		return isMember
	}

	return false
}

func (s *VoteService) canViewResults(vote *model.Vote, userID string) bool {
	switch vote.ResultsVisibility {
	case model.ResultsVisibilityLive:
		return true
	case model.ResultsVisibilityAfterClose:
		return vote.Status == model.VoteStatusClosed
	case model.ResultsVisibilityAdminOnly:
		return vote.CreatedBy == userID
	}
	return false
}

func (s *VoteService) computeResults(vote *model.Vote, options []*model.VoteOption, ballots []*model.VoteBallot) *model.VoteResult {
	result := &model.VoteResult{
		VoteID:   vote.ID,
		VoteType: vote.VoteType,
	}

	// Count abstains
	nonAbstainBallots := make([]*model.VoteBallot, 0)
	for _, b := range ballots {
		if b.IsAbstain {
			result.TotalAbstains++
		} else {
			nonAbstainBallots = append(nonAbstainBallots, b)
		}
	}
	result.TotalBallots = len(ballots)

	// Create option map
	optionMap := make(map[string]*model.VoteOption)
	for _, opt := range options {
		optionMap[opt.ID] = opt
	}

	switch vote.VoteType {
	case model.VoteTypeFPTP:
		result.OptionResults = s.computeFPTP(options, nonAbstainBallots)
	case model.VoteTypeRankedChoice:
		result.OptionResults, result.RoundDetails = s.computeRankedChoice(options, nonAbstainBallots)
	case model.VoteTypeApproval, model.VoteTypeMultiSelect:
		result.OptionResults = s.computeApproval(options, nonAbstainBallots, vote.VoteType)
	}

	// Set winner
	for i := range result.OptionResults {
		if result.OptionResults[i].IsWinner {
			result.Winner = &result.OptionResults[i].OptionID
			break
		}
	}

	return result
}

func (s *VoteService) computeFPTP(options []*model.VoteOption, ballots []*model.VoteBallot) []model.OptionResult {
	counts := make(map[string]int)
	for _, opt := range options {
		counts[opt.ID] = 0
	}

	for _, ballot := range ballots {
		if optID, ok := ballot.BallotData["option_id"].(string); ok {
			counts[optID]++
		}
	}

	total := len(ballots)
	results := make([]model.OptionResult, 0, len(options))
	for _, opt := range options {
		count := counts[opt.ID]
		pct := 0.0
		if total > 0 {
			pct = float64(count) / float64(total) * 100
		}
		results = append(results, model.OptionResult{
			OptionID:   opt.ID,
			OptionText: opt.OptionText,
			VoteCount:  count,
			Percentage: pct,
		})
	}

	// Sort by count descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].VoteCount > results[j].VoteCount
	})

	// Set ranks and winner
	for i := range results {
		results[i].Rank = i + 1
		if i == 0 && results[i].VoteCount > 0 {
			results[i].IsWinner = true
		}
	}

	return results
}

func (s *VoteService) computeRankedChoice(options []*model.VoteOption, ballots []*model.VoteBallot) ([]model.OptionResult, []model.RoundDetail) {
	// Initial counts - track active candidates
	active := make(map[string]bool)
	for _, opt := range options {
		active[opt.ID] = true
	}

	rounds := make([]model.RoundDetail, 0)
	majority := len(ballots)/2 + 1

	// Keep running until we have a winner
	for round := 1; len(active) > 1; round++ {
		counts := make(map[string]int)
		for id := range active {
			counts[id] = 0
		}

		// Count first choices among active candidates
		for _, ballot := range ballots {
			rankings, ok := ballot.BallotData["rankings"].([]interface{})
			if !ok {
				continue
			}

			// Find first active choice
			for _, r := range rankings {
				optID, ok := r.(string)
				if !ok {
					continue
				}
				if active[optID] {
					counts[optID]++
					break
				}
			}
		}

		roundDetail := model.RoundDetail{
			Round:        round,
			OptionCounts: counts,
		}

		// Check for majority
		for id, count := range counts {
			if count >= majority {
				// Winner found
				rounds = append(rounds, roundDetail)
				return s.buildRankedChoiceResults(options, counts, active, id), rounds
			}
		}

		// Find lowest count
		minCount := len(ballots) + 1
		var eliminatedID string
		for id, count := range counts {
			if count < minCount {
				minCount = count
				eliminatedID = id
			}
		}

		// Eliminate lowest
		if eliminatedID != "" {
			roundDetail.EliminatedID = &eliminatedID
			roundDetail.EliminatedCount = minCount
			delete(active, eliminatedID)
		}

		rounds = append(rounds, roundDetail)
	}

	// Only one candidate left
	var winnerID string
	for id := range active {
		winnerID = id
		break
	}

	// Final count
	counts := make(map[string]int)
	for _, opt := range options {
		counts[opt.ID] = 0
	}
	for _, ballot := range ballots {
		rankings, ok := ballot.BallotData["rankings"].([]interface{})
		if !ok {
			continue
		}
		for _, r := range rankings {
			optID, ok := r.(string)
			if !ok {
				continue
			}
			if active[optID] {
				counts[optID]++
				break
			}
		}
	}

	return s.buildRankedChoiceResults(options, counts, active, winnerID), rounds
}

func (s *VoteService) buildRankedChoiceResults(options []*model.VoteOption, counts map[string]int, active map[string]bool, winnerID string) []model.OptionResult {
	total := 0
	for _, c := range counts {
		total += c
	}

	results := make([]model.OptionResult, 0, len(options))
	for _, opt := range options {
		count := counts[opt.ID]
		pct := 0.0
		if total > 0 {
			pct = float64(count) / float64(total) * 100
		}
		results = append(results, model.OptionResult{
			OptionID:     opt.ID,
			OptionText:   opt.OptionText,
			VoteCount:    count,
			Percentage:   pct,
			IsWinner:     opt.ID == winnerID,
			IsEliminated: !active[opt.ID],
		})
	}

	// Sort by count descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].VoteCount > results[j].VoteCount
	})

	for i := range results {
		results[i].Rank = i + 1
	}

	return results
}

func (s *VoteService) computeApproval(options []*model.VoteOption, ballots []*model.VoteBallot, voteType model.VoteType) []model.OptionResult {
	counts := make(map[string]int)
	for _, opt := range options {
		counts[opt.ID] = 0
	}

	key := "approved_options"
	if voteType == model.VoteTypeMultiSelect {
		key = "selected_options"
	}

	for _, ballot := range ballots {
		selected, ok := ballot.BallotData[key].([]interface{})
		if !ok {
			continue
		}
		for _, sel := range selected {
			if optID, ok := sel.(string); ok {
				counts[optID]++
			}
		}
	}

	total := len(ballots)
	results := make([]model.OptionResult, 0, len(options))
	for _, opt := range options {
		count := counts[opt.ID]
		pct := 0.0
		if total > 0 {
			pct = float64(count) / float64(total) * 100
		}
		results = append(results, model.OptionResult{
			OptionID:   opt.ID,
			OptionText: opt.OptionText,
			VoteCount:  count,
			Percentage: pct,
		})
	}

	// Sort by count descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].VoteCount > results[j].VoteCount
	})

	for i := range results {
		results[i].Rank = i + 1
		if i == 0 && results[i].VoteCount > 0 {
			results[i].IsWinner = true
		}
	}

	return results
}
