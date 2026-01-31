package service

import (
	"context"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// ResonanceRepository defines the interface for resonance storage
type ResonanceRepository interface {
	AwardPoints(ctx context.Context, entry *model.ResonanceLedgerEntry) error
	HasAwardedPoints(ctx context.Context, userID, stat, sourceObjectID string) (bool, error)
	GetUserLedger(ctx context.Context, userID string, limit, offset int) ([]*model.ResonanceLedgerEntry, error)
	GetUserScore(ctx context.Context, userID string) (*model.ResonanceScore, error)
	RecalculateUserScore(ctx context.Context, userID string) (*model.ResonanceScore, error)
	GetDailyCap(ctx context.Context, userID string, date string) (*model.ResonanceDailyCap, error)
	IncrementDailyCap(ctx context.Context, userID string, date string, stat model.ResonanceStat, amount int) error
	GetSupportPairCount(ctx context.Context, helperID, receiverID string) (int, error)
	IncrementSupportPairCount(ctx context.Context, helperID, receiverID string) error
	// Nexus calculation methods
	GetAllActiveUserIDs(ctx context.Context) ([]string, error)
	GetUserCirclesForNexus(ctx context.Context, userID string) ([]*model.NexusCircleData, error)
	GetCirclePairOverlap(ctx context.Context, circleID1, circleID2 string) (int, error)
}

// ResonanceService handles resonance scoring business logic
type ResonanceService struct {
	repo ResonanceRepository
}

// ResonanceServiceConfig holds configuration for the resonance service
type ResonanceServiceConfig struct {
	Repo ResonanceRepository
}

// NewResonanceService creates a new resonance service
func NewResonanceService(cfg ResonanceServiceConfig) *ResonanceService {
	return &ResonanceService{
		repo: cfg.Repo,
	}
}

// GetUserScore retrieves a user's resonance score
func (s *ResonanceService) GetUserScore(ctx context.Context, userID string) (*model.ResonanceScore, error) {
	return s.repo.GetUserScore(ctx, userID)
}

// GetUserLedger retrieves a user's resonance ledger
func (s *ResonanceService) GetUserLedger(ctx context.Context, userID string, limit, offset int) ([]*model.ResonanceLedgerEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.GetUserLedger(ctx, userID, limit, offset)
}

// RecalculateScore recalculates a user's total score
func (s *ResonanceService) RecalculateScore(ctx context.Context, userID string) (*model.ResonanceScore, error) {
	return s.repo.RecalculateUserScore(ctx, userID)
}

// AwardQuesting awards Questing points for verified event completion
func (s *ResonanceService) AwardQuesting(ctx context.Context, userID, eventID string, earlyConfirm, onTimeCheckin bool) error {
	today := time.Now().Format("2006-01-02")

	// Check daily cap
	cap, err := s.repo.GetDailyCap(ctx, userID, today)
	if err != nil {
		return err
	}

	remaining := model.DailyCapQuesting - cap.QuestingEarned
	if remaining <= 0 {
		return nil // Cap reached, no points awarded
	}

	// Base points for completion
	points := model.PointsQuestingBase

	// Early confirm bonus
	if earlyConfirm {
		points += model.PointsQuestingEarlyConfirm
	}

	// On-time checkin bonus
	if onTimeCheckin {
		points += model.PointsQuestingCheckin
	}

	// Apply cap
	if points > remaining {
		points = remaining
	}

	// Check idempotency
	sourceID := "event:" + eventID
	awarded, err := s.repo.HasAwardedPoints(ctx, userID, string(model.ResonanceStatQuesting), sourceID)
	if err != nil {
		return err
	}
	if awarded {
		return nil // Already awarded
	}

	// Award points
	entry := &model.ResonanceLedgerEntry{
		UserID:         userID,
		Stat:           model.ResonanceStatQuesting,
		Points:         points,
		SourceObjectID: sourceID,
		ReasonCode:     model.ReasonQuestingCompletion,
	}

	if err := s.repo.AwardPoints(ctx, entry); err != nil {
		return err
	}

	// Update daily cap
	if err := s.repo.IncrementDailyCap(ctx, userID, today, model.ResonanceStatQuesting, points); err != nil {
		return err
	}

	// Recalculate score async (both return values intentionally ignored - fire and forget)
	go func() { _, _ = s.repo.RecalculateUserScore(context.Background(), userID) }()

	return nil
}

// AwardMana awards Mana points for helpful support sessions
func (s *ResonanceService) AwardMana(ctx context.Context, helperID, receiverID, hangoutID string, helpfulRating string, earlyConfirm, hasHelpfulTag bool) error {
	// Only award if receiver rated as helpful
	rating := model.HelpfulnessRating(helpfulRating)
	if !rating.IsHelpful() {
		return nil
	}

	today := time.Now().Format("2006-01-02")

	// Check daily cap
	cap, err := s.repo.GetDailyCap(ctx, helperID, today)
	if err != nil {
		return err
	}

	remaining := model.DailyCapMana - cap.ManaEarned
	if remaining <= 0 {
		return nil
	}

	// Get pairwise session count for diminishing returns
	pairCount, err := s.repo.GetSupportPairCount(ctx, helperID, receiverID)
	if err != nil {
		return err
	}

	// Calculate multiplier based on session count
	multiplier := 1.0
	if pairCount >= 6 {
		multiplier = 0.25
	} else if pairCount >= 3 {
		multiplier = 0.5
	}

	// Base points
	points := int(float64(model.PointsManaBase) * multiplier)

	// Early confirm bonus
	if earlyConfirm {
		points += int(float64(model.PointsManaEarlyConfirm) * multiplier)
	}

	// Helpful tag bonus
	if hasHelpfulTag {
		points += int(float64(model.PointsManaFeedbackTag) * multiplier)
	}

	// Apply cap
	if points > remaining {
		points = remaining
	}

	if points <= 0 {
		return nil
	}

	// Check idempotency
	sourceID := "hangout:" + hangoutID
	awarded, err := s.repo.HasAwardedPoints(ctx, helperID, string(model.ResonanceStatMana), sourceID)
	if err != nil {
		return err
	}
	if awarded {
		return nil
	}

	// Award points
	entry := &model.ResonanceLedgerEntry{
		UserID:         helperID,
		Stat:           model.ResonanceStatMana,
		Points:         points,
		SourceObjectID: sourceID,
		ReasonCode:     model.ReasonManaSupport,
	}

	if err := s.repo.AwardPoints(ctx, entry); err != nil {
		return err
	}

	// Update daily cap
	if err := s.repo.IncrementDailyCap(ctx, helperID, today, model.ResonanceStatMana, points); err != nil {
		return err
	}

	// Increment pair count
	if err := s.repo.IncrementSupportPairCount(ctx, helperID, receiverID); err != nil {
		return err
	}

	// Recalculate score async (both return values intentionally ignored - fire and forget)
	go func() { _, _ = s.repo.RecalculateUserScore(context.Background(), helperID) }()

	return nil
}

// AwardWayfinder awards Wayfinder points for hosting verified events
func (s *ResonanceService) AwardWayfinder(ctx context.Context, hostID, eventID string, verifiedAttendees int, earlyConfirm bool) error {
	today := time.Now().Format("2006-01-02")

	// Check daily cap
	cap, err := s.repo.GetDailyCap(ctx, hostID, today)
	if err != nil {
		return err
	}

	remaining := model.DailyCapWayfinder - cap.WayfinderEarned
	if remaining <= 0 {
		return nil
	}

	// Cap attendees at 4 to prevent mega-event farming
	attendees := verifiedAttendees
	if attendees > 4 {
		attendees = 4
	}

	// Base points + per-attendee bonus
	points := model.PointsWayfinderBase + (model.PointsWayfinderPerAttendee * attendees)

	// Early confirm bonus
	if earlyConfirm {
		points += model.PointsWayfinderEarlyConfirm
	}

	// Apply cap
	if points > remaining {
		points = remaining
	}

	// Check idempotency
	sourceID := "event:" + eventID
	awarded, err := s.repo.HasAwardedPoints(ctx, hostID, string(model.ResonanceStatWayfinder), sourceID)
	if err != nil {
		return err
	}
	if awarded {
		return nil
	}

	// Award points
	entry := &model.ResonanceLedgerEntry{
		UserID:         hostID,
		Stat:           model.ResonanceStatWayfinder,
		Points:         points,
		SourceObjectID: sourceID,
		ReasonCode:     model.ReasonWayfinderHosting,
	}

	if err := s.repo.AwardPoints(ctx, entry); err != nil {
		return err
	}

	// Update daily cap
	if err := s.repo.IncrementDailyCap(ctx, hostID, today, model.ResonanceStatWayfinder, points); err != nil {
		return err
	}

	// Recalculate score async (both return values intentionally ignored - fire and forget)
	go func() { _, _ = s.repo.RecalculateUserScore(context.Background(), hostID) }()

	return nil
}

// AwardAttunement awards Attunement points for answering questions
func (s *ResonanceService) AwardAttunement(ctx context.Context, userID, questionID string) error {
	today := time.Now().Format("2006-01-02")

	// Check daily cap
	cap, err := s.repo.GetDailyCap(ctx, userID, today)
	if err != nil {
		return err
	}

	remaining := model.DailyCapAttunement - cap.AttunementEarned
	if remaining <= 0 {
		return nil
	}

	points := model.PointsAttunementQuestion
	if points > remaining {
		points = remaining
	}

	// Check idempotency (first-time answer only)
	sourceID := "question:" + questionID
	awarded, err := s.repo.HasAwardedPoints(ctx, userID, string(model.ResonanceStatAttunement), sourceID)
	if err != nil {
		return err
	}
	if awarded {
		return nil
	}

	// Award points
	entry := &model.ResonanceLedgerEntry{
		UserID:         userID,
		Stat:           model.ResonanceStatAttunement,
		Points:         points,
		SourceObjectID: sourceID,
		ReasonCode:     model.ReasonAttunementQuestion,
	}

	if err := s.repo.AwardPoints(ctx, entry); err != nil {
		return err
	}

	// Update daily cap
	if err := s.repo.IncrementDailyCap(ctx, userID, today, model.ResonanceStatAttunement, points); err != nil {
		return err
	}

	// Recalculate score async (both return values intentionally ignored - fire and forget)
	go func() { _, _ = s.repo.RecalculateUserScore(context.Background(), userID) }()

	return nil
}

// AwardMonthlyProfileRefresh awards Attunement points for monthly profile update
func (s *ResonanceService) AwardMonthlyProfileRefresh(ctx context.Context, userID string) error {
	today := time.Now().Format("2006-01-02")
	month := time.Now().Format("2006-01")

	// Check daily cap
	cap, err := s.repo.GetDailyCap(ctx, userID, today)
	if err != nil {
		return err
	}

	remaining := model.DailyCapAttunement - cap.AttunementEarned
	if remaining <= 0 {
		return nil
	}

	points := model.PointsAttunementProfileRefresh
	if points > remaining {
		points = remaining
	}

	// Check idempotency (once per month)
	sourceID := "month:" + month
	awarded, err := s.repo.HasAwardedPoints(ctx, userID, string(model.ResonanceStatAttunement), sourceID)
	if err != nil {
		return err
	}
	if awarded {
		return nil
	}

	// Award points
	entry := &model.ResonanceLedgerEntry{
		UserID:         userID,
		Stat:           model.ResonanceStatAttunement,
		Points:         points,
		SourceObjectID: sourceID,
		ReasonCode:     model.ReasonAttunementProfileRefresh,
	}

	if err := s.repo.AwardPoints(ctx, entry); err != nil {
		return err
	}

	// Update daily cap
	if err := s.repo.IncrementDailyCap(ctx, userID, today, model.ResonanceStatAttunement, points); err != nil {
		return err
	}

	// Recalculate score async (both return values intentionally ignored - fire and forget)
	go func() { _, _ = s.repo.RecalculateUserScore(context.Background(), userID) }()

	return nil
}

// AwardNexus awards Nexus points (called monthly by batch job)
// This is complex and involves circle activity calculations
func (s *ResonanceService) AwardNexus(ctx context.Context, userID string, circleContributions []model.CircleNexusContribution) error {
	month := time.Now().Format("2006-01")

	// Calculate total nexus from contributions
	totalNexus := 0
	for _, contrib := range circleContributions {
		totalNexus += contrib.Points
	}

	// Apply monthly cap
	if totalNexus > model.MonthlyCapNexus {
		totalNexus = model.MonthlyCapNexus
	}

	if totalNexus <= 0 {
		return nil
	}

	// Check idempotency
	sourceID := "nexus:" + month
	awarded, err := s.repo.HasAwardedPoints(ctx, userID, string(model.ResonanceStatNexus), sourceID)
	if err != nil {
		return err
	}
	if awarded {
		return nil
	}

	// Award points
	entry := &model.ResonanceLedgerEntry{
		UserID:         userID,
		Stat:           model.ResonanceStatNexus,
		Points:         totalNexus,
		SourceObjectID: sourceID,
		ReasonCode:     model.ReasonNexusMonthly,
	}

	if err := s.repo.AwardPoints(ctx, entry); err != nil {
		return err
	}

	// Recalculate score
	_, err = s.repo.RecalculateUserScore(ctx, userID)
	return err
}

// GetAllActiveUserIDs returns IDs of users active in the last 30 days
func (s *ResonanceService) GetAllActiveUserIDs(ctx context.Context) ([]string, error) {
	return s.repo.GetAllActiveUserIDs(ctx)
}

// GetUserCirclesForNexus returns circle activity data for Nexus calculation
func (s *ResonanceService) GetUserCirclesForNexus(ctx context.Context, userID string) ([]*model.NexusCircleData, error) {
	return s.repo.GetUserCirclesForNexus(ctx, userID)
}

// GetCirclePairOverlap returns count of users active in both circles
func (s *ResonanceService) GetCirclePairOverlap(ctx context.Context, circleID1, circleID2 string) (int, error) {
	return s.repo.GetCirclePairOverlap(ctx, circleID1, circleID2)
}
