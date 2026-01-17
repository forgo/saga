package model

import "time"

// Question represents a matching question
type Question struct {
	ID                    string           `json:"id"`
	Text                  string           `json:"text"`
	Category              string           `json:"category"` // values, social, lifestyle, communication
	Options               []QuestionOption `json:"options"`
	IsDealBreakerEligible bool             `json:"is_dealbreaker_eligible"`
	SortOrder             int              `json:"sort_order"`
	Active                bool             `json:"active"`
	// Circle-specific question (nil = global question)
	CircleID              *string          `json:"circle_id,omitempty"`
	CreatedBy             *string          `json:"created_by,omitempty"` // For circle questions
	CreatedOn             time.Time        `json:"created_on"`
}

// QuestionOption represents an answer option for a question
type QuestionOption struct {
	Value string `json:"value"` // Machine-readable value
	Label string `json:"label"` // Human-readable label
	// ImplicitBias: internal flag (not shown to users) indicating this option
	// reflects less community-oriented, more insular/isolationist thinking.
	// Used to track users who accumulate many such answers.
	ImplicitBias float64 `json:"-"` // -1.0 (antisocial) to +1.0 (prosocial), 0 = neutral
}

// QuestionCategory constants
const (
	QuestionCategoryValues        = "values"
	QuestionCategorySocial        = "social"
	QuestionCategoryLifestyle     = "lifestyle"
	QuestionCategoryCommunication = "communication"
)

// QuestionCategoryInfo provides display information for a category
type QuestionCategoryInfo struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Icon  string `json:"icon,omitempty"`
}

// GetQuestionCategories returns all question categories with display info
func GetQuestionCategories() []QuestionCategoryInfo {
	return []QuestionCategoryInfo{
		{ID: QuestionCategoryValues, Label: "Values & Ethics", Icon: "heart.fill"},
		{ID: QuestionCategorySocial, Label: "Social Style", Icon: "person.2.fill"},
		{ID: QuestionCategoryLifestyle, Label: "Lifestyle", Icon: "house.fill"},
		{ID: QuestionCategoryCommunication, Label: "Communication", Icon: "bubble.left.and.bubble.right.fill"},
	}
}

// Answer represents a user's answer to a question
type Answer struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	QuestionID        string    `json:"question_id"`
	SelectedOption    string    `json:"selected_option"`    // User's own answer
	AcceptableOptions []string  `json:"acceptable_options"` // What they accept from others
	Importance        string    `json:"importance"`         // How much this matters
	IsDealBreaker     bool      `json:"is_dealbreaker"`     // Hard requirement
	// AlignmentWeight: how much you want others to answer similarly (0-1, default 0.5)
	// 0 = don't care if they match, 1 = strongly prefer they match my answer
	AlignmentWeight float64 `json:"alignment_weight"`
	// YikesOptions: options that are red flags if the other person selects them
	// Different from acceptable_options: these trigger a warning/lower score
	YikesOptions []string  `json:"yikes_options,omitempty"`
	CreatedOn    time.Time `json:"created_on"`
	UpdatedOn    time.Time `json:"updated_on"`
}

// AnswerWithQuestion includes the question details
type AnswerWithQuestion struct {
	Answer   Answer   `json:"answer"`
	Question Question `json:"question"`
}

// Importance levels for compatibility weighting
const (
	ImportanceIrrelevant = "irrelevant" // Weight: 0
	ImportanceLittle     = "little"     // Weight: 1
	ImportanceSomewhat   = "somewhat"   // Weight: 10
	ImportanceVery       = "very"       // Weight: 50
	ImportanceMandatory  = "mandatory"  // Weight: 250
)

// ImportanceWeight returns the numeric weight for an importance level
func ImportanceWeight(importance string) int {
	switch importance {
	case ImportanceIrrelevant:
		return 0
	case ImportanceLittle:
		return 1
	case ImportanceSomewhat:
		return 10
	case ImportanceVery:
		return 50
	case ImportanceMandatory:
		return 250
	default:
		return 10 // Default to somewhat
	}
}

