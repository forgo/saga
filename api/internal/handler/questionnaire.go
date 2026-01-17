package handler

import (
	"errors"
	"net/http"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// QuestionnaireHandler handles questionnaire endpoints
type QuestionnaireHandler struct {
	questionnaireService *service.QuestionnaireService
	compatibilityService *service.CompatibilityService
}

// NewQuestionnaireHandler creates a new questionnaire handler
func NewQuestionnaireHandler(
	questionnaireService *service.QuestionnaireService,
	compatibilityService *service.CompatibilityService,
) *QuestionnaireHandler {
	return &QuestionnaireHandler{
		questionnaireService: questionnaireService,
		compatibilityService: compatibilityService,
	}
}

// ListQuestions handles GET /v1/questions - list questions
func (h *QuestionnaireHandler) ListQuestions(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	var questions []*model.Question
	var err error

	if category != "" {
		questions, err = h.questionnaireService.GetQuestionsByCategory(r.Context(), category)
	} else {
		questions, err = h.questionnaireService.GetAllQuestions(r.Context())
	}

	if err != nil {
		WriteError(w, model.NewInternalError("failed to list questions"))
		return
	}

	WriteCollection(w, http.StatusOK, questions, nil, map[string]string{
		"self":       "/v1/questions",
		"categories": "/v1/questions/categories",
	})
}

// GetCategories handles GET /v1/questions/categories - list question categories
func (h *QuestionnaireHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories := model.GetQuestionCategories()
	WriteCollection(w, http.StatusOK, categories, nil, map[string]string{
		"self": "/v1/questions/categories",
	})
}

// GetQuestion handles GET /v1/questions/{questionId} - get a specific question
func (h *QuestionnaireHandler) GetQuestion(w http.ResponseWriter, r *http.Request) {
	questionID := r.PathValue("questionId")
	if questionID == "" {
		WriteError(w, model.NewBadRequestError("question ID required"))
		return
	}

	question, err := h.questionnaireService.GetQuestion(r.Context(), questionID)
	if err != nil {
		h.handleQuestionnaireError(w, err)
		return
	}

	WriteData(w, http.StatusOK, question, map[string]string{
		"self": "/v1/questions/" + questionID,
	})
}

// GetUserAnswers handles GET /v1/profile/answers - get own answers
func (h *QuestionnaireHandler) GetUserAnswers(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	answers, err := h.questionnaireService.GetUserAnswers(r.Context(), userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get answers"))
		return
	}

	WriteCollection(w, http.StatusOK, answers, nil, map[string]string{
		"self": "/v1/profile/answers",
	})
}

// GetUserAnswersWithQuestions handles GET /v1/profile/answers/detailed - get answers with question details
func (h *QuestionnaireHandler) GetUserAnswersWithQuestions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	answers, err := h.questionnaireService.GetUserAnswersWithQuestions(r.Context(), userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get answers"))
		return
	}

	WriteCollection(w, http.StatusOK, answers, nil, map[string]string{
		"self": "/v1/profile/answers/detailed",
	})
}

// GetQuestionProgress handles GET /v1/profile/questions/progress - get question progress
func (h *QuestionnaireHandler) GetQuestionProgress(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	progress, err := h.questionnaireService.GetQuestionProgress(r.Context(), userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get progress"))
		return
	}

	WriteData(w, http.StatusOK, progress, map[string]string{
		"self": "/v1/profile/questions/progress",
	})
}

// AnswerQuestion handles POST /v1/questions/{questionId}/answer - answer a question
func (h *QuestionnaireHandler) AnswerQuestion(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	questionID := r.PathValue("questionId")
	if questionID == "" {
		WriteError(w, model.NewBadRequestError("question ID required"))
		return
	}

	var req model.AnswerQuestionRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	// Validate
	var fieldErrors []model.FieldError
	if req.SelectedOption == "" {
		fieldErrors = append(fieldErrors, model.FieldError{
			Field:   "selected_option",
			Message: "selected_option is required",
		})
	}
	if len(fieldErrors) > 0 {
		WriteError(w, model.NewValidationError(fieldErrors))
		return
	}

	answer, err := h.questionnaireService.AnswerQuestion(r.Context(), userID, questionID, &req)
	if err != nil {
		h.handleQuestionnaireError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, answer, map[string]string{
		"self": "/v1/questions/" + questionID + "/answer",
	})
}

