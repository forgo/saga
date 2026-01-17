package service

import (
	"context"
	"errors"

	"github.com/forgo/saga/api/internal/model"
)

// Questionnaire service errors
var (
	ErrQuestionNotFound      = errors.New("question not found")
	ErrAnswerNotFound        = errors.New("answer not found")
	ErrInvalidOption         = errors.New("invalid option selected")
	ErrInvalidImportance     = errors.New("invalid importance level")
	ErrDealBreakerNotAllowed = errors.New("this question cannot be a dealbreaker")
	ErrInvalidAlignmentWeight = errors.New("alignment weight must be between 0 and 1")
)

// QuestionnaireRepository defines the interface for questionnaire storage
type QuestionnaireRepository interface {
	GetAllQuestions(ctx context.Context) ([]*model.Question, error)
	GetQuestionsByCategory(ctx context.Context, category string) ([]*model.Question, error)
	GetCircleQuestions(ctx context.Context, circleID string) ([]*model.Question, error)
	GetQuestionByID(ctx context.Context, id string) (*model.Question, error)
	CreateQuestion(ctx context.Context, question *model.Question) error
	GetUserAnswer(ctx context.Context, userID, questionID string) (*model.Answer, error)
	GetUserAnswers(ctx context.Context, userID string) ([]*model.Answer, error)
	GetUserAnswersWithQuestions(ctx context.Context, userID string) ([]*model.AnswerWithQuestion, error)
	CreateAnswer(ctx context.Context, answer *model.Answer) error
	UpdateAnswer(ctx context.Context, userID, questionID string, updates map[string]interface{}) (*model.Answer, error)
	DeleteAnswer(ctx context.Context, userID, questionID string) error
	GetSharedAnswers(ctx context.Context, userAID, userBID string) (map[string][2]*model.Answer, error)
	GetUserBiasProfile(ctx context.Context, userID string) (*model.UserBiasProfile, error)
	UpdateUserBiasProfile(ctx context.Context, userID string, accumulatedBias float64, answerCount int) error
	GetQuestionProgress(ctx context.Context, userID string) (*model.QuestionProgress, error)
	GetCircleValues(ctx context.Context, id string) (*model.CircleValues, error)
	GetCircleValuesByCircle(ctx context.Context, circleID string) ([]*model.CircleValues, error)
	CreateCircleValues(ctx context.Context, cv *model.CircleValues) error
}

// QuestionnaireService handles questionnaire business logic
type QuestionnaireService struct {
	repo QuestionnaireRepository
}

// QuestionnaireServiceConfig holds configuration for the questionnaire service
type QuestionnaireServiceConfig struct {
	Repo QuestionnaireRepository
}

// NewQuestionnaireService creates a new questionnaire service
func NewQuestionnaireService(cfg QuestionnaireServiceConfig) *QuestionnaireService {
	return &QuestionnaireService{
		repo: cfg.Repo,
	}
}

// GetAllQuestions retrieves all active questions
func (s *QuestionnaireService) GetAllQuestions(ctx context.Context) ([]*model.Question, error) {
	return s.repo.GetAllQuestions(ctx)
}

// GetQuestionsByCategory retrieves questions by category
func (s *QuestionnaireService) GetQuestionsByCategory(ctx context.Context, category string) ([]*model.Question, error) {
	if !isValidQuestionCategory(category) {
		return []*model.Question{}, nil
	}
	return s.repo.GetQuestionsByCategory(ctx, category)
}

// GetQuestion retrieves a single question by ID
func (s *QuestionnaireService) GetQuestion(ctx context.Context, id string) (*model.Question, error) {
	question, err := s.repo.GetQuestionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if question == nil {
		return nil, ErrQuestionNotFound
	}
	return question, nil
}

// GetUserAnswers retrieves all answers for a user
func (s *QuestionnaireService) GetUserAnswers(ctx context.Context, userID string) ([]*model.Answer, error) {
	return s.repo.GetUserAnswers(ctx, userID)
}

// GetUserAnswersWithQuestions retrieves all answers with question details
func (s *QuestionnaireService) GetUserAnswersWithQuestions(ctx context.Context, userID string) ([]*model.AnswerWithQuestion, error) {
	return s.repo.GetUserAnswersWithQuestions(ctx, userID)
}

// GetQuestionProgress retrieves a user's progress
func (s *QuestionnaireService) GetQuestionProgress(ctx context.Context, userID string) (*model.QuestionProgress, error) {
	return s.repo.GetQuestionProgress(ctx, userID)
}

