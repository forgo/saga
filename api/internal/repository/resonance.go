package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// Ensure database package is used for AtomicBatch
var _ = database.NewAtomicBatch

// ResonanceRepository handles resonance scoring data access
type ResonanceRepository struct {
	db database.Database
}

// NewResonanceRepository creates a new resonance repository
func NewResonanceRepository(db database.Database) *ResonanceRepository {
	return &ResonanceRepository{db: db}
}

// ErrAlreadyAwarded is returned when trying to award points that were already given
var ErrAlreadyAwarded = errors.New("points already awarded for this source")

// AwardPoints adds points to a user's resonance (idempotent)
// Returns ErrAlreadyAwarded if points have already been given for this (user, stat, source) combo
func (r *ResonanceRepository) AwardPoints(ctx context.Context, entry *model.ResonanceLedgerEntry) error {
	// First check idempotency to avoid hitting unique constraint
	awarded, err := r.HasAwardedPoints(ctx, entry.UserID, string(entry.Stat), entry.SourceObjectID)
	if err != nil {
		return err
	}
	if awarded {
		return ErrAlreadyAwarded
	}

	query := `
		CREATE resonance_ledger CONTENT {
			user: type::record($user_id),
			stat: $stat,
			points: $points,
			source_object_id: $source_object_id,
			reason_code: $reason_code,
			created_on: time::now()
		}
	`

	vars := map[string]interface{}{
		"user_id":          entry.UserID,
		"stat":             entry.Stat,
		"points":           entry.Points,
		"source_object_id": entry.SourceObjectID,
		"reason_code":      entry.ReasonCode,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		// Double-check for unique constraint violation (race condition)
		if isUniqueConstraintError(err) {
			return ErrAlreadyAwarded
		}
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	entry.ID = created.ID
	entry.CreatedOn = created.CreatedOn
	return nil
}

// TryAwardPoints attempts to award points, returns (awarded, error)
// This is the preferred method as it handles idempotency transparently
func (r *ResonanceRepository) TryAwardPoints(ctx context.Context, entry *model.ResonanceLedgerEntry) (bool, error) {
	err := r.AwardPoints(ctx, entry)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrAlreadyAwarded) {
		return false, nil // Not an error, just already awarded
	}
	return false, err
}

// isUniqueConstraintError is defined in helpers.go

// HasAwardedPoints checks if points have already been awarded for a source
func (r *ResonanceRepository) HasAwardedPoints(ctx context.Context, userID, stat, sourceObjectID string) (bool, error) {
	query := `
		SELECT count() as count FROM resonance_ledger
		WHERE user = type::record($user_id)
		AND stat = $stat
		AND source_object_id = $source_object_id
		GROUP ALL
	`
	vars := map[string]interface{}{
		"user_id":          userID,
		"stat":             stat,
		"source_object_id": sourceObjectID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	if data, ok := result.(map[string]interface{}); ok {
		count := getInt(data, "count")
		return count > 0, nil
	}

	return false, nil
}

// GetUserLedger retrieves a user's resonance ledger entries
func (r *ResonanceRepository) GetUserLedger(ctx context.Context, userID string, limit, offset int) ([]*model.ResonanceLedgerEntry, error) {
	query := `
		SELECT * FROM resonance_ledger
		WHERE user = type::record($user_id)
		ORDER BY created_on DESC
		LIMIT $limit START $offset
	`
	vars := map[string]interface{}{
		"user_id": userID,
		"limit":   limit,
		"offset":  offset,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseLedgerResult(result)
}

// GetUserScore retrieves a user's cached resonance score
func (r *ResonanceRepository) GetUserScore(ctx context.Context, userID string) (*model.ResonanceScore, error) {
	query := `SELECT * FROM resonance_score WHERE user = type::record($user_id)`
	vars := map[string]interface{}{"user_id": userID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// Return empty score
			return &model.ResonanceScore{
				UserID: userID,
			}, nil
		}
		return nil, err
	}

	return r.parseScoreResult(result)
}

// RecalculateUserScore recalculates and caches a user's total score
func (r *ResonanceRepository) RecalculateUserScore(ctx context.Context, userID string) (*model.ResonanceScore, error) {
	// Sum up by stat
	query := `
		SELECT stat, math::sum(points) as total
		FROM resonance_ledger
		WHERE user = type::record($user_id)
		GROUP BY stat
	`
	vars := map[string]interface{}{"user_id": userID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	score := &model.ResonanceScore{
		UserID:         userID,
		LastCalculated: time.Now(),
	}

	// Parse stats
	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					if data, ok := item.(map[string]interface{}); ok {
						stat := getString(data, "stat")
						total := getInt(data, "total")

						switch model.ResonanceStat(stat) {
						case model.ResonanceStatQuesting:
							score.Questing = total
						case model.ResonanceStatMana:
							score.Mana = total
						case model.ResonanceStatWayfinder:
							score.Wayfinder = total
						case model.ResonanceStatAttunement:
							score.Attunement = total
						case model.ResonanceStatNexus:
							score.Nexus = total
						}
					}
				}
			}
		}
	}

	score.Total = score.Questing + score.Mana + score.Wayfinder + score.Attunement + score.Nexus

	// Upsert cached score
	upsertQuery := `
		UPSERT resonance_score
		SET
			user = type::record($user_id),
			total = $total,
			questing = $questing,
			mana = $mana,
			wayfinder = $wayfinder,
			attunement = $attunement,
			nexus = $nexus,
			last_calculated = time::now()
		WHERE user = type::record($user_id)
	`
	vars = map[string]interface{}{
		"user_id":    userID,
		"total":      score.Total,
		"questing":   score.Questing,
		"mana":       score.Mana,
		"wayfinder":  score.Wayfinder,
		"attunement": score.Attunement,
		"nexus":      score.Nexus,
	}

	if err := r.db.Execute(ctx, upsertQuery, vars); err != nil {
		return nil, err
	}

	return score, nil
}