// CompatibilityScore represents the match score between two users
type CompatibilityScore struct {
	UserAID     string  `json:"user_a_id"`
	UserBID     string  `json:"user_b_id"`
	Score       float64 `json:"score"`        // 0-100
	AToB        float64 `json:"a_to_b_score"` // How well B matches A's preferences
	BToA        float64 `json:"b_to_a_score"` // How well A matches B's preferences
	SharedCount int     `json:"shared_count"` // Number of shared questions
	DealBreaker bool    `json:"deal_breaker"` // True if any dealbreaker violated
}

// CompatibilityBreakdown provides detailed scoring info
type CompatibilityBreakdown struct {
	CompatibilityScore
	CategoryScores map[string]float64     `json:"category_scores,omitempty"` // Per-category breakdown
	DealBreakers   []DealBreakerViolation `json:"deal_breakers,omitempty"`   // Which ones violated
}

// DealBreakerViolation represents a dealbreaker that was violated
type DealBreakerViolation struct {
	QuestionID    string `json:"question_id"`
	QuestionText  string `json:"question_text"`
	UserAnswer    string `json:"user_answer"`    // What the user requires
	PartnerAnswer string `json:"partner_answer"` // What the other person answered
}

// Question constraints
const (
	MinQuestionsForDiscovery = 3  // Must answer at least 3 questions
	MaxQuestionsToDisplay    = 50 // Limit questions shown at once
)

// Alignment weight defaults
const (
	DefaultAlignmentWeight = 0.5 // Neutral: some preference for matching
	MinAlignmentWeight     = 0.0 // Don't care if they match
	MaxAlignmentWeight     = 1.0 // Strongly prefer they match
)

// Implicit bias thresholds
const (
	// BiasThresholdWarning: accumulated bias score that triggers a gentle nudge
	BiasThresholdWarning = -5.0
	// BiasThresholdConcern: accumulated bias score that suggests community mismatch
	BiasThresholdConcern = -10.0
)

// Required categories - user must answer at least one from each
var RequiredCategories = []string{
	QuestionCategoryValues,
	QuestionCategorySocial,
}

// AnswerQuestionRequest represents a request to answer a question
type AnswerQuestionRequest struct {
	SelectedOption    string   `json:"selected_option"`
	AcceptableOptions []string `json:"acceptable_options,omitempty"` // Default: all options
	Importance        string   `json:"importance,omitempty"`         // Default: somewhat
	IsDealBreaker     bool     `json:"is_dealbreaker,omitempty"`
	AlignmentWeight   *float64 `json:"alignment_weight,omitempty"` // Default: 0.5
	YikesOptions      []string `json:"yikes_options,omitempty"`    // Red flag answers
}

// UpdateAnswerRequest represents a request to update an answer
type UpdateAnswerRequest struct {
	SelectedOption    *string  `json:"selected_option,omitempty"`
	AcceptableOptions []string `json:"acceptable_options,omitempty"`
	Importance        *string  `json:"importance,omitempty"`
	IsDealBreaker     *bool    `json:"is_dealbreaker,omitempty"`
	AlignmentWeight   *float64 `json:"alignment_weight,omitempty"`
	YikesOptions      []string `json:"yikes_options,omitempty"`
}

// QuestionProgress tracks user's progress in answering questions
type QuestionProgress struct {
	TotalQuestions     int            `json:"total_questions"`
	AnsweredCount      int            `json:"answered_count"`
	ByCategory         map[string]int `json:"by_category"`         // Answered per category
	RequiredCategories []string       `json:"required_categories"` // Which required categories are missing
	CanDiscover        bool           `json:"can_discover"`        // Has met minimum requirements
}

