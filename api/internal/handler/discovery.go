package handler

import (
	"net/http"
	"strconv"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// DiscoveryHandler handles discovery endpoints for global people matching
type DiscoveryHandler struct {
	discoveryService *service.DiscoveryService
}

// NewDiscoveryHandler creates a new discovery handler
func NewDiscoveryHandler(discoveryService *service.DiscoveryService) *DiscoveryHandler {
	return &DiscoveryHandler{
		discoveryService: discoveryService,
	}
}

// DiscoverPeople handles GET /v1/discover/people - find compatible people
// Query parameters:
//   - lat: center latitude (optional)
//   - lng: center longitude (optional)
//   - radius: search radius in km (optional, default: 25)
//   - hangout_type: filter by hangout type (optional, can repeat)
//   - interest_id: filter by specific interest (optional)
//   - min_compatibility: minimum compatibility score 0-100 (optional)
//   - limit: max results (optional, default: 20, max: 50)
//   - offset: pagination offset (optional)
func (h *DiscoveryHandler) DiscoverPeople(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	filter := service.PeopleDiscoveryFilter{}

	// Parse location
	if lat := r.URL.Query().Get("lat"); lat != "" {
		latF, err := strconv.ParseFloat(lat, 64)
		if err == nil {
			filter.CenterLat = &latF
		}
	}
	if lng := r.URL.Query().Get("lng"); lng != "" {
		lngF, err := strconv.ParseFloat(lng, 64)
		if err == nil {
			filter.CenterLng = &lngF
		}
	}

	// Parse radius
	if radius := r.URL.Query().Get("radius"); radius != "" {
		radiusF, err := strconv.ParseFloat(radius, 64)
		if err == nil && radiusF > 0 {
			filter.RadiusKm = radiusF
		}
	}

	// Parse hangout types (can be multiple)
	hangoutTypes := r.URL.Query()["hangout_type"]
	for _, ht := range hangoutTypes {
		filter.HangoutTypes = append(filter.HangoutTypes, model.HangoutType(ht))
	}

	// Parse interest filter
	if interestID := r.URL.Query().Get("interest_id"); interestID != "" {
		filter.InterestID = &interestID
	}

	// Parse compatibility threshold
	if minCompat := r.URL.Query().Get("min_compatibility"); minCompat != "" {
		compatF, err := strconv.ParseFloat(minCompat, 64)
		if err == nil && compatF >= 0 && compatF <= 100 {
			filter.MinCompatibility = compatF
		}
	}

	// Parse pagination
	if limit := r.URL.Query().Get("limit"); limit != "" {
		limitI, err := strconv.Atoi(limit)
		if err == nil && limitI > 0 {
			filter.Limit = limitI
		}
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		offsetI, err := strconv.Atoi(offset)
		if err == nil && offsetI >= 0 {
			filter.Offset = offsetI
		}
	}

	// Parse require_shared_answer
	if reqShared := r.URL.Query().Get("require_shared_answer"); reqShared == "true" {
		filter.RequireSharedAnswer = true
	}

	response, err := h.discoveryService.DiscoverPeople(r.Context(), userID, filter)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to discover people"))
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

// DiscoverByInterest handles GET /v1/discover/interest/{interestId} - find people with a specific interest
func (h *DiscoveryHandler) DiscoverByInterest(w http.ResponseWriter, r *http.Request) {
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

	// Parse limit
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	results, err := h.discoveryService.DiscoverByInterest(r.Context(), userID, interestID, limit)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to discover by interest"))
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"results":     results,
		"total_count": len(results),
	})
}

// DiscoverTeachLearn handles GET /v1/discover/teach-learn - find teach/learn matches
func (h *DiscoveryHandler) DiscoverTeachLearn(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	// Parse limit
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	results, err := h.discoveryService.FindTeachLearnMatches(r.Context(), userID, limit)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to find teach/learn matches"))
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"results":     results,
		"total_count": len(results),
	})
}

// GetHangoutTypes handles GET /v1/discover/hangout-types - get available hangout types
func (h *DiscoveryHandler) GetHangoutTypes(w http.ResponseWriter, r *http.Request) {
	types := model.GetHangoutTypeInfo()
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"hangout_types": types,
	})
}
