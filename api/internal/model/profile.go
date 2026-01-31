package model

import "time"

// UserProfile represents extended profile information for a user
type UserProfile struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Bio        *string    `json:"bio,omitempty"`
	Tagline    *string    `json:"tagline,omitempty"`
	Languages  []string   `json:"languages,omitempty"`
	Timezone   *string    `json:"timezone,omitempty"`
	Location   *Location  `json:"location,omitempty"`
	Visibility string     `json:"visibility"` // guilds, public, private
	LastActive *time.Time `json:"last_active,omitempty"`
	CreatedOn  time.Time  `json:"created_on"`
	UpdatedOn  time.Time  `json:"updated_on"`

	// Discovery eligibility tracking
	// User must answer 3+ questions from required categories (values, social, lifestyle, communication)
	DiscoveryEligible       bool     `json:"discovery_eligible"`
	CategoriesCompleted     []string `json:"categories_completed,omitempty"`
	QuestionCount           int      `json:"question_count"`
	ProfileCompletionScore  float64  `json:"profile_completion_score"`

	// Optional populated fields
	Username  *string `json:"username,omitempty"`
	Firstname *string `json:"firstname,omitempty"`
}

// ToPublic converts a UserProfile to its privacy-respecting public representation
func (p *UserProfile) ToPublic() *PublicProfile {
	pub := &PublicProfile{
		UserID:            p.UserID,
		Username:          p.Username,
		Firstname:         p.Firstname,
		Bio:               p.Bio,
		Tagline:           p.Tagline,
		Languages:         p.Languages,
		DiscoveryEligible: p.DiscoveryEligible,
	}

	if p.Location != nil {
		pub.City = p.Location.City
		pub.Country = p.Location.Country
	}

	pub.ActivityStatus = GetActivityStatus(p.LastActive)

	return pub
}

// Location stores geographic information with privacy controls
// IMPORTANT: lat/lng are stored internally but NEVER exposed to other users
type Location struct {
	// Internal only - never expose to other users via API
	Lat float64 `json:"-"`
	Lng float64 `json:"-"`

	// Public information - shown to other users
	City         string  `json:"city,omitempty"`
	Neighborhood *string `json:"neighborhood,omitempty"` // User-controlled, optional
	Country      string  `json:"country,omitempty"`
	CountryCode  string  `json:"country_code,omitempty"`
}

// LocationInternal is used internally for storing/retrieving from database
type LocationInternal struct {
	Lat          float64 `json:"lat"`
	Lng          float64 `json:"lng"`
	City         string  `json:"city,omitempty"`
	Neighborhood *string `json:"neighborhood,omitempty"`
	Country      string  `json:"country,omitempty"`
	CountryCode  string  `json:"country_code,omitempty"`
}

// ToPublic converts internal location to public representation (no coordinates)
func (l *LocationInternal) ToPublic() *Location {
	return &Location{
		City:         l.City,
		Neighborhood: l.Neighborhood,
		Country:      l.Country,
		CountryCode:  l.CountryCode,
	}
}

// ActivityStatus represents how recently a user was active
type ActivityStatus string

const (
	ActivityStatusNow       ActivityStatus = "active_now"       // < 10 minutes
	ActivityStatusRecently  ActivityStatus = "active_recently"  // < 30 minutes
	ActivityStatusThisHour  ActivityStatus = "active_this_hour" // < 1 hour
	ActivityStatusToday     ActivityStatus = "active_today"     // < 24 hours
	ActivityStatusYesterday ActivityStatus = "active_yesterday" // 24-48 hours
	ActivityStatusThisWeek  ActivityStatus = "active_this_week" // 2-7 days
	ActivityStatusAway      ActivityStatus = "away"             // > 7 days
)

// GetActivityStatus calculates activity status from last active time
func GetActivityStatus(lastActive *time.Time) ActivityStatus {
	if lastActive == nil {
		return ActivityStatusAway
	}

	since := time.Since(*lastActive)

	switch {
	case since < 10*time.Minute:
		return ActivityStatusNow
	case since < 30*time.Minute:
		return ActivityStatusRecently
	case since < time.Hour:
		return ActivityStatusThisHour
	case since < 24*time.Hour:
		return ActivityStatusToday
	case since < 48*time.Hour:
		return ActivityStatusYesterday
	case since < 7*24*time.Hour:
		return ActivityStatusThisWeek
	default:
		return ActivityStatusAway
	}
}

