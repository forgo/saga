package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// ReviewRepository handles review data access
type ReviewRepository struct {
	db database.Database
}

// NewReviewRepository creates a new review repository
func NewReviewRepository(db database.Database) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// Create creates a new review
func (r *ReviewRepository) Create(ctx context.Context, review *model.Review) error {
	query := `
		CREATE review CONTENT {
			reviewer: type::record($reviewer_id),
			reviewee: type::record($reviewee_id),
			context: $context,
			reference_id: $reference_id,
			would_meet_again: $would_meet_again,
			positive_tags: $positive_tags,
			improvement_tags: $improvement_tags,
			private_note: $private_note,
			created_on: time::now()
		}
	`

	vars := map[string]interface{}{
		"reviewer_id":      review.ReviewerID,
		"reviewee_id":      review.RevieweeID,
		"context":          review.Context,
		"reference_id":     review.ReferenceID,
		"would_meet_again": review.WouldMeetAgain,
		"positive_tags":    review.PositiveTags,
		"improvement_tags": review.ImprovementTags,
		"private_note":     review.PrivateNote,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	review.ID = created.ID
	review.CreatedOn = created.CreatedOn
	return nil
}

// GetByID retrieves a review by ID
func (r *ReviewRepository) GetByID(ctx context.Context, id string) (*model.Review, error) {
	// Direct record access - more efficient than WHERE id =
	query := `SELECT * FROM type::record($id)`
	vars := map[string]interface{}{"id": id}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.parseReviewResult(result)
}

// GetReviewsGiven retrieves reviews given by a user
func (r *ReviewRepository) GetReviewsGiven(ctx context.Context, userID string, limit, offset int) ([]*model.Review, error) {
	query := `
		SELECT * FROM review
		WHERE reviewer = type::record($user_id)
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

	return r.parseReviewsResult(result)
}

// GetReviewsReceived retrieves reviews received by a user
func (r *ReviewRepository) GetReviewsReceived(ctx context.Context, userID string, limit, offset int) ([]*model.Review, error) {
	query := `
		SELECT * FROM review
		WHERE reviewee = type::record($user_id)
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

	return r.parseReviewsResult(result)
}

// HasReviewed checks if a user has already reviewed another for a specific reference
func (r *ReviewRepository) HasReviewed(ctx context.Context, reviewerID, revieweeID, referenceID string) (bool, error) {
	query := `
		SELECT count() as count FROM review
		WHERE reviewer = type::record($reviewer_id)
		AND reviewee = type::record($reviewee_id)
		AND reference_id = $reference_id
		GROUP ALL
	`
	vars := map[string]interface{}{
		"reviewer_id":  reviewerID,
		"reviewee_id":  revieweeID,
		"reference_id": referenceID,
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

// GetReputation retrieves aggregated reputation for a user
func (r *ReviewRepository) GetReputation(ctx context.Context, userID string) (*model.Reputation, error) {
	// Count total reviews and would_meet_again
	countQuery := `
		SELECT
			count() as total,
			count(would_meet_again = true) as would_meet_again
		FROM review
		WHERE reviewee = type::record($user_id)
		GROUP ALL
	`
	vars := map[string]interface{}{"user_id": userID}

	countResult, err := r.db.QueryOne(ctx, countQuery, vars)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}

	reputation := &model.Reputation{
		UserID: userID,
	}

	if data, ok := countResult.(map[string]interface{}); ok {
		reputation.TotalReviews = getInt(data, "total")
		reputation.WouldMeetAgain = getInt(data, "would_meet_again")
		if reputation.TotalReviews > 0 {
			reputation.WouldMeetAgainPct = float64(reputation.WouldMeetAgain) / float64(reputation.TotalReviews) * 100
		}
	}

	// Get top positive tags
	tagsQuery := `
		SELECT tag, count() as count FROM (
			SELECT positive_tags as tag FROM review
			WHERE reviewee = type::record($user_id)
			SPLIT tag
		)
		GROUP BY tag
		ORDER BY count DESC
		LIMIT 5
	`

	tagsResult, err := r.db.Query(ctx, tagsQuery, vars)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}

	reputation.TopPositiveTags = r.parseTagCounts(tagsResult)

	return reputation, nil
}

// GetReputationDisplay retrieves display-ready reputation for profiles
func (r *ReviewRepository) GetReputationDisplay(ctx context.Context, userID string) (*model.ReputationDisplay, error) {
	rep, err := r.GetReputation(ctx, userID)
	if err != nil {
		return nil, err
	}

	display := &model.ReputationDisplay{
		EventsHosted: rep.EventsHosted,
		TopTags:      rep.TopPositiveTags,
	}

	// Format the ratio string
	if rep.TotalReviews > 0 {
		display.WouldReturnRatio = formatWouldReturnRatio(rep.WouldMeetAgain, rep.TotalReviews)
	}

	return display, nil
}

// GetEventHostCount counts events hosted by a user
func (r *ReviewRepository) GetEventHostCount(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT count() as count FROM event
		WHERE created_by = type::record($user_id)
		GROUP ALL
	`
	vars := map[string]interface{}{"user_id": userID}

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

// Helper functions

func (r *ReviewRepository) parseReviewResult(result interface{}) (*model.Review, error) {
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	review := &model.Review{
		ID:             convertSurrealID(data["id"]),
		ReviewerID:     convertSurrealID(data["reviewer"]),
		RevieweeID:     convertSurrealID(data["reviewee"]),
		Context:        getString(data, "context"),
		WouldMeetAgain: getBool(data, "would_meet_again"),
		PositiveTags:   getStringSlice(data, "positive_tags"),
		ImprovementTags: getStringSlice(data, "improvement_tags"),
	}

	if refID := getString(data, "reference_id"); refID != "" {
		review.ReferenceID = &refID
	}

	if note := getString(data, "private_note"); note != "" {
		review.PrivateNote = &note
	}

	if t := getTime(data, "created_on"); t != nil {
		review.CreatedOn = *t
	}

	return review, nil
}

func (r *ReviewRepository) parseReviewsResult(result []interface{}) ([]*model.Review, error) {
	reviews := make([]*model.Review, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					review, err := r.parseReviewResult(item)
					if err != nil {
						continue
					}
					reviews = append(reviews, review)
				}
			}
		}
	}

	return reviews, nil
}

func (r *ReviewRepository) parseTagCounts(result []interface{}) []model.TagCount {
	tags := make([]model.TagCount, 0)
	positiveTags := model.GetPositiveTags()
	tagLabels := make(map[string]string)
	for _, t := range positiveTags {
		tagLabels[t.Tag] = t.Label
	}

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					if data, ok := item.(map[string]interface{}); ok {
						tag := getString(data, "tag")
						count := getInt(data, "count")
						label := tagLabels[tag]
						if label == "" {
							label = tag
						}
						tags = append(tags, model.TagCount{
							Tag:   tag,
							Label: label,
							Count: count,
						})
					}
				}
			}
		}
	}

	return tags
}

func formatWouldReturnRatio(wouldMeet, total int) string {
	if total == 0 {
		return ""
	}
	return fmt.Sprintf("%d of %d would return", wouldMeet, total)
}
