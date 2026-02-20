package model

import "time"

// Interest represents a taggable interest/hobby/skill
type Interest struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"` // hobby, skill, language, sport, social, learning, outdoors
	Icon      *string   `json:"icon,omitempty"`
	CreatedOn time.Time `json:"created_on"`
}

// InterestCategory constants
const (
	InterestCategoryHobby    = "hobby"
	InterestCategorySkill    = "skill"
	InterestCategoryLanguage = "language"
	InterestCategorySport    = "sport"
	InterestCategorySocial   = "social"
	InterestCategoryLearning = "learning"
	InterestCategoryOutdoors = "outdoors"
	InterestCategoryCuisine  = "cuisine"
	InterestCategoryMusic    = "music"
	InterestCategoryArt      = "art"
	InterestCategoryTech     = "tech"
)

// InterestLevel represents proficiency/familiarity level
type InterestLevel string

const (
	InterestLevelCurious     InterestLevel = "curious"     // Just getting started
	InterestLevelInterested  InterestLevel = "interested"  // Actively interested
	InterestLevelExperienced InterestLevel = "experienced" // Significant experience
	InterestLevelExpert      InterestLevel = "expert"      // Can teach others
)

// UserInterest represents a user's relationship with an interest
type UserInterest struct {
	ID           string        `json:"id"` // Relation ID
	UserID       string        `json:"user_id"`
	InterestID   string        `json:"interest_id"`
	Level        InterestLevel `json:"level"`
	WantsToTeach bool          `json:"wants_to_teach"`
	WantsToLearn bool          `json:"wants_to_learn"`
	Intent       *string       `json:"intent,omitempty"` // "I want to organize hikes"
	CreatedOn    time.Time     `json:"created_on"`
	// Populated by repository joins
	Name     string  `json:"name,omitempty"`
	Category string  `json:"category,omitempty"`
	Icon     *string `json:"icon,omitempty"`
}

// UserInterestWithDetails includes the interest details
type UserInterestWithDetails struct {
	UserInterest
	Interest Interest `json:"interest"`
}

// TeachLearnMatch represents a potential teacher-learner pairing
type TeachLearnMatch struct {
	Interest   Interest `json:"interest"`
	TeacherID  string   `json:"teacher_id"`
	LearnerID  string   `json:"learner_id"`
	MatchScore float64  `json:"match_score,omitempty"` // Based on compatibility
}

// Interest constraints
const (
	MaxInterestsPerUser = 50
	MaxIntentLength     = 200
)

// AddInterestRequest represents a request to add an interest
type AddInterestRequest struct {
	InterestID   string  `json:"interest_id"`
	Level        string  `json:"level,omitempty"` // Default: interested
	WantsToTeach *bool   `json:"wants_to_teach,omitempty"`
	WantsToLearn *bool   `json:"wants_to_learn,omitempty"`
	Intent       *string `json:"intent,omitempty"`
}

// UpdateInterestRequest represents a request to update a user's interest
type UpdateInterestRequest struct {
	Level        *string `json:"level,omitempty"`
	WantsToTeach *bool   `json:"wants_to_teach,omitempty"`
	WantsToLearn *bool   `json:"wants_to_learn,omitempty"`
	Intent       *string `json:"intent,omitempty"`
}

// InterestFilter for searching/filtering interests
type InterestFilter struct {
	Categories []string `json:"categories,omitempty"`
	SearchTerm string   `json:"search,omitempty"`
}

// SharedInterests represents common interests between two users
type SharedInterests struct {
	UserAID    string                  `json:"user_a_id"`
	UserBID    string                  `json:"user_b_id"`
	Shared     []SharedInterestDetail  `json:"shared"`
	TeachLearn []TeachLearnOpportunity `json:"teach_learn,omitempty"`
}

// SharedInterestDetail shows how two users share an interest
type SharedInterestDetail struct {
	Interest    Interest      `json:"interest"`
	UserALevel  InterestLevel `json:"user_a_level"`
	UserBLevel  InterestLevel `json:"user_b_level"`
	UserAIntent *string       `json:"user_a_intent,omitempty"`
	UserBIntent *string       `json:"user_b_intent,omitempty"`
}

// TeachLearnOpportunity represents a teaching/learning match opportunity
type TeachLearnOpportunity struct {
	Interest  Interest `json:"interest"`
	TeacherID string   `json:"teacher_id"` // The one who wants to teach
	LearnerID string   `json:"learner_id"` // The one who wants to learn
}

// InterestMatch represents a potential match based on teaching/learning
type InterestMatch struct {
	UserID       string `json:"user_id"`
	InterestID   string `json:"interest_id"`
	InterestName string `json:"interest_name"`
	Category     string `json:"category"`
	Level        string `json:"level"`
	WantsToTeach bool   `json:"wants_to_teach"`
	WantsToLearn bool   `json:"wants_to_learn"`
}

// SharedInterestUser represents a user with shared interests
type SharedInterestUser struct {
	UserID          string   `json:"user_id"`
	SharedCount     int      `json:"shared_count"`
	SharedInterests []string `json:"shared_interests"` // Interest names
}

// InterestStats provides statistics about a user's interests
type InterestStats struct {
	TotalCount    int            `json:"total_count"`
	ByCategory    map[string]int `json:"by_category"`
	TeachingCount int            `json:"teaching_count"`
	LearningCount int            `json:"learning_count"`
}

// CategoryInfo provides display information for a category
type CategoryInfo struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Icon  string `json:"icon,omitempty"`
}

// GetInterestCategories returns all interest categories with display info
func GetInterestCategories() []CategoryInfo {
	return []CategoryInfo{
		{ID: InterestCategoryHobby, Label: "Hobbies", Icon: "paintpalette.fill"},
		{ID: InterestCategorySport, Label: "Sports & Fitness", Icon: "figure.run"},
		{ID: InterestCategorySocial, Label: "Social", Icon: "person.3.fill"},
		{ID: InterestCategoryLearning, Label: "Learning", Icon: "book.fill"},
		{ID: InterestCategoryOutdoors, Label: "Outdoors", Icon: "leaf.fill"},
		{ID: InterestCategorySkill, Label: "Skills", Icon: "hammer.fill"},
		{ID: InterestCategoryLanguage, Label: "Languages", Icon: "globe"},
		{ID: InterestCategoryCuisine, Label: "Food & Cuisine", Icon: "fork.knife"},
		{ID: InterestCategoryMusic, Label: "Music", Icon: "music.note"},
		{ID: InterestCategoryArt, Label: "Art & Design", Icon: "paintbrush.fill"},
		{ID: InterestCategoryTech, Label: "Technology", Icon: "desktopcomputer"},
	}
}
