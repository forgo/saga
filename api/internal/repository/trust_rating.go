package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// TrustRatingRepository handles trust rating data access
type TrustRatingRepository struct {
	db database.Database
}

// NewTrustRatingRepository creates a new trust rating repository
func NewTrustRatingRepository(db database.Database) *TrustRatingRepository {
	return &TrustRatingRepository{db: db}
}

// Create creates a new trust rating
func (r *TrustRatingRepository) Create(ctx context.Context, rating *model.TrustRating) error {
	// Set review visibility based on trust level
	visibility := model.ReviewVisibilityPublic
	if rating.TrustLevel == model.TrustLevelDistrust {
		visibility = model.ReviewVisibilityAdminOnly
	}

	query := `
		CREATE trust_rating CONTENT {
			rater_id: type::record($rater_id),
			ratee_id: type::record($ratee_id),
			anchor_type: $anchor_type,
			anchor_id: $anchor_id,
			trust_level: $trust_level,
			trust_review: $trust_review,
			review_visibility: $visibility,
			created_on: time::now(),
			updated_on: time::now()
		}
	`
	vars := map[string]interface{}{
		"rater_id":     rating.RaterID,
		"ratee_id":     rating.RateeID,
		"anchor_type":  rating.AnchorType,
		"anchor_id":    rating.AnchorID,
		"trust_level":  rating.TrustLevel,
		"trust_review": rating.TrustReview,
		"visibility":   visibility,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		if isUniqueConstraintError(err) {
			return fmt.Errorf("trust rating already exists for this anchor")
		}
		return fmt.Errorf("failed to create trust rating: %w", err)
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return fmt.Errorf("failed to extract created rating: %w", err)
	}

	rating.ID = created.ID
	rating.CreatedOn = created.CreatedOn
	rating.UpdatedOn = created.UpdatedOn
	rating.ReviewVisibility = visibility
	return nil
}

// GetByID retrieves a trust rating by ID
func (r *TrustRatingRepository) GetByID(ctx context.Context, id string) (*model.TrustRating, error) {
	query := `SELECT * FROM type::record($id)`
	result, err := r.db.QueryOne(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get trust rating: %w", err)
	}

	return r.parseTrustRating(result)
}

// GetByRaterRateeAnchor finds a rating by rater, ratee, and anchor
func (r *TrustRatingRepository) GetByRaterRateeAnchor(ctx context.Context, raterID, rateeID, anchorType, anchorID string) (*model.TrustRating, error) {
	query := `
		SELECT * FROM trust_rating
		WHERE rater_id = type::record($rater_id)
		AND ratee_id = type::record($ratee_id)
		AND anchor_type = $anchor_type
		AND anchor_id = $anchor_id
	`
	vars := map[string]interface{}{
		"rater_id":    raterID,
		"ratee_id":    rateeID,
		"anchor_type": anchorType,
		"anchor_id":   anchorID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get trust rating: %w", err)
	}

	return r.parseTrustRating(result)
}

// Update updates a trust rating
func (r *TrustRatingRepository) Update(ctx context.Context, id string, trustLevel model.TrustLevel, trustReview string) (*model.TrustRating, error) {
	// Set review visibility based on trust level
	visibility := model.ReviewVisibilityPublic
	if trustLevel == model.TrustLevelDistrust {
		visibility = model.ReviewVisibilityAdminOnly
	}

	query := `
		UPDATE type::record($id) SET
			trust_level = $trust_level,
			trust_review = $trust_review,
			review_visibility = $visibility,
			updated_on = time::now()
		RETURN AFTER
	`
	vars := map[string]interface{}{
		"id":           id,
		"trust_level":  trustLevel,
		"trust_review": trustReview,
		"visibility":   visibility,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to update trust rating: %w", err)
	}

	return r.parseTrustRating(result)
}

// Delete deletes a trust rating (sets back to neutral/undecided state)
func (r *TrustRatingRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE type::record($id)`
	if err := r.db.Execute(ctx, query, map[string]interface{}{"id": id}); err != nil {
		return fmt.Errorf("failed to delete trust rating: %w", err)
	}
	return nil
}

// GetReceivedRatings retrieves ratings received by a user (public only)
func (r *TrustRatingRepository) GetReceivedRatings(ctx context.Context, userID string, limit, offset int) ([]*model.TrustRating, error) {
	query := `
		SELECT * FROM trust_rating
		WHERE ratee_id = type::record($user_id)
		AND review_visibility = "public"
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
		return nil, fmt.Errorf("failed to get received ratings: %w", err)
	}

	return r.parseTrustRatings(result)
}

