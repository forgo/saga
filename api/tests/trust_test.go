// Package tests contains end-to-end acceptance tests for the Saga API.
package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/repository"
	"github.com/forgo/saga/api/internal/service"
	"github.com/forgo/saga/api/internal/testing/fixtures"
	"github.com/forgo/saga/api/internal/testing/testdb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
FEATURE: Trust Ratings
DOMAIN: Trust & Reputation

ACCEPTANCE CRITERIA:
===================

AC-TRUST-001: Create Trust Rating
  GIVEN a verified event with both users confirmed attendance
  WHEN user A rates user B as "trust" with a review
  THEN the rating is created successfully
  AND the rating appears in user B's received ratings

AC-TRUST-002: Create Distrust Rating
  GIVEN a verified event with both users confirmed attendance
  WHEN user A rates user B as "distrust" with a review
  THEN the rating is created
  AND the review is admin-only visibility

AC-TRUST-003: Cannot Rate Without Anchor
  GIVEN two users who have never attended the same event
  WHEN user A attempts to rate user B
  THEN the request fails with anchor not verified error

AC-TRUST-004: Cannot Self-Rate
  GIVEN a verified event attendance
  WHEN the user attempts to rate themselves
  THEN the request fails with cannot rate yourself error

AC-TRUST-005: Cannot Rate Same User Twice (for same anchor)
  GIVEN existing rating from A to B for event X
  WHEN A rates B again for event X
  THEN the request fails with conflict error

AC-TRUST-006: Review Required
  GIVEN valid anchor
  WHEN user rates without review
  THEN the request fails with validation error

AC-TRUST-007: Review Max Length
  GIVEN valid anchor
  WHEN user rates with review > 240 chars
  THEN the request fails with validation error

AC-TRUST-008: Update Rating - 30 Day Cooldown
  GIVEN rating created less than 30 days ago
  WHEN user attempts to change trust level
  THEN the request fails with cooldown error

AC-TRUST-009: Update Rating - After Cooldown
  GIVEN rating created more than 30 days ago
  WHEN user changes trust level
  THEN rating updated successfully

AC-TRUST-010: Delete Rating (Unset)
  GIVEN existing rating
  WHEN user deletes rating
  THEN rating is removed
  AND state returns to neutral

AC-TRUST-011: Daily Rate Limit
  GIVEN user has created 10 ratings today
  WHEN user creates 11th rating
  THEN the request fails with rate limit error

AC-TRUST-012: Endorsement - Valid
  GIVEN trust rating on event X
  AND endorser also attended event X
  WHEN endorser endorses the rating
  THEN endorsement is created

AC-TRUST-013: Endorsement - Not Attendee
  GIVEN trust rating on event X
  AND potential endorser did NOT attend X
  WHEN they attempt to endorse
  THEN the request fails with forbidden error

AC-TRUST-014: Endorsement - No Duplicates
  GIVEN existing endorsement from user A
  WHEN user A endorses same rating again
  THEN the request fails with conflict error

AC-TRUST-015: Get Trust Aggregate
  GIVEN user has received 5 trust, 1 distrust ratings
  WHEN requesting aggregate
  THEN returns trust_count: 5, distrust_count: 1

AC-TRUST-016: Admin Distrust Signals
  GIVEN users with high distrust counts
  WHEN admin queries distrust signals
  THEN flagged users returned for review
*/

// createTrustRatingService creates a TrustRatingService instance for testing
func createTrustRatingService(t *testing.T, tdb *testdb.TestDB) *service.TrustRatingService {
	t.Helper()

	repo := repository.NewTrustRatingRepository(tdb.DB)

	return service.NewTrustRatingService(service.TrustRatingServiceConfig{
		Repo: repo,
	})
}

