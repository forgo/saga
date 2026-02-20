package repository

import (
	"context"
	"errors"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// QuestionnaireRepository handles questionnaire data access
type QuestionnaireRepository struct {
	db database.Database
}

// NewQuestionnaireRepository creates a new questionnaire repository
func NewQuestionnaireRepository(db database.Database) *QuestionnaireRepository {
	return &QuestionnaireRepository{db: db}
}

// GetAllQuestions retrieves all active questions
func (r *QuestionnaireRepository) GetAllQuestions(ctx context.Context) ([]*model.Question, error) {
	query := `SELECT * FROM question WHERE active = true AND circle_id = NONE ORDER BY sort_order`

	result, err := r.db.Query(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	return r.parseQuestionsResult(result)
}

// GetQuestionsByCategory retrieves questions by category
func (r *QuestionnaireRepository) GetQuestionsByCategory(ctx context.Context, category string) ([]*model.Question, error) {
	query := `SELECT * FROM question WHERE category = $category AND active = true AND circle_id = NONE ORDER BY sort_order`
	vars := map[string]interface{}{"category": category}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseQuestionsResult(result)
}

// GetCircleQuestions retrieves questions for a specific circle
func (r *QuestionnaireRepository) GetCircleQuestions(ctx context.Context, circleID string) ([]*model.Question, error) {
	query := `SELECT * FROM question WHERE circle_id = type::record($circle_id) AND active = true ORDER BY sort_order`
	vars := map[string]interface{}{"circle_id": circleID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseQuestionsResult(result)
}

// GetQuestionByID retrieves a question by ID
func (r *QuestionnaireRepository) GetQuestionByID(ctx context.Context, id string) (*model.Question, error) {
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

	return r.parseQuestionResult(result)
}

// CreateQuestion creates a new question (for circle questions)
func (r *QuestionnaireRepository) CreateQuestion(ctx context.Context, question *model.Question) error {
	query := `
		CREATE question CONTENT {
			text: $text,
			category: $category,
			options: $options,
			is_dealbreaker_eligible: $is_dealbreaker_eligible,
			sort_order: $sort_order,
			active: true,
			circle_id: $circle_id,
			created_by: $created_by,
			created_on: time::now()
		}
	`

	// Convert options to proper format
	optionsData := make([]map[string]interface{}, len(question.Options))
	for i, opt := range question.Options {
		optionsData[i] = map[string]interface{}{
			"value":         opt.Value,
			"label":         opt.Label,
			"implicit_bias": opt.ImplicitBias,
		}
	}

	vars := map[string]interface{}{
		"text":                    question.Text,
		"category":                question.Category,
		"options":                 optionsData,
		"is_dealbreaker_eligible": question.IsDealBreakerEligible,
		"sort_order":              question.SortOrder,
		"circle_id":               question.CircleID,
		"created_by":              question.CreatedBy,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	question.ID = created.ID
	question.CreatedOn = created.CreatedOn
	return nil
}

// GetUserAnswer retrieves a user's answer to a specific question
func (r *QuestionnaireRepository) GetUserAnswer(ctx context.Context, userID, questionID string) (*model.Answer, error) {
	query := `SELECT * FROM answer WHERE user = type::record($user_id) AND question = type::record($question_id)`
	vars := map[string]interface{}{
		"user_id":     userID,
		"question_id": questionID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.parseAnswerResult(result)
}

// GetUserAnswers retrieves all answers for a user
func (r *QuestionnaireRepository) GetUserAnswers(ctx context.Context, userID string) ([]*model.Answer, error) {
	query := `SELECT * FROM answer WHERE user = type::record($user_id)`
	vars := map[string]interface{}{"user_id": userID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseAnswersResult(result)
}

// GetUserAnswersWithQuestions retrieves all answers with question details
func (r *QuestionnaireRepository) GetUserAnswersWithQuestions(ctx context.Context, userID string) ([]*model.AnswerWithQuestion, error) {
	query := `
		SELECT
			*,
			question.* as question_data
		FROM answer
		WHERE user = type::record($user_id)
	`
	vars := map[string]interface{}{"user_id": userID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseAnswersWithQuestionsResult(result)
}

// CreateAnswer creates a new answer
func (r *QuestionnaireRepository) CreateAnswer(ctx context.Context, answer *model.Answer) error {
	query := `
		CREATE answer CONTENT {
			user: type::record($user_id),
			question: type::record($question_id),
			selected_option: $selected_option,
			acceptable_options: $acceptable_options,
			importance: $importance,
			is_dealbreaker: $is_dealbreaker,
			alignment_weight: $alignment_weight,
			yikes_options: $yikes_options,
			created_on: time::now(),
			updated_on: time::now()
		}
	`

	vars := map[string]interface{}{
		"user_id":            answer.UserID,
		"question_id":        answer.QuestionID,
		"selected_option":    answer.SelectedOption,
		"acceptable_options": answer.AcceptableOptions,
		"importance":         answer.Importance,
		"is_dealbreaker":     answer.IsDealBreaker,
		"alignment_weight":   answer.AlignmentWeight,
		"yikes_options":      answer.YikesOptions,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	answer.ID = created.ID
	answer.CreatedOn = created.CreatedOn
	answer.UpdatedOn = created.UpdatedOn
	return nil
}

// UpdateAnswer updates an existing answer
func (r *QuestionnaireRepository) UpdateAnswer(ctx context.Context, userID, questionID string, updates map[string]interface{}) (*model.Answer, error) {
	query := `UPDATE answer SET updated_on = time::now()`
	vars := map[string]interface{}{
		"user_id":     userID,
		"question_id": questionID,
	}

	if selectedOption, ok := updates["selected_option"]; ok {
		query += ", selected_option = $selected_option"
		vars["selected_option"] = selectedOption
	}
	if acceptableOptions, ok := updates["acceptable_options"]; ok {
		query += ", acceptable_options = $acceptable_options"
		vars["acceptable_options"] = acceptableOptions
	}
	if importance, ok := updates["importance"]; ok {
		query += ", importance = $importance"
		vars["importance"] = importance
	}
	if isDealBreaker, ok := updates["is_dealbreaker"]; ok {
		query += ", is_dealbreaker = $is_dealbreaker"
		vars["is_dealbreaker"] = isDealBreaker
	}
	if alignmentWeight, ok := updates["alignment_weight"]; ok {
		query += ", alignment_weight = $alignment_weight"
		vars["alignment_weight"] = alignmentWeight
	}
	if yikesOptions, ok := updates["yikes_options"]; ok {
		query += ", yikes_options = $yikes_options"
		vars["yikes_options"] = yikesOptions
	}

	query += ` WHERE user = type::record($user_id) AND question = type::record($question_id) RETURN AFTER`

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseAnswerResult(result)
}

// DeleteAnswer deletes a user's answer
func (r *QuestionnaireRepository) DeleteAnswer(ctx context.Context, userID, questionID string) error {
	query := `DELETE answer WHERE user = type::record($user_id) AND question = type::record($question_id)`
	vars := map[string]interface{}{
		"user_id":     userID,
		"question_id": questionID,
	}

	return r.db.Execute(ctx, query, vars)
}

// GetSharedAnswers retrieves answers for questions both users have answered
func (r *QuestionnaireRepository) GetSharedAnswers(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error) {
	query := `
		SELECT * FROM answer WHERE user = type::record($user_a_id)
		AND question IN (SELECT question FROM answer WHERE user = type::record($user_b_id)
	`
	vars := map[string]interface{}{
		"user_a_id": userAID,
		"user_b_id": userBID,
	}

	resultA, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}
	answersA, err := r.parseAnswersResult(resultA)
	if err != nil {
		return nil, err
	}

	// Get user B's answers for the same questions
	questionIDs := make([]string, len(answersA))
	for i, a := range answersA {
		questionIDs[i] = a.QuestionID
	}

	query = `SELECT * FROM answer WHERE user = type::record($user_id) AND question IN $question_ids`
	vars = map[string]interface{}{
		"user_id":      userBID,
		"question_ids": questionIDs,
	}

	resultB, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}
	answersB, err := r.parseAnswersResult(resultB)
	if err != nil {
		return nil, err
	}

	// Build map by question ID
	answersMap := make(map[string][2]*model.Answer)
	for _, a := range answersA {
		pair := answersMap[a.QuestionID]
		pair[0] = a
		answersMap[a.QuestionID] = pair
	}
	for _, b := range answersB {
		pair := answersMap[b.QuestionID]
		pair[1] = b
		answersMap[b.QuestionID] = pair
	}

	return answersMap, nil
}

// GetUserBiasProfile retrieves a user's bias profile
func (r *QuestionnaireRepository) GetUserBiasProfile(ctx context.Context, userID string) (*model.UserBiasProfile, error) {
	query := `SELECT * FROM user_bias_profile WHERE user = type::record($user_id)`
	vars := map[string]interface{}{"user_id": userID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	profile := &model.UserBiasProfile{
		UserID:          convertSurrealID(data["user"]),
		AccumulatedBias: getFloat(data, "accumulated_bias"),
		AnswerCount:     getInt(data, "answer_count"),
		Status:          getString(data, "status"),
	}

	if profile.AnswerCount > 0 {
		profile.AverageBias = profile.AccumulatedBias / float64(profile.AnswerCount)
	}

	if t := getTime(data, "last_updated"); t != nil {
		profile.LastUpdated = *t
	}

	return profile, nil
}

// UpdateUserBiasProfile updates or creates a user's bias profile
func (r *QuestionnaireRepository) UpdateUserBiasProfile(ctx context.Context, userID string, accumulatedBias float64, answerCount int) error {
	// Determine status based on thresholds
	status := model.BiasStatusNormal
	averageBias := accumulatedBias
	if answerCount > 0 {
		averageBias = accumulatedBias / float64(answerCount)
	}
	if averageBias <= model.BiasThresholdConcern/10 { // Use average bias for status
		status = model.BiasStatusConcern
	} else if averageBias <= model.BiasThresholdWarning/10 {
		status = model.BiasStatusWarning
	}

	query := `
		UPSERT user_bias_profile
		SET
			user = type::record($user_id),
			accumulated_bias = $accumulated_bias,
			answer_count = $answer_count,
			status = $status,
			last_updated = time::now()
		WHERE user = type::record($user_id)
	`

	vars := map[string]interface{}{
		"user_id":          userID,
		"accumulated_bias": accumulatedBias,
		"answer_count":     answerCount,
		"status":           status,
	}

	return r.db.Execute(ctx, query, vars)
}

// GetQuestionProgress retrieves a user's progress in answering questions
func (r *QuestionnaireRepository) GetQuestionProgress(ctx context.Context, userID string) (*model.QuestionProgress, error) {
	// Get total questions
	totalQuery := `SELECT count() as total FROM question WHERE active = true AND circle_id = NONE GROUP ALL`
	totalResult, err := r.db.QueryOne(ctx, totalQuery, nil)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}

	totalQuestions := 0
	if data, ok := totalResult.(map[string]interface{}); ok {
		totalQuestions = getInt(data, "total")
	}

	// Get answered count by category
	answeredQuery := `
		SELECT
			question.category as category,
			count() as count
		FROM answer
		WHERE user = type::record($user_id)
		GROUP BY question.category
	`
	vars := map[string]interface{}{"user_id": userID}
	answeredResult, err := r.db.Query(ctx, answeredQuery, vars)
	if err != nil {
		return nil, err
	}

	byCategory := make(map[string]int)
	answeredCount := 0

	for _, res := range answeredResult {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					if data, ok := item.(map[string]interface{}); ok {
						cat := getString(data, "category")
						count := getInt(data, "count")
						byCategory[cat] = count
						answeredCount += count
					}
				}
			}
		}
	}

	// Check required categories
	missingRequired := []string{}
	for _, cat := range model.RequiredCategories {
		if byCategory[cat] == 0 {
			missingRequired = append(missingRequired, cat)
		}
	}

	canDiscover := answeredCount >= model.MinQuestionsForDiscovery && len(missingRequired) == 0

	return &model.QuestionProgress{
		TotalQuestions:     totalQuestions,
		AnsweredCount:      answeredCount,
		ByCategory:         byCategory,
		RequiredCategories: missingRequired,
		CanDiscover:        canDiscover,
	}, nil
}

// Helper functions

func (r *QuestionnaireRepository) parseQuestionResult(result interface{}) (*model.Question, error) {
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
	if circleID, ok := data["circle_id"]; ok && circleID != nil {
		id := convertSurrealID(circleID)
		data["circle_id"] = id
	}
	if createdBy, ok := data["created_by"]; ok && createdBy != nil {
		id := convertSurrealID(createdBy)
		data["created_by"] = id
	}

	// Parse options with implicit_bias
	var options []model.QuestionOption
	if opts, ok := data["options"].([]interface{}); ok {
		for _, opt := range opts {
			if optMap, ok := opt.(map[string]interface{}); ok {
				qo := model.QuestionOption{
					Value:        getString(optMap, "value"),
					Label:        getString(optMap, "label"),
					ImplicitBias: getFloat(optMap, "implicit_bias"),
				}
				options = append(options, qo)
			}
		}
	}

	question := &model.Question{
		ID:                    convertSurrealID(data["id"]),
		Text:                  getString(data, "text"),
		Category:              getString(data, "category"),
		Options:               options,
		IsDealBreakerEligible: getBool(data, "is_dealbreaker_eligible"),
		SortOrder:             getInt(data, "sort_order"),
		Active:                getBool(data, "active"),
	}

	if circleID, ok := data["circle_id"].(string); ok && circleID != "" {
		question.CircleID = &circleID
	}
	if createdBy, ok := data["created_by"].(string); ok && createdBy != "" {
		question.CreatedBy = &createdBy
	}
	if t := getTime(data, "created_on"); t != nil {
		question.CreatedOn = *t
	}

	return question, nil
}

func (r *QuestionnaireRepository) parseQuestionsResult(result []interface{}) ([]*model.Question, error) {
	questions := make([]*model.Question, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					question, err := r.parseQuestionResult(item)
					if err != nil {
						continue
					}
					questions = append(questions, question)
				}
				continue
			}
		}

		question, err := r.parseQuestionResult(res)
		if err != nil {
			continue
		}
		questions = append(questions, question)
	}

	return questions, nil
}

