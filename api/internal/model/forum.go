package model

import "time"

// Forum represents a discussion forum attached to an Adventure or Event
type Forum struct {
	ID          string    `json:"id"`
	AdventureID *string   `json:"adventure_id,omitempty"` // Attached to adventure
	EventID     *string   `json:"event_id,omitempty"`     // Attached to event
	PostCount   int       `json:"post_count,omitempty"`   // Computed
	CreatedOn   time.Time `json:"created_on"`
}

// ForumPost represents a post in a forum
type ForumPost struct {
	ID        string              `json:"id"`
	ForumID   string              `json:"forum_id"`
	AuthorID  string              `json:"author_id"` // User ID
	Content   string              `json:"content"`
	ReplyToID *string             `json:"reply_to_id,omitempty"` // For threaded replies
	IsPinned  bool                `json:"is_pinned"`
	Reactions map[string][]string `json:"reactions,omitempty"` // emoji -> user_ids
	Mentions  []string            `json:"mentions,omitempty"`  // User IDs mentioned
	CreatedOn time.Time           `json:"created_on"`
	UpdatedOn time.Time           `json:"updated_on"`
	DeletedOn *time.Time          `json:"deleted_on,omitempty"` // Soft delete
}

// ForumPostWithAuthor includes author information
type ForumPostWithAuthor struct {
	Post       ForumPost   `json:"post"`
	Author     UserSummary `json:"author"`
	ReplyCount int         `json:"reply_count,omitempty"`
}

// UserSummary provides minimal user info for display
type UserSummary struct {
	ID        string  `json:"id"`
	Username  *string `json:"username,omitempty"`
	Firstname *string `json:"firstname,omitempty"`
	Lastname  *string `json:"lastname,omitempty"`
}

// ForumThread represents a post with all its replies
type ForumThread struct {
	RootPost ForumPostWithAuthor   `json:"root_post"`
	Replies  []ForumPostWithAuthor `json:"replies"`
}

// ForumWithPosts includes the forum with paginated posts
type ForumWithPosts struct {
	Forum      Forum                 `json:"forum"`
	Posts      []ForumPostWithAuthor `json:"posts"`
	TotalPosts int                   `json:"total_posts"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
}

// Constraints
const (
	MaxPostsPerForum     = 1000
	MaxPostContentLength = 5000
	MaxMentionsPerPost   = 20
	MaxReactionsPerPost  = 100
	DefaultForumPageSize = 50
)

// CreateForumPostRequest represents a request to create a forum post
type CreateForumPostRequest struct {
	Content   string   `json:"content"`
	ReplyToID *string  `json:"reply_to_id,omitempty"`
	Mentions  []string `json:"mentions,omitempty"` // User IDs to mention
}

// UpdateForumPostRequest represents a request to update a forum post
type UpdateForumPostRequest struct {
	Content  *string  `json:"content,omitempty"`
	Mentions []string `json:"mentions,omitempty"`
}

// ReactToPostRequest represents a request to react to a post
type ReactToPostRequest struct {
	Emoji string `json:"emoji"` // Single emoji
}

// PinPostRequest represents a request to pin/unpin a post
type PinPostRequest struct {
	Pinned bool `json:"pinned"`
}

// ForumSearchFilters for searching forum posts
type ForumSearchFilters struct {
	AuthorID   *string `json:"author_id,omitempty"`
	Query      *string `json:"query,omitempty"` // Text search
	PinnedOnly bool    `json:"pinned_only,omitempty"`
	RootOnly   bool    `json:"root_only,omitempty"` // Only root posts, no replies
}

// ForumPagination for paginated forum requests
type ForumPagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}
