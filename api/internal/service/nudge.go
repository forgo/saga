package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// NudgeService handles nudge generation and delivery
type NudgeService struct {
	availabilityRepo AvailabilityRepository
	poolRepo         PoolRepository
	eventHub         *EventHub
	pushService      *PushService
	configs          map[model.NudgeType]model.NudgeConfig
}

// NudgeServiceConfig holds configuration for the nudge service
type NudgeServiceConfig struct {
	AvailabilityRepo AvailabilityRepository
	PoolRepo         PoolRepository
	EventHub         *EventHub
	PushService      *PushService
}

// NewNudgeService creates a new nudge service
func NewNudgeService(cfg NudgeServiceConfig) *NudgeService {
	return &NudgeService{
		availabilityRepo: cfg.AvailabilityRepo,
		poolRepo:         cfg.PoolRepo,
		eventHub:         cfg.EventHub,
		pushService:      cfg.PushService,
		configs:          model.DefaultNudgeConfigs,
	}
}

// ProcessPendingNudges checks for and sends any due nudges
// This should be called periodically by a background job
func (s *NudgeService) ProcessPendingNudges(ctx context.Context) error {
	// Process different nudge types
	if err := s.processPendingMatchNudges(ctx); err != nil {
		log.Printf("Error processing pending match nudges: %v", err)
	}

	if err := s.processStaleHangoutNudges(ctx); err != nil {
		log.Printf("Error processing stale hangout nudges: %v", err)
	}

	if err := s.processUpcomingHangoutNudges(ctx); err != nil {
		log.Printf("Error processing upcoming hangout nudges: %v", err)
	}

	if err := s.processPendingRequestNudges(ctx); err != nil {
		log.Printf("Error processing pending request nudges: %v", err)
	}

	if err := s.processPoolMatchNudges(ctx); err != nil {
		log.Printf("Error processing pool match nudges: %v", err)
	}

	return nil
}

// processPendingMatchNudges sends reminders for matches not yet acted upon
func (s *NudgeService) processPendingMatchNudges(ctx context.Context) error {
	config := s.configs[model.NudgeTypePendingMatch]
	if !config.Enabled {
		return nil
	}

	// Get pool matches that are pending and older than delay threshold
	cutoff := time.Now().Add(-config.DelayAfter)

	if s.poolRepo == nil {
		return nil
	}

	matches, err := s.poolRepo.GetStaleMatches(ctx, cutoff, model.MatchStatusPending)
	if err != nil {
		return err
	}

	for _, match := range matches {
		for _, userID := range match.MemberUserIDs {
			nudge := s.buildNudge(model.NudgeTypePendingMatch, userID, match)
			s.sendNudge(ctx, nudge)
		}
	}

	return nil
}

// processStaleHangoutNudges sends reminders for scheduled hangouts that haven't been completed
func (s *NudgeService) processStaleHangoutNudges(ctx context.Context) error {
	config := s.configs[model.NudgeTypeStaleHangout]
	if !config.Enabled {
		return nil
	}

	if s.availabilityRepo == nil {
		return nil
	}

	// Get hangouts scheduled in the past that are still marked as "scheduled"
	cutoff := time.Now().Add(-config.DelayAfter)
	hangouts, err := s.availabilityRepo.GetStaleHangouts(ctx, cutoff, model.HangoutStatusScheduled)
	if err != nil {
		return err
	}

	for _, hangout := range hangouts {
		for _, userID := range hangout.Participants {
			nudge := s.buildHangoutNudge(model.NudgeTypeStaleHangout, userID, hangout)
			s.sendNudge(ctx, nudge)
		}
	}

	return nil
}

// processUpcomingHangoutNudges sends reminders for hangouts happening soon
func (s *NudgeService) processUpcomingHangoutNudges(ctx context.Context) error {
	config := s.configs[model.NudgeTypeUpcomingHangout]
	if !config.Enabled {
		return nil
	}

	if s.availabilityRepo == nil {
		return nil
	}

	// Get hangouts scheduled in the next 2 hours
	windowStart := time.Now()
	windowEnd := time.Now().Add(2 * time.Hour)

	hangouts, err := s.availabilityRepo.GetUpcomingHangouts(ctx, windowStart, windowEnd)
	if err != nil {
		return err
	}

	for _, hangout := range hangouts {
		for _, userID := range hangout.Participants {
			nudge := s.buildHangoutNudge(model.NudgeTypeUpcomingHangout, userID, hangout)
			s.sendNudge(ctx, nudge)
		}
	}

	return nil
}

