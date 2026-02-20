package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mrand "math/rand/v2"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"

	"golang.org/x/crypto/bcrypt"
)

// SeederService generates mock data for testing and development
type SeederService struct {
	db database.Database
}

// NewSeederService creates a new seeder service
func NewSeederService(db database.Database) *SeederService {
	return &SeederService{db: db}
}

// SeedUsersRequest configures user seeding
type SeedUsersRequest struct {
	Count  int          `json:"count"`
	Region *BoundingBox `json:"region,omitempty"`
	// ActivityDistribution specifies percentage of users in each activity status
	// Keys: "active_now", "active_today", "active_this_week", "away"
	ActivityDistribution map[string]int `json:"activity_distribution,omitempty"`
	// Prefix for seeded user emails to identify them for cleanup
	Prefix string `json:"prefix,omitempty"`
}

// SeedGuildsRequest configures guild seeding
type SeedGuildsRequest struct {
	Count           int    `json:"count"`
	MembersPerGuild int    `json:"members_per_guild,omitempty"`
	Visibility      string `json:"visibility,omitempty"` // "public" or "private"
	Prefix          string `json:"prefix,omitempty"`
}

// SeedEventsRequest configures event seeding
type SeedEventsRequest struct {
	Count   int    `json:"count"`
	GuildID string `json:"guild_id,omitempty"` // If empty, creates events across all seeded guilds
	Status  string `json:"status,omitempty"`   // "published", "draft", "cancelled"
	Prefix  string `json:"prefix,omitempty"`
}

// SeedScenarioRequest runs a predefined scenario
type SeedScenarioRequest struct {
	Scenario string `json:"scenario"` // e.g., "sf_discovery_pool", "active_guild", "event_with_attendees"
}

// SeedResult contains the results of a seeding operation
type SeedResult struct {
	Created  int      `json:"created"`
	IDs      []string `json:"ids"`
	Duration int64    `json:"duration_ms"`
}

// CleanupResult contains the results of a cleanup operation
type CleanupResult struct {
	Deleted  int   `json:"deleted"`
	Duration int64 `json:"duration_ms"`
}

// Default bounding boxes for common cities
var (
	BoundingBoxSF = BoundingBox{
		MinLat: 37.7079,
		MaxLat: 37.8324,
		MinLng: -122.5149,
		MaxLng: -122.3570,
	}
	BoundingBoxNYC = BoundingBox{
		MinLat: 40.4961,
		MaxLat: 40.9155,
		MinLng: -74.2557,
		MaxLng: -73.7004,
	}
	BoundingBoxLA = BoundingBox{
		MinLat: 33.7037,
		MaxLat: 34.3373,
		MinLng: -118.6682,
		MaxLng: -118.1553,
	}
)