// AnswerQuestion creates or updates an answer to a question
func (s *QuestionnaireService) AnswerQuestion(ctx context.Context, userID, questionID string, req *model.AnswerQuestionRequest) (*model.Answer, error) {
	// Get the question
	question, err := s.repo.GetQuestionByID(ctx, questionID)
	if err != nil {
		return nil, err
	}
	if question == nil {
		return nil, ErrQuestionNotFound
	}

	// Validate selected option
	if !isValidOption(question, req.SelectedOption) {
		return nil, ErrInvalidOption
	}

	// Validate acceptable options
	for _, opt := range req.AcceptableOptions {
		if !isValidOption(question, opt) {
			return nil, ErrInvalidOption
		}
	}

	// Validate yikes options
	for _, opt := range req.YikesOptions {
		if !isValidOption(question, opt) {
			return nil, ErrInvalidOption
		}
	}

	// Validate importance
	importance := req.Importance
	if importance == "" {
		importance = model.ImportanceSomewhat
	}
	if !isValidImportance(importance) {
		return nil, ErrInvalidImportance
	}

	// Validate dealbreaker
	if req.IsDealBreaker && !question.IsDealBreakerEligible {
		return nil, ErrDealBreakerNotAllowed
	}

	// Validate alignment weight
	alignmentWeight := model.DefaultAlignmentWeight
	if req.AlignmentWeight != nil {
		if *req.AlignmentWeight < model.MinAlignmentWeight || *req.AlignmentWeight > model.MaxAlignmentWeight {
			return nil, ErrInvalidAlignmentWeight
		}
		alignmentWeight = *req.AlignmentWeight
	}

	// Default acceptable options to all if not specified
	acceptableOptions := req.AcceptableOptions
	if len(acceptableOptions) == 0 {
		for _, opt := range question.Options {
			acceptableOptions = append(acceptableOptions, opt.Value)
		}
	}

	// Check if answer exists
	existingAnswer, err := s.repo.GetUserAnswer(ctx, userID, questionID)
	if err != nil {
		return nil, err
	}

	var answer *model.Answer
	if existingAnswer != nil {
		// Update existing answer
		updates := map[string]interface{}{
			"selected_option":    req.SelectedOption,
			"acceptable_options": acceptableOptions,
			"importance":         importance,
			"is_dealbreaker":     req.IsDealBreaker,
			"alignment_weight":   alignmentWeight,
			"yikes_options":      req.YikesOptions,
		}
		answer, err = s.repo.UpdateAnswer(ctx, userID, questionID, updates)
	} else {
		// Create new answer
		answer = &model.Answer{
			UserID:            userID,
			QuestionID:        questionID,
			SelectedOption:    req.SelectedOption,
			AcceptableOptions: acceptableOptions,
			Importance:        importance,
			IsDealBreaker:     req.IsDealBreaker,
			AlignmentWeight:   alignmentWeight,
			YikesOptions:      req.YikesOptions,
		}
		err = s.repo.CreateAnswer(ctx, answer)
	}

	if err != nil {
		return nil, err
	}

	// Update bias profile
	go s.updateBiasProfile(context.Background(), userID)

	return answer, nil
}

// UpdateAnswer updates an existing answer
func (s *QuestionnaireService) UpdateAnswer(ctx context.Context, userID, questionID string, req *model.UpdateAnswerRequest) (*model.Answer, error) {
	// Get the question
	question, err := s.repo.GetQuestionByID(ctx, questionID)
	if err != nil {
		return nil, err
	}
	if question == nil {
		return nil, ErrQuestionNotFound
	}

	// Check answer exists
	existing, err := s.repo.GetUserAnswer(ctx, userID, questionID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrAnswerNotFound
	}

	updates := make(map[string]interface{})

	if req.SelectedOption != nil {
		if !isValidOption(question, *req.SelectedOption) {
			return nil, ErrInvalidOption
		}
		updates["selected_option"] = *req.SelectedOption
	}

	if req.AcceptableOptions != nil {
		for _, opt := range req.AcceptableOptions {
			if !isValidOption(question, opt) {
				return nil, ErrInvalidOption
			}
		}
		updates["acceptable_options"] = req.AcceptableOptions
	}

	if req.Importance != nil {
		if !isValidImportance(*req.Importance) {
			return nil, ErrInvalidImportance
		}
		updates["importance"] = *req.Importance
	}

	if req.IsDealBreaker != nil {
		if *req.IsDealBreaker && !question.IsDealBreakerEligible {
			return nil, ErrDealBreakerNotAllowed
		}
		updates["is_dealbreaker"] = *req.IsDealBreaker
	}

	if req.AlignmentWeight != nil {
		if *req.AlignmentWeight < model.MinAlignmentWeight || *req.AlignmentWeight > model.MaxAlignmentWeight {
			return nil, ErrInvalidAlignmentWeight
		}
		updates["alignment_weight"] = *req.AlignmentWeight
	}

	if req.YikesOptions != nil {
		for _, opt := range req.YikesOptions {
			if !isValidOption(question, opt) {
				return nil, ErrInvalidOption
			}
		}
		updates["yikes_options"] = req.YikesOptions
	}

	if len(updates) == 0 {
		return existing, nil
	}

	answer, err := s.repo.UpdateAnswer(ctx, userID, questionID, updates)
	if err != nil {
		return nil, err
	}

	// Update bias profile if selected option changed
	if _, ok := updates["selected_option"]; ok {
		go s.updateBiasProfile(context.Background(), userID)
	}

	return answer, nil
}