// setupVerifiedEventWithBothAttendees creates a verified event with both users as confirmed attendees
func setupVerifiedEventWithBothAttendees(t *testing.T, f *fixtures.Factory, tdb *testdb.TestDB, userA, userB *model.User) *model.Event {
	t.Helper()
	ctx := context.Background()

	// Create a guild first
	guild := f.CreatePublicGuild(t, userA)

	// Add userB to guild
	f.AddMemberToGuild(t, userB, guild)

	// Create a verified event
	event := f.CreateVerifiedEvent(t, guild, userA)

	// Create RSVPs for both users
	f.CreateRSVP(t, event, userA, model.RSVPStatusApproved)
	f.CreateRSVP(t, event, userB, model.RSVPStatusApproved)

	// Confirm attendance for both
	f.ConfirmEventCompletion(t, event, userA)
	f.ConfirmEventCompletion(t, event, userB)

	// Create the fn::can_rate_user function if it doesn't exist
	// This function checks if both users attended and the event is verified
	// Note: All parameters are strings - we convert to records internally
	fnQuery := `
		DEFINE FUNCTION OVERWRITE fn::can_rate_user($rater_id: string, $ratee_id: string, $anchor_type: string, $anchor_id: string) {
			IF $anchor_type = "event" {
				LET $event_rec = type::record($anchor_id);
				LET $rater_rec = type::record($rater_id);
				LET $ratee_rec = type::record($ratee_id);
				LET $event = (SELECT completion_verified FROM $event_rec);
				IF $event[0].completion_verified != true {
					RETURN false;
				};
				LET $rater_rsvp = (SELECT * FROM event_rsvp WHERE event_id = $event_rec AND user_id = $rater_rec AND status = "approved" AND completion_confirmed IS NOT NULL LIMIT 1);
				LET $ratee_rsvp = (SELECT * FROM event_rsvp WHERE event_id = $event_rec AND user_id = $ratee_rec AND status = "approved" AND completion_confirmed IS NOT NULL LIMIT 1);
				RETURN array::len($rater_rsvp) > 0 AND array::len($ratee_rsvp) > 0;
			};
			RETURN false;
		}
	`
	if err := tdb.DB.Execute(ctx, fnQuery, nil); err != nil {
		// Function might already exist, that's OK
		t.Logf("Note: fn::can_rate_user function setup: %v", err)
	}

	return event
}

func TestTrustRating_Create(t *testing.T) {
	// AC-TRUST-001: Create Trust Rating
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	// Create two users
	userA := f.CreateUser(t)
	userB := f.CreateUser(t)

	// Setup verified event with both users as confirmed attendees
	event := setupVerifiedEventWithBothAttendees(t, f, tdb, userA, userB)

	// User A rates User B as trust
	rating, err := trustService.Create(ctx, userA.ID, &model.CreateTrustRatingRequest{
		RateeID:     userB.ID,
		AnchorType:  string(model.TrustAnchorEvent),
		AnchorID:    event.ID,
		TrustLevel:  string(model.TrustLevelTrust),
		TrustReview: "Great person to meet! Really enjoyed our conversation.",
	})

	require.NoError(t, err)
	require.NotNil(t, rating)

	assert.NotEmpty(t, rating.ID)
	assert.Equal(t, userA.ID, rating.RaterID)
	assert.Equal(t, userB.ID, rating.RateeID)
	assert.Equal(t, model.TrustLevelTrust, rating.TrustLevel)
	assert.Equal(t, model.ReviewVisibilityPublic, rating.ReviewVisibility)
}

func TestTrustRating_CreateDistrust(t *testing.T) {
	// AC-TRUST-002: Create Distrust Rating
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	userA := f.CreateUser(t)
	userB := f.CreateUser(t)
	event := setupVerifiedEventWithBothAttendees(t, f, tdb, userA, userB)

	// User A rates User B as distrust
	rating, err := trustService.Create(ctx, userA.ID, &model.CreateTrustRatingRequest{
		RateeID:     userB.ID,
		AnchorType:  string(model.TrustAnchorEvent),
		AnchorID:    event.ID,
		TrustLevel:  string(model.TrustLevelDistrust),
		TrustReview: "Made me feel uncomfortable during the event.",
	})

	require.NoError(t, err)
	require.NotNil(t, rating)

	assert.Equal(t, model.TrustLevelDistrust, rating.TrustLevel)
	// Distrust reviews should be admin-only
	assert.Equal(t, model.ReviewVisibilityAdminOnly, rating.ReviewVisibility)
}

