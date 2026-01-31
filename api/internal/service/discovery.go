package service

import (
	"context"
	"sort"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// BlockChecker defines the interface for block checking in discovery
type BlockChecker interface {
	IsBlockedEitherWay(ctx context.Context, userID1, userID2 string) (bool, error)
}

// DiscoveryService handles global people matching across the platform
// This service is NOT circle-bound - it finds compatible people anywhere
type DiscoveryService struct {
	availabilityRepo  AvailabilityRepository
	compatibilityRepo QuestionnaireRepository
	interestRepo      InterestRepository
	profileRepo       ProfileRepository
	blockChecker      BlockChecker
	geoService        *GeoService
}

// DiscoveryServiceConfig holds configuration for the discovery service
type DiscoveryServiceConfig struct {
	AvailabilityRepo  AvailabilityRepository
	CompatibilityRepo QuestionnaireRepository
	InterestRepo      InterestRepository
	ProfileRepo       ProfileRepository
	BlockChecker      BlockChecker
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(cfg DiscoveryServiceConfig) *DiscoveryService {
	return &DiscoveryService{
		availabilityRepo:  cfg.AvailabilityRepo,
		compatibilityRepo: cfg.CompatibilityRepo,
		interestRepo:      cfg.InterestRepo,
		profileRepo:       cfg.ProfileRepo,
		blockChecker:      cfg.BlockChecker,
		geoService:        NewGeoService(),
	}
}

// isBlocked checks if two users have blocked each other
func (s *DiscoveryService) isBlocked(ctx context.Context, userID1, userID2 string) bool {
	if s.blockChecker == nil {
		return false
	}
	blocked, err := s.blockChecker.IsBlockedEitherWay(ctx, userID1, userID2)
	if err != nil {
		return false // Fail open to avoid breaking discovery on errors
	}
	return blocked
}

// PeopleDiscoveryFilter defines criteria for finding people
type PeopleDiscoveryFilter struct {
	// Location-based filtering
	CenterLat *float64 `json:"center_lat,omitempty"`
	CenterLng *float64 `json:"center_lng,omitempty"`
	RadiusKm  float64  `json:"radius_km,omitempty"` // Default: 25km

	// Time-based filtering (for availability)
	StartAfter *time.Time `json:"start_after,omitempty"`
	EndBefore  *time.Time `json:"end_before,omitempty"`

	// Activity filtering
	HangoutTypes []model.HangoutType `json:"hangout_types,omitempty"`
	InterestID   *string             `json:"interest_id,omitempty"` // Specific interest

	// Matching preferences
	MinCompatibility    float64 `json:"min_compatibility,omitempty"` // Minimum compatibility % (0-100)
	RequireSharedAnswer bool    `json:"require_shared_answer"`       // Must have answered at least one question

	// Result controls
	Limit  int `json:"limit,omitempty"`  // Default: 20, max: 50
	Offset int `json:"offset,omitempty"` // Pagination
}

// DiscoveryResult represents a person match with scoring details
type DiscoveryResult struct {
	UserID string `json:"user_id"`

	// Profile info (privacy-respecting)
	Profile *model.PublicProfile `json:"profile,omitempty"`

	// Availability details (if they have one)
	Availability *model.AvailabilityPublic `json:"availability,omitempty"`

	// Scoring components
	CompatibilityScore float64                `json:"compatibility_score"` // 0-100
	SharedInterests    []SharedInterestBrief  `json:"shared_interests,omitempty"`
	Distance           model.DistanceBucket   `json:"distance,omitempty"`
	ActivityRecency    model.FreshnessBucket  `json:"activity_recency,omitempty"`

	// Aggregate score used for ranking
	MatchScore float64 `json:"match_score"` // Combined weighted score
}

// SharedInterestBrief is a compact representation of a shared interest
type SharedInterestBrief struct {
	InterestID   string `json:"interest_id"`
	InterestName string `json:"interest_name"`
	Category     string `json:"category"`
	// Whether this is a teach/learn opportunity
	TeachLearnMatch bool `json:"teach_learn_match,omitempty"`
}

// DiscoveryResponse wraps results with metadata
type DiscoveryResponse struct {
	Results     []DiscoveryResult `json:"results"`
	TotalCount  int               `json:"total_count"`
	HasMore     bool              `json:"has_more"`
	Offset      int               `json:"offset"`
	SearchedAt  time.Time         `json:"searched_at"`
}

// DiscoverPeople finds compatible people based on the filter criteria
// This is the main entry point for global people matching
func (s *DiscoveryService) DiscoverPeople(ctx context.Context, requesterID string, filter PeopleDiscoveryFilter) (*DiscoveryResponse, error) {
	// Apply defaults
	if filter.Limit <= 0 || filter.Limit > 50 {
		filter.Limit = 20
	}
	if filter.RadiusKm <= 0 {
		filter.RadiusKm = DefaultSearchRadiusKm
	}
	if filter.RadiusKm > MaxSearchRadiusKm {
		filter.RadiusKm = MaxSearchRadiusKm
	}

	// Step 1: Get available people (those with active/public availability)
	candidates, err := s.findCandidates(ctx, requesterID, filter)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return &DiscoveryResponse{
			Results:    []DiscoveryResult{},
			TotalCount: 0,
			HasMore:    false,
			SearchedAt: time.Now(),
		}, nil
	}

	// Step 2: Enrich with compatibility scores
	results, err := s.enrichWithScores(ctx, requesterID, candidates, filter)
	if err != nil {
		return nil, err
	}

	// Step 3: Filter by minimum compatibility (if specified)
	if filter.MinCompatibility > 0 {
		filtered := make([]DiscoveryResult, 0, len(results))
		for _, r := range results {
			if r.CompatibilityScore >= filter.MinCompatibility {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}

	// Step 4: Calculate combined match scores and sort
	s.calculateMatchScores(results)
	sort.Slice(results, func(i, j int) bool {
		return results[i].MatchScore > results[j].MatchScore
	})

	// Step 5: Apply pagination
	totalCount := len(results)
	hasMore := false

	if filter.Offset > 0 {
		if filter.Offset >= len(results) {
			results = []DiscoveryResult{}
		} else {
			results = results[filter.Offset:]
		}
	}

	if len(results) > filter.Limit {
		results = results[:filter.Limit]
		hasMore = true
	}

	return &DiscoveryResponse{
		Results:    results,
		TotalCount: totalCount,
		HasMore:    hasMore,
		Offset:     filter.Offset,
		SearchedAt: time.Now(),
	}, nil
}

// findCandidates retrieves potential matches based on availability and location
func (s *DiscoveryService) findCandidates(ctx context.Context, requesterID string, filter PeopleDiscoveryFilter) ([]*model.Availability, error) {
	// Build time range (default: now to 24 hours from now)
	startTime := time.Now()
	endTime := time.Now().Add(24 * time.Hour)

	if filter.StartAfter != nil {
		startTime = *filter.StartAfter
	}
	if filter.EndBefore != nil {
		endTime = *filter.EndBefore
	}

	// Calculate how many candidates to fetch from database
	// We need more than (offset + limit) to account for:
	// - Block filtering (some will be filtered out)
	// - Compatibility filtering (some may not meet threshold)
	// - Deduplication
	// Use 3x multiplier with a minimum of 50 and maximum of 200
	candidateLimit := (filter.Offset + filter.Limit) * 3
	if candidateLimit < 50 {
		candidateLimit = 50
	}
	if candidateLimit > 200 {
		candidateLimit = 200
	}

	var candidates []*model.Availability

	// If location provided, search nearby
	if filter.CenterLat != nil && filter.CenterLng != nil {
		bbox := s.geoService.GetBoundingBox(*filter.CenterLat, *filter.CenterLng, filter.RadiusKm)
		nearby, err := s.availabilityRepo.GetNearby(
			ctx,
			bbox.MinLat, bbox.MaxLat,
			bbox.MinLng, bbox.MaxLng,
			startTime, endTime,
			requesterID,
			candidateLimit, // Dynamic limit based on pagination needs
		)
		if err != nil {
			return nil, err
		}
		candidates = nearby
	}

	// If hangout type specified, filter or search by type
	if len(filter.HangoutTypes) > 0 {
		if len(candidates) == 0 {
			// No location search, get by type
			// Divide limit among hangout types
			limitPerType := candidateLimit / len(filter.HangoutTypes)
			if limitPerType < 20 {
				limitPerType = 20
			}
			for _, ht := range filter.HangoutTypes {
				typed, err := s.availabilityRepo.GetByHangoutType(ctx, string(ht), requesterID, limitPerType)
				if err != nil {
					return nil, err
				}
				candidates = append(candidates, typed...)
			}
		} else {
			// Filter location results by type
			filtered := make([]*model.Availability, 0, len(candidates))
			typeSet := make(map[model.HangoutType]bool)
			for _, ht := range filter.HangoutTypes {
				typeSet[ht] = true
			}
			for _, c := range candidates {
				if typeSet[c.HangoutType] {
					filtered = append(filtered, c)
				}
			}
			candidates = filtered
		}
	}

	// If interest specified, filter by interest match
	if filter.InterestID != nil {
		filtered := make([]*model.Availability, 0)
		for _, c := range candidates {
			if c.InterestID != nil && *c.InterestID == *filter.InterestID {
				filtered = append(filtered, c)
			}
		}
		candidates = filtered
	}

	return candidates, nil
}

// enrichWithScores adds compatibility scores, shared interests, and profile info
func (s *DiscoveryService) enrichWithScores(ctx context.Context, requesterID string, candidates []*model.Availability, filter PeopleDiscoveryFilter) ([]DiscoveryResult, error) {
	results := make([]DiscoveryResult, 0, len(candidates))

	// Get requester's location for distance calculations
	var requesterLat, requesterLng float64
	if filter.CenterLat != nil && filter.CenterLng != nil {
		requesterLat = *filter.CenterLat
		requesterLng = *filter.CenterLng
	}

	// Get requester's interests for matching
	requesterInterests, _ := s.interestRepo.GetUserInterests(ctx, requesterID)
	requesterInterestSet := make(map[string]*model.UserInterest)
	for _, ui := range requesterInterests {
		requesterInterestSet[ui.InterestID] = ui
	}

	for _, candidate := range candidates {
		// SECURITY: Skip blocked users
		if s.isBlocked(ctx, requesterID, candidate.UserID) {
			continue
		}

		result := DiscoveryResult{
			UserID: candidate.UserID,
		}

		// Get compatibility score
		if s.compatibilityRepo != nil {
			sharedAnswers, err := s.compatibilityRepo.GetSharedAnswers(ctx, requesterID, candidate.UserID)
			if err == nil && len(sharedAnswers) > 0 {
				// Calculate compatibility
				score := s.calculateCompatibilityFromAnswers(sharedAnswers)
				result.CompatibilityScore = score
			} else if filter.RequireSharedAnswer {
				continue // Skip if no shared answers and it's required
			}
		}

		// Get shared interests
		if s.interestRepo != nil {
			candidateInterests, err := s.interestRepo.GetUserInterests(ctx, candidate.UserID)
			if err == nil {
				for _, ci := range candidateInterests {
					if ri, ok := requesterInterestSet[ci.InterestID]; ok {
						// Check for teach/learn opportunity
						teachLearn := (ri.WantsToLearn && ci.WantsToTeach) ||
							(ri.WantsToTeach && ci.WantsToLearn)

						result.SharedInterests = append(result.SharedInterests, SharedInterestBrief{
							InterestID:      ci.InterestID,
							InterestName:    ci.Name,
							Category:        ci.Category,
							TeachLearnMatch: teachLearn,
						})
					}
				}
			}
		}

		// Calculate distance bucket (privacy-preserving)
		if candidate.Location != nil && filter.CenterLat != nil {
			distance := s.geoService.HaversineDistance(
				requesterLat, requesterLng,
				candidate.Location.Lat, candidate.Location.Lng,
			)
			result.Distance = model.GetDistanceBucket(distance)
		}

		// Get public profile
		if s.profileRepo != nil {
			profile, err := s.profileRepo.GetByUserID(ctx, candidate.UserID)
			if err == nil && profile != nil {
				result.Profile = profile.ToPublic()
			}
		}

		// Build public availability view
		result.Availability = &model.AvailabilityPublic{
			ID:                  candidate.ID,
			UserID:              candidate.UserID,
			Status:              string(candidate.Status),
			StartTime:           candidate.StartTime,
			EndTime:             candidate.EndTime,
			Distance:            result.Distance,
			HangoutType:         candidate.HangoutType,
			ActivityDescription: candidate.ActivityDescription,
			ActivityVenue:       candidate.ActivityVenue,
			InterestID:          candidate.InterestID,
			MaxPeople:           candidate.MaxPeople,
			Note:                candidate.Note,
		}

		// Determine activity recency
		result.ActivityRecency = model.FreshnessBucketActiveNow // They have active availability

		results = append(results, result)
	}

	return results, nil
}

// calculateCompatibilityFromAnswers computes compatibility score from shared answers
func (s *DiscoveryService) calculateCompatibilityFromAnswers(sharedAnswers map[string][2]*model.Answer) float64 {
	if len(sharedAnswers) == 0 {
		return 0
	}

	var totalWeight float64
	var earnedPoints float64
	dealBreakerViolated := false

	for _, answers := range sharedAnswers {
		answerA := answers[0]
		answerB := answers[1]
		if answerA == nil || answerB == nil {
			continue
		}

		// Check for dealbreaker violations
		if answerA.IsDealBreaker && !containsString(answerA.AcceptableOptions, answerB.SelectedOption) {
			dealBreakerViolated = true
		}
		if answerB.IsDealBreaker && !containsString(answerB.AcceptableOptions, answerA.SelectedOption) {
			dealBreakerViolated = true
		}

		// Calculate weights
		weightA := float64(model.ImportanceWeight(answerA.Importance)) * (0.5 + answerA.AlignmentWeight*0.5)
		weightB := float64(model.ImportanceWeight(answerB.Importance)) * (0.5 + answerB.AlignmentWeight*0.5)

		totalWeight += weightA + weightB

		// Check acceptability
		if containsString(answerA.AcceptableOptions, answerB.SelectedOption) {
			earnedPoints += weightA
		}
		if containsString(answerB.AcceptableOptions, answerA.SelectedOption) {
			earnedPoints += weightB
		}

		// Yikes penalty
		if containsString(answerA.YikesOptions, answerB.SelectedOption) {
			earnedPoints -= weightA * 0.25
		}
		if containsString(answerB.YikesOptions, answerA.SelectedOption) {
			earnedPoints -= weightB * 0.25
		}
	}

	if dealBreakerViolated {
		return 0
	}

	if totalWeight == 0 {
		return 0
	}

	score := (earnedPoints / totalWeight) * 100
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}

	return score
}

// calculateMatchScores computes a combined score for ranking
func (s *DiscoveryService) calculateMatchScores(results []DiscoveryResult) {
	for i := range results {
		r := &results[i]

		// Base score from compatibility (0-100)
		score := r.CompatibilityScore

		// Bonus for shared interests (up to +20)
		interestBonus := float64(len(r.SharedInterests)) * 4
		if interestBonus > 20 {
			interestBonus = 20
		}
		score += interestBonus

		// Extra bonus for teach/learn opportunities (+5 each, max +15)
		teachLearnBonus := 0.0
		for _, si := range r.SharedInterests {
			if si.TeachLearnMatch {
				teachLearnBonus += 5
			}
		}
		if teachLearnBonus > 15 {
			teachLearnBonus = 15
		}
		score += teachLearnBonus

		// Distance bonus (closer = better, up to +10)
		distanceBonus := 0.0
		switch r.Distance {
		case model.DistanceNearby:
			distanceBonus = 10
		case model.Distance2km:
			distanceBonus = 8
		case model.Distance5km:
			distanceBonus = 5
		case model.Distance10km:
			distanceBonus = 2
		}
		score += distanceBonus

		r.MatchScore = score
	}
}

// DiscoverByInterest finds people with a specific shared interest
func (s *DiscoveryService) DiscoverByInterest(ctx context.Context, requesterID, interestID string, limit int) ([]DiscoveryResult, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	// Get users with this interest
	usersWithInterest, err := s.interestRepo.GetUsersWithInterest(ctx, interestID)
	if err != nil {
		return nil, err
	}

	// Get requester's interest level for this interest
	requesterInterests, _ := s.interestRepo.GetUserInterests(ctx, requesterID)
	var requesterInterest *model.UserInterest
	for _, ui := range requesterInterests {
		if ui.InterestID == interestID {
			requesterInterest = ui
			break
		}
	}

	results := make([]DiscoveryResult, 0)
	for _, ui := range usersWithInterest {
		if ui.UserID == requesterID {
			continue // Skip self
		}

		// SECURITY: Skip blocked users
		if s.isBlocked(ctx, requesterID, ui.UserID) {
			continue
		}

		result := DiscoveryResult{
			UserID: ui.UserID,
		}

		// Check for teach/learn match
		teachLearn := false
		if requesterInterest != nil {
			teachLearn = (requesterInterest.WantsToLearn && ui.WantsToTeach) ||
				(requesterInterest.WantsToTeach && ui.WantsToLearn)
		}

		result.SharedInterests = []SharedInterestBrief{{
			InterestID:      ui.InterestID,
			InterestName:    ui.Name,
			Category:        ui.Category,
			TeachLearnMatch: teachLearn,
		}}

		// Get compatibility score
		if s.compatibilityRepo != nil {
			sharedAnswers, err := s.compatibilityRepo.GetSharedAnswers(ctx, requesterID, ui.UserID)
			if err == nil {
				result.CompatibilityScore = s.calculateCompatibilityFromAnswers(sharedAnswers)
			}
		}

		// Get public profile
		if s.profileRepo != nil {
			profile, err := s.profileRepo.GetByUserID(ctx, ui.UserID)
			if err == nil && profile != nil {
				result.Profile = profile.ToPublic()
			}
		}

		results = append(results, result)
	}

	// Calculate scores and sort
	s.calculateMatchScores(results)
	sort.Slice(results, func(i, j int) bool {
		return results[i].MatchScore > results[j].MatchScore
	})

	// Apply limit
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// FindTeachLearnMatches finds people with complementary teach/learn interests
func (s *DiscoveryService) FindTeachLearnMatches(ctx context.Context, requesterID string, limit int) ([]DiscoveryResult, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	// Get requester's interests
	requesterInterests, err := s.interestRepo.GetUserInterests(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	// Separate into teach and learn sets
	wantsToTeach := make(map[string]bool)
	wantsToLearn := make(map[string]bool)
	for _, ui := range requesterInterests {
		if ui.WantsToTeach {
			wantsToTeach[ui.InterestID] = true
		}
		if ui.WantsToLearn {
			wantsToLearn[ui.InterestID] = true
		}
	}

	// Find potential teachers for things requester wants to learn
	results := make(map[string]*DiscoveryResult) // keyed by user ID

	for interestID := range wantsToLearn {
		teachers, err := s.interestRepo.GetTeachersForInterest(ctx, interestID)
		if err != nil {
			continue
		}
		for _, teacher := range teachers {
			if teacher.UserID == requesterID {
				continue
			}
			if _, exists := results[teacher.UserID]; !exists {
				results[teacher.UserID] = &DiscoveryResult{
					UserID:          teacher.UserID,
					SharedInterests: []SharedInterestBrief{},
				}
			}
			results[teacher.UserID].SharedInterests = append(
				results[teacher.UserID].SharedInterests,
				SharedInterestBrief{
					InterestID:      teacher.InterestID,
					InterestName:    teacher.Name,
					Category:        teacher.Category,
					TeachLearnMatch: true,
				},
			)
		}
	}

	// Find potential learners for things requester wants to teach
	for interestID := range wantsToTeach {
		learners, err := s.interestRepo.GetLearnersForInterest(ctx, interestID)
		if err != nil {
			continue
		}
		for _, learner := range learners {
			if learner.UserID == requesterID {
				continue
			}
			if _, exists := results[learner.UserID]; !exists {
				results[learner.UserID] = &DiscoveryResult{
					UserID:          learner.UserID,
					SharedInterests: []SharedInterestBrief{},
				}
			}
			// Check if this interest already added
			alreadyAdded := false
			for _, si := range results[learner.UserID].SharedInterests {
				if si.InterestID == learner.InterestID {
					alreadyAdded = true
					break
				}
			}
			if !alreadyAdded {
				results[learner.UserID].SharedInterests = append(
					results[learner.UserID].SharedInterests,
					SharedInterestBrief{
						InterestID:      learner.InterestID,
						InterestName:    learner.Name,
						Category:        learner.Category,
						TeachLearnMatch: true,
					},
				)
			}
		}
	}

	// Convert to slice and enrich
	resultSlice := make([]DiscoveryResult, 0, len(results))
	for _, r := range results {
		// SECURITY: Skip blocked users
		if s.isBlocked(ctx, requesterID, r.UserID) {
			continue
		}

		// Get compatibility score
		if s.compatibilityRepo != nil {
			sharedAnswers, err := s.compatibilityRepo.GetSharedAnswers(ctx, requesterID, r.UserID)
			if err == nil {
				r.CompatibilityScore = s.calculateCompatibilityFromAnswers(sharedAnswers)
			}
		}

		// Get public profile
		if s.profileRepo != nil {
			profile, err := s.profileRepo.GetByUserID(ctx, r.UserID)
			if err == nil && profile != nil {
				r.Profile = profile.ToPublic()
			}
		}

		resultSlice = append(resultSlice, *r)
	}

	// Calculate scores and sort
	s.calculateMatchScores(resultSlice)
	sort.Slice(resultSlice, func(i, j int) bool {
		return resultSlice[i].MatchScore > resultSlice[j].MatchScore
	})

	// Apply limit
	if len(resultSlice) > limit {
		resultSlice = resultSlice[:limit]
	}

	return resultSlice, nil
}