// DistanceBucket represents approximate distance (for privacy)
type DistanceBucket string

const (
	DistanceNearby   DistanceBucket = "nearby" // < 1 km
	Distance2km      DistanceBucket = "~2km"   // 1-2 km
	Distance5km      DistanceBucket = "~5km"   // 2-5 km
	Distance10km     DistanceBucket = "~10km"  // 5-10 km
	Distance20kmPlus DistanceBucket = ">20km"  // > 20 km
)

// FreshnessBucket represents activity recency tiers
type FreshnessBucket string

const (
	FreshnessBucketActiveNow      FreshnessBucket = "active_now"
	FreshnessBucketActiveRecently FreshnessBucket = "active_recently"
	FreshnessBucketActiveToday    FreshnessBucket = "active_today"
	FreshnessBucketActiveThisWeek FreshnessBucket = "active_this_week"
	FreshnessBucketAway           FreshnessBucket = "away"
)

// GetDistanceBucket converts exact distance to privacy-preserving bucket
func GetDistanceBucket(distanceKm float64) DistanceBucket {
	switch {
	case distanceKm < 1:
		return DistanceNearby
	case distanceKm < 2:
		return Distance2km
	case distanceKm < 5:
		return Distance5km
	case distanceKm < 10:
		return Distance10km
	default:
		return Distance20kmPlus
	}
}

// PublicProfile is what other users see (with privacy protections)
type PublicProfile struct {
	UserID            string          `json:"user_id"`
	Username          *string         `json:"username,omitempty"`
	Firstname         *string         `json:"firstname,omitempty"`
	Bio               *string         `json:"bio,omitempty"`
	Tagline           *string         `json:"tagline,omitempty"`
	Languages         []string        `json:"languages,omitempty"`
	City              string          `json:"city,omitempty"`
	Country           string          `json:"country,omitempty"`
	Distance          DistanceBucket  `json:"distance,omitempty"`       // Approximate only
	ActivityStatus    ActivityStatus  `json:"activity_status,omitempty"`
	Compatibility     *float64        `json:"compatibility,omitempty"`  // 0-100% if calculated
	DiscoveryEligible bool            `json:"discovery_eligible"`       // Eligible for discovery
}

// IsEligibleForDiscovery checks if a user profile meets discovery requirements
func (p *UserProfile) IsEligibleForDiscovery() bool {
	if p.QuestionCount < MinQuestionsForEligibility {
		return false
	}

	// Check all required categories are completed
	for _, required := range RequiredQuestionCategories {
		found := false
		for _, completed := range p.CategoriesCompleted {
			if completed == required {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// Visibility options
const (
	VisibilityGuilds  = "guilds"  // Only guild members can see
	VisibilityPublic  = "public"  // Anyone can see
	VisibilityPrivate = "private" // Only self can see
)

// Profile constraints
const (
	MaxBioLength     = 500
	MaxTaglineLength = 100
	MaxLanguages     = 10
)

// Required question categories for discovery eligibility
// User must answer at least 1 question from each category
var RequiredQuestionCategories = []string{
	"values",
	"social",
	"lifestyle",
	"communication",
}

// MinQuestionsForEligibility is the minimum total questions needed for discovery
const MinQuestionsForEligibility = 3

// CreateProfileRequest represents a request to create/update a profile
type CreateProfileRequest struct {
	Bio        *string          `json:"bio,omitempty"`
	Tagline    *string          `json:"tagline,omitempty"`
	Languages  []string         `json:"languages,omitempty"`
	Timezone   *string          `json:"timezone,omitempty"`
	Location   *LocationRequest `json:"location,omitempty"`
	Visibility *string          `json:"visibility,omitempty"`
}

// LocationRequest is used when user updates their location
type LocationRequest struct {
	Lat          float64 `json:"lat"`
	Lng          float64 `json:"lng"`
	City         string  `json:"city"`
	Neighborhood *string `json:"neighborhood,omitempty"`
	Country      string  `json:"country"`
	CountryCode  string  `json:"country_code"`
}

// UpdateProfileRequest represents a request to partially update a profile
type UpdateProfileRequest struct {
	Bio        *string          `json:"bio,omitempty"`
	Tagline    *string          `json:"tagline,omitempty"`
	Languages  []string         `json:"languages,omitempty"`
	Timezone   *string          `json:"timezone,omitempty"`
	Location   *LocationRequest `json:"location,omitempty"`
	Visibility *string          `json:"visibility,omitempty"`
}