// DeleteAnswer deletes an answer
func (s *QuestionnaireService) DeleteAnswer(ctx context.Context, userID, questionID string) error {
	err := s.repo.DeleteAnswer(ctx, userID, questionID)
	if err != nil {
		return err
	}

	// Update bias profile
	go s.updateBiasProfile(context.Background(), userID)

	return nil
}

// GetUserBiasProfile retrieves a user's bias profile (internal only)
func (s *QuestionnaireService) GetUserBiasProfile(ctx context.Context, userID string) (*model.UserBiasProfile, error) {
	return s.repo.GetUserBiasProfile(ctx, userID)
}

// updateBiasProfile recalculates a user's bias score based on all answers
func (s *QuestionnaireService) updateBiasProfile(ctx context.Context, userID string) {
	// Get all answers with questions
	awqs, err := s.repo.GetUserAnswersWithQuestions(ctx, userID)
	if err != nil {
		return
	}

	var accumulatedBias float64
	answerCount := 0

	for _, awq := range awqs {
		// Find the selected option's bias
		for _, opt := range awq.Question.Options {
			if opt.Value == awq.Answer.SelectedOption {
				accumulatedBias += opt.ImplicitBias
				answerCount++
				break
			}
		}
	}

	// Update the profile
	_ = s.repo.UpdateUserBiasProfile(ctx, userID, accumulatedBias, answerCount)
}

// CreateCircleQuestion creates a question specific to a circle
func (s *QuestionnaireService) CreateCircleQuestion(ctx context.Context, circleID, createdBy string, req *model.CreateCircleQuestionRequest) (*model.Question, error) {
	// Validate category
	if !isValidQuestionCategory(req.Category) {
		return nil, errors.New("invalid category")
	}

	// Validate options
	if len(req.Options) < 2 {
		return nil, errors.New("question must have at least 2 options")
	}

	question := &model.Question{
		Text:                  req.Text,
		Category:              req.Category,
		Options:               req.Options,
		IsDealBreakerEligible: req.IsDealBreakerEligible,
		Active:                true,
		CircleID:              &circleID,
		CreatedBy:             &createdBy,
	}

	if err := s.repo.CreateQuestion(ctx, question); err != nil {
		return nil, err
	}

	return question, nil
}

// GetCircleQuestions retrieves questions for a specific circle
func (s *QuestionnaireService) GetCircleQuestions(ctx context.Context, circleID string) ([]*model.Question, error) {
	return s.repo.GetCircleQuestions(ctx, circleID)
}

// CreateCircleValues creates a circle values questionnaire
func (s *QuestionnaireService) CreateCircleValues(ctx context.Context, circleID, createdBy string, req *model.CreateCircleValuesRequest) (*model.CircleValues, error) {
	// Validate questions exist
	for _, qID := range req.Questions {
		q, err := s.repo.GetQuestionByID(ctx, qID)
		if err != nil {
			return nil, err
		}
		if q == nil {
			return nil, ErrQuestionNotFound
		}
	}

	cv := &model.CircleValues{
		CircleID:    circleID,
		Name:        req.Name,
		Description: req.Description,
		Questions:   req.Questions,
		Required:    req.Required,
		CreatedBy:   createdBy,
	}

	if err := s.repo.CreateCircleValues(ctx, cv); err != nil {
		return nil, err
	}

	return cv, nil
}

// GetCircleValues retrieves circle values by ID
func (s *QuestionnaireService) GetCircleValues(ctx context.Context, id string) (*model.CircleValues, error) {
	return s.repo.GetCircleValues(ctx, id)
}

// GetCircleValuesByCircle retrieves all values questionnaires for a circle
func (s *QuestionnaireService) GetCircleValuesByCircle(ctx context.Context, circleID string) ([]*model.CircleValues, error) {
	return s.repo.GetCircleValuesByCircle(ctx, circleID)
}

// Helper functions

func isValidQuestionCategory(category string) bool {
	switch category {
	case model.QuestionCategoryValues,
		model.QuestionCategorySocial,
		model.QuestionCategoryLifestyle,
		model.QuestionCategoryCommunication:
		return true
	default:
		return false
	}
}

func isValidOption(question *model.Question, option string) bool {
	for _, opt := range question.Options {
		if opt.Value == option {
			return true
		}
	}
	return false
}

func isValidImportance(importance string) bool {
	switch importance {
	case model.ImportanceIrrelevant,
		model.ImportanceLittle,
		model.ImportanceSomewhat,
		model.ImportanceVery,
		model.ImportanceMandatory:
		return true
	default:
		return false
	}
}
