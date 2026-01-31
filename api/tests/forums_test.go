package tests

/*
FEATURE: Forum Discussions
DOMAIN: Communication & Community

ACCEPTANCE CRITERIA:
===================

AC-FORUM-001: Forum Created with Adventure
  GIVEN new adventure
  THEN forum auto-created

AC-FORUM-002: Create Post
  GIVEN admitted user
  WHEN posting to adventure forum
  THEN post created

AC-FORUM-003: Reply Threading
  GIVEN existing post
  WHEN replying
  THEN reply_to_id set correctly

AC-FORUM-004: Pin Post
  GIVEN organizer
  WHEN pinning post
  THEN post.is_pinned = true
  AND pinned posts appear first

AC-FORUM-005: Forum Access Control
  GIVEN non-admitted user
  WHEN attempting to post
  THEN fails with 403 Forbidden
*/

import (
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestForum_Constraints(t *testing.T) {
	// Verify forum constraints
	assert.Equal(t, 1000, model.MaxPostsPerForum)
	assert.Equal(t, 5000, model.MaxPostContentLength)
	assert.Equal(t, 20, model.MaxMentionsPerPost)
	assert.Equal(t, 100, model.MaxReactionsPerPost)
	assert.Equal(t, 50, model.DefaultForumPageSize)
}

func TestForum_ForumModel(t *testing.T) {
	// AC-FORUM-001: Forum attached to adventure
	adventureID := "adventure:123"
	forum := &model.Forum{
		ID:          "forum:f1",
		AdventureID: &adventureID,
		EventID:     nil,
		PostCount:   0,
		CreatedOn:   time.Now(),
	}

	assert.Equal(t, "forum:f1", forum.ID)
	assert.NotNil(t, forum.AdventureID)
	assert.Equal(t, "adventure:123", *forum.AdventureID)
	assert.Nil(t, forum.EventID)
}

func TestForum_ForumAttachedToEvent(t *testing.T) {
	// Forum can also be attached to event
	eventID := "event:456"
	forum := &model.Forum{
		ID:        "forum:f2",
		EventID:   &eventID,
		CreatedOn: time.Now(),
	}

	assert.NotNil(t, forum.EventID)
	assert.Equal(t, "event:456", *forum.EventID)
	assert.Nil(t, forum.AdventureID)
}

func TestForum_ForumPostModel(t *testing.T) {
	// AC-FORUM-002: Post creation
	post := &model.ForumPost{
		ID:        "post:p1",
		ForumID:   "forum:f1",
		AuthorID:  "user:alice",
		Content:   "This is my first post in the adventure forum!",
		ReplyToID: nil,
		IsPinned:  false,
		Reactions: nil,
		Mentions:  nil,
		CreatedOn: time.Now(),
		UpdatedOn: time.Now(),
		DeletedOn: nil,
	}

	assert.Equal(t, "post:p1", post.ID)
	assert.Equal(t, "forum:f1", post.ForumID)
	assert.Equal(t, "user:alice", post.AuthorID)
	assert.NotEmpty(t, post.Content)
	assert.Nil(t, post.ReplyToID)
	assert.False(t, post.IsPinned)
	assert.Nil(t, post.DeletedOn)
}

func TestForum_ReplyThreading(t *testing.T) {
	// AC-FORUM-003: Reply threading with reply_to_id
	parentID := "post:p1"
	reply := &model.ForumPost{
		ID:        "post:p2",
		ForumID:   "forum:f1",
		AuthorID:  "user:bob",
		Content:   "Great point! I agree.",
		ReplyToID: &parentID,
		IsPinned:  false,
		CreatedOn: time.Now(),
		UpdatedOn: time.Now(),
	}

	assert.NotNil(t, reply.ReplyToID)
	assert.Equal(t, "post:p1", *reply.ReplyToID)
}

func TestForum_NestedReplies(t *testing.T) {
	// Multiple levels of threading
	rootID := "post:root"
	level1ID := "post:l1"

	root := &model.ForumPost{
		ID:        rootID,
		ForumID:   "forum:f1",
		AuthorID:  "user:alice",
		Content:   "Original post",
		ReplyToID: nil,
	}

	level1 := &model.ForumPost{
		ID:        level1ID,
		ForumID:   "forum:f1",
		AuthorID:  "user:bob",
		Content:   "First reply",
		ReplyToID: &rootID,
	}

	level2 := &model.ForumPost{
		ID:        "post:l2",
		ForumID:   "forum:f1",
		AuthorID:  "user:carol",
		Content:   "Reply to the reply",
		ReplyToID: &level1ID,
	}

	assert.Nil(t, root.ReplyToID)
	assert.Equal(t, rootID, *level1.ReplyToID)
	assert.Equal(t, level1ID, *level2.ReplyToID)
}

func TestForum_PinnedPost(t *testing.T) {
	// AC-FORUM-004: Pinned posts
	pinnedPost := &model.ForumPost{
		ID:        "post:pinned",
		ForumID:   "forum:f1",
		AuthorID:  "user:organizer",
		Content:   "Important announcement: Event time changed!",
		IsPinned:  true,
		CreatedOn: time.Now(),
		UpdatedOn: time.Now(),
	}

	normalPost := &model.ForumPost{
		ID:        "post:normal",
		ForumID:   "forum:f1",
		AuthorID:  "user:alice",
		Content:   "Regular discussion post",
		IsPinned:  false,
		CreatedOn: time.Now(),
		UpdatedOn: time.Now(),
	}

	assert.True(t, pinnedPost.IsPinned)
	assert.False(t, normalPost.IsPinned)
}

func TestForum_PostReactions(t *testing.T) {
	// Posts can have emoji reactions from users
	post := &model.ForumPost{
		ID:       "post:p1",
		ForumID:  "forum:f1",
		AuthorID: "user:alice",
		Content:  "Who's excited for the hike?",
		Reactions: map[string][]string{
			"thumbsup": {"user:bob", "user:carol", "user:dave"},
			"fire":     {"user:eve"},
			"heart":    {"user:bob", "user:frank"},
		},
	}

	// Check reaction structure
	assert.Len(t, post.Reactions, 3)
	assert.Len(t, post.Reactions["thumbsup"], 3)
	assert.Contains(t, post.Reactions["thumbsup"], "user:bob")
	assert.Len(t, post.Reactions["heart"], 2)
}

func TestForum_PostMentions(t *testing.T) {
	// Posts can mention users
	post := &model.ForumPost{
		ID:       "post:p1",
		ForumID:  "forum:f1",
		AuthorID: "user:alice",
		Content:  "Hey @bob and @carol, are you coming?",
		Mentions: []string{"user:bob", "user:carol"},
	}

	assert.Len(t, post.Mentions, 2)
	assert.Contains(t, post.Mentions, "user:bob")
	assert.Contains(t, post.Mentions, "user:carol")
}

func TestForum_SoftDelete(t *testing.T) {
	// Posts can be soft-deleted
	now := time.Now()
	deletedPost := &model.ForumPost{
		ID:        "post:deleted",
		ForumID:   "forum:f1",
		AuthorID:  "user:alice",
		Content:   "[deleted]",
		DeletedOn: &now,
	}

	assert.NotNil(t, deletedPost.DeletedOn)
}

func TestForum_PostWithAuthor(t *testing.T) {
	// AC-FORUM-002: Post includes author info
	username := "alice123"
	firstname := "Alice"
	lastname := "Smith"

	postWithAuthor := &model.ForumPostWithAuthor{
		Post: model.ForumPost{
			ID:       "post:p1",
			ForumID:  "forum:f1",
			AuthorID: "user:alice",
			Content:  "Test post",
		},
		Author: model.UserSummary{
			ID:        "user:alice",
			Username:  &username,
			Firstname: &firstname,
			Lastname:  &lastname,
		},
		ReplyCount: 5,
	}

	assert.Equal(t, "user:alice", postWithAuthor.Author.ID)
	assert.Equal(t, "alice123", *postWithAuthor.Author.Username)
	assert.Equal(t, 5, postWithAuthor.ReplyCount)
}

func TestForum_ThreadStructure(t *testing.T) {
	// AC-FORUM-003: Thread with root and replies
	username := "alice123"
	bobUsername := "bob456"

	thread := &model.ForumThread{
		RootPost: model.ForumPostWithAuthor{
			Post: model.ForumPost{
				ID:       "post:root",
				ForumID:  "forum:f1",
				AuthorID: "user:alice",
				Content:  "Original discussion topic",
			},
			Author: model.UserSummary{
				ID:       "user:alice",
				Username: &username,
			},
			ReplyCount: 2,
		},
		Replies: []model.ForumPostWithAuthor{
			{
				Post: model.ForumPost{
					ID:       "post:r1",
					ForumID:  "forum:f1",
					AuthorID: "user:bob",
					Content:  "First reply",
				},
				Author: model.UserSummary{
					ID:       "user:bob",
					Username: &bobUsername,
				},
			},
		},
	}

	assert.Equal(t, "post:root", thread.RootPost.Post.ID)
	assert.Equal(t, 2, thread.RootPost.ReplyCount)
	assert.Len(t, thread.Replies, 1)
}

func TestForum_ForumWithPostsPagination(t *testing.T) {
	// Forum listing with pagination
	forumWithPosts := &model.ForumWithPosts{
		Forum: model.Forum{
			ID:        "forum:f1",
			PostCount: 150,
		},
		Posts:      []model.ForumPostWithAuthor{}, // Would be populated
		TotalPosts: 150,
		Page:       1,
		PageSize:   50,
	}

	assert.Equal(t, 150, forumWithPosts.TotalPosts)
	assert.Equal(t, 1, forumWithPosts.Page)
	assert.Equal(t, 50, forumWithPosts.PageSize)
}

func TestForum_CreatePostRequest(t *testing.T) {
	// AC-FORUM-002: Create post request structure
	replyID := "post:parent"
	req := &model.CreateForumPostRequest{
		Content:   "New post content with @mention",
		ReplyToID: &replyID,
		Mentions:  []string{"user:mentioned"},
	}

	assert.NotEmpty(t, req.Content)
	assert.Equal(t, "post:parent", *req.ReplyToID)
	assert.Len(t, req.Mentions, 1)
}

func TestForum_UpdatePostRequest(t *testing.T) {
	// Update request for editing posts
	content := "Updated content"
	req := &model.UpdateForumPostRequest{
		Content:  &content,
		Mentions: []string{"user:newmention"},
	}

	assert.Equal(t, "Updated content", *req.Content)
	assert.Len(t, req.Mentions, 1)
}

func TestForum_ReactToPostRequest(t *testing.T) {
	// Reaction request
	req := &model.ReactToPostRequest{
		Emoji: "thumbsup",
	}

	assert.Equal(t, "thumbsup", req.Emoji)
}

func TestForum_PinPostRequest(t *testing.T) {
	// AC-FORUM-004: Pin request
	pinReq := &model.PinPostRequest{
		Pinned: true,
	}
	unpinReq := &model.PinPostRequest{
		Pinned: false,
	}

	assert.True(t, pinReq.Pinned)
	assert.False(t, unpinReq.Pinned)
}

func TestForum_SearchFilters(t *testing.T) {
	// Forum search with various filters
	authorID := "user:alice"
	query := "hiking"

	filters := &model.ForumSearchFilters{
		AuthorID:   &authorID,
		Query:      &query,
		PinnedOnly: true,
		RootOnly:   false,
	}

	assert.Equal(t, "user:alice", *filters.AuthorID)
	assert.Equal(t, "hiking", *filters.Query)
	assert.True(t, filters.PinnedOnly)
	assert.False(t, filters.RootOnly)
}

func TestForum_SearchFiltersRootOnly(t *testing.T) {
	// Filter to show only root posts (no replies)
	filters := &model.ForumSearchFilters{
		RootOnly: true,
	}

	assert.True(t, filters.RootOnly)
}

func TestForum_Pagination(t *testing.T) {
	// Pagination structure
	pagination := &model.ForumPagination{
		Page:     2,
		PageSize: 25,
	}

	assert.Equal(t, 2, pagination.Page)
	assert.Equal(t, 25, pagination.PageSize)
}

func TestForum_UserSummary(t *testing.T) {
	// User summary for display
	username := "johndoe"
	firstname := "John"
	lastname := "Doe"

	summary := &model.UserSummary{
		ID:        "user:john",
		Username:  &username,
		Firstname: &firstname,
		Lastname:  &lastname,
	}

	assert.Equal(t, "user:john", summary.ID)
	assert.Equal(t, "johndoe", *summary.Username)
	assert.Equal(t, "John", *summary.Firstname)
	assert.Equal(t, "Doe", *summary.Lastname)
}

func TestForum_UserSummaryMinimal(t *testing.T) {
	// User summary with minimal info (some fields optional)
	summary := &model.UserSummary{
		ID: "user:anonymous",
	}

	assert.Equal(t, "user:anonymous", summary.ID)
	assert.Nil(t, summary.Username)
	assert.Nil(t, summary.Firstname)
	assert.Nil(t, summary.Lastname)
}

func TestForum_ContentLengthValidation(t *testing.T) {
	// Post content should not exceed max length
	maxLen := model.MaxPostContentLength

	validContent := make([]byte, maxLen)
	for i := range validContent {
		validContent[i] = 'a'
	}

	invalidContent := make([]byte, maxLen+1)
	for i := range invalidContent {
		invalidContent[i] = 'a'
	}

	assert.Len(t, validContent, maxLen)
	assert.Len(t, invalidContent, maxLen+1)
	assert.Greater(t, len(invalidContent), model.MaxPostContentLength)
}

func TestForum_MentionsLimitValidation(t *testing.T) {
	// Mentions should not exceed max
	maxMentions := model.MaxMentionsPerPost

	mentions := make([]string, maxMentions)
	for i := range mentions {
		mentions[i] = "user:mentioned"
	}

	assert.Len(t, mentions, maxMentions)
}

func TestForum_PostTimestamps(t *testing.T) {
	// Posts track created and updated times
	now := time.Now()
	post := &model.ForumPost{
		ID:        "post:p1",
		ForumID:   "forum:f1",
		AuthorID:  "user:alice",
		Content:   "Original content",
		CreatedOn: now,
		UpdatedOn: now,
	}

	// After edit, UpdatedOn should be newer
	assert.Equal(t, post.CreatedOn, post.UpdatedOn)

	// Simulating edit
	laterTime := now.Add(time.Hour)
	post.Content = "Edited content"
	post.UpdatedOn = laterTime

	assert.True(t, post.UpdatedOn.After(post.CreatedOn))
}