func (r *QuestionnaireRepository) parseAnswerResult(result interface{}) (*model.Answer, error) {
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

	answer := &model.Answer{
		ID:              convertSurrealID(data["id"]),
		UserID:          convertSurrealID(data["user"]),
		QuestionID:      convertSurrealID(data["question"]),
		SelectedOption:  getString(data, "selected_option"),
		Importance:      getString(data, "importance"),
		IsDealBreaker:   getBool(data, "is_dealbreaker"),
		AlignmentWeight: getFloat(data, "alignment_weight"),
	}

	// Parse acceptable_options array
	if opts, ok := data["acceptable_options"].([]interface{}); ok {
		for _, opt := range opts {
			if s, ok := opt.(string); ok {
				answer.AcceptableOptions = append(answer.AcceptableOptions, s)
			}
		}
	}

	// Parse yikes_options array
	if opts, ok := data["yikes_options"].([]interface{}); ok {
		for _, opt := range opts {
			if s, ok := opt.(string); ok {
				answer.YikesOptions = append(answer.YikesOptions, s)
			}
		}
	}

	if t := getTime(data, "created_on"); t != nil {
		answer.CreatedOn = *t
	}
	if t := getTime(data, "updated_on"); t != nil {
		answer.UpdatedOn = *t
	}

	return answer, nil
}