// GetDailyCap retrieves a user's daily cap usage
func (r *ResonanceRepository) GetDailyCap(ctx context.Context, userID string, date string) (*model.ResonanceDailyCap, error) {
	query := `SELECT * FROM resonance_daily_cap WHERE user = type::record($user_id) AND date = $date`
	vars := map[string]interface{}{
		"user_id": userID,
		"date":    date,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return &model.ResonanceDailyCap{
				UserID: userID,
				Date:   date,
			}, nil
		}
		return nil, err
	}

	return r.parseDailyCapResult(result)
}

// IncrementDailyCap increments a user's daily cap usage for a stat
func (r *ResonanceRepository) IncrementDailyCap(ctx context.Context, userID string, date string, stat model.ResonanceStat, amount int) error {
	field := ""
	switch stat {
	case model.ResonanceStatQuesting:
		field = "questing_earned"
	case model.ResonanceStatMana:
		field = "mana_earned"
	case model.ResonanceStatWayfinder:
		field = "wayfinder_earned"
	case model.ResonanceStatAttunement:
		field = "attunement_earned"
	default:
		return nil // Nexus doesn't have daily cap
	}

	// SurrealDB 3.0 UPSERT doesn't work with WHERE clause properly
	// Use IF/ELSE pattern instead
	query := fmt.Sprintf(`
		LET $existing = SELECT * FROM resonance_daily_cap WHERE user = type::record($user_id) AND date = $date;
		IF array::len($existing) = 0 {
			CREATE resonance_daily_cap SET
				user = type::record($user_id),
				date = $date,
				%s = $amount
		} ELSE {
			UPDATE resonance_daily_cap SET
				%s += $amount
			WHERE user = type::record($user_id) AND date = $date
		}
	`, field, field)

	vars := map[string]interface{}{
		"user_id": userID,
		"date":    date,
		"amount":  amount,
	}

	_, err := r.db.Query(ctx, query, vars)
	return err
}