// processPendingRequestNudges notifies users about pending hangout requests
func (s *NudgeService) processPendingRequestNudges(ctx context.Context) error {
	config := s.configs[model.NudgeTypePendingRequest]
	if !config.Enabled {
		return nil
	}

	if s.availabilityRepo == nil {
		return nil
	}

	// Get pending hangout requests
	requests, err := s.availabilityRepo.GetAllPendingRequests(ctx)
	if err != nil {
		return err
	}

	for _, req := range requests {
		// Get the availability to find the owner
		av, err := s.availabilityRepo.GetByID(ctx, req.AvailabilityID)
		if err != nil || av == nil {
			continue
		}

		nudge := s.buildRequestNudge(model.NudgeTypePendingRequest, av.UserID, req)
		s.sendNudge(ctx, nudge)
	}

	return nil
}

// processPoolMatchNudges handles pool-specific match nudges
func (s *NudgeService) processPoolMatchNudges(ctx context.Context) error {
	config := s.configs[model.NudgeTypePoolMatchStale]
	if !config.Enabled {
		return nil
	}

	if s.poolRepo == nil {
		return nil
	}

	// Get stale pool matches
	cutoff := time.Now().Add(-config.DelayAfter)
	matches, err := s.poolRepo.GetStaleMatches(ctx, cutoff, model.MatchStatusPending)
	if err != nil {
		return err
	}

	for _, match := range matches {
		for _, userID := range match.MemberUserIDs {
			nudge := s.buildNudge(model.NudgeTypePoolMatchStale, userID, match)
			s.sendNudge(ctx, nudge)
		}
	}

	return nil
}

// buildNudge creates a nudge for a pool match
func (s *NudgeService) buildNudge(nudgeType model.NudgeType, userID string, match *model.MatchResult) *model.Nudge {
	template := model.NudgeTemplates[nudgeType]

	// Build partner names (excluding current user)
	var partnerNames []string
	for i, memberUserID := range match.MemberUserIDs {
		if memberUserID != userID && i < len(match.MemberNames) {
			partnerNames = append(partnerNames, match.MemberNames[i])
		}
	}

	partnersStr := "your match"
	if len(partnerNames) > 0 {
		partnersStr = joinNames(partnerNames)
	}

	return &model.Nudge{
		UserID:  userID,
		Type:    nudgeType,
		Channel: s.configs[nudgeType].Channel,
		Title:   template.Title,
		Message: fmt.Sprintf(template.Message, partnersStr),
		Data: model.NudgeData{
			MatchID:      &match.ID,
			PoolID:       &match.PoolID,
			PartnerNames: partnerNames,
		},
		SentAt: time.Now(),
	}
}

// buildHangoutNudge creates a nudge for a hangout
func (s *NudgeService) buildHangoutNudge(nudgeType model.NudgeType, userID string, hangout *model.Hangout) *model.Nudge {
	template := model.NudgeTemplates[nudgeType]

	// Build partner description
	var partnerNames []string
	for _, pID := range hangout.Participants {
		if pID != userID {
			partnerNames = append(partnerNames, pID) // Would use actual names in production
		}
	}

	partnersStr := "someone"
	if len(partnerNames) > 0 {
		partnersStr = joinNames(partnerNames)
	}

	return &model.Nudge{
		UserID:  userID,
		Type:    nudgeType,
		Channel: s.configs[nudgeType].Channel,
		Title:   template.Title,
		Message: fmt.Sprintf(template.Message, partnersStr),
		Data: model.NudgeData{
			HangoutID:      &hangout.ID,
			PartnerUserIDs: hangout.Participants,
			ScheduledTime:  &hangout.ScheduledTime,
		},
		SentAt: time.Now(),
	}
}

