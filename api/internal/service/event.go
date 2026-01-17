package service

import (
	"context"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// Error definitions moved to errors.go

// EventRepositoryInterface defines the repository interface
type EventRepositoryInterface interface {
	Create(ctx context.Context, event *model.Event) error
	Get(ctx context.Context, eventID string) (*model.Event, error)
	Update(ctx context.Context, eventID string, updates map[string]interface{}) (*model.Event, error)
	Delete(ctx context.Context, eventID string) error
	GetByGuild(ctx context.Context, guildID string, filters *model.EventSearchFilters) ([]*model.Event, error)
	GetPublicEvents(ctx context.Context, filters *model.EventSearchFilters, limit int) ([]*model.Event, error)
	CreateHost(ctx context.Context, host *model.EventHost) error
	GetHosts(ctx context.Context, eventID string) ([]*model.EventHost, error)
	IsHost(ctx context.Context, eventID, userID string) (bool, error)
	CreateRSVP(ctx context.Context, rsvp *model.EventRSVP) error
	GetRSVP(ctx context.Context, eventID, userID string) (*model.EventRSVP, error)
	UpdateRSVP(ctx context.Context, rsvpID string, updates map[string]interface{}) (*model.EventRSVP, error)
	GetRSVPsByEvent(ctx context.Context, eventID string) ([]*model.EventRSVP, error)
	GetPendingRSVPs(ctx context.Context, eventID string) ([]*model.EventRSVP, error)
	CountApprovedRSVPs(ctx context.Context, eventID string) (int, error)
}

// CompatibilityServiceForEvent is the compatibility service interface
type CompatibilityServiceForEvent interface {
	CalculateCompatibility(ctx context.Context, userAID, userBID string) (*model.CompatibilityScore, error)
	CalculateYikesSummary(ctx context.Context, userAID, userBID string) (*model.YikesSummary, error)
}

// QuestionnaireServiceForEvent is the questionnaire service interface
type QuestionnaireServiceForEvent interface {
	GetUserAnswers(ctx context.Context, userID string) ([]*model.Answer, error)
}

// EventRoleServiceForEvent is the event role service interface
type EventRoleServiceForEvent interface {
	CreateDefaultRole(ctx context.Context, eventID, hostUserID string, maxSlots int) (*model.EventRole, error)
}

// EventService handles event business logic
type EventService struct {
	repo                 EventRepositoryInterface
	compatibilityService CompatibilityServiceForEvent
	questionnaireService QuestionnaireServiceForEvent
	eventRoleService     EventRoleServiceForEvent
}

// NewEventService creates a new event service
func NewEventService(
	repo EventRepositoryInterface,
	compatibilityService CompatibilityServiceForEvent,
	questionnaireService QuestionnaireServiceForEvent,
	eventRoleService EventRoleServiceForEvent,
) *EventService {
	return &EventService{
		repo:                 repo,
		compatibilityService: compatibilityService,
		questionnaireService: questionnaireService,
		eventRoleService:     eventRoleService,
	}
}

// CreateEvent creates a new event
func (s *EventService) CreateEvent(ctx context.Context, userID string, req *model.CreateEventRequest) (*model.Event, error) {
	event := &model.Event{
		GuildID:            req.GuildID,
		Title:              req.Title,
		Description:        req.Description,
		Location:           req.Location,
		StartTime:          req.StartTime,
		EndTime:            req.EndTime,
		Template:           req.Template,
		Visibility:         req.Visibility,
		MaxAttendees:       req.MaxAttendees,
		WaitlistEnabled:    req.WaitlistEnabled,
		CoverImage:         req.CoverImage,
		ThemeColor:         req.ThemeColor,
		ValuesRequired:     req.ValuesRequired,
		ValuesQuestions:    req.ValuesQuestions,
		AutoApproveAligned: req.AutoApproveAligned,
		YikesThreshold:     req.YikesThreshold,
		IsSupportEvent:     req.IsSupportEvent,
		Status:             model.EventStatusPublished,
		CreatedBy:          userID,
	}

	// Set default yikes threshold
	if event.YikesThreshold == 0 && event.ValuesRequired {
		event.YikesThreshold = model.DefaultYikesThreshold
	}

	if err := s.repo.Create(ctx, event); err != nil {
		return nil, err
	}

	// Add creator as primary host
	host := &model.EventHost{
		EventID: event.ID,
		UserID:  userID,
		Role:    model.HostRolePrimary,
		AddedBy: userID,
	}
	// Create host (non-fatal error)
	_ = s.repo.CreateHost(ctx, host)

	// Create default "Guest" role
	maxSlots := 0
	if event.MaxAttendees != nil {
		maxSlots = *event.MaxAttendees
	}
	if s.eventRoleService != nil {
		_, _ = s.eventRoleService.CreateDefaultRole(ctx, event.ID, userID, maxSlots)
	}

	return event, nil
}

// GetEvent retrieves an event by ID
func (s *EventService) GetEvent(ctx context.Context, eventID string) (*model.Event, error) {
	event, err := s.repo.Get(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, ErrEventNotFound
	}
	return event, nil
}

// GetEventWithDetails retrieves an event with all details
func (s *EventService) GetEventWithDetails(ctx context.Context, eventID, userID string) (*model.EventWithDetails, error) {
	event, err := s.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	hosts, err := s.repo.GetHosts(ctx, eventID)
	if err != nil {
		return nil, err
	}

	approvedCount, _ := s.repo.CountApprovedRSVPs(ctx, eventID)
	pendingRSVPs, _ := s.repo.GetPendingRSVPs(ctx, eventID)

	details := &model.EventWithDetails{
		Event:          *event,
		Hosts:          make([]model.EventHost, 0, len(hosts)),
		AttendeesCount: approvedCount,
		WaitlistCount:  len(pendingRSVPs),
	}

	for _, host := range hosts {
		details.Hosts = append(details.Hosts, *host)
	}

	// Get user's RSVP if authenticated
	if userID != "" {
		rsvp, _ := s.repo.GetRSVP(ctx, eventID, userID)
		details.UserRSVP = rsvp
	}

	return details, nil
}

// UpdateEvent updates an event (host only)
func (s *EventService) UpdateEvent(ctx context.Context, userID, eventID string, req *model.UpdateEventRequest) (*model.Event, error) {
	isHost, err := s.repo.IsHost(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}
	if !isHost {
		return nil, ErrNotEventHost
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Location != nil {
		updates["location"] = map[string]interface{}{
			"name":         req.Location.Name,
			"address":      req.Location.Address,
			"neighborhood": req.Location.Neighborhood,
			"city":         req.Location.City,
			"lat":          req.Location.Lat,
			"lng":          req.Location.Lng,
			"is_virtual":   req.Location.IsVirtual,
			"meet_link":    req.Location.MeetLink,
		}
	}
	if req.StartTime != nil {
		updates["start_time"] = *req.StartTime
	}
	if req.EndTime != nil {
		updates["end_time"] = *req.EndTime
	}
	if req.MaxAttendees != nil {
		updates["max_attendees"] = *req.MaxAttendees
	}
	if req.WaitlistEnabled != nil {
		updates["waitlist_enabled"] = *req.WaitlistEnabled
	}
	if req.CoverImage != nil {
		updates["cover_image"] = *req.CoverImage
	}
	if req.ThemeColor != nil {
		updates["theme_color"] = *req.ThemeColor
	}
	if req.ValuesRequired != nil {
		updates["values_required"] = *req.ValuesRequired
	}
	if req.ValuesQuestions != nil {
		updates["values_questions"] = req.ValuesQuestions
	}
	if req.AutoApproveAligned != nil {
		updates["auto_approve_aligned"] = *req.AutoApproveAligned
	}
	if req.YikesThreshold != nil {
		updates["yikes_threshold"] = *req.YikesThreshold
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		return s.GetEvent(ctx, eventID)
	}

	return s.repo.Update(ctx, eventID, updates)
}

// CancelEvent cancels an event (host only)
func (s *EventService) CancelEvent(ctx context.Context, userID, eventID string) error {
	isHost, err := s.repo.IsHost(ctx, eventID, userID)
	if err != nil {
		return err
	}
	if !isHost {
		return ErrNotEventHost
	}

	_, err = s.repo.Update(ctx, eventID, map[string]interface{}{
		"status": model.EventStatusCancelled,
	})
	return err
}

// AddHost adds a co-host to an event
func (s *EventService) AddHost(ctx context.Context, userID, eventID, newHostID string) (*model.EventHost, error) {
	isHost, err := s.repo.IsHost(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}
	if !isHost {
		return nil, ErrNotEventHost
	}

	// Check max hosts
	hosts, err := s.repo.GetHosts(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if len(hosts) >= model.MaxEventHosts {
		return nil, ErrMaxHostsReached
	}

	// Check if already a host
	for _, h := range hosts {
		if h.UserID == newHostID {
			return nil, ErrAlreadyHost
		}
	}

	host := &model.EventHost{
		EventID: eventID,
		UserID:  newHostID,
		Role:    model.HostRoleCoHost,
		AddedBy: userID,
	}

	if err := s.repo.CreateHost(ctx, host); err != nil {
		return nil, err
	}

	return host, nil
}

// RSVP creates or updates an RSVP for an event
func (s *EventService) RSVP(ctx context.Context, userID, eventID string, req *model.RSVPRequest) (*model.EventRSVP, error) {
	event, err := s.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Check if user already has an RSVP
	existingRSVP, err := s.repo.GetRSVP(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}

	// Handle not going - simple case
	if req.RSVPType == model.RSVPTypeNotGoing {
		if existingRSVP != nil {
			return s.repo.UpdateRSVP(ctx, existingRSVP.ID, map[string]interface{}{
				"status":    model.RSVPStatusCancelled,
				"rsvp_type": model.RSVPTypeNotGoing,
			})
		}
		return nil, nil // No existing RSVP, nothing to cancel
	}

	// If already has approved RSVP, don't allow re-RSVP
	if existingRSVP != nil && existingRSVP.Status == model.RSVPStatusApproved {
		return nil, ErrAlreadyRSVPd
	}

	// Check capacity
	if event.MaxAttendees != nil {
		currentCount, _ := s.repo.CountApprovedRSVPs(ctx, eventID)
		totalRequested := 1 + req.PlusOnes
		if currentCount+totalRequested > *event.MaxAttendees {
			if !event.WaitlistEnabled {
				return nil, ErrEventFull
			}
		}
	}

	// Check values alignment
	var valuesCheck *model.EventValuesCheck
	if event.ValuesRequired {
		valuesCheck, err = s.CheckValuesAlignment(ctx, userID, event)
		if err != nil {
			return nil, err
		}
	}

	// Determine RSVP status based on values check and event settings
	status := model.RSVPStatusPending
	var waitingReason *string

	if event.ValuesRequired && valuesCheck != nil {
		switch valuesCheck.Recommendation {
		case model.ValuesRecommendAutoApprove:
			if event.AutoApproveAligned {
				status = model.RSVPStatusApproved
			}
		case model.ValuesRecommendNeedsReview:
			reason := model.WaitingReasonValuesReview
			waitingReason = &reason
		case model.ValuesRecommendDeclineSuggested:
			reason := model.WaitingReasonYikes
			waitingReason = &reason
		}
	} else if !event.ValuesRequired {
		// No values check required - check if host approval needed
		// For public events, auto-approve unless at capacity
		if event.Visibility == model.EventVisibilityPublic {
			if event.MaxAttendees == nil {
				status = model.RSVPStatusApproved
			} else {
				currentCount, _ := s.repo.CountApprovedRSVPs(ctx, eventID)
				if currentCount+1+req.PlusOnes <= *event.MaxAttendees {
					status = model.RSVPStatusApproved
				} else if event.WaitlistEnabled {
					status = model.RSVPStatusWaitlisted
					reason := model.WaitingReasonCapacity
					waitingReason = &reason
				}
			}
		}
	}

	// Create or update RSVP
	if existingRSVP != nil {
		updates := map[string]interface{}{
			"status":        status,
			"rsvp_type":     req.RSVPType,
			"plus_ones":     req.PlusOnes,
			"plus_one_names": req.PlusOneNames,
		}
		if waitingReason != nil {
			updates["waiting_reason"] = *waitingReason
		}
		if valuesCheck != nil {
			updates["values_aligned"] = valuesCheck.IsAligned
			updates["alignment_score"] = valuesCheck.AlignmentScore
			updates["yikes_count"] = valuesCheck.YikesCount
		}
		return s.repo.UpdateRSVP(ctx, existingRSVP.ID, updates)
	}

	rsvp := &model.EventRSVP{
		EventID:       eventID,
		UserID:        userID,
		Status:        status,
		RSVPType:      req.RSVPType,
		WaitingReason: waitingReason,
		PlusOnes:      req.PlusOnes,
		PlusOneNames:  req.PlusOneNames,
	}

	if valuesCheck != nil {
		rsvp.ValuesAligned = valuesCheck.IsAligned
		rsvp.AlignmentScore = valuesCheck.AlignmentScore
		rsvp.YikesCount = valuesCheck.YikesCount
	}

	if err := s.repo.CreateRSVP(ctx, rsvp); err != nil {
		return nil, err
	}

	return rsvp, nil
}

// CheckValuesAlignment checks a user's values alignment with an event
func (s *EventService) CheckValuesAlignment(ctx context.Context, userID string, event *model.Event) (*model.EventValuesCheck, error) {
	check := &model.EventValuesCheck{
		UserID:  userID,
		EventID: event.ID,
	}

	// Get user's answers
	answers, err := s.questionnaireService.GetUserAnswers(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build map of user's answers
	userAnswers := make(map[string]*model.Answer)
	for _, answer := range answers {
		userAnswers[answer.QuestionID] = answer
	}

	// Check required questions
	if len(event.ValuesQuestions) > 0 {
		missingQuestions := make([]string, 0)
		for _, qID := range event.ValuesQuestions {
			if _, ok := userAnswers[qID]; !ok {
				missingQuestions = append(missingQuestions, qID)
			}
		}
		check.MissingQuestions = missingQuestions

		if len(missingQuestions) > 0 {
			check.IsAligned = false
			check.Recommendation = model.ValuesRecommendNeedsReview
			return check, nil
		}
	}

	// Calculate compatibility with event creator
	if s.compatibilityService != nil {
		compat, err := s.compatibilityService.CalculateCompatibility(ctx, event.CreatedBy, userID)
		if err == nil && compat != nil {
			check.AlignmentScore = compat.Score
		}

		yikes, err := s.compatibilityService.CalculateYikesSummary(ctx, event.CreatedBy, userID)
		if err == nil && yikes != nil {
			check.YikesCount = yikes.YikesCount
			check.YikesCategories = yikes.Categories
		}
	}

	// Determine recommendation
	if check.YikesCount > event.YikesThreshold {
		check.IsAligned = false
		check.Recommendation = model.ValuesRecommendDeclineSuggested
	} else if check.AlignmentScore >= 70 && check.YikesCount == 0 {
		check.IsAligned = true
		check.Recommendation = model.ValuesRecommendAutoApprove
	} else {
		check.IsAligned = check.AlignmentScore >= 50
		check.Recommendation = model.ValuesRecommendNeedsReview
	}

	return check, nil
}

// RespondToRSVP allows host to approve or decline an RSVP
func (s *EventService) RespondToRSVP(ctx context.Context, hostUserID, eventID, rsvpUserID string, req *model.RespondToRSVPRequest) (*model.EventRSVP, error) {
	isHost, err := s.repo.IsHost(ctx, eventID, hostUserID)
	if err != nil {
		return nil, err
	}
	if !isHost {
		return nil, ErrNotEventHost
	}

	rsvp, err := s.repo.GetRSVP(ctx, eventID, rsvpUserID)
	if err != nil {
		return nil, err
	}
	if rsvp == nil {
		return nil, ErrRSVPNotFound
	}

	now := time.Now()
	updates := map[string]interface{}{
		"responded_by": hostUserID,
		"responded_on": now,
	}

	if req.Approved {
		updates["status"] = model.RSVPStatusApproved
		updates["waiting_reason"] = nil
	} else {
		updates["status"] = model.RSVPStatusDeclined
	}

	if req.Note != nil {
		updates["host_note"] = *req.Note
	}

	return s.repo.UpdateRSVP(ctx, rsvp.ID, updates)
}

// CancelRSVP allows a user to cancel their own RSVP
func (s *EventService) CancelRSVP(ctx context.Context, userID, eventID string) error {
	rsvp, err := s.repo.GetRSVP(ctx, eventID, userID)
	if err != nil {
		return err
	}
	if rsvp == nil {
		return ErrRSVPNotFound
	}

	_, err = s.repo.UpdateRSVP(ctx, rsvp.ID, map[string]interface{}{
		"status": model.RSVPStatusCancelled,
	})
	return err
}

// GetPendingRSVPs retrieves pending RSVPs for host review
func (s *EventService) GetPendingRSVPs(ctx context.Context, userID, eventID string) ([]*model.EventRSVP, error) {
	isHost, err := s.repo.IsHost(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}
	if !isHost {
		return nil, ErrNotEventHost
	}

	return s.repo.GetPendingRSVPs(ctx, eventID)
}

// GetGuildEvents retrieves events for a guild
func (s *EventService) GetGuildEvents(ctx context.Context, guildID string, filters *model.EventSearchFilters) ([]*model.Event, error) {
	return s.repo.GetByGuild(ctx, guildID, filters)
}

// GetPublicEvents retrieves public events
func (s *EventService) GetPublicEvents(ctx context.Context, filters *model.EventSearchFilters, limit int) ([]*model.Event, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.GetPublicEvents(ctx, filters, limit)
}

// ConfirmCompletion marks event attendance as confirmed (for Resonance)
func (s *EventService) ConfirmCompletion(ctx context.Context, userID, eventID string, completed bool) error {
	rsvp, err := s.repo.GetRSVP(ctx, eventID, userID)
	if err != nil {
		return err
	}
	if rsvp == nil {
		return ErrRSVPNotFound
	}

	_, err = s.repo.UpdateRSVP(ctx, rsvp.ID, map[string]interface{}{
		"completion_confirmed": completed,
	})
	return err
}

// Checkin records event check-in time (for Resonance)
func (s *EventService) Checkin(ctx context.Context, userID, eventID string) error {
	rsvp, err := s.repo.GetRSVP(ctx, eventID, userID)
	if err != nil {
		return err
	}
	if rsvp == nil {
		return ErrRSVPNotFound
	}

	now := time.Now()
	_, err = s.repo.UpdateRSVP(ctx, rsvp.ID, map[string]interface{}{
		"checkin_time": now,
	})
	return err
}

// SubmitFeedback submits helpfulness feedback for support events
func (s *EventService) SubmitFeedback(ctx context.Context, userID, eventID string, req *model.EventFeedbackRequest) error {
	rsvp, err := s.repo.GetRSVP(ctx, eventID, userID)
	if err != nil {
		return err
	}
	if rsvp == nil {
		return ErrRSVPNotFound
	}

	_, err = s.repo.UpdateRSVP(ctx, rsvp.ID, map[string]interface{}{
		"helpfulness_rating": req.HelpfulnessRating,
		"helpfulness_tags":   req.Tags,
	})
	return err
}