// Sample data for realistic generation
var (
	firstNames = []string{
		"Emma", "Liam", "Olivia", "Noah", "Ava", "Ethan", "Sophia", "Mason",
		"Isabella", "William", "Mia", "James", "Charlotte", "Benjamin", "Amelia",
		"Lucas", "Harper", "Henry", "Evelyn", "Alexander", "Abigail", "Michael",
		"Emily", "Daniel", "Elizabeth", "Jacob", "Sofia", "Logan", "Avery", "Jackson",
		"Ella", "Sebastian", "Scarlett", "Aiden", "Grace", "Matthew", "Chloe", "Samuel",
		"Victoria", "David", "Riley", "Joseph", "Aria", "Carter", "Lily", "Owen",
		"Aurora", "Wyatt", "Zoey", "John", "Penelope", "Jack", "Layla", "Luke",
	}
	lastNames = []string{
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
		"Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson",
		"Thomas", "Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson",
		"White", "Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson", "Walker",
		"Young", "Allen", "King", "Wright", "Scott", "Torres", "Nguyen", "Hill", "Flores",
		"Green", "Adams", "Nelson", "Baker", "Hall", "Rivera", "Campbell", "Mitchell",
	}
	guildNames = []string{
		"Adventure Seekers", "Weekend Warriors", "Urban Explorers", "Mountain Climbers",
		"Beach Lovers", "Game Night Crew", "Book Club", "Hiking Enthusiasts",
		"Foodies United", "Photography Club", "Fitness Friends", "Art Appreciators",
		"Music Makers", "Tech Talks", "Wine Tasters", "Coffee Connoisseurs",
		"Yoga Circle", "Running Club", "Cycling Group", "Board Game Buffs",
	}
	eventTitles = []string{
		"Weekly Meetup", "Game Night", "Hiking Trip", "Coffee Chat",
		"Movie Night", "Dinner Party", "Beach Day", "Museum Visit",
		"Picnic in the Park", "Karaoke Night", "Trivia Tuesday", "Book Discussion",
		"Wine Tasting", "Cooking Class", "Art Workshop", "Photography Walk",
		"Yoga Session", "Running Club", "Cycling Adventure", "Board Game Marathon",
	}
	bios = []string{
		"Love exploring new places and meeting interesting people.",
		"Passionate about outdoor adventures and good conversations.",
		"Always up for trying something new!",
		"Coffee enthusiast and weekend explorer.",
		"Looking to make meaningful connections.",
		"Adventurer at heart, homebody on rainy days.",
		"Tech nerd by day, social butterfly by night.",
		"Foodie seeking fellow culinary adventurers.",
		"Bookworm who also loves hiking.",
		"Music lover and amateur photographer.",
	}
	taglines = []string{
		"Let's explore!", "Adventure awaits", "Life is short, let's connect",
		"Always curious", "Seeking kindred spirits", "Ready for anything",
		"Making memories", "Living my best life", "Open to possibilities",
		"Here for good times",
	}
)

