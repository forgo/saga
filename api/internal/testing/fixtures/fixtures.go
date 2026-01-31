// Package fixtures provides test data factories for e2e testing.
//
// Each factory method creates entities with sensible defaults while allowing
// customization via option functions. Factories handle database insertion
// and return fully populated models.
//
// Usage:
//
//	f := fixtures.New(tdb.DB)
//	user := f.CreateUser(t)
//	guild := f.CreateGuild(t, user)
//	event := f.CreateEvent(t, guild, user)
package fixtures

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"

	"golang.org/x/crypto/bcrypt"
)

// Factory creates test entities in the database
type Factory struct {
	db database.Database
}

// New creates a new fixture factory
func New(db database.Database) *Factory {
	return &Factory{db: db}
}

// randomID generates a random hex ID
func randomID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ctx returns a context with timeout
func ctx() context.Context {
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// Store cancel to prevent leak warning
	_ = cancel
	return c
}

// ============================================================================
// User Fixtures
// ============================================================================

// UserOpts customizes user creation
type UserOpts struct {
	Email         string
	Username      string
	Password      string
	Role          model.UserRole
	EmailVerified bool
}

// CreateUser creates a user with optional customizations
func (f *Factory) CreateUser(t *testing.T, opts ...func(*UserOpts)) *model.User {
	t.Helper()

	o := &UserOpts{
		Email:         fmt.Sprintf("user_%s@test.local", randomID()),
		Username:      fmt.Sprintf("user_%s", randomID()),
		Password:      "testpass123",
		Role:          model.UserRoleUser,
		EmailVerified: true,
	}
	for _, fn := range opts {
		fn(o)
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(o.Password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("fixtures: failed to hash password: %v", err)
	}

	query := `
		CREATE user CONTENT {
			email: $email,
			username: $username,
			hash: $hash,
			role: $role,
			email_verified: $email_verified,
			created_on: time::now(),
			updated_on: time::now()
		}
	`
	vars := map[string]interface{}{
		"email":          o.Email,
		"username":       o.Username,
		"hash":           string(hash),
		"role":           string(o.Role),
		"email_verified": o.EmailVerified,
	}

	results, err := f.db.Query(ctx(), query, vars)
	if err != nil {
		t.Fatalf("fixtures: failed to create user: %v", err)
	}

	user := parseUserResult(t, results)
	user.Hash = nil // Don't expose hash in fixture
	return user
}

// CreateAdmin creates an admin user
func (f *Factory) CreateAdmin(t *testing.T) *model.User {
	return f.CreateUser(t, func(o *UserOpts) {
		o.Role = model.UserRoleAdmin
	})
}

// CreateModerator creates a moderator user
func (f *Factory) CreateModerator(t *testing.T) *model.User {
	return f.CreateUser(t, func(o *UserOpts) {
		o.Role = model.UserRoleModerator
	})
}

// ============================================================================
// Guild Fixtures
// ============================================================================

// GuildOpts customizes guild creation
type GuildOpts struct {
	Name        string
	Description string
	Visibility  string
}

// CreateGuild creates a guild with the given user as admin member
func (f *Factory) CreateGuild(t *testing.T, admin *model.User, opts ...func(*GuildOpts)) *model.Guild {
	t.Helper()

	o := &GuildOpts{
		Name:        fmt.Sprintf("Guild %s", randomID()),
		Description: "Test guild description",
		Visibility:  model.GuildVisibilityPrivate,
	}
	for _, fn := range opts {
		fn(o)
	}

	// Create guild
	guildQuery := `
		CREATE guild CONTENT {
			name: $name,
			description: $description,
			visibility: $visibility,
			created_on: time::now(),
			updated_on: time::now()
		}
	`
	results, err := f.db.Query(ctx(), guildQuery, map[string]interface{}{
		"name":        o.Name,
		"description": o.Description,
		"visibility":  o.Visibility,
	})
	if err != nil {
		t.Fatalf("fixtures: failed to create guild: %v", err)
	}

	guild := parseGuildResult(t, results)

	// Create member for admin
	memberQuery := `
		CREATE member CONTENT {
			name: $name,
			email: $email,
			user: type::record($user_id),
			created_on: time::now(),
			updated_on: time::now()
		}
	`
	memberResults, err := f.db.Query(ctx(), memberQuery, map[string]interface{}{
		"name":    admin.Email,
		"email":   admin.Email,
		"user_id": admin.ID,
	})
	if err != nil {
		t.Fatalf("fixtures: failed to create member: %v", err)
	}
	memberID := parseIDFromResult(t, memberResults)

	// Link member to guild via responsible_for relation with admin role
	relateQuery := `
		LET $m = type::record($member_id);
		LET $g = type::record($guild_id);
		RELATE $m->responsible_for->$g SET role = "admin";
	`
	if err := f.db.Execute(ctx(), relateQuery, map[string]interface{}{
		"member_id": memberID,
		"guild_id":  guild.ID,
	}); err != nil {
		t.Fatalf("fixtures: failed to link member to guild: %v", err)
	}

	return guild
}

// CreatePublicGuild creates a public guild
func (f *Factory) CreatePublicGuild(t *testing.T, admin *model.User) *model.Guild {
	return f.CreateGuild(t, admin, func(o *GuildOpts) {
		o.Visibility = model.GuildVisibilityPublic
	})
}

// AddMemberToGuild adds a user as a member of a guild with default "member" role
func (f *Factory) AddMemberToGuild(t *testing.T, user *model.User, guild *model.Guild) {
	f.addMemberToGuildWithRole(t, user, guild, "member")
}

// AddMemberToGuildAsAdmin adds a user as an admin of a guild
func (f *Factory) AddMemberToGuildAsAdmin(t *testing.T, user *model.User, guild *model.Guild) {
	f.addMemberToGuildWithRole(t, user, guild, "admin")
}

// addMemberToGuildWithRole adds a user to a guild with a specific role
func (f *Factory) addMemberToGuildWithRole(t *testing.T, user *model.User, guild *model.Guild, role string) {
	t.Helper()

	memberQuery := `
		CREATE member CONTENT {
			name: $name,
			email: $email,
			user: type::record($user_id),
			created_on: time::now(),
			updated_on: time::now()
		}
	`
	memberResults, err := f.db.Query(ctx(), memberQuery, map[string]interface{}{
		"name":    user.Email,
		"email":   user.Email,
		"user_id": user.ID,
	})
	if err != nil {
		t.Fatalf("fixtures: failed to create member: %v", err)
	}
	memberID := parseIDFromResult(t, memberResults)

	relateQuery := `
		LET $m = type::record($member_id);
		LET $g = type::record($guild_id);
		RELATE $m->responsible_for->$g SET role = $role;
	`
	if err := f.db.Execute(ctx(), relateQuery, map[string]interface{}{
		"member_id": memberID,
		"guild_id":  guild.ID,
		"role":      role,
	}); err != nil {
		t.Fatalf("fixtures: failed to link member to guild: %v", err)
	}
}

// ============================================================================
// Event Fixtures
// ============================================================================

// EventOpts customizes event creation
type EventOpts struct {
	Title          string
	Template       string
	Visibility     string
	StartTime      time.Time
	EndTime        *time.Time
	MaxAttendees   *int
	IsSupportEvent bool
	Status         string
	AdventureID    string
}

// WithEventVisibility sets event visibility
func WithEventVisibility(vis string) func(*EventOpts) {
	return func(o *EventOpts) {
		o.Visibility = vis
	}
}

// CreateEvent creates an event in a guild
func (f *Factory) CreateEvent(t *testing.T, guild *model.Guild, host *model.User, opts ...func(*EventOpts)) *model.Event {
	t.Helper()

	// First look up the member record for the host user in this guild
	// Use a single query to avoid multi-statement result parsing issues
	memberQuery := `SELECT in AS id FROM responsible_for WHERE out = type::record($guild_id) AND in.user = type::record($user_id) LIMIT 1`
	memberResults, err := f.db.Query(ctx(), memberQuery, map[string]interface{}{
		"user_id":  host.ID,
		"guild_id": guild.ID,
	})
	if err != nil {
		t.Fatalf("fixtures: failed to find member for host: %v", err)
	}
	memberID := parseIDFromResult(t, memberResults)
	if memberID == "" {
		t.Fatalf("fixtures: host is not a member of the guild")
	}

	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	o := &EventOpts{
		Title:      fmt.Sprintf("Event %s", randomID()),
		Template:   model.EventTemplateCasual,
		Visibility: model.EventVisibilityGuilds,
		StartTime:  startTime,
		EndTime:    &endTime,
		Status:     model.EventStatusPublished,
	}
	for _, fn := range opts {
		fn(o)
	}

	// Set confirmation deadline 48h after end time
	var confirmDeadline *time.Time
	if o.EndTime != nil {
		d := o.EndTime.Add(48 * time.Hour)
		confirmDeadline = &d
	}

	vars := map[string]interface{}{
		"guild_id":              guild.ID,
		"title":                 o.Title,
		"template":              o.Template,
		"visibility":            o.Visibility,
		"starts_at":             o.StartTime,
		"ends_at":               o.EndTime,
		"is_support_event":      o.IsSupportEvent,
		"status":                o.Status,
		"created_by":            memberID,
		"confirmation_deadline": confirmDeadline,
	}

	// Build query dynamically to handle optional fields
	query := `
		CREATE event SET
			guild_id = type::record($guild_id),
			title = $title,
			template = $template,
			visibility = $visibility,
			starts_at = $starts_at,
			ends_at = $ends_at,
			is_support_event = $is_support_event,
			status = $status,
			created_by = type::record($created_by),
			confirmation_deadline = $confirmation_deadline,
			confirmed_count = 0,
			requires_confirmation = true,
			completion_verified = false,
			attendee_count = 0,
			created_on = time::now(),
			updated_on = time::now()`

	// Add optional max_attendees if specified
	if o.MaxAttendees != nil {
		query += ", max_attendees = $max_attendees"
		vars["max_attendees"] = *o.MaxAttendees
	}

	results, err := f.db.Query(ctx(), query, vars)
	if err != nil {
		t.Fatalf("fixtures: failed to create event: %v", err)
	}

	return parseEventResult(t, results)
}

// CreateEvent creates an event in a guild (overload that takes adventure as first parameter)
func (f *Factory) CreateEventForAdventure(t *testing.T, adventure *model.Adventure, guild *model.Guild, host *model.User, opts ...func(*EventOpts)) *model.Event {
	t.Helper()

	// First look up the member record for the host user in this guild
	memberQuery := `SELECT in AS id FROM responsible_for WHERE out = type::record($guild_id) AND in.user = type::record($user_id) LIMIT 1`
	memberResults, err := f.db.Query(ctx(), memberQuery, map[string]interface{}{
		"user_id":  host.ID,
		"guild_id": guild.ID,
	})
	if err != nil {
		t.Fatalf("fixtures: failed to find member for host: %v", err)
	}
	memberID := parseIDFromResult(t, memberResults)
	if memberID == "" {
		t.Fatalf("fixtures: host is not a member of the guild")
	}

	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	o := &EventOpts{
		Title:       fmt.Sprintf("Event %s", randomID()),
		Template:    model.EventTemplateCasual,
		Visibility:  model.EventVisibilityGuilds,
		StartTime:   startTime,
		EndTime:     &endTime,
		Status:      model.EventStatusPublished,
		AdventureID: adventure.ID,
	}
	for _, fn := range opts {
		fn(o)
	}

	// Set confirmation deadline 48h after end time
	var confirmDeadline *time.Time
	if o.EndTime != nil {
		d := o.EndTime.Add(48 * time.Hour)
		confirmDeadline = &d
	}

	vars := map[string]interface{}{
		"guild_id":              guild.ID,
		"adventure_id":          o.AdventureID,
		"title":                 o.Title,
		"template":              o.Template,
		"visibility":            o.Visibility,
		"starts_at":             o.StartTime,
		"ends_at":               o.EndTime,
		"is_support_event":      o.IsSupportEvent,
		"status":                o.Status,
		"created_by":            memberID,
		"confirmation_deadline": confirmDeadline,
	}

	query := `
		CREATE event SET
			guild_id = type::record($guild_id),
			adventure_id = type::record($adventure_id),
			title = $title,
			template = $template,
			visibility = $visibility,
			starts_at = $starts_at,
			ends_at = $ends_at,
			is_support_event = $is_support_event,
			status = $status,
			created_by = type::record($created_by),
			confirmation_deadline = $confirmation_deadline,
			confirmed_count = 0,
			requires_confirmation = true,
			completion_verified = false,
			attendee_count = 0,
			created_on = time::now(),
			updated_on = time::now()`

	if o.MaxAttendees != nil {
		query += ", max_attendees = $max_attendees"
		vars["max_attendees"] = *o.MaxAttendees
	}

	results, err := f.db.Query(ctx(), query, vars)
	if err != nil {
		t.Fatalf("fixtures: failed to create event: %v", err)
	}

	return parseEventResult(t, results)
}

// CreateVerifiedEvent creates an event with completion_verified=true
func (f *Factory) CreateVerifiedEvent(t *testing.T, guild *model.Guild, host *model.User) *model.Event {
	t.Helper()

	// Create event that ended in the past
	startTime := time.Now().Add(-4 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	maxAttendees := 2

	event := f.CreateEvent(t, guild, host, func(o *EventOpts) {
		o.StartTime = startTime
		o.EndTime = &endTime
		o.MaxAttendees = &maxAttendees
		o.Status = model.EventStatusCompleted
	})

	// Mark as verified
	query := `UPDATE type::record($event_id) SET completion_verified = true, completion_verified_on = time::now()`
	if err := f.db.Execute(ctx(), query, map[string]interface{}{"event_id": event.ID}); err != nil {
		t.Fatalf("fixtures: failed to verify event: %v", err)
	}

	event.CompletionVerified = true
	return event
}

// CreateRSVP creates an RSVP for a user to an event
func (f *Factory) CreateRSVP(t *testing.T, event *model.Event, user *model.User, status string) *model.EventRSVP {
	t.Helper()

	query := `
		CREATE event_rsvp CONTENT {
			event_id: type::record($event_id),
			user_id: type::record($user_id),
			status: $status,
			rsvp_type: "going",
			requested_on: time::now(),
			updated_on: time::now(),
			plus_ones: 0
		}
	`
	results, err := f.db.Query(ctx(), query, map[string]interface{}{
		"event_id": event.ID,
		"user_id":  user.ID,
		"status":   status,
	})
	if err != nil {
		t.Fatalf("fixtures: failed to create RSVP: %v", err)
	}

	return parseRSVPResult(t, results)
}

// ConfirmEventCompletion marks a user's RSVP as completion confirmed
func (f *Factory) ConfirmEventCompletion(t *testing.T, event *model.Event, user *model.User) {
	t.Helper()

	query := `
		UPDATE event_rsvp SET completion_confirmed = time::now()
		WHERE event_id = type::record($event_id) AND user_id = type::record($user_id)
	`
	if err := f.db.Execute(ctx(), query, map[string]interface{}{
		"event_id": event.ID,
		"user_id":  user.ID,
	}); err != nil {
		t.Fatalf("fixtures: failed to confirm event completion: %v", err)
	}

	// Increment confirmed count on event
	updateEvent := `UPDATE type::record($event_id) SET confirmed_count += 1`
	if err := f.db.Execute(ctx(), updateEvent, map[string]interface{}{"event_id": event.ID}); err != nil {
		t.Fatalf("fixtures: failed to update event confirmed count: %v", err)
	}
}

// ============================================================================
// Trust Rating Fixtures
// ============================================================================

// TrustRatingOpts customizes trust rating creation
type TrustRatingOpts struct {
	TrustLevel string // trust, distrust
	Review     string
}

// CreateTrustRating creates a trust rating between users (requires verified event anchor)
func (f *Factory) CreateTrustRating(t *testing.T, rater, ratee *model.User, anchorEvent *model.Event, opts ...func(*TrustRatingOpts)) {
	t.Helper()

	o := &TrustRatingOpts{
		TrustLevel: "trust",
		Review:     "Great person to hang out with!",
	}
	for _, fn := range opts {
		fn(o)
	}

	query := `
		CREATE trust_rating CONTENT {
			rater_user_id: $rater_id,
			ratee_user_id: $ratee_id,
			anchor_event_id: $anchor_event_id,
			trust_level: $trust_level,
			review: $review,
			visibility: "public",
			created_on: time::now(),
			updated_on: time::now()
		}
	`
	if err := f.db.Execute(ctx(), query, map[string]interface{}{
		"rater_id":        rater.ID,
		"ratee_id":        ratee.ID,
		"anchor_event_id": anchorEvent.ID,
		"trust_level":     o.TrustLevel,
		"review":          o.Review,
	}); err != nil {
		t.Fatalf("fixtures: failed to create trust rating: %v", err)
	}
}

// ============================================================================
// Vote Fixtures
// ============================================================================

// VoteOpts customizes vote creation
type VoteOpts struct {
	Title       string
	VoteType    model.VoteType
	ScopeType   model.VoteScopeType
	OpensAt     time.Time
	ClosesAt    time.Time
	Status      model.VoteStatus
	AllowAbstain bool
	MaxOptionsSelectable *int
}

// CreateVote creates a vote
func (f *Factory) CreateVote(t *testing.T, creator *model.User, guild *model.Guild, opts ...func(*VoteOpts)) *model.Vote {
	t.Helper()

	o := &VoteOpts{
		Title:     fmt.Sprintf("Vote %s", randomID()),
		VoteType:  model.VoteTypeFPTP,
		ScopeType: model.VoteScopeGuild,
		OpensAt:   time.Now(),
		ClosesAt:  time.Now().Add(7 * 24 * time.Hour),
		Status:    model.VoteStatusDraft,
		AllowAbstain: false,
	}
	for _, fn := range opts {
		fn(o)
	}

	var scopeID *string
	if guild != nil {
		scopeID = &guild.ID
	}

	query := `
		CREATE vote CONTENT {
			scope_type: $scope_type,
			scope_id: $scope_id,
			created_by: $created_by,
			title: $title,
			vote_type: $vote_type,
			opens_at: $opens_at,
			closes_at: $closes_at,
			status: $status,
			results_visibility: "after_close",
			allow_abstain: $allow_abstain,
			max_options_selectable: $max_options_selectable,
			created_on: time::now(),
			updated_on: time::now()
		}
	`
	results, err := f.db.Query(ctx(), query, map[string]interface{}{
		"scope_type":             string(o.ScopeType),
		"scope_id":               scopeID,
		"created_by":             creator.ID,
		"title":                  o.Title,
		"vote_type":              string(o.VoteType),
		"opens_at":               o.OpensAt,
		"closes_at":              o.ClosesAt,
		"status":                 string(o.Status),
		"allow_abstain":          o.AllowAbstain,
		"max_options_selectable": o.MaxOptionsSelectable,
	})
	if err != nil {
		t.Fatalf("fixtures: failed to create vote: %v", err)
	}

	return parseVoteResult(t, results)
}

// CreateOpenVote creates a vote that is already open
func (f *Factory) CreateOpenVote(t *testing.T, creator *model.User, guild *model.Guild) *model.Vote {
	return f.CreateVote(t, creator, guild, func(o *VoteOpts) {
		o.Status = model.VoteStatusOpen
		o.OpensAt = time.Now().Add(-1 * time.Hour)
	})
}

// AddVoteOption adds an option to a vote
func (f *Factory) AddVoteOption(t *testing.T, vote *model.Vote, creator *model.User, text string) *model.VoteOption {
	t.Helper()

	query := `
		CREATE vote_option CONTENT {
			vote_id: $vote_id,
			option_text: $option_text,
			sort_order: 0,
			created_by: $created_by,
			created_on: time::now()
		}
	`
	results, err := f.db.Query(ctx(), query, map[string]interface{}{
		"vote_id":     vote.ID,
		"option_text": text,
		"created_by":  creator.ID,
	})
	if err != nil {
		t.Fatalf("fixtures: failed to create vote option: %v", err)
	}

	return parseVoteOptionResult(t, results)
}

// CastBallot casts a ballot for a user
func (f *Factory) CastBallot(t *testing.T, vote *model.Vote, voter *model.User, ballotData model.BallotData) *model.VoteBallot {
	t.Helper()

	query := `
		CREATE vote_ballot CONTENT {
			vote_id: $vote_id,
			voter_user_id: $voter_id,
			voter_snapshot: { username: $username },
			ballot_data: $ballot_data,
			is_abstain: false,
			created_on: time::now()
		}
	`
	username := ""
	if voter.Username != nil {
		username = *voter.Username
	}

	results, err := f.db.Query(ctx(), query, map[string]interface{}{
		"vote_id":     vote.ID,
		"voter_id":    voter.ID,
		"username":    username,
		"ballot_data": ballotData,
	})
	if err != nil {
		t.Fatalf("fixtures: failed to cast ballot: %v", err)
	}

	return parseBallotResult(t, results)
}

// ============================================================================
// Adventure Fixtures
// ============================================================================

// AdventureOpts customizes adventure creation
type AdventureOpts struct {
	Title       string
	Description string
	Visibility  string
}

// WithAdventureVisibility sets adventure visibility
func WithAdventureVisibility(vis string) func(*AdventureOpts) {
	return func(o *AdventureOpts) {
		o.Visibility = vis
	}
}

// CreateAdventure creates an adventure in a guild
func (f *Factory) CreateAdventure(t *testing.T, guild *model.Guild, organizer *model.User, opts ...func(*AdventureOpts)) *model.Adventure {
	t.Helper()

	o := &AdventureOpts{
		Title:       fmt.Sprintf("Adventure %s", randomID()),
		Description: "Test adventure",
		Visibility:  string(model.AdventureVisibilityGuilds),
	}
	for _, fn := range opts {
		fn(o)
	}

	startDate := time.Now().Add(24 * time.Hour)
	endDate := startDate.Add(7 * 24 * time.Hour)

	query := `
		CREATE adventure SET
			title = $title,
			description = $description,
			visibility = $visibility,
			start_date = $start_date,
			end_date = $end_date,
			organizer_type = "guild",
			organizer_id = $guild_id,
			organizer_user_id = type::record($user_id),
			created_by_id = type::record($user_id),
			status = "active",
			created_on = time::now(),
			updated_on = time::now()
	`
	results, err := f.db.Query(ctx(), query, map[string]interface{}{
		"title":       o.Title,
		"description": o.Description,
		"visibility":  o.Visibility,
		"start_date":  startDate,
		"end_date":    endDate,
		"guild_id":    "guild:" + guild.ID,
		"user_id":     organizer.ID,
	})
	if err != nil {
		t.Fatalf("fixtures: failed to create adventure: %v", err)
	}

	return parseAdventureResult(t, results)
}

// ============================================================================
// Block Fixtures
// ============================================================================

// CreateBlock creates a block between users
func (f *Factory) CreateBlock(t *testing.T, blocker, blocked *model.User) {
	t.Helper()

	query := `
		CREATE block CONTENT {
			blocker_user_id: $blocker_id,
			blocked_user_id: $blocked_id,
			created_on: time::now()
		}
	`
	if err := f.db.Execute(ctx(), query, map[string]interface{}{
		"blocker_id": blocker.ID,
		"blocked_id": blocked.ID,
	}); err != nil {
		t.Fatalf("fixtures: failed to create block: %v", err)
	}
}

// ============================================================================
// Result Parsing Helpers
// ============================================================================

func parseUserResult(t *testing.T, results []interface{}) *model.User {
	t.Helper()
	data := extractFirstResult(t, results)
	return &model.User{
		ID:            getString(data, "id"),
		Email:         getString(data, "email"),
		Username:      getStringPtr(data, "username"),
		Role:          model.UserRole(getString(data, "role")),
		EmailVerified: getBool(data, "email_verified"),
		CreatedOn:     getTime(data, "created_on"),
		UpdatedOn:     getTime(data, "updated_on"),
	}
}

func parseGuildResult(t *testing.T, results []interface{}) *model.Guild {
	t.Helper()
	data := extractFirstResult(t, results)
	return &model.Guild{
		ID:          getString(data, "id"),
		Name:        getString(data, "name"),
		Description: getString(data, "description"),
		Visibility:  getString(data, "visibility"),
		CreatedOn:   getTime(data, "created_on"),
		UpdatedOn:   getTime(data, "updated_on"),
	}
}

func parseEventResult(t *testing.T, results []interface{}) *model.Event {
	t.Helper()
	data := extractFirstResult(t, results)
	return &model.Event{
		ID:                   getString(data, "id"),
		Title:                getString(data, "title"),
		Template:             getString(data, "template"),
		Visibility:           getString(data, "visibility"),
		Status:               getString(data, "status"),
		CreatedBy:            getString(data, "created_by"),
		StartTime:            getTime(data, "starts_at"),
		CompletionVerified:   getBool(data, "completion_verified"),
		ConfirmedCount:       getInt(data, "confirmed_count"),
		AttendeeCount:        getInt(data, "attendee_count"),
		CreatedOn:            getTime(data, "created_on"),
		UpdatedOn:            getTime(data, "updated_on"),
	}
}

func parseRSVPResult(t *testing.T, results []interface{}) *model.EventRSVP {
	t.Helper()
	data := extractFirstResult(t, results)
	return &model.EventRSVP{
		ID:          getString(data, "id"),
		EventID:     getString(data, "event_id"),
		UserID:      getString(data, "user_id"),
		Status:      getString(data, "status"),
		RSVPType:    getString(data, "rsvp_type"),
		RequestedOn: getTime(data, "requested_on"),
		UpdatedOn:   getTime(data, "updated_on"),
	}
}

func parseVoteResult(t *testing.T, results []interface{}) *model.Vote {
	t.Helper()
	data := extractFirstResult(t, results)
	return &model.Vote{
		ID:        getString(data, "id"),
		ScopeType: model.VoteScopeType(getString(data, "scope_type")),
		ScopeID:   getStringPtr(data, "scope_id"),
		CreatedBy: getString(data, "created_by"),
		Title:     getString(data, "title"),
		VoteType:  model.VoteType(getString(data, "vote_type")),
		Status:    model.VoteStatus(getString(data, "status")),
		OpensAt:   getTime(data, "opens_at"),
		ClosesAt:  getTime(data, "closes_at"),
		CreatedOn: getTime(data, "created_on"),
		UpdatedOn: getTime(data, "updated_on"),
	}
}

func parseVoteOptionResult(t *testing.T, results []interface{}) *model.VoteOption {
	t.Helper()
	data := extractFirstResult(t, results)
	return &model.VoteOption{
		ID:         getString(data, "id"),
		VoteID:     getString(data, "vote_id"),
		OptionText: getString(data, "option_text"),
		SortOrder:  getInt(data, "sort_order"),
		CreatedBy:  getString(data, "created_by"),
		CreatedOn:  getTime(data, "created_on"),
	}
}

func parseBallotResult(t *testing.T, results []interface{}) *model.VoteBallot {
	t.Helper()
	data := extractFirstResult(t, results)
	return &model.VoteBallot{
		ID:          getString(data, "id"),
		VoteID:      getString(data, "vote_id"),
		VoterUserID: getString(data, "voter_user_id"),
		IsAbstain:   getBool(data, "is_abstain"),
		CreatedOn:   getTime(data, "created_on"),
	}
}

func parseAdventureResult(t *testing.T, results []interface{}) *model.Adventure {
	t.Helper()
	data := extractFirstResult(t, results)
	return &model.Adventure{
		ID:              getString(data, "id"),
		Title:           getString(data, "title"),
		Description:     getStringPtr(data, "description"),
		StartDate:       getTime(data, "start_date"),
		EndDate:         getTime(data, "end_date"),
		Visibility:      model.AdventureVisibility(getString(data, "visibility")),
		OrganizerType:   model.AdventureOrganizerType(getString(data, "organizer_type")),
		OrganizerID:     getString(data, "organizer_id"),
		OrganizerUserID: getString(data, "organizer_user_id"),
		CreatedByID:     getString(data, "created_by_id"),
		Status:          model.AdventureStatus(getString(data, "status")),
		CreatedOn:       getTime(data, "created_on"),
		UpdatedOn:       getTime(data, "updated_on"),
	}
}

func parseIDFromResult(t *testing.T, results []interface{}) string {
	t.Helper()
	data := extractFirstResult(t, results)
	return getString(data, "id")
}

// ============================================================================
// Data Extraction Helpers
// ============================================================================

func extractFirstResult(t *testing.T, results []interface{}) map[string]interface{} {
	t.Helper()
	if len(results) == 0 {
		t.Fatal("fixtures: no results returned")
	}

	// Handle SurrealDB response wrapper
	resp, ok := results[0].(map[string]interface{})
	if !ok {
		t.Fatalf("fixtures: unexpected result type: %T", results[0])
	}

	result, ok := resp["result"]
	if !ok {
		t.Fatal("fixtures: no result in response")
	}

	// Handle array result
	if arr, ok := result.([]interface{}); ok {
		if len(arr) == 0 {
			t.Fatal("fixtures: empty result array")
		}
		data, ok := arr[0].(map[string]interface{})
		if !ok {
			t.Fatalf("fixtures: unexpected array item type: %T", arr[0])
		}
		return data
	}

	// Handle single result
	data, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("fixtures: unexpected result type: %T", result)
	}
	return data
}

func getString(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	// Handle SurrealDB 3 record ID type - could be a struct or map
	if v := data[key]; v != nil {
		// Try to get the ID as a map with "tb" (table) and "id" fields
		if m, ok := v.(map[string]interface{}); ok {
			if tb, ok := m["tb"].(string); ok {
				if id := m["id"]; id != nil {
					return fmt.Sprintf("%s:%v", tb, id)
				}
			}
		}
		// Fallback: use string conversion but fix the format if needed
		s := fmt.Sprintf("%v", v)
		// Convert "{table id}" to "table:id"
		if len(s) > 2 && s[0] == '{' && s[len(s)-1] == '}' {
			inner := s[1 : len(s)-1]
			for i, c := range inner {
				if c == ' ' {
					return inner[:i] + ":" + inner[i+1:]
				}
			}
		}
		return s
	}
	return ""
}

func getStringPtr(data map[string]interface{}, key string) *string {
	if v, ok := data[key].(string); ok {
		return &v
	}
	return nil
}

func getBool(data map[string]interface{}, key string) bool {
	if v, ok := data[key].(bool); ok {
		return v
	}
	return false
}

func getInt(data map[string]interface{}, key string) int {
	if v, ok := data[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getTime(data map[string]interface{}, key string) time.Time {
	if v, ok := data[key].(string); ok {
		t, _ := time.Parse(time.RFC3339Nano, v)
		return t
	}
	return time.Time{}
}