func (r *QuestionnaireRepository) parseAnswersResult(result []interface{}) ([]*model.Answer, error) {
	answers := make([]*model.Answer, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					answer, err := r.parseAnswerResult(item)
					if err != nil {
						continue
					}
					answers = append(answers, answer)
				}
				continue
			}
		}

		answer, err := r.parseAnswerResult(res)
		if err != nil {
			continue
		}
		answers = append(answers, answer)
	}

	return answers, nil
}

func (r *QuestionnaireRepository) parseAnswersWithQuestionsResult(result []interface{}) ([]*model.AnswerWithQuestion, error) {
	awqs := make([]*model.AnswerWithQuestion, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					data, ok := item.(map[string]interface{})
					if !ok {
						continue
					}

					answer, err := r.parseAnswerResult(data)
					if err != nil {
						continue
					}

					var question model.Question
					if qData, ok := data["question_data"].(map[string]interface{}); ok {
						q, err := r.parseQuestionResult(qData)
						if err == nil && q != nil {
							question = *q
						}
					}

					awqs = append(awqs, &model.AnswerWithQuestion{
						Answer:   *answer,
						Question: question,
					})
				}
			}
		}
	}

	return awqs, nil
}

// CircleValues methods

// GetCircleValues retrieves circle values questionnaire by ID
func (r *QuestionnaireRepository) GetCircleValues(ctx context.Context, id string) (*model.CircleValues, error) {
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

	return r.parseCircleValuesResult(result)
}