// SeedUsers creates mock users with profiles
func (s *SeederService) SeedUsers(ctx context.Context, req SeedUsersRequest) (*SeedResult, error) {
	start := time.Now()

	if req.Count <= 0 || req.Count > 1000 {
		return nil, fmt.Errorf("count must be between 1 and 1000")
	}

	if req.Prefix == "" {
		req.Prefix = "seed_"
	}

	if req.Region == nil {
		req.Region = &BoundingBoxSF
	}

	// Default activity distribution
	if req.ActivityDistribution == nil {
		req.ActivityDistribution = map[string]int{
			"active_now":       20,
			"active_today":     30,
			"active_this_week": 30,
			"away":             20,
		}
	}

	ids := make([]string, 0, req.Count)
	password := "testpass123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	for i := 0; i < req.Count; i++ {
		randID := randomID()
		email := fmt.Sprintf("%s%s@test.local", req.Prefix, randID)
		username := fmt.Sprintf("%s%s", req.Prefix, randID)
		firstName := firstNames[mrand.IntN(len(firstNames))]
		lastName := lastNames[mrand.IntN(len(lastNames))]

		// Create user
		userQuery := `
			CREATE user CONTENT {
				email: $email,
				username: $username,
				hash: $hash,
				firstname: $firstname,
				lastname: $lastname,
				role: "user",
				email_verified: true,
				created_on: time::now(),
				updated_on: time::now()
			}
		`
		results, err := s.db.Query(ctx, userQuery, map[string]interface{}{
			"email":     email,
			"username":  username,
			"hash":      string(hash),
			"firstname": firstName,
			"lastname":  lastName,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		userID := extractID(results)
		if userID == "" {
			return nil, fmt.Errorf("failed to extract user ID")
		}
		ids = append(ids, userID)

		// Generate location within bounds
		lat := req.Region.MinLat + mrand.Float64()*(req.Region.MaxLat-req.Region.MinLat)
		lng := req.Region.MinLng + mrand.Float64()*(req.Region.MaxLng-req.Region.MinLng)

		// Determine last active based on distribution
		lastActive := generateLastActive(req.ActivityDistribution)

		// Create profile
		bio := bios[mrand.IntN(len(bios))]
		tagline := taglines[mrand.IntN(len(taglines))]

		profileQuery := `
			CREATE profile CONTENT {
				user_id: type::record($user_id),
				bio: $bio,
				tagline: $tagline,
				visibility: "public",
				location: {
					lat: $lat,
					lng: $lng,
					city: $city,
					country: "United States",
					country_code: "US"
				},
				last_active: $last_active,
				discovery_eligible: true,
				question_count: 5,
				categories_completed: ["values", "social", "lifestyle", "communication"],
				profile_completion_score: 0.8,
				created_on: time::now(),
				updated_on: time::now()
			}
		`
		_, err = s.db.Query(ctx, profileQuery, map[string]interface{}{
			"user_id":     userID,
			"bio":         bio,
			"tagline":     tagline,
			"lat":         lat,
			"lng":         lng,
			"city":        "San Francisco",
			"last_active": lastActive,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create profile: %w", err)
		}
	}

	return &SeedResult{
		Created:  len(ids),
		IDs:      ids,
		Duration: time.Since(start).Milliseconds(),
	}, nil
}

// SeedGuilds creates mock guilds with members
func (s *SeederService) SeedGuilds(ctx context.Context, req SeedGuildsRequest) (*SeedResult, error) {
	start := time.Now()

	if req.Count <= 0 || req.Count > 100 {
		return nil, fmt.Errorf("count must be between 1 and 100")
	}

	if req.Prefix == "" {
		req.Prefix = "seed_"
	}

	if req.Visibility == "" {
		req.Visibility = model.GuildVisibilityPublic
	}

	if req.MembersPerGuild <= 0 {
		req.MembersPerGuild = 5
	}

	// First, get or create seed users to be members
	userQuery := fmt.Sprintf(`SELECT id FROM user WHERE email CONTAINS '%s' LIMIT %d`, req.Prefix, req.MembersPerGuild*req.Count)
	userResults, err := s.db.Query(ctx, userQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	userIDs := extractIDs(userResults)
	if len(userIDs) == 0 {
		// Create some users first
		seedResult, err := s.SeedUsers(ctx, SeedUsersRequest{
			Count:  req.MembersPerGuild * req.Count,
			Prefix: req.Prefix,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to seed users for guilds: %w", err)
		}
		userIDs = seedResult.IDs
	}

	ids := make([]string, 0, req.Count)

	for i := 0; i < req.Count; i++ {
		name := fmt.Sprintf("%s%s", req.Prefix, guildNames[mrand.IntN(len(guildNames))])
		description := fmt.Sprintf("A community for %s enthusiasts", name)

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
		results, err := s.db.Query(ctx, guildQuery, map[string]interface{}{
			"name":        name,
			"description": description,
			"visibility":  req.Visibility,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create guild: %w", err)
		}

		guildID := extractID(results)
		if guildID == "" {
			return nil, fmt.Errorf("failed to extract guild ID")
		}
		ids = append(ids, guildID)

		// Add members
		startIdx := (i * req.MembersPerGuild) % len(userIDs)
		for j := 0; j < req.MembersPerGuild && j < len(userIDs); j++ {
			userID := userIDs[(startIdx+j)%len(userIDs)]
			role := "member"
			if j == 0 {
				role = "admin"
			}

			// Create member record
			memberQuery := `
				CREATE member CONTENT {
					name: $email,
					email: $email,
					user: type::record($user_id),
					created_on: time::now(),
					updated_on: time::now()
				}
			`
			email := fmt.Sprintf("member_%s@test.local", randomID())
			memberResults, err := s.db.Query(ctx, memberQuery, map[string]interface{}{
				"email":   email,
				"user_id": userID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create member: %w", err)
			}

			memberID := extractID(memberResults)
			if memberID == "" {
				continue
			}

			// Link member to guild
			relateQuery := `
				LET $m = type::record($member_id);
				LET $g = type::record($guild_id);
				RELATE $m->responsible_for->$g SET role = $role;
			`
			err = s.db.Execute(ctx, relateQuery, map[string]interface{}{
				"member_id": memberID,
				"guild_id":  guildID,
				"role":      role,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to link member to guild: %w", err)
			}
		}
	}

	return &SeedResult{
		Created:  len(ids),
		IDs:      ids,
		Duration: time.Since(start).Milliseconds(),
	}, nil
}

// SeedEvents creates mock events
func (s *SeederService) SeedEvents(ctx context.Context, req SeedEventsRequest) (*SeedResult, error) {
	start := time.Now()

	if req.Count <= 0 || req.Count > 100 {
		return nil, fmt.Errorf("count must be between 1 and 100")
	}

	if req.Prefix == "" {
		req.Prefix = "seed_"
	}

	if req.Status == "" {
		req.Status = model.EventStatusPublished
	}

	// Get guilds to create events in
	var guildIDs []string
	if req.GuildID != "" {
		guildIDs = []string{req.GuildID}
	} else {
		guildQuery := fmt.Sprintf(`SELECT id FROM guild WHERE name CONTAINS '%s' LIMIT 10`, req.Prefix)
		guildResults, err := s.db.Query(ctx, guildQuery, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to query guilds: %w", err)
		}
		guildIDs = extractIDs(guildResults)
	}

	if len(guildIDs) == 0 {
		// Create a guild first
		guildResult, err := s.SeedGuilds(ctx, SeedGuildsRequest{
			Count:  1,
			Prefix: req.Prefix,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to seed guild for events: %w", err)
		}
		guildIDs = guildResult.IDs
	}

	ids := make([]string, 0, req.Count)

	for i := 0; i < req.Count; i++ {
		guildID := guildIDs[i%len(guildIDs)]

		// Get a member of this guild to be the host
		memberQuery := `SELECT in AS id FROM responsible_for WHERE out = type::record($guild_id) LIMIT 1`
		memberResults, err := s.db.Query(ctx, memberQuery, map[string]interface{}{
			"guild_id": guildID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to find guild member: %w", err)
		}

		memberID := extractID(memberResults)
		if memberID == "" {
			continue
		}

		title := fmt.Sprintf("%s%s", req.Prefix, eventTitles[mrand.IntN(len(eventTitles))])
		startTime := time.Now().Add(time.Duration(mrand.IntN(14)+1) * 24 * time.Hour)
		endTime := startTime.Add(2 * time.Hour)
		confirmDeadline := endTime.Add(48 * time.Hour)

		eventQuery := `
			CREATE event SET
				guild_id = type::record($guild_id),
				title = $title,
				template = "casual",
				visibility = "guilds",
				starts_at = $starts_at,
				ends_at = $ends_at,
				is_support_event = false,
				status = $status,
				created_by = type::record($created_by),
				confirmation_deadline = $confirmation_deadline,
				confirmed_count = 0,
				requires_confirmation = true,
				completion_verified = false,
				attendee_count = 0,
				created_on = time::now(),
				updated_on = time::now()
		`
		results, err := s.db.Query(ctx, eventQuery, map[string]interface{}{
			"guild_id":              guildID,
			"title":                 title,
			"starts_at":             startTime,
			"ends_at":               endTime,
			"status":                req.Status,
			"created_by":            memberID,
			"confirmation_deadline": confirmDeadline,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create event: %w", err)
		}

		eventID := extractID(results)
		if eventID != "" {
			ids = append(ids, eventID)
		}
	}

	return &SeedResult{
		Created:  len(ids),
		IDs:      ids,
		Duration: time.Since(start).Milliseconds(),
	}, nil
}

// SeedScenario runs a predefined scenario
func (s *SeederService) SeedScenario(ctx context.Context, req SeedScenarioRequest) (*SeedResult, error) {
	start := time.Now()
	var totalCreated int
	var allIDs []string

	switch req.Scenario {
	case "sf_discovery_pool":
		// 20 users in SF for discovery testing
		result, err := s.SeedUsers(ctx, SeedUsersRequest{
			Count:  20,
			Region: &BoundingBoxSF,
			Prefix: "sf_",
			ActivityDistribution: map[string]int{
				"active_now":       40,
				"active_today":     30,
				"active_this_week": 20,
				"away":             10,
			},
		})
		if err != nil {
			return nil, err
		}
		totalCreated = result.Created
		allIDs = result.IDs

	case "active_guild":
		// A guild with 10 members and 5 events
		userResult, err := s.SeedUsers(ctx, SeedUsersRequest{
			Count:  10,
			Prefix: "guild_",
		})
		if err != nil {
			return nil, err
		}
		allIDs = append(allIDs, userResult.IDs...)
		totalCreated += userResult.Created

		guildResult, err := s.SeedGuilds(ctx, SeedGuildsRequest{
			Count:           1,
			MembersPerGuild: 10,
			Prefix:          "guild_",
		})
		if err != nil {
			return nil, err
		}
		allIDs = append(allIDs, guildResult.IDs...)
		totalCreated += guildResult.Created

		eventResult, err := s.SeedEvents(ctx, SeedEventsRequest{
			Count:   5,
			GuildID: guildResult.IDs[0],
			Prefix:  "guild_",
		})
		if err != nil {
			return nil, err
		}
		allIDs = append(allIDs, eventResult.IDs...)
		totalCreated += eventResult.Created

	case "event_with_attendees":
		// An event with 20 attendees for testing RSVPs
		userResult, err := s.SeedUsers(ctx, SeedUsersRequest{
			Count:  20,
			Prefix: "event_",
		})
		if err != nil {
			return nil, err
		}
		allIDs = append(allIDs, userResult.IDs...)
		totalCreated += userResult.Created

		guildResult, err := s.SeedGuilds(ctx, SeedGuildsRequest{
			Count:           1,
			MembersPerGuild: 20,
			Prefix:          "event_",
		})
		if err != nil {
			return nil, err
		}
		allIDs = append(allIDs, guildResult.IDs...)
		totalCreated += guildResult.Created

		eventResult, err := s.SeedEvents(ctx, SeedEventsRequest{
			Count:   1,
			GuildID: guildResult.IDs[0],
			Prefix:  "event_",
		})
		if err != nil {
			return nil, err
		}
		allIDs = append(allIDs, eventResult.IDs...)
		totalCreated += eventResult.Created

	default:
		return nil, fmt.Errorf("unknown scenario: %s", req.Scenario)
	}

	return &SeedResult{
		Created:  totalCreated,
		IDs:      allIDs,
		Duration: time.Since(start).Milliseconds(),
	}, nil
}

// Cleanup removes all seeded data with the given prefix
func (s *SeederService) Cleanup(ctx context.Context, prefix string) (*CleanupResult, error) {
	start := time.Now()

	if prefix == "" {
		prefix = "seed_"
	}

	var totalDeleted int

	// Delete events
	eventQuery := fmt.Sprintf(`DELETE event WHERE title CONTAINS '%s'`, prefix)
	if err := s.db.Execute(ctx, eventQuery, nil); err != nil {
		return nil, fmt.Errorf("failed to delete events: %w", err)
	}

	// Delete responsible_for relations for members we're about to delete
	relQuery := fmt.Sprintf(`DELETE responsible_for WHERE in.email CONTAINS '%s'`, prefix)
	if err := s.db.Execute(ctx, relQuery, nil); err != nil {
		return nil, fmt.Errorf("failed to delete relations: %w", err)
	}

	// Delete members
	memberQuery := fmt.Sprintf(`DELETE member WHERE email CONTAINS '%s'`, prefix)
	if err := s.db.Execute(ctx, memberQuery, nil); err != nil {
		return nil, fmt.Errorf("failed to delete members: %w", err)
	}

	// Delete guilds
	guildQuery := fmt.Sprintf(`DELETE guild WHERE name CONTAINS '%s'`, prefix)
	if err := s.db.Execute(ctx, guildQuery, nil); err != nil {
		return nil, fmt.Errorf("failed to delete guilds: %w", err)
	}

	// Delete profiles for seeded users
	profileQuery := fmt.Sprintf(`DELETE profile WHERE user_id.email CONTAINS '%s'`, prefix)
	if err := s.db.Execute(ctx, profileQuery, nil); err != nil {
		return nil, fmt.Errorf("failed to delete profiles: %w", err)
	}

	// Delete users
	userQuery := fmt.Sprintf(`DELETE user WHERE email CONTAINS '%s'`, prefix)
	if err := s.db.Execute(ctx, userQuery, nil); err != nil {
		return nil, fmt.Errorf("failed to delete users: %w", err)
	}

	// Count what we deleted (approximation based on prefix)
	countQuery := fmt.Sprintf(`SELECT count() FROM user WHERE email CONTAINS '%s' GROUP ALL`, prefix)
	results, _ := s.db.Query(ctx, countQuery, nil)
	if len(results) > 0 {
		// If we get here, count should be 0 since we deleted them
		totalDeleted = 0
	}

	return &CleanupResult{
		Deleted:  totalDeleted,
		Duration: time.Since(start).Milliseconds(),
	}, nil
}

// Helper functions

func randomID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func generateLastActive(distribution map[string]int) time.Time {
	total := 0
	for _, v := range distribution {
		total += v
	}
	if total == 0 {
		return time.Now().Add(-24 * time.Hour)
	}

	r := mrand.IntN(total)
	cumulative := 0

	for status, weight := range distribution {
		cumulative += weight
		if r < cumulative {
			switch status {
			case "active_now":
				return time.Now().Add(-time.Duration(mrand.IntN(10)) * time.Minute)
			case "active_today":
				return time.Now().Add(-time.Duration(mrand.IntN(24)) * time.Hour)
			case "active_this_week":
				return time.Now().Add(-time.Duration(mrand.IntN(7)*24) * time.Hour)
			case "away":
				return time.Now().Add(-time.Duration(mrand.IntN(30)+7) * 24 * time.Hour)
			}
		}
	}

	return time.Now().Add(-24 * time.Hour)
}

func extractID(results []interface{}) string {
	if len(results) == 0 {
		return ""
	}

	resp, ok := results[0].(map[string]interface{})
	if !ok {
		return ""
	}

	result, ok := resp["result"]
	if !ok {
		return ""
	}

	// Handle array result
	if arr, ok := result.([]interface{}); ok {
		if len(arr) == 0 {
			return ""
		}
		data, ok := arr[0].(map[string]interface{})
		if !ok {
			return ""
		}
		return formatID(data["id"])
	}

	// Handle single result
	data, ok := result.(map[string]interface{})
	if !ok {
		return ""
	}
	return formatID(data["id"])
}

func extractIDs(results []interface{}) []string {
	var ids []string
	if len(results) == 0 {
		return ids
	}

	resp, ok := results[0].(map[string]interface{})
	if !ok {
		return ids
	}

	result, ok := resp["result"]
	if !ok {
		return ids
	}

	arr, ok := result.([]interface{})
	if !ok {
		return ids
	}

	for _, item := range arr {
		data, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if id := formatID(data["id"]); id != "" {
			ids = append(ids, id)
		}
	}

	return ids
}

func formatID(v interface{}) string {
	if v == nil {
		return ""
	}

	if s, ok := v.(string); ok {
		return s
	}

	// Handle SurrealDB 3 record ID type
	if m, ok := v.(map[string]interface{}); ok {
		if tb, ok := m["tb"].(string); ok {
			if id := m["id"]; id != nil {
				return fmt.Sprintf("%s:%v", tb, id)
			}
		}
	}

	// Fallback: convert "{table id}" to "table:id"
	s := fmt.Sprintf("%v", v)
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
