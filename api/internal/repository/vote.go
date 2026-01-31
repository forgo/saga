package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// VoteRepository handles vote data access
type VoteRepository struct {
	db database.Database
}

// NewVoteRepository creates a new vote repository
func NewVoteRepository(db database.Database) *VoteRepository {
	return &VoteRepository{db: db}
}

// Vote CRUD

// Create creates a new vote
func (r *VoteRepository) Create(ctx context.Context, vote *model.Vote) error {
	// Build query dynamically to avoid NULL values for optional fields
	vars := map[string]interface{}{
		"scope_type":         vote.ScopeType,
		"created_by":         vote.CreatedBy,
		"title":              vote.Title,
		"vote_type":          vote.VoteType,
		"opens_at":           vote.OpensAt,
		"closes_at":          vote.ClosesAt,
		"status":             vote.Status,
		"results_visibility": vote.ResultsVisibility,
		"allow_abstain":      vote.AllowAbstain,
	}

	// Build optional fields
	optionalFields := ""
	if vote.ScopeID != nil && *vote.ScopeID != "" {
		optionalFields += ",\n\t\t\tscope_id: type::record($scope_id)"
		vars["scope_id"] = *vote.ScopeID
	}
	if vote.Description != nil && *vote.Description != "" {
		optionalFields += ",\n\t\t\tdescription: $description"
		vars["description"] = *vote.Description
	}
	if vote.MaxOptionsSelectable != nil {
		optionalFields += ",\n\t\t\tmax_options_selectable: $max_options"
		vars["max_options"] = *vote.MaxOptionsSelectable
	}

	query := `
		CREATE vote CONTENT {
			scope_type: $scope_type,
			created_by: type::record($created_by),
			title: $title,
			vote_type: $vote_type,
			opens_at: $opens_at,
			closes_at: $closes_at,
			status: $status,
			results_visibility: $results_visibility,
			allow_abstain: $allow_abstain,
			created_on: time::now(),
			updated_on: time::now()` + optionalFields + `
		}
	`

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return fmt.Errorf("failed to create vote: %w", err)
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return fmt.Errorf("failed to extract created vote: %w", err)
	}

	vote.ID = created.ID
	vote.CreatedOn = created.CreatedOn
	vote.UpdatedOn = created.UpdatedOn
	return nil
}

// GetByID retrieves a vote by ID
func (r *VoteRepository) GetByID(ctx context.Context, id string) (*model.Vote, error) {
	query := `SELECT * FROM type::record($id)`
	result, err := r.db.QueryOne(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get vote: %w", err)
	}

	return r.parseVote(result)
}

// GetByGuild retrieves votes for a guild
func (r *VoteRepository) GetByGuild(ctx context.Context, guildID string, status *model.VoteStatus, limit, offset int) ([]*model.Vote, error) {
	query := `
		SELECT * FROM vote
		WHERE scope_type = "guild"
		AND scope_id = type::record($guild_id)
	`
	vars := map[string]interface{}{
		"guild_id": guildID,
		"limit":    limit,
		"offset":   offset,
	}

	if status != nil {
		query += ` AND status = $status`
		vars["status"] = *status
	}

	query += ` ORDER BY created_on DESC LIMIT $limit START $offset`

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild votes: %w", err)
	}

	return r.parseVotes(result)
}

// GetGlobalVotes retrieves global votes
func (r *VoteRepository) GetGlobalVotes(ctx context.Context, status *model.VoteStatus, limit, offset int) ([]*model.Vote, error) {
	query := `
		SELECT * FROM vote
		WHERE scope_type = "global"
	`
	vars := map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}

	if status != nil {
		query += ` AND status = $status`
		vars["status"] = *status
	}

	query += ` ORDER BY created_on DESC LIMIT $limit START $offset`

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get global votes: %w", err)
	}

	return r.parseVotes(result)
}

// GetVotesToOpen retrieves votes that should be opened (opens_at <= now, status = draft)
func (r *VoteRepository) GetVotesToOpen(ctx context.Context) ([]*model.Vote, error) {
	query := `
		SELECT * FROM vote
		WHERE status = "draft"
		AND opens_at <= time::now()
	`

	result, err := r.db.Query(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get votes to open: %w", err)
	}

	return r.parseVotes(result)
}

