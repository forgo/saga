package service

import (
	"context"
	"errors"

	"github.com/forgo/saga/api/internal/model"
)

// Profile service errors
var (
	ErrProfileNotFound    = errors.New("profile not found")
	ErrProfileExists      = errors.New("profile already exists")
	ErrInvalidVisibility  = errors.New("invalid visibility setting")
	ErrBioTooLong         = errors.New("bio exceeds maximum length")
	ErrTaglineTooLong     = errors.New("tagline exceeds maximum length")
	ErrTooManyLanguages   = errors.New("too many languages")
)

// ProfileRepository defines the interface for profile storage
type ProfileRepository interface {
	Create(ctx context.Context, profile *model.UserProfile) error
	GetByUserID(ctx context.Context, userID string) (*model.UserProfile, error)
	Update(ctx context.Context, userID string, updates map[string]interface{}) (*model.UserProfile, error)
	UpdateLastActive(ctx context.Context, userID string) error
	Delete(ctx context.Context, userID string) error
	GetNearby(ctx context.Context, minLat, maxLat, minLng, maxLng float64, limit int) ([]*model.UserProfile, error)
	GetLocationInternal(ctx context.Context, userID string) (*model.LocationInternal, error)
}

// ProfileModerationRepository defines moderation checks for profiles
type ProfileModerationRepository interface {
	IsBlockedEitherWay(ctx context.Context, userID1, userID2 string) (bool, error)
}

// ProfileGuildRepository defines guild membership checks for profiles
type ProfileGuildRepository interface {
	GetGuildsForUser(ctx context.Context, userID string) ([]*model.Guild, error)
}

// ProfileService handles profile business logic
type ProfileService struct {
	profileRepo    ProfileRepository
	userRepo       UserRepository
	moderationRepo ProfileModerationRepository
	guildRepo      ProfileGuildRepository
	geoService     *GeoService
}

// ProfileServiceConfig holds configuration for the profile service
type ProfileServiceConfig struct {
	ProfileRepo    ProfileRepository
	UserRepo       UserRepository
	ModerationRepo ProfileModerationRepository
	GuildRepo      ProfileGuildRepository
}

// NewProfileService creates a new profile service
func NewProfileService(cfg ProfileServiceConfig) *ProfileService {
	return &ProfileService{
		profileRepo:    cfg.ProfileRepo,
		userRepo:       cfg.UserRepo,
		moderationRepo: cfg.ModerationRepo,
		guildRepo:      cfg.GuildRepo,
		geoService:     NewGeoService(),
	}
}

// isBlocked checks if two users have blocked each other
func (s *ProfileService) isBlocked(ctx context.Context, userID1, userID2 string) bool {
	if s.moderationRepo == nil {
		return false
	}
	blocked, err := s.moderationRepo.IsBlockedEitherWay(ctx, userID1, userID2)
	if err != nil {
		return false // Fail open to avoid breaking profile viewing on errors
	}
	return blocked
}

// sharesGuild checks if two users share at least one guild
func (s *ProfileService) sharesGuild(ctx context.Context, userID1, userID2 string) bool {
	if s.guildRepo == nil {
		return false
	}

	guilds1, err := s.guildRepo.GetGuildsForUser(ctx, userID1)
	if err != nil || len(guilds1) == 0 {
		return false
	}

	guilds2, err := s.guildRepo.GetGuildsForUser(ctx, userID2)
	if err != nil || len(guilds2) == 0 {
		return false
	}

	// Check for overlap
	guild1Set := make(map[string]bool)
	for _, g := range guilds1 {
		guild1Set[g.ID] = true
	}
	for _, g := range guilds2 {
		if guild1Set[g.ID] {
			return true
		}
	}
	return false
}

// GetProfile retrieves a user's own profile
func (s *ProfileService) GetProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	profile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrProfileNotFound
	}
	return profile, nil
}

