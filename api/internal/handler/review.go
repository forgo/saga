package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// ReviewHandler handles review and reputation endpoints
type ReviewHandler struct {
	reviewService *service.ReviewService
}

// NewReviewHandler creates a new review handler
func NewReviewHandler(reviewService *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
	}
}

// CreateReview handles POST /v1/reviews - leave feedback
func (h *ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var req model.CreateReviewRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	// Validate
	var fieldErrors []model.FieldError
	if req.RevieweeID == "" {
		fieldErrors = append(fieldErrors, model.FieldError{
			Field:   "reviewee_id",
			Message: "reviewee_id is required",
		})
	}
	if req.Context == "" {
		fieldErrors = append(fieldErrors, model.FieldError{
			Field:   "context",
			Message: "context is required",
		})
	}
	if len(fieldErrors) > 0 {
		WriteError(w, model.NewValidationError(fieldErrors))
		return
	}

	review, err := h.reviewService.CreateReview(r.Context(), userID, &req)
	if err != nil {
		h.handleReviewError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, review, map[string]string{
		"self": "/v1/reviews/" + review.ID,
	})
}

// GetReview handles GET /v1/reviews/{reviewId} - get a specific review
func (h *ReviewHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	reviewID := r.PathValue("reviewId")
	if reviewID == "" {
		WriteError(w, model.NewBadRequestError("review ID required"))
		return
	}

	review, err := h.reviewService.GetReview(r.Context(), reviewID)
	if err != nil {
		h.handleReviewError(w, err)
		return
	}

	WriteData(w, http.StatusOK, review, map[string]string{
		"self": "/v1/reviews/" + reviewID,
	})
}

// GetReviewsGiven handles GET /v1/profile/reviews/given - reviews you've left
func (h *ReviewHandler) GetReviewsGiven(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	limit := 20
	if r.URL.Query().Get("limit") != "" {
		if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	offset := 0
	if r.URL.Query().Get("offset") != "" {
		if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && o >= 0 {
			offset = o
		}
	}

	reviews, err := h.reviewService.GetReviewsGiven(r.Context(), userID, limit, offset)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get reviews"))
		return
	}

	WriteCollection(w, http.StatusOK, reviews, nil, map[string]string{
		"self": "/v1/profile/reviews/given",
	})
}

// GetReviewsReceived handles GET /v1/profile/reviews/received - reviews you've received
func (h *ReviewHandler) GetReviewsReceived(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	limit := 20
	if r.URL.Query().Get("limit") != "" {
		if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	offset := 0
	if r.URL.Query().Get("offset") != "" {
		if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && o >= 0 {
			offset = o
		}
	}

	reviews, err := h.reviewService.GetReviewsReceived(r.Context(), userID, limit, offset)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get reviews"))
		return
	}

	WriteCollection(w, http.StatusOK, reviews, nil, map[string]string{
		"self": "/v1/profile/reviews/received",
	})
}

// GetMyReputation handles GET /v1/profile/reputation - get own reputation
func (h *ReviewHandler) GetMyReputation(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	reputation, err := h.reviewService.GetReputation(r.Context(), userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get reputation"))
		return
	}

	WriteData(w, http.StatusOK, reputation, map[string]string{
		"self": "/v1/profile/reputation",
	})
}

// GetUserReputation handles GET /v1/users/{userId}/reputation - get another user's reputation
func (h *ReviewHandler) GetUserReputation(w http.ResponseWriter, r *http.Request) {
	_ = middleware.GetUserID(r.Context())
	if middleware.GetUserID(r.Context()) == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	targetUserID := r.PathValue("userId")
	if targetUserID == "" {
		WriteError(w, model.NewBadRequestError("user ID required"))
		return
	}

	display, err := h.reviewService.GetReputationDisplay(r.Context(), targetUserID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get reputation"))
		return
	}

	WriteData(w, http.StatusOK, display, map[string]string{
		"self": "/v1/users/" + targetUserID + "/reputation",
	})
}

// GetPositiveTags handles GET /v1/reviews/tags/positive - list positive tags
func (h *ReviewHandler) GetPositiveTags(w http.ResponseWriter, r *http.Request) {
	tags := model.GetPositiveTags()
	WriteCollection(w, http.StatusOK, tags, nil, map[string]string{
		"self": "/v1/reviews/tags/positive",
	})
}

// GetImprovementTags handles GET /v1/reviews/tags/improvement - list improvement tags
func (h *ReviewHandler) GetImprovementTags(w http.ResponseWriter, r *http.Request) {
	tags := model.GetImprovementTags()
	WriteCollection(w, http.StatusOK, tags, nil, map[string]string{
		"self": "/v1/reviews/tags/improvement",
	})
}

func (h *ReviewHandler) handleReviewError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrReviewNotFound):
		WriteError(w, model.NewNotFoundError("review"))
	case errors.Is(err, service.ErrCannotReviewSelf):
		WriteError(w, model.NewBadRequestError("cannot review yourself"))
	case errors.Is(err, service.ErrAlreadyReviewed):
		WriteError(w, model.NewConflictError("already reviewed for this reference"))
	case errors.Is(err, service.ErrInvalidReviewContext):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "context", Message: "invalid review context"},
		}))
	case errors.Is(err, service.ErrTooManyTags):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "tags", Message: "too many tags (max 8)"},
		}))
	case errors.Is(err, service.ErrPrivateNoteTooLong):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "private_note", Message: "private note too long (max 500 characters)"},
		}))
	default:
		WriteError(w, model.NewInternalError("review operation failed"))
	}
}
