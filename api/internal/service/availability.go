package service

import (
	"context"
	"errors"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// Availability service errors
var (
	ErrAvailabilityNotFound   = errors.New("availability not found")
	ErrHangoutRequestNotFound = errors.New("hangout request not found")
	ErrHangoutNotFound        = errors.New("hangout not found")
	ErrInvalidHangoutType     = errors.New("invalid hangout type")
	ErrInvalidTimeRange       = errors.New("end time must be after start time")
	ErrNoteTooShort           = errors.New("note must be at least 20 characters")
	ErrAlreadyRequested       = errors.New("already requested this hangout")
	ErrCannotRequestOwn       = errors.New("cannot request your own availability")
)

// AvailabilityRepository defines the interface for availability storage
type AvailabilityRepository interface {
	Create(ctx context.Context, av *model.Availability) error
	GetByID(ctx context.Context, id string) (*model.Availability, error)
	GetByUser(ctx context.Context, userID string) ([]*model.Availability, error)
	GetNearby(ctx context.Context, minLat, maxLat, minLng, maxLng float64, startTime, endTime time.Time, excludeUserID string, limit int) ([]*model.Availability, error)
	GetByHangoutType(ctx context.Context, hangoutType string, excludeUserID string, limit int) ([]*model.Availability, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) (*model.Availability, error)
	Delete(ctx context.Context, id string) error
	CreateHangoutRequest(ctx context.Context, req *model.HangoutRequest) error
	GetHangoutRequest(ctx context.Context, id string) (*model.HangoutRequest, error)
	GetPendingRequests(ctx context.Context, availabilityID string) ([]*model.HangoutRequest, error)
	UpdateHangoutRequestStatus(ctx context.Context, id, status string) error
	CreateHangout(ctx context.Context, hangout *model.Hangout) error
	GetHangout(ctx context.Context, id string) (*model.Hangout, error)
	GetUserHangouts(ctx context.Context, userID string, limit int) ([]*model.Hangout, error)
	UpdateHangoutStatus(ctx context.Context, id, status string) error
	// Nudge-related
	GetStaleHangouts(ctx context.Context, cutoff time.Time, status string) ([]*model.Hangout, error)
	GetUpcomingHangouts(ctx context.Context, windowStart, windowEnd time.Time) ([]*model.Hangout, error)
	GetAllPendingRequests(ctx context.Context) ([]*model.HangoutRequest, error)
	GetPendingRequestsForUser(ctx context.Context, userID string) ([]*model.HangoutRequest, error)
	GetUserUpcomingHangouts(ctx context.Context, userID string, windowStart, windowEnd time.Time) ([]*model.Hangout, error)
}

// AvailabilityService handles availability business logic
type AvailabilityService struct {
	repo       AvailabilityRepository
	geoService *GeoService
}

// AvailabilityServiceConfig holds configuration for the availability service
type AvailabilityServiceConfig struct {
	Repo AvailabilityRepository
}

// NewAvailabilityService creates a new availability service
func NewAvailabilityService(cfg AvailabilityServiceConfig) *AvailabilityService {
	return &AvailabilityService{
		repo:       cfg.Repo,
		geoService: NewGeoService(),
	}
}

// CreateAvailability creates a new availability window
func (s *AvailabilityService) CreateAvailability(ctx context.Context, userID string, req *model.CreateAvailabilityRequest) (*model.Availability, error) {
	// Validate hangout type
	if !isValidHangoutType(req.HangoutType) {
		return nil, ErrInvalidHangoutType
	}

	// Parse times
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, errors.New("invalid start_time format")
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return nil, errors.New("invalid end_time format")
	}

	// Validate time range
	if endTime.Before(startTime) || endTime.Equal(startTime) {
		return nil, ErrInvalidTimeRange
	}

	// Default expiry to 24 hours from now
	expiresAt := time.Now().Add(24 * time.Hour)

	// Default max people
	maxPeople := 1
	if req.MaxPeople != nil && *req.MaxPeople > 0 {
		maxPeople = *req.MaxPeople
	}

	// Default visibility
	visibility := "circles"
	if req.Visibility != nil && *req.Visibility != "" {
		visibility = *req.Visibility
	}

	av := &model.Availability{
		UserID:              userID,
		Status:              model.AvailabilityStatusAvailable,
		StartTime:           startTime,
		EndTime:             endTime,
		HangoutType:         model.HangoutType(req.HangoutType),
		ActivityDescription: req.ActivityDescription,
		ActivityVenue:       req.ActivityVenue,
		InterestID:          req.InterestID,
		MaxPeople:           maxPeople,
		Note:                req.Note,
		Visibility:          visibility,
		ExpiresAt:           expiresAt,
	}

	if req.Location != nil {
		av.Location = &model.AvailabilityLocation{
			Lat:    req.Location.Lat,
			Lng:    req.Location.Lng,
			Radius: req.Location.Radius,
		}
	}

	if err := s.repo.Create(ctx, av); err != nil {
		return nil, err
	}

	return av, nil
}