// GetOrCreateProfile retrieves or creates a profile for a user
func (s *ProfileService) GetOrCreateProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	profile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if profile != nil {
		return profile, nil
	}

	// Create default profile
	profile = &model.UserProfile{
		UserID:     userID,
		Visibility: model.VisibilityCircles,
		Languages:  []string{},
	}

	if err := s.profileRepo.Create(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

// UpdateProfile updates a user's profile
func (s *ProfileService) UpdateProfile(ctx context.Context, userID string, req *model.UpdateProfileRequest) (*model.UserProfile, error) {
	// Validate fields
	if req.Bio != nil && len(*req.Bio) > model.MaxBioLength {
		return nil, ErrBioTooLong
	}
	if req.Tagline != nil && len(*req.Tagline) > model.MaxTaglineLength {
		return nil, ErrTaglineTooLong
	}
	if len(req.Languages) > model.MaxLanguages {
		return nil, ErrTooManyLanguages
	}
	if req.Visibility != nil && !isValidVisibility(*req.Visibility) {
		return nil, ErrInvalidVisibility
	}

	// Ensure profile exists
	_, err := s.GetOrCreateProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Bio != nil {
		updates["bio"] = *req.Bio
	}
	if req.Tagline != nil {
		updates["tagline"] = *req.Tagline
	}
	if len(req.Languages) > 0 {
		updates["languages"] = req.Languages
	}
	if req.Timezone != nil {
		updates["timezone"] = *req.Timezone
	}
	if req.Location != nil {
		updates["location"] = map[string]interface{}{
			"lat":          req.Location.Lat,
			"lng":          req.Location.Lng,
			"city":         req.Location.City,
			"neighborhood": req.Location.Neighborhood,
			"country":      req.Location.Country,
			"country_code": req.Location.CountryCode,
		}
	}
	if req.Visibility != nil {
		updates["visibility"] = *req.Visibility
	}

	return s.profileRepo.Update(ctx, userID, updates)
}

// GetPublicProfile retrieves another user's public profile with privacy controls
func (s *ProfileService) GetPublicProfile(ctx context.Context, viewerID, targetUserID string, viewerLocation *model.LocationInternal) (*model.PublicProfile, error) {
	// SECURITY: Check if users have blocked each other
	if s.isBlocked(ctx, viewerID, targetUserID) {
		return nil, ErrProfileNotFound
	}

	// Get target profile
	profile, err := s.profileRepo.GetByUserID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrProfileNotFound
	}

	// Check visibility
	if profile.Visibility == model.VisibilityPrivate {
		return nil, ErrProfileNotFound
	}

	// SECURITY: Check guild membership if visibility is "guilds" (formerly "circles")
	if profile.Visibility == model.VisibilityCircles {
		if !s.sharesGuild(ctx, viewerID, targetUserID) {
			return nil, ErrProfileNotFound
		}
	}

	// Get user details
	user, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	// Build public profile
	public := &model.PublicProfile{
		UserID:    targetUserID,
		Firstname: user.Firstname,
		Bio:       profile.Bio,
		Tagline:   profile.Tagline,
		Languages: profile.Languages,
	}

	// Add city/country (never exact location)
	if profile.Location != nil {
		public.City = profile.Location.City
		public.Country = profile.Location.Country
	}

	// Calculate distance bucket if both have locations
	if viewerLocation != nil && profile.Location != nil {
		targetLocation := &model.LocationInternal{
			Lat: profile.Location.Lat,
			Lng: profile.Location.Lng,
		}
		distance := s.geoService.DistanceBetweenLocations(viewerLocation, targetLocation)
		if distance >= 0 {
			public.Distance = s.geoService.GetDistanceBucket(distance)
		}
	}

	// Add activity status
	public.ActivityStatus = model.GetActivityStatus(profile.LastActive)

	return public, nil
}

// UpdateLastActive updates the user's last active timestamp
func (s *ProfileService) UpdateLastActive(ctx context.Context, userID string) error {
	return s.profileRepo.UpdateLastActive(ctx, userID)
}

// GetNearbyProfiles finds profiles near a location
func (s *ProfileService) GetNearbyProfiles(ctx context.Context, viewerID string, centerLat, centerLng, radiusKm float64, limit int) ([]*model.PublicProfile, error) {
	// Get bounding box for initial DB filter
	bbox := s.geoService.GetBoundingBox(centerLat, centerLng, radiusKm)

	// Query profiles in bounding box
	profiles, err := s.profileRepo.GetNearby(ctx, bbox.MinLat, bbox.MaxLat, bbox.MinLng, bbox.MaxLng, limit*2)
	if err != nil {
		return nil, err
	}

	viewerLocation := &model.LocationInternal{Lat: centerLat, Lng: centerLng}
	result := make([]*model.PublicProfile, 0, len(profiles))

	for _, profile := range profiles {
		// Skip self
		if profile.UserID == viewerID {
			continue
		}

		// SECURITY: Skip blocked users (also checked in GetPublicProfile, but skip early)
		if s.isBlocked(ctx, viewerID, profile.UserID) {
			continue
		}

		// Get public profile with distance
		public, err := s.GetPublicProfile(ctx, viewerID, profile.UserID, viewerLocation)
		if err != nil {
			continue
		}

		result = append(result, public)

		if len(result) >= limit {
			break
		}
	}

	return result, nil
}

// GetLocationInternal gets a user's internal location (for distance calculations)
func (s *ProfileService) GetLocationInternal(ctx context.Context, userID string) (*model.LocationInternal, error) {
	return s.profileRepo.GetLocationInternal(ctx, userID)
}

// Helper functions

func isValidVisibility(v string) bool {
	return v == model.VisibilityCircles ||
		v == model.VisibilityPublic ||
		v == model.VisibilityPrivate
}
