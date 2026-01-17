package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// InterestRepository handles interest data access
type InterestRepository struct {
	db database.Database
}

// NewInterestRepository creates a new interest repository
func NewInterestRepository(db database.Database) *InterestRepository {
	return &InterestRepository{db: db}
}

// GetAll retrieves all interests
func (r *InterestRepository) GetAll(ctx context.Context) ([]*model.Interest, error) {
	query := `SELECT * FROM interest ORDER BY category, name`

	result, err := r.db.Query(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	return r.parseInterestsResult(result)
}

// GetByCategory retrieves interests by category
func (r *InterestRepository) GetByCategory(ctx context.Context, category string) ([]*model.Interest, error) {
	query := `SELECT * FROM interest WHERE category = $category ORDER BY name`
	vars := map[string]interface{}{"category": category}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseInterestsResult(result)
}

// GetByID retrieves an interest by ID
func (r *InterestRepository) GetByID(ctx context.Context, id string) (*model.Interest, error) {
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

	return r.parseInterestResult(result)
}

// Create creates a new interest
func (r *InterestRepository) Create(ctx context.Context, interest *model.Interest) error {
	query := `
		CREATE interest CONTENT {
			name: $name,
			category: $category,
			icon: $icon,
			created_on: time::now()
		}
	`

	vars := map[string]interface{}{
		"name":     interest.Name,
		"category": interest.Category,
		"icon":     interest.Icon,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	interest.ID = created.ID
	interest.CreatedOn = created.CreatedOn
	return nil
}

// AddUserInterest adds an interest to a user
func (r *InterestRepository) AddUserInterest(ctx context.Context, userID, interestID string, req *model.AddInterestRequest) error {
	query := `
		RELATE $user_id->has_interest->$interest_id CONTENT {
			level: $level,
			wants_to_teach: $wants_to_teach,
			wants_to_learn: $wants_to_learn,
			intent: $intent,
			created_on: time::now()
		}
	`

	vars := map[string]interface{}{
		"user_id":        userID,
		"interest_id":    interestID,
		"level":          req.Level,
		"wants_to_teach": req.WantsToTeach,
		"wants_to_learn": req.WantsToLearn,
		"intent":         req.Intent,
	}

	return r.db.Execute(ctx, query, vars)
}

// UpdateUserInterest updates a user's interest relationship
func (r *InterestRepository) UpdateUserInterest(ctx context.Context, userID, interestID string, updates map[string]interface{}) error {
	query := `UPDATE has_interest SET `

	vars := map[string]interface{}{
		"user_id":     userID,
		"interest_id": interestID,
	}

	first := true
	if level, ok := updates["level"]; ok {
		query += "level = $level"
		vars["level"] = level
		first = false
	}
	if wantsToTeach, ok := updates["wants_to_teach"]; ok {
		if !first {
			query += ", "
		}
		query += "wants_to_teach = $wants_to_teach"
		vars["wants_to_teach"] = wantsToTeach
		first = false
	}
	if wantsToLearn, ok := updates["wants_to_learn"]; ok {
		if !first {
			query += ", "
		}
		query += "wants_to_learn = $wants_to_learn"
		vars["wants_to_learn"] = wantsToLearn
		first = false
	}
	if intent, ok := updates["intent"]; ok {
		if !first {
			query += ", "
		}
		query += "intent = $intent"
		vars["intent"] = intent
	}

	query += ` WHERE in = type::record($user_id) AND out = type::record($interest_id)`

	return r.db.Execute(ctx, query, vars)
}

// RemoveUserInterest removes an interest from a user
func (r *InterestRepository) RemoveUserInterest(ctx context.Context, userID, interestID string) error {
	query := `DELETE has_interest WHERE in = type::record($user_id) AND out = type::record($interest_id)`
	vars := map[string]interface{}{
		"user_id":     userID,
		"interest_id": interestID,
	}

	return r.db.Execute(ctx, query, vars)
}

// GetUserInterests retrieves all interests for a user
func (r *InterestRepository) GetUserInterests(ctx context.Context, userID string) ([]*model.UserInterest, error) {
	query := `
		SELECT
			out.id as interest_id,
			out.name as name,
			out.category as category,
			out.icon as icon,
			level,
			wants_to_teach,
			wants_to_learn,
			intent,
			created_on
		FROM has_interest
		WHERE in = type::record($user_id)
	`
	vars := map[string]interface{}{"user_id": userID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseUserInterestsResult(result)
}

// FindTeachingMatches finds users who want to learn what the user can teach
func (r *InterestRepository) FindTeachingMatches(ctx context.Context, userID string, limit int) ([]*model.InterestMatch, error) {
	query := `
		SELECT
			in as user_id,
			out.id as interest_id,
			out.name as interest_name,
			out.category as category,
			level,
			wants_to_learn
		FROM has_interest
		WHERE out IN (
			SELECT out FROM has_interest
			WHERE in = type::record($user_id) AND wants_to_teach = true
		)
		AND wants_to_learn = true
		AND in != type::record($user_id)
		LIMIT $limit
	`
	vars := map[string]interface{}{
		"user_id": userID,
		"limit":   limit,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseInterestMatchesResult(result)
}

// FindLearningMatches finds users who can teach what the user wants to learn
func (r *InterestRepository) FindLearningMatches(ctx context.Context, userID string, limit int) ([]*model.InterestMatch, error) {
	query := `
		SELECT
			in as user_id,
			out.id as interest_id,
			out.name as interest_name,
			out.category as category,
			level,
			wants_to_teach
		FROM has_interest
		WHERE out IN (
			SELECT out FROM has_interest
			WHERE in = type::record($user_id) AND wants_to_learn = true
		)
		AND wants_to_teach = true
		AND in != type::record($user_id)
		LIMIT $limit
	`
	vars := map[string]interface{}{
		"user_id": userID,
		"limit":   limit,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseInterestMatchesResult(result)
}

// FindSharedInterests finds users with shared interests
func (r *InterestRepository) FindSharedInterests(ctx context.Context, userID string, limit int) ([]*model.SharedInterestUser, error) {
	query := `
		SELECT
			in as user_id,
			count() as shared_count,
			array::group(out.name) as shared_interests
		FROM has_interest
		WHERE out IN (
			SELECT out FROM has_interest WHERE in = type::record($user_id)
		)
		AND in != type::record($user_id)
		GROUP BY in
		ORDER BY shared_count DESC
		LIMIT $limit
	`
	vars := map[string]interface{}{
		"user_id": userID,
		"limit":   limit,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseSharedInterestsResult(result)
}

// Helper functions

func (r *InterestRepository) parseInterestResult(result interface{}) (*model.Interest, error) {
	if result == nil {
		return nil, database.ErrNotFound
	}

	// Handle array wrapper
	if arr, ok := result.([]interface{}); ok {
		if len(arr) == 0 {
			return nil, database.ErrNotFound
		}
		result = arr[0]
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	if id, ok := data["id"]; ok {
		data["id"] = convertSurrealID(id)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var interest model.Interest
	if err := json.Unmarshal(jsonBytes, &interest); err != nil {
		return nil, err
	}

	return &interest, nil
}

func (r *InterestRepository) parseInterestsResult(result []interface{}) ([]*model.Interest, error) {
	interests := make([]*model.Interest, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					interest, err := r.parseInterestResult(item)
					if err != nil {
						continue
					}
					interests = append(interests, interest)
				}
				continue
			}
		}

		interest, err := r.parseInterestResult(res)
		if err != nil {
			continue
		}
		interests = append(interests, interest)
	}

	return interests, nil
}

func (r *InterestRepository) parseUserInterestsResult(result []interface{}) ([]*model.UserInterest, error) {
	userInterests := make([]*model.UserInterest, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					data, ok := item.(map[string]interface{})
					if !ok {
						continue
					}

					ui := &model.UserInterest{
						InterestID:   convertSurrealID(data["interest_id"]),
						Name:         getString(data, "name"),
						Category:     getString(data, "category"),
						Level:        model.InterestLevel(getString(data, "level")),
						WantsToTeach: getBool(data, "wants_to_teach"),
						WantsToLearn: getBool(data, "wants_to_learn"),
					}
					if icon, ok := data["icon"].(string); ok {
						ui.Icon = &icon
					}
					if intent, ok := data["intent"].(string); ok {
						ui.Intent = &intent
					}
					userInterests = append(userInterests, ui)
				}
			}
		}
	}

	return userInterests, nil
}

func (r *InterestRepository) parseInterestMatchesResult(result []interface{}) ([]*model.InterestMatch, error) {
	matches := make([]*model.InterestMatch, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					data, ok := item.(map[string]interface{})
					if !ok {
						continue
					}

					match := &model.InterestMatch{
						UserID:       convertSurrealID(data["user_id"]),
						InterestID:   convertSurrealID(data["interest_id"]),
						InterestName: getString(data, "interest_name"),
						Category:     getString(data, "category"),
						Level:        getString(data, "level"),
						WantsToTeach: getBool(data, "wants_to_teach"),
						WantsToLearn: getBool(data, "wants_to_learn"),
					}
					matches = append(matches, match)
				}
			}
		}
	}

	return matches, nil
}

