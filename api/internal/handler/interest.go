package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// InterestHandler handles interest endpoints
type InterestHandler struct {
	interestService *service.InterestService
}

// NewInterestHandler creates a new interest handler
func NewInterestHandler(interestService *service.InterestService) *InterestHandler {
	return &InterestHandler{
		interestService: interestService,
	}
}

// ListInterests handles GET /v1/interests - list all interests
func (h *InterestHandler) ListInterests(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	var interests []*model.Interest
	var err error

	if category != "" {
		interests, err = h.interestService.GetInterestsByCategory(r.Context(), category)
	} else {
		interests, err = h.interestService.GetAllInterests(r.Context())
	}

	if err != nil {
		WriteError(w, model.NewInternalError("failed to list interests"))
		return
	}

	WriteCollection(w, http.StatusOK, interests, nil, map[string]string{
		"self": "/v1/interests",
	})
}

// GetCategories handles GET /v1/interests/categories - list interest categories
func (h *InterestHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories := model.GetInterestCategories()
	WriteCollection(w, http.StatusOK, categories, nil, map[string]string{
		"self": "/v1/interests/categories",
	})
}

// GetUserInterests handles GET /v1/profile/interests - get own interests
func (h *InterestHandler) GetUserInterests(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	interests, err := h.interestService.GetUserInterests(r.Context(), userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get interests"))
		return
	}

	WriteCollection(w, http.StatusOK, interests, nil, map[string]string{
		"self": "/v1/profile/interests",
	})
}

// AddUserInterest handles POST /v1/profile/interests - add interest to profile
func (h *InterestHandler) AddUserInterest(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var req model.AddInterestRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	// Validate
	var fieldErrors []model.FieldError
	if req.InterestID == "" {
		fieldErrors = append(fieldErrors, model.FieldError{
			Field:   "interest_id",
			Message: "interest_id is required",
		})
	}
	if len(fieldErrors) > 0 {
		WriteError(w, model.NewValidationError(fieldErrors))
		return
	}

	err := h.interestService.AddUserInterest(r.Context(), userID, req.InterestID, &req)
	if err != nil {
		h.handleInterestError(w, err)
		return
	}

	// Return success - get the user's interests to return
	interests, _ := h.interestService.GetUserInterests(r.Context(), userID)
	WriteCollection(w, http.StatusCreated, interests, nil, map[string]string{
		"self": "/v1/profile/interests",
	})
}

// UpdateUserInterest handles PATCH /v1/profile/interests/{interestId} - update interest settings
func (h *InterestHandler) UpdateUserInterest(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	interestID := r.PathValue("interestId")
	if interestID == "" {
		WriteError(w, model.NewBadRequestError("interest ID required"))
		return
	}

	var req model.UpdateInterestRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	err := h.interestService.UpdateUserInterest(r.Context(), userID, interestID, &req)
	if err != nil {
		h.handleInterestError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveUserInterest handles DELETE /v1/profile/interests/{interestId}
func (h *InterestHandler) RemoveUserInterest(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	interestID := r.PathValue("interestId")
	if interestID == "" {
		WriteError(w, model.NewBadRequestError("interest ID required"))
		return
	}

	if err := h.interestService.RemoveUserInterest(r.Context(), userID, interestID); err != nil {
		h.handleInterestError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// FindTeachingMatches handles GET /v1/interests/matches/teaching - find people to teach
func (h *InterestHandler) FindTeachingMatches(w http.ResponseWriter, r *http.Request) {
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

	matches, err := h.interestService.FindTeachingMatches(r.Context(), userID, limit)
	if err != nil {
		h.handleInterestError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, matches, nil, map[string]string{
		"self": "/v1/interests/matches/teaching",
	})
}

// FindLearningMatches handles GET /v1/interests/matches/learning - find teachers
func (h *InterestHandler) FindLearningMatches(w http.ResponseWriter, r *http.Request) {
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

	matches, err := h.interestService.FindLearningMatches(r.Context(), userID, limit)
	if err != nil {
		h.handleInterestError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, matches, nil, map[string]string{
		"self": "/v1/interests/matches/learning",
	})
}

// FindSharedInterests handles GET /v1/interests/shared - find users with shared interests
func (h *InterestHandler) FindSharedInterests(w http.ResponseWriter, r *http.Request) {
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

	users, err := h.interestService.FindSharedInterests(r.Context(), userID, limit)
	if err != nil {
		h.handleInterestError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, users, nil, map[string]string{
		"self": "/v1/interests/shared",
	})
}

// GetInterestStats handles GET /v1/profile/interests/stats - get user's interest statistics
func (h *InterestHandler) GetInterestStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	stats, err := h.interestService.GetInterestStats(r.Context(), userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get interest stats"))
		return
	}

	WriteData(w, http.StatusOK, stats, map[string]string{
		"self": "/v1/profile/interests/stats",
	})
}

func (h *InterestHandler) handleInterestError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInterestNotFound):
		WriteError(w, model.NewNotFoundError("interest"))
	case errors.Is(err, service.ErrInterestAlreadyExists):
		WriteError(w, model.NewConflictError("interest already added"))
	case errors.Is(err, service.ErrInvalidInterestLevel):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "level", Message: "invalid interest level"},
		}))
	default:
		WriteError(w, model.NewInternalError("interest operation failed"))
	}
}