// buildRequestNudge creates a nudge for a hangout request
func (s *NudgeService) buildRequestNudge(nudgeType model.NudgeType, userID string, req *model.HangoutRequest) *model.Nudge {
	template := model.NudgeTemplates[nudgeType]

	return &model.Nudge{
		UserID:  userID,
		Type:    nudgeType,
		Channel: s.configs[nudgeType].Channel,
		Title:   template.Title,
		Message: fmt.Sprintf(template.Message, "Someone"),
		Data: model.NudgeData{
			AvailabilityID: &req.AvailabilityID,
			PartnerUserID:  &req.RequesterID,
		},
		SentAt: time.Now(),
	}
}

// sendNudge delivers the nudge via the appropriate channel
func (s *NudgeService) sendNudge(ctx context.Context, nudge *model.Nudge) {
	switch nudge.Channel {
	case model.NudgeChannelSSE:
		s.sendSSENudge(nudge)
	case model.NudgeChannelPush:
		// Try push notification first
		if s.pushService != nil && s.pushService.IsEnabled() {
			notification := &PushNotification{
				Title: nudge.Title,
				Body:  nudge.Message,
				Data:  s.convertNudgeDataToStrings(nudge),
			}
			if _, err := s.pushService.SendToUser(ctx, nudge.UserID, notification); err != nil {
				// Fall back to SSE on push failure
				log.Printf("[NudgeService] Push failed for user %s, falling back to SSE: %v", nudge.UserID, err)
				s.sendSSENudge(nudge)
			}
		} else {
			// Push not available, fall back to SSE
			s.sendSSENudge(nudge)
		}
	}
}

// convertNudgeDataToStrings converts nudge data to string map for push payload
func (s *NudgeService) convertNudgeDataToStrings(nudge *model.Nudge) map[string]string {
	result := make(map[string]string)
	result["nudge_type"] = string(nudge.Type)
	result["sent_at"] = nudge.SentAt.Format(time.RFC3339)

	// Data is a value type, check individual fields
	if nudge.Data.MatchID != nil {
		result["match_id"] = *nudge.Data.MatchID
	}
	if nudge.Data.AvailabilityID != nil {
		result["availability_id"] = *nudge.Data.AvailabilityID
	}
	if nudge.Data.PartnerUserID != nil {
		result["partner_user_id"] = *nudge.Data.PartnerUserID
	}

	return result
}

// sendSSENudge sends nudge via server-sent events
func (s *NudgeService) sendSSENudge(nudge *model.Nudge) {
	if s.eventHub == nil {
		return
	}

	// Create SSE event
	event := Event{
		Type: EventNudge,
		Data: map[string]interface{}{
			"nudge_type": nudge.Type,
			"title":      nudge.Title,
			"message":    nudge.Message,
			"data":       nudge.Data,
			"sent_at":    nudge.SentAt,
		},
	}

	s.eventHub.SendToUser(nudge.UserID, event)
}

// GetNudgeSummary returns a summary of actionable items for a user
func (s *NudgeService) GetNudgeSummary(ctx context.Context, userID string) (*model.NudgeSummary, error) {
	summary := &model.NudgeSummary{
		UserID: userID,
	}

	// Count pending pool matches
	if s.poolRepo != nil {
		matches, err := s.poolRepo.GetUserPendingMatches(ctx, userID)
		if err == nil {
			summary.PendingMatches = len(matches)
		}
	}

	// Count pending hangout requests
	if s.availabilityRepo != nil {
		requests, err := s.availabilityRepo.GetPendingRequestsForUser(ctx, userID)
		if err == nil {
			summary.PendingRequests = len(requests)
		}
	}

	// Count upcoming hangouts
	if s.availabilityRepo != nil {
		hangouts, err := s.availabilityRepo.GetUserUpcomingHangouts(ctx, userID, time.Now(), time.Now().Add(24*time.Hour))
		if err == nil {
			summary.UpcomingHangouts = len(hangouts)
		}
	}

	summary.TotalActionable = summary.PendingMatches + summary.PendingRequests

	return summary, nil
}

// Helper function to join names nicely
func joinNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	if len(names) == 1 {
		return names[0]
	}
	if len(names) == 2 {
		return names[0] + " and " + names[1]
	}
	return names[0] + ", " + names[1] + ", and others"
}
