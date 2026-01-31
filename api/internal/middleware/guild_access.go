package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/forgo/saga/api/internal/model"
)

// GuildMembershipChecker defines the interface for checking guild membership
type GuildMembershipChecker interface {
	IsMember(ctx context.Context, userID, guildID string) (bool, error)
}

// GuildIDKey is the context key for guild ID
const GuildIDKey contextKey = "guildID"

// GetGuildID extracts the guild ID from context
func GetGuildID(ctx context.Context) string {
	if id, ok := ctx.Value(GuildIDKey).(string); ok {
		return id
	}
	return ""
}

// GuildAccess returns a middleware that validates guild membership
// It expects the guild ID to be in the URL path parameter
func GuildAccess(checker GuildMembershipChecker) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from auth context
			userID := GetUserID(r.Context())
			if userID == "" {
				model.NewUnauthorizedError("authentication required").WriteJSON(w)
				return
			}

			// Extract guild ID from URL path
			guildID := extractGuildID(r.URL.Path)
			if guildID == "" {
				model.NewBadRequestError("invalid guild ID").WriteJSON(w)
				return
			}

			// Check membership
			isMember, err := checker.IsMember(r.Context(), userID, guildID)
			if err != nil {
				// Log error but return 404 to not leak information
				model.NewNotFoundError("guild").WriteJSON(w)
				return
			}

			if !isMember {
				// Return 404 instead of 403 to not leak guild existence
				model.NewNotFoundError("guild").WriteJSON(w)
				return
			}

			// Add guild ID to context
			ctx := context.WithValue(r.Context(), GuildIDKey, guildID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractGuildID extracts the guild ID from URL path
// Expected formats:
// - /v1/guilds/{guildId}
// - /v1/guilds/{guildId}/people
// - /v1/guilds/{guildId}/people/{personId}
// etc.
func extractGuildID(path string) string {
	parts := strings.Split(path, "/")

	// Find "guilds" in path and get the next segment
	for i, part := range parts {
		if part == "guilds" && i+1 < len(parts) {
			guildID := parts[i+1]
			// Validate it looks like an ID (not empty, not a sub-resource name)
			if guildID != "" && guildID != "people" && guildID != "activities" && guildID != "timers" && guildID != "members" && guildID != "events" {
				return guildID
			}
		}
	}

	return ""
}

// PersonIDKey is the context key for person ID
const PersonIDKey contextKey = "personID"

// GetPersonID extracts the person ID from context
func GetPersonID(ctx context.Context) string {
	if id, ok := ctx.Value(PersonIDKey).(string); ok {
		return id
	}
	return ""
}

// extractPersonID extracts the person ID from URL path
func extractPersonID(path string) string {
	parts := strings.Split(path, "/")

	// Find "people" in path and get the next segment
	for i, part := range parts {
		if part == "people" && i+1 < len(parts) {
			personID := parts[i+1]
			if personID != "" && personID != "timers" {
				return personID
			}
		}
	}

	return ""
}

// ActivityIDKey is the context key for activity ID
const ActivityIDKey contextKey = "activityID"

// GetActivityID extracts the activity ID from context
func GetActivityID(ctx context.Context) string {
	if id, ok := ctx.Value(ActivityIDKey).(string); ok {
		return id
	}
	return ""
}

// TimerIDKey is the context key for timer ID
const TimerIDKey contextKey = "timerID"

// GetTimerID extracts the timer ID from context
func GetTimerID(ctx context.Context) string {
	if id, ok := ctx.Value(TimerIDKey).(string); ok {
		return id
	}
	return ""
}

// ExtractPathParams extracts guild, person, activity, and timer IDs from path
// and adds them to context
func ExtractPathParams(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		ctx := r.Context()

		// Extract guild ID
		if guildID := extractGuildID(path); guildID != "" {
			ctx = context.WithValue(ctx, GuildIDKey, guildID)
		}

		// Extract person ID
		if personID := extractPersonID(path); personID != "" {
			ctx = context.WithValue(ctx, PersonIDKey, personID)
		}

		// Extract activity ID
		if activityID := extractActivityID(path); activityID != "" {
			ctx = context.WithValue(ctx, ActivityIDKey, activityID)
		}

		// Extract timer ID
		if timerID := extractTimerID(path); timerID != "" {
			ctx = context.WithValue(ctx, TimerIDKey, timerID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractActivityID(path string) string {
	parts := strings.Split(path, "/")

	for i, part := range parts {
		if part == "activities" && i+1 < len(parts) {
			activityID := parts[i+1]
			if activityID != "" {
				return activityID
			}
		}
	}

	return ""
}

func extractTimerID(path string) string {
	parts := strings.Split(path, "/")

	for i, part := range parts {
		if part == "timers" && i+1 < len(parts) {
			timerID := parts[i+1]
			if timerID != "" && timerID != "reset" {
				return timerID
			}
		}
	}

	return ""
}