// GetUsersWithInterest retrieves all users with a specific interest
func (r *InterestRepository) GetUsersWithInterest(ctx context.Context, interestID string) ([]*model.UserInterest, error) {
	query := `
		SELECT
			in.id as user_id,
			out.id as interest_id,
			out.name as name,
			out.category as category,
			out.icon as icon,
			level,
			wants_to_teach,
			wants_to_learn,
			intent,
			created_on
		FROM has_interest
		WHERE out = type::record($interest_id)
	`
	vars := map[string]interface{}{"interest_id": interestID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseUserInterestsResultWithUserID(result)
}

// GetTeachersForInterest retrieves users who want to teach a specific interest
func (r *InterestRepository) GetTeachersForInterest(ctx context.Context, interestID string) ([]*model.UserInterest, error) {
	query := `
		SELECT
			in.id as user_id,
			out.id as interest_id,
			out.name as name,
			out.category as category,
			out.icon as icon,
			level,
			wants_to_teach,
			wants_to_learn,
			intent,
			created_on
		FROM has_interest
		WHERE out = type::record($interest_id) AND wants_to_teach = true
	`
	vars := map[string]interface{}{"interest_id": interestID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseUserInterestsResultWithUserID(result)
}

// GetLearnersForInterest retrieves users who want to learn a specific interest
func (r *InterestRepository) GetLearnersForInterest(ctx context.Context, interestID string) ([]*model.UserInterest, error) {
	query := `
		SELECT
			in.id as user_id,
			out.id as interest_id,
			out.name as name,
			out.category as category,
			out.icon as icon,
			level,
			wants_to_teach,
			wants_to_learn,
			intent,
			created_on
		FROM has_interest
		WHERE out = type::record($interest_id) AND wants_to_learn = true
	`
	vars := map[string]interface{}{"interest_id": interestID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseUserInterestsResultWithUserID(result)
}

// parseUserInterestsResultWithUserID parses user interest results that include user_id
func (r *InterestRepository) parseUserInterestsResultWithUserID(result []interface{}) ([]*model.UserInterest, error) {
	userInterests := make([]*model.UserInterest, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					data, ok := item.(map[string]interface{})
					if !ok {
						continue
					}

					ui := &model.UserInterest{
						UserID:       convertSurrealID(data["user_id"]),
						InterestID:   convertSurrealID(data["interest_id"]),
						Name:         getString(data, "name"),
						Category:     getString(data, "category"),
						Level:        model.InterestLevel(getString(data, "level")),
						WantsToTeach: getBool(data, "wants_to_teach"),
						WantsToLearn: getBool(data, "wants_to_learn"),
					}
					if icon, ok := data["icon"].(string); ok {
						ui.Icon = &icon
					}
					if intent, ok := data["intent"].(string); ok {
						ui.Intent = &intent
					}
					userInterests = append(userInterests, ui)
				}
			}
		}
	}

	return userInterests, nil
}

func (r *InterestRepository) parseSharedInterestsResult(result []interface{}) ([]*model.SharedInterestUser, error) {
	users := make([]*model.SharedInterestUser, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					data, ok := item.(map[string]interface{})
					if !ok {
						continue
					}

					user := &model.SharedInterestUser{
						UserID:      convertSurrealID(data["user_id"]),
						SharedCount: getInt(data, "shared_count"),
					}

					if interests, ok := data["shared_interests"].([]interface{}); ok {
						for _, i := range interests {
							if s, ok := i.(string); ok {
								user.SharedInterests = append(user.SharedInterests, s)
							}
						}
					}

					users = append(users, user)
				}
			}
		}
	}

	return users, nil
}