// GetGivenRatings retrieves ratings given by a user
func (r *TrustRatingRepository) GetGivenRatings(ctx context.Context, userID string, limit, offset int) ([]*model.TrustRating, error) {
	query := `
		SELECT * FROM trust_rating
		WHERE rater_id = type::record($user_id)
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
		return nil, fmt.Errorf("failed to get given ratings: %w", err)
	}

	return r.parseTrustRatings(result)
}

// GetAggregate retrieves aggregated trust stats for a user
func (r *TrustRatingRepository) GetAggregate(ctx context.Context, userID string) (*model.TrustAggregate, error) {
	// Use type::record to properly cast the string to a record for comparison
	query := `
		SELECT
			count(trust_level = "trust") as trust_count,
			count(trust_level = "distrust") as distrust_count
		FROM trust_rating
		WHERE ratee_id = type::record($user_id)
		GROUP ALL
	`
	vars := map[string]interface{}{"user_id": userID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, fmt.Errorf("failed to get aggregate: %w", err)
	}

	agg := &model.TrustAggregate{UserID: userID}

	if data, ok := result.(map[string]interface{}); ok {
		agg.TrustCount = getInt(data, "trust_count")
		agg.DistrustCount = getInt(data, "distrust_count")
	}

	// Get endorsement count
	endorseQuery := `
		SELECT count() as cnt FROM trust_endorsement
		WHERE trust_rating_id IN (SELECT id FROM trust_rating WHERE ratee_id = type::record($user_id))
		GROUP ALL
	`
	endorseResult, err := r.db.QueryOne(ctx, endorseQuery, vars)
	if err == nil {
		if data, ok := endorseResult.(map[string]interface{}); ok {
			agg.EndorsementCount = getInt(data, "cnt")
		}
	}

	agg.NetTrust = agg.TrustCount - agg.DistrustCount
	return agg, nil
}

// GetDailyCount gets the current day's rating count for a user
func (r *TrustRatingRepository) GetDailyCount(ctx context.Context, userID string) (int, error) {
	today := time.Now().Format("2006-01-02")
	query := `
		SELECT count FROM trust_rating_daily_count
		WHERE user_id = type::record($user_id) AND date = $date
	`
	vars := map[string]interface{}{
		"user_id": userID,
		"date":    today,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get daily count: %w", err)
	}

	if data, ok := result.(map[string]interface{}); ok {
		return getInt(data, "count"), nil
	}
	return 0, nil
}

// CanRate checks if a user can rate another based on anchor validation
func (r *TrustRatingRepository) CanRate(ctx context.Context, raterID, rateeID, anchorType, anchorID string) (bool, error) {
	query := `fn::can_rate_user($rater_id, $ratee_id, $anchor_type, $anchor_id)`
	vars := map[string]interface{}{
		"rater_id":    raterID,
		"ratee_id":    rateeID,
		"anchor_type": anchorType,
		"anchor_id":   anchorID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return false, fmt.Errorf("failed to check rating eligibility: %w", err)
	}

	if canRate, ok := result.(bool); ok {
		return canRate, nil
	}
	return false, nil
}

// Endorsement operations

// CreateEndorsement creates a new endorsement on a trust rating
func (r *TrustRatingRepository) CreateEndorsement(ctx context.Context, endorsement *model.TrustEndorsement) error {
	// Build query conditionally to avoid NULL for optional note field
	var query string
	vars := map[string]interface{}{
		"rating_id":        endorsement.TrustRatingID,
		"endorser_id":      endorsement.EndorserID,
		"endorsement_type": endorsement.EndorsementType,
	}

	if endorsement.Note != nil && *endorsement.Note != "" {
		query = `
			CREATE trust_endorsement CONTENT {
				trust_rating_id: type::record($rating_id),
				endorser_id: type::record($endorser_id),
				endorsement_type: $endorsement_type,
				note: $note,
				created_on: time::now()
			}
		`
		vars["note"] = *endorsement.Note
	} else {
		query = `
			CREATE trust_endorsement CONTENT {
				trust_rating_id: type::record($rating_id),
				endorser_id: type::record($endorser_id),
				endorsement_type: $endorsement_type,
				created_on: time::now()
			}
		`
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		if isUniqueConstraintError(err) {
			return fmt.Errorf("endorsement already exists")
		}
		return fmt.Errorf("failed to create endorsement: %w", err)
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return fmt.Errorf("failed to extract created endorsement: %w", err)
	}

	endorsement.ID = created.ID
	endorsement.CreatedOn = created.CreatedOn
	return nil
}

// GetEndorsementsByRating retrieves all endorsements for a rating
func (r *TrustRatingRepository) GetEndorsementsByRating(ctx context.Context, ratingID string) ([]*model.TrustEndorsement, error) {
	query := `
		SELECT * FROM trust_endorsement
		WHERE trust_rating_id = type::record($rating_id)
		ORDER BY created_on DESC
	`
	vars := map[string]interface{}{"rating_id": ratingID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get endorsements: %w", err)
	}

	return r.parseEndorsements(result)
}

// GetEndorsementCounts gets agree/disagree counts for a rating
func (r *TrustRatingRepository) GetEndorsementCounts(ctx context.Context, ratingID string) (agree, disagree int, err error) {
	query := `
		SELECT
			count(endorsement_type = "agree") as agree,
			count(endorsement_type = "disagree") as disagree
		FROM trust_endorsement
		WHERE trust_rating_id = type::record($rating_id)
		GROUP ALL
	`
	vars := map[string]interface{}{"rating_id": ratingID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return 0, 0, nil
		}
		return 0, 0, fmt.Errorf("failed to get endorsement counts: %w", err)
	}

	if data, ok := result.(map[string]interface{}); ok {
		return getInt(data, "agree"), getInt(data, "disagree"), nil
	}
	return 0, 0, nil
}

// HasEndorsed checks if a user has already endorsed a rating
func (r *TrustRatingRepository) HasEndorsed(ctx context.Context, endorserID, ratingID string) (bool, error) {
	query := `
		SELECT count() as count FROM trust_endorsement
		WHERE endorser_id = type::record($endorser_id)
		AND trust_rating_id = type::record($rating_id)
		GROUP ALL
	`
	vars := map[string]interface{}{
		"endorser_id": endorserID,
		"rating_id":   ratingID,
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

// Admin functions

// GetDistrustSignals retrieves users with significant distrust for admin review
func (r *TrustRatingRepository) GetDistrustSignals(ctx context.Context, minDistrust int, limit int) ([]*model.DistrustSignal, error) {
	// SurrealDB 3.0 doesn't support HAVING clause, so we use a subquery approach
	// First aggregate, then filter in application layer
	query := `
		SELECT
			ratee_id,
			count(trust_level = "trust") as trust_count,
			count(trust_level = "distrust") as distrust_count,
			array::first(trust_review) as latest_reason,
			array::first(created_on) as latest_rating_on
		FROM trust_rating
		GROUP BY ratee_id
		ORDER BY distrust_count DESC
	`

	result, err := r.db.Query(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get distrust signals: %w", err)
	}

	// Filter in application layer since HAVING is not supported
	signals, err := r.parseDistrustSignals(result)
	if err != nil {
		return nil, err
	}

	// Apply minimum distrust filter and limit
	filtered := make([]*model.DistrustSignal, 0)
	for _, s := range signals {
		if s.DistrustCount >= minDistrust {
			filtered = append(filtered, s)
			if len(filtered) >= limit {
				break
			}
		}
	}

	return filtered, nil
}

// Parsing helpers

func (r *TrustRatingRepository) parseTrustRating(result interface{}) (*model.TrustRating, error) {
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	rating := &model.TrustRating{
		ID:               convertSurrealID(data["id"]),
		RaterID:          convertSurrealID(data["rater_id"]),
		RateeID:          convertSurrealID(data["ratee_id"]),
		AnchorType:       model.TrustAnchorType(getString(data, "anchor_type")),
		AnchorID:         getString(data, "anchor_id"),
		TrustLevel:       model.TrustLevel(getString(data, "trust_level")),
		TrustReview:      getString(data, "trust_review"),
		ReviewVisibility: model.ReviewVisibility(getString(data, "review_visibility")),
	}

	if t := getTime(data, "created_on"); t != nil {
		rating.CreatedOn = *t
	}
	if t := getTime(data, "updated_on"); t != nil {
		rating.UpdatedOn = *t
	}

	return rating, nil
}

func (r *TrustRatingRepository) parseTrustRatings(result []interface{}) ([]*model.TrustRating, error) {
	ratings := make([]*model.TrustRating, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					rating, err := r.parseTrustRating(item)
					if err != nil {
						continue
					}
					ratings = append(ratings, rating)
				}
			}
		}
	}

	return ratings, nil
}

func (r *TrustRatingRepository) parseEndorsements(result []interface{}) ([]*model.TrustEndorsement, error) {
	endorsements := make([]*model.TrustEndorsement, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					if data, ok := item.(map[string]interface{}); ok {
						endorsement := &model.TrustEndorsement{
							ID:              convertSurrealID(data["id"]),
							TrustRatingID:   convertSurrealID(data["trust_rating_id"]),
							EndorserID:      convertSurrealID(data["endorser_id"]),
							EndorsementType: model.EndorsementType(getString(data, "endorsement_type")),
						}
						if note := getString(data, "note"); note != "" {
							endorsement.Note = &note
						}
						if t := getTime(data, "created_on"); t != nil {
							endorsement.CreatedOn = *t
						}
						endorsements = append(endorsements, endorsement)
					}
				}
			}
		}
	}

	return endorsements, nil
}

func (r *TrustRatingRepository) parseDistrustSignals(result []interface{}) ([]*model.DistrustSignal, error) {
	signals := make([]*model.DistrustSignal, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					if data, ok := item.(map[string]interface{}); ok {
						signal := &model.DistrustSignal{
							UserID:        convertSurrealID(data["ratee_id"]),
							TrustCount:    getInt(data, "trust_count"),
							DistrustCount: getInt(data, "distrust_count"),
							LatestReason:  getString(data, "latest_reason"),
						}
						signal.NetTrust = signal.TrustCount - signal.DistrustCount
						if t := getTime(data, "latest_rating_on"); t != nil {
							signal.LatestRatingOn = *t
						}
						signals = append(signals, signal)
					}
				}
			}
		}
	}

	return signals, nil
}
