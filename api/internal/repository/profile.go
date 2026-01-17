package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// ProfileRepository handles user profile data access
type ProfileRepository struct {
	db database.Database
}

// NewProfileRepository creates a new profile repository
func NewProfileRepository(db database.Database) *ProfileRepository {
	return &ProfileRepository{db: db}
}

// Create creates a new user profile
func (r *ProfileRepository) Create(ctx context.Context, profile *model.UserProfile) error {
	// Build query dynamically to avoid NULL vs NONE issues for optional fields
	setClause := `user = type::record($user_id), visibility = $visibility, created_on = time::now(), updated_on = time::now()`
	vars := map[string]interface{}{
		"user_id":    profile.UserID,
		"visibility": profile.Visibility,
	}

	// Add discovery eligibility fields
	setClause += ", discovery_eligible = $discovery_eligible"
	vars["discovery_eligible"] = profile.DiscoveryEligible

	// Add optional fields only when they have values
	if profile.Bio != nil {
		setClause += ", bio = $bio"
		vars["bio"] = *profile.Bio
	}
	if profile.Tagline != nil {
		setClause += ", tagline = $tagline"
		vars["tagline"] = *profile.Tagline
	}
	if len(profile.Languages) > 0 {
		setClause += ", languages = $languages"
		vars["languages"] = profile.Languages
	}
	if profile.Timezone != nil {
		setClause += ", timezone = $timezone"
		vars["timezone"] = *profile.Timezone
	}
	if profile.Location != nil {
		setClause += ", location = $location"
		vars["location"] = map[string]interface{}{
			"lat":          profile.Location.Lat,
			"lng":          profile.Location.Lng,
			"city":         profile.Location.City,
			"neighborhood": profile.Location.Neighborhood,
			"country":      profile.Location.Country,
			"country_code": profile.Location.CountryCode,
		}
	}

	query := "CREATE user_profile SET " + setClause
	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	// Extract created profile ID
	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	profile.ID = created.ID
	profile.CreatedOn = created.CreatedOn
	profile.UpdatedOn = created.UpdatedOn
	return nil
}

