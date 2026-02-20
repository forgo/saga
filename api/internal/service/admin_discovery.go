package service

import (
	"context"
	"fmt"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// AdminDiscoveryService wraps existing discovery and compatibility services
// for admin-only operations (exposes exact coordinates, etc.)
type AdminDiscoveryService struct {
	db                   database.Database
	discoveryService     *DiscoveryService
	compatibilityService *CompatibilityService
	geoService           *GeoService
}

// NewAdminDiscoveryService creates a new admin discovery service
func NewAdminDiscoveryService(
	db database.Database,
	discoveryService *DiscoveryService,
	compatibilityService *CompatibilityService,
) *AdminDiscoveryService {
	return &AdminDiscoveryService{
		db:                   db,
		discoveryService:     discoveryService,
		compatibilityService: compatibilityService,
		geoService:           NewGeoService(),
	}
}

// AdminMapUser represents a user with optional coordinates for admin map display
type AdminMapUser struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	Username    *string `json:"username,omitempty"`
	Firstname   *string `json:"firstname,omitempty"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	City        string  `json:"city,omitempty"`
	HasLocation bool    `json:"has_location"`
}

// GetUsersWithLocations returns all users, enriched with location data when available.
// Checks both user_profile (app-created) and profile (seeder-created) tables for locations.
func (s *AdminDiscoveryService) GetUsersWithLocations(ctx context.Context, limit int) ([]AdminMapUser, error) {
	if limit <= 0 {
		limit = 200
	}
	if limit > 500 {
		limit = 500
	}

	// Step 1: Get all users
	userQuery := `
		SELECT id, email, username, firstname
		FROM user
		ORDER BY created_on DESC
		LIMIT $limit
	`
	userResults, err := s.db.Query(ctx, userQuery, map[string]interface{}{
		"limit": limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	userRows := extractResultArray(userResults)
	users := make([]AdminMapUser, 0, len(userRows))
	userMap := make(map[string]*AdminMapUser, len(userRows))

	for _, row := range userRows {
		id := getStringField(row, "id")
		if id == "" {
			continue
		}
		u := &AdminMapUser{
			ID:        id,
			Email:     getStringField(row, "email"),
			Username:  getOptStringField(row, "username"),
			Firstname: getOptStringField(row, "firstname"),
		}
		userMap[id] = u
	}

	// Step 2: Enrich with locations from user_profile table (app-created profiles)
	profileQuery := `
		SELECT user as user_ref, location
		FROM user_profile
		WHERE location != NONE AND location.lat != NONE
	`
	profileResults, err := s.db.Query(ctx, profileQuery, nil)
	if err == nil {
		for _, row := range extractResultArray(profileResults) {
			userID := getStringField(row, "user_ref")
			if u, ok := userMap[userID]; ok && !u.HasLocation {
				if loc, ok := row["location"].(map[string]interface{}); ok {
					if lat, ok := loc["lat"].(float64); ok {
						if lng, ok := loc["lng"].(float64); ok {
							u.Lat = lat
							u.Lng = lng
							u.HasLocation = true
							if city, ok := loc["city"].(string); ok {
								u.City = city
							}
						}
					}
				}
			}
		}
	}

	// Step 3: Enrich with locations from profile table (seeder-created profiles)
	seederQuery := `
		SELECT user_id as user_ref, location
		FROM profile
		WHERE location != NONE AND location.lat != NONE
	`
	seederResults, err := s.db.Query(ctx, seederQuery, nil)
	if err == nil {
		for _, row := range extractResultArray(seederResults) {
			userID := getStringField(row, "user_ref")
			if u, ok := userMap[userID]; ok && !u.HasLocation {
				if loc, ok := row["location"].(map[string]interface{}); ok {
					if lat, ok := loc["lat"].(float64); ok {
						if lng, ok := loc["lng"].(float64); ok {
							u.Lat = lat
							u.Lng = lng
							u.HasLocation = true
							if city, ok := loc["city"].(string); ok {
								u.City = city
							}
						}
					}
				}
			}
		}
	}

	for _, u := range userMap {
		users = append(users, *u)
	}

	return users, nil
}

// AdminDiscoveryRequest defines the simulation request body
type AdminDiscoveryRequest struct {
	ViewerID            string  `json:"viewer_id"`
	RadiusKm            float64 `json:"radius_km,omitempty"`
	MinCompatibility    float64 `json:"min_compatibility,omitempty"`
	RequireSharedAnswer bool    `json:"require_shared_answer"`
	Limit               int     `json:"limit,omitempty"`
}

// AdminDiscoveryResultItem enriches a discovery result with exact coordinates
type AdminDiscoveryResultItem struct {
	UserID             string                `json:"user_id"`
	Email              string                `json:"email,omitempty"`
	Username           *string               `json:"username,omitempty"`
	Firstname          *string               `json:"firstname,omitempty"`
	Lat                float64               `json:"lat"`
	Lng                float64               `json:"lng"`
	City               string                `json:"city,omitempty"`
	CompatibilityScore float64               `json:"compatibility_score"`
	MatchScore         float64               `json:"match_score"`
	DistanceKm         float64               `json:"distance_km"`
	SharedInterests    []SharedInterestBrief `json:"shared_interests,omitempty"`
}

// AdminDiscoveryResponse wraps simulation results
type AdminDiscoveryResponse struct {
	Results    []AdminDiscoveryResultItem `json:"results"`
	TotalCount int                        `json:"total_count"`
	ViewerLat  float64                    `json:"viewer_lat"`
	ViewerLng  float64                    `json:"viewer_lng"`
	RadiusKm   float64                    `json:"radius_km"`
}

// SimulateDiscovery runs the discovery algorithm from a viewer's perspective
// and enriches results with exact coordinates for map display
func (s *AdminDiscoveryService) SimulateDiscovery(ctx context.Context, req AdminDiscoveryRequest) (*AdminDiscoveryResponse, error) {
	if req.ViewerID == "" {
		return nil, fmt.Errorf("viewer_id is required")
	}

	// Get viewer's location — check user_profile first, then seeder's profile table
	viewerLat, viewerLng, err := s.getUserLocation(ctx, req.ViewerID)
	if err != nil {
		return nil, fmt.Errorf("viewer has no location: %w", err)
	}

	// Build filter for DiscoverPeople
	filter := PeopleDiscoveryFilter{
		CenterLat:           &viewerLat,
		CenterLng:           &viewerLng,
		RadiusKm:            req.RadiusKm,
		MinCompatibility:    req.MinCompatibility,
		RequireSharedAnswer: req.RequireSharedAnswer,
		Limit:               req.Limit,
	}

	if filter.RadiusKm <= 0 {
		filter.RadiusKm = DefaultSearchRadiusKm
	}
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	// Run the actual discovery algorithm
	discoveryResp, err := s.discoveryService.DiscoverPeople(ctx, req.ViewerID, filter)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	// Enrich results with exact coordinates and user info
	results := make([]AdminDiscoveryResultItem, 0, len(discoveryResp.Results))

	for _, dr := range discoveryResp.Results {
		item := AdminDiscoveryResultItem{
			UserID:             dr.UserID,
			CompatibilityScore: dr.CompatibilityScore,
			MatchScore:         dr.MatchScore,
			SharedInterests:    dr.SharedInterests,
		}

		// Get exact coordinates from profile (checks both tables)
		if lat, lng, err := s.getUserLocation(ctx, dr.UserID); err == nil {
			item.Lat = lat
			item.Lng = lng
		}
		// Get city from whichever profile table has it
		item.City = s.getUserCity(ctx, dr.UserID)

		// Calculate exact distance
		item.DistanceKm = s.geoService.HaversineDistance(viewerLat, viewerLng, item.Lat, item.Lng)

		// Get user email/username
		userQuery := `SELECT email, username, firstname FROM user WHERE id = type::record($user_id)`
		userResults, err := s.db.Query(ctx, userQuery, map[string]interface{}{
			"user_id": dr.UserID,
		})
		if err == nil {
			userRows := extractResultArray(userResults)
			if len(userRows) > 0 {
				item.Email = getStringField(userRows[0], "email")
				item.Username = getOptStringField(userRows[0], "username")
				item.Firstname = getOptStringField(userRows[0], "firstname")
			}
		}

		results = append(results, item)
	}

	return &AdminDiscoveryResponse{
		Results:    results,
		TotalCount: discoveryResp.TotalCount,
		ViewerLat:  viewerLat,
		ViewerLng:  viewerLng,
		RadiusKm:   filter.RadiusKm,
	}, nil
}

// AdminCompatibilityResponse combines breakdown + yikes summary
type AdminCompatibilityResponse struct {
	Breakdown *model.CompatibilityBreakdown `json:"breakdown"`
	Yikes     *model.YikesSummary           `json:"yikes"`
}

// getUserLocation resolves a user's lat/lng from user_profile or seeder's profile table
func (s *AdminDiscoveryService) getUserLocation(ctx context.Context, userID string) (float64, float64, error) {
	// Try user_profile first (app-created)
	q1 := `SELECT location FROM user_profile WHERE user = type::record($user_id) LIMIT 1`
	if lat, lng, ok := s.extractLocation(ctx, q1, userID); ok {
		return lat, lng, nil
	}

	// Try seeder's profile table
	q2 := `SELECT location FROM profile WHERE user_id = type::record($user_id) LIMIT 1`
	if lat, lng, ok := s.extractLocation(ctx, q2, userID); ok {
		return lat, lng, nil
	}

	return 0, 0, fmt.Errorf("no location found for user %s", userID)
}

// extractLocation runs a query and extracts lat/lng from the location field
func (s *AdminDiscoveryService) extractLocation(ctx context.Context, query, userID string) (float64, float64, bool) {
	results, err := s.db.Query(ctx, query, map[string]interface{}{"user_id": userID})
	if err != nil {
		return 0, 0, false
	}
	rows := extractResultArray(results)
	if len(rows) == 0 {
		return 0, 0, false
	}
	loc, ok := rows[0]["location"].(map[string]interface{})
	if !ok {
		return 0, 0, false
	}
	lat, latOk := loc["lat"].(float64)
	lng, lngOk := loc["lng"].(float64)
	if !latOk || !lngOk {
		return 0, 0, false
	}
	return lat, lng, true
}

// getUserCity resolves a user's city from their profile
func (s *AdminDiscoveryService) getUserCity(ctx context.Context, userID string) string {
	for _, query := range []string{
		`SELECT location.city as city FROM user_profile WHERE user = type::record($user_id) LIMIT 1`,
		`SELECT location.city as city FROM profile WHERE user_id = type::record($user_id) LIMIT 1`,
	} {
		results, err := s.db.Query(ctx, query, map[string]interface{}{"user_id": userID})
		if err != nil {
			continue
		}
		rows := extractResultArray(results)
		if len(rows) > 0 {
			if city := getStringField(rows[0], "city"); city != "" {
				return city
			}
		}
	}
	return ""
}

// GetCompatibility returns a detailed compatibility breakdown between two users
func (s *AdminDiscoveryService) GetCompatibility(ctx context.Context, userAID, userBID string) (*AdminCompatibilityResponse, error) {
	breakdown, err := s.compatibilityService.CalculateCompatibilityBreakdown(ctx, userAID, userBID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate compatibility: %w", err)
	}

	yikes, err := s.compatibilityService.CalculateYikesSummary(ctx, userAID, userBID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate yikes: %w", err)
	}

	return &AdminCompatibilityResponse{
		Breakdown: breakdown,
		Yikes:     yikes,
	}, nil
}