// GetSupportPairCount retrieves the count of mana sessions between two users
func (r *ResonanceRepository) GetSupportPairCount(ctx context.Context, helperID, receiverID string) (int, error) {
	query := `
		SELECT count FROM support_pair_count
		WHERE helper = type::record($helper_id) AND receiver = type::record($receiver_id)
	`
	vars := map[string]interface{}{
		"helper_id":   helperID,
		"receiver_id": receiverID,
	}

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

// IncrementSupportPairCount increments the support session count between two users
func (r *ResonanceRepository) IncrementSupportPairCount(ctx context.Context, helperID, receiverID string) error {
	// SurrealDB 3.0 UPSERT doesn't work with WHERE clause properly
	// Use IF/ELSE pattern instead
	query := `
		LET $existing = SELECT * FROM support_pair_count WHERE helper = type::record($helper_id) AND receiver = type::record($receiver_id);
		IF array::len($existing) = 0 {
			CREATE support_pair_count SET
				helper = type::record($helper_id),
				receiver = type::record($receiver_id),
				count = 1,
				last_session = time::now()
		} ELSE {
			UPDATE support_pair_count SET
				count += 1,
				last_session = time::now()
			WHERE helper = type::record($helper_id) AND receiver = type::record($receiver_id)
		}
	`
	vars := map[string]interface{}{
		"helper_id":   helperID,
		"receiver_id": receiverID,
	}

	_, err := r.db.Query(ctx, query, vars)
	return err
}

// ResetExpiredSupportPairCounts resets counts older than 30 days
func (r *ResonanceRepository) ResetExpiredSupportPairCounts(ctx context.Context) error {
	query := `
		UPDATE support_pair_count
		SET count = 0
		WHERE last_session < time::now() - 30d
	`

	return r.db.Execute(ctx, query, nil)
}

// AwardPointsAtomic atomically awards points, updates daily cap, and recalculates score
// This ensures all operations succeed or fail together, preventing exploitation
func (r *ResonanceRepository) AwardPointsAtomic(ctx context.Context, entry *model.ResonanceLedgerEntry, stat model.ResonanceStat) error {
	// Use a single transaction with multiple statements
	batch := database.NewAtomicBatch()

	// 1. Create ledger entry (will fail on unique constraint if already awarded)
	batch.Add(`
		CREATE resonance_ledger CONTENT {
			user: type::record($user_id),
			stat: $stat,
			points: $points,
			source_object_id: $source_object_id,
			reason_code: $reason_code,
			created_on: time::now()
		}
	`, map[string]interface{}{
		"user_id":          entry.UserID,
		"stat":             entry.Stat,
		"points":           entry.Points,
		"source_object_id": entry.SourceObjectID,
		"reason_code":      entry.ReasonCode,
	})

	// 2. Upsert daily cap
	today := time.Now().Format("2006-01-02")
	field := ""
	switch stat {
	case model.ResonanceStatQuesting:
		field = "questing_earned"
	case model.ResonanceStatMana:
		field = "mana_earned"
	case model.ResonanceStatWayfinder:
		field = "wayfinder_earned"
	case model.ResonanceStatAttunement:
		field = "attunement_earned"
	}

	if field != "" {
		batch.Add(fmt.Sprintf(`
			UPSERT resonance_daily_cap
			SET
				user = type::record($user_id),
				date = $date,
				%s = %s + $points
			WHERE user = type::record($user_id) AND date = $date
		`, field, field), map[string]interface{}{
			"user_id": entry.UserID,
			"date":    today,
			"points":  entry.Points,
		})
	}

	// 3. Update cached score (incremental update)
	batch.Add(`
		LET $existing = (SELECT * FROM resonance_score WHERE user = type::record($user_id);
		IF count($existing) = 0 THEN {
			CREATE resonance_score SET
				user = type::record($user_id),
				total = $points,
				questing = IF $stat = "questing" THEN $points ELSE 0 END,
				mana = IF $stat = "mana" THEN $points ELSE 0 END,
				wayfinder = IF $stat = "wayfinder" THEN $points ELSE 0 END,
				attunement = IF $stat = "attunement" THEN $points ELSE 0 END,
				nexus = IF $stat = "nexus" THEN $points ELSE 0 END,
				last_calculated = time::now()
		} ELSE {
			UPDATE resonance_score SET
				total += $points,
				questing += IF $stat = "questing" THEN $points ELSE 0 END,
				mana += IF $stat = "mana" THEN $points ELSE 0 END,
				wayfinder += IF $stat = "wayfinder" THEN $points ELSE 0 END,
				attunement += IF $stat = "attunement" THEN $points ELSE 0 END,
				nexus += IF $stat = "nexus" THEN $points ELSE 0 END,
				last_calculated = time::now()
			WHERE user = type::record($user_id)
		}
	`, map[string]interface{}{
		"user_id": entry.UserID,
		"stat":    string(stat),
		"points":  entry.Points,
	})

	return batch.Execute(ctx, r.db)
}

// Helper functions

func (r *ResonanceRepository) parseLedgerResult(result []interface{}) ([]*model.ResonanceLedgerEntry, error) {
	entries := make([]*model.ResonanceLedgerEntry, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					entry, err := r.parseLedgerEntryResult(item)
					if err != nil {
						continue
					}
					entries = append(entries, entry)
				}
			}
		}
	}

	return entries, nil
}

func (r *ResonanceRepository) parseLedgerEntryResult(result interface{}) (*model.ResonanceLedgerEntry, error) {
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	entry := &model.ResonanceLedgerEntry{
		ID:             convertSurrealID(data["id"]),
		UserID:         convertSurrealID(data["user"]),
		Stat:           model.ResonanceStat(getString(data, "stat")),
		Points:         getInt(data, "points"),
		SourceObjectID: getString(data, "source_object_id"),
		ReasonCode:     getString(data, "reason_code"),
	}

	if t := getTime(data, "created_on"); t != nil {
		entry.CreatedOn = *t
	}

	return entry, nil
}

func (r *ResonanceRepository) parseScoreResult(result interface{}) (*model.ResonanceScore, error) {
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	score := &model.ResonanceScore{
		UserID:     convertSurrealID(data["user"]),
		Total:      getInt(data, "total"),
		Questing:   getInt(data, "questing"),
		Mana:       getInt(data, "mana"),
		Wayfinder:  getInt(data, "wayfinder"),
		Attunement: getInt(data, "attunement"),
		Nexus:      getInt(data, "nexus"),
	}

	if t := getTime(data, "last_calculated"); t != nil {
		score.LastCalculated = *t
	}

	return score, nil
}