// GetAvailability retrieves an availability by ID
func (s *AvailabilityService) GetAvailability(ctx context.Context, id string) (*model.Availability, error) {
	av, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if av == nil {
		return nil, ErrAvailabilityNotFound
	}
	return av, nil
}

// GetUserAvailabilities retrieves all active availabilities for a user
func (s *AvailabilityService) GetUserAvailabilities(ctx context.Context, userID string) ([]*model.Availability, error) {
	return s.repo.GetByUser(ctx, userID)
}

// FindNearbyAvailabilities finds availabilities near a location
func (s *AvailabilityService) FindNearbyAvailabilities(ctx context.Context, userID string, lat, lng, radiusKm float64, startTime, endTime time.Time, limit int) ([]*model.Availability, error) {
	bbox := s.geoService.GetBoundingBox(lat, lng, radiusKm)

	if limit <= 0 || limit > 50 {
		limit = 20
	}

	return s.repo.GetNearby(ctx, bbox.MinLat, bbox.MaxLat, bbox.MinLng, bbox.MaxLng, startTime, endTime, userID, limit)
}

// FindByHangoutType finds availabilities by type
func (s *AvailabilityService) FindByHangoutType(ctx context.Context, userID string, hangoutType string, limit int) ([]*model.Availability, error) {
	if !isValidHangoutType(hangoutType) {
		return nil, ErrInvalidHangoutType
	}

	if limit <= 0 || limit > 50 {
		limit = 20
	}

	return s.repo.GetByHangoutType(ctx, hangoutType, userID, limit)
}

// UpdateAvailability updates an availability
func (s *AvailabilityService) UpdateAvailability(ctx context.Context, userID, id string, req *model.UpdateAvailabilityRequest) (*model.Availability, error) {
	// Verify ownership
	av, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if av == nil {
		return nil, ErrAvailabilityNotFound
	}
	if av.UserID != userID {
		return nil, ErrAvailabilityNotFound // Don't reveal it exists
	}

	updates := make(map[string]interface{})

	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.EndTime != nil {
		endTime, err := time.Parse(time.RFC3339, *req.EndTime)
		if err != nil {
			return nil, errors.New("invalid end_time format")
		}
		updates["end_time"] = endTime
	}
	if req.ActivityDescription != nil {
		updates["activity_description"] = *req.ActivityDescription
	}
	if req.ActivityVenue != nil {
		updates["activity_venue"] = *req.ActivityVenue
	}
	if req.MaxPeople != nil {
		updates["max_people"] = *req.MaxPeople
	}
	if req.Note != nil {
		updates["note"] = *req.Note
	}

	if len(updates) == 0 {
		return av, nil
	}

	return s.repo.Update(ctx, id, updates)
}

// DeleteAvailability deletes an availability
func (s *AvailabilityService) DeleteAvailability(ctx context.Context, userID, id string) error {
	// Verify ownership
	av, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if av == nil {
		return ErrAvailabilityNotFound
	}
	if av.UserID != userID {
		return ErrAvailabilityNotFound
	}

	return s.repo.Delete(ctx, id)
}