// GetVotesToClose retrieves votes that should be closed (closes_at <= now, status = open)
func (r *VoteRepository) GetVotesToClose(ctx context.Context) ([]*model.Vote, error) {
	query := `
		SELECT * FROM vote
		WHERE status = "open"
		AND closes_at <= time::now()
	`

	result, err := r.db.Query(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get votes to close: %w", err)
	}

	return r.parseVotes(result)
}

// Update updates a vote (only when draft)
func (r *VoteRepository) Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Vote, error) {
	query := `UPDATE type::record($id) SET updated_on = time::now()`
	vars := map[string]interface{}{"id": id}

	for key, value := range updates {
		query += fmt.Sprintf(`, %s = $%s`, key, key)
		vars[key] = value
	}

	query += ` RETURN AFTER`

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to update vote: %w", err)
	}

	return r.parseVote(result)
}

// UpdateStatus updates vote status
func (r *VoteRepository) UpdateStatus(ctx context.Context, id string, status model.VoteStatus) error {
	query := `UPDATE type::record($id) SET status = $status, updated_on = time::now()`
	vars := map[string]interface{}{
		"id":     id,
		"status": status,
	}

	if err := r.db.Execute(ctx, query, vars); err != nil {
		return fmt.Errorf("failed to update vote status: %w", err)
	}
	return nil
}

