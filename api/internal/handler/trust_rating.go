package handler

import (
	"net/http"
	"strconv"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// TrustRatingHandler handles trust rating HTTP requests
type TrustRatingHandler struct {
	svc *service.TrustRatingService
}

// NewTrustRatingHandler creates a new trust rating handler
func NewTrustRatingHandler(svc *service.TrustRatingService) *TrustRatingHandler {
	return &TrustRatingHandler{svc: svc}
}

// Create handles POST /v1/trust-ratings
func (h *TrustRatingHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var req model.CreateTrustRatingRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	rating, err := h.svc.Create(ctx, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, rating, nil)
}

// GetByID handles GET /v1/trust-ratings/{ratingId}
func (h *TrustRatingHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	ratingID := r.PathValue("ratingId")

	rating, err := h.svc.GetByID(ctx, ratingID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, rating, nil)
}

// Update handles PATCH /v1/trust-ratings/{ratingId}
func (h *TrustRatingHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	ratingID := r.PathValue("ratingId")

	var req model.UpdateTrustRatingRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	rating, err := h.svc.Update(ctx, ratingID, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, rating, nil)
}

// Delete handles DELETE /v1/trust-ratings/{ratingId}
func (h *TrustRatingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	ratingID := r.PathValue("ratingId")

	if err := h.svc.Delete(ctx, ratingID, userID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetReceivedRatings handles GET /v1/users/{userId}/trust-ratings/received
func (h *TrustRatingHandler) GetReceivedRatings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	targetUserID := r.PathValue("userId")

	limit, offset := getPaginationParams(r)

	ratings, err := h.svc.GetReceivedRatings(ctx, targetUserID, limit, offset)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, ratings, nil, nil)
}

// GetGivenRatings handles GET /v1/users/{userId}/trust-ratings/given
func (h *TrustRatingHandler) GetGivenRatings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	targetUserID := r.PathValue("userId")

	limit, offset := getPaginationParams(r)

	ratings, err := h.svc.GetGivenRatings(ctx, targetUserID, limit, offset)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, ratings, nil, nil)
}

// GetAggregate handles GET /v1/users/{userId}/trust-aggregate
func (h *TrustRatingHandler) GetAggregate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	targetUserID := r.PathValue("userId")

	aggregate, err := h.svc.GetAggregate(ctx, targetUserID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusOK, aggregate, nil)
}

// CreateEndorsement handles POST /v1/trust-ratings/{ratingId}/endorsements
func (h *TrustRatingHandler) CreateEndorsement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}
	ratingID := r.PathValue("ratingId")

	var req model.CreateEndorsementRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	endorsement, err := h.svc.CreateEndorsement(ctx, ratingID, userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteData(w, http.StatusCreated, endorsement, nil)
}

// GetEndorsements handles GET /v1/trust-ratings/{ratingId}/endorsements
func (h *TrustRatingHandler) GetEndorsements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ratingID := r.PathValue("ratingId")

	endorsements, err := h.svc.GetEndorsements(ctx, ratingID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, endorsements, nil, nil)
}

// GetDistrustSignals handles GET /v1/admin/distrust-signals
func (h *TrustRatingHandler) GetDistrustSignals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	minDistrust := 3
	if v := r.URL.Query().Get("min_distrust"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			minDistrust = parsed
		}
	}

	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	signals, err := h.svc.GetDistrustSignals(ctx, minDistrust, limit)
	if err != nil {
		h.handleError(w, err)
		return
	}

	WriteCollection(w, http.StatusOK, signals, nil, nil)
}

// handleError converts service errors to HTTP responses
func (h *TrustRatingHandler) handleError(w http.ResponseWriter, err error) {
	if pd, ok := err.(*model.ProblemDetails); ok {
		WriteError(w, pd)
		return
	}
	WriteError(w, model.NewInternalError("internal server error"))
}

// Helper function for pagination
func getPaginationParams(r *http.Request) (limit, offset int) {
	limit = 50
	offset = 0

	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if v := r.URL.Query().Get("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}