// UpdateAnswer handles PATCH /v1/questions/{questionId}/answer - update an answer
func (h *QuestionnaireHandler) UpdateAnswer(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	questionID := r.PathValue("questionId")
	if questionID == "" {
		WriteError(w, model.NewBadRequestError("question ID required"))
		return
	}

	var req model.UpdateAnswerRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	answer, err := h.questionnaireService.UpdateAnswer(r.Context(), userID, questionID, &req)
	if err != nil {
		h.handleQuestionnaireError(w, err)
		return
	}

	WriteData(w, http.StatusOK, answer, map[string]string{
		"self": "/v1/questions/" + questionID + "/answer",
	})
}

// DeleteAnswer handles DELETE /v1/questions/{questionId}/answer - delete an answer
func (h *QuestionnaireHandler) DeleteAnswer(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	questionID := r.PathValue("questionId")
	if questionID == "" {
		WriteError(w, model.NewBadRequestError("question ID required"))
		return
	}

	if err := h.questionnaireService.DeleteAnswer(r.Context(), userID, questionID); err != nil {
		h.handleQuestionnaireError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetCompatibility handles GET /v1/compatibility/{userId} - get compatibility with another user
func (h *QuestionnaireHandler) GetCompatibility(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	targetUserID := r.PathValue("userId")
	if targetUserID == "" {
		WriteError(w, model.NewBadRequestError("user ID required"))
		return
	}

	if userID == targetUserID {
		WriteError(w, model.NewBadRequestError("cannot calculate compatibility with yourself"))
		return
	}

	score, err := h.compatibilityService.CalculateCompatibility(r.Context(), userID, targetUserID)
	if err != nil {
		h.handleQuestionnaireError(w, err)
		return
	}

	WriteData(w, http.StatusOK, score, map[string]string{
		"self": "/v1/compatibility/" + targetUserID,
	})
}

// GetYikesSummary handles GET /v1/compatibility/{userId}/yikes - get yikes flags
func (h *QuestionnaireHandler) GetYikesSummary(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	targetUserID := r.PathValue("userId")
	if targetUserID == "" {
		WriteError(w, model.NewBadRequestError("user ID required"))
		return
	}

	summary, err := h.compatibilityService.CalculateYikesSummary(r.Context(), userID, targetUserID)
	if err != nil {
		h.handleQuestionnaireError(w, err)
		return
	}

	WriteData(w, http.StatusOK, summary, map[string]string{
		"self": "/v1/compatibility/" + targetUserID + "/yikes",
	})
}

func (h *QuestionnaireHandler) handleQuestionnaireError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrQuestionNotFound):
		WriteError(w, model.NewNotFoundError("question"))
	case errors.Is(err, service.ErrAnswerNotFound):
		WriteError(w, model.NewNotFoundError("answer"))
	case errors.Is(err, service.ErrInvalidOption):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "selected_option", Message: "invalid option for this question"},
		}))
	case errors.Is(err, service.ErrInvalidImportance):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "importance", Message: "invalid importance level"},
		}))
	case errors.Is(err, service.ErrDealBreakerNotAllowed):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "is_dealbreaker", Message: "this question is not eligible for dealbreaker status"},
		}))
	case errors.Is(err, service.ErrInvalidAlignmentWeight):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "alignment_weight", Message: "alignment weight must be between 0 and 1"},
		}))
	default:
		WriteError(w, model.NewInternalError("questionnaire operation failed"))
	}
}