// UserBiasProfile tracks accumulated implicit bias from questionnaire answers
// This is INTERNAL only - never exposed to users or other users
type UserBiasProfile struct {
	UserID string `json:"user_id"`
	// AccumulatedBias: sum of ImplicitBias values from selected answers
	// Negative values indicate more antisocial/insular tendencies
	AccumulatedBias float64 `json:"-"`
	// AnswerCount: number of answers contributing to the bias score
	AnswerCount int `json:"-"`
	// AverageBias: AccumulatedBias / AnswerCount (normalized score)
	AverageBias float64 `json:"-"`
	// Status: normal, warning, concern (based on thresholds)
	Status string `json:"-"`
	// LastUpdated: when bias was last recalculated
	LastUpdated time.Time `json:"-"`
}

// BiasStatus constants
const (
	BiasStatusNormal  = "normal"  // No concerns
	BiasStatusWarning = "warning" // Gentle nudge toward community values
	BiasStatusConcern = "concern" // May not be a good fit for this community
)

// CircleValues represents circle-specific value questions
// Circles can define their own questions to screen members
type CircleValues struct {
	ID          string    `json:"id"`
	CircleID    string    `json:"circle_id"`
	Name        string    `json:"name"`        // e.g., "Photography Club Values"
	Description *string   `json:"description"` // What this questionnaire is about
	Questions   []string  `json:"questions"`   // IDs of questions in this set
	Required    bool      `json:"required"`    // Must complete to join/remain in circle
	CreatedBy   string    `json:"created_by"`  // Circle admin who created it
	CreatedOn   time.Time `json:"created_on"`
	UpdatedOn   time.Time `json:"updated_on"`
}

// CircleValuesResponse tracks a member's responses to circle values
type CircleValuesResponse struct {
	ID            string    `json:"id"`
	CircleID      string    `json:"circle_id"`
	UserID        string    `json:"user_id"`
	ValuesID      string    `json:"values_id"` // Which CircleValues set
	Answers       []Answer  `json:"answers"`
	CompletedOn   time.Time `json:"completed_on"`
	BiasScore     float64   `json:"-"` // Internal: accumulated bias for circle questions
	ApprovedBy    *string   `json:"approved_by,omitempty"`    // Admin who approved (if manual review)
	ApprovedOn    *time.Time `json:"approved_on,omitempty"`
	RejectedBy    *string   `json:"rejected_by,omitempty"`    // Admin who rejected (if manual review)
	RejectionNote *string   `json:"rejection_note,omitempty"` // Private note to user
}

// CircleValuesStatus constants
const (
	CircleValuesStatusPending  = "pending"  // Awaiting admin review
	CircleValuesStatusApproved = "approved" // Can join/stay in circle
	CircleValuesStatusRejected = "rejected" // Values mismatch
)

// CreateCircleValuesRequest represents a request to create circle values
type CreateCircleValuesRequest struct {
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	Questions   []string `json:"questions"` // Question IDs to include
	Required    bool     `json:"required"`
}

// CreateCircleQuestionRequest represents a request to create a circle-specific question
type CreateCircleQuestionRequest struct {
	Text                  string           `json:"text"`
	Category              string           `json:"category"`
	Options               []QuestionOption `json:"options"`
	IsDealBreakerEligible bool             `json:"is_dealbreaker_eligible"`
}

// YikesSummary provides info about red flags in a compatibility match
type YikesSummary struct {
	HasYikes    bool     `json:"has_yikes"`
	YikesCount  int      `json:"yikes_count"`
	Categories  []string `json:"categories,omitempty"`  // Which categories have yikes
	Severity    string   `json:"severity,omitempty"`    // mild, moderate, severe
}

// YikesSeverity constants
const (
	YikesSeverityNone     = ""
	YikesSeverityMild     = "mild"     // 1-2 yikes
	YikesSeverityModerate = "moderate" // 3-4 yikes
	YikesSeveritySevere   = "severe"   // 5+ yikes
)

// GetYikesSeverity returns severity based on count
func GetYikesSeverity(count int) string {
	switch {
	case count == 0:
		return YikesSeverityNone
	case count <= 2:
		return YikesSeverityMild
	case count <= 4:
		return YikesSeverityModerate
	default:
		return YikesSeveritySevere
	}
}