// GetByUserID retrieves a profile by user ID
func (r *ProfileRepository) GetByUserID(ctx context.Context, userID string) (*model.UserProfile, error) {
	query := `SELECT * FROM user_profile WHERE user = type::record($user_id) LIMIT 1`
	vars := map[string]interface{}{"user_id": userID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.parseProfileResult(result)
}

// Update updates a user profile
func (r *ProfileRepository) Update(ctx context.Context, userID string, updates map[string]interface{}) (*model.UserProfile, error) {
	// Build dynamic update query
	query := `UPDATE user_profile SET updated_on = time::now()`

	vars := map[string]interface{}{
		"user_id": userID,
	}

	if bio, ok := updates["bio"]; ok {
		query += ", bio = $bio"
		vars["bio"] = bio
	}
	if tagline, ok := updates["tagline"]; ok {
		query += ", tagline = $tagline"
		vars["tagline"] = tagline
	}
	if languages, ok := updates["languages"]; ok {
		query += ", languages = $languages"
		vars["languages"] = languages
	}
	if timezone, ok := updates["timezone"]; ok {
		query += ", timezone = $timezone"
		vars["timezone"] = timezone
	}
	if location, ok := updates["location"]; ok {
		query += ", location = $location"
		vars["location"] = location
	}
	if visibility, ok := updates["visibility"]; ok {
		query += ", visibility = $visibility"
		vars["visibility"] = visibility
	}
	if discoveryEligible, ok := updates["discovery_eligible"]; ok {
		query += ", discovery_eligible = $discovery_eligible"
		vars["discovery_eligible"] = discoveryEligible
	}

	query += ` WHERE user = type::record($user_id) RETURN AFTER`

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseProfileResult(result)
}

// UpdateLastActive updates the last active timestamp
func (r *ProfileRepository) UpdateLastActive(ctx context.Context, userID string) error {
	query := `UPDATE user_profile SET last_active = time::now(), updated_on = time::now() WHERE user = type::record($user_id)`
	vars := map[string]interface{}{"user_id": userID}

	return r.db.Execute(ctx, query, vars)
}

// Delete deletes a user profile
func (r *ProfileRepository) Delete(ctx context.Context, userID string) error {
	query := `DELETE user_profile WHERE user = type::record($user_id)`
	vars := map[string]interface{}{"user_id": userID}

	return r.db.Execute(ctx, query, vars)
}

// GetNearby finds profiles within a bounding box (for initial filtering)
func (r *ProfileRepository) GetNearby(ctx context.Context, minLat, maxLat, minLng, maxLng float64, limit int) ([]*model.UserProfile, error) {
	query := `
		SELECT * FROM user_profile
		WHERE location != NONE
			AND location.lat >= $min_lat
			AND location.lat <= $max_lat
			AND location.lng >= $min_lng
			AND location.lng <= $max_lng
			AND visibility != "private"
		ORDER BY last_active DESC
		LIMIT $limit
	`

	vars := map[string]interface{}{
		"min_lat": minLat,
		"max_lat": maxLat,
		"min_lng": minLng,
		"max_lng": maxLng,
		"limit":   limit,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseProfilesResult(result)
}

// GetByVisibility finds profiles with a specific visibility setting
func (r *ProfileRepository) GetByVisibility(ctx context.Context, visibility string, limit int) ([]*model.UserProfile, error) {
	query := `
		SELECT * FROM user_profile
		WHERE visibility = $visibility
		ORDER BY last_active DESC
		LIMIT $limit
	`

	vars := map[string]interface{}{
		"visibility": visibility,
		"limit":      limit,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseProfilesResult(result)
}

// Helper functions

func (r *ProfileRepository) parseProfileResult(result interface{}) (*model.UserProfile, error) {
	if result == nil {
		return nil, database.ErrNotFound
	}

	// Navigate through SurrealDB response structure
	if resp, ok := result.(map[string]interface{}); ok {
		if status, ok := resp["status"].(string); ok && status == "OK" {
			if resultData, ok := resp["result"].([]interface{}); ok {
				if len(resultData) == 0 {
					return nil, database.ErrNotFound
				}
				result = resultData[0]
			}
		}
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

	// Handle SurrealDB's complex ID format
	if id, ok := data["id"]; ok {
		data["id"] = convertSurrealID(id)
	}
	if userID, ok := data["user"]; ok {
		data["user_id"] = convertSurrealID(userID)
		delete(data, "user")
	}

	// Convert to JSON and back to struct
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var profile model.UserProfile
	if err := json.Unmarshal(jsonBytes, &profile); err != nil {
		return nil, err
	}

	// Parse location manually to handle internal/public split
	if locData, ok := data["location"].(map[string]interface{}); ok {
		profile.Location = &model.Location{
			City:        getString(locData, "city"),
			Country:     getString(locData, "country"),
			CountryCode: getString(locData, "country_code"),
		}
		if neighborhood, ok := locData["neighborhood"].(string); ok {
			profile.Location.Neighborhood = &neighborhood
		}
		// Note: lat/lng stored but not exposed in Location struct (json:"-")
		if lat, ok := locData["lat"].(float64); ok {
			profile.Location.Lat = lat
		}
		if lng, ok := locData["lng"].(float64); ok {
			profile.Location.Lng = lng
		}
	}

	return &profile, nil
}

func (r *ProfileRepository) parseProfilesResult(result []interface{}) ([]*model.UserProfile, error) {
	profiles := make([]*model.UserProfile, 0)

	for _, res := range result {
		// Handle SurrealDB response wrapper
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					profile, err := r.parseProfileResult(item)
					if err != nil {
						continue
					}
					profiles = append(profiles, profile)
				}
				continue
			}
		}

		profile, err := r.parseProfileResult(res)
		if err != nil {
			continue
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// GetLocationInternal returns the internal location data (with coordinates) for calculations
func (r *ProfileRepository) GetLocationInternal(ctx context.Context, userID string) (*model.LocationInternal, error) {
	query := `SELECT location FROM user_profile WHERE user = type::record($user_id) LIMIT 1`
	vars := map[string]interface{}{"user_id": userID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// Parse location data
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	locData, ok := data["location"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	location := &model.LocationInternal{
		Lat:         getFloat(locData, "lat"),
		Lng:         getFloat(locData, "lng"),
		City:        getString(locData, "city"),
		Country:     getString(locData, "country"),
		CountryCode: getString(locData, "country_code"),
	}

	if neighborhood, ok := locData["neighborhood"].(string); ok {
		location.Neighborhood = &neighborhood
	}

	return location, nil
}

// Helpers getString, getFloat, getTime, getBool, getStringSlice, getInt are defined in helpers.go
