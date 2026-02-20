package model

import "time"

// Review represents feedback between users (tag-based, not star ratings)
type Review struct {
	ID              string    `json:"id"`
	ReviewerID      string    `json:"reviewer_id"`
	RevieweeID      string    `json:"reviewee_id"`
	Context         string    `json:"context"`                // hosted, was_guest, event, matched, hangout
	ReferenceID     *string   `json:"reference_id,omitempty"` // event:xyz, hangout:abc
	WouldMeetAgain  bool      `json:"would_meet_again"`
	PositiveTags    []string  `json:"positive_tags,omitempty"`
	ImprovementTags []string  `json:"improvement_tags,omitempty"`
	PrivateNote     *string   `json:"private_note,omitempty"` // Only visible to reviewee
	CreatedOn       time.Time `json:"created_on"`
}

// ReviewContext constants
const (
	ReviewContextHosted   = "hosted"    // You hosted an event
	ReviewContextWasGuest = "was_guest" // You attended someone's event
	ReviewContextEvent    = "event"     // General event attendance
	ReviewContextMatched  = "matched"   // Pool/Donut match
	ReviewContextHangout  = "hangout"   // 1:1 or small group hangout
)

// PositiveTag constants - things that went well
const (
	TagWellOrganized      = "well_organized"
	TagInclusive          = "inclusive"
	TagGoodConversation   = "good_conversation"
	TagGoodVenue          = "good_venue"
	TagRightGroupSize     = "right_group_size"
	TagOnTime             = "on_time"
	TagGoodFood           = "good_food"
	TagMadeConnections    = "made_connections"
	TagWelcomingNewcomers = "welcoming_newcomers"
	TagGoodListener       = "good_listener"
	TagReliable           = "reliable"
	TagBringsSnacks       = "brings_snacks"
	TagGreatHost          = "great_host"
	TagInterestingPerson  = "interesting_person"
	TagFunEnergy          = "fun_energy"
)

// ImprovementTag constants - suggestions for improvement
const (
	TagMoreStructure       = "more_structure"
	TagBetterVenue         = "better_venue"
	TagDifferentTime       = "different_time"
	TagSmallerGroup        = "smaller_group"
	TagLargerGroup         = "larger_group"
	TagMoreLeadTime        = "more_lead_time"
	TagClearerExpectations = "clearer_expectations"
	TagDietaryOptions      = "dietary_options"
	TagBetterCommunication = "better_communication"
	TagMoreActivities      = "more_activities"
)

// TagInfo provides display information for tags
type TagInfo struct {
	Tag   string `json:"tag"`
	Label string `json:"label"`
	Icon  string `json:"icon,omitempty"`
}

// GetPositiveTags returns all positive tags with display info
func GetPositiveTags() []TagInfo {
	return []TagInfo{
		{Tag: TagWellOrganized, Label: "Well organized", Icon: "checkmark.circle.fill"},
		{Tag: TagInclusive, Label: "Inclusive atmosphere", Icon: "person.3.fill"},
		{Tag: TagGoodConversation, Label: "Great conversation", Icon: "bubble.left.and.bubble.right.fill"},
		{Tag: TagGoodVenue, Label: "Good venue/location", Icon: "mappin.circle.fill"},
		{Tag: TagRightGroupSize, Label: "Right group size", Icon: "person.2.fill"},
		{Tag: TagOnTime, Label: "Started/ended on time", Icon: "clock.fill"},
		{Tag: TagGoodFood, Label: "Good food/drinks", Icon: "fork.knife"},
		{Tag: TagMadeConnections, Label: "Made new connections", Icon: "link"},
		{Tag: TagWelcomingNewcomers, Label: "Welcomes newcomers", Icon: "hand.wave.fill"},
		{Tag: TagGoodListener, Label: "Good listener", Icon: "ear.fill"},
		{Tag: TagReliable, Label: "Reliable", Icon: "checkmark.seal.fill"},
		{Tag: TagBringsSnacks, Label: "Brings snacks", Icon: "takeoutbag.and.cup.and.straw.fill"},
		{Tag: TagGreatHost, Label: "Great host", Icon: "star.fill"},
		{Tag: TagInterestingPerson, Label: "Interesting person", Icon: "sparkles"},
		{Tag: TagFunEnergy, Label: "Fun energy", Icon: "bolt.fill"},
	}
}

// GetImprovementTags returns all improvement tags with display info
func GetImprovementTags() []TagInfo {
	return []TagInfo{
		{Tag: TagMoreStructure, Label: "More structured activities"},
		{Tag: TagBetterVenue, Label: "Better venue"},
		{Tag: TagDifferentTime, Label: "Different time"},
		{Tag: TagSmallerGroup, Label: "Smaller group"},
		{Tag: TagLargerGroup, Label: "Larger group"},
		{Tag: TagMoreLeadTime, Label: "More lead time for planning"},
		{Tag: TagClearerExpectations, Label: "Clearer expectations"},
		{Tag: TagDietaryOptions, Label: "Dietary options"},
		{Tag: TagBetterCommunication, Label: "Better communication"},
		{Tag: TagMoreActivities, Label: "More activities"},
	}
}

// Reputation represents aggregated feedback for a user
type Reputation struct {
	UserID            string     `json:"user_id"`
	TotalReviews      int        `json:"total_reviews"`
	WouldMeetAgain    int        `json:"would_meet_again"`     // Count of positive
	WouldMeetAgainPct float64    `json:"would_meet_again_pct"` // Percentage
	TopPositiveTags   []TagCount `json:"top_positive_tags"`
	EventsHosted      int        `json:"events_hosted"`
	EventsAttended    int        `json:"events_attended"`
}

// TagCount represents a tag with its count
type TagCount struct {
	Tag   string `json:"tag"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

// ReputationDisplay is what's shown on profiles (not star ratings!)
type ReputationDisplay struct {
	EventsHosted     int        `json:"events_hosted"`
	TopTags          []TagCount `json:"top_tags"`           // Top 5 positive tags with counts
	WouldReturnRatio string     `json:"would_return_ratio"` // "11 of 12 attendees would return"
}

// Review constraints
const (
	MaxPrivateNoteLength = 500
	MaxTagsPerReview     = 8
)

// CreateReviewRequest represents a request to leave feedback
type CreateReviewRequest struct {
	RevieweeID      string   `json:"reviewee_id"`
	Context         string   `json:"context"`
	ReferenceID     *string  `json:"reference_id,omitempty"`
	WouldMeetAgain  bool     `json:"would_meet_again"`
	PositiveTags    []string `json:"positive_tags,omitempty"`
	ImprovementTags []string `json:"improvement_tags,omitempty"`
	PrivateNote     *string  `json:"private_note,omitempty"`
}

// EventFeedbackFlow represents the post-event feedback questions
type EventFeedbackFlow struct {
	EventID  string `json:"event_id"`
	Attended string `json:"attended"` // yes, no, partial
	// Then positive tags, improvement tags, would_attend_again, optional private note
}

// PostEventFeedbackRequest is the full feedback submission
type PostEventFeedbackRequest struct {
	EventID          string   `json:"event_id"`
	Attended         string   `json:"attended"` // yes, no, partial
	PositiveTags     []string `json:"positive_tags,omitempty"`
	ImprovementTags  []string `json:"improvement_tags,omitempty"`
	WouldAttendAgain string   `json:"would_attend_again"` // yes, maybe, no
	PrivateNote      *string  `json:"private_note,omitempty"`
}
