package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// ProfileHandler handles profile endpoints
type ProfileHandler struct {
	profileService *service.ProfileService
}

// NewProfileHandler creates a new profile handler
func NewProfileHandler(profileService *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
	}
}

// ProfileResponse represents a profile in API responses
type ProfileResponse struct {
	UserID     string   `json:"user_id"`
	Bio        *string  `json:"bio,omitempty"`
	Tagline    *string  `json:"tagline,omitempty"`
	Languages  []string `json:"languages,omitempty"`
	Timezone   *string  `json:"timezone,omitempty"`
	City       string   `json:"city,omitempty"`
	Country    string   `json:"country,omitempty"`
	Visibility string   `json:"visibility"`
	CreatedOn  string   `json:"created_on"`
	UpdatedOn  string   `json:"updated_on"`
}

// PublicProfileResponse is what other users see
type PublicProfileResponse struct {
	UserID         string  `json:"user_id"`
	Firstname      *string `json:"firstname,omitempty"`
	Bio            *string `json:"bio,omitempty"`
	Tagline        *string `json:"tagline,omitempty"`
	Languages      []string `json:"languages,omitempty"`
	City           string  `json:"city,omitempty"`
	Country        string  `json:"country,omitempty"`
	Distance       string  `json:"distance,omitempty"`
	ActivityStatus string  `json:"activity_status,omitempty"`
	Compatibility  *float64 `json:"compatibility,omitempty"`
}

// Get handles GET /v1/profile - get own profile
func (h *ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	profile, err := h.profileService.GetOrCreateProfile(r.Context(), userID)
	if err != nil {
		h.handleProfileError(w, err)
		return
	}

	WriteData(w, http.StatusOK, toProfileResponse(profile), map[string]string{
		"self":      "/v1/profile",
		"interests": "/v1/profile/interests",
	})
}

// Update handles PATCH /v1/profile - update own profile
func (h *ProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var req model.UpdateProfileRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	// Validate
	var fieldErrors []model.FieldError
	if req.Bio != nil && len(*req.Bio) > model.MaxBioLength {
		fieldErrors = append(fieldErrors, model.FieldError{
			Field:   "bio",
			Message: "bio must be at most 500 characters",
		})
	}
	if req.Tagline != nil && len(*req.Tagline) > model.MaxTaglineLength {
		fieldErrors = append(fieldErrors, model.FieldError{
			Field:   "tagline",
			Message: "tagline must be at most 100 characters",
		})
	}
	if len(req.Languages) > model.MaxLanguages {
		fieldErrors = append(fieldErrors, model.FieldError{
			Field:   "languages",
			Message: "maximum 10 languages allowed",
		})
	}
	if req.Visibility != nil && !isValidVisibility(*req.Visibility) {
		fieldErrors = append(fieldErrors, model.FieldError{
			Field:   "visibility",
			Message: "visibility must be 'circles', 'public', or 'private'",
		})
	}

	if len(fieldErrors) > 0 {
		WriteError(w, model.NewValidationError(fieldErrors))
		return
	}

	profile, err := h.profileService.UpdateProfile(r.Context(), userID, &req)
	if err != nil {
		h.handleProfileError(w, err)
		return
	}

	WriteData(w, http.StatusOK, toProfileResponse(profile), map[string]string{
		"self":      "/v1/profile",
		"interests": "/v1/profile/interests",
	})
}

// GetUser handles GET /v1/users/{userId}/profile - get another user's public profile
func (h *ProfileHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	viewerID := middleware.GetUserID(r.Context())
	if viewerID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	targetUserID := r.PathValue("userId")
	if targetUserID == "" {
		WriteError(w, model.NewBadRequestError("user ID required"))
		return
	}

	// Get viewer's location for distance calculation
	viewerLocation, _ := h.profileService.GetLocationInternal(r.Context(), viewerID)

	profile, err := h.profileService.GetPublicProfile(r.Context(), viewerID, targetUserID, viewerLocation)
	if err != nil {
		h.handleProfileError(w, err)
		return
	}

	WriteData(w, http.StatusOK, toPublicProfileResponse(profile), map[string]string{
		"self": "/v1/users/" + targetUserID + "/profile",
	})
}

// GetNearby handles GET /v1/discover/people - find people nearby
func (h *ProfileHandler) GetNearby(w http.ResponseWriter, r *http.Request) {
	viewerID := middleware.GetUserID(r.Context())
	if viewerID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	// Get viewer's location
	viewerLocation, err := h.profileService.GetLocationInternal(r.Context(), viewerID)
	if err != nil || viewerLocation == nil {
		WriteError(w, model.NewBadRequestError("location required to discover nearby people"))
		return
	}

	// Parse query params
	radiusKm := 25.0 // Default
	if r.URL.Query().Get("radius_km") != "" {
		if radius, err := strconv.ParseFloat(r.URL.Query().Get("radius_km"), 64); err == nil {
			radiusKm = radius
		}
	}

	limit := 20 // Default
	if r.URL.Query().Get("limit") != "" {
		if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	profiles, err := h.profileService.GetNearbyProfiles(r.Context(), viewerID, viewerLocation.Lat, viewerLocation.Lng, radiusKm, limit)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to find nearby people"))
		return
	}

	// Convert to response format
	response := make([]PublicProfileResponse, 0, len(profiles))
	for _, p := range profiles {
		response = append(response, *toPublicProfileResponse(p))
	}

	WriteCollection(w, http.StatusOK, response, nil, map[string]string{
		"self": "/v1/discover/people",
	})
}

func (h *ProfileHandler) handleProfileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrProfileNotFound):
		WriteError(w, model.NewNotFoundError("profile"))
	case errors.Is(err, service.ErrInvalidVisibility):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "visibility", Message: "invalid visibility setting"},
		}))
	case errors.Is(err, service.ErrBioTooLong):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "bio", Message: "bio exceeds maximum length"},
		}))
	case errors.Is(err, service.ErrTaglineTooLong):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "tagline", Message: "tagline exceeds maximum length"},
		}))
	default:
		WriteError(w, model.NewInternalError("profile operation failed"))
	}
}

func toProfileResponse(p *model.UserProfile) ProfileResponse {
	resp := ProfileResponse{
		UserID:     p.UserID,
		Bio:        p.Bio,
		Tagline:    p.Tagline,
		Languages:  p.Languages,
		Timezone:   p.Timezone,
		Visibility: p.Visibility,
		CreatedOn:  p.CreatedOn.Format("2006-01-02T15:04:05Z"),
		UpdatedOn:  p.UpdatedOn.Format("2006-01-02T15:04:05Z"),
	}

	if p.Location != nil {
		resp.City = p.Location.City
		resp.Country = p.Location.Country
	}

	return resp
}

func toPublicProfileResponse(p *model.PublicProfile) *PublicProfileResponse {
	return &PublicProfileResponse{
		UserID:         p.UserID,
		Firstname:      p.Firstname,
		Bio:            p.Bio,
		Tagline:        p.Tagline,
		Languages:      p.Languages,
		City:           p.City,
		Country:        p.Country,
		Distance:       string(p.Distance),
		ActivityStatus: string(p.ActivityStatus),
		Compatibility:  p.Compatibility,
	}
}

func isValidVisibility(v string) bool {
	return v == model.VisibilityGuilds ||
		v == model.VisibilityPublic ||
		v == model.VisibilityPrivate
}