func TestTrustRating_CannotRateWithoutAnchor(t *testing.T) {
	// AC-TRUST-003: Cannot Rate Without Anchor
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	userA := f.CreateUser(t)
	userB := f.CreateUser(t)

	// Create the function that always returns false
	fnQuery := `
		DEFINE FUNCTION OVERWRITE fn::can_rate_user($rater_id: string, $ratee_id: string, $anchor_type: string, $anchor_id: string) {
			RETURN false;
		}
	`
	_ = tdb.DB.Execute(ctx, fnQuery, nil)

	// Try to rate without a verified event anchor
	_, err := trustService.Create(ctx, userA.ID, &model.CreateTrustRatingRequest{
		RateeID:     userB.ID,
		AnchorType:  string(model.TrustAnchorEvent),
		AnchorID:    "event:nonexistent",
		TrustLevel:  string(model.TrustLevelTrust),
		TrustReview: "This should fail",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot rate")
}

func TestTrustRating_CannotSelfRate(t *testing.T) {
	// AC-TRUST-004: Cannot Self-Rate
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t)

	// Try to rate yourself
	_, err := trustService.Create(ctx, user.ID, &model.CreateTrustRatingRequest{
		RateeID:     user.ID,
		AnchorType:  string(model.TrustAnchorEvent),
		AnchorID:    "event:fake",
		TrustLevel:  string(model.TrustLevelTrust),
		TrustReview: "I'm awesome!",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot rate yourself")
}

func TestTrustRating_CannotRateTwiceForSameAnchor(t *testing.T) {
	// AC-TRUST-005: Cannot Rate Same User Twice (for same anchor)
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	userA := f.CreateUser(t)
	userB := f.CreateUser(t)
	event := setupVerifiedEventWithBothAttendees(t, f, tdb, userA, userB)

	// First rating
	_, err := trustService.Create(ctx, userA.ID, &model.CreateTrustRatingRequest{
		RateeID:     userB.ID,
		AnchorType:  string(model.TrustAnchorEvent),
		AnchorID:    event.ID,
		TrustLevel:  string(model.TrustLevelTrust),
		TrustReview: "First rating",
	})
	require.NoError(t, err)

	// Try to rate again for same anchor
	_, err = trustService.Create(ctx, userA.ID, &model.CreateTrustRatingRequest{
		RateeID:     userB.ID,
		AnchorType:  string(model.TrustAnchorEvent),
		AnchorID:    event.ID,
		TrustLevel:  string(model.TrustLevelDistrust),
		TrustReview: "Changed my mind",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestTrustRating_ReviewRequired(t *testing.T) {
	// AC-TRUST-006: Review Required
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	userA := f.CreateUser(t)
	userB := f.CreateUser(t)

	// Try to rate without review
	_, err := trustService.Create(ctx, userA.ID, &model.CreateTrustRatingRequest{
		RateeID:     userB.ID,
		AnchorType:  string(model.TrustAnchorEvent),
		AnchorID:    "event:fake",
		TrustLevel:  string(model.TrustLevelTrust),
		TrustReview: "", // Empty review
	})

	require.Error(t, err)
	// Validation error - ProblemDetails
	var problemErr *model.ProblemDetails
	assert.ErrorAs(t, err, &problemErr)
}

func TestTrustRating_ReviewMaxLength(t *testing.T) {
	// AC-TRUST-007: Review Max Length
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	userA := f.CreateUser(t)
	userB := f.CreateUser(t)

	// Try to rate with review > 240 chars
	longReview := strings.Repeat("a", model.MaxTrustReviewLength+1)
	_, err := trustService.Create(ctx, userA.ID, &model.CreateTrustRatingRequest{
		RateeID:     userB.ID,
		AnchorType:  string(model.TrustAnchorEvent),
		AnchorID:    "event:fake",
		TrustLevel:  string(model.TrustLevelTrust),
		TrustReview: longReview,
	})

	require.Error(t, err)
	var problemErr *model.ProblemDetails
	assert.ErrorAs(t, err, &problemErr)
}

func TestTrustRating_UpdateCooldown(t *testing.T) {
	// AC-TRUST-008: Update Rating - 30 Day Cooldown
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	userA := f.CreateUser(t)
	userB := f.CreateUser(t)
	event := setupVerifiedEventWithBothAttendees(t, f, tdb, userA, userB)

	// Create initial rating
	rating, err := trustService.Create(ctx, userA.ID, &model.CreateTrustRatingRequest{
		RateeID:     userB.ID,
		AnchorType:  string(model.TrustAnchorEvent),
		AnchorID:    event.ID,
		TrustLevel:  string(model.TrustLevelTrust),
		TrustReview: "Initial rating",
	})
	require.NoError(t, err)

	// Try to update within 30 days
	newLevel := string(model.TrustLevelDistrust)
	_, err = trustService.Update(ctx, rating.ID, userA.ID, &model.UpdateTrustRatingRequest{
		TrustLevel: &newLevel,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be changed until")
}

func TestTrustRating_UpdateAfterCooldown(t *testing.T) {
	// AC-TRUST-009: Update Rating - After Cooldown
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	userA := f.CreateUser(t)
	userB := f.CreateUser(t)
	event := setupVerifiedEventWithBothAttendees(t, f, tdb, userA, userB)

	// Create initial rating
	rating, err := trustService.Create(ctx, userA.ID, &model.CreateTrustRatingRequest{
		RateeID:     userB.ID,
		AnchorType:  string(model.TrustAnchorEvent),
		AnchorID:    event.ID,
		TrustLevel:  string(model.TrustLevelTrust),
		TrustReview: "Initial rating",
	})
	require.NoError(t, err)

	// Backdate the rating to 31 days ago
	backdateQuery := `UPDATE type::record($id) SET updated_on = $old_date`
	oldDate := time.Now().AddDate(0, 0, -(model.TrustRatingCooldownDays + 1))
	err = tdb.DB.Execute(ctx, backdateQuery, map[string]interface{}{
		"id":       rating.ID,
		"old_date": oldDate,
	})
	require.NoError(t, err)

	// Now update should succeed
	newLevel := string(model.TrustLevelDistrust)
	newReview := "After more interactions, I changed my mind"
	updated, err := trustService.Update(ctx, rating.ID, userA.ID, &model.UpdateTrustRatingRequest{
		TrustLevel:  &newLevel,
		TrustReview: &newReview,
	})

	require.NoError(t, err)
	assert.Equal(t, model.TrustLevelDistrust, updated.TrustLevel)
}

func TestTrustRating_Delete(t *testing.T) {
	// AC-TRUST-010: Delete Rating (Unset)
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	userA := f.CreateUser(t)
	userB := f.CreateUser(t)
	event := setupVerifiedEventWithBothAttendees(t, f, tdb, userA, userB)

	// Create rating
	rating, err := trustService.Create(ctx, userA.ID, &model.CreateTrustRatingRequest{
		RateeID:     userB.ID,
		AnchorType:  string(model.TrustAnchorEvent),
		AnchorID:    event.ID,
		TrustLevel:  string(model.TrustLevelTrust),
		TrustReview: "Rating to delete",
	})
	require.NoError(t, err)

	// Delete the rating
	err = trustService.Delete(ctx, rating.ID, userA.ID)
	require.NoError(t, err)

	// Verify it's gone
	_, err = trustService.GetByID(ctx, rating.ID, userA.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestTrustRating_DeleteNotOwner(t *testing.T) {
	// AC-TRUST-010 (variation): Cannot delete someone else's rating
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	userA := f.CreateUser(t)
	userB := f.CreateUser(t)
	userC := f.CreateUser(t)
	event := setupVerifiedEventWithBothAttendees(t, f, tdb, userA, userB)

	// Create rating from A to B
	rating, err := trustService.Create(ctx, userA.ID, &model.CreateTrustRatingRequest{
		RateeID:     userB.ID,
		AnchorType:  string(model.TrustAnchorEvent),
		AnchorID:    event.ID,
		TrustLevel:  string(model.TrustLevelTrust),
		TrustReview: "My rating",
	})
	require.NoError(t, err)

	// User C tries to delete
	err = trustService.Delete(ctx, rating.ID, userC.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not your rating")
}

func TestTrustRating_DailyLimit(t *testing.T) {
	// AC-TRUST-011: Daily Rate Limit
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	rater := f.CreateUser(t)

	// Seed the daily count to be at limit
	today := time.Now().Format("2006-01-02")
	seedQuery := `
		CREATE trust_rating_daily_count CONTENT {
			user_id: type::record($user_id),
			date: $date,
			count: $count
		}
	`
	err := tdb.DB.Execute(ctx, seedQuery, map[string]interface{}{
		"user_id": rater.ID,
		"date":    today,
		"count":   model.MaxTrustRatingsPerDay,
	})
	require.NoError(t, err)

	// Try to create another rating
	ratee := f.CreateUser(t)
	_, err = trustService.Create(ctx, rater.ID, &model.CreateTrustRatingRequest{
		RateeID:     ratee.ID,
		AnchorType:  string(model.TrustAnchorEvent),
		AnchorID:    "event:fake",
		TrustLevel:  string(model.TrustLevelTrust),
		TrustReview: "One too many",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "daily rating limit")
}

func TestTrustRating_Endorsement(t *testing.T) {
	// AC-TRUST-012: Endorsement - Valid
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	userA := f.CreateUser(t)
	userB := f.CreateUser(t)
	userC := f.CreateUser(t) // Endorser

	// Setup event with all three users
	guild := f.CreatePublicGuild(t, userA)
	f.AddMemberToGuild(t, userB, guild)
	f.AddMemberToGuild(t, userC, guild)

	event := f.CreateVerifiedEvent(t, guild, userA)

	f.CreateRSVP(t, event, userA, model.RSVPStatusApproved)
	f.CreateRSVP(t, event, userB, model.RSVPStatusApproved)
	f.CreateRSVP(t, event, userC, model.RSVPStatusApproved)

	f.ConfirmEventCompletion(t, event, userA)
	f.ConfirmEventCompletion(t, event, userB)
	f.ConfirmEventCompletion(t, event, userC)

	// Create the can_rate_user function
	fnQuery := `
		DEFINE FUNCTION OVERWRITE fn::can_rate_user($rater_id: string, $ratee_id: string, $anchor_type: string, $anchor_id: string) {
			IF $anchor_type = "event" {
				LET $event_rec = type::record($anchor_id);
				LET $rater_rec = type::record($rater_id);
				LET $ratee_rec = type::record($ratee_id);
				LET $event = (SELECT completion_verified FROM $event_rec);
				IF $event[0].completion_verified != true {
					RETURN false;
				};
				LET $rater_rsvp = (SELECT * FROM event_rsvp WHERE event_id = $event_rec AND user_id = $rater_rec AND status = "approved" AND completion_confirmed IS NOT NULL LIMIT 1);
				LET $ratee_rsvp = (SELECT * FROM event_rsvp WHERE event_id = $event_rec AND user_id = $ratee_rec AND status = "approved" AND completion_confirmed IS NOT NULL LIMIT 1);
				RETURN array::len($rater_rsvp) > 0 AND array::len($ratee_rsvp) > 0;
			};
			RETURN false;
		}
	`
	_ = tdb.DB.Execute(ctx, fnQuery, nil)

	// User A rates User B
	rating, err := trustService.Create(ctx, userA.ID, &model.CreateTrustRatingRequest{
		RateeID:     userB.ID,
		AnchorType:  string(model.TrustAnchorEvent),
		AnchorID:    event.ID,
		TrustLevel:  string(model.TrustLevelTrust),
		TrustReview: "Great person!",
	})
	require.NoError(t, err)

	// User C endorses the rating (they also attended)
	endorsement, err := trustService.CreateEndorsement(ctx, rating.ID, userC.ID, &model.CreateEndorsementRequest{
		EndorsementType: string(model.EndorsementAgree),
	})

	require.NoError(t, err)
	require.NotNil(t, endorsement)
	assert.Equal(t, rating.ID, endorsement.TrustRatingID)
	assert.Equal(t, userC.ID, endorsement.EndorserID)
	assert.Equal(t, model.EndorsementAgree, endorsement.EndorsementType)
}

func TestTrustRating_EndorsementNotAttendee(t *testing.T) {
	// AC-TRUST-013: Endorsement - Not Attendee
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	userA := f.CreateUser(t)
	userB := f.CreateUser(t)
	nonAttendee := f.CreateUser(t)

	event := setupVerifiedEventWithBothAttendees(t, f, tdb, userA, userB)

	// User A rates User B
	rating, err := trustService.Create(ctx, userA.ID, &model.CreateTrustRatingRequest{
		RateeID:     userB.ID,
		AnchorType:  string(model.TrustAnchorEvent),
		AnchorID:    event.ID,
		TrustLevel:  string(model.TrustLevelTrust),
		TrustReview: "Great person!",
	})
	require.NoError(t, err)

	// Non-attendee tries to endorse
	_, err = trustService.CreateEndorsement(ctx, rating.ID, nonAttendee.ID, &model.CreateEndorsementRequest{
		EndorsementType: string(model.EndorsementAgree),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must have attended")
}

func TestTrustRating_EndorsementNoDuplicates(t *testing.T) {
	// AC-TRUST-014: Endorsement - No Duplicates
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	userA := f.CreateUser(t)
	userB := f.CreateUser(t)
	userC := f.CreateUser(t)

	// Setup event with all three users
	guild := f.CreatePublicGuild(t, userA)
	f.AddMemberToGuild(t, userB, guild)
	f.AddMemberToGuild(t, userC, guild)

	event := f.CreateVerifiedEvent(t, guild, userA)

	f.CreateRSVP(t, event, userA, model.RSVPStatusApproved)
	f.CreateRSVP(t, event, userB, model.RSVPStatusApproved)
	f.CreateRSVP(t, event, userC, model.RSVPStatusApproved)

	f.ConfirmEventCompletion(t, event, userA)
	f.ConfirmEventCompletion(t, event, userB)
	f.ConfirmEventCompletion(t, event, userC)

	// Create the can_rate_user function
	fnQuery := `
		DEFINE FUNCTION OVERWRITE fn::can_rate_user($rater_id: string, $ratee_id: string, $anchor_type: string, $anchor_id: string) {
			IF $anchor_type = "event" {
				LET $event_rec = type::record($anchor_id);
				LET $rater_rec = type::record($rater_id);
				LET $ratee_rec = type::record($ratee_id);
				LET $event = (SELECT completion_verified FROM $event_rec);
				IF $event[0].completion_verified != true {
					RETURN false;
				};
				LET $rater_rsvp = (SELECT * FROM event_rsvp WHERE event_id = $event_rec AND user_id = $rater_rec AND status = "approved" AND completion_confirmed IS NOT NULL LIMIT 1);
				LET $ratee_rsvp = (SELECT * FROM event_rsvp WHERE event_id = $event_rec AND user_id = $ratee_rec AND status = "approved" AND completion_confirmed IS NOT NULL LIMIT 1);
				RETURN array::len($rater_rsvp) > 0 AND array::len($ratee_rsvp) > 0;
			};
			RETURN false;
		}
	`
	_ = tdb.DB.Execute(ctx, fnQuery, nil)

	rating, err := trustService.Create(ctx, userA.ID, &model.CreateTrustRatingRequest{
		RateeID:     userB.ID,
		AnchorType:  string(model.TrustAnchorEvent),
		AnchorID:    event.ID,
		TrustLevel:  string(model.TrustLevelTrust),
		TrustReview: "Great person!",
	})
	require.NoError(t, err)

	// First endorsement
	_, err = trustService.CreateEndorsement(ctx, rating.ID, userC.ID, &model.CreateEndorsementRequest{
		EndorsementType: string(model.EndorsementAgree),
	})
	require.NoError(t, err)

	// Second endorsement - should fail
	_, err = trustService.CreateEndorsement(ctx, rating.ID, userC.ID, &model.CreateEndorsementRequest{
		EndorsementType: string(model.EndorsementDisagree),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already endorsed")
}

func TestTrustRating_GetAggregate(t *testing.T) {
	// AC-TRUST-015: Get Trust Aggregate
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	// Create target user
	targetUser := f.CreateUser(t)

	// Create ratings directly in the database for aggregate test
	for i := 0; i < 5; i++ {
		rater := f.CreateUser(t)
		query := `
			CREATE trust_rating CONTENT {
				rater_id: type::record($rater_id),
				ratee_id: type::record($ratee_id),
				anchor_type: "event",
				anchor_id: $anchor_id,
				trust_level: "trust",
				trust_review: $review,
				review_visibility: "public",
				created_on: time::now(),
				updated_on: time::now()
			}
		`
		err := tdb.DB.Execute(ctx, query, map[string]interface{}{
			"rater_id":  rater.ID,
			"ratee_id":  targetUser.ID,
			"anchor_id": fmt.Sprintf("event:test%d", i),
			"review":    fmt.Sprintf("Trust review %d", i),
		})
		require.NoError(t, err)
	}

	// Add one distrust rating
	distruster := f.CreateUser(t)
	query := `
		CREATE trust_rating CONTENT {
			rater_id: type::record($rater_id),
			ratee_id: type::record($ratee_id),
			anchor_type: "event",
			anchor_id: "event:distrust",
			trust_level: "distrust",
			trust_review: "Distrust review",
			review_visibility: "admin_only",
			created_on: time::now(),
			updated_on: time::now()
		}
	`
	err := tdb.DB.Execute(ctx, query, map[string]interface{}{
		"rater_id": distruster.ID,
		"ratee_id": targetUser.ID,
	})
	require.NoError(t, err)

	// Get aggregate
	aggregate, err := trustService.GetAggregate(ctx, targetUser.ID)

	require.NoError(t, err)
	require.NotNil(t, aggregate)

	assert.Equal(t, 5, aggregate.TrustCount)
	assert.Equal(t, 1, aggregate.DistrustCount)
	assert.Equal(t, 4, aggregate.NetTrust) // 5 - 1 = 4
}

func TestTrustRating_DistrustSignals(t *testing.T) {
	// AC-TRUST-016: Admin Distrust Signals
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	// Create a user with high distrust count
	targetUser := f.CreateUser(t)

	// Create 5 distrust ratings for this user
	for i := 0; i < 5; i++ {
		rater := f.CreateUser(t)
		query := `
			CREATE trust_rating CONTENT {
				rater_id: type::record($rater_id),
				ratee_id: type::record($ratee_id),
				anchor_type: "event",
				anchor_id: $anchor_id,
				trust_level: "distrust",
				trust_review: $review,
				review_visibility: "admin_only",
				created_on: time::now(),
				updated_on: time::now()
			}
		`
		err := tdb.DB.Execute(ctx, query, map[string]interface{}{
			"rater_id":  rater.ID,
			"ratee_id":  targetUser.ID,
			"anchor_id": fmt.Sprintf("event:distrust%d", i),
			"review":    fmt.Sprintf("Concerning behavior %d", i),
		})
		require.NoError(t, err)
	}

	// Query distrust signals (min 3 distrust)
	signals, err := trustService.GetDistrustSignals(ctx, 3, 10)

	require.NoError(t, err)
	require.NotEmpty(t, signals)

	// Find our target user in signals
	var found bool
	for _, signal := range signals {
		if signal.UserID == targetUser.ID {
			found = true
			assert.GreaterOrEqual(t, signal.DistrustCount, 3)
			break
		}
	}
	assert.True(t, found, "Target user should be in distrust signals")
}

func TestTrustRating_CannotEndorseOwnRating(t *testing.T) {
	// Additional test: Cannot endorse your own rating
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	trustService := createTrustRatingService(t, tdb)
	ctx := context.Background()

	userA := f.CreateUser(t)
	userB := f.CreateUser(t)
	event := setupVerifiedEventWithBothAttendees(t, f, tdb, userA, userB)

	// User A rates User B
	rating, err := trustService.Create(ctx, userA.ID, &model.CreateTrustRatingRequest{
		RateeID:     userB.ID,
		AnchorType:  string(model.TrustAnchorEvent),
		AnchorID:    event.ID,
		TrustLevel:  string(model.TrustLevelTrust),
		TrustReview: "Great!",
	})
	require.NoError(t, err)

	// User A tries to endorse their own rating
	_, err = trustService.CreateEndorsement(ctx, rating.ID, userA.ID, &model.CreateEndorsementRequest{
		EndorsementType: string(model.EndorsementAgree),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot endorse your own rating")
}