// Delete deletes a vote
func (r *VoteRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE type::record($id)`
	if err := r.db.Execute(ctx, query, map[string]interface{}{"id": id}); err != nil {
		return fmt.Errorf("failed to delete vote: %w", err)
	}
	return nil
}

// Vote Options

// CreateOption creates a vote option
func (r *VoteRepository) CreateOption(ctx context.Context, option *model.VoteOption) error {
	// Build query dynamically to avoid NULL values for optional fields
	vars := map[string]interface{}{
		"vote_id":     option.VoteID,
		"option_text": option.OptionText,
		"sort_order":  option.SortOrder,
		"created_by":  option.CreatedBy,
	}

	optionalFields := ""
	if option.OptionDescription != nil && *option.OptionDescription != "" {
		optionalFields = ",\n\t\t\toption_description: $option_description"
		vars["option_description"] = *option.OptionDescription
	}

	query := `
		CREATE vote_option CONTENT {
			vote_id: type::record($vote_id),
			option_text: $option_text,
			sort_order: $sort_order,
			created_by: type::record($created_by),
			created_on: time::now()` + optionalFields + `
		}
	`

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return fmt.Errorf("failed to create option: %w", err)
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return fmt.Errorf("failed to extract created option: %w", err)
	}

	option.ID = created.ID
	option.CreatedOn = created.CreatedOn
	return nil
}

// GetOptionByID retrieves an option by ID
func (r *VoteRepository) GetOptionByID(ctx context.Context, id string) (*model.VoteOption, error) {
	query := `SELECT * FROM type::record($id)`
	result, err := r.db.QueryOne(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get option: %w", err)
	}

	return r.parseOption(result)
}

// GetOptionsByVote retrieves all options for a vote
func (r *VoteRepository) GetOptionsByVote(ctx context.Context, voteID string) ([]*model.VoteOption, error) {
	query := `
		SELECT * FROM vote_option
		WHERE vote_id = type::record($vote_id)
		ORDER BY sort_order ASC
	`
	vars := map[string]interface{}{"vote_id": voteID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}

	return r.parseOptions(result)
}

// UpdateOption updates an option
func (r *VoteRepository) UpdateOption(ctx context.Context, id string, updates *model.UpdateVoteOptionRequest) (*model.VoteOption, error) {
	query := `UPDATE type::record($id) SET `
	vars := map[string]interface{}{"id": id}
	first := true

	if updates.OptionText != nil {
		query += `option_text = $option_text`
		vars["option_text"] = *updates.OptionText
		first = false
	}
	if updates.OptionDescription != nil {
		if !first {
			query += `, `
		}
		query += `option_description = $option_description`
		vars["option_description"] = *updates.OptionDescription
		first = false
	}
	if updates.SortOrder != nil {
		if !first {
			query += `, `
		}
		query += `sort_order = $sort_order`
		vars["sort_order"] = *updates.SortOrder
	}

	query += ` RETURN AFTER`

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to update option: %w", err)
	}

	return r.parseOption(result)
}

// DeleteOption deletes an option
func (r *VoteRepository) DeleteOption(ctx context.Context, id string) error {
	query := `DELETE type::record($id)`
	if err := r.db.Execute(ctx, query, map[string]interface{}{"id": id}); err != nil {
		return fmt.Errorf("failed to delete option: %w", err)
	}
	return nil
}

// Ballots

// CreateBallot creates a ballot
func (r *VoteRepository) CreateBallot(ctx context.Context, ballot *model.VoteBallot) error {
	// Convert voter snapshot to map for SurrealDB object field
	voterSnapshot := map[string]interface{}{
		"username":     ballot.VoterSnapshot.Username,
		"display_name": ballot.VoterSnapshot.DisplayName,
	}

	query := `
		CREATE vote_ballot CONTENT {
			vote_id: type::record($vote_id),
			voter_user_id: type::record($voter_user_id),
			voter_snapshot: $voter_snapshot,
			ballot_data: $ballot_data,
			is_abstain: $is_abstain,
			created_on: time::now()
		}
	`
	vars := map[string]interface{}{
		"vote_id":        ballot.VoteID,
		"voter_user_id":  ballot.VoterUserID,
		"voter_snapshot": voterSnapshot,
		"ballot_data":    map[string]interface{}(ballot.BallotData),
		"is_abstain":     ballot.IsAbstain,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		if isUniqueConstraintError(err) {
			return fmt.Errorf("ballot already cast for this vote")
		}
		return fmt.Errorf("failed to create ballot: %w", err)
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return fmt.Errorf("failed to extract created ballot: %w", err)
	}

	ballot.ID = created.ID
	ballot.CreatedOn = created.CreatedOn
	return nil
}

// GetBallotByVoter retrieves a user's ballot for a vote
func (r *VoteRepository) GetBallotByVoter(ctx context.Context, voteID, userID string) (*model.VoteBallot, error) {
	query := `
		SELECT * FROM vote_ballot
		WHERE vote_id = type::record($vote_id)
		AND voter_user_id = type::record($user_id)
	`
	vars := map[string]interface{}{
		"vote_id": voteID,
		"user_id": userID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get ballot: %w", err)
	}

	return r.parseBallot(result)
}

// GetBallotsByVote retrieves all ballots for a vote
func (r *VoteRepository) GetBallotsByVote(ctx context.Context, voteID string) ([]*model.VoteBallot, error) {
	query := `
		SELECT * FROM vote_ballot
		WHERE vote_id = type::record($vote_id)
		ORDER BY created_on ASC
	`
	vars := map[string]interface{}{"vote_id": voteID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get ballots: %w", err)
	}

	return r.parseBallots(result)
}

// DeleteBallot deletes a ballot (allows revoting)
func (r *VoteRepository) DeleteBallot(ctx context.Context, id string) error {
	query := `DELETE type::record($id)`
	if err := r.db.Execute(ctx, query, map[string]interface{}{"id": id}); err != nil {
		return fmt.Errorf("failed to delete ballot: %w", err)
	}
	return nil
}

// HasVoted checks if a user has voted
func (r *VoteRepository) HasVoted(ctx context.Context, voteID, userID string) (bool, error) {
	query := `
		SELECT count() as count FROM vote_ballot
		WHERE vote_id = type::record($vote_id)
		AND voter_user_id = type::record($user_id)
		GROUP ALL
	`
	vars := map[string]interface{}{
		"vote_id": voteID,
		"user_id": userID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	if data, ok := result.(map[string]interface{}); ok {
		return getInt(data, "count") > 0, nil
	}
	return false, nil
}

// CountBallots counts ballots for a vote
func (r *VoteRepository) CountBallots(ctx context.Context, voteID string) (int, error) {
	query := `
		SELECT count() as count FROM vote_ballot
		WHERE vote_id = type::record($vote_id)
		GROUP ALL
	`
	vars := map[string]interface{}{"vote_id": voteID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	if data, ok := result.(map[string]interface{}); ok {
		return getInt(data, "count"), nil
	}
	return 0, nil
}

// Parsing helpers

func (r *VoteRepository) parseVote(result interface{}) (*model.Vote, error) {
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	vote := &model.Vote{
		ID:                convertSurrealID(data["id"]),
		ScopeType:         model.VoteScopeType(getString(data, "scope_type")),
		CreatedBy:         convertSurrealID(data["created_by"]),
		Title:             getString(data, "title"),
		VoteType:          model.VoteType(getString(data, "vote_type")),
		Status:            model.VoteStatus(getString(data, "status")),
		ResultsVisibility: model.ResultsVisibility(getString(data, "results_visibility")),
		AllowAbstain:      getBool(data, "allow_abstain"),
	}

	if scopeID := convertSurrealID(data["scope_id"]); scopeID != "" {
		vote.ScopeID = &scopeID
	}
	if desc := getString(data, "description"); desc != "" {
		vote.Description = &desc
	}
	if maxOpts := getInt(data, "max_options_selectable"); maxOpts > 0 {
		vote.MaxOptionsSelectable = &maxOpts
	}
	if t := getTime(data, "opens_at"); t != nil {
		vote.OpensAt = *t
	}
	if t := getTime(data, "closes_at"); t != nil {
		vote.ClosesAt = *t
	}
	if t := getTime(data, "created_on"); t != nil {
		vote.CreatedOn = *t
	}
	if t := getTime(data, "updated_on"); t != nil {
		vote.UpdatedOn = *t
	}

	return vote, nil
}

func (r *VoteRepository) parseVotes(result []interface{}) ([]*model.Vote, error) {
	votes := make([]*model.Vote, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					vote, err := r.parseVote(item)
					if err != nil {
						continue
					}
					votes = append(votes, vote)
				}
			}
		}
	}

	return votes, nil
}

func (r *VoteRepository) parseOption(result interface{}) (*model.VoteOption, error) {
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	option := &model.VoteOption{
		ID:         convertSurrealID(data["id"]),
		VoteID:     convertSurrealID(data["vote_id"]),
		OptionText: getString(data, "option_text"),
		SortOrder:  getInt(data, "sort_order"),
		CreatedBy:  convertSurrealID(data["created_by"]),
	}

	if desc := getString(data, "option_description"); desc != "" {
		option.OptionDescription = &desc
	}
	if t := getTime(data, "created_on"); t != nil {
		option.CreatedOn = *t
	}

	return option, nil
}

func (r *VoteRepository) parseOptions(result []interface{}) ([]*model.VoteOption, error) {
	options := make([]*model.VoteOption, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					option, err := r.parseOption(item)
					if err != nil {
						continue
					}
					options = append(options, option)
				}
			}
		}
	}

	return options, nil
}

func (r *VoteRepository) parseBallot(result interface{}) (*model.VoteBallot, error) {
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	ballot := &model.VoteBallot{
		ID:          convertSurrealID(data["id"]),
		VoteID:      convertSurrealID(data["vote_id"]),
		VoterUserID: convertSurrealID(data["voter_user_id"]),
		IsAbstain:   getBool(data, "is_abstain"),
	}

	// Parse voter snapshot
	if snapshotData, ok := data["voter_snapshot"].(map[string]interface{}); ok {
		ballot.VoterSnapshot = model.VoterSnapshot{
			Username:    getString(snapshotData, "username"),
			DisplayName: getString(snapshotData, "display_name"),
		}
	}

	// Parse ballot data
	if ballotData, ok := data["ballot_data"].(map[string]interface{}); ok {
		ballot.BallotData = ballotData
	}

	if t := getTime(data, "created_on"); t != nil {
		ballot.CreatedOn = *t
	}

	return ballot, nil
}

func (r *VoteRepository) parseBallots(result []interface{}) ([]*model.VoteBallot, error) {
	ballots := make([]*model.VoteBallot, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					ballot, err := r.parseBallot(item)
					if err != nil {
						continue
					}
					ballots = append(ballots, ballot)
				}
			}
		}
	}

	return ballots, nil
}