// RequestHangout creates a request to join someone's availability
func (s *AvailabilityService) RequestHangout(ctx context.Context, requesterID, availabilityID, note string) (*model.HangoutRequest, error) {
	// Get the availability
	av, err := s.repo.GetByID(ctx, availabilityID)
	if err != nil {
		return nil, err
	}
	if av == nil {
		return nil, ErrAvailabilityNotFound
	}

	// Can't request your own
	if av.UserID == requesterID {
		return nil, ErrCannotRequestOwn
	}

	// Validate note length (minimum 20 chars to prevent shallow interactions)
	if len(note) < model.MinHangoutNoteLength {
		return nil, ErrNoteTooShort
	}

	req := &model.HangoutRequest{
		AvailabilityID: availabilityID,
		RequesterID:    requesterID,
		Note:           note,
		Status:         model.HangoutRequestStatusPending,
	}

	if err := s.repo.CreateHangoutRequest(ctx, req); err != nil {
		return nil, err
	}

	return req, nil
}

// GetPendingRequests retrieves pending requests for an availability
func (s *AvailabilityService) GetPendingRequests(ctx context.Context, userID, availabilityID string) ([]*model.HangoutRequest, error) {
	// Verify ownership
	av, err := s.repo.GetByID(ctx, availabilityID)
	if err != nil {
		return nil, err
	}
	if av == nil {
		return nil, ErrAvailabilityNotFound
	}
	if av.UserID != userID {
		return nil, ErrAvailabilityNotFound
	}

	return s.repo.GetPendingRequests(ctx, availabilityID)
}

// RespondToRequest accepts or declines a hangout request
func (s *AvailabilityService) RespondToRequest(ctx context.Context, userID, requestID string, accept bool) (*model.Hangout, error) {
	// Get the request
	req, err := s.repo.GetHangoutRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, ErrHangoutRequestNotFound
	}

	// Get the availability and verify ownership
	av, err := s.repo.GetByID(ctx, req.AvailabilityID)
	if err != nil {
		return nil, err
	}
	if av == nil || av.UserID != userID {
		return nil, ErrAvailabilityNotFound
	}

	if accept {
		// Update request status
		if err := s.repo.UpdateHangoutRequestStatus(ctx, requestID, model.HangoutRequestStatusAccepted); err != nil {
			return nil, err
		}

		// Create the hangout
		hangout := &model.Hangout{
			Participants:        []string{av.UserID, req.RequesterID},
			AvailabilityID:      &av.ID,
			HangoutType:         av.HangoutType,
			ActivityDescription: av.ActivityDescription,
			ScheduledTime:       av.StartTime,
			IsSupportSession:    av.HangoutType == model.HangoutTypeTalkItOut || av.HangoutType == model.HangoutTypeHereToListen,
			Status:              model.HangoutStatusScheduled,
		}

		if av.Location != nil {
			hangout.Location = &model.HangoutLocation{
				Lat: av.Location.Lat,
				Lng: av.Location.Lng,
			}
			if av.ActivityVenue != nil {
				hangout.Location.Venue = av.ActivityVenue
			}
		}

		if err := s.repo.CreateHangout(ctx, hangout); err != nil {
			return nil, err
		}

		return hangout, nil
	}

	// Declined
	if err := s.repo.UpdateHangoutRequestStatus(ctx, requestID, model.HangoutRequestStatusDeclined); err != nil {
		return nil, err
	}

	return nil, nil
}

// GetUserHangouts retrieves hangouts for a user
func (s *AvailabilityService) GetUserHangouts(ctx context.Context, userID string, limit int) ([]*model.Hangout, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.repo.GetUserHangouts(ctx, userID, limit)
}

// UpdateHangoutStatus updates a hangout status (complete, cancel, etc.)
func (s *AvailabilityService) UpdateHangoutStatus(ctx context.Context, userID, hangoutID, status string) error {
	// Get the hangout and verify participation
	hangout, err := s.repo.GetHangout(ctx, hangoutID)
	if err != nil {
		return err
	}
	if hangout == nil {
		return ErrHangoutNotFound
	}

	// Check if user is a participant
	isParticipant := false
	for _, p := range hangout.Participants {
		if p == userID {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return ErrHangoutNotFound
	}

	return s.repo.UpdateHangoutStatus(ctx, hangoutID, status)
}

// Helper functions

func isValidHangoutType(t string) bool {
	switch model.HangoutType(t) {
	case model.HangoutTypeTalkItOut,
		model.HangoutTypeHereToListen,
		model.HangoutTypeConcreteActivity,
		model.HangoutTypeMutualInterest,
		model.HangoutTypeMeetAnyone:
		return true
	default:
		return false
	}
}