func (r *ResonanceRepository) parseDailyCapResult(result interface{}) (*model.ResonanceDailyCap, error) {
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	cap := &model.ResonanceDailyCap{
		UserID:           convertSurrealID(data["user"]),
		Date:             getString(data, "date"),
		QuestingEarned:   getInt(data, "questing_earned"),
		ManaEarned:       getInt(data, "mana_earned"),
		WayfinderEarned:  getInt(data, "wayfinder_earned"),
		AttunementEarned: getInt(data, "attunement_earned"),
	}

	return cap, nil
}

// GetUserCirclesForNexus retrieves circle activity data for a user's Nexus calculation
func (r *ResonanceRepository) GetUserCirclesForNexus(ctx context.Context, userID string) ([]*model.NexusCircleData, error) {
	// This query gets all circles the user is a member of with activity stats
	query := `
		LET $user = type::record($user_id);
		LET $thirty_days_ago = time::now() - 30d;

		SELECT
			id as circle_id,
			name as circle_name,
			(SELECT count() FROM event WHERE circle = $parent.id AND start_time >= $thirty_days_ago GROUP ALL)[0].count OR 0 as total_events,
			(SELECT count() FROM member WHERE circle = $parent.id AND user IN (
				SELECT DISTINCT user FROM event_rsvp
				WHERE event.circle = $parent.id
				AND event.start_time >= $thirty_days_ago
				AND completion_confirmed IS NOT NONE
			) GROUP ALL)[0].count OR 0 as active_members,
			(SELECT count() FROM event_rsvp
				WHERE user = $user
				AND event.circle = $parent.id
				AND event.start_time >= $thirty_days_ago
				AND completion_confirmed IS NOT NONE
				GROUP ALL
			)[0].count OR 0 as user_completions
		FROM circle
		WHERE id IN (SELECT circle FROM member WHERE user = $user)
	`
	vars := map[string]interface{}{
		"user_id": userID,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseNexusCircleData(result)
}

func (r *ResonanceRepository) parseNexusCircleData(result []interface{}) ([]*model.NexusCircleData, error) {
	circles := make([]*model.NexusCircleData, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					if data, ok := item.(map[string]interface{}); ok {
						completions := getInt(data, "user_completions")
						activityFactor := float64(completions) / 3.0
						if activityFactor > 1.0 {
							activityFactor = 1.0
						}

						totalEvents := getInt(data, "total_events")
						activeMembers := getInt(data, "active_members")
						isActive := totalEvents >= 2 && activeMembers >= 3

						circles = append(circles, &model.NexusCircleData{
							CircleID:        convertSurrealID(data["circle_id"]),
							CircleName:      getString(data, "circle_name"),
							ActiveMembers:   activeMembers,
							TotalEvents:     totalEvents,
							UserCompletions: completions,
							ActivityFactor:  activityFactor,
							IsActive:        isActive,
						})
					}
				}
			}
		}
	}

	return circles, nil
}

// GetCirclePairOverlap returns count of users active in both circles
func (r *ResonanceRepository) GetCirclePairOverlap(ctx context.Context, circleID1, circleID2 string) (int, error) {
	query := `
		LET $thirty_days_ago = time::now() - 30d;

		LET $active_in_c1 = (SELECT DISTINCT user FROM event_rsvp
			WHERE event.circle = type::record($circle_id_1)
			AND event.start_time >= $thirty_days_ago
			AND completion_confirmed IS NOT NONE);

		LET $active_in_c2 = (SELECT DISTINCT user FROM event_rsvp
			WHERE event.circle = type::record($circle_id_2)
			AND event.start_time >= $thirty_days_ago
			AND completion_confirmed IS NOT NONE);

		SELECT count() as count FROM $active_in_c1 WHERE user IN $active_in_c2 GROUP ALL
	`
	vars := map[string]interface{}{
		"circle_id_1": circleID1,
		"circle_id_2": circleID2,
	}

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

// GetAllActiveUserIDs returns IDs of users who have been active in the last 30 days
func (r *ResonanceRepository) GetAllActiveUserIDs(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT user.id as user_id FROM event_rsvp
		WHERE event.start_time >= time::now() - 30d
		AND completion_confirmed IS NOT NONE
	`

	result, err := r.db.Query(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0)
	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					if data, ok := item.(map[string]interface{}); ok {
						if userID := convertSurrealID(data["user_id"]); userID != "" {
							userIDs = append(userIDs, userID)
						}
					}
				}
			}
		}
	}

	return userIDs, nil
}