// GetCircleValuesByCircle retrieves all values questionnaires for a circle
func (r *QuestionnaireRepository) GetCircleValuesByCircle(ctx context.Context, circleID string) ([]*model.CircleValues, error) {
	query := `SELECT * FROM circle_values WHERE circle_id = type::record($circle_id)`
	vars := map[string]interface{}{"circle_id": circleID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseCircleValuesResults(result)
}

// CreateCircleValues creates a new circle values questionnaire
func (r *QuestionnaireRepository) CreateCircleValues(ctx context.Context, cv *model.CircleValues) error {
	query := `
		CREATE circle_values CONTENT {
			circle_id: type::record($circle_id),
			name: $name,
			description: $description,
			questions: $questions,
			required: $required,
			created_by: type::record($created_by),
			created_on: time::now(),
			updated_on: time::now()
		}
	`

	vars := map[string]interface{}{
		"circle_id":   cv.CircleID,
		"name":        cv.Name,
		"description": cv.Description,
		"questions":   cv.Questions,
		"required":    cv.Required,
		"created_by":  cv.CreatedBy,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	cv.ID = created.ID
	cv.CreatedOn = created.CreatedOn
	cv.UpdatedOn = created.UpdatedOn
	return nil
}

func (r *QuestionnaireRepository) parseCircleValuesResult(result interface{}) (*model.CircleValues, error) {
	if result == nil {
		return nil, database.ErrNotFound
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	cv := &model.CircleValues{
		ID:        convertSurrealID(data["id"]),
		CircleID:  convertSurrealID(data["circle_id"]),
		Name:      getString(data, "name"),
		Required:  getBool(data, "required"),
		CreatedBy: convertSurrealID(data["created_by"]),
	}

	if desc, ok := data["description"].(string); ok {
		cv.Description = &desc
	}

	// Parse questions array
	if qs, ok := data["questions"].([]interface{}); ok {
		for _, q := range qs {
			cv.Questions = append(cv.Questions, convertSurrealID(q))
		}
	}

	if t := getTime(data, "created_on"); t != nil {
		cv.CreatedOn = *t
	}
	if t := getTime(data, "updated_on"); t != nil {
		cv.UpdatedOn = *t
	}

	return cv, nil
}

func (r *QuestionnaireRepository) parseCircleValuesResults(result []interface{}) ([]*model.CircleValues, error) {
	cvs := make([]*model.CircleValues, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					cv, err := r.parseCircleValuesResult(item)
					if err != nil {
						continue
					}
					cvs = append(cvs, cv)
				}
			}
		}
	}

	return cvs, nil
}
